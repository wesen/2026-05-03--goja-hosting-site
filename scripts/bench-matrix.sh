#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCENARIOS="null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed"
RATES="5/s"
DURATION="10s"
WARMUP_DURATION="0s"
REPEAT="1"
START_PORT="18080"
START_METRICS_PORT="19090"
OUT_ROOT=""
BINARY=""
CAPTURE_PPROF="0"
PPROF_SECONDS="5"
OTEL_ENABLED="0"
OTEL_ENDPOINT="http://127.0.0.1:4318/v1/traces"
OTEL_SAMPLE_RATIO="0.01"
VEGETA_BIN="${VEGETA_BIN:-vegeta}"

usage() {
  cat <<'EOF_USAGE'
Usage: scripts/bench-matrix.sh [options]

Runs a repeatable matrix of goja-site Vegeta benchmark scenarios. Each cell
calls scripts/bench-vegeta.sh and writes a per-run directory plus a matrix
summary at the root.

Options:
  --scenarios CSV       Scenario names (default: null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed)
  --rates CSV           Vegeta rates, e.g. 5/s,10/s,25/s (default: 5/s)
  --duration DURATION   Measured duration per run (default: 10s)
  --warmup-duration D   Warmup before each measured run (default: 0s)
  --repeat N            Runs per scenario/rate cell (default: 1)
  --start-port PORT     First app port; increments per run (default: 18080)
  --start-metrics-port  First metrics port; increments per run (default: 19090)
  --out-root DIR        Matrix output root (default: bench/results/matrix-<timestamp>)
  --binary PATH         Existing goja-site binary to reuse
  --pprof               Capture pprof artifacts for each run
  --pprof-seconds N     CPU profile seconds when --pprof is set (default: 5)
  --otel                Enable OpenTelemetry for each server process
  --otel-endpoint URL   OTLP HTTP endpoint (default: http://127.0.0.1:4318/v1/traces)
  --otel-sample-ratio N Trace sample ratio (default: 0.01)
  -h, --help            Show this help
EOF_USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --scenarios) SCENARIOS="$2"; shift 2 ;;
    --rates) RATES="$2"; shift 2 ;;
    --duration) DURATION="$2"; shift 2 ;;
    --warmup-duration) WARMUP_DURATION="$2"; shift 2 ;;
    --repeat) REPEAT="$2"; shift 2 ;;
    --start-port) START_PORT="$2"; shift 2 ;;
    --start-metrics-port) START_METRICS_PORT="$2"; shift 2 ;;
    --out-root) OUT_ROOT="$2"; shift 2 ;;
    --binary) BINARY="$2"; shift 2 ;;
    --pprof) CAPTURE_PPROF="1"; shift ;;
    --pprof-seconds) PPROF_SECONDS="$2"; shift 2 ;;
    --otel) OTEL_ENABLED="1"; shift ;;
    --otel-endpoint) OTEL_ENDPOINT="$2"; shift 2 ;;
    --otel-sample-ratio) OTEL_SAMPLE_RATIO="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if ! [[ "$REPEAT" =~ ^[0-9]+$ ]] || [[ "$REPEAT" -lt 1 ]]; then
  echo "--repeat must be a positive integer" >&2
  exit 2
fi

cd "$ROOT"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
if [[ -z "$OUT_ROOT" ]]; then
  OUT_ROOT="bench/results/matrix-${STAMP}"
fi
mkdir -p "$OUT_ROOT"
OUT_ROOT="$(cd "$OUT_ROOT" && pwd)"

IFS=',' read -r -a SCENARIO_LIST <<<"$SCENARIOS"
IFS=',' read -r -a RATE_LIST <<<"$RATES"

RUN_INDEX=0
RUN_LIST="$OUT_ROOT/runs.tsv"
: >"$RUN_LIST"

sanitize_rate() {
  echo "$1" | tr '/:' '__'
}

for scenario in "${SCENARIO_LIST[@]}"; do
  scenario="$(echo "$scenario" | xargs)"
  [[ -n "$scenario" ]] || continue
  for rate in "${RATE_LIST[@]}"; do
    rate="$(echo "$rate" | xargs)"
    [[ -n "$rate" ]] || continue
    for run in $(seq 1 "$REPEAT"); do
      port=$((START_PORT + RUN_INDEX))
      metrics_port=$((START_METRICS_PORT + RUN_INDEX))
      safe_rate="$(sanitize_rate "$rate")"
      out_dir="$OUT_ROOT/${scenario}/rate-${safe_rate}/run-${run}"
      mkdir -p "$out_dir"
      args=(
        scripts/bench-vegeta.sh
        --scenario "$scenario"
        --duration "$DURATION"
        --warmup-duration "$WARMUP_DURATION"
        --rate "$rate"
        --port "$port"
        --metrics-port "$metrics_port"
        --out-dir "$out_dir"
      )
      if [[ -n "$BINARY" ]]; then
        args+=(--binary "$BINARY")
      fi
      if [[ "$CAPTURE_PPROF" == "1" ]]; then
        args+=(--pprof --pprof-seconds "$PPROF_SECONDS")
      fi
      if [[ "$OTEL_ENABLED" == "1" ]]; then
        args+=(--otel --otel-endpoint "$OTEL_ENDPOINT" --otel-sample-ratio "$OTEL_SAMPLE_RATIO")
      fi
      echo "==> scenario=${scenario} rate=${rate} run=${run} port=${port} metrics_port=${metrics_port}"
      VEGETA_BIN="$VEGETA_BIN" "${args[@]}"
      printf '%s\t%s\t%s\t%s\n' "$scenario" "$rate" "$run" "$out_dir" >>"$RUN_LIST"
      RUN_INDEX=$((RUN_INDEX + 1))
    done
  done
done

python3 - "$OUT_ROOT" "$RUN_LIST" <<'PY'
import csv, json, os, statistics, sys
out_root, run_list = sys.argv[1:3]
rows = []
with open(run_list, encoding="utf-8") as f:
    for scenario, rate, run, out_dir in csv.reader(f, delimiter="\t"):
        report_path = os.path.join(out_dir, "vegeta.json")
        with open(report_path, encoding="utf-8") as rf:
            report = json.load(rf)
        lat = report.get("latencies", {})
        rows.append({
            "scenario": scenario,
            "rate_target": rate,
            "run": int(run),
            "requests": report.get("requests", 0),
            "throughput": report.get("throughput", 0.0),
            "success": report.get("success", 0.0),
            "p50_ms": lat.get("50th", 0) / 1_000_000,
            "p95_ms": lat.get("95th", 0) / 1_000_000,
            "p99_ms": lat.get("99th", 0) / 1_000_000,
            "max_ms": lat.get("max", 0) / 1_000_000,
            "status_codes": report.get("status_codes", {}),
            "errors": report.get("errors", []),
            "out_dir": os.path.relpath(out_dir, out_root),
        })

with open(os.path.join(out_root, "matrix-summary.json"), "w", encoding="utf-8") as f:
    json.dump(rows, f, indent=2, sort_keys=True)
    f.write("\n")

md = ["# goja-site benchmark matrix", "", f"Output root: `{out_root}`", "", "## Runs", "", "| scenario | rate | run | requests | success | throughput | p50 ms | p95 ms | p99 ms | max ms | status | errors | dir |", "|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|---|---|"]
for r in rows:
    md.append("| {scenario} | {rate_target} | {run} | {requests} | {success:.4f} | {throughput:.2f} | {p50_ms:.2f} | {p95_ms:.2f} | {p99_ms:.2f} | {max_ms:.2f} | `{status_json}` | `{errors_json}` | `{out_dir}` |".format(
        status_json=json.dumps(r["status_codes"], sort_keys=True),
        errors_json=json.dumps(r["errors"]),
        **r,
    ))

if rows:
    md += ["", "## Aggregates", "", "| scenario | rate | runs | success avg | p95 avg ms | p95 max ms | p99 avg ms |", "|---|---:|---:|---:|---:|---:|---:|"]
    groups = {}
    for r in rows:
        groups.setdefault((r["scenario"], r["rate_target"]), []).append(r)
    for (scenario, rate), values in sorted(groups.items()):
        md.append("| {} | {} | {} | {:.4f} | {:.2f} | {:.2f} | {:.2f} |".format(
            scenario,
            rate,
            len(values),
            statistics.mean(v["success"] for v in values),
            statistics.mean(v["p95_ms"] for v in values),
            max(v["p95_ms"] for v in values),
            statistics.mean(v["p99_ms"] for v in values),
        ))

with open(os.path.join(out_root, "matrix-summary.md"), "w", encoding="utf-8") as f:
    f.write("\n".join(md) + "\n")
PY

echo "matrix complete: $OUT_ROOT"
echo "summary: $OUT_ROOT/matrix-summary.md"
