#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

TICKET_DIR="ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
MATRIX_ID="${MULTI_VM_MATRIX_ID:-multi-vm-saturation-${STAMP}}"
OUT_ROOT="${MULTI_VM_OUT_ROOT:-bench/results/${MATRIX_ID}}"
REPORT_MD="${MULTI_VM_REPORT_MD:-${TICKET_DIR}/reference/03-multi-vm-saturation-sweep-report.md}"
DURATION="${MULTI_VM_DURATION:-10s}"
WARMUP_DURATION="${MULTI_VM_WARMUP_DURATION:-3s}"
START_PORT="${MULTI_VM_START_PORT:-18900}"
START_METRICS_PORT="${MULTI_VM_START_METRICS_PORT:-19900}"
TMP_DIR="$(mktemp -d -t goja-site-multi-vm-saturation.XXXXXX)"
BINARY="$TMP_DIR/goja-site"

cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

mkdir -p "$OUT_ROOT"
go build -o "$BINARY" ./cmd/goja-site

run_index=0
run_cell() {
  local scenario="$1"
  local vm_count="$2"
  local distribution="$3"
  local rate="$4"
  local safe_rate
  safe_rate="$(echo "$rate" | tr '/:' '__')"
  local out_dir="$OUT_ROOT/${scenario}/${distribution}/${vm_count}vm/rate-${safe_rate}"
  local port=$((START_PORT + run_index))
  local metrics_port=$((START_METRICS_PORT + run_index))
  echo "==> scenario=${scenario} vm_count=${vm_count} distribution=${distribution} rate=${rate} port=${port} metrics_port=${metrics_port}"
  "${TICKET_DIR}/scripts/01-run-multi-vm-vegeta.sh" \
    --scenario "$scenario" \
    --vm-count "$vm_count" \
    --distribution "$distribution" \
    --rate "$rate" \
    --duration "$DURATION" \
    --warmup-duration "$WARMUP_DURATION" \
    --port "$port" \
    --metrics-port "$metrics_port" \
    --binary "$BINARY" \
    --out-dir "$out_dir"
  run_index=$((run_index + 1))
}

# Cheap path saturation: total offered rate across all configured hot VMs.
for vm_count in 1 2 4 8; do
  for rate in 400/s 800/s 1200/s 2000/s; do
    run_cell null "$vm_count" even-hot "$rate"
  done
done

# Real rendering path saturation. This intentionally avoids kanban-action for now
# because single-VM action refresh already bends around 80/s.
for vm_count in 1 2 4 8; do
  for rate in 100/s 200/s 400/s; do
    run_cell kanban-fragment "$vm_count" even-hot "$rate"
  done
done

"${TICKET_DIR}/scripts/03-render-multi-vm-summary.py" \
  --root "$OUT_ROOT" \
  --out "$REPORT_MD" \
  --title "Multi-VM serve-multi saturation sweep report: ${MATRIX_ID}"

echo "multi-vm saturation sweep complete: $OUT_ROOT"
echo "report: $REPORT_MD"
