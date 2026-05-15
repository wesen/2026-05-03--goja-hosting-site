#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCENARIO="null"
DURATION="30s"
RATE="50/s"
PORT="18080"
METRICS_PORT="19090"
OUT_DIR=""
BINARY=""
KEEP_DB="0"
CAPTURE_PPROF="0"
PPROF_SECONDS="5"
OTEL_ENABLED="0"
OTEL_ENDPOINT="http://127.0.0.1:4318/v1/traces"
OTEL_SAMPLE_RATIO="0.01"
WARMUP_DURATION="0s"
VEGETA_BIN="${VEGETA_BIN:-vegeta}"

usage() {
  cat <<'EOF_USAGE'
Usage: scripts/bench-vegeta.sh [options]

Options:
  --scenario NAME       null | render | render-flat-1000 | render-attrs-1000 | db-read | db-read-100 | db-write | db-write-batch-10 | multi | kanban-page | kanban-fragment | kanban-fragment-10 | kanban-fragment-500 | kanban-action | kanban-mixed (default: null)
  --duration DURATION   Vegeta measurement duration, e.g. 30s (default: 30s)
  --warmup-duration D   Optional warmup before measured run, e.g. 10s (default: 0s)
  --rate RATE           Vegeta rate, e.g. 50/s or 100/1s (default: 50/s)
  --port PORT           goja-site app port (default: 18080)
  --metrics-port PORT   private metrics port (default: 19090)
  --out-dir DIR         result directory (default: bench/results/<timestamp>-<scenario>)
  --binary PATH         existing goja-site binary to use instead of building tmp binary
  --pprof               enable diagnostics pprof and capture CPU/heap/goroutine profiles
  --pprof-seconds N     CPU profile duration in seconds when --pprof is set (default: 5)
  --otel                enable OpenTelemetry tracing for the goja-site process
  --otel-endpoint URL   OTLP HTTP traces endpoint (default: http://127.0.0.1:4318/v1/traces)
  --otel-sample-ratio N trace sample ratio between 0 and 1 (default: 0.01)
  --keep-db             keep temporary DB files
  -h, --help            show this help

Requires Vegeta: go install github.com/tsenart/vegeta/v12@latest
EOF_USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --scenario) SCENARIO="$2"; shift 2 ;;
    --duration) DURATION="$2"; shift 2 ;;
    --warmup-duration) WARMUP_DURATION="$2"; shift 2 ;;
    --rate) RATE="$2"; shift 2 ;;
    --port) PORT="$2"; shift 2 ;;
    --metrics-port) METRICS_PORT="$2"; shift 2 ;;
    --out-dir) OUT_DIR="$2"; shift 2 ;;
    --binary) BINARY="$2"; shift 2 ;;
    --pprof) CAPTURE_PPROF="1"; shift ;;
    --pprof-seconds) PPROF_SECONDS="$2"; shift 2 ;;
    --otel) OTEL_ENABLED="1"; shift ;;
    --otel-endpoint) OTEL_ENDPOINT="$2"; shift 2 ;;
    --otel-sample-ratio) OTEL_SAMPLE_RATIO="$2"; shift 2 ;;
    --keep-db) KEEP_DB="1"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if ! command -v "$VEGETA_BIN" >/dev/null 2>&1; then
  echo "vegeta not found. Install with: go install github.com/tsenart/vegeta/v12@latest" >&2
  exit 127
fi

cd "$ROOT"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
if [[ -z "$OUT_DIR" ]]; then
  OUT_DIR="bench/results/${STAMP}-${SCENARIO}"
fi
mkdir -p "$OUT_DIR"
OUT_DIR="$(cd "$OUT_DIR" && pwd)"
TMP_DIR="$(mktemp -d -t goja-site-bench.XXXXXX)"
APP_ADDR="127.0.0.1:${PORT}"
BASE_URL="http://${APP_ADDR}"
METRICS_ADDR="127.0.0.1:${METRICS_PORT}"
METRICS_URL="http://${METRICS_ADDR}/metrics"
PPROF_ARGS=()
if [[ "$CAPTURE_PPROF" == "1" ]]; then
  PPROF_ARGS=(--pprof)
fi
OTEL_ARGS=()
if [[ "$OTEL_ENABLED" == "1" ]]; then
  OTEL_ARGS=(--otel-enabled --otel-endpoint "$OTEL_ENDPOINT" --otel-sample-ratio "$OTEL_SAMPLE_RATIO")
fi
LOG_PATH="$OUT_DIR/server.log"
TARGETS_PATH="$TMP_DIR/targets.txt"
DB_PATH="$TMP_DIR/app.db"

cleanup() {
  local status=$?
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  if [[ "$KEEP_DB" != "1" ]]; then
    rm -f "$DB_PATH"
    rm -rf "$TMP_DIR"
  else
    echo "kept temp directory: $TMP_DIR" >&2
  fi
  if [[ $status -ne 0 ]]; then
    echo "bench output dir: $OUT_DIR" >&2
    echo "server log: $LOG_PATH" >&2
    if [[ -f "$LOG_PATH" ]]; then
      tail -200 "$LOG_PATH" >&2 || true
    fi
  fi
}
trap cleanup EXIT

if [[ -z "$BINARY" ]]; then
  BINARY="$TMP_DIR/goja-site"
  go build -o "$BINARY" ./cmd/goja-site
fi

write_targets() {
  case "$SCENARIO" in
    null)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/
EOF_TARGETS
      ;;
    render)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/render?n=100
EOF_TARGETS
      ;;
    render-flat-1000)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/flat?n=1000
EOF_TARGETS
      ;;
    render-attrs-1000)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/attrs?n=1000
EOF_TARGETS
      ;;
    db-read)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/read?n=10
EOF_TARGETS
      ;;
    db-read-100)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/read?n=100
EOF_TARGETS
      ;;
    db-write)
      cat >"$TMP_DIR/write-one.json" <<'EOF_JSON'
{"n":1}
EOF_JSON
      cat >"$TARGETS_PATH" <<EOF_TARGETS
POST ${BASE_URL}/write
Content-Type: application/json
@${TMP_DIR}/write-one.json
EOF_TARGETS
      ;;
    db-write-batch-10)
      cat >"$TMP_DIR/write-batch-10.json" <<'EOF_JSON'
{"n":10}
EOF_JSON
      cat >"$TARGETS_PATH" <<EOF_TARGETS
POST ${BASE_URL}/write
Content-Type: application/json
@${TMP_DIR}/write-batch-10.json
EOF_TARGETS
      ;;
    multi)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/
Host: site-a.bench.example.test

GET ${BASE_URL}/
Host: site-b.bench.example.test
EOF_TARGETS
      ;;
    kanban-page)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/
EOF_TARGETS
      ;;
    kanban-fragment|kanban-fragment-10|kanban-fragment-500)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/_kanban/bench/fragment
EOF_TARGETS
      ;;
    kanban-action)
      cat >"$TMP_DIR/kanban-move.json" <<'EOF_JSON'
{"cardId":"1","to":{"columnId":"done","index":0}}
EOF_JSON
      cat >"$TARGETS_PATH" <<EOF_TARGETS
POST ${BASE_URL}/_kanban/bench/action/cardMoved
Content-Type: application/json
Cookie: goja_session=bench-session
@${TMP_DIR}/kanban-move.json
EOF_TARGETS
      ;;
    kanban-mixed)
      cat >"$TMP_DIR/kanban-move.json" <<'EOF_JSON'
{"cardId":"1","to":{"columnId":"done","index":0}}
EOF_JSON
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/_kanban/bench/fragment

POST ${BASE_URL}/_kanban/bench/action/cardMoved
Content-Type: application/json
Cookie: goja_session=bench-session
@${TMP_DIR}/kanban-move.json

GET ${BASE_URL}/
GET ${BASE_URL}/_kanban/bench/fragment
GET ${BASE_URL}/_kanban/bench/fragment
GET ${BASE_URL}/_kanban/bench/fragment
GET ${BASE_URL}/_kanban/bench/fragment
GET ${BASE_URL}/_kanban/bench/fragment
GET ${BASE_URL}/_kanban/bench/fragment
GET ${BASE_URL}/_kanban/bench/fragment

POST ${BASE_URL}/_kanban/bench/action/cardMoved
Content-Type: application/json
Cookie: goja_session=bench-session
@${TMP_DIR}/kanban-move.json
EOF_TARGETS
      ;;
    *) echo "unsupported scenario: $SCENARIO" >&2; exit 2 ;;
  esac
}

start_server() {
  case "$SCENARIO" in
    null)
      "$BINARY" serve --db "$DB_PATH" --scripts bench/scripts/null-route --db-policy simple --allow-writes --addr "$APP_ADDR" --metrics-addr "$METRICS_ADDR" "${PPROF_ARGS[@]}" "${OTEL_ARGS[@]}" >"$LOG_PATH" 2>&1 &
      ;;
    render)
      "$BINARY" serve --db "$DB_PATH" --scripts bench/scripts/render-route --db-policy simple --allow-writes --addr "$APP_ADDR" --metrics-addr "$METRICS_ADDR" "${PPROF_ARGS[@]}" "${OTEL_ARGS[@]}" >"$LOG_PATH" 2>&1 &
      ;;
    render-flat-1000|render-attrs-1000)
      "$BINARY" serve --db "$DB_PATH" --scripts bench/scripts/render-shapes --db-policy simple --allow-writes --addr "$APP_ADDR" --metrics-addr "$METRICS_ADDR" "${PPROF_ARGS[@]}" "${OTEL_ARGS[@]}" >"$LOG_PATH" 2>&1 &
      ;;
    db-read|db-write|db-read-100|db-write-batch-10)
      "$BINARY" serve --db "$DB_PATH" --scripts bench/scripts/db-read-write --db-policy simple --allow-writes --addr "$APP_ADDR" --metrics-addr "$METRICS_ADDR" "${PPROF_ARGS[@]}" "${OTEL_ARGS[@]}" >"$LOG_PATH" 2>&1 &
      ;;
    kanban-page|kanban-fragment|kanban-action|kanban-mixed)
      "$BINARY" serve --db "$DB_PATH" --scripts bench/scripts/kanban-board --db-policy simple --allow-writes --addr "$APP_ADDR" --metrics-addr "$METRICS_ADDR" "${PPROF_ARGS[@]}" "${OTEL_ARGS[@]}" >"$LOG_PATH" 2>&1 &
      ;;
    kanban-fragment-10)
      "$BINARY" serve --db "$DB_PATH" --scripts bench/scripts/kanban-board-10 --db-policy simple --allow-writes --addr "$APP_ADDR" --metrics-addr "$METRICS_ADDR" "${PPROF_ARGS[@]}" "${OTEL_ARGS[@]}" >"$LOG_PATH" 2>&1 &
      ;;
    kanban-fragment-500)
      "$BINARY" serve --db "$DB_PATH" --scripts bench/scripts/kanban-board-500 --db-policy simple --allow-writes --addr "$APP_ADDR" --metrics-addr "$METRICS_ADDR" "${PPROF_ARGS[@]}" "${OTEL_ARGS[@]}" >"$LOG_PATH" 2>&1 &
      ;;
    multi)
      local cfg="$TMP_DIR/sites.yaml"
      cat >"$cfg" <<EOF_CFG
addr: "${APP_ADDR}"
dataDir: "${TMP_DIR}/multi-data"
baseDomain: "bench.example.test"
dev: false
sites:
  - name: site-a
    host: site-a.bench.example.test
    dbPolicy: simple
    allowWrites: true
    scripts:
      - bench/scripts/null-route
  - name: site-b
    host: site-b.bench.example.test
    dbPolicy: simple
    allowWrites: true
    scripts:
      - bench/scripts/render-route
EOF_CFG
      "$BINARY" serve-multi --config "$cfg" --metrics-addr "$METRICS_ADDR" "${PPROF_ARGS[@]}" "${OTEL_ARGS[@]}" >"$LOG_PATH" 2>&1 &
      ;;
  esac
  SERVER_PID=$!
}

wait_ready() {
  local url="${BASE_URL}/"
  local curl_args=(-fsS "$url")
  if [[ "$SCENARIO" == "multi" ]]; then
    curl_args=(-fsS -H 'Host: site-a.bench.example.test' "$url")
  fi
  for _ in $(seq 1 100); do
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
      echo "goja-site exited before becoming ready" >&2
      return 1
    fi
    if curl "${curl_args[@]}" >/dev/null 2>&1; then
      if ! kill -0 "$SERVER_PID" 2>/dev/null; then
        echo "goja-site exited during readiness check" >&2
        return 1
      fi
      return 0
    fi
    sleep 0.1
  done
  echo "server did not become ready: $url" >&2
  return 1
}

scrape_metrics() {
  local dest="$1"
  curl -fsS "$METRICS_URL" >"$dest" || true
}

capture_pprof_after_run() {
  if [[ "$CAPTURE_PPROF" != "1" ]]; then
    return 0
  fi
  curl -fsS "http://${METRICS_ADDR}/debug/pprof/heap" -o "$OUT_DIR/heap.pprof" || true
  curl -fsS "http://${METRICS_ADDR}/debug/pprof/goroutine?debug=1" -o "$OUT_DIR/goroutine.txt" || true
  curl -fsS "http://${METRICS_ADDR}/debug/pprof/allocs" -o "$OUT_DIR/allocs.pprof" || true
}

write_targets
start_server
wait_ready
cp "$TARGETS_PATH" "$OUT_DIR/targets.txt"

if [[ "$WARMUP_DURATION" != "0" && "$WARMUP_DURATION" != "0s" ]]; then
  "$VEGETA_BIN" attack -targets="$TARGETS_PATH" -rate="$RATE" -duration="$WARMUP_DURATION" >/dev/null
fi

scrape_metrics "$OUT_DIR/metrics-before.prom"

RESULT_BIN="$OUT_DIR/vegeta.bin"
RESULT_JSON="$OUT_DIR/vegeta.json"
RESULT_TEXT="$OUT_DIR/vegeta.txt"

if [[ "$CAPTURE_PPROF" == "1" ]]; then
  curl -fsS "http://${METRICS_ADDR}/debug/pprof/profile?seconds=${PPROF_SECONDS}" -o "$OUT_DIR/cpu.pprof" &
  PPROF_PID=$!
fi

"$VEGETA_BIN" attack -targets="$TARGETS_PATH" -rate="$RATE" -duration="$DURATION" \
  | tee "$RESULT_BIN" \
  | "$VEGETA_BIN" report | tee "$RESULT_TEXT"
if [[ -n "${PPROF_PID:-}" ]]; then
  wait "$PPROF_PID" || true
fi
"$VEGETA_BIN" report -type=json "$RESULT_BIN" >"$RESULT_JSON"
scrape_metrics "$OUT_DIR/metrics-after.prom"
capture_pprof_after_run

python3 - "$OUT_DIR" "$SCENARIO" "$DURATION" "$WARMUP_DURATION" "$RATE" "$BASE_URL" "$METRICS_URL" "$CAPTURE_PPROF" "$PPROF_SECONDS" "$OTEL_ENABLED" "$OTEL_ENDPOINT" "$OTEL_SAMPLE_RATIO" "$BINARY" <<'PY'
import json, os, platform, socket, subprocess, sys, time
(out_dir, scenario, duration, warmup, rate, base_url, metrics_url, pprof, pprof_seconds, otel, otel_endpoint, otel_sample_ratio, binary) = sys.argv[1:14]
def run(cmd):
    try:
        return subprocess.check_output(cmd, text=True, stderr=subprocess.STDOUT).strip()
    except Exception as e:
        return str(e)
meta = {
    "created_at_utc": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
    "scenario": scenario,
    "duration": duration,
    "warmup_duration": warmup,
    "rate": rate,
    "base_url": base_url,
    "metrics_url": metrics_url,
    "pprof_capture": pprof == "1",
    "pprof_seconds": pprof_seconds,
    "otel_enabled": otel == "1",
    "otel_endpoint": otel_endpoint,
    "otel_sample_ratio": otel_sample_ratio,
    "binary": binary,
    "git_commit": run(["git", "rev-parse", "HEAD"]),
    "git_dirty": bool(run(["git", "status", "--porcelain"])),
    "go_version": run(["go", "version"]),
    "hostname": socket.gethostname(),
    "platform": platform.platform(),
}
with open(os.path.join(out_dir, "metadata.json"), "w", encoding="utf-8") as f:
    json.dump(meta, f, indent=2, sort_keys=True)
    f.write("\n")
PY

python3 - "$OUT_DIR/metrics-before.prom" "$OUT_DIR/metrics-after.prom" >"$OUT_DIR/metrics-delta.txt" <<'PY' || true
import re, sys
metric_re = re.compile(r'^([a-zA-Z_:][a-zA-Z0-9_:]*)(\{[^}]*\})?\s+([-+0-9.eE]+)$')
def load(path):
    out = {}
    with open(path, encoding='utf-8') as f:
        for line in f:
            if line.startswith('#'):
                continue
            m = metric_re.match(line.strip())
            if not m:
                continue
            name, labels, value = m.groups()
            if not name.startswith('goja_site_'):
                continue
            try:
                out[(name, labels or '')] = float(value)
            except ValueError:
                pass
    return out
before, after = load(sys.argv[1]), load(sys.argv[2])
for key in sorted(after):
    delta = after[key] - before.get(key, 0.0)
    if delta:
        print(f"{key[0]}{key[1]} {delta:g}")
PY

cat >"$OUT_DIR/summary.md" <<EOF_SUMMARY
# goja-site Vegeta benchmark

- Scenario: ${SCENARIO}
- Duration: ${DURATION}
- Warmup duration: ${WARMUP_DURATION}
- Rate: ${RATE}
- Base URL: ${BASE_URL}
- Metrics URL: ${METRICS_URL}
- pprof capture: ${CAPTURE_PPROF}
- pprof seconds: ${PPROF_SECONDS}
- OpenTelemetry enabled: ${OTEL_ENABLED}
- OpenTelemetry endpoint: ${OTEL_ENDPOINT}
- OpenTelemetry sample ratio: ${OTEL_SAMPLE_RATIO}
- Commit: $(git rev-parse HEAD)
- Dirty worktree: $(if [[ -n "$(git status --porcelain)" ]]; then echo true; else echo false; fi)
- Go version: $(go version)
- Vegeta: $($VEGETA_BIN --version 2>&1 | head -1)

## Report

\`\`\`text
$(cat "$RESULT_TEXT")
\`\`\`

## Artifacts

- Raw results: vegeta.bin
- JSON report: vegeta.json
- Text report: vegeta.txt
- Run metadata: metadata.json
- Metrics before: metrics-before.prom
- Metrics after: metrics-after.prom
- Metrics delta: metrics-delta.txt
- Server log: server.log
- Targets: targets.txt
- CPU profile: cpu.pprof (when --pprof is set)
- Heap profile: heap.pprof (when --pprof is set)
- Goroutine profile: goroutine.txt (when --pprof is set)
EOF_SUMMARY

echo "benchmark complete: $OUT_DIR"
