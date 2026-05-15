---
Title: Anti-overfit benchmark report - anti-overfit-v1-20260515T182902Z
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - benchmarking
    - performance
    - kanban
    - optimization
DocType: reference
Intent: historical
Owners: []
RelatedFiles:
    - Path: bench/scripts/kanban-board-10/app.js
      Note: 10-card Kanban scaling fixture
    - Path: bench/scripts/kanban-board-500/app.js
      Note: 500-card Kanban scaling fixture
    - Path: bench/scripts/render-shapes/app.js
      Note: generic UI DSL render shape fixture
    - Path: scripts/bench-vegeta.sh
      Note: scenario support for anti-overfit benchmark matrix
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/scripts/04-run-anti-overfit-matrix.sh
      Note: anti-overfit matrix runner
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/scripts/05-render-anti-overfit-report.py
      Note: anti-overfit report renderer
ExternalSources: []
Summary: Results from Anti-overfit matrix v1 covering Kanban sizes, renderer shapes, and DB workloads.
LastUpdated: 2026-05-15T18:55:39Z
WhatFor: ""
WhenToUse: ""
---


# Anti-overfit benchmark report: anti-overfit-v1-20260515T182902Z

Raw matrix root: `bench/results/anti-overfit-v1-20260515T182902Z`

This report broadens the post-simplification evidence beyond the original 120-card Kanban fixture. It checks three workload classes: Kanban size scaling, generic UI DSL render shapes, and database-heavy request paths.

## Matrix shape

| field | value |
|---|---|
| scenarios | db-read-100, db-write-batch-10, kanban-fragment, kanban-fragment-10, kanban-fragment-500, render-attrs-1000, render-flat-1000 |
| rates | 25/s, 100/s, 250/s |
| runs | 63 |
| repeat count per cell | 3 |

## Executive summary

The matrix found cells that deserve follow-up because they missed 98% throughput ratio, had non-100% success, or exceeded 250 ms average p95:

| scenario | rate | success avg | throughput ratio avg | p95 avg ms | p95 max ms |
|---|---|---|---|---|---|
| render-attrs-1000 | 250/s | 0.71 | 0.24 | 30001.02 | 30001.09 |
| db-write-batch-10 | 250/s | 1 | 0.38 | 23260.57 | 24828.24 |
| kanban-fragment-500 | 250/s | 1 | 0.41 | 20858.40 | 22539.80 |
| render-attrs-1000 | 100/s | 1 | 0.51 | 13841.81 | 17316.67 |
| render-flat-1000 | 250/s | 1 | 0.69 | 6524.56 | 7267.49 |
| kanban-fragment-500 | 100/s | 1 | 0.73 | 5460.73 | 5706.05 |
| db-write-batch-10 | 100/s | 1 | 0.87 | 2554.00 | 5547.67 |
| db-read-100 | 250/s | 1 | 0.88 | 2278.69 | 4040.21 |
| kanban-fragment | 250/s | 1 | 0.99 | 435.02 | 1194.05 |


## Key findings

- The 10-card Kanban fragment stayed extremely cheap through `250/s`: average p95 `2.14 ms`, throughput ratio `1.00`.
- The 120-card Kanban fragment remained healthy at `100/s` with average p95 `9.43 ms`; at `250/s` it still delivered nearly full throughput but showed queueing, with average p95 `435.02 ms` and one repeat reaching p95 `1194.05 ms`.
- The 500-card Kanban fragment was healthy at `25/s` with average p95 `21.08 ms`, but saturated by `100/s`: throughput ratio `0.73`, average p95 `5460.73 ms`.
- Generic large render shapes are now the strongest warning against overfitting: `render-flat-1000` saturated at `250/s`, and `render-attrs-1000` saturated already at `100/s` and timed out at `250/s`.
- DB reads with 100 queries per request stayed healthy through `100/s`, then queued at `250/s`.
- Batch writes of 10 inserts per request showed queueing already at `100/s`, so write batching/transaction behavior is a separate follow-up from render optimization.

The main conclusion is that removing `preciseMoveForm` fixed the original Kanban-specific markup explosion, but the broader suite found three remaining workload families to investigate separately: very large rendered trees, attribute-heavy rendered trees, and write-heavy SQLite actions.

## Workload-family comparison

| family | cells | p95 min ms | p95 avg ms | p95 max ms | min throughput ratio |
|---|---|---|---|---|---|
| Kanban size scaling | 9 | 2.02 | 2977.66 | 20858.40 | 0.41 |
| UI shape rendering | 6 | 16.12 | 8420.73 | 30001.02 | 0.24 |
| Database-heavy | 6 | 13.73 | 4699.49 | 23260.57 | 0.38 |

## Scenario/rate aggregates

| family | scenario | rate | runs | requests | success avg | throughput avg | throughput ratio avg | p50 avg ms | p95 avg ms | p95 max ms | p99 avg ms | max ms | errors |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Database-heavy | db-read-100 | 25/s | 3 | 1125 | 1 | 25.06 | 1.00 | 4.22 | 13.73 | 14.74 | 21.08 | 24.67 | 0 |
| Database-heavy | db-read-100 | 100/s | 3 | 4500 | 1 | 100.02 | 1.00 | 3.78 | 23.15 | 29.53 | 34.60 | 57.85 | 0 |
| Database-heavy | db-read-100 | 250/s | 3 | 11250 | 1 | 218.97 | 0.88 | 719.92 | 2278.69 | 4040.21 | 2356.88 | 4111.56 | 0 |
| Database-heavy | db-write-batch-10 | 25/s | 3 | 1125 | 1 | 25.04 | 1.00 | 12.21 | 66.81 | 168.62 | 123.48 | 395.19 | 0 |
| Database-heavy | db-write-batch-10 | 100/s | 3 | 4500 | 1 | 86.95 | 0.87 | 1643.74 | 2554.00 | 5547.67 | 2607.34 | 5693.73 | 0 |
| Database-heavy | db-write-batch-10 | 250/s | 3 | 11250 | 1 | 95.17 | 0.38 | 12129.61 | 23260.57 | 24828.24 | 24224.66 | 26067.82 | 0 |
| Kanban size scaling | kanban-fragment | 25/s | 3 | 1125 | 1 | 25.06 | 1.00 | 3.39 | 7.84 | 8.64 | 10.97 | 16.13 | 0 |
| Kanban size scaling | kanban-fragment | 100/s | 3 | 4500 | 1 | 100.04 | 1.00 | 3.21 | 9.43 | 9.77 | 26.51 | 63.45 | 0 |
| Kanban size scaling | kanban-fragment | 250/s | 3 | 11250 | 1 | 247.71 | 0.99 | 176.20 | 435.02 | 1194.05 | 496.33 | 1297.35 | 0 |
| Kanban size scaling | kanban-fragment-10 | 25/s | 3 | 1125 | 1 | 25.07 | 1.00 | 0.79 | 2.02 | 2.17 | 3.13 | 5.83 | 0 |
| Kanban size scaling | kanban-fragment-10 | 100/s | 3 | 4500 | 1 | 100.06 | 1.00 | 0.83 | 2.24 | 2.45 | 3.92 | 18.06 | 0 |
| Kanban size scaling | kanban-fragment-10 | 250/s | 3 | 11250 | 1 | 250.05 | 1.00 | 0.70 | 2.14 | 2.28 | 4.52 | 61.78 | 0 |
| Kanban size scaling | kanban-fragment-500 | 25/s | 3 | 1125 | 1 | 25.05 | 1.00 | 9.55 | 21.08 | 24.90 | 32.69 | 67.17 | 0 |
| Kanban size scaling | kanban-fragment-500 | 100/s | 3 | 4500 | 1 | 73.12 | 0.73 | 3302.34 | 5460.73 | 5706.05 | 5576.06 | 5767.81 | 0 |
| Kanban size scaling | kanban-fragment-500 | 250/s | 3 | 11250 | 1 | 102.31 | 0.41 | 10770.72 | 20858.40 | 22539.80 | 21548.95 | 23424.75 | 0 |
| UI shape rendering | render-attrs-1000 | 25/s | 3 | 1125 | 1 | 25.03 | 1.00 | 16.29 | 37.33 | 38.80 | 50.05 | 67.49 | 0 |
| UI shape rendering | render-attrs-1000 | 100/s | 3 | 4500 | 1 | 50.99 | 0.51 | 6347.97 | 13841.81 | 17316.67 | 14615.66 | 18777.60 | 0 |
| UI shape rendering | render-attrs-1000 | 250/s | 3 | 11250 | 0.71 | 59.49 | 0.24 | 21267.61 | 30001.02 | 30001.09 | 30001.75 | 30008.34 | 4 |
| UI shape rendering | render-flat-1000 | 25/s | 3 | 1125 | 1 | 25.06 | 1.00 | 5.51 | 16.12 | 16.79 | 24.47 | 33.17 | 0 |
| UI shape rendering | render-flat-1000 | 100/s | 3 | 4500 | 1 | 100.01 | 1.00 | 10.06 | 103.53 | 257.36 | 131.80 | 288.29 | 0 |
| UI shape rendering | render-flat-1000 | 250/s | 3 | 11250 | 1 | 171.67 | 0.69 | 3658.68 | 6524.56 | 7267.49 | 6774.36 | 7878.96 | 0 |

## Highest-rate cell per scenario

| scenario | rate | success avg | throughput avg | throughput ratio avg | p95 avg ms | p99 avg ms |
|---|---|---|---|---|---|---|
| db-read-100 | 250/s | 1 | 218.97 | 0.88 | 2278.69 | 2356.88 |
| db-write-batch-10 | 250/s | 1 | 95.17 | 0.38 | 23260.57 | 24224.66 |
| kanban-fragment | 250/s | 1 | 247.71 | 0.99 | 435.02 | 496.33 |
| kanban-fragment-10 | 250/s | 1 | 250.05 | 1.00 | 2.14 | 4.52 |
| kanban-fragment-500 | 250/s | 1 | 102.31 | 0.41 | 20858.40 | 21548.95 |
| render-attrs-1000 | 250/s | 0.71 | 59.49 | 0.24 | 30001.02 | 30001.75 |
| render-flat-1000 | 250/s | 1 | 171.67 | 0.69 | 6524.56 | 6774.36 |

## Slowest individual runs by p95

| scenario | rate | run | success | throughput | throughput ratio | p50 ms | p95 ms | p99 ms | max ms | dir |
|---|---|---|---|---|---|---|---|---|---|---|
| render-attrs-1000 | 250/s | 3 | 0.63 | 52.25 | 0.21 | 23770.14 | 30001.09 | 30001.85 | 30008.34 | render-attrs-1000/rate-250_s/run-3 |
| render-attrs-1000 | 250/s | 2 | 0.74 | 61.96 | 0.25 | 20279.34 | 30001.00 | 30001.83 | 30004.05 | render-attrs-1000/rate-250_s/run-2 |
| render-attrs-1000 | 250/s | 1 | 0.77 | 64.27 | 0.26 | 19753.36 | 30000.96 | 30001.57 | 30006.41 | render-attrs-1000/rate-250_s/run-1 |
| db-write-batch-10 | 250/s | 1 | 1 | 91.32 | 0.37 | 11484.09 | 24828.24 | 25816.58 | 26067.82 | db-write-batch-10/rate-250_s/run-1 |
| db-write-batch-10 | 250/s | 3 | 1 | 94.13 | 0.38 | 13582.08 | 23555.21 | 24609.07 | 24843.96 | db-write-batch-10/rate-250_s/run-3 |
| kanban-fragment-500 | 250/s | 1 | 1 | 97.60 | 0.39 | 13919.44 | 22539.80 | 23313.59 | 23424.75 | kanban-fragment-500/rate-250_s/run-1 |
| db-write-batch-10 | 250/s | 2 | 1 | 100.07 | 0.40 | 11322.65 | 21398.27 | 22248.33 | 22477.99 | db-write-batch-10/rate-250_s/run-2 |
| kanban-fragment-500 | 250/s | 2 | 1 | 103.98 | 0.42 | 10215.02 | 20348.09 | 20915.17 | 21070.44 | kanban-fragment-500/rate-250_s/run-2 |
| kanban-fragment-500 | 250/s | 3 | 1 | 105.34 | 0.42 | 8177.69 | 19687.30 | 20418.07 | 20603.79 | kanban-fragment-500/rate-250_s/run-3 |
| render-attrs-1000 | 100/s | 1 | 1 | 44.42 | 0.44 | 6142.09 | 17316.67 | 18683.83 | 18777.60 | render-attrs-1000/rate-100_s/run-1 |
| render-attrs-1000 | 100/s | 2 | 1 | 53.79 | 0.54 | 6182.62 | 12156.01 | 12779.70 | 12894.67 | render-attrs-1000/rate-100_s/run-2 |
| render-attrs-1000 | 100/s | 3 | 1 | 54.77 | 0.55 | 6719.19 | 12052.77 | 12383.44 | 12400.68 | render-attrs-1000/rate-100_s/run-3 |
| render-flat-1000 | 250/s | 2 | 1 | 163.94 | 0.66 | 3574.07 | 7267.49 | 7785.77 | 7878.96 | render-flat-1000/rate-250_s/run-2 |
| render-flat-1000 | 250/s | 1 | 1 | 174.11 | 0.70 | 3851.39 | 6261.72 | 6393.41 | 6546.77 | render-flat-1000/rate-250_s/run-1 |
| render-flat-1000 | 250/s | 3 | 1 | 176.96 | 0.71 | 3550.58 | 6044.48 | 6143.91 | 6195.89 | render-flat-1000/rate-250_s/run-3 |

## Interpretation

Use this matrix as a regression and discovery tool, not a final capacity claim. The tested duration is intentionally short enough to run during development. If a workload family shows a knee here, follow up with a targeted longer run and pprof. If all cells pass, the next useful expansion is mixed multi-site traffic or a browser-side timing matrix.

## Recommended follow-up

1. Add threshold checking around this matrix so it can fail on p95, success, response-byte, or throughput-ratio regressions.
2. Add a 1000-card Kanban fragment once 500-card behavior is understood.
3. Add mixed `serve-multi` traffic with null, render, DB, and Kanban sites in the same process.
4. Capture pprof only for the first cell that exceeds the threshold, rather than profiling every run.
