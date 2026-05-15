---
Title: 'Multi-VM serve-multi quick stress report: multi-vm-quick-20260515T163935Z'
Ticket: GOJA-MULTI-VM-STRESS
Status: active
Topics:
    - benchmarking
    - stress-testing
    - goja
    - observability
DocType: reference
Intent: historical
Owners: []
RelatedFiles:
    - Path: bench/results/multi-vm-quick-20260515T163935Z
      Note: ignored raw Vegeta/metrics artifacts for quick multi-VM sweep
ExternalSources: []
Summary: Markdown summary generated from multi-VM serve-multi stress run artifacts.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Multi-VM serve-multi quick stress report: multi-vm-quick-20260515T163935Z

Result root: `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/multi-vm-quick-20260515T163935Z`

## Runs

| scenario | vm_count | distribution | rate | requests | throughput | success | p50 ms | p95 ms | p99 ms | max ms | status | errors | dir |
|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|---|---|---|
| kanban-fragment | 1 | even-hot | 50/s | 500 | 50.06 | 1.0000 | 7.67 | 20.31 | 28.29 | 48.48 | `{"200": 500}` | `[]` | `kanban-fragment/even-hot/1vm/rate-50_s` |
| kanban-fragment | 2 | even-hot | 50/s | 500 | 50.06 | 1.0000 | 8.33 | 16.95 | 26.49 | 49.44 | `{"200": 500}` | `[]` | `kanban-fragment/even-hot/2vm/rate-50_s` |
| kanban-fragment | 4 | even-hot | 50/s | 500 | 50.06 | 1.0000 | 7.90 | 18.01 | 32.38 | 62.71 | `{"200": 500}` | `[]` | `kanban-fragment/even-hot/4vm/rate-50_s` |
| kanban-fragment | 8 | even-hot | 50/s | 500 | 50.03 | 1.0000 | 8.75 | 19.00 | 26.10 | 27.36 | `{"200": 500}` | `[]` | `kanban-fragment/even-hot/8vm/rate-50_s` |
| null | 1 | even-hot | 200/s | 2000 | 200.09 | 1.0000 | 0.28 | 0.66 | 0.94 | 1.43 | `{"200": 2000}` | `[]` | `null/even-hot/1vm/rate-200_s` |
| null | 2 | even-hot | 200/s | 2000 | 200.10 | 1.0000 | 0.26 | 0.62 | 0.88 | 3.44 | `{"200": 2000}` | `[]` | `null/even-hot/2vm/rate-200_s` |
| null | 4 | even-hot | 200/s | 2000 | 200.07 | 1.0000 | 0.30 | 0.73 | 1.13 | 3.97 | `{"200": 2000}` | `[]` | `null/even-hot/4vm/rate-200_s` |
| null | 8 | even-hot | 200/s | 2000 | 200.09 | 1.0000 | 0.34 | 0.86 | 2.19 | 10.98 | `{"200": 2000}` | `[]` | `null/even-hot/8vm/rate-200_s` |
| null | 2 | one-hot | 200/s | 2000 | 200.09 | 1.0000 | 0.30 | 0.72 | 1.07 | 3.31 | `{"200": 2000}` | `[]` | `null/one-hot/2vm/rate-200_s` |
| null | 4 | one-hot | 200/s | 2000 | 200.08 | 1.0000 | 0.31 | 0.68 | 1.06 | 10.49 | `{"200": 2000}` | `[]` | `null/one-hot/4vm/rate-200_s` |
| null | 8 | one-hot | 200/s | 2000 | 200.09 | 1.0000 | 0.30 | 0.70 | 1.09 | 4.64 | `{"200": 2000}` | `[]` | `null/one-hot/8vm/rate-200_s` |

## Headline findings

The quick multi-VM validation sweep succeeded. All 11 measured runs returned HTTP 200 only, with 100% Vegeta success and no Vegeta error sets.

The minimal `null` route stayed healthy at a total offered rate of 200/s across 1, 2, 4, and 8 configured Goja VMs. The p95 stayed below 1 ms in all `null` cells. The 8-VM even-hot case showed a higher max latency outlier of 10.98 ms, but the p95 remained 0.86 ms.

The one-hot idle-VM check also stayed healthy. With 2, 4, and 8 loaded VMs but all traffic sent to `site-001`, p95 stayed around 0.68-0.72 ms at 200/s. This quick run does not show meaningful idle-VM overhead at this VM count for the minimal route.

The heavier `kanban-fragment` route stayed healthy at 50/s total across 1, 2, 4, and 8 VMs. p95 ranged from 16.95 ms to 20.31 ms, with 100% success. This validates that Host-header distribution across several mounted Kanban fixtures works and that the quick multi-VM harness can exercise real Goja/Kanban work.

This is a validation sweep, not a saturation sweep. The next experiment should raise total rate for `null` and `kanban-fragment`, then add `kanban-action` only at rates chosen with the known single-VM knee in mind.

## Interpretation notes

- `vm_count` is the number of configured `serve-multi` sites, which means the number of Goja runtimes/VMs loaded by the process.
- `even-hot` distributes Vegeta targets across all hosts in the target file.
- `one-hot` sends traffic only to the first site while the remaining VMs stay idle.
- `skewed` repeats site 1 nine times, then includes each other site once.
- Throughput is total process throughput across the generated target set, not per-site throughput.
