#!/usr/bin/env python3
"""Render a Markdown summary for multi-VM stress run directories."""
from __future__ import annotations

import argparse
import json
from pathlib import Path


def read_json(path: Path):
    if not path.exists():
        return {}
    with path.open(encoding="utf-8") as f:
        return json.load(f)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", required=True, help="Matrix root containing per-run directories")
    parser.add_argument("--out", required=True, help="Markdown output path")
    parser.add_argument("--title", default="Multi-VM serve-multi stress summary")
    args = parser.parse_args()

    root = Path(args.root).resolve()
    rows = []
    for metadata_path in sorted(root.glob("**/metadata.json")):
        run_dir = metadata_path.parent
        meta = read_json(metadata_path)
        vegeta = read_json(run_dir / "vegeta.json")
        lat = vegeta.get("latencies", {}) if isinstance(vegeta, dict) else {}
        rows.append({
            "dir": str(run_dir.relative_to(root)),
            "vm_count": meta.get("vm_count"),
            "scenario": meta.get("scenario"),
            "distribution": meta.get("distribution"),
            "rate": meta.get("rate"),
            "duration": meta.get("duration"),
            "warmup": meta.get("warmup_duration"),
            "requests": vegeta.get("requests", 0),
            "throughput": vegeta.get("throughput", 0.0),
            "success": vegeta.get("success", 0.0),
            "p50_ms": lat.get("50th", 0) / 1_000_000,
            "p95_ms": lat.get("95th", 0) / 1_000_000,
            "p99_ms": lat.get("99th", 0) / 1_000_000,
            "max_ms": lat.get("max", 0) / 1_000_000,
            "status": json.dumps(vegeta.get("status_codes", {}), sort_keys=True),
            "errors": json.dumps(vegeta.get("errors", []), sort_keys=True),
        })

    rows.sort(key=lambda r: (str(r["scenario"]), str(r["distribution"]), int(r["vm_count"] or 0), str(r["rate"]), r["dir"]))

    out = Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    md = [
        "---",
        f"Title: {args.title}",
        "Ticket: GOJA-MULTI-VM-STRESS",
        "Status: active",
        "Topics:",
        "    - benchmarking",
        "    - stress-testing",
        "    - goja",
        "    - observability",
        "DocType: reference",
        "Intent: historical",
        "Summary: \"Markdown summary generated from multi-VM serve-multi stress run artifacts.\"",
        "---",
        "",
        f"# {args.title}",
        "",
        f"Result root: `{root}`",
        "",
        "## Runs",
        "",
        "| scenario | vm_count | distribution | rate | requests | throughput | success | p50 ms | p95 ms | p99 ms | max ms | status | errors | dir |",
        "|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|---|---|---|",
    ]
    for r in rows:
        md.append(
            "| {scenario} | {vm_count} | {distribution} | {rate} | {requests} | {throughput:.2f} | {success:.4f} | {p50_ms:.2f} | {p95_ms:.2f} | {p99_ms:.2f} | {max_ms:.2f} | `{status}` | `{errors}` | `{dir}` |".format(**r)
        )

    md += [
        "",
        "## Interpretation notes",
        "",
        "- `vm_count` is the number of configured `serve-multi` sites, which means the number of Goja runtimes/VMs loaded by the process.",
        "- `even-hot` distributes Vegeta targets across all hosts in the target file.",
        "- `one-hot` sends traffic only to the first site while the remaining VMs stay idle.",
        "- `skewed` repeats site 1 nine times, then includes each other site once.",
        "- Throughput is total process throughput across the generated target set, not per-site throughput.",
    ]
    out.write_text("\n".join(md) + "\n", encoding="utf-8")
    print(f"wrote {out}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
