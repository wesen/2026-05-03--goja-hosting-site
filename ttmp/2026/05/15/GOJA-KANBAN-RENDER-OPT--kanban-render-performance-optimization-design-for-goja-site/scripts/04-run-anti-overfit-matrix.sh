#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
MATRIX_ID="${MATRIX_ID:-anti-overfit-v1-${STAMP}}"
TICKET_DIR="ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site"
OUT_ROOT="${OUT_ROOT:-bench/results/${MATRIX_ID}}"
DURATION="${DURATION:-15s}"
WARMUP_DURATION="${WARMUP_DURATION:-3s}"
RATES="${RATES:-25/s,100/s,250/s}"
REPEAT="${REPEAT:-3}"
SCENARIOS="${SCENARIOS:-kanban-fragment-10,kanban-fragment,kanban-fragment-500,render-flat-1000,render-attrs-1000,db-read-100,db-write-batch-10}"
START_PORT="${START_PORT:-22080}"
START_METRICS_PORT="${START_METRICS_PORT:-23090}"
BINARY="${BINARY:-$(mktemp -t goja-site-anti-overfit.XXXXXX)}"
KEEP_BINARY="${KEEP_BINARY:-0}"

cleanup() {
  if [[ "$KEEP_BINARY" != "1" && -n "${BINARY:-}" && -f "$BINARY" ]]; then
    rm -f "$BINARY"
  fi
}
trap cleanup EXIT

mkdir -p "$TICKET_DIR/archive"

echo "building goja-site binary: $BINARY"
go build -o "$BINARY" ./cmd/goja-site

scripts/bench-matrix.sh \
  --scenarios "$SCENARIOS" \
  --rates "$RATES" \
  --duration "$DURATION" \
  --warmup-duration "$WARMUP_DURATION" \
  --repeat "$REPEAT" \
  --start-port "$START_PORT" \
  --start-metrics-port "$START_METRICS_PORT" \
  --out-root "$OUT_ROOT" \
  --binary "$BINARY"

python3 "$TICKET_DIR/scripts/05-render-anti-overfit-report.py" \
  --matrix-id "$MATRIX_ID" \
  --matrix-root "$OUT_ROOT" \
  --out "$TICKET_DIR/reference/03-anti-overfit-benchmark-report.md"

echo "anti-overfit matrix complete"
echo "matrix id: $MATRIX_ID"
echo "raw root: $OUT_ROOT"
echo "report: $TICKET_DIR/reference/03-anti-overfit-benchmark-report.md"
