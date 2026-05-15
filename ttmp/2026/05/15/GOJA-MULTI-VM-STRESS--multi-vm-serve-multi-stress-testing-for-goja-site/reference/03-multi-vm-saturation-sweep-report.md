---
Title: 'Multi-VM serve-multi saturation sweep report: multi-vm-saturation-20260515T164805Z'
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
    - Path: bench/results/multi-vm-saturation-20260515T164805Z
      Note: ignored raw Vegeta/metrics artifacts for multi-VM saturation sweep
ExternalSources: []
Summary: Markdown summary generated from multi-VM serve-multi stress run artifacts.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Multi-VM serve-multi saturation sweep report: multi-vm-saturation-20260515T164805Z

Result root: `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/multi-vm-saturation-20260515T164805Z`

## Headline findings

The saturation sweep found an inflection point for the real Kanban fragment workload, but not for the minimal `null` route in the tested range.

All 28 measured runs returned HTTP 200 only with 100% Vegeta success and no Vegeta error sets. The failure mode, where present, was queueing and tail-latency growth rather than HTTP failure.

### Minimal `null` route

The `null` route did not saturate through 2000/s total offered rate for 1, 2, 4, or 8 configured VMs. p95 stayed below 0.6 ms in every `null` cell. This means the serve-multi Host dispatch path and minimal Goja handler path are not the first bottleneck at these rates.

### Kanban fragment route

The `kanban-fragment` route shows the first clear saturation behavior. At 100/s total, all VM counts remained usable, with p95 between 21 ms and 66 ms. At 200/s total, the result depended on VM count: 1 VM fell badly behind, 2 VMs showed significant queueing, 4 VMs stayed much healthier, and 8 VMs showed high p95 variance. At 400/s total, every VM count saturated heavily.

The strongest inflection points are:

| Case | Signal |
|---|---|
| 1 VM, 100/s -> 200/s | p95 jumps from 58.91 ms to 5901.66 ms; throughput falls to 124.64/s. |
| 2 VMs, 100/s -> 200/s | p95 jumps from 65.83 ms to 720.15 ms; throughput falls to 183.80/s. |
| 4 VMs, 200/s -> 400/s | p95 jumps from 179.41 ms to 5861.77 ms; throughput falls to 246.53/s. |
| 8 VMs, 200/s -> 400/s | p95 jumps from 564.94 ms to 5559.25 ms; throughput falls to 249.44/s. |

The approximate useful ceiling for this fixture is therefore around 100/s for 1 VM, around 100-200/s for 2 VMs, around 200/s for 4 VMs, and below 400/s for 8 VMs. The 8-VM 200/s cell had throughput close to target but p95 564.94 ms, so it should be treated as degraded even though throughput ratio remained high.

### Scaling interpretation

Adding VMs helps, but the 400/s rows show a process-level or shared-resource ceiling around 240-250/s for this large-response Kanban fragment fixture. The workload returns about 246 KB per request, so response rendering and response writing are both significant. The multi-VM architecture improves the owner-loop bottleneck, but it does not remove CPU, allocation, GC, and response-output costs.

## Runs

| scenario | vm_count | distribution | rate | requests | throughput | success | p50 ms | p95 ms | p99 ms | max ms | status | errors | dir |
|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|---|---|---|
| kanban-fragment | 1 | even-hot | 100/s | 1000 | 100.04 | 1.0000 | 12.18 | 58.91 | 77.78 | 89.31 | `{"200": 1000}` | `[]` | `kanban-fragment/even-hot/1vm/rate-100_s` |
| kanban-fragment | 1 | even-hot | 200/s | 2000 | 124.64 | 1.0000 | 2736.46 | 5901.66 | 6009.52 | 6052.11 | `{"200": 2000}` | `[]` | `kanban-fragment/even-hot/1vm/rate-200_s` |
| kanban-fragment | 1 | even-hot | 400/s | 4000 | 140.28 | 1.0000 | 9813.36 | 17523.84 | 18421.36 | 18516.61 | `{"200": 4000}` | `[]` | `kanban-fragment/even-hot/1vm/rate-400_s` |
| kanban-fragment | 2 | even-hot | 100/s | 1000 | 100.02 | 1.0000 | 7.36 | 65.83 | 190.53 | 219.83 | `{"200": 1000}` | `[]` | `kanban-fragment/even-hot/2vm/rate-100_s` |
| kanban-fragment | 2 | even-hot | 200/s | 2000 | 183.80 | 1.0000 | 429.62 | 720.15 | 831.36 | 899.64 | `{"200": 2000}` | `[]` | `kanban-fragment/even-hot/2vm/rate-200_s` |
| kanban-fragment | 2 | even-hot | 400/s | 4000 | 193.19 | 1.0000 | 6155.08 | 10175.94 | 10580.35 | 10710.09 | `{"200": 4000}` | `[]` | `kanban-fragment/even-hot/2vm/rate-400_s` |
| kanban-fragment | 4 | even-hot | 100/s | 1000 | 100.04 | 1.0000 | 7.60 | 51.29 | 202.67 | 272.39 | `{"200": 1000}` | `[]` | `kanban-fragment/even-hot/4vm/rate-100_s` |
| kanban-fragment | 4 | even-hot | 200/s | 2000 | 199.87 | 1.0000 | 13.81 | 179.41 | 214.69 | 240.71 | `{"200": 2000}` | `[]` | `kanban-fragment/even-hot/4vm/rate-200_s` |
| kanban-fragment | 4 | even-hot | 400/s | 4000 | 246.53 | 1.0000 | 3126.74 | 5861.77 | 6097.19 | 6237.54 | `{"200": 4000}` | `[]` | `kanban-fragment/even-hot/4vm/rate-400_s` |
| kanban-fragment | 8 | even-hot | 100/s | 1000 | 100.01 | 1.0000 | 7.90 | 21.30 | 93.61 | 136.77 | `{"200": 1000}` | `[]` | `kanban-fragment/even-hot/8vm/rate-100_s` |
| kanban-fragment | 8 | even-hot | 200/s | 2000 | 198.41 | 1.0000 | 20.75 | 564.94 | 665.56 | 719.93 | `{"200": 2000}` | `[]` | `kanban-fragment/even-hot/8vm/rate-200_s` |
| kanban-fragment | 8 | even-hot | 400/s | 4000 | 249.44 | 1.0000 | 2804.86 | 5559.25 | 6235.44 | 6412.28 | `{"200": 4000}` | `[]` | `kanban-fragment/even-hot/8vm/rate-400_s` |
| null | 1 | even-hot | 1200/s | 12000 | 1200.03 | 1.0000 | 0.15 | 0.33 | 0.58 | 21.79 | `{"200": 12000}` | `[]` | `null/even-hot/1vm/rate-1200_s` |
| null | 1 | even-hot | 2000/s | 19999 | 2000.05 | 1.0000 | 0.20 | 0.41 | 0.94 | 13.14 | `{"200": 19999}` | `[]` | `null/even-hot/1vm/rate-2000_s` |
| null | 1 | even-hot | 400/s | 4000 | 400.07 | 1.0000 | 0.18 | 0.48 | 0.79 | 8.78 | `{"200": 4000}` | `[]` | `null/even-hot/1vm/rate-400_s` |
| null | 1 | even-hot | 800/s | 8000 | 800.13 | 1.0000 | 0.15 | 0.39 | 0.61 | 3.83 | `{"200": 8000}` | `[]` | `null/even-hot/1vm/rate-800_s` |
| null | 2 | even-hot | 1200/s | 12000 | 1200.00 | 1.0000 | 0.16 | 0.37 | 0.58 | 1.81 | `{"200": 12000}` | `[]` | `null/even-hot/2vm/rate-1200_s` |
| null | 2 | even-hot | 2000/s | 20000 | 2000.06 | 1.0000 | 0.16 | 0.33 | 0.53 | 1.93 | `{"200": 20000}` | `[]` | `null/even-hot/2vm/rate-2000_s` |
| null | 2 | even-hot | 400/s | 4000 | 400.12 | 1.0000 | 0.17 | 0.49 | 0.78 | 1.92 | `{"200": 4000}` | `[]` | `null/even-hot/2vm/rate-400_s` |
| null | 2 | even-hot | 800/s | 8000 | 800.15 | 1.0000 | 0.17 | 0.43 | 0.66 | 3.00 | `{"200": 8000}` | `[]` | `null/even-hot/2vm/rate-800_s` |
| null | 4 | even-hot | 1200/s | 12000 | 1200.02 | 1.0000 | 0.16 | 0.36 | 0.57 | 3.29 | `{"200": 12000}` | `[]` | `null/even-hot/4vm/rate-1200_s` |
| null | 4 | even-hot | 2000/s | 19998 | 2000.01 | 1.0000 | 0.17 | 0.34 | 0.61 | 8.43 | `{"200": 19998}` | `[]` | `null/even-hot/4vm/rate-2000_s` |
| null | 4 | even-hot | 400/s | 4000 | 400.08 | 1.0000 | 0.17 | 0.40 | 0.70 | 3.90 | `{"200": 4000}` | `[]` | `null/even-hot/4vm/rate-400_s` |
| null | 4 | even-hot | 800/s | 8000 | 800.09 | 1.0000 | 0.17 | 0.43 | 0.68 | 3.09 | `{"200": 8000}` | `[]` | `null/even-hot/4vm/rate-800_s` |
| null | 8 | even-hot | 1200/s | 12000 | 1200.06 | 1.0000 | 0.17 | 0.40 | 0.68 | 6.86 | `{"200": 12000}` | `[]` | `null/even-hot/8vm/rate-1200_s` |
| null | 8 | even-hot | 2000/s | 19999 | 2000.01 | 1.0000 | 0.17 | 0.34 | 0.59 | 3.95 | `{"200": 19999}` | `[]` | `null/even-hot/8vm/rate-2000_s` |
| null | 8 | even-hot | 400/s | 4000 | 400.07 | 1.0000 | 0.17 | 0.51 | 0.85 | 2.99 | `{"200": 4000}` | `[]` | `null/even-hot/8vm/rate-400_s` |
| null | 8 | even-hot | 800/s | 8000 | 800.13 | 1.0000 | 0.18 | 0.47 | 0.70 | 3.17 | `{"200": 8000}` | `[]` | `null/even-hot/8vm/rate-800_s` |

## Interpretation notes

- `vm_count` is the number of configured `serve-multi` sites, which means the number of Goja runtimes/VMs loaded by the process.
- `even-hot` distributes Vegeta targets across all hosts in the target file.
- `one-hot` sends traffic only to the first site while the remaining VMs stay idle.
- `skewed` repeats site 1 nine times, then includes each other site once.
- Throughput is total process throughput across the generated target set, not per-site throughput.
