---
Title: Client Interactivity and UI DSL Redesign Guide
Ticket: GOJA-HOSTING-SITE
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
      Note: Current Kanban app limitations and target implementation surface
    - Path: pkg/uidsl/module.go
      Note: Current UI DSL module registration and tag constructor surface
    - Path: pkg/uidsl/render.go
      Note: Current safe HTML renderer and escaping behavior
    - Path: pkg/web/express_module.go
      Note: Current Express-style app API including app.static
    - Path: pkg/web/host.go
      Note: Current static dispatch and Goja runtime-owner request dispatch
    - Path: pkg/web/request_response.go
      Note: Current JS request and response API surface
ExternalSources: []
Summary: Design for adding search, precise card moves, and client-side behavior support to goja-site's server-rendered UI DSL.
LastUpdated: 2026-05-03T14:25:00-04:00
WhatFor: Guide an intern through redesigning the UI DSL and Kanban app so server-rendered pages can include small client-side behaviors.
WhenToUse: Use before implementing search, drag/drop or explicit move controls, browser-side JavaScript assets, JSON APIs, or client behavior helpers.
---


# Client Interactivity and UI DSL Redesign Guide

## Executive summary

The current `goja-site` Kanban application proves that Go can host a JavaScript website through `go-go-goja`, an Express-style HTTP module, a preconfigured SQLite database module, static asset serving, and a server-side HTML UI DSL. It is not yet a satisfying application. Search is missing, and moving a card is only a small checkbox-style form that advances the card to the next hard-coded column. The user now wants search and the ability to move cards to a specific location.

The right redesign is **not** to turn the server-side `ui.dsl` into React. The better design is a layered model:

1. Keep `ui.dsl` as a pure, safe server-side HTML builder.
2. Add a small **client behavior DSL** for declaring browser-side assets, actions, and data bindings.
3. Serve a tiny browser runtime (`/goja-site/client.js`) or app-specific script (`/app.js`) that reads declarative `data-*` attributes and calls JSON endpoints.
4. Extend the Express-style module and Kanban app with JSON endpoints for search and precise movement.
5. Treat drag/drop and richer interactions as progressive enhancement: the page still renders useful HTML without JavaScript, but JavaScript makes search and movement immediate.

The intern should think of this as moving from:

```text
Server JS returns HTML only
```

to:

```text
Server JS returns HTML + declarative behavior metadata + serves small browser JS + JSON endpoints
```

This preserves the simplicity of the current system while making the app actually interactive.

## Problem statement

The current Kanban board has two functional gaps:

1. **Search/filtering is not implemented.** The toolbar has a `Filter` button, but there is no input, no query parameter handling, no `/api/cards?search=...`, and no client-side filtering behavior.
2. **Card movement is not precise.** A card has a checkbox-like form that posts hidden fields to advance to the next column. The user cannot choose a destination column and position. There is no drag/drop support. There is no API that says “move card 4 to column `done` before card 9” or “move card 4 to index 2 in column `todo`.”

The current implementation also reveals a deeper DSL limitation. The server-side UI DSL can create HTML nodes and arbitrary attributes, but it does not have a first-class way to describe browser behavior such as:

- “When this input changes, query `/api/cards` and rerender the board.”
- “When this card is dropped into a column, POST a move command.”
- “This button opens a menu.”
- “This section should update from JSON without a full page reload.”

We need a client-side story that still fits the Goja-hosted, JavaScript-authored model.

## Current-state analysis with evidence

### The runtime is already capable of serving server-rendered JS websites

`pkg/app/server.go` builds the runtime with database modules, support modules, and runtime registrars. The server creates a `web.Host`, preconfigures `database` and `db`, enables `fs`, `path`, `time`, and `timer`, and registers the Express and UI DSL modules.

Important evidence:

- The host is constructed with `uidsl.RenderAny` as renderer (`pkg/app/server.go:55`).
- Preconfigured `database` and `db` modules are registered as native module specs (`pkg/app/server.go:56-70`).
- `fs`, `path`, `time`, and `timer` are selected through module middleware (`pkg/app/server.go:71`).
- `web.NewExpressRegistrar(host)` and `uidsl.NewRegistrar()` are registered with the go-go-goja factory (`pkg/app/server.go:72`).

This means adding client behavior should not require changing the runtime owner model. We can add modules and assets at the web/UI layer.

### The Express-style module is small but sufficient for JSON endpoints

`pkg/web/express_module.go` exports `require("express")`, `express.app()`, route methods, and `app.static(...)`.

Important evidence:

- `express` is registered as a runtime module (`pkg/web/express_module.go:17-23`).
- The app object exports HTTP methods (`pkg/web/express_module.go:31-43`).
- `app.static(prefix, dir)` maps a URL prefix to a filesystem directory (`pkg/web/express_module.go:44-50`).

`pkg/web/request_response.go` gives handlers useful request/response primitives.

Important evidence:

- `RequestDTO` includes `Method`, `URL`, `Path`, `Query`, `Params`, `Headers`, `Cookies`, `IP`, `Body`, and `RawBody` (`pkg/web/request_response.go:16-27`).
- `RequestDTO.Map()` converts those fields into lower-case JavaScript keys (`pkg/web/request_response.go:29-42`).
- `Response.JSObject` exports `status`, `set`, `type`, `json`, `send`, `html`, `redirect`, and `end` (`pkg/web/request_response.go:87-102`).

This is enough to implement:

```javascript
app.get('/api/cards', (req, res) => res.json(listCards(req.query)))
app.post('/api/cards/:id/move', (req, res) => res.json(moveCard(req.params.id, req.body)))
```

No Go backend change is strictly required for search or precise movement if form/JSON bodies are already parsed.

### Static files already work, so client-side JavaScript can be served today

`pkg/web/host.go` now has static mounts:

- `Host` stores `static []StaticMount` (`pkg/web/host.go:23-29`).
- `RegisterStatic(prefix, dir)` appends a `http.FileServer` mount (`pkg/web/host.go:39-42`).
- `ServeHTTP` checks static mounts before JS routes (`pkg/web/host.go:44-50`).

The Kanban app uses this already:

- `app.static("/assets", "examples/kanban/assets")` (`examples/kanban/scripts/app.js:5-6`).

Therefore an app-specific browser script can be added immediately:

```javascript
app.static('/assets', 'examples/kanban/assets');
app.get('/app.js', (req, res) => res.type('application/javascript').send(clientScript()));
```

The question is not whether client JavaScript is possible; it is how to make it pleasant and safe to author from the same server-side JavaScript code.

### The current UI DSL is a pure HTML AST renderer

`pkg/uidsl/module.go` registers `ui.dsl` and exports tag functions.

Important evidence:

- The registrar exposes both `ui.dsl` and `ui` aliases (`pkg/uidsl/module.go:15-18`).
- The tag list contains standard tags including `img`, `time`, SVG tags, layout tags, forms, and text tags (`pkg/uidsl/module.go:21`).
- Every tag constructor calls `elementFromCall` and returns an `*Element` (`pkg/uidsl/module.go:24-29`, `pkg/uidsl/module.go:39-49`).
- Helpers include `fragment`, `text`, `raw`, `render`, and `page` (`pkg/uidsl/module.go:30-36`).

`pkg/uidsl/render.go` renders the AST safely:

- `RenderAny` normalizes and renders arbitrary Goja values (`pkg/uidsl/render.go:13-23`).
- Text content is escaped (`pkg/uidsl/render.go:112-115`).
- Attributes are sorted and escaped (`pkg/uidsl/render.go:128-160`).
- `raw` is an explicit escape hatch (`pkg/uidsl/render.go:114-115`).

This purity is good. It makes HTML generation predictable and safe. The redesign should not pollute this layer with application-specific ideas like “Kanban card” or “move action.” Instead, add a parallel behavior layer.

### The current Kanban app is server-rendered and form-post based

The Kanban app currently defines its schema and board in one JavaScript file.

Important evidence:

- The schema has `status`, `position`, `tag`, `due_date`, `done`, and `image` (`examples/kanban/scripts/app.js:19-32`).
- Seeded cards are inserted with status and position (`examples/kanban/scripts/app.js:42-54`).
- `boardPage()` reads all cards with `SELECT * FROM cards ORDER BY position, id` (`examples/kanban/scripts/app.js:142-144`).
- The board renders each column with `columns.map(([status, label]) => columnView(cards, status, label))` (`examples/kanban/scripts/app.js:167`).
- Card movement is hard-coded by `nextStatus(status)` (`examples/kanban/scripts/app.js:127-132`).
- The card check form posts hidden `status` and `done` values (`examples/kanban/scripts/app.js:106-114`).
- The move endpoint updates only `status`, `done`, and `updated_at` (`examples/kanban/scripts/app.js:207-212`).

The schema has `position`, but the UI and API do not use it for precise movement. Search and filtering are absent.

## Gap analysis

### Functional gaps

Current behavior:

- Can render a board.
- Can add cards.
- Can advance a card to the next column.
- Can expose a JSON card list.

Needed behavior:

- Search cards by text/tag/status/date.
- Filter board without a full reload.
- Move a card to a chosen column.
- Move a card to a chosen position inside a column.
- Persist the new order.
- Preferably drag/drop cards, but provide a non-drag fallback.

### UI DSL gaps

The HTML DSL can render attributes like `data-card-id`, `draggable`, `aria-*`, and `data-action`. It does not need deep changes for static markup.

Missing design-level concepts:

- Page-level script registration.
- Safe inline JSON state bootstrapping.
- Declarative event/action metadata.
- A standard client runtime contract.
- Helpers for common browser behaviors such as fetch-submit, live search, and drag/drop.

### Server/API gaps

The Express-style server is sufficient for simple JSON endpoints, but the Kanban app needs new endpoints and conventions:

- `GET /api/cards?search=&status=&tag=`
- `POST /api/cards`
- `PATCH /api/cards/:id`
- `POST /api/cards/:id/move`
- possibly `POST /api/cards/reorder`

It also needs response/request support that is pleasant for API clients. Current JSON parsing exists, but method override or PATCH support should be tested in the browser. The route registrar already exports `patch` (`pkg/web/express_module.go:33`), so the main missing piece is application code and tests.

## Proposed architecture

### High-level model

```text
+-----------------------------------------------------------+
| Server-side JS app                                        |
|                                                           |
|  require("ui.dsl")      require("express")      db       |
|       |                       |                 |          |
|       v                       v                 v          |
|  HTML page + data-* attrs   JSON endpoints      SQLite     |
|       |                       ^                            |
|       v                       | fetch                      |
|  Browser receives HTML + /app.js                           |
+------------------------------|----------------------------+
                               |
                               v
+-----------------------------------------------------------+
| Browser                                                   |
|                                                           |
|  app.js / goja-site client runtime                        |
|   - live search input                                     |
|   - fetch JSON cards                                      |
|   - render/update board DOM                               |
|   - drag/drop or move controls                            |
|   - call /api/cards/:id/move                              |
+-----------------------------------------------------------+
```

The server remains the source of truth. The browser gets just enough JavaScript to make the board interactive.

### Redesign principle: split `ui.dsl` into three layers

#### Layer 1: `ui.dsl` — HTML structure

This remains what it is today: safe constructors for HTML nodes.

Example:

```javascript
ui.article({
  class: 'kanban-card',
  draggable: true,
  'data-card-id': card.id,
  'data-action': 'drag-card'
}, ...)
```

Layer 1 should be stable, simple, and boring.

#### Layer 2: `ui.behavior` or `ui.client` — declarative behavior helpers

Add a module or namespace that produces conventional attributes and page assets.

Possible API:

```javascript
const client = require('ui.client');

client.page({
  title: 'Trail Notes',
  scripts: [client.runtime(), client.module('/app.js')],
  state: { columns, initialCards },
}, body)
```

or, if keeping one module:

```javascript
ui.page({
  title: 'Trail Notes',
  scripts: [ui.script({ src: '/app.js', defer: true })],
  state: { columns, initialCards },
}, body)
```

Behavior attribute helpers:

```javascript
client.searchBox({
  target: '#board',
  endpoint: '/api/cards',
  param: 'search',
})

client.draggable({ type: 'card', id: card.id })
client.dropZone({ type: 'column', status: status })
client.action('move-card', { id: card.id })
```

These helpers should output ordinary attributes, not special Go objects:

```javascript
client.draggable({ type: 'card', id: card.id })
// returns:
// { draggable: true, 'data-drag-type': 'card', 'data-card-id': id }
```

This keeps rendering simple.

#### Layer 3: browser runtime — imperative behavior

A tiny browser script reads attributes and does the work:

- `input[data-live-search]` triggers debounced fetch.
- `[data-card-id][draggable=true]` participates in drag/drop.
- `[data-drop-status]` accepts drops.
- forms with `data-fetch-form` submit with `fetch` and update the board.

This script is regular browser JavaScript. It can be app-specific first, then generalized.

## Concrete Kanban redesign

### Data model

The current `cards` table has enough fields for a first implementation:

```sql
CREATE TABLE cards (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  description TEXT,
  status TEXT NOT NULL DEFAULT 'todo',
  position INTEGER NOT NULL DEFAULT 0,
  tag TEXT NOT NULL DEFAULT 'Planning',
  due_date TEXT NOT NULL DEFAULT '',
  done INTEGER NOT NULL DEFAULT 0,
  image TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
)
```

But the semantics of `position` need to be tightened.

Recommended invariant:

- Positions are per-column ordered numbers.
- Lower position means earlier in the column.
- After every move, normalize positions in the affected source and destination columns to `10, 20, 30, ...`.
- Using gaps lets the app insert between cards without renumbering every time, but renormalizing is simpler for an intern.

### Search API

Endpoint:

```javascript
app.get('/api/cards', (req, res) => {
  const search = String(req.query.search || '').trim().toLowerCase();
  const status = String(req.query.status || '').trim();
  const tag = String(req.query.tag || '').trim();
  res.json(listCards({ search, status, tag }));
});
```

Pseudocode:

```javascript
function listCards(filters) {
  const where = [];
  const args = [];

  if (filters.search) {
    where.push('(lower(title) LIKE ? OR lower(description) LIKE ? OR lower(tag) LIKE ?)');
    const q = `%${filters.search}%`;
    args.push(q, q, q);
  }

  if (filters.status) {
    where.push('status = ?');
    args.push(filters.status);
  }

  if (filters.tag) {
    where.push('tag = ?');
    args.push(filters.tag);
  }

  const sql = `SELECT * FROM cards ${where.length ? 'WHERE ' + where.join(' AND ') : ''} ORDER BY status, position, id`;
  return db.query(sql, args);
}
```

Important: `database.Query` flattens argument arrays today through the go-go-goja database module, but the intern should verify array argument behavior with a test before relying on `db.query(sql, args)`. If it does not flatten as expected, use variadic calls or explicit helper wrappers.

### Precise move API

Endpoint:

```javascript
app.post('/api/cards/:id/move', (req, res) => {
  const id = Number(req.params.id);
  const toStatus = String(req.body.toStatus || req.body.status || 'todo');
  const toIndex = Number(req.body.toIndex || 0);
  const result = moveCard({ id, toStatus, toIndex });
  res.json({ ok: true, card: result.card, cards: listCards({}) });
});
```

Payload:

```json
{
  "toStatus": "done",
  "toIndex": 2
}
```

Move algorithm:

```javascript
function moveCard({ id, toStatus, toIndex }) {
  const card = db.query('SELECT * FROM cards WHERE id = ?', id)[0];
  if (!card) throw new Error(`card ${id} not found`);

  const fromStatus = card.status;
  const done = toStatus === 'done' ? 1 : 0;

  // Remove from old ordered list.
  const destination = db.query(
    'SELECT * FROM cards WHERE status = ? AND id != ? ORDER BY position, id',
    toStatus, id
  );

  const clamped = Math.max(0, Math.min(toIndex, destination.length));
  destination.splice(clamped, 0, { ...card, status: toStatus, done });

  // Update moved card status first.
  db.exec(
    'UPDATE cards SET status = ?, done = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?',
    toStatus, done, id
  );

  // Renumber destination column.
  destination.forEach((c, index) => {
    db.exec('UPDATE cards SET position = ? WHERE id = ?', (index + 1) * 10, c.id);
  });

  // If moved across columns, renumber source column too.
  if (fromStatus !== toStatus) {
    const source = db.query('SELECT * FROM cards WHERE status = ? ORDER BY position, id', fromStatus);
    source.forEach((c, index) => {
      db.exec('UPDATE cards SET position = ? WHERE id = ?', (index + 1) * 10, c.id);
    });
  }

  return { card: db.query('SELECT * FROM cards WHERE id = ?', id)[0] };
}
```

### Browser behavior: live search

Markup:

```javascript
ui.input({
  id: 'search',
  class: 'search-input',
  placeholder: 'Search trail notes...',
  'data-live-search': true,
  'data-target': '#board',
  'data-endpoint': '/api/cards',
})
```

Browser pseudocode:

```javascript
const search = document.querySelector('[data-live-search]');
search.addEventListener('input', debounce(async () => {
  const q = search.value.trim();
  const cards = await fetchJSON(`/api/cards?search=${encodeURIComponent(q)}`);
  renderBoard(cards);
}, 150));
```

Rendering options:

1. Browser renders cards from JSON using client templates.
2. Browser asks server for an HTML fragment.
3. Browser does simple hide/show based on existing DOM.

Recommended first implementation: **simple hide/show for search**, then JSON rendering in phase 2.

Why: hide/show requires minimal new DSL support and keeps the server-rendered board as source of DOM.

Markup:

```javascript
ui.article({
  class: 'kanban-card',
  'data-card-id': card.id,
  'data-search-text': `${card.title} ${card.description} ${card.tag}`.toLowerCase(),
}, ...)
```

Client code:

```javascript
function applySearch(q) {
  q = q.toLowerCase();
  document.querySelectorAll('[data-card-id]').forEach(card => {
    card.hidden = q && !card.dataset.searchText.includes(q);
  });
  updateColumnCounts();
}
```

This can ship quickly and does not need an API. Later, implement server-backed search for large boards.

### Browser behavior: drag/drop precise movement

Markup for a card:

```javascript
ui.article({
  class: 'kanban-card',
  draggable: true,
  'data-card-id': card.id,
  'data-card-status': card.status,
}, ...)
```

Markup for a column/card list:

```javascript
ui.div({
  class: 'card-list',
  id: `column-${status}`,
  'data-drop-status': status,
}, filtered.map(cardView), ...)
```

Browser pseudocode:

```javascript
let draggedCard = null;

document.addEventListener('dragstart', event => {
  const card = event.target.closest('[data-card-id]');
  if (!card) return;
  draggedCard = card;
  event.dataTransfer.setData('text/plain', card.dataset.cardId);
});

document.addEventListener('dragover', event => {
  const list = event.target.closest('[data-drop-status]');
  if (!list) return;
  event.preventDefault();
  const before = cardAfterPointer(list, event.clientY);
  if (before) list.insertBefore(draggedCard, before);
  else list.appendChild(draggedCard);
});

document.addEventListener('drop', async event => {
  const list = event.target.closest('[data-drop-status]');
  if (!list || !draggedCard) return;
  event.preventDefault();

  const toStatus = list.dataset.dropStatus;
  const cards = [...list.querySelectorAll('[data-card-id]')];
  const toIndex = cards.indexOf(draggedCard);
  const id = draggedCard.dataset.cardId;

  const response = await fetch(`/api/cards/${id}/move`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ toStatus, toIndex }),
  });
  if (!response.ok) {
    location.reload();
    return;
  }
  updateColumnCounts();
});
```

### Browser behavior: non-drag fallback

Drag/drop can be hard to use on touch devices and with keyboards. Add a fallback select menu on each card:

```javascript
ui.form({ class: 'move-form', 'data-fetch-form': true, action: `/api/cards/${card.id}/move`, method: 'post' },
  ui.select({ name: 'toStatus' }, columns.map(...)),
  ui.input({ name: 'toIndex', type: 'number', min: 0, value: currentIndex }),
  ui.button('Move')
)
```

The browser runtime can intercept `data-fetch-form` and submit with fetch, while the server can also accept normal form posts.

## UI DSL redesign in detail

### Keep HTML constructors as-is

No breaking change should be made to:

```javascript
ui.div(attrs, ...children)
ui.article(attrs, ...children)
ui.page({ title }, ...children)
```

The current design is already capable of emitting all required markup. The redesign adds convenience helpers, not a replacement.

### Add page-level script/style/state support

Current `ui.page` treats `link`, `style`, `meta`, and `title` nodes as head nodes (`pkg/uidsl/module.go:51-71`). It does not treat `script` as a head tag because `headTags` currently only includes `meta`, `link`, `style`, and `title` (`pkg/uidsl/module.go:22`).

Recommendation:

1. Add `script` to `headTags` when it has `src`, `defer`, or `type: 'module'`.
2. Add a helper for safe bootstrapped JSON state.
3. Add an ergonomic helper for client scripts.

Proposed API:

```javascript
ui.page({ title: 'Trail Notes' },
  ui.clientState('kanbanInitialState', { columns, cards }),
  ui.clientScript('/app.js', { type: 'module' }),
  body
)
```

Rendered output:

```html
<script type="application/json" id="kanbanInitialState">{"columns":...}</script>
<script type="module" src="/app.js"></script>
```

Go implementation sketch:

```go
_ = exports.Set("clientScript", func(src string, options map[string]any) *Element {
  attrs := map[string]any{"src": src, "defer": true}
  for k, v := range options { attrs[k] = v }
  return &Element{Tag: "script", Attrs: attrs}
})

_ = exports.Set("clientState", func(id string, value goja.Value) (*Element, error) {
  data, err := json.Marshal(value.Export())
  if err != nil { return nil, err }
  return &Element{
    Tag: "script",
    Attrs: map[string]any{"type": "application/json", "id": id},
    Children: []Node{&RawHTML{Value: safeJSONScript(data)}},
  }, nil
})
```

Important: JSON inside script tags must escape `</script>` and related sequences. Use a dedicated helper, not plain `RawHTML` with arbitrary strings.

### Add behavior-attribute helpers

These helpers return attribute maps. The existing `ui.div({ ...attrs }, ...)` call can accept them if JavaScript spreads are used.

Proposed API:

```javascript
ui.attrs({ class: 'kanban-card' }, ui.behavior.draggableCard(card.id), ui.behavior.searchText(card))
```

Or simpler:

```javascript
const b = ui.behavior;

ui.article({
  class: 'kanban-card',
  ...b.draggable({ type: 'card', id: card.id }),
  ...b.searchText(`${card.title} ${card.description} ${card.tag}`),
}, ...)
```

Exports:

```javascript
b.draggable({ type, id })
b.dropZone({ type, status })
b.liveSearch({ target, param })
b.fetchForm({ target })
b.action(name, payload)
b.searchText(text)
```

Example output:

```javascript
b.dropZone({ type: 'column', status: 'done' })
// => {
//   'data-drop-type': 'column',
//   'data-drop-status': 'done'
// }
```

This is optional syntactic sugar. The intern can first write raw attributes directly in the Kanban app, then extract helpers after tests pass.

### Add an optional browser runtime asset

The DSL can include a standard client runtime:

```javascript
ui.clientRuntime()
```

Rendered as:

```html
<script src="/goja-site/client.js" defer></script>
```

Server support options:

1. Built-in Go handler always serves `/goja-site/client.js`.
2. Express module exposes `app.clientRuntime()` to register it.
3. App serves its own `/app.js` route first, then built-in runtime later.

Recommendation for next implementation: start with app-specific `/app.js`. Generalize to built-in runtime after the interaction patterns stabilize.

## Implementation plan

### Phase 1: Make current form behavior honest and useful

Goal: users can choose where a card goes without client JavaScript.

Tasks:

1. Add visible move form on each card:
   - destination column select,
   - position select or numeric input,
   - submit button.
2. Change `/cards/:id/move` to accept `toStatus` and `toIndex`.
3. Implement `moveCard({ id, toStatus, toIndex })`.
4. Normalize positions after moves.
5. Add search as a normal GET form:
   - `/ ? search=permit`
   - `boardPage(req.query)` filters SQL.
6. Keep this phase fully server-rendered.

Why first: it makes the app actually work even without client JS.

### Phase 2: Add JSON APIs

Goal: browser JS can update without full reload.

Tasks:

1. Add `GET /api/cards?search=&status=&tag=`.
2. Add `POST /api/cards/:id/move` returning JSON when `Accept: application/json` or path starts with `/api`.
3. Add `POST /api/cards` returning JSON.
4. Add API tests using `httptest` or Playwright request API.
5. Keep form endpoints for progressive enhancement.

### Phase 3: Add app-specific browser script

Goal: search and movement feel interactive.

Tasks:

1. Add route:

```javascript
app.get('/app.js', (req, res) => res.type('application/javascript').send(clientScript()));
```

2. Add `ui.script({ src: '/app.js', defer: true })` to `ui.page`.
3. Add `data-search-text` to cards.
4. Add live search hide/show.
5. Add drag/drop movement.
6. Add fetch submission for move forms.
7. Add Playwright tests for:
   - search hides/shows matching cards,
   - moving card to `Done` at top persists after reload,
   - no console errors.

### Phase 4: Extract reusable client DSL helpers

Goal: app code becomes nicer and reusable for other JS websites.

Tasks:

1. Add `ui.clientScript(src, options)`.
2. Add `ui.clientState(id, value)` with safe JSON escaping.
3. Add `ui.behavior` helper object or module.
4. Update Kanban to use helpers instead of raw `data-*` attributes.
5. Add tests for rendered script/state output.

### Phase 5: Optional built-in client runtime

Goal: a future app can opt into common behavior without writing browser JS from scratch.

Tasks:

1. Add `pkg/clientruntime/client.js` embedded in Go.
2. Add host route `/goja-site/client.js`.
3. Add `ui.clientRuntime()` helper.
4. Implement generic behaviors:
   - `[data-live-search]`,
   - `[data-fetch-form]`,
   - `[data-draggable]`,
   - `[data-drop-zone]`.
5. Keep app-specific custom code possible.

## Recommended first implementation sketch

### Update `boardPage` to accept filters

```javascript
function boardPage(query = {}) {
  const filters = normalizeFilters(query);
  const cards = listCards(filters);
  return ui.page({ title: 'Trail Notes: Cascade Loop' },
    ui.link({ rel: 'stylesheet', href: '/style.css' }),
    ui.script({ src: '/app.js', defer: true }),
    layout(cards, filters)
  );
}

app.get('/', (req, res) => res.html(boardPage(req.query)));
```

### Add search input

```javascript
ui.form({ class: 'search-form', method: 'get', action: '/' },
  ui.input({
    id: 'search',
    name: 'search',
    value: filters.search,
    placeholder: 'Search notes...',
    'data-live-search': true,
  }),
  ui.button({ type: 'submit' }, 'Search')
)
```

### Add precise move form

```javascript
function moveForm(card, index) {
  return ui.form({
    class: 'move-form',
    method: 'post',
    action: `/cards/${card.id}/move`,
    'data-fetch-form': true,
  },
    ui.select({ name: 'toStatus' }, columns.map(([value, label]) =>
      ui.option({ value, selected: value === card.status }, label)
    )),
    ui.select({ name: 'toIndex' }, positionsForColumn(card.status).map(i =>
      ui.option({ value: i, selected: i === index }, `#${i + 1}`)
    )),
    ui.button({ type: 'submit' }, 'Move')
  );
}
```

### Add app-specific client script

```javascript
function clientScript() {
  return `
    const debounce = (fn, ms) => {
      let t;
      return (...args) => { clearTimeout(t); t = setTimeout(() => fn(...args), ms); };
    };

    function applySearch(query) {
      query = query.toLowerCase().trim();
      document.querySelectorAll('[data-card-id]').forEach(card => {
        const text = card.dataset.searchText || '';
        card.hidden = query && !text.includes(query);
      });
      updateCounts();
    }

    function updateCounts() {
      document.querySelectorAll('[data-status]').forEach(column => {
        const count = column.querySelectorAll('[data-card-id]:not([hidden])').length;
        const badge = column.querySelector('[data-count]');
        if (badge) badge.textContent = String(count);
      });
    }

    document.addEventListener('input', event => {
      const input = event.target.closest('[data-live-search]');
      if (!input) return;
      applySearch(input.value);
    });
  `;
}
```

This first client script can be small. Do not start with a generic framework.

## Testing strategy

### Go tests

Add/extend tests for:

- `ui.clientState` JSON escaping if implemented.
- `ui.clientScript` rendering.
- route/request parsing for JSON move API.
- static script route if built into host.

Commands:

```bash
GOTOOLCHAIN=go1.26.2 go test ./... -count=1
```

### Playwright tests

Manual Playwright scenario:

1. Start server with a fresh DB.
2. Navigate to `/`.
3. Search for `permit`.
4. Assert only `Book permit` is visible.
5. Clear search.
6. Move `Book permit` to `Done` position `0`.
7. Assert it appears as first card in `Done`.
8. Reload page.
9. Assert the position persisted.
10. Assert browser console has no errors.

Pseudocode:

```javascript
await page.goto('http://127.0.0.1:60024/');
await page.getByPlaceholder('Search notes...').fill('permit');
await expect(page.getByText('Book permit')).toBeVisible();
await expect(page.getByText('Research campsites')).toBeHidden();

const card = page.locator('[data-card-id]', { hasText: 'Book permit' });
await card.locator('select[name="toStatus"]').selectOption('done');
await card.locator('select[name="toIndex"]').selectOption('0');
await card.getByRole('button', { name: 'Move' }).click();

const doneCards = page.locator('section[data-status="done"] article[data-card-id]');
await expect(doneCards.first()).toContainText('Book permit');
await page.reload();
await expect(doneCards.first()).toContainText('Book permit');
```

## Risks and tradeoffs

### Risk: inventing a framework too early

Do not build a large client framework before the Kanban interactions are implemented. First write app-specific browser JS. Extract generic helpers only after two or three behaviors are proven.

### Risk: unsafe inline script/state rendering

Inline JSON state is easy to get wrong. Never render raw `JSON.stringify` into `<script>` without escaping `</script>`. Prefer external `/app.js` and JSON endpoints initially.

### Risk: static paths are process-relative

Current `app.static('/assets', 'examples/kanban/assets')` is easy to demo but not robust. A future design should resolve static paths relative to the script directory or an explicit configured app root.

### Risk: drag/drop accessibility

Drag/drop is not enough. Keep a form/select fallback for keyboard and touch users.

### Risk: position race conditions

With multiple users, two moves can conflict. SQLite transactions should eventually wrap move operations. For the local demo, sequential request handling is acceptable, but the algorithm should be designed so it can be put inside a transaction later.

## File references

- `pkg/app/server.go:55-72`: runtime construction with renderer, database modules, support modules, and Express/UI registrars.
- `pkg/web/express_module.go:17-23`: registration of the Express-style module.
- `pkg/web/express_module.go:31-50`: exported app methods and `app.static`.
- `pkg/web/host.go:39-50`: static mount registration and dispatch.
- `pkg/web/host.go:60-81`: HTTP request dispatch into Goja through the runtime owner.
- `pkg/web/request_response.go:16-42`: request object fields and JavaScript map conversion.
- `pkg/web/request_response.go:87-102`: response methods exposed to JavaScript.
- `pkg/uidsl/module.go:15-21`: `ui.dsl` registration and exported tag list.
- `pkg/uidsl/module.go:24-36`: exported tag constructors and helpers.
- `pkg/uidsl/render.go:74-125`: HTML node rendering.
- `pkg/uidsl/render.go:128-160`: attribute rendering and escaping.
- `examples/kanban/scripts/app.js:19-36`: current card schema and migration.
- `examples/kanban/scripts/app.js:106-124`: current card view and limited movement control.
- `examples/kanban/scripts/app.js:127-132`: current next-status movement model.
- `examples/kanban/scripts/app.js:142-168`: current board render path.
- `examples/kanban/scripts/app.js:186-188`: current JSON card list endpoint.
- `examples/kanban/scripts/app.js:207-212`: current move endpoint that lacks precise position semantics.

## Bottom line

The UI DSL should be redesigned around **progressive client behavior**, not replaced with a client framework. The immediate implementation should:

1. Add server-rendered search and precise move forms.
2. Add JSON APIs for cards and moves.
3. Add an app-specific `/app.js` for live search and drag/drop.
4. Then extract small helpers such as `ui.clientScript`, `ui.clientState`, and `ui.behavior.*` after the app-specific code proves the shape.

This keeps `goja-site` aligned with its core promise: write small websites in JavaScript, persist them with SQLite, render them with a safe HTML DSL, and add just enough browser JavaScript when the interface needs direct manipulation.
