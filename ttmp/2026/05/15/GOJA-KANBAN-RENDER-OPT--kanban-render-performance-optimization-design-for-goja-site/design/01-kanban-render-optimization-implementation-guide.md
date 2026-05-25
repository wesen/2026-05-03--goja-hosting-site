---
Title: Kanban render optimization implementation guide
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - goja
    - kanban
    - performance
    - optimization
    - benchmarking
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: bench/scripts/kanban-board/app.js
      Note: benchmark fixture that enables preciseMove and dragDrop
    - Path: pkg/kanbanddsl/builder.go
      Note: JavaScript builder API for features.preciseMove
    - Path: pkg/kanbanddsl/client_runtime.go
      Note: browser runtime for forms
    - Path: pkg/kanbanddsl/render.go
      Note: main render path and preciseMoveForm implementation
    - Path: pkg/kanbanddsl/types.go
      Note: FeatureSpec and BoardConfig types to extend
ExternalSources: []
Summary: Intern-ready implementation guide for reducing Kanban render cost in goja-site, starting with preciseMoveForm and response-size optimizations.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Kanban render optimization implementation guide

This guide explains the concrete optimization work recommended after the single-VM and multi-VM stress tests. It is written for a new engineer who needs to understand what the Kanban subsystem is, why it became the first performance bottleneck, which files matter, and how to implement and validate an optimization safely.

The immediate recommendation is to reduce the cost of rendering full Kanban board HTML, starting with `preciseMoveForm`. The broader recommendation is to measure every change with the existing benchmark harness before treating it as a performance improvement.

## Implementation update: opinionated simplification

After this guide was first written, the project decision changed: backward compatibility is not required, and the Kanban DSL should remain strongly opinionated. The implemented direction is therefore simpler than the original mode-based proposal. The core DSL no longer exposes `features.preciseMove()` and no longer renders full movement forms on every card. Movement remains a semantic `cardMoved` action. The frontend runtime now provides the accessible card action surface with a compact `Actions` button, keyboard menu behavior, live-region announcements, and focus restoration.

The earlier mode names in this document (`eager`, `none`, `button`, `lazy`) should be read as design history, not as the current implementation target. The current target is: no eager per-card precise forms in the core DSL.

## 1. The system being optimized

`goja-site` hosts trusted JavaScript sites inside Goja runtimes. A JavaScript site can import modules such as `express`, `ui.dsl`, `database`, and `kanban.dsl`. The Kanban DSL lets a site author describe a board in JavaScript and mount it as HTTP routes.

The benchmark Kanban fixture is here:

```text
bench/scripts/kanban-board/app.js
```

It creates 120 cards and builds a board:

```javascript
const board = kanban.board("bench")
  .columns(cols => cols
    .column("todo").title("To Do").done()
    .column("doing").title("Doing").done()
    .column("done").title("Done").done())
  .data(data => data
    .cards(() => cards)
    .id(card => String(card.id))
    .column(card => card.status)
    .position(card => card.position)
    .searchText(card => card.title + " " + card.status + " " + card.size))
  .features(features => features.search().preciseMove().dragDrop())
  .render(render => render.card(card => ui.div(...)))
  .actions(actions => actions.cardMoved(event => { ... }))
  .build();
```

The important feature call is:

```javascript
.features(features => features.search().preciseMove().dragDrop())
```

`preciseMove()` enables an HTML form on every card. That form contains destination-column and destination-position controls. The benchmark profile showed that generating those controls is expensive under load.

## 2. Relevant files

Read these files before changing code:

| File | Purpose |
|---|---|
| `pkg/kanbanddsl/types.go` | Core data structures: `BoardConfig`, `FeatureSpec`, `RenderSpec`, `Board`, and `renderedCard`. |
| `pkg/kanbanddsl/builder.go` | JavaScript builder API that site authors call from Goja. |
| `pkg/kanbanddsl/render.go` | Board, column, card, and `preciseMoveForm` rendering. Main optimization target. |
| `pkg/kanbanddsl/mount.go` | Mounted HTTP routes: client script, fragment endpoint, action endpoint. |
| `pkg/kanbanddsl/client_runtime.go` | Browser runtime for search, precise move form submit, drag/drop, action posting, and HTML replacement. |
| `pkg/kanbanddsl/observer.go` | Observer interface for Kanban metrics. |
| `pkg/observability/kanban.go` | Prometheus observer implementation for Kanban timings and rendered bytes. |
| `bench/scripts/kanban-board/app.js` | Synthetic 120-card benchmark fixture. |
| `scripts/bench-vegeta.sh` | Single-site benchmark runner with metrics and pprof. |
| `ttmp/.../GOJA-MULTI-VM-STRESS/scripts/01-run-multi-vm-vegeta.sh` | Multi-VM benchmark runner with generated `serve-multi` configs. |

## 3. Current render flow

The mounted fragment route is in `pkg/kanbanddsl/mount.go`. The request path is:

```text
GET /_kanban/bench/fragment
```

The route calls:

```go
node, err := b.Render(...)
res.html(node)
```

The action route is:

```text
POST /_kanban/bench/action/cardMoved
```

The action route calls `b.Dispatch`, then usually re-renders the board:

```go
result, err := b.Dispatch(action, bodyObj)
out := map[string]any{"ok": true}
refresh := shouldRefresh(out["refresh"])
if refresh {
    node, err := b.Render(...)
    html, err := uidsl.RenderAny(b.vm, b.vm.ToValue(node))
    out["html"] = html
}
res.json(out)
```

`Board.Render` is in `pkg/kanbanddsl/render.go`. It loads cards, groups them by column, renders columns, renders cards, and returns a UI DSL node tree:

```go
func (b *Board) Render(ctx goja.Value) (uidsl.Node, error) {
    cards, err := b.loadCards(ctx)
    byColumn := groupAndSort(cards)
    children := []uidsl.Node{}
    columnNodes := []uidsl.Node{}
    for _, col := range b.cfg.Columns {
        columnNodes = append(columnNodes, b.renderColumn(ctx, col, byColumn[col.ID]))
    }
    children = append(children, boardElement(columnNodes))
    return root, nil
}
```

Each column renders each card:

```go
for i, card := range cards {
    listChildren = append(listChildren, b.renderCard(ctx, card, i, len(cards)))
}
```

Each card may render the precise move form:

```go
children := []uidsl.Node{body}
if b.cfg.Features.PreciseMove && !b.cfg.Features.ReadOnly {
    children = append(children, b.preciseMoveForm(card, index, columnCount))
}
```

## 4. Why `preciseMoveForm` is expensive

`preciseMoveForm` builds several UI DSL nodes for every card. For each card it creates:

- two hidden inputs,
- one hidden index input,
- a destination-column `select`,
- a destination-position `select`,
- one submit button,
- one `option` node per board column,
- one `option` node per position in the card's current column.

The current implementation is:

```go
func (b *Board) preciseMoveForm(card renderedCard, index, columnCount int) uidsl.Node {
    statusOptions := []uidsl.Node{}
    for _, col := range b.cfg.Columns {
        statusOptions = append(statusOptions, option(col.ID, col.Title, col.ID == card.ColumnID))
    }

    positionOptions := []uidsl.Node{}
    count := columnCount
    if count < 1 { count = 1 }
    for i := 0; i < count; i++ {
        positionOptions = append(positionOptions, option(i, fmt.Sprintf("#%d", i+1), i == index))
    }

    return form(hiddenInputs, selects, button)
}
```

For a 120-card board split across three columns, this creates a large number of small nodes and maps. The UI DSL renderer then traverses those nodes, renders attributes, escapes strings, writes buffers, and allocates intermediate objects. The pprof data shows this path directly:

```text
uidsl.renderNode                                      12.57s cumulative, 30.58%
kanbanddsl.(*Board).preciseMoveForm                   11.57s cumulative, 28.15%
uidsl.renderAttrs                                     10.09s cumulative, 24.55%
uidsl.attrValue                                        4.04s cumulative, 9.83%
runtime.mallocgc                                       9.42s cumulative, 22.92%
runtime.gcDrain                                        9.09s cumulative, 22.12%
```

The underlying issue is not only that one function is slow. The deeper issue is that the current board render creates complete interaction markup for every card on every render. That makes the render cost proportional to the number of cards multiplied by the amount of per-card control markup.

## 5. Immediate optimization goal

The first implementation slice should make precise movement controls optional at render time without breaking existing boards.

The goal is not to remove the feature. The goal is to let site authors choose whether every card should include full precise movement controls in the initial HTML.

Recommended new modes:

| Mode | Behavior | Expected performance effect | User impact |
|---|---|---|---|
| `eager` | Current behavior: render full form for every card. | Baseline. | Full no-JS form controls always present. |
| `none` | Do not render precise move forms. | Largest immediate reduction in HTML and allocations. | Users rely on drag/drop or custom UI. |
| `button` | Render a small button or placeholder per card. Browser opens/generates controls on demand. | Significant reduction in initial HTML. | Requires client-side runtime behavior. |
| `lazy` | Render a placeholder with enough data for client-side form construction. | Similar to `button`, more explicit. | Requires client-side runtime behavior. |

The safest first patch is `none` plus a benchmark fixture that disables precise move. The next patch can add `button` or `lazy`.

## 6. API design

The existing JavaScript API has:

```javascript
.features(features => features.search().preciseMove().dragDrop())
```

Add an optional mode to `preciseMove` while preserving the old call:

```javascript
features.preciseMove()              // existing behavior: eager
features.preciseMove("eager")       // explicit current behavior
features.preciseMove("none")        // disable rendered forms
features.preciseMove("button")      // render small trigger, generate controls on demand
features.preciseMove({ mode: "none" })
features.preciseMove({ mode: "button" })
```

Go type sketch:

```go
type PreciseMoveMode string

const (
    PreciseMoveEager  PreciseMoveMode = "eager"
    PreciseMoveNone   PreciseMoveMode = "none"
    PreciseMoveButton PreciseMoveMode = "button"
)

type FeatureSpec struct {
    Search          SearchSpec
    PreciseMove     bool              // keep for compatibility
    PreciseMoveMode PreciseMoveMode   // new
    DragDrop        bool
    CreateCard      bool
    CardMenu        bool
    ReadOnly        bool
}
```

Compatibility rule:

```text
PreciseMove == true and PreciseMoveMode == "" means eager.
PreciseMove == false means no precise move controls.
```

This keeps old code working and makes the new mode explicit.

## 7. Rendering pseudocode

The first implementation should concentrate the decision inside `renderCard`:

```go
func (b *Board) renderCard(ctx goja.Value, card renderedCard, index, columnCount int) uidsl.Node {
    body := renderCardBody(...)
    children := []uidsl.Node{body}

    if b.shouldRenderPreciseMoveForm() {
        children = append(children, b.preciseMoveForm(card, index, columnCount))
    } else if b.shouldRenderPreciseMoveButton() {
        children = append(children, b.preciseMoveButton(card, index, columnCount))
    }

    return article(attrs, children)
}
```

The helpers should keep the semantics visible:

```go
func (b *Board) preciseMoveMode() PreciseMoveMode {
    if !b.cfg.Features.PreciseMove || b.cfg.Features.ReadOnly {
        return PreciseMoveNone
    }
    if b.cfg.Features.PreciseMoveMode == "" {
        return PreciseMoveEager
    }
    return b.cfg.Features.PreciseMoveMode
}

func (b *Board) shouldRenderPreciseMoveForm() bool {
    return b.preciseMoveMode() == PreciseMoveEager
}
```

For `button` mode, render only stable data attributes:

```go
func (b *Board) preciseMoveButton(card renderedCard, index, columnCount int) uidsl.Node {
    return &uidsl.Element{
        Tag: "button",
        Attrs: map[string]any{
            "type": "button",
            "class": "kb-move-trigger",
            "data-kb-move-trigger": true,
            "data-kb-card-id": card.ID,
            "data-kb-from-column-id": card.ColumnID,
            "data-kb-from-index": index,
        },
        Children: []uidsl.Node{&uidsl.Text{Value: "Move"}},
    }
}
```

The client runtime can then construct a form or menu only when the user clicks the trigger. That moves work from every render to the specific user interaction.

## 8. Builder implementation plan

The builder code lives in:

```text
pkg/kanbanddsl/builder.go
```

Find the feature builder methods. Add parsing for the optional argument. The exact implementation should follow the existing builder style, but the logic should look like this:

```go
func (f *featureBuilder) preciseMove(call goja.FunctionCall) goja.Value {
    f.board.cfg.Features.PreciseMove = true
    mode := parsePreciseMoveMode(call.Argument(0))
    f.board.cfg.Features.PreciseMoveMode = mode
    return f.obj
}
```

Parsing rules:

```go
func parsePreciseMoveMode(v goja.Value) PreciseMoveMode {
    if missingValue(v) {
        return PreciseMoveEager
    }
    if s := strings.TrimSpace(v.String()); s != "" && s != "undefined" {
        return validatePreciseMoveMode(s)
    }
    if obj := v.ToObject(vm); obj != nil {
        return validatePreciseMoveMode(obj.Get("mode").String())
    }
    return PreciseMoveEager
}
```

Invalid modes should fail early during board construction. Do not silently fall back to eager for a typo such as `"buton"`.

## 9. Client runtime implementation plan

The client runtime is:

```text
pkg/kanbanddsl/client_runtime.go
```

It currently handles:

- search input,
- form submit for `[data-kb-move-form]`,
- drag/drop,
- action posting,
- replacing the board root when an action response contains `payload.html`.

For `button` mode, add one delegated click listener:

```javascript
document.addEventListener('click', event => {
  const trigger = event.target.closest('[data-kb-move-trigger]');
  if (!trigger) return;
  event.preventDefault();
  openMoveMenu(trigger);
});
```

The initial implementation can be intentionally small:

```javascript
function openMoveMenu(trigger) {
  const card = trigger.closest('[data-kb-card-id]');
  const board = boardFor(trigger);
  if (!card || !board) return;

  // Phase 1: no rich UI. Use a minimal prompt/select only if needed.
  // Phase 2: render a small positioned menu from board metadata.
}
```

If there is no client-side implementation yet, `none` mode can still be valuable for benchmark and drag/drop-heavy use cases. Do not block the first optimization on a full client-side menu.

## 10. Benchmark fixtures

Add new benchmark fixtures rather than mutating the existing baseline fixture. This preserves comparability.

Recommended files:

```text
bench/scripts/kanban-board-no-precise/app.js
bench/scripts/kanban-board-button-precise/app.js
```

Start with `kanban-board-no-precise`:

```javascript
.features(features => features.search().dragDrop())
```

or, if the new API exists:

```javascript
.features(features => features.search().preciseMove("none").dragDrop())
```

The fixture should keep the same number of cards, same columns, same card renderer, and same action behavior. Only the precise-move rendering mode should change. Otherwise the benchmark will not isolate the optimization.

## 11. Validation plan

### Unit tests

Update builder tests in:

```text
pkg/kanbanddsl/builder_test.go
```

Add cases:

```text
preciseMove() renders data-kb-move-form
preciseMove("eager") renders data-kb-move-form
preciseMove("none") does not render data-kb-move-form
preciseMove("button") renders data-kb-move-trigger and not data-kb-move-form
invalid preciseMove mode fails board construction
```

### Mounted-route tests

Update or add tests in:

```text
pkg/kanbanddsl/mount_test.go
```

Verify that action endpoints still work when precise move forms are not rendered. Drag/drop uses JSON `cardMoved` actions, so it should not require the form markup.

### Benchmark validation

Use the existing baseline first:

```bash
scripts/bench-vegeta.sh --scenario kanban-fragment --duration 10s --warmup-duration 3s --rate 100/s --port 19100 --metrics-port 20100
```

Then add a new scenario for the optimized fixture. If the benchmark harness does not yet support arbitrary Kanban fixture names, extend it with scenarios:

```text
kanban-fragment-no-precise
kanban-action-no-precise
```

The comparison matrix should include:

```text
single VM:
  kanban-fragment baseline vs no-precise at 100/s,200/s
  kanban-action baseline vs no-precise at 50/s,80/s,100/s

multi VM:
  kanban-fragment baseline vs no-precise with 4 VMs at 200/s,400/s
```

### pprof validation

Capture pprof for one degraded baseline cell and one matching optimized cell:

```text
baseline:  kanban-fragment, 4 VMs, 400/s
optimized: kanban-fragment-no-precise, 4 VMs, 400/s
```

The expected improvement is:

- lower cumulative time in `kanbanddsl.(*Board).preciseMoveForm`,
- lower cumulative time in `uidsl.renderNode`,
- lower cumulative time in `uidsl.renderAttrs`,
- lower allocation and GC cost,
- lower response bytes,
- lower p95 and higher achieved throughput.

## 12. Pros and cons of the optimization

### Pros

- It directly targets a measured hotspot.
- It can be implemented behind a backward-compatible API.
- It gives site authors control over server-rendered interaction markup.
- It reduces HTML size and allocation pressure when full forms are unnecessary.
- It preserves the current eager mode for accessibility, no-JavaScript operation, or sites that require full HTML controls.
- It creates an API foundation for later lazy or client-side interaction modes.

### Cons

- Disabling precise forms changes available HTML controls for users who rely on them.
- A lazy or button mode requires more client-side runtime behavior.
- Client-side generated controls need careful accessibility design.
- More modes increase the test matrix.
- If site authors choose `none` without providing another movement interaction, they can remove useful functionality.
- The optimization reduces one major cost but does not solve full-board refresh for action responses by itself.

The safest decision is to keep eager rendering as the compatibility default and require explicit opt-in for lighter modes.

## 13. Intern implementation checklist

1. Read `pkg/kanbanddsl/types.go`, `builder.go`, `render.go`, `mount.go`, and `client_runtime.go`.
2. Read `bench/scripts/kanban-board/app.js` and confirm that it enables `preciseMove()`.
3. Add `PreciseMoveMode` to `FeatureSpec`.
4. Extend the JavaScript builder API for `features.preciseMove(mode)`.
5. Add render helpers for mode selection.
6. Implement `none` mode first.
7. Add tests proving eager compatibility and no-form rendering for `none`.
8. Add a `kanban-board-no-precise` benchmark fixture.
9. Add benchmark harness scenario names if needed.
10. Run unit tests and smoke benchmarks.
11. Run before/after saturation cells.
12. Capture pprof for matching baseline and optimized cells.
13. Update the optimization ticket diary and changelog with exact commands and results.

## 14. Definition of done

The first optimization slice is done when:

- Existing `features.preciseMove()` boards render exactly as before.
- A new mode can suppress full precise move form rendering.
- Tests cover the compatibility and optimized modes.
- The benchmark fixture isolates only the precise-move change.
- Before/after reports show response size, latency, throughput, and pprof differences.
- The report explains whether the optimization is worth keeping as a public API.
