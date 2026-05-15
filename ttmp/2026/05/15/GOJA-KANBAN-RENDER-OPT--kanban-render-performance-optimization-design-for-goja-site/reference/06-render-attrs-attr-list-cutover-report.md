---
Title: render-attrs-1000 attr-list cutover report
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
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/module.go
      Note: attrsFromValue constructs render-ready attrs
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/node.go
      Note: Element Attrs cut over to []Attr
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/render.go
      Note: renderAttrs now writes []Attr
    - Path: pkg/kanbanddsl/render.go
      Note: updated for uidsl.Attrs cutover
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/archive/render-attrs-attrlist-20260515T200104Z/allocs-top.txt
      Note: post-cutover allocation profile
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/archive/render-attrs-attrlist-20260515T200104Z/cpu-top.txt
      Note: post-cutover CPU profile
ExternalSources: []
Summary: Report for the no-compatibility cutover from map attrs to render-ready []Attr in ui.dsl.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# render-attrs-1000 attr-list cutover report

The user clarified that we should not preserve a compatibility map path. I cut `ui.dsl` over to a render-ready attr list representation.

## Implementation

Local `go-go-goja` commit:

```text
9083b8f3bad9f8807c5e4c3c0076c74f2c33a196
perf: render ui attrs from attr list
```

Main structural change:

```go
type Attr struct {
    Key   string
    Value string
    Bool  bool
}

type Element struct {
    Tag      string
    Attrs    []Attr
    Children []Node
}
```

`renderAttrs` now writes `[]Attr` directly and no longer allocates/sorts a key slice during rendering.

I also updated `goja-site/pkg/kanbanddsl/render.go` to construct `uidsl.Element` values with `uidsl.Attrs(map[string]any{...})` at the construction boundary.

## Microbenchmark signal

Post-cutover microbenchmarks in `go-go-goja/modules/uidsl`:

| Benchmark | Result |
|---|---:|
| `BenchmarkRenderAttrsAttrs9` | ~`0.73–0.90 us/op`, `680 B/op`, `5 allocs/op` |
| Previous `BenchmarkRenderAttrsAttrs9` | ~`2.2–2.6 us/op`, `936 B/op`, `12 allocs/op` |
| `BenchmarkRenderPageAttrs1000` | ~`12.4–14.4 ms/op`, ~`7.24 MB/op`, ~`104,892 allocs/op` |
| `BenchmarkRenderPageFlat1000` | ~`4.0 ms/op`, ~`2.18 MB/op`, ~`36,761 allocs/op` |

The render-time attr writer improved clearly. Page-level allocation did not improve much because attrs are still exported from JS object literals and converted into render-ready attr slices during node construction.

## Load-test validation

Artifact root:

```text
archive/render-attrs-attrlist-20260515T200104Z
```

Run shape:

```text
render-attrs-1000
100/s
30s measured
5s warmup
10s CPU profile
```

Result:

| Scenario | Throughput | Success | p50 | p95 | p99 | Max |
|---|---:|---:|---:|---:|---:|---:|
| `render-attrs-1000` after attr-list cutover | `71.79/s` | `100%` | `5.271s` | `11.222s` | `11.714s` | `11.799s` |

Comparison:

| Version | Throughput | p95 |
|---|---:|---:|
| Original follow-up pprof | `67.91/s` | `13.357s` |
| Single-export slice | `59.70/s` | `19.337s` |
| Attr-list cutover | `71.79/s` | `11.222s` |

The cutover helps, but `render-attrs-1000` is still not healthy at `100/s`.

## Interpretation

The no-compatibility attr-list cutover successfully removes render-time attr map iteration and sort costs. That is visible in the microbenchmark. The macro benchmark remains saturated because the bigger remaining costs occur earlier:

- JavaScript object construction for every element and attr object,
- one generic `Value.Export()` per attr object,
- Goja map/object allocation and property writes,
- creation of render-ready attr slices for each node,
- response buffering and GC.

## Next recommendation

The next useful step is to reduce JS object construction/export at the DSL boundary, not to further optimize `renderAttrs`.

Candidate directions:

1. Add specialized native constructors for common attrs patterns, such as `ui.divClass(className, child...)` or `ui.el(tag, attrsArray, child...)`.
2. Introduce a JS-side compiled node representation where attrs are stored as arrays rather than object literals.
3. Render directly from Goja values for hot paths instead of normalizing into Go nodes first.
4. Pre-size response buffers or stream response output after construction/export cost is reduced.

Any next optimization should be validated against both `render-attrs-1000` and Kanban workloads.
