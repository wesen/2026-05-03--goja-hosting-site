#!/usr/bin/env bash
set -euo pipefail

# Quick stress validation pass (~2-3 minutes plus startup/import/report overhead).
# Purpose: prove the stress matrix, SQLite import, and SQL-backed report work
# before running the longer hour-scale experiment.

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

TICKET_DIR="ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
MATRIX_ID="${BENCH_MATRIX_ID:-stress-quick-${STAMP}}"
OUT_ROOT="${BENCH_OUT_ROOT:-bench/results/${MATRIX_ID}}"
DB_PATH="${BENCH_SQLITE_DB:-${TICKET_DIR}/archive/stress-benchmarks.sqlite}"
REPORT_MD="${BENCH_REPORT_MD:-${TICKET_DIR}/reference/02-quick-stress-sweep-sqlite-report.md}"
SCENARIOS="${BENCH_SCENARIOS:-null,render,db-write,kanban-action}"
RATES="${BENCH_RATES:-50/s,100/s,200/s}"
DURATION="${BENCH_DURATION:-10s}"
WARMUP_DURATION="${BENCH_WARMUP_DURATION:-3s}"
REPEAT="${BENCH_REPEAT:-1}"
START_PORT="${BENCH_START_PORT:-18400}"
START_METRICS_PORT="${BENCH_START_METRICS_PORT:-19400}"
TMP_DIR="$(mktemp -d -t goja-site-stress-quick.XXXXXX)"
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
  --title "Quick Stress Sweep SQLite Report"

echo "Quick stress sweep written to: $OUT_ROOT"
echo "SQLite DB: $DB_PATH"
echo "SQLite matrix id: $MATRIX_ID"
echo "Markdown report: $REPORT_MD"
