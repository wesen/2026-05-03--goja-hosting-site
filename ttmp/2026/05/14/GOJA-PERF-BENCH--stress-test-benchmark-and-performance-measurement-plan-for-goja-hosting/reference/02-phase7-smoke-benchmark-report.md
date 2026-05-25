---
Title: Phase 7 Smoke Benchmark Report
Ticket: GOJA-PERF-BENCH
Status: active
Topics:
    - benchmarking
    - vegeta
    - kanban
    - prometheus
DocType: reference
Intent: historical
Owners: []
RelatedFiles:
    - Path: bench/scenarios.yaml
      Note: canonical benchmark scenario matrix
    - Path: bench/scripts/kanban-board/app.js
      Note: valid mounted Kanban benchmark fixture
    - Path: scripts/bench-matrix.sh
      Note: matrix runner used for Phase 7 smoke validation
    - Path: scripts/bench-vegeta.sh
      Note: single-scenario benchmark runner enhanced for warmup
ExternalSources: []
Summary: First Phase 7 benchmark runner smoke report validating null, render, DB, and Kanban benchmark scenarios.
LastUpdated: 2026-05-14T19:12:00-04:00
WhatFor: Use this as evidence that the new benchmark matrix runner and Kanban target/actions are valid before running longer benchmark matrices.
WhenToUse: Before comparing longer benchmark results or changing benchmark fixtures.
---





# Phase 7 Smoke Benchmark Report

## Purpose

This report records the first Phase 7 benchmark validity pass. The goal was not to establish production capacity. The goal was to prove that the benchmark runner, targets, fixture sites, and Kanban action payloads are valid before investing time in longer benchmark runs.

## Implementation Added

The Phase 7 runner work added:

- `bench/scenarios.yaml`: canonical scenario matrix and suggested smoke/short/saturation profiles.
- `scripts/bench-matrix.sh`: scenario/rate/repeat matrix runner over `scripts/bench-vegeta.sh`.
- `scripts/bench-vegeta.sh` enhancements:
  - `--warmup-duration`,
  - richer `metadata.json`,
  - `metrics-delta.txt`,
  - Kanban scenarios,
  - copied per-run targets.
- `bench/scripts/kanban-board/app.js`: a mounted 120-card Kanban board fixture.
- `bench/targets/kanban-*.txt`: static targets for page, fragment, action, and mixed Kanban workload documentation.
- `deploy/observability/prometheus.example.yaml`: local scrape config.
- `deploy/observability/grafana-dashboard-goja-site-benchmark.json`: starter benchmark dashboard.
- Ticket scripts:
  - `scripts/run-phase7-smoke-matrix.sh`,
  - `scripts/run-phase7-short-matrix.sh`.

## Validation Commands

Unit/integration tests:

```bash
go test ./...
```

Dedicated valid Kanban action smoke:

```bash
scripts/bench-vegeta.sh \
  --scenario kanban-action \
  --duration 2s \
  --rate 2/s \
  --port 18210 \
  --metrics-port 19210
```

Full smoke matrix:

```bash
scripts/bench-matrix.sh \
  --scenarios null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed \
  --rates 2/s \
  --duration 2s \
  --warmup-duration 0s \
  --repeat 1 \
  --start-port 18240 \
  --start-metrics-port 19240 \
  --out-root bench/results/phase7-smoke-manual2
```

Kanban mixed action verification after target ordering fix:

```bash
scripts/bench-vegeta.sh \
  --scenario kanban-mixed \
  --duration 2s \
  --rate 2/s \
  --port 18260 \
  --metrics-port 19260
```

Metrics evidence from the mixed run:

```text
goja_site_kanban_action_duration_seconds_count{action="cardMoved",board="bench",refresh="true",site="default"} 1
goja_site_kanban_fragment_duration_seconds_count{board="bench",site="default"} 2
goja_site_kanban_render_duration_seconds_count{board="bench",reason="action_refresh",site="default"} 1
goja_site_kanban_render_duration_seconds_count{board="bench",reason="fragment",site="default"} 2
```

This proves the mixed scenario hits both valid fragment and valid `cardMoved` action paths, even in a tiny four-request smoke run.

## Smoke Matrix Results

The smoke matrix output root was:

```text
bench/results/phase7-smoke-manual2
```

These result directories are intentionally ignored by git, but the summary was:

| scenario | rate | run | requests | success | throughput | p50 ms | p95 ms | p99 ms | max ms | status | errors |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|---|
| null | 2/s | 1 | 4 | 1.0000 | 2.67 | 0.64 | 1.18 | 1.18 | 1.18 | `{"200": 4}` | `[]` |
| render | 2/s | 1 | 4 | 1.0000 | 2.66 | 2.18 | 2.75 | 2.75 | 2.75 | `{"200": 4}` | `[]` |
| db-read | 2/s | 1 | 4 | 1.0000 | 2.66 | 1.29 | 1.49 | 1.49 | 1.49 | `{"200": 4}` | `[]` |
| db-write | 2/s | 1 | 4 | 1.0000 | 2.66 | 4.33 | 6.52 | 6.52 | 6.52 | `{"200": 4}` | `[]` |
| kanban-fragment | 2/s | 1 | 4 | 1.0000 | 2.64 | 11.71 | 29.18 | 29.18 | 29.18 | `{"200": 4}` | `[]` |
| kanban-action | 2/s | 1 | 4 | 1.0000 | 2.65 | 16.12 | 25.65 | 25.65 | 25.65 | `{"200": 4}` | `[]` |
| kanban-mixed | 2/s | 1 | 4 | 1.0000 | 2.65 | 8.57 | 10.57 | 10.57 | 10.57 | `{"200": 4}` | `[]` |

## Interpretation

All smoke scenarios returned `200` for every request and had empty Vegeta error sets. This means the fixtures and targets are valid enough for longer runs.

The numbers are not capacity numbers because each scenario only issued four requests. They are useful only for correctness and sanity:

- null route starts and serves successfully,
- render route returns HTML,
- DB read/write routes are valid,
- Kanban fragment endpoint renders mounted board HTML,
- Kanban action endpoint accepts the `cardMoved` payload,
- Kanban mixed workload can hit fragment and action paths.

## Important Caveat

The first full matrix run succeeded but summary generation failed with a Python formatting bug:

```text
TypeError: str.format() got multiple values for keyword argument 'errors'
```

The bug was in `scripts/bench-matrix.sh`, where the row dictionary already contained an `errors` key and the formatter also passed `errors=...`. I fixed this by renaming formatter arguments to `status_json` and `errors_json`, then reran the smoke matrix successfully.

## Next Benchmarking Steps

1. Run the short matrix from the ticket script:

```bash
ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/scripts/run-phase7-short-matrix.sh
```

2. Review `matrix-summary.md` for each scenario/rate.
3. Identify the first scenario/rate where p95 latency or errors bend upward.
4. Rerun that specific scenario with `--pprof`.
5. Only use tracing at low sample rates or short 100%-sample diagnostic runs.

## Conclusion

The Phase 7 benchmark harness is now ready for real short and saturation runs. The next valuable work is collecting comparable baseline matrices, not adding more fine-grained spans.
