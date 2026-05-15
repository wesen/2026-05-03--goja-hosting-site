#!/usr/bin/env python3
"""Render an anti-overfit benchmark Markdown report from bench-matrix output."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import statistics
from pathlib import Path
from typing import Any


def rate_key(rate: str) -> float:
    match = re.match(r"([0-9.]+)\s*/", rate)
    return float(match.group(1)) if match else 0.0


def fmt(value: Any, digits: int = 2) -> str:
    if value is None:
        return ""
    if isinstance(value, float):
        return f"{value:.{digits}f}"
    return str(value).replace("|", "\\|")


def table(headers: list[str], rows: list[list[Any]]) -> str:
    out = ["| " + " | ".join(headers) + " |", "|" + "|".join("---" for _ in headers) + "|"]
    for row in rows:
        out.append("| " + " | ".join(fmt(v) for v in row) + " |")
    return "\n".join(out)


def target_rps(rate: str) -> float:
    return rate_key(rate)


def scenario_family(scenario: str) -> str:
    if scenario.startswith("kanban"):
        return "Kanban size scaling"
    if scenario.startswith("render"):
        return "UI shape rendering"
    if scenario.startswith("db"):
        return "Database-heavy"
    return "Other"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--matrix-id", required=True)
    parser.add_argument("--matrix-root", required=True)
    parser.add_argument("--out", required=True)
    args = parser.parse_args()

    matrix_root = Path(args.matrix_root)
    rows = json.loads((matrix_root / "matrix-summary.json").read_text(encoding="utf-8"))
    for row in rows:
        row["target_rps"] = target_rps(row["rate_target"])
        row["throughput_ratio"] = (row["throughput"] / row["target_rps"]) if row["target_rps"] else 0.0
        row["family"] = scenario_family(row["scenario"])

    groups: dict[tuple[str, str], list[dict[str, Any]]] = {}
    for row in rows:
        groups.setdefault((row["scenario"], row["rate_target"]), []).append(row)

    aggregates: list[dict[str, Any]] = []
    for (scenario, rate), values in sorted(groups.items(), key=lambda item: (item[0][0], rate_key(item[0][1]))):
        aggregates.append({
            "scenario": scenario,
            "family": scenario_family(scenario),
            "rate": rate,
            "runs": len(values),
            "requests": sum(v["requests"] for v in values),
            "success_avg": statistics.mean(v["success"] for v in values),
            "throughput_avg": statistics.mean(v["throughput"] for v in values),
            "throughput_ratio_avg": statistics.mean(v["throughput_ratio"] for v in values),
            "p50_avg": statistics.mean(v["p50_ms"] for v in values),
            "p95_avg": statistics.mean(v["p95_ms"] for v in values),
            "p95_max": max(v["p95_ms"] for v in values),
            "p99_avg": statistics.mean(v["p99_ms"] for v in values),
            "max_ms": max(v["max_ms"] for v in values),
            "errors": sum(len(v.get("errors") or []) for v in values),
        })

    slowest = sorted(rows, key=lambda r: r["p95_ms"], reverse=True)[:15]
    weak = [a for a in aggregates if a["success_avg"] < 1 or a["throughput_ratio_avg"] < 0.98 or a["p95_avg"] > 250]
    weak = sorted(weak, key=lambda a: (a["throughput_ratio_avg"], -a["p95_avg"]))

    best_by_scenario: dict[str, dict[str, Any]] = {}
    for a in aggregates:
        cur = best_by_scenario.get(a["scenario"])
        if cur is None or rate_key(a["rate"]) > rate_key(cur["rate"]):
            best_by_scenario[a["scenario"]] = a

    family_rows = []
    for family in ["Kanban size scaling", "UI shape rendering", "Database-heavy"]:
        vals = [a for a in aggregates if a["family"] == family]
        if vals:
            family_rows.append([
                family,
                len(vals),
                min(v["p95_avg"] for v in vals),
                statistics.mean(v["p95_avg"] for v in vals),
                max(v["p95_avg"] for v in vals),
                min(v["throughput_ratio_avg"] for v in vals),
            ])

    out = [
        "---",
        f"Title: Anti-overfit benchmark report - {args.matrix_id}",
        "Ticket: GOJA-KANBAN-RENDER-OPT",
        "Status: active",
        "Topics:",
        "    - benchmarking",
        "    - performance",
        "    - kanban",
        "    - optimization",
        "DocType: reference",
        "Intent: historical",
        'Summary: "Results from Anti-overfit matrix v1 covering Kanban sizes, renderer shapes, and DB workloads."',
        f"LastUpdated: {dt.datetime.now(dt.UTC).strftime('%Y-%m-%dT%H:%M:%SZ')}",
        "---",
        "",
        f"# Anti-overfit benchmark report: {args.matrix_id}",
        "",
        f"Raw matrix root: `{matrix_root}`",
        "",
        "This report broadens the post-simplification evidence beyond the original 120-card Kanban fixture. It checks three workload classes: Kanban size scaling, generic UI DSL render shapes, and database-heavy request paths.",
        "",
        "## Matrix shape",
        "",
        table(["field", "value"], [
            ["scenarios", ", ".join(sorted({r["scenario"] for r in rows}))],
            ["rates", ", ".join(sorted({r["rate_target"] for r in rows}, key=rate_key))],
            ["runs", len(rows)],
            ["repeat count per cell", max(len(v) for v in groups.values()) if groups else 0],
        ]),
        "",
        "## Executive summary",
        "",
    ]

    if weak:
        out.append("The matrix found cells that deserve follow-up because they missed 98% throughput ratio, had non-100% success, or exceeded 250 ms average p95:")
        out.append("")
        out.append(table(["scenario", "rate", "success avg", "throughput ratio avg", "p95 avg ms", "p95 max ms"], [
            [a["scenario"], a["rate"], a["success_avg"], a["throughput_ratio_avg"], a["p95_avg"], a["p95_max"]]
            for a in weak[:20]
        ]))
    else:
        out.append("All measured cells held 100% success, kept average throughput ratio at or above 98%, and stayed under 250 ms average p95. That means the post-simplification render path is not obviously overfit to the original Kanban fixture at this matrix size.")
    out.extend([
        "",
        "## Workload-family comparison",
        "",
        table(["family", "cells", "p95 min ms", "p95 avg ms", "p95 max ms", "min throughput ratio"], family_rows),
        "",
        "## Scenario/rate aggregates",
        "",
        table(["family", "scenario", "rate", "runs", "requests", "success avg", "throughput avg", "throughput ratio avg", "p50 avg ms", "p95 avg ms", "p95 max ms", "p99 avg ms", "max ms", "errors"], [
            [a["family"], a["scenario"], a["rate"], a["runs"], a["requests"], a["success_avg"], a["throughput_avg"], a["throughput_ratio_avg"], a["p50_avg"], a["p95_avg"], a["p95_max"], a["p99_avg"], a["max_ms"], a["errors"]]
            for a in aggregates
        ]),
        "",
        "## Highest-rate cell per scenario",
        "",
        table(["scenario", "rate", "success avg", "throughput avg", "throughput ratio avg", "p95 avg ms", "p99 avg ms"], [
            [a["scenario"], a["rate"], a["success_avg"], a["throughput_avg"], a["throughput_ratio_avg"], a["p95_avg"], a["p99_avg"]]
            for _, a in sorted(best_by_scenario.items())
        ]),
        "",
        "## Slowest individual runs by p95",
        "",
        table(["scenario", "rate", "run", "success", "throughput", "throughput ratio", "p50 ms", "p95 ms", "p99 ms", "max ms", "dir"], [
            [r["scenario"], r["rate_target"], r["run"], r["success"], r["throughput"], r["throughput_ratio"], r["p50_ms"], r["p95_ms"], r["p99_ms"], r["max_ms"], r["out_dir"]]
            for r in slowest
        ]),
        "",
        "## Interpretation",
        "",
        "Use this matrix as a regression and discovery tool, not a final capacity claim. The tested duration is intentionally short enough to run during development. If a workload family shows a knee here, follow up with a targeted longer run and pprof. If all cells pass, the next useful expansion is mixed multi-site traffic or a browser-side timing matrix.",
        "",
        "## Recommended follow-up",
        "",
        "1. Add threshold checking around this matrix so it can fail on p95, success, response-byte, or throughput-ratio regressions.",
        "2. Add a 1000-card Kanban fragment once 500-card behavior is understood.",
        "3. Add mixed `serve-multi` traffic with null, render, DB, and Kanban sites in the same process.",
        "4. Capture pprof only for the first cell that exceeds the threshold, rather than profiling every run.",
    ])

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text("\n".join(out) + "\n", encoding="utf-8")
    print(f"wrote report: {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
