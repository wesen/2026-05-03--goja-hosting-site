---
Title: Post-simplification Kanban benchmark report
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - goja
    - kanban
    - performance
    - benchmarking
DocType: reference
Intent: historical
Owners: []
RelatedFiles:
    - Path: bench/results/kanban-simplified-20260515T173954Z
      Note: ignored raw benchmark output for post-simplification runs
ExternalSources: []
Summary: Benchmark report after removing eager preciseMove forms and adding compact frontend accessibility actions.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Post-simplification Kanban benchmark report

## What changed before this benchmark

The Kanban DSL was simplified. The server no longer renders a full `preciseMoveForm` on every card. The benchmark fixture now uses:

```javascript
.features(features => features.search().dragDrop())
```

The renderer now emits a small card `Actions` button with ARIA attributes instead of a full form. The client runtime implements the accessible action menu, keyboard handling, live-region announcements, and focus restoration.

## Command

```bash
ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/scripts/02-run-post-simplification-benchmarks.sh
```

## Result root

```text
bench/results/kanban-simplified-20260515T173954Z
```

The raw result directory is ignored by Git. This report records the key numbers.

## Results

| Case | Requests | Offered rate | Throughput | Success | p50 | p95 | p99 | max | Mean response bytes |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| single `kanban-fragment` | 2000 | 200/s | 200.03/s | 100% | 2.494 ms | 9.887 ms | 13.489 ms | 19.032 ms | 60,831 |
| single `kanban-action` | 1000 | 100/s | 100.06/s | 100% | 2.500 ms | 8.165 ms | 11.388 ms | 20.651 ms | 77,035 |
| multi 4VM `kanban-fragment` | 4000 | 400/s | 399.76/s | 100% | 2.792 ms | 7.760 ms | 14.725 ms | 51.730 ms | 60,831 |

All runs returned HTTP 200 only and no Vegeta errors.

## Comparison to previous saturation evidence

The old `kanban-fragment` response was about 245,946 bytes. The simplified fragment response is about 60,831 bytes.

The old multi-VM 4VM `kanban-fragment` 400/s saturation cell had:

```text
throughput 246.53/s
p95 5861.77 ms
mean response bytes 245,946
```

The same shape after simplification has:

```text
throughput 399.76/s
p95 7.76 ms
mean response bytes 60,831
```

The old single-VM `kanban-action` 100/s stress result had p95 above 1.5 seconds. The simplified `kanban-action` 100/s run has p95 8.165 ms.

## Interpretation

The result strongly supports the design decision. Eagerly rendering `preciseMoveForm` on every card was not a minor cost. Removing it reduced response size by roughly 75% for the fragment path and turned a heavily saturated 4VM/400s run into a healthy run.

The improvement is consistent with the pprof evidence. The previous profiles showed `preciseMoveForm`, `uidsl.renderNode`, `uidsl.renderAttrs`, allocation, and GC as the dominant costs. Removing the form reduces the number of nodes, attributes, options, strings, buffers, and response bytes created on every render.

## Remaining caveat

This benchmark validates the server-side performance win and the basic Go tests for the new client runtime markers. It does not replace browser-based accessibility testing. The next accessibility validation should exercise keyboard focus, menu navigation, live-region announcement behavior, and focus restoration in a browser or Playwright test.
