---
Title: Multi-VM Kanban Fragment pprof Report
Ticket: GOJA-MULTI-VM-STRESS
Status: active
Topics:
    - benchmarking
    - stress-testing
    - goja
    - pprof
DocType: reference
Intent: historical
Owners: []
RelatedFiles:
    - Path: ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/archive/pprof-kanban-fragment-4vm-400rps-20260515T170138Z/cpu-top.txt
      Note: CPU pprof top for degraded multi-VM Kanban fragment run
ExternalSources: []
Summary: pprof evidence for degraded serve-multi kanban-fragment workload with 4 Goja VMs at 400/s.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Multi-VM Kanban Fragment pprof Report

## Purpose

The saturation sweep showed that `kanban-fragment` with 4 VMs at 400/s is clearly degraded: throughput falls below offered rate and p95 becomes multi-second. This report records pprof evidence for that multi-VM degraded cell.

## Important harness fix

The first pprof attempt produced a CPU profile with zero samples because `01-run-multi-vm-vegeta.sh` captured `/debug/pprof/profile` after the Vegeta attack completed. The script was fixed so CPU profiling starts immediately before the measured attack and overlaps the load window.

## Command

```bash
ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/scripts/01-run-multi-vm-vegeta.sh \
  --scenario kanban-fragment \
  --vm-count 4 \
  --distribution even-hot \
  --rate 400/s \
  --duration 10s \
  --warmup-duration 3s \
  --port 19001 \
  --metrics-port 20001 \
  --out-dir ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/archive/pprof-kanban-fragment-4vm-400rps-20260515T170138Z \
  --pprof \
  --pprof-seconds 10
```

## Artifacts

```text
ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/archive/pprof-kanban-fragment-4vm-400rps-20260515T170138Z
```

Key files:

```text
cpu.pprof
cpu-top.txt
heap.pprof
heap-top.txt
allocs.pprof
allocs-top.txt
goroutine.txt
metrics-before.prom
metrics-after.prom
metrics-delta.txt
sites.yaml
targets.txt
vegeta.json
vegeta.txt
```

The raw `vegeta.bin` was removed before commit because it was about 940 MiB.

## Vegeta result

```text
Requests      [total, rate, throughput]         3999, 399.99, 214.66
Duration      [total, attack, wait]             18.629s, 9.998s, 8.632s
Latencies     [min, mean, 50, 90, 95, 99, max]  19.766ms, 4.564s, 4.725s, 7.61s, 8.066s, 8.516s, 8.632s
Bytes In      [total, mean]                     983538054, 245946.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:3999
Error Set:
```

This reproduces the saturation pattern: no HTTP errors, but achieved throughput is far below offered rate and p95 is around 8 seconds.

## CPU profile headline

The CPU profile contains 41.10 seconds of samples over a 10 second wall-clock profile, so the workload was using roughly four cores during the profile window.

Top relevant cumulative costs:

```text
github.com/go-go-golems/go-go-goja/modules/uidsl.renderNode     12.57s cumulative, 30.58%
github.com/go-go-golems/goja-site/pkg/kanbanddsl.(*Board).preciseMoveForm 11.57s cumulative, 28.15%
github.com/go-go-golems/go-go-goja/modules/uidsl.renderAttrs    10.09s cumulative, 24.55%
github.com/go-go-golems/go-go-goja/modules/uidsl.attrValue       4.04s cumulative, 9.83%
github.com/dop251/goja.(*vm).run                                3.47s cumulative, 8.44%
```

Allocation and GC are also prominent:

```text
runtime.mallocgc             9.42s cumulative, 22.92%
runtime.gcDrain              9.09s cumulative, 22.12%
runtime.mallocgcSmallScanNoHeader 7.11s cumulative, 17.30%
runtime.newobject            4.81s cumulative, 11.70%
runtime.scanSpan             4.49s cumulative, 10.92%
```

The profile is consistent with the single-VM Kanban action profile: UI DSL rendering, attribute rendering, precise move form generation, allocation, and GC dominate the expensive path.

## Heap profile headline

The in-use heap profile was dominated by HTTP connection buffering:

```text
bufio.NewReaderSize  11308.06kB, 32.86%
bufio.NewWriterSize   9764.08kB, 28.37%
runtime.mallocgc      4612.42kB, 13.40%
```

This is expected for a run with many active HTTP responses and large response bodies. The CPU profile is more useful for explaining service-time cost; the heap profile confirms that buffered response handling is visible during saturated large-response operation.

## Interpretation

The degraded multi-VM cell is not bottlenecked by `serve-multi` Host dispatch or by HTTP errors. It is dominated by per-request rendering and response work. Multiple VMs allow multiple owner loops to execute in parallel, which is why the CPU profile shows about four cores worth of samples. But each VM still performs the same large board render, and every request returns about 246 KB of HTML.

The main cost centers are:

1. `uidsl.renderNode`: recursive node rendering for the board.
2. `uidsl.renderAttrs`: attribute serialization for rendered elements.
3. `kanbanddsl.(*Board).preciseMoveForm`: repeated generation of movement controls/options.
4. Allocation and GC caused by render tree and HTML/string construction.
5. HTTP buffering and response writing for large responses.

## Recommended next step

The next optimization target should not be `serve-multi` dispatch. The evidence points to reducing the amount of HTML generated and returned per Kanban request.

Most likely improvements:

1. Avoid full board fragment rendering for high-frequency interactions.
2. Make `preciseMoveForm` lazy, cached, or optional.
3. Cache stable per-column/per-board option markup.
4. Reduce allocations in `uidsl.renderNode`, `renderAttrs`, and `attrValue`.
5. Consider smaller response formats for dynamic updates.
