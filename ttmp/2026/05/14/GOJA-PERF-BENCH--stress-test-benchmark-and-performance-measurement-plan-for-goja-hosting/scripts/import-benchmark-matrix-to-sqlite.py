#!/usr/bin/env python3
"""Import a goja-site bench-matrix result directory into SQLite."""

from __future__ import annotations

import argparse
import csv
import datetime as dt
import json
import os
import re
import shlex
import sqlite3
import sys
from pathlib import Path
from typing import Any

METRIC_RE = re.compile(r"^([a-zA-Z_:][a-zA-Z0-9_:]*)(\{[^}]*\})?\s+([-+0-9.eE]+)$")


def read_text(path: Path) -> str:
    if not path.exists():
        return ""
    return path.read_text(encoding="utf-8")


def read_json(path: Path) -> Any:
    if not path.exists():
        return {}
    with path.open(encoding="utf-8") as f:
        return json.load(f)


def load_runs(matrix_root: Path) -> list[dict[str, str]]:
    rows: list[dict[str, str]] = []
    with (matrix_root / "runs.tsv").open(encoding="utf-8") as f:
        for scenario, rate, run, out_dir in csv.reader(f, delimiter="\t"):
            rows.append({"scenario": scenario, "rate": rate, "run": run, "out_dir": out_dir})
    return rows


def parse_metric_delta_line(line: str) -> tuple[str, dict[str, str], float] | None:
    m = METRIC_RE.match(line.strip())
    if not m:
        return None
    name, labels_text, value_text = m.groups()
    labels: dict[str, str] = {}
    if labels_text:
        inner = labels_text[1:-1]
        for part in re.finditer(r'([^=,]+)="((?:[^"\\]|\\.)*)"', inner):
            labels[part.group(1)] = bytes(part.group(2), "utf-8").decode("unicode_escape")
    try:
        value = float(value_text)
    except ValueError:
        return None
    return name, labels, value


def first_metadata(matrix_root: Path, rows: list[dict[str, str]]) -> dict[str, Any]:
    if not rows:
        return {}
    out_dir = Path(rows[0]["out_dir"])
    if not out_dir.is_absolute():
        out_dir = matrix_root / out_dir
    return read_json(out_dir / "metadata.json")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--matrix-root", required=True, help="bench-matrix result root containing runs.tsv")
    parser.add_argument("--matrix-id", help="stable matrix id; default is matrix root directory name")
    parser.add_argument("--db", required=True, help="SQLite database path")
    parser.add_argument("--schema", default=str(Path(__file__).with_name("sqlite-benchmark-schema.sql")))
    parser.add_argument("--command-line", default=" ".join(shlex.quote(x) for x in sys.argv))
    args = parser.parse_args()

    matrix_root = Path(args.matrix_root).resolve()
    matrix_id = args.matrix_id or matrix_root.name
    db_path = Path(args.db).resolve()
    db_path.parent.mkdir(parents=True, exist_ok=True)

    rows = load_runs(matrix_root)
    meta = first_metadata(matrix_root, rows)
    scenarios = ",".join(dict.fromkeys(row["scenario"] for row in rows))
    rates = ",".join(dict.fromkeys(row["rate"] for row in rows))
    repeat_count = max((int(row["run"]) for row in rows), default=0)
    duration = str(meta.get("duration", ""))
    warmup = str(meta.get("warmup_duration", ""))
    repo_commit = str(meta.get("git_commit", ""))
    git_dirty = 1 if bool(meta.get("git_dirty", False)) else 0
    created_at = dt.datetime.now(dt.UTC).strftime("%Y-%m-%dT%H:%M:%SZ")

    with sqlite3.connect(db_path) as conn:
        conn.execute("PRAGMA foreign_keys = ON")
        conn.executescript(Path(args.schema).read_text(encoding="utf-8"))
        conn.execute("DELETE FROM benchmark_matrices WHERE matrix_id = ?", (matrix_id,))
        conn.execute(
            """
            INSERT INTO benchmark_matrices (
              matrix_id, created_at_utc, out_root, repo_commit, git_dirty,
              duration, warmup_duration, scenarios, rates, repeat_count,
              command_line, source_summary_json, source_summary_md
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
            (
                matrix_id,
                created_at,
                str(matrix_root),
                repo_commit,
                git_dirty,
                duration,
                warmup,
                scenarios,
                rates,
                repeat_count,
                args.command_line,
                read_text(matrix_root / "matrix-summary.json"),
                read_text(matrix_root / "matrix-summary.md"),
            ),
        )

        for row in rows:
            out_dir = Path(row["out_dir"])
            if not out_dir.is_absolute():
                out_dir = matrix_root / out_dir
            metadata = read_json(out_dir / "metadata.json")
            vegeta = read_json(out_dir / "vegeta.json")
            metrics_delta_text = read_text(out_dir / "metrics-delta.txt")
            latencies = vegeta.get("latencies", {}) if isinstance(vegeta, dict) else {}
            bytes_in = vegeta.get("bytes_in", {}) if isinstance(vegeta, dict) else {}
            bytes_out = vegeta.get("bytes_out", {}) if isinstance(vegeta, dict) else {}
            cur = conn.execute(
                """
                INSERT INTO benchmark_runs (
                  matrix_id, scenario, rate_target, run_number, out_dir,
                  created_at_utc, duration, warmup_duration, requests, throughput,
                  success_ratio, p50_ms, p95_ms, p99_ms, max_ms,
                  bytes_in_total, bytes_out_total, status_codes_json, errors_json,
                  metadata_json, vegeta_json, metrics_delta_text
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    matrix_id,
                    row["scenario"],
                    row["rate"],
                    int(row["run"]),
                    str(out_dir),
                    str(metadata.get("created_at_utc", "")) or None,
                    str(metadata.get("duration", duration)),
                    str(metadata.get("warmup_duration", warmup)),
                    int(vegeta.get("requests", 0)),
                    float(vegeta.get("throughput", 0.0)),
                    float(vegeta.get("success", 0.0)),
                    float(latencies.get("50th", 0)) / 1_000_000,
                    float(latencies.get("95th", 0)) / 1_000_000,
                    float(latencies.get("99th", 0)) / 1_000_000,
                    float(latencies.get("max", 0)) / 1_000_000,
                    int(bytes_in.get("total", 0)),
                    int(bytes_out.get("total", 0)),
                    json.dumps(vegeta.get("status_codes", {}), sort_keys=True),
                    json.dumps(vegeta.get("errors", []), sort_keys=True),
                    json.dumps(metadata, sort_keys=True),
                    json.dumps(vegeta, sort_keys=True),
                    metrics_delta_text,
                ),
            )
            run_id = cur.lastrowid
            for line in metrics_delta_text.splitlines():
                parsed = parse_metric_delta_line(line)
                if not parsed:
                    continue
                name, labels, delta = parsed
                conn.execute(
                    """
                    INSERT INTO benchmark_metric_deltas (
                      run_id, matrix_id, scenario, rate_target, metric_name, labels_json, delta_value
                    ) VALUES (?, ?, ?, ?, ?, ?, ?)
                    """,
                    (run_id, matrix_id, row["scenario"], row["rate"], name, json.dumps(labels, sort_keys=True), delta),
                )
        conn.commit()

    print(f"imported matrix {matrix_id!r} into SQLite database {db_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
