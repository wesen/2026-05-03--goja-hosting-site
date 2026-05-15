#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

TICKET_DIR="ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
RUN_ID="${RUN_ID:-kanban-simplified-${STAMP}}"
OUT_ROOT="${OUT_ROOT:-bench/results/${RUN_ID}}"
TMP_DIR="$(mktemp -d -t goja-kanban-simplified.XXXXXX)"
BINARY="$TMP_DIR/goja-site"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

mkdir -p "$OUT_ROOT"
go build -o "$BINARY" ./cmd/goja-site

scripts/bench-vegeta.sh \
  --scenario kanban-fragment \
  --duration 10s \
  --warmup-duration 3s \
  --rate 200/s \
  --port 19110 \
  --metrics-port 20110 \
  --binary "$BINARY" \
  --out-dir "$OUT_ROOT/single-kanban-fragment-200s"

scripts/bench-vegeta.sh \
  --scenario kanban-action \
  --duration 10s \
  --warmup-duration 3s \
  --rate 100/s \
  --port 19111 \
  --metrics-port 20111 \
  --binary "$BINARY" \
  --out-dir "$OUT_ROOT/single-kanban-action-100s"

ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/scripts/01-run-multi-vm-vegeta.sh \
  --scenario kanban-fragment \
  --vm-count 4 \
  --distribution even-hot \
  --rate 400/s \
  --duration 10s \
  --warmup-duration 3s \
  --port 19112 \
  --metrics-port 20112 \
  --binary "$BINARY" \
  --out-dir "$OUT_ROOT/multi-4vm-kanban-fragment-400s"

echo "$OUT_ROOT"
