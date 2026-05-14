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
scripts/bench-vegeta.sh --scenario multi --duration 10s --rate 100/s
```

The wrapper starts `goja-site`, enables the private metrics listener, captures metrics before and after the run, stores raw Vegeta binary output, and writes a small Markdown summary in `bench/results/`.

## Other useful tools

- `fortio`: Go-based HTTP/gRPC load testing with useful histograms and server modes.
- `hey`: Go-based quick fixed-concurrency HTTP checks.
- `bombardier`: Go-based saturation testing.
- `k6`: Go runtime with JavaScript test scripts; excellent for rich user workflows but not Go-authored scenarios.

## Fixture scripts

- `bench/scripts/null-route`: constant `ok` route for host/router overhead.
- `bench/scripts/render-route`: UI DSL render route with configurable node count.
- `bench/scripts/db-read-write`: SQLite read/write routes for early DB load scenarios.

Generated outputs under `bench/results/` are ignored by git except `.gitignore`.
