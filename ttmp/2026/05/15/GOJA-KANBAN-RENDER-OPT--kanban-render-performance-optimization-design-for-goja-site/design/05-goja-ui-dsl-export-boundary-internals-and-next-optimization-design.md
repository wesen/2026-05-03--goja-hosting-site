---
Title: Goja UI DSL export boundary internals and next optimization design
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - goja
    - uidsl
    - performance
    - profiling
    - optimization
    - internals
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../others/goja/array.go
      Note: dense array string key and export behavior
    - Path: ../../../../../../../../others/goja/object.go
      Note: ordinary object storage
    - Path: ../../../../../../../../others/goja/object_goreflect.go
      Note: Go-reflect object export behavior
    - Path: ../../../../../../../../others/goja/value.go
      Note: Value interface
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/module.go
      Note: UI DSL native function boundary and attrsFromValue
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/node.go
      Note: Element and Attr representation after cutover
    - Path: ../../../../../../../go-go-golems/go-go-goja/modules/uidsl/render.go
      Note: renderAttrs after attr-list cutover
    - Path: bench/scripts/render-shapes/app.js
      Note: render-attrs-1000 benchmark fixture
ExternalSources: []
Summary: Intern-ready internals guide explaining the render-attrs-1000 bottleneck, Goja Value.Export mechanics, JS-to-Go UI DSL boundaries, and candidate next optimization designs.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Goja UI DSL export boundary internals and next optimization design

This document explains why `render-attrs-1000` is still slow after the Kanban-specific optimization and after cutting `ui.dsl` attributes over to a render-ready `[]Attr` representation. The purpose is not to propose one small patch. The purpose is to give a new intern enough context to understand the request path, the Goja internals, the current UI DSL bridge, and the likely design space for the next optimization.

The short version is this:

```text
The remaining bottleneck is not final HTML attribute rendering.
The remaining bottleneck is the repeated construction and export of many small JavaScript objects across the Goja -> Go native-module boundary.
```

A single `render-attrs-1000` request builds roughly 1000 outer nodes and 1000 inner nodes. Each outer node has a JavaScript object literal for attributes. Each call to `ui.div({ ...attrs... }, child)` crosses from JavaScript into Go. At that crossing, the native module currently converts the JavaScript attrs object into Go data by using Goja's generic export machinery. That machinery is correct and general, but it is not cheap. It enumerates properties, preserves JavaScript semantics, recursively exports values, creates Go maps and slices, and uses caches to handle object cycles.

That generality is the root of the cost. UI attributes are not arbitrary JavaScript object graphs. They are usually a small set of strings, booleans, numbers, class lists, and style maps. We are paying for a general bridge when the hot path has a narrow shape.

## 1. The request timeline

This section locates the bottleneck in the full browser-request path. The important thing to see is that the hot path happens before final HTML rendering.

```text
Browser
  |
  |  HTTP GET /attrs?n=1000
  v
Go net/http
  |
  |  route dispatch
  v
goja-site server
  |
  |  call JS handler in one site VM / owner loop
  v
Goja VM
  |
  |  JS app route runs:
  |    attrPage(1000)
  |
  |  loop 1000 times:
  |    ui.div({ many attrs }, ui.span(...))
  |       |
  |       | native Go function call
  |       v
  |    uidsl.elementFromCall
  |       |
  |       | BOTTLENECK REGION:
  |       | - JS object literal already allocated in Goja
  |       | - native call receives goja.Value arguments
  |       | - attrs object crosses JS -> Go
  |       | - goja.Value.Export / Object.Export enumerates keys
  |       | - Go maps/slices/strings are allocated
  |       | - Go uidsl.Element is built
  |       v
  |    Go uidsl.Element{Attrs: []Attr, Children: ...}
  |
  |  returns full page node tree
  v
Go uidsl renderer
  |
  |  renderNode / renderAttrs
  |  now faster after []Attr cutover
  v
bytes.Buffer / response string
  |
  |  allocate/grow ~250 KB response
  v
HTTP response
  |
  v
Browser receives HTML
```

The final `renderAttrs` step used to allocate and sort a key slice for every element. The `[]Attr` cutover improved that part. The profile still shows that the request spends too much time earlier, while building and exporting the node tree.

## 2. The benchmark that exposes the issue

The fixture lives in:

```text
bench/scripts/render-shapes/app.js
```

The hot route is:

```javascript
app.get("/attrs", (req, res) => {
  res.html(attrPage(Number(req.query.n || 1000)));
});
```

The page builder creates 1000 elements with many attributes:

```javascript
function attrPage(n) {
  const children = [];
  for (let i = 0; i < n; i++) {
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
  }
  return ui.html(
    ui.head(ui.title("goja render attrs bench")),
    ui.body(ui.main(ui.h1("Attribute-heavy render benchmark"), ui.section({ class: "attr-root", role: "list" }, children)))
  );
}
```

This benchmark is synthetic, but it is intentionally ordinary. Many real server-rendered pages are large lists of elements with IDs, classes, ARIA attributes, and data attributes. The benchmark asks a direct question: can `goja-site` repeatedly cross the JavaScript-to-Go UI DSL boundary at high frequency?

Current macro results at `100/s`:

| Version | Throughput | p50 | p95 | p99 | Success |
|---|---:|---:|---:|---:|---:|
| Original follow-up pprof | `67.91/s` | `5.999s` | `13.357s` | `13.929s` | `100%` |
| Single-export slice | `59.70/s` | `9.853s` | `19.337s` | `20.114s` | `100%` |
| `[]Attr` cutover | `71.79/s` | `5.271s` | `11.222s` | `11.714s` | `100%` |

The `[]Attr` cutover helps but does not change the basic shape. The server still cannot keep up with `100/s` on this route.

## 3. The current UI DSL boundary

The UI DSL module is implemented in the local `go-go-goja` repository:

```text
../go-go-golems/go-go-goja/modules/uidsl/module.go
../go-go-golems/go-go-goja/modules/uidsl/node.go
../go-go-golems/go-go-goja/modules/uidsl/render.go
../go-go-golems/go-go-goja/modules/uidsl/components.go
../go-go-golems/go-go-goja/modules/uidsl/table.go
```

The module registers functions such as `ui.div`, `ui.span`, and `ui.section`:

```go
func Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)
    for _, tag := range tags {
        tag := tag
        _ = exports.Set(tag, func(call goja.FunctionCall) goja.Value {
            return vm.ToValue(elementFromCall(vm, tag, call))
        })
    }
}
```

When JavaScript calls:

```javascript
ui.div({ class: "x" }, "hello")
```

Goja invokes the Go closure with a `goja.FunctionCall`. That call contains a slice of `goja.Value` arguments:

```go
type FunctionCall struct {
    This      Value
    Arguments []Value
}
```

The current native constructor does this:

```go
func elementFromCall(vm *goja.Runtime, tag string, call goja.FunctionCall) *Element {
    var attrs []Attr
    args := call.Arguments
    if len(args) > 0 {
        if decoded, ok := attrsFromValue(vm, args[0]); ok {
            attrs = decoded
            args = args[1:]
        }
    }
    return &Element{Tag: tag, Attrs: attrs, Children: nodesFromArgs(args)}
}
```

After the attr-list cutover, the Go-side node representation is:

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

That representation is faster to render. It does not, by itself, remove the cost of converting JavaScript object literals into `[]Attr` values.

## 4. What `Value.Export()` does in Goja

The relevant upstream source is in:

```text
/home/manuel/code/others/goja/value.go
/home/manuel/code/others/goja/object.go
/home/manuel/code/others/goja/array.go
/home/manuel/code/others/goja/object_goreflect.go
/home/manuel/code/others/goja/runtime.go
```

The Goja public interface for JavaScript values is in `value.go`:

```go
type Value interface {
    ToInteger() int64
    ToString() Value
    String() string
    ToFloat() float64
    ToNumber() Value
    ToBoolean() bool
    ToObject(*Runtime) *Object
    SameAs(Value) bool
    Equals(Value) bool
    StrictEquals(Value) bool
    Export() interface{}
    ExportType() reflect.Type
    // ...
}
```

For primitive values, export is simple:

```go
func (i valueInt) Export() interface{}   { return int64(i) }
func (b valueBool) Export() interface{}  { return bool(b) }
func (f valueFloat) Export() interface{} { return float64(f) }
func (s asciiString) Export() interface{} { return string(s) }
```

For objects, the path is different. `Object.Export()` delegates to the object's internal implementation:

```go
func (o *Object) Export() interface{} {
    return o.self.export(&objectExportCtx{})
}
```

The default ordinary object export path is in `object.go`:

```go
func (o *baseObject) export(ctx *objectExportCtx) interface{} {
    if v, exists := ctx.get(o.val); exists {
        return v
    }
    keys := o.stringKeys(false, nil)
    m := make(map[string]interface{}, len(keys))
    ctx.put(o.val, m)
    for _, itemName := range keys {
        itemNameStr := itemName.String()
        v := o.val.self.getStr(itemName.string(), nil)
        if v != nil {
            m[itemNameStr] = exportValue(v, ctx)
        } else {
            m[itemNameStr] = nil
        }
    }

    return m
}
```

This is the essential algorithm:

```text
Object.Export()
  -> check export cache to avoid cycles
  -> enumerate own enumerable string keys
  -> allocate Go map with len(keys)
  -> cache map before recursion
  -> for each key:
       - convert key to Go string
       - get property value by name
       - recursively export the value
       - store in Go map
  -> return map[string]interface{}
```

That is correct for arbitrary JavaScript objects. It is expensive for hot-path UI attributes because it does a full general object export for a small object whose shape is known.

## 5. Why export needs a cache

Goja's export context is:

```go
type objectExportCtx struct {
    cache map[*Object]interface{}
}
```

The cache is used here:

```go
func (ctx *objectExportCtx) get(key *Object) (interface{}, bool) { ... }
func (ctx *objectExportCtx) put(key *Object, value interface{}) { ... }
```

The cache is necessary because JavaScript object graphs can be cyclic:

```javascript
const a = {};
a.self = a;
a.Export(); // must not recurse forever
```

For UI attributes, cycles are not useful. But the generic export function cannot assume that. It must allocate and maintain the context. This is an example of the central tradeoff: generic correctness costs more than a specialized UI path.

## 6. Key enumeration overhead

Ordinary object export starts with:

```go
keys := o.stringKeys(false, nil)
```

`baseObject.stringKeys` is in `object.go`:

```go
func (o *baseObject) stringKeys(all bool, keys []Value) []Value {
    o.ensurePropOrder()
    if all {
        for _, k := range o.propNames {
            keys = append(keys, stringValueFromRaw(k))
        }
    } else {
        for _, k := range o.propNames {
            prop := o.values[k]
            if prop, ok := prop.(*valueProperty); ok && !prop.enumerable {
                continue
            }
            keys = append(keys, stringValueFromRaw(k))
        }
    }
    return keys
}
```

Before enumeration, Goja ensures ECMAScript property order:

```go
func (o *baseObject) ensurePropOrder() {
    if o.lastSortedPropLen < len(o.propNames) {
        o.fixPropOrder()
    }
}
```

`fixPropOrder` exists because ECMAScript has specific ordering rules for integer-like keys. Even if our attribute object uses ordinary non-integer keys, the generic path has to be prepared to enforce the rule.

For `render-attrs-1000`, this overhead is repeated thousands of times. Each attrs object has keys such as:

```text
class
id
data-index
data-kind
data-group
data-label
aria-label
role
tabindex
```

These are not complex keys. But the generic exporter cannot treat them as a special case.

## 7. Object construction overhead before export

The profiler also shows costs such as:

```text
goja.(*baseObject)._put
goja.(*baseObject)._putProp
goja.newBaseObjectObj
goja.(*baseObject).init
goja.asciiString.Concat
goja.stringValueFromRaw
```

These do not come from exporting alone. They come from building the JavaScript objects in the first place.

When JavaScript evaluates:

```javascript
{
  class: "attr-node state-" + (i % 5),
  id: "attr-node-" + i,
  "data-index": String(i),
  role: "listitem"
}
```

Goja must allocate an object, create internal property storage, add properties, maintain property-name order, and store values. The ordinary object representation in `object.go` includes:

```go
type baseObject struct {
    class      string
    val        *Object
    prototype  *Object
    extensible bool

    values    map[unistring.String]Value
    propNames []unistring.String

    lastSortedPropLen, idxPropCount int
    // ...
}
```

Adding a property goes through `setOwnStr` or `_putProp`-style logic. That updates the `values` map and appends to `propNames`. This is necessary for general JavaScript objects. It is costly if the page builder creates thousands of short-lived objects only to immediately export them to Go.

The key point:

```text
Even a perfect export path would not eliminate the cost of constructing thousands of JS object literals.
```

A larger design may need to avoid object literals on the hot path.

## 8. Array export behavior

The user asked whether manipulating arrays on the JavaScript side may be more efficient. The answer is: arrays are promising, but not automatically free.

Goja's dense array export path is in `array.go`:

```go
func (a *arrayObject) export(ctx *objectExportCtx) interface{} {
    if v, exists := ctx.get(a.val); exists {
        return v
    }
    arr := make([]interface{}, a.length)
    ctx.put(a.val, arr)
    if a.propValueCount == 0 && a.length == uint32(len(a.values)) && uint32(a.objCount) == a.length {
        for i, v := range a.values {
            if v != nil {
                arr[i] = exportValue(v, ctx)
            }
        }
    } else {
        for i := uint32(0); i < a.length; i++ {
            v := a.getIdx(valueInt(i), nil)
            if v != nil {
                arr[i] = exportValue(v, ctx)
            }
        }
    }
    return arr
}
```

Dense arrays have a fast path:

```text
if the array has no property descriptors, length matches storage, and every slot is an object slot:
    iterate backing values directly
else:
    get each index through general property access
```

An attrs representation such as this:

```javascript
ui.divA([
  ["class", "attr-node state-1"],
  ["id", "attr-node-1"],
  ["data-index", "1"]
], child)
```

could avoid object property-name enumeration, but it may introduce many nested arrays. Exporting nested arrays still recursively exports each inner array. That means it might trade object-key enumeration for array allocation and nested export.

A flatter array is likely better:

```javascript
ui.divA([
  "class", "attr-node state-1",
  "id", "attr-node-1",
  "data-index", "1"
], child)
```

That representation has fewer objects. It could let the native module decode pairs by index:

```go
func attrsFromFlatArray(vm *goja.Runtime, v goja.Value) []Attr {
    obj := v.ToObject(vm)
    length := int(obj.Get("length").ToInteger())
    attrs := make([]Attr, 0, length/2)
    for i := 0; i+1 < length; i += 2 {
        key := obj.Get(strconv.Itoa(i)).String()
        value := obj.Get(strconv.Itoa(i+1)).String()
        attrs = append(attrs, Attr{Key: key, Value: value})
    }
    return attrs
}
```

This pseudocode is not necessarily optimal because `Object.Get("0")` still performs property lookup. But it illustrates the direction: use a representation that avoids arbitrary object export and recursive map allocation.

## 9. Direct property reads versus `Export()`

A direct decoder could use:

```go
obj := v.ToObject(vm)
keys := obj.Keys()
for _, key := range keys {
    value := obj.Get(key)
    // decode primitive directly
}
```

This avoids `Object.Export()` creating a Go map. We tried a version of this during the first optimization slice. It did not win for this workload. The profile shifted into:

```text
goja.(*Object).Keys
goja.(*enumerableIter).next
goja.(*objectPropIter).next
per-property conversion
```

That result is important. It tells us that simply replacing `Export()` with public `Object.Keys()` plus `Get()` is not enough. It still uses Goja's generic object-key machinery and performs property lookups from Go for every key.

The practical lesson:

```text
If we keep JavaScript object literals as the attrs representation, public Goja APIs may not provide a cheap enough escape hatch.
```

The next meaningful win may require either changing the JavaScript-facing representation or adding a specialized internal Goja API that can expose object storage more directly.

## 10. Go-side builders

A Go-side builder means avoiding JavaScript object literals and JavaScript arrays for the hottest repeated structure. Instead of having JavaScript construct arbitrary values that Go then decodes, JavaScript calls methods that mutate a Go-owned builder.

A sketch:

```javascript
const b = ui.builder();
for (let i = 0; i < 1000; i++) {
  b.divStart();
  b.attr("class", "attr-node state-" + (i % 5));
  b.attr("id", "attr-node-" + i);
  b.attr("data-index", String(i));
  b.attr("role", "listitem");
  b.spanStart();
  b.attr("class", "label");
  b.text("node-" + i);
  b.end();
  b.end();
}
return b.finish();
```

The Go side would receive many native calls, but each call passes simple primitives rather than complex object graphs. The builder stores `[]Attr` and `[]Node` directly.

A lower-call-count version could batch attrs:

```javascript
b.el("div",
  "class", "attr-node state-" + (i % 5),
  "id", "attr-node-" + i,
  "data-index", String(i),
  "role", "listitem",
  child,
);
```

Potential benefits:

- Avoids JS object literal allocation for attrs.
- Avoids generic `Object.Export()` for attrs.
- Creates Go `Attr` values directly.
- Can pre-size slices in the builder.

Potential costs:

- Many native calls can also be expensive.
- The API is less idiomatic JavaScript than object-literal attrs.
- The builder needs a careful stack discipline and good errors.
- It is a larger API change.

A Go-side builder is likely faster if it reduces object allocation and export enough to offset native call overhead. It should be tested with microbenchmarks before becoming the default DSL.

## 11. Specialized constructors

A less invasive alternative is to add specialized constructors for common patterns:

```javascript
ui.divClass("attr-node state-1", child)
ui.divIdClass("attr-node-1", "attr-node state-1", child)
ui.divAttrs("class", "x", "id", "y", "role", "listitem", child)
```

The most general of these is a flat vararg attr constructor:

```javascript
ui.el("div",
  ui.attrs("class", "x", "id", "y", "role", "listitem"),
  child
)
```

or:

```javascript
ui.divKV("class", "x", "id", "y", "role", "listitem", child)
```

Potential benefits:

- Avoids object literal enumeration.
- Can decode primitive arguments directly from `goja.Value` without map export.
- Keeps the functional style of `ui.div(...)`.
- Lets authors opt into faster paths only where needed.

Potential costs:

- API surface can become messy if every tag gets many variants.
- Child-vs-attr boundary becomes harder to parse for varargs.
- TypeScript definitions and documentation must be updated.

For `render-attrs-1000`, the highest-value specialized constructor would be one that accepts attrs as a flat primitive list. That avoids both JavaScript object literal allocation and nested array allocation.

## 12. Array-based attrs

An array-based representation is a compromise between object literals and Go-side builders.

Possible shapes:

```javascript
// nested pairs
ui.div([ ["class", "x"], ["id", "y"] ], child)

// flat pairs
ui.div([ "class", "x", "id", "y" ], child)

// tagged attrs wrapper
ui.div(ui.attrs("class", "x", "id", "y"), child)
```

The `ui.attrs(...)` wrapper is useful because it disambiguates attrs from children. It could return a Go-wrapped attrs object rather than a JavaScript array:

```javascript
ui.div(ui.attrs("class", "x", "id", "y"), child)
```

The Go implementation could make `ui.attrs` return a Go value that already contains `[]Attr`. Then `ui.div` can detect that value as attrs without exporting an object literal.

Pseudocode:

```go
type AttrsValue struct {
    Attrs []Attr
}

exports.Set("attrs", func(call goja.FunctionCall) goja.Value {
    attrs := decodeFlatAttrs(call.Arguments)
    return vm.ToValue(&AttrsValue{Attrs: attrs})
})

func attrsFromValue(v goja.Value) ([]Attr, bool) {
    if x, ok := v.Export().(*AttrsValue); ok {
        return x.Attrs, true
    }
    // fallback or error depending on desired strictness
}
```

This is promising because it turns the attrs conversion into one Go-native operation and gives `ui.div` a cheap way to recognize precompiled attrs. However, if `ui.attrs(...)` is called inside the loop for every element, it still allocates an `AttrsValue` per element. The benchmark would tell us whether that is still cheaper than object literal export.

## 13. Rendering directly from Goja values

Another direction is to avoid normalizing into a Go node tree before rendering. Instead, the renderer could walk a compact JavaScript representation directly.

For example, the JS side could produce arrays like:

```javascript
["div", ["class", "x", "id", "y"], [
  ["span", ["class", "label"], ["node-1"]]
]]
```

The Go renderer would consume this shape without converting the entire graph into `Element` structs first. It would recursively read tag, attrs, and children and write HTML.

Potential benefits:

- Avoids allocating Go `Element` objects for every node.
- Avoids creating a second Go-side tree when the JS tree already exists.
- Could stream directly to the response writer.

Potential costs:

- Repeated indexed reads from Goja values may be expensive through public APIs.
- Error handling becomes more complex.
- The representation is less ergonomic for JS authors unless hidden behind helper functions.
- This may require deeper Goja internals access to be fast.

This direction is invasive but may be the right long-term architecture if server-rendered pages are a core workload.

## 14. Direct Goja internals access

The user specifically called out that this may require work directly inside Goja. That is plausible.

The public Goja APIs are correct but general:

```go
Value.Export()
Object.Keys()
Object.Get(name)
Runtime.ExportTo(...)
```

For hot UI rendering, we may want lower-level operations such as:

```go
// hypothetical API
obj.FastOwnEnumerableStringProperties(func(key string, value goja.Value) bool)
```

Such an API would let the caller iterate own enumerable string properties without constructing a `[]Value` key list and without recursively exporting the object. It would still need to respect Goja's semantics, but it could avoid intermediate maps and reduce allocation.

A possible internal helper in Goja might look conceptually like:

```go
func (o *Object) ForEachOwnEnumerableStringProperty(fn func(name string, value Value) bool) {
    // ensure property order once
    // iterate baseObject.propNames directly for ordinary objects
    // skip non-enumerable valueProperty entries
    // call fn(name, value)
}
```

This is not a trivial upstream change because Goja supports multiple object implementations:

- ordinary objects (`baseObject`),
- arrays (`arrayObject`),
- proxies,
- dynamic objects,
- Go-reflect objects,
- typed arrays,
- functions,
- maps and sets.

A safe API must either work for all of them or clearly document that it is an optimized path for ordinary objects with fallback. If introduced upstream, it needs tests around property order, enumerability, proxies, accessors, and panics/exceptions.

## 15. Which approach is likely to be fastest?

The likely order from least invasive to most invasive is:

| Approach | Expected impact | Invasiveness | Notes |
|---|---:|---:|---|
| Keep object literals and use `Export()` | Baseline | Low | Correct, ergonomic, too slow for hot attrs. |
| Keep object literals and use `Object.Keys()` + `Get()` | Low/uncertain | Low-medium | Tried once; shifted cost into key iteration and property access. |
| `[]Attr` render-ready cutover | Low-medium | Medium | Helps final rendering; macro still slow. Done. |
| `ui.attrs(...)` returning Go-owned attrs | Medium | Medium | Avoids object literal export if authors use it. Still creates attrs wrapper per node. |
| Flat vararg constructors | Medium-high | Medium-high | Avoids object literals and arrays; API design risk. |
| Go-side builder | High | High | Avoids object graphs; may pay native-call overhead. Needs benchmark. |
| Direct render from compact JS arrays | High but uncertain | High | Avoids Go node tree; public Goja indexed reads may still cost. |
| New Goja internals API for fast property iteration | High | Very high | Could preserve object-literal ergonomics, but requires careful upstream design. |

The most promising two experiments are:

1. `ui.attrs(...)` returning a Go-owned attrs wrapper.
2. A flat vararg constructor that avoids object literals entirely.

The most ambitious experiment is a Goja internals API that exposes ordinary object properties without full export.

## 16. Concrete experiment plan

Do not rewrite the whole DSL first. Build small competing prototypes and benchmark them.

### Experiment A: `ui.attrs(...)` wrapper

JavaScript:

```javascript
ui.div(ui.attrs(
  "class", "attr-node state-" + (i % 5),
  "id", "attr-node-" + i,
  "data-index", String(i),
  "role", "listitem"
), child)
```

Go sketch:

```go
type AttrsValue struct { Attrs []Attr }

func attrsFunction(call goja.FunctionCall) goja.Value {
    attrs := decodeFlatPairs(call.Arguments)
    return vm.ToValue(&AttrsValue{Attrs: attrs})
}

func attrsFromValue(v goja.Value) ([]Attr, bool) {
    if wrapper, ok := v.Export().(*AttrsValue); ok {
        return wrapper.Attrs, true
    }
    return nil, false
}
```

Benchmark against object-literal attrs.

### Experiment B: flat constructor

JavaScript:

```javascript
ui.divKV(
  "class", "attr-node state-" + (i % 5),
  "id", "attr-node-" + i,
  "data-index", String(i),
  "role", "listitem",
  ui.spanKV("class", "label", "node-" + i)
)
```

This needs a clean child boundary. One possible rule is: key/value pairs must come first and children start after a sentinel:

```javascript
ui.divKV("class", "x", "id", "y", ui.children(child1, child2))
```

or:

```javascript
ui.el("div", ui.attrs(...), child1, child2)
```

### Experiment C: Go builder

JavaScript:

```javascript
const b = ui.builder();
b.open("div");
b.attr("class", "x");
b.text("hello");
b.close();
return b.node();
```

Benchmark two versions:

- many small native calls,
- batched `open(tag, attrs...)` calls.

### Experiment D: Goja fast property iteration API

Prototype locally in `/home/manuel/code/others/goja` or a fork branch before proposing upstream.

Hypothetical API:

```go
func (o *Object) ForEachOwnEnumerableStringProperty(fn func(name string, value Value) bool)
```

Then `ui.dsl` can decode attrs without `Export()` and without `Object.Keys()` allocating a key slice.

This experiment needs a careful correctness suite.

## 17. Benchmark criteria

Each experiment should be judged by three levels of evidence.

### Microbenchmarks

Run in `go-go-goja/modules/uidsl`:

```text
BenchmarkElementFromCallAttrs9
BenchmarkRenderPageAttrs1000
BenchmarkRenderPageFlat1000
```

Add new variants:

```text
BenchmarkElementFromCallAttrsWrapper9
BenchmarkElementFromCallFlatPairs9
BenchmarkBuilderAttrs1000
```

Track:

```text
ns/op
B/op
allocs/op
```

### Macro benchmark

Run in `goja-site`:

```text
render-attrs-1000 at 25/s, 100/s
render-flat-1000 at 100/s, 250/s
kanban-fragment at 100/s, 250/s
kanban-fragment-500 at 100/s
```

The first success target is:

```text
render-attrs-1000 at 100/s should approach throughput ratio >= 0.95 and p95 < 250ms.
```

That is ambitious. If the first experiment does not hit it, the profile should at least show a clear reduction in object export and GC.

### Profile evidence

After each serious experiment, capture pprof and compare:

```text
go tool pprof -top cpu.pprof
go tool pprof -top allocs.pprof
```

The next successful design should reduce:

```text
goja.(*baseObject).export
goja.(*Object).Export
goja.(*baseObject)._put
goja.(*baseObject).stringKeys
goja.(*Object).Keys
runtime.mallocgc
runtime.gcDrain
```

If these remain dominant, the optimization did not attack the actual boundary.

## 18. API design guidance

If we add a faster UI DSL API, it should not become a pile of special cases. The current object-literal API is ergonomic:

```javascript
ui.div({ class: "x" }, child)
```

A faster API should be explicit enough that authors understand why it exists:

```javascript
ui.div(ui.attrs("class", "x", "id", "y"), child)
```

This says: attrs are a DSL value, not an arbitrary object. That is a good conceptual boundary. It also leaves room for the implementation to store attrs as Go-owned `[]Attr` rather than JavaScript objects.

Avoid adding dozens of ad hoc helpers such as:

```text
ui.divClass
ui.divClassId
ui.spanClass
ui.sectionRoleClass
```

Those can be fast, but they make the DSL harder to teach and maintain. Prefer one or two composable primitives:

```text
ui.attrs(...)
ui.el(tag, attrs, children...)
ui.builder()
```

## 19. What an intern should remember

The important mental model is this:

- `Value.Export()` is a general JavaScript-to-Go conversion API. It is correct for arbitrary values, but expensive for thousands of small object literals.
- Ordinary object export enumerates keys, allocates a Go map, recursively exports values, and uses a cache to handle cycles.
- Arrays have a dense export fast path, but nested arrays still allocate and recursively export. A flat representation is more promising than nested pairs.
- Public `Object.Keys()` plus `Get()` is not automatically faster than `Export()` because it still uses generic key iteration and property lookup machinery.
- The `[]Attr` cutover made final HTML attribute writing faster, but the macro bottleneck remains earlier at object construction/export time.
- A Go-side builder or attrs wrapper may be faster because it changes the representation rather than only optimizing the renderer.
- A deeper Goja internals API could preserve object-literal ergonomics, but it is invasive and must be designed with JavaScript semantics in mind.

## 20. Recommended next step, but not now

Because the next step is likely invasive, we should stop here for the current optimization thread and keep the evidence. When we resume, start with experiments, not a rewrite.

Recommended order:

1. Prototype `ui.attrs(...)` returning a Go-owned attrs wrapper.
2. Prototype a flat-pair constructor or `ui.el(tag, attrs, children...)` form.
3. Compare both with microbenchmarks and `render-attrs-1000` macro benchmarks.
4. Only then consider modifying Goja internals for fast ordinary-object property iteration.

The next design should be accepted only if it improves the macro benchmark and keeps the DSL teachable.
