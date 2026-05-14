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
VEGETA_BIN="${VEGETA_BIN:-vegeta}"

usage() {
  cat <<'EOF_USAGE'
Usage: scripts/bench-vegeta.sh [options]

Options:
  --scenario NAME       null | render | db-read | db-write | multi (default: null)
  --duration DURATION   Vegeta duration, e.g. 30s (default: 30s)
  --rate RATE           Vegeta rate, e.g. 50/s or 100/1s (default: 50/s)
  --port PORT           goja-site app port (default: 18080)
  --metrics-port PORT   private metrics port (default: 19090)
  --out-dir DIR         result directory (default: bench/results/<timestamp>-<scenario>)
  --binary PATH         existing goja-site binary to use instead of building tmp binary
  --keep-db             keep temporary DB files
  -h, --help            show this help

Requires Vegeta: go install github.com/tsenart/vegeta/v12@latest
EOF_USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --scenario) SCENARIO="$2"; shift 2 ;;
    --duration) DURATION="$2"; shift 2 ;;
    --rate) RATE="$2"; shift 2 ;;
    --port) PORT="$2"; shift 2 ;;
    --metrics-port) METRICS_PORT="$2"; shift 2 ;;
    --out-dir) OUT_DIR="$2"; shift 2 ;;
    --binary) BINARY="$2"; shift 2 ;;
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
    db-read)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/read?n=10
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
    multi)
      cat >"$TARGETS_PATH" <<EOF_TARGETS
GET ${BASE_URL}/
Host: site-a.bench.example.test

GET ${BASE_URL}/
Host: site-b.bench.example.test
EOF_TARGETS
      ;;
    *) echo "unsupported scenario: $SCENARIO" >&2; exit 2 ;;
  esac
}

start_server() {
  case "$SCENARIO" in
    null)
      "$BINARY" serve --db "$DB_PATH" --scripts bench/scripts/null-route --db-policy simple --allow-writes --addr "$APP_ADDR" --metrics-addr "$METRICS_ADDR" >"$LOG_PATH" 2>&1 &
      ;;
    render)
      "$BINARY" serve --db "$DB_PATH" --scripts bench/scripts/render-route --db-policy simple --allow-writes --addr "$APP_ADDR" --metrics-addr "$METRICS_ADDR" >"$LOG_PATH" 2>&1 &
      ;;
    db-read|db-write)
      "$BINARY" serve --db "$DB_PATH" --scripts bench/scripts/db-read-write --db-policy simple --allow-writes --addr "$APP_ADDR" --metrics-addr "$METRICS_ADDR" >"$LOG_PATH" 2>&1 &
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
      "$BINARY" serve-multi --config "$cfg" --metrics-addr "$METRICS_ADDR" >"$LOG_PATH" 2>&1 &
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
    if curl "${curl_args[@]}" >/dev/null 2>&1; then
      return 0
    fi
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
      echo "goja-site exited before becoming ready" >&2
      return 1
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

write_targets
start_server
wait_ready
scrape_metrics "$OUT_DIR/metrics-before.prom"

RESULT_BIN="$OUT_DIR/vegeta.bin"
RESULT_JSON="$OUT_DIR/vegeta.json"
RESULT_TEXT="$OUT_DIR/vegeta.txt"

"$VEGETA_BIN" attack -targets="$TARGETS_PATH" -rate="$RATE" -duration="$DURATION" \
  | tee "$RESULT_BIN" \
  | "$VEGETA_BIN" report | tee "$RESULT_TEXT"
"$VEGETA_BIN" report -type=json "$RESULT_BIN" >"$RESULT_JSON"
scrape_metrics "$OUT_DIR/metrics-after.prom"

cat >"$OUT_DIR/summary.md" <<EOF_SUMMARY
# goja-site Vegeta benchmark

- Scenario: ${SCENARIO}
- Duration: ${DURATION}
- Rate: ${RATE}
- Base URL: ${BASE_URL}
- Metrics URL: ${METRICS_URL}
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
- Metrics before: metrics-before.prom
- Metrics after: metrics-after.prom
- Server log: server.log
- Targets: targets.txt
EOF_SUMMARY
cp "$TARGETS_PATH" "$OUT_DIR/targets.txt"

echo "benchmark complete: $OUT_DIR"
