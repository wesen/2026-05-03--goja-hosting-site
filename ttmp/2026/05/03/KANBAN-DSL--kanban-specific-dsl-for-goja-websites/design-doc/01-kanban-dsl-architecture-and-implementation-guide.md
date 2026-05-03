---
Title: Kanban DSL Architecture and Implementation Guide
Ticket: KANBAN-DSL
Status: active
Topics:
    - go
    - goja
    - javascript
    - web
    - glazed
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/kanban/scripts/app.js
      Note: Current hand-written Kanban logic that kanban.dsl should replace
    - Path: pkg/app/server.go
      Note: Runtime composition point for adding kanban.dsl registrar
    - Path: pkg/uidsl/module.go
      Note: Current ui.dsl module pattern and tag constructors that kanban.dsl should reuse
    - Path: pkg/uidsl/render.go
      Note: Safe HTML renderer that kanban.dsl should delegate to
    - Path: pkg/web/express_module.go
      Note: Express app object that board.mount(app
    - Path: pkg/web/host.go
      Note: Runtime-owner request dispatch and static mounting behavior
    - Path: pkg/web/request_response.go
      Note: JavaScript request/response API used for callback action routes
ExternalSources: []
Summary: Design for a kanban.dsl module that renders flexible Kanban boards, mixes with ui.dsl, ships a client runtime, and calls server-side Goja callbacks for actions such as cardMoved.
LastUpdated: 2026-05-03T15:15:00-04:00
WhatFor: Guide an intern through implementing a Kanban-specific DSL for goja-site with server-side callbacks and client-managed interactions.
WhenToUse: Use before implementing kanban.dsl, callback dispatch, the Kanban browser runtime, board rendering helpers, or Kanban examples.
---


# Kanban DSL Architecture and Implementation Guide

## Executive summary

The current `goja-site` Kanban example demonstrates the raw ingredients: `ui.dsl` can render HTML, `express` can register endpoints, `database` can persist cards, static assets can serve images and browser JavaScript, and a hand-written `/app.js` can add search and movement. The next step is to avoid writing custom browser JavaScript for every board while still keeping the power to design different boards.

This guide proposes a new `kanban.dsl` module. It is a Kanban-specific server-side DSL exposed to Goja JavaScript as `require("kanban.dsl")`. A script author should be able to write:

```javascript
const db = require("database");
const express = require("express");
const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

const app = express.app();

const board = kanban.board("trail-notes", {
  title: "Trail Notes: Cascade Loop",
  columns: [
    { id: "todo", title: "To Do" },
    { id: "progress", title: "In Progress" },
    { id: "done", title: "Done" },
    { id: "someday", title: "Someday" },
  ],
  cards(ctx) {
    return db.query("SELECT * FROM cards ORDER BY position, id");
  },
  search(card, query) {
    const haystack = `${card.title} ${card.description} ${card.tag}`.toLowerCase();
    return haystack.includes(query.toLowerCase());
  },
  renderCard(card, ctx) {
    return ui.fragment(
      ui.h3(card.title),
      ui.p(card.description),
      card.image ? ui.img({ class: "card-image", src: card.image, alt: "" }) : null,
      ui.div({ class: "card-meta" }, ui.span({ class: "tag" }, card.tag), ui.time(card.due_date))
    );
  },
  actions: {
    cardMoved(event) {
      moveCardInDatabase(event.cardId, event.to.columnId, event.to.index);
      return { ok: true, refresh: true };
    },
    cardCreated(event) {
      createCard(event.input);
      return { ok: true, refresh: true };
    },
  },
});

board.mount(app, "/_kanban");

app.get("/", (req, res) => {
  res.html(ui.page({ title: "Trail Notes" }, board.render({ query: req.query })));
});
```

The browser should then understand how to:

- live-search the board,
- drag/drop cards,
- move cards to exact columns/positions,
- submit create/edit/delete actions,
- call server-side callbacks such as `cardMoved`,
- replace board HTML with server-rendered fragments,
- preserve freedom for each app to customize card layout with `ui.dsl`.

The core idea is **server-authored UI with a generic Kanban browser runtime**:

```text
Server-side Goja JS describes the board and callbacks.
kanban.dsl renders HTML + data attributes + client runtime script.
Browser runtime reads data attributes and posts action events.
kanban.dsl dispatches events back into server-side Goja callbacks.
Callbacks update DB/application state and return refresh instructions.
Browser runtime replaces board fragments rendered by server-side JS.
```

This is special because it bridges browser interactions back into a long-lived server-side Goja runtime. The design must be careful about callback identity, event envelopes, serialization, HTML rendering, concurrency, and progressive enhancement.

## Goals and non-goals

### Goals

- Provide a `kanban.dsl` module for authoring Kanban boards in server-side Goja JavaScript.
- Avoid per-app hand-written client-side JavaScript for standard board interactions.
- Allow server-side callback functions such as `cardMoved`, `cardCreated`, `cardUpdated`, `cardDeleted`, and `cardClicked`.
- Keep flexible rendering: cards, column headers, toolbar, empty states, and details can be custom `ui.dsl` components.
- Support multiple boards per page and multiple boards per app.
- Provide a generic browser runtime that knows how to call the server-side callback dispatcher.
- Support exact movement semantics: card ID, source column/index, destination column/index, and optional neighboring cards.
- Support search/filtering as a standard board feature.
- Support progressive enhancement: basic HTML should still be meaningful without JavaScript.
- Keep the existing `ui.dsl` useful and composable rather than replacing it.

### Non-goals for v1

- Not a React/Vue/Svelte clone.
- Not a full client state-management framework.
- Not a general drag/drop framework for arbitrary UI.
- Not designed for untrusted third-party scripts.
- Not real-time multi-user collaboration in v1.
- Not offline-first.
- Not a full permissions/auth system, though the protocol should leave room for authorization hooks.

## Current-state architecture and evidence

This section maps the existing implementation so the intern understands what `kanban.dsl` should build on.

### `ui.dsl` is a safe server-side HTML AST builder

`pkg/uidsl/node.go` defines the current HTML node model:

- `Node` is the interface for renderable nodes (`pkg/uidsl/node.go:3-4`).
- `Document` stores `Title`, `Head`, and `Body` (`pkg/uidsl/node.go:6-10`).
- `Element` stores `Tag`, `Attrs`, and `Children` (`pkg/uidsl/node.go:14-18`).
- `Text`, `RawHTML`, and `Fragment` complete the basic AST (`pkg/uidsl/node.go:22-32`).

`pkg/uidsl/module.go` exposes this to JavaScript:

- `ui.dsl` and `ui` are registered as native modules (`pkg/uidsl/module.go:15-18`).
- A large list of standard tags is exported (`pkg/uidsl/module.go:21`).
- Each tag function creates an `Element` with attributes and children (`pkg/uidsl/module.go:24-49`).
- Helpers include `fragment`, `text`, `raw`, `render`, and `page` (`pkg/uidsl/module.go:30-36`).

`pkg/uidsl/render.go` renders safely:

- `RenderAny` normalizes a Goja value and renders it (`pkg/uidsl/render.go:13-23`).
- `NormalizeExport` turns strings, arrays, nodes, and primitive values into nodes (`pkg/uidsl/render.go:32-52`).
- Text is escaped (`pkg/uidsl/render.go:112-115`).
- Attributes are escaped and sorted (`pkg/uidsl/render.go:128-160`).
- `RawHTML` is the only raw escape hatch (`pkg/uidsl/render.go:114-115`).

`kanban.dsl` should reuse this model. It should not reimplement HTML escaping.

### The Express-style module can mount routes but currently only JavaScript handlers

`pkg/web/express_module.go` registers `require("express")`:

- `ExpressRegistrar` registers the module into the runtime require registry (`pkg/web/express_module.go:17-23`).
- `express.app()` returns an app object (`pkg/web/express_module.go:26-31`).
- The app object supports route registration for `get`, `post`, `put`, `patch`, `delete`, and `all` (`pkg/web/express_module.go:31-43`).
- It also supports `app.static(prefix, dir)` (`pkg/web/express_module.go:44-50`).

The `app.post(...)` API expects a JavaScript function. A future `kanban.dsl` board can mount endpoints by calling `app.get` and `app.post` from Go-exposed methods, or by returning JS functions to the script author.

### The host dispatches requests into Goja through a runtime owner

`pkg/web/host.go` owns route dispatch:

- `Host` stores a route registry, static mounts, renderer, runtime owner, and dev flag (`pkg/web/host.go:23-29`).
- Static mounts are checked before JavaScript routes (`pkg/web/host.go:44-50`).
- Request dispatch calls `h.owner.Call(...)` so Goja access is serialized (`pkg/web/host.go:66-81`).

This is critical for `kanban.dsl`. Browser events may arrive from many HTTP goroutines. Server-side callbacks must run on the Goja runtime owner, never directly from arbitrary goroutines.

### Request and response APIs already support JSON-style callbacks

`pkg/web/request_response.go` defines the JavaScript-facing request and response objects:

- The request object includes method, URL, path, query, params, headers, cookies, IP, body, and raw body (`pkg/web/request_response.go:16-27`).
- `RequestDTO.Map()` converts field names to lower-case JavaScript keys (`pkg/web/request_response.go:29-42`).
- The response object exports `status`, `set`, `type`, `json`, `send`, `html`, `redirect`, and `end` (`pkg/web/request_response.go:87-102`).

`kanban.dsl` can use JSON POST bodies for action envelopes without modifying the core request/response layer.

### The current app-specific Kanban implementation is too verbose

The current example in `examples/kanban/scripts/app.js` implements many things manually:

- Board columns (`examples/kanban/scripts/app.js:8-13`).
- Database schema and migration (`examples/kanban/scripts/app.js:28-45`).
- Card search and filtering (`examples/kanban/scripts/app.js:67-97`).
- Position normalization and movement (`examples/kanban/scripts/app.js:99-130`).
- HTML/CSS for the board (`examples/kanban/scripts/app.js:137-288`).
- Browser runtime string returned by `clientScript()` (`examples/kanban/scripts/app.js:290-379`).
- Routes for page, CSS, app JS, API cards, create card, and move card (`examples/kanban/scripts/app.js:384-436`).

This is precisely what `kanban.dsl` should reduce. The author should not have to write 90 lines of generic browser drag/drop/search code for every Kanban board.

## Core design idea

### The board is a server-side object with three faces

A `kanban.dsl` board has three responsibilities:

1. **Render face**: produce `ui.dsl` nodes for the current board state.
2. **Protocol face**: expose generic HTTP endpoints for fragments and actions.
3. **Callback face**: dispatch events from the browser back to server-side Goja callbacks.

```text
+------------------------------+
| kanban.board(...)            |
|                              |
|  render(ctx) -> ui.Node      |
|  mount(app, prefix)          |
|  dispatch(actionEnvelope)    |
|                              |
|  callbacks:                  |
|    cardMoved(event)          |
|    cardCreated(event)        |
|    cardUpdated(event)        |
+------------------------------+
```

### The browser runtime should not know application data details

The generic client runtime should know Kanban mechanics, not app-specific database fields. It should understand:

- board ID,
- card ID,
- column ID,
- source index,
- destination index,
- visible search query,
- action names,
- endpoint URLs.

It should not know:

- SQL schema,
- how cards are rendered,
- what tags mean,
- how to persist cards,
- app-specific validation rules.

Those belong in server-side callbacks.

### Server-rendered fragments avoid client template complexity

The client runtime can update the board in two ways:

1. Optimistically rearrange DOM, then reload or refresh a fragment.
2. Ask the server for a re-rendered HTML fragment after every action.

Recommendation for v1: **server-rendered fragment refresh**.

Why:

- Card rendering can use arbitrary `ui.dsl` components.
- The client runtime does not need a card template language.
- Server-side callbacks can change any field, not only status/position.
- It naturally supports mixed `kanban.dsl` + `ui.dsl` content.

Protocol:

```text
Browser posts cardMoved event
  -> server callback updates DB
  -> kanban.dsl rerenders board fragment
  -> server returns { ok: true, html: "<div class=...>...</div>" }
  -> browser replaces board root innerHTML
  -> browser rebinds behavior
```

## Proposed JavaScript API


### Preferred authoring API: fluid builder with Go-side enforcement

The preferred public API should be a **fluid builder**, not a large loosely typed options object. The object-literal API shown below is still useful as an internal shape and possible advanced escape hatch, but it is too easy for an app author to accidentally omit required fields, misspell callback names, or provide inconsistent features. A builder lets the Go native module enforce ordering, required fields, duplicate IDs, and callback signatures much earlier.

The target API should look like this:

```javascript
const board = kanban
  .board("trail-notes")
  .title("Trail Notes: Cascade Loop")
  .theme("field-notes")
  .columns(cols => cols
    .column("todo").title("To Do").done()
    .column("progress").title("In Progress").limit(3).done()
    .column("done").title("Done").terminal(true).done()
    .column("someday").title("Someday").done()
  )
  .data(data => data
    .cards(ctx => db.query("SELECT * FROM cards ORDER BY position, id"))
    .id(card => String(card.id))
    .column(card => card.status)
    .position(card => Number(card.position || 0))
    .searchText(card => `${card.title} ${card.description} ${card.tag}`)
  )
  .features(features => features
    .search({ mode: "client" })
    .preciseMove()
    .dragDrop()
    .createCard()
  )
  .render(render => render
    .card((card, ctx) => ui.fragment(
      ui.h3(card.title),
      ui.p(card.description),
      card.image ? ui.img({ src: card.image, class: "card-image", alt: "" }) : null
    ))
    .emptyColumn((column) => ui.div({ class: "empty" }, `No ${column.title} cards`))
  )
  .actions(actions => actions
    .cardMoved(event => {
      moveCard({ id: event.cardId, toStatus: event.to.columnId, toIndex: event.to.index });
      return { ok: true, refresh: true };
    })
    .cardCreated(event => {
      createCard(event.input);
      return { ok: true, refresh: true };
    })
  )
  .build();
```

The `.build()` call is important. Before `.build()`, the value is a mutable builder. After `.build()`, it is an immutable board with `.render(...)`, `.mount(...)`, `.dispatch(...)`, and `.clientScriptURL()` methods. This mirrors strong Go patterns: collect configuration, validate it once, freeze it, then use it at runtime.

#### Why a fluid builder is better than a large object literal

Object-literal API:

```javascript
kanban.board("x", {
  columns: [ ... ],
  cards() { ... },
  actions: { cardMoved() { ... } }
})
```

Problems:

- Misspelled fields silently disappear unless the Go decoder validates every key.
- Required fields are all checked at one late point.
- Duplicate column IDs are easy to miss.
- The app can specify `dragDrop: true` without a `cardMoved` callback.
- The app can define `renderCard` before core card identity functions exist.
- The Go side has less context for helpful error messages.
- The API does not guide a new intern or app author through the correct mental model.

Fluid builder API:

```javascript
kanban.board("x")
  .columns(c => c.column("todo").title("To Do").done())
  .data(d => d.cards(load).id(cardId).column(cardColumn))
  .actions(a => a.cardMoved(onMove))
  .build()
```

Benefits:

- Each method can validate its own input immediately.
- Builders can enforce sub-state: a column must call `.title(...)` before `.done()`.
- `.build()` can report accumulated missing requirements with precise messages.
- The Go implementation can expose only valid methods on each sub-builder.
- Feature methods can register implied requirements.
- The final board object can be immutable.

#### Builder lifecycle

```text
kanban.board(id)
  -> BoardBuilder (mutable)
    -> .columns(fn) creates ColumnListBuilder
      -> .column(id) creates ColumnBuilder
      -> .title(...).limit(...).done()
    -> .data(fn) creates DataBuilder
      -> .cards(fn).id(fn).column(fn).position(fn).searchText(fn)
    -> .render(fn) creates RenderBuilder
    -> .features(fn) creates FeatureBuilder
    -> .actions(fn) creates ActionBuilder
  -> .build()
    -> validate all accumulated config
    -> freeze into Board
    -> register board by ID in kanban runtime registry
```

#### Type-state-like enforcement in Go

JavaScript cannot enforce compile-time types, but Go can expose different objects with different method sets. For example, `columns.column("todo")` can return a `ColumnBuilder` that has `.title(...)`, `.limit(...)`, `.className(...)`, `.attrs(...)`, and `.done()`. The column is not appended to the board until `.done()` is called.

Go-side sketch:

```go
type BoardBuilder struct {
  vm *goja.Runtime
  runtime *Runtime
  id string
  columns []ColumnSpec
  pendingColumn *ColumnBuilder
  data DataSpec
  render RenderSpec
  features FeatureSpec
  actions map[string]goja.Callable
  errors []error
  built bool
}

type ColumnBuilder struct {
  parent *ColumnListBuilder
  spec ColumnSpec
  done bool
}

func (b *BoardBuilder) JSObject(vm *goja.Runtime) *goja.Object {
  obj := vm.NewObject()
  _ = obj.Set("title", func(title string) goja.Value {
    b.title = strings.TrimSpace(title)
    return obj
  })
  _ = obj.Set("columns", func(fn goja.Value) goja.Value {
    b.runColumnsBuilder(fn)
    return obj
  })
  _ = obj.Set("data", func(fn goja.Value) goja.Value {
    b.runDataBuilder(fn)
    return obj
  })
  _ = obj.Set("build", func() goja.Value {
    board, err := b.Build()
    if err != nil { panic(vm.NewGoError(err)) }
    return board.JSObject(vm)
  })
  return obj
}
```

The key design is that `.columns(fn)` receives a dedicated columns builder, not the same board object:

```javascript
.columns(cols => cols
  .column("todo").title("To Do").done()
  .column("done").title("Done").done()
)
```

If the app forgets `.done()`, `.build()` should fail with:

```text
kanban.board("trail-notes"): column "todo" was started but not finalized with .done()
```

#### Required validation rules

At `.build()`, enforce these rules:

Board-level:

- Board ID is non-empty and unique in the current Goja runtime.
- At least one column exists.
- `data.cards(fn)` is registered.
- `data.id(fn)` is registered or the default `card.id` field is explicitly accepted.
- `data.column(fn)` is registered or the default `card.status` field is explicitly accepted.
- If `features.dragDrop()` or `features.preciseMove()` is enabled, `actions.cardMoved(fn)` must be registered unless the board declares `.readOnly()`.
- If `features.createCard()` is enabled, `actions.cardCreated(fn)` must be registered.
- If `features.cardMenu()` is enabled, `actions.cardMenuAction(fn)` must be registered.

Column-level:

- Column ID is non-empty.
- Column IDs are unique.
- Column title is non-empty.
- Column limit is positive if provided.

Callback-level:

- Registered action callbacks are functions.
- Unknown action names are rejected unless the builder uses `.customAction(name, fn)`.
- Reserved callback names cannot be overwritten accidentally.

Rendering-level:

- `render.card(fn)` must be a function if provided.
- Default card renderer requires at least `card.title` or an explicit `.render.card(...)`.

Feature-level:

- Search mode must be `"client"` or `"server"`.
- Drag/drop implies the client runtime script is included.
- Server search implies the fragment endpoint is mounted.

#### Helpful error aggregation

Do not stop at the first error if possible. `.build()` should aggregate all missing pieces:

```text
kanban.board("sales") is invalid:
  - no columns were registered
  - data.cards(fn) is required
  - features.dragDrop() requires actions.cardMoved(fn)
```

This is much better for intern onboarding than one-at-a-time failures.

#### Method families

Recommended builder method families:

```typescript
kanban.board(id)
  .title(title)
  .description(text)
  .theme(nameOrOptions)
  .className(className)
  .attrs(attrs)
  .columns(fn)
  .data(fn)
  .features(fn)
  .render(fn)
  .actions(fn)
  .mount(app, prefix)       // convenience: build + mount if not built? see below
  .build()
```

Sub-builders:

```typescript
ColumnListBuilder
  .column(id) -> ColumnBuilder

ColumnBuilder
  .title(title)
  .description(text)
  .limit(n)
  .terminal(bool)
  .className(className)
  .attrs(attrs)
  .done() -> ColumnListBuilder

DataBuilder
  .cards(fn)
  .id(fn)
  .column(fn)
  .position(fn)
  .searchText(fn)
  .defaultFields({ id, column, position, title })

FeatureBuilder
  .search(options?)
  .preciseMove(options?)
  .dragDrop(options?)
  .createCard(options?)
  .cardMenu(options?)
  .readOnly()

RenderBuilder
  .card(fn)
  .columnHeader(fn)
  .toolbar(fn)
  .emptyColumn(fn)
  .boardShell(fn)       // advanced: wrap generated columns

ActionBuilder
  .cardMoved(fn)
  .cardCreated(fn)
  .cardUpdated(fn)
  .cardDeleted(fn)
  .cardClicked(fn)
  .cardMenuAction(fn)
  .custom(name, fn)
```

#### Build versus mount ergonomics

Prefer explicit `.build()` for clarity:

```javascript
const board = kanban.board("trail").columns(...).data(...).actions(...).build();
board.mount(app, "/_kanban");
```

But allow a convenience method only if it still validates:

```javascript
kanban.board("trail")
  .columns(...)
  .data(...)
  .actions(...)
  .mount(app, "/_kanban"); // internally calls build once
```

If `.mount(...)` auto-builds, document that it returns the built `Board`, not the mutable builder.

#### Low-level object API as an escape hatch

The large object literal API can remain as:

```javascript
kanban.fromSpec("trail", spec).build()
```

or:

```javascript
kanban.boardFromSpec("trail", spec)
```

But it should be described as advanced/compatibility API. Interns and examples should use the fluid builder.

### Legacy/low-level object authoring API

```javascript
const kanban = require("kanban.dsl");

const board = kanban.board("trail-notes", {
  columns: [
    { id: "todo", title: "To Do" },
    { id: "progress", title: "In Progress" },
    { id: "done", title: "Done" },
    { id: "someday", title: "Someday" },
  ],
  cards(ctx) {
    return db.query("SELECT * FROM cards ORDER BY position, id");
  },
  cardId(card) {
    return String(card.id);
  },
  cardColumn(card) {
    return card.status;
  },
  cardSearchText(card) {
    return `${card.title} ${card.description} ${card.tag}`;
  },
  renderCard(card, ctx) {
    return ui.fragment(
      ui.h3(card.title),
      ui.p(card.description)
    );
  },
  actions: {
    cardMoved(event) {
      moveCard({
        id: event.cardId,
        toStatus: event.to.columnId,
        toIndex: event.to.index,
      });
      return { ok: true, refresh: true };
    },
  },
});

board.mount(app, "/_kanban");

app.get("/", (req, res) => {
  res.html(ui.page({ title: "Board" }, board.render({ query: req.query })));
});
```

### Board options

```typescript
type BoardOptions = {
  title?: string;
  columns: ColumnSpec[];

  // Data access.
  cards: (ctx: BoardContext) => any[];
  cardId?: (card: any) => string;
  cardColumn?: (card: any) => string;
  cardPosition?: (card: any) => number;
  cardSearchText?: (card: any) => string;

  // Rendering hooks.
  renderCard?: (card: any, ctx: CardRenderContext) => ui.Node;
  renderColumnHeader?: (column: ColumnSpec, ctx: ColumnRenderContext) => ui.Node;
  renderToolbar?: (ctx: BoardRenderContext) => ui.Node;
  renderEmptyColumn?: (column: ColumnSpec, ctx: ColumnRenderContext) => ui.Node;

  // Feature flags.
  features?: {
    search?: boolean | SearchOptions;
    dragDrop?: boolean | DragDropOptions;
    preciseMove?: boolean | MoveOptions;
    createCard?: boolean | CreateCardOptions;
    cardMenu?: boolean | CardMenuOptions;
  };

  // Server-side callbacks invoked by browser events.
  actions?: BoardActions;

  // Styling/theming.
  className?: string;
  theme?: "field-notes" | "plain" | "minimal" | string;
  attrs?: Record<string, any>;
};
```

### Column specs

```typescript
type ColumnSpec = {
  id: string;
  title: string;
  description?: string;
  limit?: number;
  className?: string;
  attrs?: Record<string, any>;
};
```

### Action callbacks

```typescript
type BoardActions = {
  cardMoved?: (event: CardMovedEvent) => ActionResult;
  cardCreated?: (event: CardCreatedEvent) => ActionResult;
  cardUpdated?: (event: CardUpdatedEvent) => ActionResult;
  cardDeleted?: (event: CardDeletedEvent) => ActionResult;
  cardClicked?: (event: CardClickedEvent) => ActionResult;
  cardMenuAction?: (event: CardMenuActionEvent) => ActionResult;
  searchChanged?: (event: SearchChangedEvent) => ActionResult;
};
```

Example callback:

```javascript
actions: {
  cardMoved(event) {
    // event.from.columnId, event.from.index, event.to.columnId, event.to.index
    // event.visibleCardIds is useful for custom reordering.
    db.exec("UPDATE cards SET status = ? WHERE id = ?", event.to.columnId, event.cardId);
    renumberColumn(event.to.columnId, event.visibleCardIds);
    if (event.from.columnId !== event.to.columnId) renumberColumn(event.from.columnId);
    return { ok: true, toast: "Moved card", refresh: true };
  }
}
```

### Action result

```typescript
type ActionResult = {
  ok?: boolean;
  refresh?: boolean;
  html?: string;       // server-rendered board fragment, optional
  patch?: DOMPatch[];  // future extension
  toast?: string;
  error?: string;
  data?: any;
};
```

V1 recommendation:

- Callback returns `{ ok: true, refresh: true }`.
- The `kanban.dsl` action dispatcher rerenders the board fragment and returns `{ ok: true, html, toast }`.
- Browser replaces board root HTML.

## Callback protocol

### Browser-to-server action envelope

Every browser action posts a JSON envelope:

```json
{
  "boardId": "trail-notes",
  "action": "cardMoved",
  "eventId": "01HV...",
  "cardId": "42",
  "from": {
    "columnId": "todo",
    "index": 1
  },
  "to": {
    "columnId": "done",
    "index": 0
  },
  "visibleCardIds": ["42", "7", "8"],
  "query": {
    "search": "permit",
    "status": ""
  },
  "payload": {}
}
```

Fields:

- `boardId`: stable board ID selected by server-side author.
- `action`: callback name.
- `eventId`: client-generated ID for idempotency/debugging.
- `cardId`: primary card ID for card actions.
- `from`: source column/index as observed in DOM.
- `to`: destination column/index as observed in DOM.
- `visibleCardIds`: ordering after the interaction, useful for filtered views and custom ordering.
- `query`: current search/filter UI state.
- `payload`: action-specific data.

### Server response envelope

```json
{
  "ok": true,
  "boardId": "trail-notes",
  "html": "<div class=\"kb-board\" ...>...</div>",
  "toast": "Moved card",
  "version": 12
}
```

Error response:

```json
{
  "ok": false,
  "error": "Cannot move card into Done until checklist is complete",
  "html": "<div class=\"kb-board\" ...>...</div>"
}
```

The browser runtime should display the error and refresh the board if HTML is present.

## HTTP routes mounted by `board.mount(app, prefix)`

Assume:

```javascript
board.mount(app, "/_kanban");
```

Routes:

```text
GET  /_kanban/client.js
GET  /_kanban/:boardId/fragment?search=&status=&tag=
POST /_kanban/:boardId/action/:action
```

### `GET /_kanban/client.js`

Serves the generic Kanban browser runtime.

- Content-Type: `application/javascript; charset=utf-8`
- Cache: during development no-cache; later hash/version it.

### `GET /_kanban/:boardId/fragment`

Rerenders only the board root.

Use cases:

- live search with server-side filtering,
- post-action refresh,
- manual refresh.

### `POST /_kanban/:boardId/action/:action`

Dispatches browser action to server callback.

Flow:

```text
HTTP request
  -> Express route handler from board.mount
  -> board.dispatch(req.params.action, req.body)
  -> find registered callback
  -> invoke callback in current Goja runtime
  -> rerender board fragment if requested
  -> res.json(response)
```

## How mounting can work without changing `web.Host`

`board.mount(app, prefix)` can use the existing Express-style API. The Go implementation receives the JS `app` object and calls its route methods.

Pseudocode in Go:

```go
func (b *Board) Mount(vm *goja.Runtime, app *goja.Object, prefix string) error {
  getFn, ok := goja.AssertFunction(app.Get("get"))
  if !ok { return fmt.Errorf("kanban.board.mount: app.get missing") }
  postFn, ok := goja.AssertFunction(app.Get("post"))
  if !ok { return fmt.Errorf("kanban.board.mount: app.post missing") }

  clientHandler := vm.ToValue(func(req, res goja.Value) error {
    // res.type("application/javascript").send(clientRuntime)
  })
  fragmentHandler := vm.ToValue(func(req, res goja.Value) error {
    // render fragment with req.query
  })
  actionHandler := vm.ToValue(func(req, res goja.Value) error {
    // dispatch req.params.action + req.body
  })

  _, err := getFn(app, vm.ToValue(prefix+"/client.js"), clientHandler)
  if err != nil { return err }
  _, err = getFn(app, vm.ToValue(prefix+"/:boardId/fragment"), fragmentHandler)
  if err != nil { return err }
  _, err = postFn(app, vm.ToValue(prefix+"/:boardId/action/:action"), actionHandler)
  return err
}
```

This avoids a new native route type in `pkg/web.Host` for v1.

## Go package design

Recommended package:

```text
pkg/kanbanddsl/
  registrar.go       # RuntimeModuleRegistrar for require("kanban.dsl")
  module.go          # Loader and JS exports
  board.go           # Board model and options decoding
  render.go          # Render board/columns/cards into uidsl.Node
  mount.go           # board.mount(app, prefix)
  dispatch.go        # action envelope dispatch and callback invocation
  client_runtime.go  # embedded/generator string for browser runtime
  types.go           # Go structs for specs/events/results
  render_test.go
  dispatch_test.go
  runtime_test.go
```

### Registrar

`kanban.dsl` should follow the same runtime module registrar pattern as `ui.dsl`.

```go
type Registrar struct {
  Renderer func(*goja.Runtime, goja.Value) (string, error)
}

func NewRegistrar() *Registrar { return &Registrar{} }
func (r *Registrar) ID() string { return "kanban-dsl" }
func (r *Registrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
  runtime := NewRuntime(ctx, r.Renderer)
  reg.RegisterNativeModule("kanban.dsl", runtime.Loader)
  ctx.SetValue("kanban.dsl", runtime)
  return nil
}
```

`pkg/app/server.go` would then add it alongside UI and Express:

```go
WithRuntimeModuleRegistrars(
  web.NewExpressRegistrar(host),
  uidsl.NewRegistrar(),
  kanbanddsl.NewRegistrar(),
)
```

### Runtime state

The module needs per-runtime state:

```go
type Runtime struct {
  ctx      *engine.RuntimeModuleContext
  boardsMu sync.RWMutex
  boards   map[string]*Board
  renderer Renderer
}
```

Store boards by ID. Board IDs must be unique per runtime.

### Board model

```go
type Board struct {
  ID string
  Options BoardOptions
  Runtime *Runtime

  cardsFn goja.Callable
  renderCardFn goja.Callable
  callbacks map[string]goja.Callable
}
```

The Go board stores Goja callables. This is safe only if all calls happen on the runtime owner. Because route handlers are already invoked through `Host.ServeHTTP -> owner.Call`, calls from mounted Express handlers should be on the owner path. Document and test this.

### Options decoding

Be conservative. Start by reading properties manually instead of over-decoding large maps.

```go
func newBoard(vm *goja.Runtime, id string, options goja.Value) (*Board, error) {
  obj := options.ToObject(vm)
  columns := decodeColumns(vm, obj.Get("columns"))
  cardsFn := mustFunction(obj.Get("cards"), "cards")
  renderCardFn := optionalFunction(obj.Get("renderCard"))
  actions := decodeActions(vm, obj.Get("actions"))
  return &Board{...}, nil
}
```

### Board JS object

`kanban.board(...)` returns a JS object with methods:

```javascript
board.id
board.render(ctx?)
board.fragment(ctx?)
board.mount(app, prefix?)
board.dispatch(action, envelope)
board.clientScriptURL()
```

Implementation:

```go
func (b *Board) JSObject(vm *goja.Runtime) *goja.Object {
  obj := vm.NewObject()
  _ = obj.Set("id", b.ID)
  _ = obj.Set("render", func(call goja.FunctionCall) goja.Value { ... })
  _ = obj.Set("fragment", func(call goja.FunctionCall) goja.Value { ... })
  _ = obj.Set("mount", func(app goja.Value, prefix string) error { ... })
  _ = obj.Set("dispatch", func(action string, envelope goja.Value) (map[string]any, error) { ... })
  return obj
}
```

## Rendering design

### Board HTML structure

`kanban.dsl` should generate stable class names and data attributes. Designers can override CSS, but the runtime depends on `data-*` hooks.

```html
<div class="kb-board kb-theme-field-notes" data-kb-board-id="trail-notes" data-kb-prefix="/_kanban">
  <form class="kb-search" data-kb-search-form>
    <input name="search" data-kb-search placeholder="Search...">
  </form>
  <div class="kb-columns">
    <section class="kb-column" data-kb-column-id="todo">
      <header class="kb-column-header">
        <h2>To Do</h2>
        <span data-kb-count="todo">3</span>
      </header>
      <div class="kb-card-list" data-kb-drop-zone data-kb-column-id="todo">
        <article class="kb-card" draggable="true" data-kb-card-id="42" data-kb-column-id="todo" data-kb-search-text="...">
          <!-- custom renderCard content -->
        </article>
      </div>
    </section>
  </div>
</div>
<script src="/_kanban/client.js" defer></script>
```

### Rendering pipeline

```text
board.render(ctx)
  -> normalize ctx/query
  -> call cards(ctx)
  -> filter/sort/group cards
  -> render board shell
  -> for each column render column header
  -> for each card call renderCard(card, cardCtx)
  -> wrap renderCard output in kb-card chrome and data attrs
  -> include script for client runtime if enabled
```

Pseudocode:

```go
func (b *Board) Render(ctx RenderContext) (uidsl.Node, error) {
  cards, err := b.LoadCards(ctx)
  if err != nil { return nil, err }
  grouped := b.GroupCards(cards, ctx)

  columns := []uidsl.Node{}
  for _, col := range b.Columns {
    cardNodes := []uidsl.Node{}
    for i, card := range grouped[col.ID] {
      cardBody, err := b.RenderCard(card, CardContext{Column: col, Index: i})
      if err != nil { return nil, err }
      cardNodes = append(cardNodes, wrapCard(card, cardBody, col, i))
    }
    columns = append(columns, renderColumn(col, cardNodes))
  }

  return &uidsl.Element{Tag: "div", Attrs: boardAttrs, Children: columns}, nil
}
```

### Mixing with `ui.dsl`

A user-provided `renderCard` returns any `ui.dsl` node/value:

```javascript
renderCard(card, ctx) {
  return ui.fragment(
    ui.h3(card.title),
    ui.p(card.description),
    ui.div({ class: "card-controls" }, ctx.moveSelect())
  );
}
```

`kanban.dsl` should provide helper components on `ctx`:

```typescript
type CardRenderContext = {
  boardId: string;
  column: ColumnSpec;
  index: number;
  moveSelect(): ui.Node;
  menu(items: MenuItem[]): ui.Node;
  checkbox(options?: any): ui.Node;
  attrs(extra?: object): object;
};
```

The first implementation can skip `ctx.moveSelect()` and just wrap cards with default controls. Add helpers after render basics work.

## Browser runtime design

### Responsibilities

The browser runtime should:

- find boards with `[data-kb-board-id]`,
- bind search inputs,
- bind drag/drop handlers,
- bind move forms/buttons,
- send action envelopes to `/_kanban/:boardId/action/:action`,
- request fragments from `/_kanban/:boardId/fragment`,
- replace board HTML,
- rebind behavior after replacement,
- display errors/toasts minimally.

### Client runtime pseudocode

```javascript
(function () {
  function initBoard(root) {
    bindSearch(root);
    bindDragDrop(root);
    bindMoveForms(root);
  }

  async function postAction(root, action, envelope) {
    const boardId = root.dataset.kbBoardId;
    const prefix = root.dataset.kbPrefix || '/_kanban';
    const res = await fetch(`${prefix}/${encodeURIComponent(boardId)}/action/${encodeURIComponent(action)}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
      body: JSON.stringify({ boardId, action, eventId: crypto.randomUUID(), ...envelope }),
    });
    const data = await res.json();
    if (!data.ok) showError(root, data.error || 'Action failed');
    if (data.html) replaceBoard(root, data.html);
    return data;
  }

  function replaceBoard(root, html) {
    const tmp = document.createElement('template');
    tmp.innerHTML = html.trim();
    const next = tmp.content.firstElementChild;
    root.replaceWith(next);
    initBoard(next);
  }

  document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('[data-kb-board-id]').forEach(initBoard);
  });
})();
```

### Search modes

Support two modes:

1. Client filter mode: hide/show existing cards based on `data-kb-search-text`.
2. Server fragment mode: debounce and fetch `fragment?search=...`.

Board option:

```javascript
features: {
  search: { mode: "client" } // or "server"
}
```

V1 recommendation: default to `client`; support `server` soon after.

### Movement modes

Support two modes:

1. `form`: explicit destination/position controls.
2. `dragDrop`: drag/drop direct manipulation.

Board option:

```javascript
features: {
  preciseMove: true,
  dragDrop: true,
}
```

Use both by default for accessibility.

## Server callback dispatch design

### Dispatch flow

```text
POST /_kanban/trail-notes/action/cardMoved
  -> mounted Express handler
  -> board.Dispatch("cardMoved", req.body)
  -> find callback actions.cardMoved
  -> build event object
  -> call callback in Goja
  -> normalize result
  -> if result.refresh: board.RenderFragment(query)
  -> res.json(response)
```

### Event object shape

```javascript
{
  boardId: "trail-notes",
  action: "cardMoved",
  eventId: "...",
  cardId: "42",
  from: { columnId: "todo", index: 1 },
  to: { columnId: "done", index: 0 },
  visibleCardIds: ["42", "7", "8"],
  query: { search: "permit" },
  payload: {},
  now: "2026-05-03T...Z"
}
```

Do not pass raw HTTP request/response objects to board callbacks in v1. Keep callback events serializable.

### Callback result normalization

Accept these forms:

```javascript
return { ok: true, refresh: true };
return { ok: false, error: "Nope" };
return ui.div("custom response?"); // not recommended for action callbacks
return undefined; // treat as { ok: true, refresh: true }
```

Go normalizer:

```go
func normalizeActionResult(value goja.Value) ActionResult {
  if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
    return ActionResult{OK: true, Refresh: true}
  }
  if m, ok := value.Export().(map[string]any); ok { ... }
  return ActionResult{OK: true, Refresh: true, Data: value.Export()}
}
```

## Intelligent design considerations

### 1. Callback IDs vs callback functions

Do not expose raw callback function names to the browser beyond action names. The browser should post `action: "cardMoved"`, not arbitrary JS code or callback IDs.

Good:

```json
{ "action": "cardMoved", "cardId": "42" }
```

Bad:

```json
{ "callback": "function body here" }
```

The server owns the callback registry.

### 2. Versioning and stale boards

Each board render can include a version number:

```html
<div data-kb-board-id="trail-notes" data-kb-version="17">
```

The action envelope includes `version`. If the server detects a stale version, it can still apply the action or return a refresh.

V1 can include but not enforce versions.

### 3. Idempotency

Action envelopes include `eventId`. The server can log or ignore duplicates. V1 can pass it to callbacks without storing it. Later, a callback can store processed event IDs in the database.

### 4. Security

This is a trusted local scripting framework, but callbacks mutate state. Basic guardrails:

- same-origin requests only,
- JSON content type required for action endpoint,
- optional per-board nonce in `data-kb-token`,
- callback validates card IDs and column IDs,
- no arbitrary callback names beyond registered actions.

### 5. Accessibility

Drag/drop alone is not enough. Always provide explicit controls or keyboard alternatives when `preciseMove` is enabled.

### 6. Multi-board pages

Every DOM hook must be scoped under `[data-kb-board-id]`. The browser runtime must not query global `.kb-card` without a board root.

### 7. Multiple rendering styles

Separate behavior hooks from styling classes:

- Behavior: `data-kb-card-id`, `data-kb-drop-zone`.
- Styling: `kb-card`, custom classes, themes.

This lets designers reshape the board without breaking runtime behavior.

## Implementation phases

### Phase 1: Minimal module shell

Files:

- `pkg/kanbanddsl/registrar.go`
- `pkg/kanbanddsl/module.go`
- `pkg/kanbanddsl/board.go`
- `pkg/app/server.go`

Tasks:

1. Create package `pkg/kanbanddsl`.
2. Implement `NewRegistrar()` and register `kanban.dsl`.
3. Export `kanban.board(id, options)`.
4. Store boards in runtime-scoped registry.
5. Return a board JS object with `id` and `render()`.
6. Add registrar to `pkg/app/server.go`.
7. Add runtime smoke test for `require("kanban.dsl")`.

Acceptance criteria:

```javascript
const kanban = require("kanban.dsl");
const board = kanban.board("x", { columns: [], cards: () => [] });
board.id === "x";
```

### Phase 2: Render basic boards

Files:

- `pkg/kanbanddsl/render.go`
- `pkg/kanbanddsl/types.go`
- `pkg/kanbanddsl/render_test.go`

Tasks:

1. Decode columns.
2. Call `cards(ctx)`.
3. Group cards by column using `cardColumn` or `status` default.
4. Render board, columns, card wrappers, counts.
5. Call optional `renderCard(card, ctx)`.
6. Ensure returned `ui.dsl` nodes are preserved.
7. Add tests for custom card rendering.

Acceptance criteria:

- Board HTML contains `data-kb-board-id`.
- Columns contain `data-kb-column-id`.
- Cards contain `data-kb-card-id`.
- Custom `ui.dsl` returned by `renderCard` appears inside card wrapper.

### Phase 3: Mount routes and client runtime

Files:

- `pkg/kanbanddsl/mount.go`
- `pkg/kanbanddsl/client_runtime.go`
- `pkg/kanbanddsl/runtime_test.go`

Tasks:

1. Implement `board.mount(app, prefix)`.
2. Register `GET prefix/client.js`.
3. Register `GET prefix/:boardId/fragment`.
4. Register `POST prefix/:boardId/action/:action`.
5. Make `board.render()` include client script reference.
6. Add browser runtime that initializes boards.

Acceptance criteria:

- `/ _kanban/client.js` serves JavaScript.
- `/ _kanban/<board>/fragment` returns board HTML.
- Board page loads client runtime with no console errors.

### Phase 4: Action dispatch and callbacks

Files:

- `pkg/kanbanddsl/dispatch.go`
- `pkg/kanbanddsl/dispatch_test.go`

Tasks:

1. Decode action envelope.
2. Validate board ID and action name.
3. Find callback in `actions` map.
4. Invoke callback with normalized event object.
5. Normalize callback result.
6. Rerender fragment when `refresh` is true.
7. Return JSON response.

Acceptance criteria:

- Browser POST to `cardMoved` invokes server-side JS callback.
- Callback can update DB through closure.
- Response includes fresh board HTML.

### Phase 5: Search and movement features

Files:

- `pkg/kanbanddsl/client_runtime.go`
- example app scripts

Tasks:

1. Implement client search binding.
2. Implement precise move form behavior.
3. Implement drag/drop behavior.
4. Build action envelopes for `cardMoved`.
5. Refresh board fragment after action.
6. Keep non-JS fallback route examples documented.

Acceptance criteria:

- Search filters cards.
- Move form moves card to exact column/index.
- Drag/drop posts `cardMoved` action.
- Board refreshes after action.

### Phase 6: Replace hand-written Kanban example

Files:

- `examples/kanban/scripts/app.js`
- `examples/kanban/README.md`

Tasks:

1. Rewrite current example to use `kanban.dsl`.
2. Keep the Field Notes visual theme.
3. Keep custom `renderCard` with `ui.dsl`.
4. Keep database persistence callbacks in server-side JS.
5. Remove app-specific `clientScript()`.
6. Add Playwright validation.

Acceptance criteria:

- Example app no longer hand-writes client drag/drop/search JS.
- Search works.
- Precise move works.
- CardMoved callback runs server-side and persists to SQLite.
- Playwright verifies no console errors.

## Example final Kanban app after DSL

```javascript
const db = require("database");
const express = require("express");
const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

const app = express.app();
app.static("/assets", "examples/kanban/assets");

migrate();
seedIfEmpty();

const board = kanban.board("trail-notes", {
  title: "Trail Notes: Cascade Loop",
  theme: "field-notes",
  columns: [
    { id: "todo", title: "To Do" },
    { id: "progress", title: "In Progress" },
    { id: "done", title: "Done" },
    { id: "someday", title: "Someday" },
  ],
  features: {
    search: true,
    preciseMove: true,
    dragDrop: true,
    createCard: true,
  },
  cards(ctx) {
    return listCards(ctx.query);
  },
  cardId(card) { return String(card.id); },
  cardColumn(card) { return card.status; },
  cardSearchText(card) { return `${card.title} ${card.description} ${card.tag}`; },
  renderCard(card, ctx) {
    return ui.fragment(
      ui.div({ class: "card-top" },
        ui.span({ class: "check" }, Number(card.done) ? "✓" : ""),
        ui.h3(card.title),
        ui.button({ class: "card-menu" }, "...")
      ),
      ui.p({ class: "desc" }, card.description),
      card.image ? ui.img({ class: "card-image", src: card.image, alt: "Trail map sketch" }) : null,
      ui.div({ class: "card-meta" },
        ui.span({ class: "tag" }, card.tag),
        card.due_date ? ui.time({ datetime: card.due_date }, card.due_date) : ui.span("")
      )
    );
  },
  actions: {
    cardMoved(event) {
      moveCard({ id: event.cardId, toStatus: event.to.columnId, toIndex: event.to.index });
      return { ok: true, refresh: true };
    },
    cardCreated(event) {
      createCard(event.input);
      return { ok: true, refresh: true };
    },
  },
});

board.mount(app, "/_kanban");

app.get("/", (req, res) => {
  res.html(ui.page({ title: "Trail Notes" },
    ui.link({ rel: "stylesheet", href: "/style.css" }),
    ui.main({ class: "page" },
      header(),
      board.render({ query: req.query }),
      footer()
    )
  ));
});
```

The custom app still controls:

- database schema,
- data access,
- card rendering,
- callback behavior,
- page shell,
- theme CSS.

The DSL controls:

- board markup conventions,
- search controls,
- move controls,
- drag/drop hooks,
- client runtime,
- callback protocol,
- fragment refresh.


## Example Kanban apps enabled by this DSL

This section gives concrete examples of the kinds of applications a developer should be able to create once `kanban.dsl` exists. The point is not that every app looks identical. The point is that the board mechanics are generic while data access, card rendering, page chrome, and callbacks stay app-specific server-side JavaScript.

### Example 1: Trail planning board with custom Field Notes cards

This is the direct evolution of the current example. The app stores trail-planning tasks in SQLite, renders a Field Notes-style board, uses a custom map image in one card, and persists moves through `cardMoved`.

```javascript
const db = require("database");
const express = require("express");
const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

const app = express.app();
app.static("/assets", "examples/trail-board/assets");

function migrate() {
  db.exec(`CREATE TABLE IF NOT EXISTS trail_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    notes TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    position INTEGER NOT NULL DEFAULT 0,
    tag TEXT NOT NULL DEFAULT 'Planning',
    due_date TEXT NOT NULL DEFAULT '',
    image TEXT NOT NULL DEFAULT ''
  )`);
}

function listCards(query = {}) {
  const where = [];
  const args = [];
  if (query.search) {
    const q = `%${String(query.search).toLowerCase()}%`;
    where.push(`(lower(title) LIKE ? OR lower(notes) LIKE ? OR lower(tag) LIKE ?)`);
    args.push(q, q, q);
  }
  const sql = `SELECT * FROM trail_cards ${where.length ? 'WHERE ' + where.join(' AND ') : ''} ORDER BY position, id`;
  return db.query(sql, ...args);
}

function moveTrailCard(cardId, toColumn, toIndex) {
  // Use the same normalize-column algorithm from the current app.
  const card = db.query("SELECT * FROM trail_cards WHERE id = ?", cardId)[0];
  if (!card) throw new Error(`missing card ${cardId}`);

  const destination = db.query(
    "SELECT * FROM trail_cards WHERE status = ? AND id != ? ORDER BY position, id",
    toColumn,
    cardId,
  );
  destination.splice(Math.max(0, Math.min(toIndex, destination.length)), 0, card);
  db.exec("UPDATE trail_cards SET status = ? WHERE id = ?", toColumn, cardId);
  destination.forEach((c, index) => {
    db.exec("UPDATE trail_cards SET position = ? WHERE id = ?", (index + 1) * 10, c.id);
  });
}

migrate();

const board = kanban.board("trail-notes", {
  title: "Trail Notes: Cascade Loop",
  theme: "field-notes",
  columns: [
    { id: "todo", title: "To Do" },
    { id: "progress", title: "In Progress" },
    { id: "done", title: "Done" },
    { id: "someday", title: "Someday" },
  ],
  features: {
    search: { mode: "client" },
    preciseMove: true,
    dragDrop: true,
  },
  cards(ctx) {
    return listCards(ctx.query);
  },
  cardId(card) {
    return String(card.id);
  },
  cardColumn(card) {
    return card.status;
  },
  cardSearchText(card) {
    return `${card.title} ${card.notes} ${card.tag}`;
  },
  renderCard(card, ctx) {
    return ui.fragment(
      ui.div({ class: "card-top" },
        ui.span({ class: "check" }, card.status === "done" ? "✓" : ""),
        ui.h3(card.title),
        ui.button({ class: "card-menu" }, "...")
      ),
      ui.p({ class: "desc" }, card.notes || ""),
      card.image ? ui.img({ class: "card-image", src: card.image, alt: "Trail map" }) : null,
      ui.div({ class: "card-meta" },
        ui.span({ class: "tag" }, card.tag),
        card.due_date ? ui.time({ datetime: card.due_date }, card.due_date) : ui.span("")
      )
    );
  },
  actions: {
    cardMoved(event) {
      moveTrailCard(event.cardId, event.to.columnId, event.to.index);
      return { ok: true, refresh: true, toast: "Moved trail note" };
    },
  },
});

board.mount(app, "/_kanban");

app.get("/", (req, res) => {
  res.html(ui.page({ title: "Trail Notes" },
    ui.link({ rel: "stylesheet", href: "/assets/field-notes.css" }),
    ui.main({ class: "page" },
      ui.header({ class: "hero" },
        ui.h1("Field Notes"),
        ui.p("Observations from the trail.")
      ),
      board.render({ query: req.query })
    )
  ));
});
```

What this example demonstrates:

- Custom card body with normal `ui.dsl` nodes.
- Static assets mixed into card rendering.
- `kanban.dsl` owns movement/search mechanics.
- The app owns the database schema and callback behavior.

### Example 2: Editorial publishing pipeline

A very different board can use the same mechanics. This one tracks articles through an editorial workflow. Cards show author, draft word count, priority, and publication date. Moving a card into `published` triggers a server-side callback that writes a publication timestamp and can run additional validation.

```javascript
const db = require("database");
const express = require("express");
const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

const app = express.app();

function migrate() {
  db.exec(`CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    summary TEXT,
    workflow_state TEXT NOT NULL DEFAULT 'ideas',
    priority TEXT NOT NULL DEFAULT 'normal',
    words INTEGER NOT NULL DEFAULT 0,
    position INTEGER NOT NULL DEFAULT 0,
    publish_at TEXT,
    published_at TEXT
  )`);
}

function listArticles(query = {}) {
  const where = [];
  const args = [];
  if (query.search) {
    const q = `%${String(query.search).toLowerCase()}%`;
    where.push(`(lower(title) LIKE ? OR lower(author) LIKE ? OR lower(summary) LIKE ?)`);
    args.push(q, q, q);
  }
  if (query.priority) {
    where.push("priority = ?");
    args.push(query.priority);
  }
  return db.query(
    `SELECT * FROM articles ${where.length ? 'WHERE ' + where.join(' AND ') : ''} ORDER BY position, id`,
    ...args,
  );
}

function moveArticle(event) {
  const article = db.query("SELECT * FROM articles WHERE id = ?", event.cardId)[0];
  if (!article) throw new Error(`article ${event.cardId} not found`);

  if (event.to.columnId === "published" && Number(article.words) < 500) {
    return { ok: false, error: "Cannot publish articles under 500 words." };
  }

  db.exec(
    "UPDATE articles SET workflow_state = ?, published_at = CASE WHEN ? = 'published' THEN CURRENT_TIMESTAMP ELSE published_at END WHERE id = ?",
    event.to.columnId,
    event.to.columnId,
    event.cardId,
  );
  renumberArticleColumn(event.to.columnId, event.cardId, event.to.index);
  return { ok: true, refresh: true, toast: `Moved to ${event.to.columnId}` };
}

migrate();

const editorialBoard = kanban.board("editorial", {
  title: "Editorial Pipeline",
  columns: [
    { id: "ideas", title: "Ideas" },
    { id: "drafting", title: "Drafting", limit: 5 },
    { id: "review", title: "Review" },
    { id: "scheduled", title: "Scheduled" },
    { id: "published", title: "Published" },
  ],
  features: {
    search: true,
    preciseMove: true,
    dragDrop: true,
    cardMenu: true,
  },
  cards(ctx) {
    return listArticles(ctx.query);
  },
  cardId(article) {
    return String(article.id);
  },
  cardColumn(article) {
    return article.workflow_state;
  },
  cardSearchText(article) {
    return `${article.title} ${article.author} ${article.summary} ${article.priority}`;
  },
  renderToolbar(ctx) {
    return ui.div({ class: "editorial-toolbar" },
      ui.input({ name: "search", placeholder: "Search articles...", "data-kb-search": true }),
      ui.select({ name: "priority", "data-kb-filter": "priority" },
        ui.option({ value: "" }, "All priorities"),
        ui.option({ value: "high" }, "High"),
        ui.option({ value: "normal" }, "Normal"),
        ui.option({ value: "low" }, "Low")
      )
    );
  },
  renderCard(article, ctx) {
    return ui.fragment(
      ui.h3(article.title),
      ui.p({ class: "summary" }, article.summary || "No summary yet."),
      ui.div({ class: "article-meta" },
        ui.span({ class: "author" }, `By ${article.author}`),
        ui.span({ class: ["priority", `priority-${article.priority}`] }, article.priority),
        ui.span({ class: "words" }, `${article.words} words`)
      ),
      article.publish_at ? ui.time({ datetime: article.publish_at }, `Scheduled ${article.publish_at}`) : null
    );
  },
  actions: {
    cardMoved(event) {
      return moveArticle(event);
    },
    cardMenuAction(event) {
      if (event.payload.name === "duplicate") {
        duplicateArticle(event.cardId);
        return { ok: true, refresh: true, toast: "Article duplicated" };
      }
      return { ok: false, error: "Unknown article action" };
    },
  },
});

editorialBoard.mount(app, "/_kanban");

app.get("/editorial", (req, res) => {
  res.html(ui.page({ title: "Editorial Pipeline" },
    ui.link({ rel: "stylesheet", href: "/editorial.css" }),
    ui.h1("Editorial Pipeline"),
    editorialBoard.render({ query: req.query })
  ));
});
```

What this example demonstrates:

- More than four columns.
- Custom toolbar/filter rendering.
- Validation inside a server-side `cardMoved` callback.
- Callback returns an error that the Kanban runtime should display in the browser.
- Card menus can map browser-side menu choices to server-side callbacks without exposing arbitrary code.

### Example 3: CRM sales pipeline with mixed UI widgets

A sales board needs different cards again. It tracks deals by stage, shows currency value, account owner, probability, and next action. It mixes `kanban.dsl` board mechanics with `ui.dsl` widgets such as progress bars, badges, and a side summary.

```javascript
const db = require("database");
const express = require("express");
const ui = require("ui.dsl");
const kanban = require("kanban.dsl");

const app = express.app();

function listDeals(query = {}) {
  const args = [];
  const where = [];
  if (query.search) {
    const q = `%${String(query.search).toLowerCase()}%`;
    where.push(`(lower(company) LIKE ? OR lower(contact) LIKE ? OR lower(next_action) LIKE ?)`);
    args.push(q, q, q);
  }
  return db.query(
    `SELECT * FROM deals ${where.length ? 'WHERE ' + where.join(' AND ') : ''} ORDER BY position, id`,
    ...args,
  );
}

function money(cents) {
  return `$${(Number(cents || 0) / 100).toLocaleString(undefined, { minimumFractionDigits: 0 })}`;
}

function probabilityBar(percent) {
  percent = Math.max(0, Math.min(100, Number(percent || 0)));
  return ui.div({ class: "probability" },
    ui.div({ class: "probability-fill", style: { width: `${percent}%` } }),
    ui.span(`${percent}%`)
  );
}

const salesBoard = kanban.board("sales", {
  title: "Sales Pipeline",
  columns: [
    { id: "lead", title: "Lead" },
    { id: "qualified", title: "Qualified" },
    { id: "proposal", title: "Proposal" },
    { id: "negotiation", title: "Negotiation" },
    { id: "won", title: "Won" },
    { id: "lost", title: "Lost" },
  ],
  features: {
    search: { mode: "server" },
    preciseMove: true,
    dragDrop: true,
  },
  cards(ctx) {
    return listDeals(ctx.query);
  },
  cardId(deal) {
    return String(deal.id);
  },
  cardColumn(deal) {
    return deal.stage;
  },
  cardSearchText(deal) {
    return `${deal.company} ${deal.contact} ${deal.next_action}`;
  },
  renderCard(deal, ctx) {
    return ui.fragment(
      ui.h3(deal.company),
      ui.p({ class: "contact" }, deal.contact),
      ui.div({ class: "deal-row" },
        ui.strong(money(deal.value_cents)),
        ui.span({ class: "owner" }, deal.owner)
      ),
      probabilityBar(deal.probability),
      ui.p({ class: "next-action" }, deal.next_action || "No next action"),
      ui.time({ datetime: deal.next_action_date }, deal.next_action_date || "unscheduled")
    );
  },
  actions: {
    cardMoved(event) {
      db.exec("UPDATE deals SET stage = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", event.to.columnId, event.cardId);
      renumberDeals(event.to.columnId, event.cardId, event.to.index);

      if (event.to.columnId === "won") {
        db.exec("UPDATE deals SET closed_at = CURRENT_TIMESTAMP, probability = 100 WHERE id = ?", event.cardId);
      }
      if (event.to.columnId === "lost") {
        db.exec("UPDATE deals SET closed_at = CURRENT_TIMESTAMP, probability = 0 WHERE id = ?", event.cardId);
      }

      return { ok: true, refresh: true };
    },
  },
});

salesBoard.mount(app, "/_kanban");

app.get("/sales", (req, res) => {
  const deals = listDeals(req.query);
  const total = deals.reduce((sum, deal) => sum + Number(deal.value_cents || 0), 0);

  res.html(ui.page({ title: "Sales Pipeline" },
    ui.link({ rel: "stylesheet", href: "/sales.css" }),
    ui.main({ class: "sales-page" },
      ui.aside({ class: "sales-summary" },
        ui.h2("Pipeline"),
        ui.p(`${deals.length} active deals`),
        ui.strong(money(total))
      ),
      ui.section({ class: "sales-board" }, salesBoard.render({ query: req.query }))
    )
  ));
});
```

What this example demonstrates:

- Different domain model: deals instead of tasks.
- Six columns with business semantics.
- Server-side callback applies stage-specific business rules.
- Board can be embedded next to ordinary `ui.dsl` content such as summary panels.
- Custom card body can contain mini-widgets like progress bars.

### Example 4: Personal habit board with computed cards

Not all boards need one database row per card. A habit board can compute cards from daily records. Moving a card might mean recording a daily completion state rather than changing a `status` field.

```javascript
const habitsBoard = kanban.board("habits", {
  title: "Today",
  columns: [
    { id: "planned", title: "Planned" },
    { id: "done", title: "Done" },
    { id: "skipped", title: "Skipped" },
  ],
  cards(ctx) {
    const today = new Date().toISOString().slice(0, 10);
    return db.query(`
      SELECT h.id, h.name, h.category, h.target, COALESCE(l.state, 'planned') AS state
      FROM habits h
      LEFT JOIN habit_log l ON l.habit_id = h.id AND l.day = ?
      ORDER BY h.position, h.id
    `, today);
  },
  cardId(habit) {
    return String(habit.id);
  },
  cardColumn(habit) {
    return habit.state;
  },
  renderCard(habit) {
    return ui.fragment(
      ui.h3(habit.name),
      ui.p(`${habit.category} · target: ${habit.target}`)
    );
  },
  actions: {
    cardMoved(event) {
      const today = new Date().toISOString().slice(0, 10);
      db.exec(`
        INSERT INTO habit_log(habit_id, day, state)
        VALUES (?, ?, ?)
        ON CONFLICT(habit_id, day) DO UPDATE SET state = excluded.state
      `, event.cardId, today, event.to.columnId);
      return { ok: true, refresh: true };
    },
  },
});
```

What this example demonstrates:

- Cards can be computed projections, not direct rows.
- `cardMoved` can write to a log table instead of mutating the card source table.
- `kanban.dsl` is not tied to a fixed schema.


### Builder-style example snippets

The examples above are intentionally detailed, but new examples and tests should use the fluid builder API. These shorter snippets show how the same apps read when the builder is the primary interface.

#### Trail planning board, builder style

```javascript
const trailBoard = kanban
  .board("trail-notes")
  .title("Trail Notes: Cascade Loop")
  .theme("field-notes")
  .columns(cols => cols
    .column("todo").title("To Do").done()
    .column("progress").title("In Progress").done()
    .column("done").title("Done").terminal(true).done()
    .column("someday").title("Someday").done()
  )
  .data(data => data
    .cards(ctx => listCards(ctx.query))
    .id(card => String(card.id))
    .column(card => card.status)
    .position(card => Number(card.position || 0))
    .searchText(card => `${card.title} ${card.notes} ${card.tag}`)
  )
  .features(features => features
    .search({ mode: "client" })
    .preciseMove()
    .dragDrop()
  )
  .render(render => render
    .card((card, ctx) => ui.fragment(
      ui.div({ class: "card-top" },
        ui.span({ class: "check" }, card.status === "done" ? "✓" : ""),
        ui.h3(card.title),
        ui.button({ class: "card-menu" }, "...")
      ),
      ui.p({ class: "desc" }, card.notes || ""),
      card.image ? ui.img({ class: "card-image", src: card.image, alt: "Trail map" }) : null,
      ui.div({ class: "card-meta" },
        ui.span({ class: "tag" }, card.tag),
        card.due_date ? ui.time({ datetime: card.due_date }, card.due_date) : ui.span("")
      )
    ))
  )
  .actions(actions => actions
    .cardMoved(event => {
      moveTrailCard(event.cardId, event.to.columnId, event.to.index);
      return { ok: true, refresh: true, toast: "Moved trail note" };
    })
  )
  .build();
```

This gives the Go side strong checkpoints:

- `.columns(...)` can reject duplicate IDs.
- `.features(...).dragDrop()` can require `.actions(...).cardMoved(...)`.
- `.data(...)` can ensure there is a card source and identity mapping.
- `.build()` can fail before any HTTP routes are mounted.

#### Editorial board, builder style

```javascript
const editorialBoard = kanban
  .board("editorial")
  .title("Editorial Pipeline")
  .columns(cols => cols
    .column("ideas").title("Ideas").done()
    .column("drafting").title("Drafting").limit(5).done()
    .column("review").title("Review").done()
    .column("scheduled").title("Scheduled").done()
    .column("published").title("Published").terminal(true).done()
  )
  .data(data => data
    .cards(ctx => listArticles(ctx.query))
    .id(article => String(article.id))
    .column(article => article.workflow_state)
    .position(article => Number(article.position || 0))
    .searchText(article => `${article.title} ${article.author} ${article.summary} ${article.priority}`)
  )
  .features(features => features
    .search({ mode: "server" })
    .preciseMove()
    .dragDrop()
    .cardMenu()
  )
  .render(render => render
    .toolbar(ctx => ui.div({ class: "editorial-toolbar" },
      ui.input({ name: "search", placeholder: "Search articles...", "data-kb-search": true }),
      ui.select({ name: "priority", "data-kb-filter": "priority" },
        ui.option({ value: "" }, "All priorities"),
        ui.option({ value: "high" }, "High"),
        ui.option({ value: "normal" }, "Normal"),
        ui.option({ value: "low" }, "Low")
      )
    ))
    .card((article, ctx) => ui.fragment(
      ui.h3(article.title),
      ui.p({ class: "summary" }, article.summary || "No summary yet."),
      ui.div({ class: "article-meta" },
        ui.span({ class: "author" }, `By ${article.author}`),
        ui.span({ class: ["priority", `priority-${article.priority}`] }, article.priority),
        ui.span({ class: "words" }, `${article.words} words`)
      )
    ))
  )
  .actions(actions => actions
    .cardMoved(event => moveArticle(event))
    .cardMenuAction(event => handleArticleMenu(event))
  )
  .build();
```

This example shows why builder validation matters. Since `.features(...).cardMenu()` was called, `.build()` should require `.actions(...).cardMenuAction(...)`. Since `drafting` has `.limit(5)`, the runtime can include that limit in event envelopes and optionally block moves client-side or server-side.

#### Sales board, builder style

```javascript
const salesBoard = kanban
  .board("sales")
  .title("Sales Pipeline")
  .columns(cols => cols
    .column("lead").title("Lead").done()
    .column("qualified").title("Qualified").done()
    .column("proposal").title("Proposal").done()
    .column("negotiation").title("Negotiation").done()
    .column("won").title("Won").terminal(true).done()
    .column("lost").title("Lost").terminal(true).done()
  )
  .data(data => data
    .cards(ctx => listDeals(ctx.query))
    .id(deal => String(deal.id))
    .column(deal => deal.stage)
    .position(deal => Number(deal.position || 0))
    .searchText(deal => `${deal.company} ${deal.contact} ${deal.next_action}`)
  )
  .features(features => features
    .search({ mode: "server" })
    .preciseMove()
    .dragDrop()
  )
  .render(render => render
    .card((deal, ctx) => ui.fragment(
      ui.h3(deal.company),
      ui.p({ class: "contact" }, deal.contact),
      ui.div({ class: "deal-row" },
        ui.strong(money(deal.value_cents)),
        ui.span({ class: "owner" }, deal.owner)
      ),
      probabilityBar(deal.probability),
      ui.p({ class: "next-action" }, deal.next_action || "No next action")
    ))
  )
  .actions(actions => actions
    .cardMoved(event => {
      moveDeal(event.cardId, event.to.columnId, event.to.index);
      if (event.to.columnId === "won") markDealWon(event.cardId);
      if (event.to.columnId === "lost") markDealLost(event.cardId);
      return { ok: true, refresh: true };
    })
  )
  .build();
```

This reads like a domain-specific declaration rather than a bundle of manual routes, DOM attributes, and browser code. The app still has full control over rendering and callbacks, but the Kanban interaction protocol is owned by `kanban.dsl`.

### Lessons from the examples

These examples imply several important API requirements:

- The DSL must not assume field names beyond defaults; every board can provide `cardId`, `cardColumn`, `cardPosition`, and `cardSearchText`.
- `renderCard` must accept arbitrary `ui.dsl` output.
- `cardMoved` must receive a rich event envelope rather than only `(cardId, columnId)`.
- The callback result must support both success and validation errors.
- The board must support custom toolbar rendering and ordinary `ui.dsl` page composition.
- The client runtime must be generic and data-attribute driven, not coupled to visual CSS classes.
- Fragment refresh is the safest v1 update model because it keeps rendering server-side and domain-specific.

## Testing strategy

### Go unit tests

- `TestBoardCreationRequiresUniqueID`
- `TestBoardRenderGroupsCardsByColumn`
- `TestBoardRenderUsesCustomRenderCard`
- `TestMountRegistersRoutes`
- `TestDispatchCallsCardMovedCallback`
- `TestDispatchReturnsRefreshedHTML`
- `TestClientRuntimeServedAsJavaScript`

### Runtime integration tests

Boot `go-go-goja` with:

- `ui.dsl`,
- `kanban.dsl`,
- `express`,
- test database or fake card provider.

Execute JavaScript that creates a board and mount routes. Use `httptest` to call fragment and action endpoints.

### Playwright tests

Scenarios:

1. Load board page and assert title/columns/cards.
2. Search for `permit`; assert only matching card visible.
3. Move `Book permit` to `Done`, index `0`; assert first card in Done.
4. Reload; assert persistence.
5. Drag `Gear check` to `Someday`; assert callback fired and DB changed.
6. Create card; assert callback fired and board refreshed.
7. Assert no console errors.

## Risks and mitigations

### Risk: callback dispatch becomes too magical

Mitigation: keep action names explicit and event envelopes visible. Document every callback shape.

### Risk: too much framework too early

Mitigation: implement v1 around current Kanban needs only: render, search, move, action dispatch, fragment refresh.

### Risk: Goja callable lifetimes

Callbacks are stored in board structs. They must only be invoked on the runtime owner. Add tests that dispatch through the HTTP path.

### Risk: HTML fragments and script reinitialization

Replacing a board fragment removes event listeners. The client runtime must re-run `initBoard(nextRoot)` after replacement.

### Risk: CSS coupling

Do not let client runtime depend on visual classes like `.field-card`. Depend on `data-kb-*` attributes.

### Risk: stale or filtered move ordering

When search is active, visible indices may not equal full column indices. The event should include both visible order and destination column. The callback decides how to interpret it. Default helper can move within full column by index when no search is active, and append/relative move when search is active.

### Risk: security

V1 is trusted local scripts. Still, add a board token or same-origin nonce later. Never allow browser to specify arbitrary callback source.

## Open questions

1. Should `kanban.dsl` be implemented entirely in Go or as a JS library loaded into Goja?
   - Recommendation: Go native module for callback registry and tight integration with `ui.dsl`/Go rendering.
2. Should client runtime be per-board generated or static?
   - Recommendation: static generic runtime, with behavior controlled by `data-kb-*` attributes.
3. Should actions always refresh HTML or support JSON patches?
   - Recommendation: refresh HTML in v1; patches later.
4. Should `board.mount(app)` call `app.get/post` internally or require explicit route wiring?
   - Recommendation: internal mounting for ergonomics, using existing Express app object.
5. Should card rendering happen in Go callbacks on every fragment request?
   - Yes. This is the price of flexible server-side rendering and is acceptable for small boards.

## File references

- `pkg/uidsl/node.go:3-32`: current HTML AST node model reused by Kanban rendering.
- `pkg/uidsl/module.go:15-36`: current `ui.dsl` registration and helper exports.
- `pkg/uidsl/module.go:39-49`: tag constructor implementation pattern.
- `pkg/uidsl/render.go:13-23`: `RenderAny` entrypoint for converting nodes to HTML.
- `pkg/uidsl/render.go:74-125`: node rendering logic.
- `pkg/uidsl/render.go:128-160`: attribute rendering and escaping.
- `pkg/web/express_module.go:17-23`: Express runtime module registration.
- `pkg/web/express_module.go:31-50`: Express app methods and static mounting API.
- `pkg/web/host.go:44-50`: static request dispatch.
- `pkg/web/host.go:66-81`: safe request dispatch into Goja via runtime owner.
- `pkg/web/request_response.go:16-42`: JavaScript request object contract.
- `pkg/web/request_response.go:87-102`: JavaScript response object contract.
- `pkg/app/server.go:55-72`: runtime composition point where `kanban.dsl` registrar should be added.
- `examples/kanban/scripts/app.js:67-130`: current manual search and move helpers that `kanban.dsl` should replace.
- `examples/kanban/scripts/app.js:211-288`: current manual card/column/board rendering that `kanban.dsl` should simplify.
- `examples/kanban/scripts/app.js:290-379`: current hand-written browser runtime that `kanban.dsl` should eliminate from app code.
- `examples/kanban/scripts/app.js:384-436`: current routes and callbacks that should move behind `board.mount` and `actions`.

## Final recommendation

Build `kanban.dsl` as a runtime-scoped native Go module with a fluid builder API. The builder should collect board configuration through typed sub-builders, validate it strongly in Go at `.build()`, then freeze it into an immutable board object that uses `ui.dsl` nodes for rendering and the existing Express app object for mounting routes. The module should own the generic Kanban browser runtime and callback protocol. App authors should own data access, card rendering, and server-side callbacks.

The first implementation should be intentionally small but complete:

- `kanban.board(id, options)`
- `board.render(ctx)`
- `board.mount(app, prefix)`
- generic client runtime
- `cardMoved` callback
- fragment refresh after action
- search support
- precise move support

Only after that should we add richer callbacks, menu actions, create/edit modals, built-in themes, and patch-based updates.
