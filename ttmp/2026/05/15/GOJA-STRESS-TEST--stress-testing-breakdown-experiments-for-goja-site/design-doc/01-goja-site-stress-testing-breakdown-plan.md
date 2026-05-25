---
Title: Goja Site Stress Testing Breakdown Plan
Ticket: GOJA-STRESS-TEST
Status: active
Topics:
    - benchmarking
    - stress-testing
    - sqlite
    - observability
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: scripts/bench-matrix.sh
      Note: shared benchmark matrix runner used by stress scripts
    - Path: ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/scripts/run-stress-hour-sweep.sh
      Note: hour-scale stress sweep script
    - Path: ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/scripts/run-stress-quick-sweep.sh
      Note: quick validation stress sweep before long experiment
ExternalSources: []
Summary: Plan for short validation and hour-scale stress sweeps that identify where goja-site scenarios bend or fail.
LastUpdated: 2026-05-15T15:00:00Z
WhatFor: Use this to run stress tests only after smoke/short benchmark validity is known, and to classify breakdown with SQL-backed evidence.
WhenToUse: Before running saturation or hour-scale experiments against goja-site scenarios.
---




# Goja Site Stress Testing Breakdown Plan

## Executive Summary

This ticket extends `GOJA-PERF-BENCH` from baseline benchmarking into stress testing. The goal is to answer:

> At what offered load does each scenario bend or fail, and how does it fail?

A stress test result is not just a p95 number. The report should classify breakdown using success ratio, achieved throughput versus offered rate, latency-knee growth, HTTP status/error sets, Prometheus metric deltas, and eventually pprof evidence for the first scenario that bends.

The ticket has two run scripts:

- `scripts/run-stress-quick-sweep.sh`: a short 2-3 minute validation sweep before any long run.
- `scripts/run-stress-hour-sweep.sh`: an hour-scale sweep over more scenarios and rates.

Both scripts store results in SQLite and render Markdown reports with every SQL query embedded before its result table.

## Problem Statement

The previous benchmark work established that core routes and Kanban actions are valid and that a short baseline matrix runs cleanly through 25/s. It did not find breakdown points. We need a controlled stress workflow that can safely increase load and report the first meaningful failure or saturation signals.

## Breakdown Signals

The report flags a scenario/rate cell as a breakdown candidate if any of the following occurs:

1. `success_min < 0.99`.
2. Vegeta error set is non-empty.
3. Achieved throughput is less than 95% of offered throughput.
4. p95 exceeds 100 ms for light scenarios:
   - `null`,
   - `render`,
   - `db-read`.
5. p95 exceeds 250 ms for heavier scenarios:
   - `db-write`,
   - `kanban-fragment`,
   - `kanban-action`,
   - `kanban-mixed`.
6. Max latency exceeds 1000 ms.

These thresholds are deliberately conservative first-pass signals, not production SLOs.

## Quick Sweep

The quick sweep is designed to prove the machinery works without spending an hour.

Default shape:

```text
scenarios: null,render,db-write,kanban-action
rates:     50/s,100/s,200/s
duration:  10s measured
warmup:    3s
repeat:    1
```

Approximate runtime:

```text
4 scenarios × 3 rates × (3s warmup + 10s measured) = 156s plus startup/report overhead
```

Command:

```bash
ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/scripts/run-stress-quick-sweep.sh
```

## Hour-Scale Sweep

The hour-scale sweep should only run after the quick sweep succeeds.

Default shape:

```text
scenarios: null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed
rates:     25/s,50/s,100/s,200/s,400/s,800/s
duration:  60s measured
warmup:    10s
repeat:    1
```

Approximate runtime:

```text
7 scenarios × 6 rates × (10s warmup + 60s measured) = 49 minutes plus startup/report overhead
```

Command:

```bash
ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/scripts/run-stress-hour-sweep.sh
```

## SQLite Storage

SQLite database:

```text
archive/stress-benchmarks.sqlite
```

Tables:

```text
benchmark_matrices
benchmark_runs
benchmark_metric_deltas
```

The importer stores normalized fields and raw artifacts:

- matrix metadata,
- per-run Vegeta summaries,
- raw `metadata.json`,
- raw `vegeta.json`,
- raw `metrics-delta.txt`,
- parsed metric deltas.

## Report Sections

The stress report renderer writes:

1. Matrix metadata.
2. Offered versus achieved throughput and latency.
3. Breakdown candidates.
4. Latency knee growth between adjacent rates.
5. Slowest individual runs by p95.
6. Runs with non-100% success or Vegeta errors.
7. HTTP route deltas.
8. DB operation and error deltas.
9. Kanban action/render deltas.
10. Stored artifact locations.

Every section includes the SQL query used to generate it.

## Expected Failure Modes

### Null route

Likely bottleneck: HTTP server, Goja handler dispatch, or load generator limits.

### Render route

Likely bottleneck: UI DSL object construction, normalization, and HTML rendering.

### DB read

Likely bottleneck: repeated SQLite selects and Goja-native module call overhead. The current fixture performs 10 selects per HTTP request.

### DB write

Likely bottleneck: SQLite write serialization, filesystem behavior, or lock contention.

### Kanban fragment

Likely bottleneck: rendering a 120-card board and returning a large HTML response.

### Kanban action

Likely bottleneck: action dispatch plus full board refresh render in the JSON response. This was the slowest scenario in the short baseline matrix.

### Kanban mixed

Likely bottleneck: same as fragment/action, but more representative because it mixes page, fragment, and action requests.

## Follow-up Profiling

After a stress sweep identifies a bending scenario, rerun one targeted rate with pprof:

```bash
scripts/bench-vegeta.sh \
  --scenario kanban-action \
  --duration 60s \
  --warmup-duration 10s \
  --rate 200/s \
  --pprof \
  --pprof-seconds 30
```

Use pprof only after narrowing the suspected bottleneck.

## Design Decisions

- Use SQLite instead of MySQL so the result database travels with the ticket.
- Use a quick sweep before any hour-scale run to validate scripts, ports, actions, and import/report flow.
- Build `goja-site` once per stress script and reuse the binary across matrix cells.
- Store ignored raw `bench/results/...` paths in SQLite so reports can point back to raw artifacts.
- Keep SQL visible in the Markdown report so reviewers can reproduce or modify every table.

## Open Questions

- Are 800/s rates too high for the local machine or Vegeta before the app itself bends?
- Should the hour sweep include repeats around the first observed knee, or should that be a separate targeted script?
- Should future reports include explicit threshold pass/fail columns from `bench/scenarios.yaml`?
