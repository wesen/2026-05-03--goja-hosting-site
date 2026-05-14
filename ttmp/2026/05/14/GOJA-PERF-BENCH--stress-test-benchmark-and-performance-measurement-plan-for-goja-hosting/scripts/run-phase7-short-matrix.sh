#!/usr/bin/env bash
set -euo pipefail

# Repro script for a short local benchmark matrix after smoke validity passes.
# This is intentionally metrics-only by default; add --pprof only to targeted reruns.
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT_ROOT="bench/results/phase7-short-${STAMP}"

scripts/bench-matrix.sh \
  --scenarios null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed \
  --rates 5/s,10/s,25/s \
  --duration 60s \
  --warmup-duration 10s \
  --repeat 3 \
  --start-port 18300 \
  --start-metrics-port 19300 \
  --out-root "$OUT_ROOT"

echo "Phase 7 short matrix written to: $OUT_ROOT"
