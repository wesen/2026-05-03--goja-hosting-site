#!/usr/bin/env bash
set -euo pipefail

# Repro script for the Phase 7 benchmark validity pass.
# It runs the smoke matrix, imports results into SQLite, and renders a Markdown
# report with the SQL queries used to create it.

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

TICKET_DIR="ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
MATRIX_ID="phase7-smoke-${STAMP}"
OUT_ROOT="bench/results/${MATRIX_ID}"
DB_PATH="${TICKET_DIR}/archive/phase7-benchmarks.sqlite"
REPORT_MD="${TICKET_DIR}/reference/03-phase7-smoke-sqlite-benchmark-report.md"

scripts/bench-matrix.sh \
  --scenarios null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed \
  --rates 2/s \
  --duration 2s \
  --warmup-duration 0s \
  --repeat 1 \
  --start-port 18200 \
  --start-metrics-port 19200 \
  --out-root "$OUT_ROOT"

"${TICKET_DIR}/scripts/import-benchmark-matrix-to-sqlite.py" \
  --matrix-root "$OUT_ROOT" \
  --matrix-id "$MATRIX_ID" \
  --db "$DB_PATH"

"${TICKET_DIR}/scripts/render-benchmark-report-from-sqlite.py" \
  --matrix-id "$MATRIX_ID" \
  --db "$DB_PATH" \
  --out "$REPORT_MD"

echo "Phase 7 smoke matrix written to: $OUT_ROOT"
echo "SQLite DB: $DB_PATH"
echo "SQLite matrix id: $MATRIX_ID"
echo "Markdown report: $REPORT_MD"
