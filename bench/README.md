# goja-site benchmark load tools

This directory contains small load-test fixtures and target files for the `GOJA-PERF-BENCH` ticket.

## Recommended tool: Vegeta

Vegeta is the first load-generation tool for this repo because it is written in Go, has a CLI, has a Go library for future custom scenario runners, supports constant-rate load, and can set Host headers for `serve-multi` tests.

Install:

```bash
go install github.com/tsenart/vegeta/v12@latest
```

Run a smoke benchmark through the wrapper:

```bash
scripts/bench-vegeta.sh --scenario null --duration 10s --rate 50/s
scripts/bench-vegeta.sh --scenario render --duration 10s --rate 50/s
scripts/bench-vegeta.sh --scenario db-read --duration 10s --rate 25/s
scripts/bench-vegeta.sh --scenario kanban-fragment --duration 10s --rate 10/s
scripts/bench-vegeta.sh --scenario kanban-action --duration 10s --rate 5/s
scripts/bench-vegeta.sh --scenario multi --duration 10s --rate 100/s
```

The wrapper starts `goja-site`, enables the private metrics listener, optionally performs a warmup, captures metrics before and after the measured run, stores raw Vegeta binary output, writes `metadata.json`, computes a small `metrics-delta.txt`, and writes a Markdown summary in `bench/results/`.

## Other useful tools

- `fortio`: Go-based HTTP/gRPC load testing with useful histograms and server modes.
- `hey`: Go-based quick fixed-concurrency HTTP checks.
- `bombardier`: Go-based saturation testing.
- `k6`: Go runtime with JavaScript test scripts; excellent for rich user workflows but not Go-authored scenarios.

## Matrix runner

Use the matrix runner for comparable scenario/rate sweeps:

```bash
scripts/bench-matrix.sh \
  --scenarios null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed \
  --rates 5/s,10/s,25/s \
  --duration 60s \
  --warmup-duration 10s \
  --repeat 3
```

Each matrix writes per-run artifacts plus:

- `runs.tsv`
- `matrix-summary.json`
- `matrix-summary.md`

The canonical scenario list and suggested smoke/short/saturation profiles live in `bench/scenarios.yaml`.

## Fixture scripts

- `bench/scripts/null-route`: constant `ok` route for host/router overhead.
- `bench/scripts/render-route`: UI DSL render route with configurable node count.
- `bench/scripts/db-read-write`: SQLite read/write routes for early DB load scenarios.
- `bench/scripts/kanban-board`: mounted 120-card Kanban board with valid fragment and `cardMoved` action endpoints.

Generated outputs under `bench/results/` are ignored by git except `.gitignore`.
