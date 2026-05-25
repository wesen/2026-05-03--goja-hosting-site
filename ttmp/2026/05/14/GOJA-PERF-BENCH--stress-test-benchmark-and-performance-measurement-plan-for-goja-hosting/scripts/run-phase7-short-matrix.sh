#!/usr/bin/env bash
set -euo pipefail

# Repro script for a short local benchmark matrix after smoke validity passes.
# It runs metrics-only by default, imports results into SQLite, and renders a
# Markdown report with the SQL queries used to create it.

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

TICKET_DIR="ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
MATRIX_ID="${BENCH_MATRIX_ID:-phase7-short-${STAMP}}"
OUT_ROOT="${BENCH_OUT_ROOT:-bench/results/${MATRIX_ID}}"
DB_PATH="${BENCH_SQLITE_DB:-${TICKET_DIR}/archive/phase7-benchmarks.sqlite}"
REPORT_MD="${BENCH_REPORT_MD:-${TICKET_DIR}/reference/04-phase7-short-sqlite-benchmark-report.md}"
SCENARIOS="${BENCH_SCENARIOS:-null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed}"
RATES="${BENCH_RATES:-5/s,10/s,25/s}"
DURATION="${BENCH_DURATION:-60s}"
WARMUP_DURATION="${BENCH_WARMUP_DURATION:-10s}"
REPEAT="${BENCH_REPEAT:-3}"
START_PORT="${BENCH_START_PORT:-18300}"
START_METRICS_PORT="${BENCH_START_METRICS_PORT:-19300}"
TMP_DIR="$(mktemp -d -t goja-site-short-matrix.XXXXXX)"
BINARY="$TMP_DIR/goja-site"
cleanup() {
  rm -rf "$TMP_DIR"
}
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

"${TICKET_DIR}/scripts/render-benchmark-report-from-sqlite.py" \
  --matrix-id "$MATRIX_ID" \
  --db "$DB_PATH" \
  --out "$REPORT_MD"

echo "Phase 7 short matrix written to: $OUT_ROOT"
echo "SQLite DB: $DB_PATH"
echo "SQLite matrix id: $MATRIX_ID"
echo "Markdown report: $REPORT_MD"
