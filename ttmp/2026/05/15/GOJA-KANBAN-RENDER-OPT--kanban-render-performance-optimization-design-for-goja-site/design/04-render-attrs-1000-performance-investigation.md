---
Title: render-attrs-1000 performance investigation
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - benchmarking
    - performance
    - profiling
    - goja
    - optimization
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/module.go
      Note: elementFromCall
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/node.go
      Note: Node and Element representation
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/render.go
      Note: renderNode and renderAttrs source
    - Path: bench/scripts/render-shapes/app.js
      Note: render-attrs-1000 benchmark fixture under investigation
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/allocs-top.txt
      Note: allocation profile evidence
    - Path: ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/cpu-top.txt
      Note: CPU profile evidence
ExternalSources: []
Summary: Detailed investigation of why render-attrs-1000 saturates, with source-level root causes and optimization plan for ui.dsl object and attribute conversion.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# render-attrs-1000 performance investigation

`render-attrs-1000` is the first anti-overfit benchmark that reproduced a severe, generic UI-rendering bottleneck after the Kanban-specific `preciseMoveForm` problem was removed. It matters because it is not a Kanban-only workload. It is a synthetic but plausible hosted-site shape: a page with many ordinary elements, many ordinary attributes, and a large server-rendered HTML response.

The high-level conclusion is:

```text
render-attrs-1000 is dominated by Goja object export, UI DSL element construction, attribute map conversion, allocation, and GC.
```

The next optimization should focus on `go-go-goja/modules/uidsl`, especially the path from JavaScript calls such as:

```javascript
ui.div({ class: "...", id: "...", "data-index": "..." }, child)
```

to Go values such as:

```go
&Element{Tag: "div", Attrs: map[string]any{...}, Children: []Node{...}}
```

## 1. Benchmark definition

The benchmark fixture is:

```text
bench/scripts/render-shapes/app.js
```

The relevant route is:

```javascript
app.get("/attrs", (req, res) => {
  res.html(attrPage(Number(req.query.n || 1000)));
});
```

The page builder creates 1000 sibling nodes. Each node has nine attributes and one child span:

```javascript
children.push(ui.div({
  class: "attr-node state-" + (i % 5),
  id: "attr-node-" + i,
  "data-index": String(i),
  "data-kind": "benchmark",
  "data-group": String(i % 17),
  "data-label": "attribute heavy node " + i,
  "aria-label": "Attribute heavy node " + i,
  role: "listitem",
  tabindex: "0"
}, ui.span({ class: "label" }, "node-" + i)));
```

So one response includes approximately:

- 1000 outer `div` elements,
- 1000 inner `span` elements,
- at least 10,000 attribute key/value pairs after including the section/list wrapper and span class attributes,
- a response body around 250 KB.

The benchmark is intentionally not exotic. It is a stress test for common HTML generation patterns: list-like output, ARIA attributes, data attributes, ids, classes, and text.

## 2. Observed benchmark behavior

### 2.1 Anti-overfit matrix result

From `reference/03-anti-overfit-benchmark-report.md`:

| Scenario | Rate | Success avg | Throughput ratio avg | p95 avg |
|---|---:|---:|---:|---:|
| `render-attrs-1000` | `25/s` | `1.00` | `1.00` | `37.33 ms` |
| `render-attrs-1000` | `100/s` | `1.00` | `0.51` | `13.842 s` |
| `render-attrs-1000` | `250/s` | `0.71` | `0.24` | `30.001 s` |

This shows a sharp knee between `25/s` and `100/s`. At `100/s`, the server still returns HTTP 200 for all requests, but it cannot keep up. Requests queue behind the single site/VM event loop and p95 grows into seconds. At `250/s`, some requests hit client timeouts.

### 2.2 Follow-up pprof result

From `reference/04-anti-overfit-follow-up-pprof-report.md`:

| Scenario | Rate | Requests | Throughput | Success | p50 | p95 | p99 | Max | Avg bytes in |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `render-attrs-1000` | `100/s` | 3000 | `67.91/s` | `100%` | `5.999s` | `13.357s` | `13.929s` | `14.186s` | 250,040 |

This reproduces the `100/s` saturation outside the original 63-run matrix. The bottleneck is stable enough to treat as real.

Artifacts:

```text
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/cpu.pprof
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/cpu-top.txt
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/allocs.pprof
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/allocs-top.txt
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/heap.pprof
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s/heap-top.txt
```

## 3. Source path under test

The implementation lives in the local `go-go-goja` repository:

```text
../go-go-golems/go-go-goja/modules/uidsl/module.go
../go-go-golems/go-go-goja/modules/uidsl/render.go
../go-go-golems/go-go-goja/modules/uidsl/node.go
```

The current pipeline is:

```text
JavaScript route
  -> ui.div(attrs, children...)
  -> Go native function registered in Loader
  -> elementFromCall(tag, call)
  -> isAttrs(args[0])
  -> args[0].Export()
  -> map[string]any attrs
  -> nodesFromArgs(children)
  -> child.Export()
  -> NormalizeExport
  -> Element tree
  -> res.html(...)
  -> RenderAny / renderNode / renderAttrs
  -> bytes.Buffer.String
  -> HTTP response
```

The most important code is in `module.go`:

```go
func Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)
    for _, tag := range tags {
        tag := tag
        _ = exports.Set(tag, func(call goja.FunctionCall) goja.Value { return vm.ToValue(elementFromCall(tag, call)) })
    }
    // ...
}

func elementFromCall(tag string, call goja.FunctionCall) *Element {
    attrs := map[string]any{}
    args := call.Arguments
    if len(args) > 0 && isAttrs(args[0]) {
        if m, ok := args[0].Export().(map[string]any); ok {
            attrs = m
        }
        args = args[1:]
    }
    return &Element{Tag: tag, Attrs: attrs, Children: nodesFromArgs(args)}
}

func nodesFromArgs(args []goja.Value) []Node {
    var out []Node
    for _, arg := range args {
        if arg == nil || goja.IsUndefined(arg) || goja.IsNull(arg) {
            continue
        }
        n, err := NormalizeExport(arg.Export())
        // ...
    }
    return out
}

func isAttrs(v goja.Value) bool {
    if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
        return false
    }
    switch v.Export().(type) {
    case Node, string, []any, []Node, int, int64, float64, bool:
        return false
    case map[string]any:
        return true
    default:
        return false
    }
}
```

The renderer is in `render.go`:

```go
func renderAttrs(b *bytes.Buffer, attrs map[string]any) {
    if len(attrs) == 0 {
        return
    }
    keys := make([]string, 0, len(attrs))
    for k := range attrs {
        if k != "" {
            keys = append(keys, k)
        }
    }
    sort.Strings(keys)
    for _, k := range keys {
        v := attrs[k]
        // ...
        value := attrValue(k, v)
        // ...
        b.WriteString(html.EscapeString(value))
    }
}
```

## 4. CPU profile interpretation

The CPU profile for `render-attrs-1000` at `100/s` reports:

| Function / family | Signal |
|---|---:|
| `github.com/dop251/goja.(*vm).run` | `9.98s` cumulative, `61.08%` |
| `runtime.gcDrain` | `5.29s` cumulative, `32.37%` |
| `runtime.mallocgc` | `3.63s` cumulative, `22.22%` |
| `github.com/dop251/goja.(*baseObject).export` | `3.57s` cumulative, `21.85%` |
| `uidsl.renderNode` | `2.02s` cumulative, `12.36%` |
| `uidsl.renderAttrs` | `1.83s` cumulative, `11.20%` |
| `goja.(*baseObject).stringKeys` | `1.45s` cumulative, `8.87%` |

This is not a single-function problem. The bottleneck is spread across the cost of:

1. running the JS page builder,
2. converting JS objects into Go values,
3. rendering the resulting Go node tree,
4. allocating intermediate data structures,
5. garbage-collecting those structures.

The source-level pprof listing is more specific.

### 4.1 `elementFromCall`

`go tool pprof -list elementFromCall` shows:

```text
ROUTINE github.com/go-go-golems/go-go-goja/modules/uidsl.elementFromCall
30ms flat, 4.02s cumulative, 24.60% of total
```

Important lines:

```text
line 75: if len(args) > 0 && isAttrs(args[0])       -> 2.07s cumulative
line 76: if m, ok := args[0].Export().(...)         -> 1.62s cumulative
line 81: return &Element{... Children: nodesFromArgs(args)} -> 250ms cumulative
```

This means the first argument handling dominates `elementFromCall`. It is doing expensive conversion before we even render anything.

### 4.2 `isAttrs`

`go tool pprof -list isAttrs` shows:

```text
ROUTINE github.com/go-go-golems/go-go-goja/modules/uidsl.isAttrs
10ms flat, 2.07s cumulative, 12.67% of total
```

The hot line is:

```go
switch v.Export().(type) {
```

This is a structural problem. `isAttrs` calls `Export()` only to classify the first argument. Then `elementFromCall` calls `Export()` again to actually obtain the map. For attribute objects, the code pays a full export cost twice.

### 4.3 `renderAttrs`

`go tool pprof -list renderAttrs` shows:

```text
ROUTINE github.com/go-go-golems/go-goja/modules/uidsl.renderAttrs
130ms flat, 1.83s cumulative, 11.20% of total
```

Important lines:

```text
line 132: keys := make([]string, 0, len(attrs)) -> 70ms cumulative and 496MB allocs
line 138: sort.Strings(keys)                    -> 170ms cumulative
line 151: value := attrValue(k, v)              -> 610ms cumulative
line 158: html.EscapeString(value)              -> 200ms cumulative
```

`renderAttrs` is not the largest total cost, but it is still important because it runs for every element. It allocates a key slice per element to provide deterministic attribute order, converts every value through `fmt.Sprint`-style logic, and escapes every string.

## 5. Allocation profile interpretation

The allocation profile is the clearest evidence.

| Function / family | Allocated |
|---|---:|
| `goja.(*baseObject).export` | `15511.81 MB` cumulative |
| `uidsl.elementFromCall` | `16718.55 MB` cumulative |
| `goja.(*Object).Export` | `15705.29 MB` cumulative |
| `uidsl.renderAttrs` | `2564.39 MB` cumulative |
| `bytes.growSlice` | `1722.23 MB` flat |
| `gojahttp.(*Response).writeString` | `849.58 MB` cumulative |

The source-level allocation listing for `elementFromCall` shows:

```text
line 73: attrs := map[string]any{}              -> 317.51 MB flat
line 75: isAttrs(args[0])                       -> 7.63 GB cumulative
line 76: args[0].Export()                       -> 7.62 GB cumulative
line 81: &Element{... nodesFromArgs(args)}      -> 787.23 MB cumulative
```

This means a large amount of allocation is not required by the final HTML response. It is temporary conversion work.

The source-level allocation listing for `renderAttrs` shows:

```text
line 132: make([]string, 0, len(attrs))         -> 496.57 MB flat
line 151: attrValue(k, v)                       -> 405.51 MB cumulative
line 156: b.WriteString(k)                      -> 1003.16 MB cumulative
line 158: html.EscapeString(value)              -> 212.73 MB cumulative
line 159: b.WriteByte('"')                      -> 444.41 MB cumulative
```

Some bytes.Buffer growth attribution appears on writes because writing the response expands the buffer. That is real cost, but it is downstream of having already built a large element tree.

## 6. Root-cause model

The root cause has five layers.

### 6.1 Attribute detection performs full export

Current `isAttrs` checks whether the first argument is an attrs object by doing:

```go
switch v.Export().(type) {
```

For ordinary JS object literals, `Export()` recursively enumerates properties and builds Go maps. That is exactly the expensive work we are trying to avoid. In this benchmark, it happens for every `ui.div({ ... })` and every `ui.span({ ... })`.

### 6.2 Attribute extraction performs full export again

After `isAttrs` returns true, `elementFromCall` does:

```go
if m, ok := args[0].Export().(map[string]any); ok {
    attrs = m
}
```

So the attribute object is exported twice:

1. once for classification,
2. once for extraction.

This is visible in pprof as roughly equal cumulative allocation under `isAttrs` and line 76.

### 6.3 Generic `Export()` is too general for UI DSL attrs

`Export()` must handle arbitrary JS values. That includes arrays, objects, accessors, prototypes, nested structures, and conversion semantics. UI attributes are much more constrained:

```text
string
number
boolean
null/undefined
class array
class map
style map
```

The UI DSL is using a fully general conversion mechanism for a hot, predictable shape.

### 6.4 Attribute rendering repeats per-element sorting and conversion

`renderAttrs` sorts keys on every element:

```go
keys := make([]string, 0, len(attrs))
sort.Strings(keys)
```

Sorting provides deterministic output, which is valuable for tests and stable HTML. But it allocates and sorts on every render. In `render-attrs-1000`, this happens at least 2000 times per request because each `div` contains a `span` with attrs.

`attrValue` also converts values generically:

```go
return fmt.Sprint(v)
```

For this benchmark most values are already strings. We can take cheaper paths for common primitive types.

### 6.5 The whole response is buffered as a string

The output path renders into a `bytes.Buffer`, then calls `b.String()`, then `gojahttp.Response.writeString` writes the string. The response is ~250 KB. This is visible in allocation profiles as `bytes.growSlice`, `bytes.Buffer.String`, and `gojahttp.Response.writeString`.

This is not the first optimization target, because the object export cost is larger, but streaming or pre-sizing could become important after conversion is optimized.

## 7. Why this is different from the old Kanban bottleneck

The old Kanban issue was dominated by `preciseMoveForm`: the server rendered many movement forms for every card. The fix removed an unnecessary feature shape.

`render-attrs-1000` is different. The benchmark is doing useful work: rendering many attributed elements. We cannot solve it by deleting the content. We need to make the common path cheaper.

The important distinction:

| Bottleneck | Cause | Fix class |
|---|---|---|
| Old Kanban `preciseMoveForm` | unnecessary eager markup | remove/simplify DSL feature |
| `render-attrs-1000` | expensive generic object conversion and attr rendering | optimize UI DSL internals |

## 8. Optimization plan

### 8.1 Phase 1: Stop double-exporting attrs

Replace the current two-step logic:

```go
if len(args) > 0 && isAttrs(args[0]) {
    if m, ok := args[0].Export().(map[string]any); ok {
        attrs = m
    }
    args = args[1:]
}
```

with one function that classifies and decodes at the same time:

```go
attrs, ok := attrsFromValue(vm, args[0])
if ok {
    args = args[1:]
}
```

Expected benefit:

- removes one full `Export()` per element with attributes,
- should immediately reduce the two ~7.6 GB allocation branches in `elementFromCall`.

Risk:

- Must preserve current behavior for children that are nodes, arrays, strings, numbers, booleans, and fragments.

### 8.2 Phase 2: Decode attrs directly from `goja.Object`

Use direct object access instead of generic `Export()`:

```go
func attrsFromValue(vm *goja.Runtime, v goja.Value) (map[string]any, bool) {
    if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
        return nil, false
    }
    if _, ok := v.Export().(Node); ok { // or better: detect wrapped Node without full object export
        return nil, false
    }
    obj := v.ToObject(vm)
    keys := obj.Keys()
    if len(keys) == 0 {
        return map[string]any{}, true
    }
    attrs := make(map[string]any, len(keys))
    for _, key := range keys {
        attrs[key] = attrGoValue(vm, obj.Get(key))
    }
    return attrs, true
}
```

The important part is `attrGoValue`: it should avoid arbitrary recursive export for simple primitive values.

```go
func attrGoValue(vm *goja.Runtime, v goja.Value) any {
    if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
        return nil
    }
    if b, ok := v.Export().(bool); ok { return b }
    if s, ok := v.Export().(string); ok { return s }
    // for numbers: use ToString or Export integer/float directly
    // for arrays and maps used by class/style: decode boundedly
    return v.String()
}
```

This sketch still calls `Export()` in places, but much less often and on primitives. A better final version should use Goja type predicates and `ToString()`/`ToBoolean()` where possible.

Expected benefit:

- less `goja.(*baseObject).export`,
- less `goja.(*Object).Export`,
- less `goja.(*baseObject).stringKeys`,
- less allocation pressure and GC.

Risk:

- Need to preserve class arrays and class/style object behavior.
- Need to avoid accidentally treating Go-wrapped UI nodes as attrs.
- Need tests for attr edge cases.

### 8.3 Phase 3: Pre-normalize attrs into render-ready representation

Current `Element.Attrs` is:

```go
map[string]any
```

That is flexible but expensive. A faster representation could be:

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

Then `elementFromCall` can decode/sort attrs once at construction time, and `renderAttrs` can write them without allocating a key slice on every render.

Expected benefit:

- eliminates per-render `keys := make([]string, ...)`,
- avoids repeated `sort.Strings(keys)` if attrs are already sorted,
- avoids repeated generic `fmt.Sprint` in `attrValue`,
- makes render path more predictable.

Risk:

- Larger API change because other code may construct `&Element{Attrs: map[string]any{...}}` directly.
- Migration can be done compatibly by adding a second field:

```go
type Element struct {
    Tag         string
    Attrs       map[string]any
    AttrList    []Attr
    Children    []Node
}
```

`renderAttrs` could prefer `AttrList` when present.

### 8.4 Phase 4: Pre-size output buffers or stream response

`RenderAny` currently uses:

```go
var b bytes.Buffer
renderNode(&b, n)
return b.String(), nil
```

After object conversion is cheaper, buffer growth and string copying may become larger relative costs. Options:

1. Estimate HTML size and call `b.Grow(size)`.
2. Render directly to an `io.Writer` in the HTTP response path.
3. Allow `res.html(node)` to bypass string materialization for Node values.

This should not be first, because `Object.Export` dominates today.

## 9. Tests to add before implementation

Add tests in `../go-go-golems/go-go-goja/modules/uidsl` for:

### 9.1 Attribute compatibility

```text
string attr
number attr
boolean true attr
boolean false attr
null/undefined omitted attr
value="" preserved
empty non-value attr omitted
class string
class array
class object map
style object map
aria-* attr
data-* attr
```

### 9.2 Child-vs-attrs disambiguation

Ensure these are not treated as attrs:

```javascript
ui.div("text")
ui.div(42)
ui.div(true)
ui.div(ui.span("child"))
ui.div([ui.span("child")])
ui.div(ui.fragment(ui.span("child")))
```

Ensure these are treated as attrs:

```javascript
ui.div({ class: "x" })
ui.div({})
ui.div({ hidden: true })
```

### 9.3 Render output stability

If deterministic ordering is preserved, assert exact output for representative attributes. If deterministic ordering is relaxed later, update tests intentionally rather than accidentally.

## 10. Benchmarks to add in go-go-goja

The load-test benchmark is useful but too coarse for tight iteration. Add Go benchmarks inside `go-go-goja/modules/uidsl`:

```go
BenchmarkUIDSLAttrElementFromCall_Attrs9
BenchmarkUIDSLRenderAttrs_Attrs9
BenchmarkUIDSLRenderPage_Attrs1000
BenchmarkUIDSLRenderPage_Flat1000
```

Each benchmark should report:

```text
ns/op
B/op
allocs/op
```

The target is not just lower latency; it is lower allocation count and bytes allocated per request.

## 11. Validation matrix for any fix

A fix should be accepted only if it improves `render-attrs-1000` and does not regress other anti-overfit workloads.

Run at least:

```text
render-attrs-1000 at 25/s and 100/s
render-flat-1000 at 100/s and 250/s
kanban-fragment at 100/s and 250/s
kanban-fragment-500 at 100/s
db-read-100 at 100/s
```

Profile again if `render-attrs-1000` still queues at `100/s`.

Expected success criteria for first optimization slice:

```text
render-attrs-1000 100/s throughput ratio >= 0.95
render-attrs-1000 100/s p95 < 250ms
allocation profile no longer dominated by Object.Export/baseObject.export
no correctness regressions in uidsl tests
no Kanban regression above 10% p95 or throughput
```

These are intentionally ambitious. If Phase 1 only removes double export, we may not reach all targets, but we should see a clear allocation drop.

## 12. Recommended first implementation slice

Start small:

```text
Replace isAttrs + second Export with one attrsFromValue path.
```

Do not start with response streaming or a full `Element` representation rewrite. The profile says the biggest avoidable waste is in generic conversion before rendering.

The first implementation should:

1. Add focused uidsl microbenchmarks.
2. Add compatibility tests for attrs and child disambiguation.
3. Replace `isAttrs` with `attrsFromValue` that performs only one conversion.
4. Prefer direct primitive reads and direct object key iteration where safe.
5. Re-run microbenchmarks.
6. Re-run `render-attrs-1000` at `100/s`.
7. Compare allocation profile before and after.

## 13. Open questions

- Can we reliably detect Go-wrapped `Node` values without calling full `Export()`?
- Should `Element.Attrs` remain `map[string]any`, or should we introduce a render-ready attr list?
- Is deterministic attribute ordering required in production, or only in tests?
- Can `ui.dsl` return opaque Go nodes to JS in a way that avoids Goja object export for children?
- Should `res.html(node)` render directly to the HTTP writer instead of forcing `RenderAny` to return a string?
- How much of `goja.(*vm).run` is JS object construction that could only be reduced by changing the JS-facing DSL API?

## 14. Final diagnosis

`render-attrs-1000` is not primarily an HTTP, SQLite, or Kanban problem. It is a UI DSL bridge problem.

The current bridge uses generic Goja export in a hot path where the value shapes are known and repetitive. The workload creates thousands of small JavaScript objects, converts them into Go maps and nodes, sorts and renders their attributes, then allocates a large response string. The most actionable part is the conversion boundary:

```text
JavaScript object literal -> goja.Value.Export -> map[string]any -> Element.Attrs
```

Reducing or specializing that conversion is the highest-value next optimization.
