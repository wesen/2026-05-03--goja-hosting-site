#!/usr/bin/env bash
set -euo pipefail

# Repro script for the first Phase 7 benchmark validity pass.
# Run from anywhere inside the goja-site checkout.
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT_ROOT="bench/results/phase7-smoke-${STAMP}"

scripts/bench-matrix.sh \
  --scenarios null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed \
  --rates 2/s \
  --duration 2s \
  --warmup-duration 0s \
  --repeat 1 \
  --start-port 18200 \
  --start-metrics-port 19200 \
  --out-root "$OUT_ROOT"

echo "Phase 7 smoke matrix written to: $OUT_ROOT"
