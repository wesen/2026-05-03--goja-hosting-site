---
Title: render-attrs-1000 first optimization report
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
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/attrs_bench_test.go
      Note: focused ui.dsl attr/render microbenchmarks
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/attrs_compat_test.go
      Note: attribute compatibility and child disambiguation tests
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/module.go
      Note: first optimization slice
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/archive/render-attrs-single-export-20260515T194921Z/allocs-top.txt
      Note: post-slice allocation profile
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/archive/render-attrs-single-export-20260515T194921Z/cpu-top.txt
      Note: post-slice CPU profile
ExternalSources: []
Summary: Implementation and validation report for the first ui.dsl attrs optimization slice in go-go-goja.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# render-attrs-1000 first optimization report

This report records the first implementation slice for the `render-attrs-1000` investigation.

The implementation landed in the local `go-go-goja` repository:

```text
../go-go-golems/go-go-goja
commit: 2bc72d5f04f2ea3c47b2394484d523b4c26bbf1c
message: perf: avoid double export for ui attrs
```

## What changed

Files changed in `go-go-goja`:

```text
modules/uidsl/module.go
modules/uidsl/attrs_compat_test.go
modules/uidsl/attrs_bench_test.go
```

The original hot path did this for every UI element call with attrs:

```go
if len(args) > 0 && isAttrs(args[0]) {
    if m, ok := args[0].Export().(map[string]any); ok {
        attrs = m
    }
    args = args[1:]
}
```

`isAttrs` also called `v.Export()`, so attrs were exported once for classification and again for extraction.

The new path decodes classification and extraction together:

```go
if len(args) > 0 {
    if decoded, ok := attrsFromValue(vm, args[0]); ok {
        attrs = decoded
        args = args[1:]
    }
}
```

and:

```go
func attrsFromValue(_ *goja.Runtime, v goja.Value) (map[string]any, bool) {
    if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
        return nil, false
    }
    switch exported := v.Export().(type) {
    case Node, string, []any, []Node, int, int64, float64, bool:
        return nil, false
    case map[string]any:
        return exported, true
    default:
        return nil, false
    }
}
```

A small allocation improvement was also made by using `var attrs map[string]any` instead of allocating an empty map for every element.

## Tests added

Added compatibility coverage for:

- string attrs,
- numeric attrs,
- boolean attrs,
- null and undefined omission,
- empty `value=""` preservation,
- class arrays,
- style maps,
- `aria-*` and `data-*`,
- child-vs-attrs disambiguation for text, numbers, booleans, nodes, arrays, fragments, empty attrs, and normal attrs.

This guards the JS-facing DSL behavior before deeper performance changes.

## Microbenchmarks added

Added:

```text
BenchmarkElementFromCallAttrs9
BenchmarkRenderAttrsAttrs9
BenchmarkRenderPageAttrs1000
BenchmarkRenderPageFlat1000
```

Run command:

```text
go test ./modules/uidsl -bench='Benchmark(ElementFromCallAttrs9|RenderPageAttrs1000|RenderPageFlat1000|RenderAttrsAttrs9)' -benchmem -run '^$' -count=3
```

Representative post-change results:

| Benchmark | Result |
|---|---:|
| `BenchmarkElementFromCallAttrs9` | ~`2.9–4.4 us/op`, `2160–2208 B/op`, `38–39 allocs/op` |
| `BenchmarkRenderAttrsAttrs9` | ~`2.2–2.6 us/op`, `936 B/op`, `12 allocs/op` |
| `BenchmarkRenderPageAttrs1000` | ~`13.6–18.3 ms/op`, ~`6.8 MB/op`, ~`102k–104k allocs/op` |
| `BenchmarkRenderPageFlat1000` | ~`3.5–5.3 ms/op`, ~`2.1 MB/op`, ~`35k–36k allocs/op` |

The most useful microbenchmark signal is that `ElementFromCallAttrs9` now has fewer allocations than the direct-object-decoder experiment and fewer than the pre-change double-export path. The page-level benchmark still allocates heavily, so this is only the first slice.

## Validation

In `go-go-goja`:

```text
go test ./modules/uidsl
go test ./...
```

The pre-commit hook also ran:

```text
golangci-lint run -v
go generate ./...
go test ./...
```

All passed.

## Load-test result after first slice

A follow-up `goja-site` pprof run used the local `go-go-goja` replace and the new `render-attrs-1000` implementation.

Artifact root:

```text
archive/render-attrs-single-export-20260515T194921Z
```

Command shape:

```text
scripts/bench-vegeta.sh \
  --scenario render-attrs-1000 \
  --duration 30s \
  --warmup-duration 5s \
  --rate 100/s \
  --pprof \
  --pprof-seconds 10
```

Result:

| Scenario | Rate | Requests | Throughput | Success | p50 | p95 | p99 | Max |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `render-attrs-1000` after single-export slice | `100/s` | 3000 | `59.70/s` | `100%` | `9.853s` | `19.337s` | `20.114s` | `20.257s` |

This did **not** improve the end-to-end load result. It reduced one avoidable conversion pattern, but the request path still saturates badly.

The allocation profile changed shape compared with the previous follow-up pprof:

| Before first slice | After first slice |
|---|---|
| ~`31.49 GB` total alloc-space profile | ~`23.27 GB` total alloc-space profile |
| `elementFromCall` cumulative ~`16.7 GB` | `elementFromCall` cumulative ~`8.8 GB` |
| `Object.Export` cumulative ~`15.7 GB` | `Object.Export` cumulative ~`7.8 GB` |

So the first slice achieved the intended allocation reduction, but that reduction is not enough to move the macro benchmark. The system is still dominated by JS object construction, generic export, rendering, response buffering, and GC.

## Rejected experiment during implementation

I also tried a direct `goja.Object.Keys()` + property-read decoder for attrs. It avoided full recursive object export, but it was worse in practice for this workload: `Object.Keys`, property iteration, and per-property conversion became visible and the macro run regressed. That experiment was not kept.

The kept implementation is intentionally conservative: it removes the duplicated export while preserving existing behavior.

## Current diagnosis after implementation

The first slice confirms the investigation: duplicated attr export was real and wasteful, but it was only part of the problem.

Remaining dominant costs are:

- one full generic `Object.Export` per attrs object,
- JS object construction for every node and attrs object,
- `goja.(*baseObject)._put` while building those objects,
- `renderAttrs` key sorting and value conversion,
- output buffer growth and string materialization,
- GC over all of the temporary structures.

## Recommended next slice

Do not keep shaving around `attrsFromValue`. The next meaningful optimization is a render-ready attr representation or a lower-allocation UI node construction path.

Recommended next design:

```text
Introduce a compact AttrList representation for Element attrs.
```

Possible shape:

```go
type Attr struct {
    Key   string
    Value string
    Bool  bool
}

type Element struct {
    Tag      string
    Attrs    map[string]any // compatibility path
    AttrList []Attr         // fast path
    Children []Node
}
```

Then the UI DSL native constructors can decode attrs once into sorted, render-ready `[]Attr`, and `renderAttrs` can avoid:

- map iteration,
- per-render key-slice allocation,
- per-render sorting,
- repeated generic `fmt.Sprint` conversion for common primitive attrs.

Acceptance criteria for the next slice:

```text
render-attrs-1000 100/s throughput ratio improves materially
alloc-space profile drops below the current ~23 GB run
renderAttrs no longer allocates a key slice per element on the fast path
compatibility tests continue to pass
kanban-fragment and kanban-fragment-500 do not regress
```
