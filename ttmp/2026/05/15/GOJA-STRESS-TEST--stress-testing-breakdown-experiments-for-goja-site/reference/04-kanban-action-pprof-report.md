---
Title: Kanban Action 80rps pprof Report
Ticket: GOJA-STRESS-TEST
Status: active
Topics:
    - benchmarking
    - stress-testing
    - pprof
    - kanban
DocType: reference
Intent: historical
Owners: []
RelatedFiles:
    - Path: ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/archive/pprof-kanban-action-80rps-20260515T151521Z/cpu-top.txt
      Note: CPU pprof top output for kanban-action at 80/s
ExternalSources: []
Summary: Targeted pprof report for kanban-action at 80/s after knee search identified 80/s as the first clear latency-threshold breach.
LastUpdated: 2026-05-15T15:20:00Z
WhatFor: Use this to understand the first CPU/allocation evidence for the kanban-action stress knee.
WhenToUse: After reviewing the quick stress and kanban-action knee reports.
---


# Kanban Action 80rps pprof Report

## Purpose

The quick stress sweep found that `kanban-action` was the first scenario to bend. The targeted knee search then narrowed the first clear latency-threshold breach to around `80/s`. This report records the first pprof diagnostic run at that rate.

## Command

```bash
scripts/bench-vegeta.sh \
  --scenario kanban-action \
  --duration 30s \
  --warmup-duration 5s \
  --rate 80/s \
  --port 18700 \
  --metrics-port 19700 \
  --out-dir ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/archive/pprof-kanban-action-80rps-20260515T151521Z \
  --pprof \
  --pprof-seconds 20
```

## Artifacts

Stored under:

```text
ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/archive/pprof-kanban-action-80rps-20260515T151521Z
```

Important files:

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
server.log
summary.txt
vegeta.json
vegeta.txt
```

The raw `vegeta.bin` file was intentionally removed before committing because it was about 891 MiB due to large Kanban response bodies.

## Vegeta result

```text
Requests      [total, rate, throughput]         2400, 80.03, 80.00
Duration      [total, attack, wait]             29.999s, 29.988s, 11.411ms
Latencies     [min, mean, 50, 90, 95, 99, max]  4.75ms, 36.297ms, 14.412ms, 138.113ms, 178.339ms, 221.485ms, 251.72ms
Bytes In      [total, mean]                     932940000, 388725.00
Bytes Out     [total, mean]                     120000, 50.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:2400
Error Set:
```

This particular diagnostic run was healthier than the worst 80/s knee-search repeats, but it still shows elevated tail latency relative to 50-70/s.

## CPU profile headline

The CPU profile suggests the heavy path is dominated by rendering and allocation/GC work, especially the UI DSL render path and the Kanban precise move form generation.

Top relevant entries from `cpu-top.txt`:

```text
flat    flat%   cum     cum%   function
1.17s   4.53%   2.06s   7.98%  encoding/json.appendString
0.36s   1.39%   5.79s  22.42%  github.com/go-go-golems/go-go-goja/modules/uidsl.renderNode
0.34s   1.32%   4.62s  17.89%  github.com/go-go-golems/go-go-goja/modules/uidsl.renderAttrs
0.19s   0.74%   5.27s  20.41%  github.com/go-go-golems/goja-site/pkg/kanbanddsl.(*Board).preciseMoveForm
```

GC/allocation pressure is visible in CPU as well:

```text
runtime.gcDrain                  6.85s cumulative, 26.53%
runtime.mallocgc                 5.03s cumulative, 19.48%
runtime.mallocgcSmallScanNoHeader 3.53s cumulative, 13.67%
runtime.scanSpan                 3.10s cumulative, 12.01%
runtime.scanObject               2.58s cumulative, 9.99%
```

## Heap profile headline

The heap profile is less conclusive because in-use heap is small and includes startup/runtime state. Still, it shows JSON encoding and Goja/runtime allocations:

```text
encoding/json.appendString  1801.08kB  19.48%
runtime.mallocgc            1538.25kB  16.64%
encoding/json Encoder paths 2528.86kB cumulative, 27.35%
```

## Interpretation

The first diagnostic evidence supports the hypothesis from the stress reports:

```text
kanban-action is expensive because every action refresh renders a large 120-card board and returns it as JSON-wrapped HTML.
```

The specific hotspots indicate:

1. UI DSL rendering is a major CPU path.
2. Rendering attributes and nodes allocates heavily enough to involve GC noticeably.
3. `preciseMoveForm` is expensive because every rendered card includes form controls/options for precise moves.
4. JSON encoding of the large refreshed HTML response is visible.

## Likely optimization directions

Potential follow-ups, in order of likely impact:

1. Avoid full-board HTML refresh for every action; return smaller patches or affected columns/cards.
2. Disable or lazily render `preciseMoveForm` for benchmark/production modes where drag/drop is primary.
3. Cache static column/option markup for move forms.
4. Reduce allocation in UI DSL rendering, especially `renderNode` and `renderAttrs`.
5. Reuse buffers in HTML rendering and JSON response paths.

## Caution

This is one pprof run at 80/s. The knee-search matrix showed high variance around 80-100/s, so this should guide optimization but not be treated as a final root-cause proof. The next best profiling run is to capture CPU at the first consistently degraded rate, likely 90/s or 100/s, after deciding how much tail latency is acceptable.
