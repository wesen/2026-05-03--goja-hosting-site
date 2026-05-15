#!/usr/bin/env bash
set -euo pipefail

cat <<'EOF'
This is a planning stub for the first implementation PR after the design guide.

After adding an optimized fixture such as:

  bench/scripts/kanban-board-no-precise/app.js

extend scripts/bench-vegeta.sh with scenarios such as:

  kanban-fragment-no-precise
  kanban-action-no-precise

Then run a before/after matrix like:

  scripts/bench-vegeta.sh --scenario kanban-fragment --duration 30s --warmup-duration 5s --rate 100/s --port 19100 --metrics-port 20100
  scripts/bench-vegeta.sh --scenario kanban-fragment-no-precise --duration 30s --warmup-duration 5s --rate 100/s --port 19101 --metrics-port 20101

For multi-VM comparison, use GOJA-MULTI-VM-STRESS/scripts/01-run-multi-vm-vegeta.sh with baseline and optimized scenario names once supported.
EOF
