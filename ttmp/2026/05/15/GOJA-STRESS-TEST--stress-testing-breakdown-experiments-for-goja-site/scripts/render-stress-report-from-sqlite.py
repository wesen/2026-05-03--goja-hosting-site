#!/usr/bin/env python3
"""Render a Markdown stress-test report from the SQLite benchmark store.

Every result section embeds the SQL query used to produce it. The queries are
focused on identifying breakdown: success loss, throughput shortfall, latency
knees, tail outliers, and route/module metric evidence.
"""

from __future__ import annotations

import argparse
import datetime as dt
import sqlite3
from pathlib import Path
from typing import Sequence


def format_cell(value: object) -> str:
    if value is None:
        return ""
    text = str(value).replace("\n", "<br>")
    return text.replace("|", "\\|")


def markdown_table(headers: Sequence[str], rows: Sequence[Sequence[object]]) -> str:
    out = ["| " + " | ".join(headers) + " |", "|" + "|".join("---" for _ in headers) + "|"]
    for row in rows:
        out.append("| " + " | ".join(format_cell(v) for v in row) + " |")
    return "\n".join(out)


SECTIONS: list[tuple[str, str, str]] = [
    (
        "Matrix metadata",
        "What was run, where the ignored raw artifacts live, and which commit produced the data.",
        """
SELECT matrix_id, created_at_utc, imported_at_utc, repo_commit, git_dirty,
       duration, warmup_duration, scenarios, rates, repeat_count, out_root
FROM benchmark_matrices
WHERE matrix_id = :matrix_id;
""".strip(),
    ),
    (
        "Offered vs achieved throughput and latency",
        "Primary stress table: compares requested rate to achieved throughput, success, and tail latency.",
        """
SELECT
  scenario,
  rate_target,
  CAST(REPLACE(rate_target, '/s', '') AS REAL) AS offered_rps,
  COUNT(*) AS runs,
  SUM(requests) AS requests,
  ROUND(AVG(throughput), 2) AS throughput_avg,
  ROUND(AVG(throughput) / NULLIF(CAST(REPLACE(rate_target, '/s', '') AS REAL), 0), 3) AS throughput_ratio,
  ROUND(AVG(success_ratio), 4) AS success_avg,
  ROUND(AVG(p50_ms), 2) AS p50_avg_ms,
  ROUND(AVG(p95_ms), 2) AS p95_avg_ms,
  ROUND(AVG(p99_ms), 2) AS p99_avg_ms,
  ROUND(MAX(max_ms), 2) AS max_ms
FROM benchmark_runs
WHERE matrix_id = :matrix_id
GROUP BY scenario, rate_target
ORDER BY scenario, offered_rps;
""".strip(),
    ),
    (
        "Breakdown candidates",
        "Flags cells where success falls, throughput misses offered rate, p95 exceeds conservative stress thresholds, or max latency is very high.",
        """
WITH agg AS (
  SELECT
    scenario,
    rate_target,
    CAST(REPLACE(rate_target, '/s', '') AS REAL) AS offered_rps,
    COUNT(*) AS runs,
    MIN(success_ratio) AS success_min,
    AVG(success_ratio) AS success_avg,
    AVG(throughput) AS throughput_avg,
    AVG(throughput) / NULLIF(CAST(REPLACE(rate_target, '/s', '') AS REAL), 0) AS throughput_ratio,
    AVG(p95_ms) AS p95_avg_ms,
    MAX(max_ms) AS max_ms,
    MAX(errors_json) AS errors_json
  FROM benchmark_runs
  WHERE matrix_id = :matrix_id
  GROUP BY scenario, rate_target
)
SELECT
  scenario,
  rate_target,
  ROUND(success_min, 4) AS success_min,
  ROUND(throughput_ratio, 3) AS throughput_ratio,
  ROUND(p95_avg_ms, 2) AS p95_avg_ms,
  ROUND(max_ms, 2) AS max_ms,
  CASE
    WHEN success_min < 0.99 THEN 'success_loss'
    WHEN errors_json <> '[]' THEN 'vegeta_errors'
    WHEN throughput_ratio < 0.95 THEN 'throughput_shortfall'
    WHEN scenario IN ('null', 'render', 'db-read') AND p95_avg_ms > 100 THEN 'latency_threshold'
    WHEN scenario IN ('db-write', 'kanban-fragment', 'kanban-action', 'kanban-mixed') AND p95_avg_ms > 250 THEN 'latency_threshold'
    WHEN max_ms > 1000 THEN 'tail_outlier'
    ELSE 'ok'
  END AS breakdown_signal,
  errors_json
FROM agg
WHERE breakdown_signal <> 'ok'
ORDER BY scenario, offered_rps;
""".strip(),
    ),
    (
        "Latency knee growth between adjacent rates",
        "Looks for rate-to-rate p95 growth. Large growth factors identify where a scenario starts bending even before errors appear.",
        """
WITH agg AS (
  SELECT
    scenario,
    rate_target,
    CAST(REPLACE(rate_target, '/s', '') AS REAL) AS offered_rps,
    AVG(p95_ms) AS p95_avg_ms,
    AVG(p99_ms) AS p99_avg_ms,
    AVG(throughput) AS throughput_avg,
    AVG(success_ratio) AS success_avg
  FROM benchmark_runs
  WHERE matrix_id = :matrix_id
  GROUP BY scenario, rate_target
), ordered AS (
  SELECT
    *,
    LAG(p95_avg_ms) OVER (PARTITION BY scenario ORDER BY offered_rps) AS previous_p95_avg_ms,
    LAG(offered_rps) OVER (PARTITION BY scenario ORDER BY offered_rps) AS previous_offered_rps
  FROM agg
)
SELECT
  scenario,
  previous_offered_rps || '/s -> ' || rate_target AS rate_step,
  ROUND(previous_p95_avg_ms, 2) AS previous_p95_ms,
  ROUND(p95_avg_ms, 2) AS p95_ms,
  ROUND(p95_avg_ms / NULLIF(previous_p95_avg_ms, 0), 2) AS p95_growth_factor,
  ROUND(p99_avg_ms, 2) AS p99_ms,
  ROUND(throughput_avg, 2) AS throughput_avg,
  ROUND(success_avg, 4) AS success_avg
FROM ordered
WHERE previous_p95_avg_ms IS NOT NULL
ORDER BY p95_growth_factor DESC, scenario, offered_rps;
""".strip(),
    ),
    (
        "Slowest individual runs by p95",
        "The noisiest runs and likely candidates for targeted pprof reruns.",
        """
SELECT
  scenario,
  rate_target,
  run_number,
  requests,
  ROUND(success_ratio, 4) AS success_ratio,
  ROUND(throughput, 2) AS throughput,
  ROUND(p50_ms, 2) AS p50_ms,
  ROUND(p95_ms, 2) AS p95_ms,
  ROUND(p99_ms, 2) AS p99_ms,
  ROUND(max_ms, 2) AS max_ms,
  status_codes_json,
  errors_json
FROM benchmark_runs
WHERE matrix_id = :matrix_id
ORDER BY p95_ms DESC
LIMIT 25;
""".strip(),
    ),
    (
        "Runs with non-100% success or Vegeta errors",
        "Hard failures: anything here needs immediate investigation before interpreting latency numbers.",
        """
SELECT
  scenario,
  rate_target,
  run_number,
  requests,
  ROUND(success_ratio, 4) AS success_ratio,
  status_codes_json,
  errors_json,
  out_dir
FROM benchmark_runs
WHERE matrix_id = :matrix_id
  AND (success_ratio < 1 OR errors_json <> '[]')
ORDER BY scenario, CAST(REPLACE(rate_target, '/s', '') AS REAL), run_number;
""".strip(),
    ),
    (
        "HTTP route deltas",
        "Prometheus route counters prove which route classes were actually hit at each stress rate.",
        """
SELECT
  scenario,
  rate_target,
  labels_json,
  SUM(delta_value) AS delta_sum
FROM benchmark_metric_deltas
WHERE matrix_id = :matrix_id
  AND metric_name = 'goja_site_http_requests_total'
GROUP BY scenario, rate_target, labels_json
ORDER BY scenario, CAST(REPLACE(rate_target, '/s', '') AS REAL), labels_json;
""".strip(),
    ),
    (
        "DB operation and error deltas",
        "DB work performed and DB error evidence by scenario/rate.",
        """
SELECT
  scenario,
  rate_target,
  metric_name,
  labels_json,
  SUM(delta_value) AS delta_sum
FROM benchmark_metric_deltas
WHERE matrix_id = :matrix_id
  AND metric_name IN ('goja_site_db_operations_total', 'goja_site_db_errors_total')
GROUP BY scenario, rate_target, metric_name, labels_json
ORDER BY scenario, CAST(REPLACE(rate_target, '/s', '') AS REAL), metric_name, labels_json;
""".strip(),
    ),
    (
        "Kanban action/render deltas",
        "Kanban work performed and action/fragment/render evidence by scenario/rate.",
        """
SELECT
  scenario,
  rate_target,
  metric_name,
  labels_json,
  SUM(delta_value) AS delta_sum
FROM benchmark_metric_deltas
WHERE matrix_id = :matrix_id
  AND metric_name IN (
    'goja_site_kanban_action_duration_seconds_count',
    'goja_site_kanban_fragment_duration_seconds_count',
    'goja_site_kanban_render_duration_seconds_count',
    'goja_site_kanban_errors_total'
  )
GROUP BY scenario, rate_target, metric_name, labels_json
ORDER BY scenario, CAST(REPLACE(rate_target, '/s', '') AS REAL), metric_name, labels_json;
""".strip(),
    ),
    (
        "Stored artifact locations",
        "Per-run raw artifacts. Use these directories for server logs, pprof files, raw Vegeta data, and metrics snapshots.",
        """
SELECT scenario, rate_target, run_number, out_dir
FROM benchmark_runs
WHERE matrix_id = :matrix_id
ORDER BY scenario, CAST(REPLACE(rate_target, '/s', '') AS REAL), run_number;
""".strip(),
    ),
]


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--db", required=True, help="SQLite benchmark database path")
    parser.add_argument("--matrix-id", required=True)
    parser.add_argument("--out", required=True, help="Markdown file to write")
    parser.add_argument("--title", default="Stress Test SQLite Report")
    args = parser.parse_args()

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)

    with sqlite3.connect(args.db) as conn:
        conn.row_factory = sqlite3.Row
        lines = [
            "---",
            f"Title: {args.title} - {args.matrix_id}",
            "Ticket: GOJA-STRESS-TEST",
            "Status: active",
            "Topics:",
            "    - benchmarking",
            "    - stress-testing",
            "    - sqlite",
            "    - vegeta",
            "    - prometheus",
            "DocType: reference",
            "Intent: historical",
            'Summary: "SQLite-backed stress test report with embedded SQL queries for breakdown analysis."',
            f"LastUpdated: {dt.datetime.now(dt.UTC).strftime('%Y-%m-%dT%H:%M:%SZ')}",
            "---",
            "",
            f"# {args.title}: {args.matrix_id}",
            "",
            f"This report was generated from SQLite database `{Path(args.db).resolve()}`.",
            "",
            "Each section includes the exact SQL query used to generate the table.",
            "",
            "## Breakdown criteria used in this report",
            "",
            "The report flags a breakdown candidate when any grouped scenario/rate cell shows one of:",
            "",
            "- `success_min < 0.99`,",
            "- non-empty Vegeta errors,",
            "- achieved throughput below 95% of offered rate,",
            "- p95 above 100 ms for light scenarios (`null`, `render`, `db-read`),",
            "- p95 above 250 ms for heavier scenarios (`db-write`, Kanban scenarios),",
            "- max latency above 1000 ms.",
            "",
        ]

        for title, explanation, query in SECTIONS:
            cur = conn.execute(query, {"matrix_id": args.matrix_id})
            rows = cur.fetchall()
            headers = [d[0] for d in cur.description]
            lines.extend([
                f"## {title}",
                "",
                explanation,
                "",
                "```sql",
                query,
                "```",
                "",
            ])
            if rows:
                lines.append(markdown_table(headers, [[row[h] for h in headers] for row in rows]))
            else:
                lines.append("_No rows._")
            lines.append("")

    out_path.write_text("\n".join(lines), encoding="utf-8")
    print(f"wrote report: {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
