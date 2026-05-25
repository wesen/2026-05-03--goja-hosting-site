#!/usr/bin/env python3
"""Render a Markdown benchmark report from the SQLite benchmark store."""

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


SECTIONS: list[tuple[str, str]] = [
    (
        "Matrix metadata",
        """
SELECT matrix_id, created_at_utc, imported_at_utc, repo_commit, git_dirty,
       duration, warmup_duration, scenarios, rates, repeat_count, out_root
FROM benchmark_matrices
WHERE matrix_id = :matrix_id;
""".strip(),
    ),
    (
        "Scenario/rate aggregate latency and success",
        """
SELECT
  scenario,
  rate_target,
  COUNT(*) AS runs,
  SUM(requests) AS requests,
  ROUND(AVG(success_ratio), 4) AS success_avg,
  ROUND(MIN(p95_ms), 2) AS p95_min_ms,
  ROUND(AVG(p95_ms), 2) AS p95_avg_ms,
  ROUND(MAX(p95_ms), 2) AS p95_max_ms,
  ROUND(AVG(p99_ms), 2) AS p99_avg_ms,
  ROUND(AVG(throughput), 2) AS throughput_avg
FROM benchmark_runs
WHERE matrix_id = :matrix_id
GROUP BY scenario, rate_target
ORDER BY scenario, rate_target;
""".strip(),
    ),
    (
        "Slowest individual runs by p95",
        """
SELECT
  scenario,
  rate_target,
  run_number,
  requests,
  ROUND(success_ratio, 4) AS success_ratio,
  ROUND(p50_ms, 2) AS p50_ms,
  ROUND(p95_ms, 2) AS p95_ms,
  ROUND(p99_ms, 2) AS p99_ms,
  ROUND(max_ms, 2) AS max_ms,
  status_codes_json,
  errors_json
FROM benchmark_runs
WHERE matrix_id = :matrix_id
ORDER BY p95_ms DESC
LIMIT 20;
""".strip(),
    ),
    (
        "Runs with non-100% success or errors",
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
ORDER BY scenario, rate_target, run_number;
""".strip(),
    ),
    (
        "HTTP request deltas captured from Prometheus metrics",
        """
SELECT
  scenario,
  rate_target,
  metric_name,
  labels_json,
  SUM(delta_value) AS delta_sum
FROM benchmark_metric_deltas
WHERE matrix_id = :matrix_id
  AND metric_name = 'goja_site_http_requests_total'
GROUP BY scenario, rate_target, metric_name, labels_json
ORDER BY scenario, rate_target, labels_json;
""".strip(),
    ),
    (
        "DB operation deltas captured from Prometheus metrics",
        """
SELECT
  scenario,
  rate_target,
  labels_json,
  SUM(delta_value) AS delta_sum
FROM benchmark_metric_deltas
WHERE matrix_id = :matrix_id
  AND metric_name = 'goja_site_db_operations_total'
GROUP BY scenario, rate_target, labels_json
ORDER BY scenario, rate_target, labels_json;
""".strip(),
    ),
    (
        "Kanban action and fragment metric deltas",
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
    'goja_site_kanban_render_duration_seconds_count'
  )
GROUP BY scenario, rate_target, metric_name, labels_json
ORDER BY scenario, rate_target, metric_name, labels_json;
""".strip(),
    ),
    (
        "Stored artifact locations",
        """
SELECT scenario, rate_target, run_number, out_dir
FROM benchmark_runs
WHERE matrix_id = :matrix_id
ORDER BY scenario, rate_target, run_number;
""".strip(),
    ),
]


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--db", required=True, help="SQLite benchmark database path")
    parser.add_argument("--matrix-id", required=True)
    parser.add_argument("--out", required=True, help="Markdown file to write")
    args = parser.parse_args()

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)

    with sqlite3.connect(args.db) as conn:
        conn.row_factory = sqlite3.Row
        lines = [
            "---",
            f"Title: Benchmark Matrix SQLite Report - {args.matrix_id}",
            "Ticket: GOJA-PERF-BENCH",
            "Status: active",
            "Topics:",
            "    - benchmarking",
            "    - sqlite",
            "    - vegeta",
            "    - prometheus",
            "DocType: reference",
            "Intent: historical",
            'Summary: "SQLite-backed benchmark matrix report generated from stored Vegeta and metrics-delta results."',
            f"LastUpdated: {dt.datetime.now(dt.UTC).strftime('%Y-%m-%dT%H:%M:%SZ')}",
            "---",
            "",
            f"# Benchmark Matrix SQLite Report: {args.matrix_id}",
            "",
            f"This report was generated from SQLite database `{Path(args.db).resolve()}`.",
            "",
            "Every section includes the SQL query used to generate it, followed by the query result.",
            "",
        ]

        for title, query in SECTIONS:
            cur = conn.execute(query, {"matrix_id": args.matrix_id})
            rows = cur.fetchall()
            headers = [d[0] for d in cur.description]
            lines.extend([
                f"## {title}",
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
