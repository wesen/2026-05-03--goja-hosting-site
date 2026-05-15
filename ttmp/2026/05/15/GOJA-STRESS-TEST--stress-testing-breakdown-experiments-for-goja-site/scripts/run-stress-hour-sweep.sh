#!/usr/bin/env bash
set -euo pipefail

# Hour-scale stress sweep. Do not run this until run-stress-quick-sweep.sh has
# completed successfully on the same machine.
#
# Default shape: 7 scenarios x 6 rates x 1 repeat, 10s warmup + 60s measured.
# This is roughly 50 minutes of measured/warmup time plus server startup/import
# overhead, i.e. an hour-scale experiment.

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

TICKET_DIR="ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
MATRIX_ID="${BENCH_MATRIX_ID:-stress-hour-${STAMP}}"
OUT_ROOT="${BENCH_OUT_ROOT:-bench/results/${MATRIX_ID}}"
DB_PATH="${BENCH_SQLITE_DB:-${TICKET_DIR}/archive/stress-benchmarks.sqlite}"
REPORT_MD="${BENCH_REPORT_MD:-${TICKET_DIR}/reference/03-hour-stress-sweep-sqlite-report.md}"
SCENARIOS="${BENCH_SCENARIOS:-null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed}"
RATES="${BENCH_RATES:-25/s,50/s,100/s,200/s,400/s,800/s}"
DURATION="${BENCH_DURATION:-60s}"
WARMUP_DURATION="${BENCH_WARMUP_DURATION:-10s}"
REPEAT="${BENCH_REPEAT:-1}"
START_PORT="${BENCH_START_PORT:-18500}"
START_METRICS_PORT="${BENCH_START_METRICS_PORT:-19500}"
TMP_DIR="$(mktemp -d -t goja-site-stress-hour.XXXXXX)"
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
  --title "Hour Stress Sweep SQLite Report"

echo "Hour stress sweep written to: $OUT_ROOT"
echo "SQLite DB: $DB_PATH"
echo "SQLite matrix id: $MATRIX_ID"
echo "Markdown report: $REPORT_MD"
