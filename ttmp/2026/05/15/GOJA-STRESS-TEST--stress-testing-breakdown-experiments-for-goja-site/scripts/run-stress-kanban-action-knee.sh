#!/usr/bin/env bash
set -euo pipefail

# Targeted knee search for the known first bending scenario from the quick
# stress sweep: kanban-action. This intentionally avoids the broad hour-scale
# sweep and narrows the rate range around 50/s -> 100/s.

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

TICKET_DIR="ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
MATRIX_ID="${BENCH_MATRIX_ID:-stress-kanban-action-knee-${STAMP}}"
OUT_ROOT="${BENCH_OUT_ROOT:-bench/results/${MATRIX_ID}}"
DB_PATH="${BENCH_SQLITE_DB:-${TICKET_DIR}/archive/stress-benchmarks.sqlite}"
REPORT_MD="${BENCH_REPORT_MD:-${TICKET_DIR}/reference/03-kanban-action-knee-sqlite-report.md}"
SCENARIOS="${BENCH_SCENARIOS:-kanban-action}"
RATES="${BENCH_RATES:-60/s,70/s,80/s,90/s,100/s}"
DURATION="${BENCH_DURATION:-30s}"
WARMUP_DURATION="${BENCH_WARMUP_DURATION:-5s}"
REPEAT="${BENCH_REPEAT:-3}"
START_PORT="${BENCH_START_PORT:-18600}"
START_METRICS_PORT="${BENCH_START_METRICS_PORT:-19600}"
TMP_DIR="$(mktemp -d -t goja-site-stress-knee.XXXXXX)"
BINARY="$TMP_DIR/goja-site"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

go build -o "$BINARY" ./cmd/goja-site

scripts/bench-matrix.sh \
  --scenarios "$SCENARIOS" \
  --rates "$RATES" \
  --duration "$DURATION" \
  --warmup-duration "$WARMUP_DURATION" \
  --repeat "$REPEAT" \
  --start-port "$START_PORT" \
  --start-metrics-port "$START_METRICS_PORT" \
  --binary "$BINARY" \
  --out-root "$OUT_ROOT"

"${TICKET_DIR}/scripts/import-benchmark-matrix-to-sqlite.py" \
  --matrix-root "$OUT_ROOT" \
  --matrix-id "$MATRIX_ID" \
  --db "$DB_PATH"

"${TICKET_DIR}/scripts/render-stress-report-from-sqlite.py" \
  --matrix-id "$MATRIX_ID" \
  --db "$DB_PATH" \
  --out "$REPORT_MD" \
  --title "Kanban Action Knee Stress Report"

echo "Kanban action knee sweep written to: $OUT_ROOT"
echo "SQLite DB: $DB_PATH"
echo "SQLite matrix id: $MATRIX_ID"
echo "Markdown report: $REPORT_MD"
