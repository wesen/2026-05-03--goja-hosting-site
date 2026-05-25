---
Title: Anti-overfit follow-up pprof report
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - benchmarking
    - performance
    - profiling
    - optimization
DocType: reference
Intent: historical
Owners: []
RelatedFiles:
    - Path: bench/scripts/kanban-board-500/app.js
      Note: 500-card Kanban workload used for pprof
    - Path: bench/scripts/render-shapes/app.js
      Note: attribute-heavy UI render workload used for pprof
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/archive/pprof-anti-overfit-20260515T191339Z/kanban-fragment-500-100_s/allocs-top.txt
      Note: allocation top for 500-card Kanban render
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/archive/pprof-anti-overfit-20260515T191339Z/kanban-fragment-500-100_s/cpu-top.txt
      Note: CPU top for 500-card Kanban render
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/allocs-top.txt
      Note: allocation top for attribute-heavy render
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/cpu-top.txt
      Note: CPU top for attribute-heavy render
ExternalSources: []
Summary: Follow-up pprof comparison for render-attrs-1000 at 100/s and kanban-fragment-500 at 100/s after Anti-overfit matrix v1.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Anti-overfit follow-up pprof report

After Anti-overfit matrix v1, two cells deserved profiling:

1. `render-attrs-1000` at `100/s`, because the attribute-heavy generic UI DSL shape saturated before the other render shapes.
2. `kanban-fragment-500` at `100/s`, because the 500-card Kanban scaling fixture saturated in the short matrix and represents the large-board version of the original optimization target.

Both profiles were captured in the same pass so we can compare generic UI rendering against large Kanban rendering instead of choosing one prematurely.

## Run metadata

| Field | Value |
|---|---|
| Artifact root | `archive/pprof-anti-overfit-20260515T191339Z` |
| Duration | `30s` measured |
| Warmup | `5s` |
| CPU profile window | `10s` during load |
| Scenarios | `render-attrs-1000`, `kanban-fragment-500` |
| Rate | `100/s` |
| Raw Vegeta binaries | removed after text/json/profile artifacts were generated |

## Vegeta result comparison

| Scenario | Rate | Requests | Throughput | Success | p50 | p95 | p99 | Max | Avg bytes in |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `render-attrs-1000` | `100/s` | 3000 | `67.91/s` | `100%` | `5.999s` | `13.357s` | `13.929s` | `14.186s` | 250,040 |
| `kanban-fragment-500` | `100/s` | 3000 | `100.01/s` | `100%` | `17.385ms` | `66.472ms` | `82.266ms` | `107.746ms` | 251,460 |

The important surprise is that `kanban-fragment-500` did not reproduce the severe queueing from the short anti-overfit matrix in this longer pprof run. It held the target rate with low p95. That suggests the earlier `kanban-fragment-500` 100/s result may have been noisy or sensitive to run ordering, GC state, CPU contention, or neighboring benchmark cells. The generic `render-attrs-1000` result reproduced a clear bottleneck and is therefore the stronger immediate optimization target.

## CPU profile: render-attrs-1000 at 100/s

Artifact paths:

```text
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/cpu.pprof
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/cpu-top.txt
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/allocs-top.txt
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/heap-top.txt
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/goroutine.txt
```

Top CPU observations:

| Function / family | Signal |
|---|---|
| `github.com/dop251/goja.(*vm).run` | `9.98s` cumulative, `61.08%` of CPU samples |
| `runtime.gcDrain` | `5.29s` cumulative, `32.37%` |
| `runtime.mallocgc` | `3.63s` cumulative, `22.22%` |
| `github.com/dop251/goja.(*baseObject).export` | `3.57s` cumulative, `21.85%` |
| `github.com/go-go-golems/go-go-goja/modules/uidsl.renderNode` | `2.02s` cumulative, `12.36%` |
| `github.com/go-go-golems/go-go-goja/modules/uidsl.renderAttrs` | `1.83s` cumulative, `11.20%` |
| `github.com/dop251/goja.(*baseObject).stringKeys` | `1.45s` cumulative, `8.87%` |

The CPU profile is not a single narrow hotspot. It is an allocation/export/GC profile. The route constructs many JS UI objects with many attributes. Go then repeatedly exports Goja objects into Go-side UI DSL nodes and renders their attribute maps. That makes the cost show up across Goja object export, string key enumeration, map access/assignment, rendering, allocation, and GC.

## Allocation profile: render-attrs-1000 at 100/s

The allocation profile is more decisive than the CPU top list:

| Function / family | Allocated |
|---|---:|
| `github.com/dop251/goja.(*baseObject).export` | `15511.81 MB` cumulative |
| `github.com/go-go-golems/go-go-goja/modules/uidsl.elementFromCall` | `16718.55 MB` cumulative |
| `github.com/dop251/goja.(*Object).Export` | `15705.29 MB` cumulative |
| `github.com/go-go-golems/go-go-goja/modules/uidsl.renderAttrs` | `2564.39 MB` cumulative |
| `bytes.growSlice` | `1722.23 MB` flat |
| `gojahttp.(*Response).writeString` | `849.58 MB` cumulative |

Interpretation:

- The generic UI DSL path spends heavily converting Goja object values to Go values.
- Attribute-heavy nodes amplify `Object.Export`, `baseObject.export`, `stringKeys`, and `renderAttrs`.
- Response serialization is visible, but not the whole story. The object construction/export path dominates allocations.
- A good next optimization should avoid generic `Object.Export` for known UI DSL node and attribute shapes.

## CPU profile: kanban-fragment-500 at 100/s

Artifact paths:

```text
archive/pprof-anti-overfit-20260515T191339Z/kanban-fragment-500-100_s/cpu.pprof
archive/pprof-anti-overfit-20260515T191339Z/kanban-fragment-500-100_s/cpu-top.txt
archive/pprof-anti-overfit-20260515T191339Z/kanban-fragment-500-100_s/allocs-top.txt
archive/pprof-anti-overfit-20260515T191339Z/kanban-fragment-500-100_s/heap-top.txt
archive/pprof-anti-overfit-20260515T191339Z/kanban-fragment-500-100_s/goroutine.txt
```

Top CPU observations:

| Function / family | Signal |
|---|---|
| `github.com/dop251/goja.(*vm).run` | `3.47s` cumulative, `25.31%` |
| `runtime.gcDrain` | `4.03s` cumulative, `29.39%` |
| `runtime.mallocgc` | `2.89s` cumulative, `21.08%` |
| `github.com/go-go-golems/go-go-goja/modules/uidsl.renderNode` | `2.19s` cumulative, `15.97%` |
| `github.com/go-go-golems/go-go-goja/modules/uidsl.renderAttrs` | `1.95s` cumulative, `14.22%` |
| `kanbanddsl.(*Board).loadCards` | `1.60s` cumulative, `11.67%` |
| `kanbanddsl.(*Board).renderCard` | `4.30s` cumulative, `31.36%` |

The large Kanban profile is now a normal large-render profile, not the old `preciseMoveForm` profile. The previous pathological server-rendered movement-form hotspot is gone. Remaining cost is distributed across:

- calling the JS card renderer for 500 cards,
- loading/exporting card data,
- rendering UI DSL nodes and attributes,
- output buffer allocation,
- GC.

## Allocation profile: kanban-fragment-500 at 100/s

| Function / family | Allocated |
|---|---:|
| `kanbanddsl.(*Board).Render` | `12184.67 MB` cumulative |
| `kanbanddsl.(*Board).renderColumn` | `10197.25 MB` cumulative |
| `kanbanddsl.(*Board).renderCard` | `10082.97 MB` cumulative |
| `uidsl.elementFromCall` | `2936.94 MB` cumulative |
| `uidsl.renderAttrs` | `2303.27 MB` cumulative |
| `bytes.growSlice` | `1767.51 MB` flat |
| `gojahttp.(*Response).writeString` | `838.57 MB` cumulative |
| `kanbanddsl.(*Board).cardActionsButton` | `787.15 MB` flat |
| `kanbanddsl.(*Board).loadCards` | `1480.42 MB` cumulative |
| `kanbanddsl.callString` | `679.59 MB` cumulative |

Interpretation:

- The remaining Kanban cost is proportional full-board rendering work.
- `cardActionsButton` is visible but no longer pathological; it is a compact per-card control, not a per-card matrix of movement forms.
- The next Kanban-specific optimization would not be another movement-form removal. It would be either partial/semantic action responses or reducing per-card render/export overhead.

## Comparison and recommendation

| Candidate | Evidence strength | Why |
|---|---|---|
| `render-attrs-1000` generic UI attr/export optimization | Strong | Reproduced severe queueing at `100/s`; allocation profile clearly points to Goja object export, attribute enumeration, and UI DSL object conversion. |
| `kanban-fragment-500` large-board optimization | Medium | The short matrix showed queueing at `100/s`, but this pprof run was healthy. Still useful as a size-scaling regression, but less urgent as the next optimization target. |
| Semantic action response protocol | Still valuable | It reduces full refresh cost for interactions, but this pprof pair is about fragment rendering, not action response size. |

Recommended next implementation target:

```text
Optimize generic UI DSL object/attribute conversion before doing another Kanban-specific render change.
```

Concrete design direction:

1. Inspect `go-go-goja/modules/uidsl` for `elementFromCall`, `NormalizeExport`, and `renderAttrs`.
2. Avoid `Object.Export` for common UI element calls when possible.
3. Decode attrs from Goja objects directly with bounded, typed accessors instead of generic export-to-map conversion.
4. Consider a compact internal node representation with pre-normalized attributes.
5. Benchmark against both `render-attrs-1000` and `kanban-fragment-500` so the fix helps generic UI and does not regress Kanban.

## Validation commands

```text
scripts/bench-vegeta.sh --scenario render-attrs-1000 --duration 30s --warmup-duration 5s --rate 100/s --pprof --pprof-seconds 10
scripts/bench-vegeta.sh --scenario kanban-fragment-500 --duration 30s --warmup-duration 5s --rate 100/s --pprof --pprof-seconds 10
go tool pprof -top cpu.pprof
go tool pprof -top allocs.pprof
```
