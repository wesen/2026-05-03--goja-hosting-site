---
Title: Investigation Diary
Ticket: KANBAN-DSL
Status: active
Topics:
    - go
    - goja
    - javascript
    - web
    - glazed
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Chronological investigation and design diary for the kanban.dsl ticket."
LastUpdated: 2026-05-03T15:30:00-04:00
WhatFor: "Record why kanban.dsl is designed as a fluid builder with server-side callbacks and a browser runtime."
WhenToUse: "Read before implementing pkg/kanbanddsl or migrating the Kanban example."
---

# Diary

## Goal

This diary records the design work for a Kanban-specific DSL that runs in server-side Goja JavaScript, renders flexible boards with `ui.dsl`, ships a generic browser runtime, and dispatches browser actions such as `cardMoved` back into server-side callbacks.

## Step 1: Created KANBAN-DSL ticket and architecture guide

I created a new docmgr ticket for `kanban.dsl` and wrote a long-form architecture and implementation guide. The design starts from the current `goja-site` implementation: `ui.dsl` renders safe HTML, `express` handles routes, and the Kanban example currently hand-writes search, movement, JSON endpoints, and client JavaScript.

The guide proposes a runtime-scoped native Go module, `require("kanban.dsl")`, that owns generic board mechanics while letting app authors keep control over data access, card rendering, and server-side callbacks.

### Prompt Context

**User prompt (verbatim):** "let's invent a kanban specific "kanban.dsl" module to avoid having to write client side JS in the app itself, that however leaves us a lot of flexibility in designing various kanban boards, and gives us the freedom to mix it with ui.dsl components too.

Create a new docmgr ticket for a kanban specific DSL that for example allows action callback (such as cardMoved, etc...) to be written in the server side goja JS, and the frontend manages to know how to call it.

This is quite special, so think hard and brainstorm wide and do some really intelligent work here.

 Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the                                                                                      
  system needed to understand what it is, with prose paragraphs and                                                                                                                                         
bullet                                                                                                                                                                                                      
  point sand pseudocode and diagrams and api references and                                                                                                                                                 
file                                                                                                                                                                                                        
                                                                                                                                                                                                            
references.                                                                                                                                                                                                 
    It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Design a special-purpose Kanban DSL module that removes hand-written client JS from apps while preserving flexible server-side rendering and callbacks.

**Inferred user intent:** The user wants a serious architecture that bridges browser Kanban interactions to server-side Goja callbacks in a reusable, ergonomic, and enforceable way.

**Commit (code):** N/A — documentation/design step.

### What I did

- Created ticket `KANBAN-DSL`.
- Created design doc `01-kanban-dsl-architecture-and-implementation-guide.md`.
- Inspected current implementation files for evidence and references.
- Wrote a proposed `kanban.dsl` architecture with rendering, mounting, callback protocol, browser runtime, phases, risks, and file references.

### Why

- The current Kanban example has too much hand-written app-specific browser code.
- The desired abstraction needs careful protocol design: browser events must map safely to server-side Goja callbacks.

### What worked

- The current `ui.dsl` and `express` surfaces are enough to build on.
- The server already serializes Goja request dispatch through the runtime owner.

### What didn't work

- No implementation was attempted in this design step.

### What I learned

- The key abstraction is not a fixed Kanban schema. It is a board runtime with data access hooks, render hooks, and callback hooks.

### What was tricky to build

- The browser must not know app-specific data structures, yet it must send enough information for precise movement. The event envelope therefore includes card ID, from/to columns and indices, visible card IDs, and query state.

### What warrants a second pair of eyes

- Callback protocol shape.
- Fragment refresh strategy.
- Whether mounting should call existing Express app methods or integrate deeper into `web.Host`.

### What should be done in the future

- Implement the module in phases from `tasks.md`.

### Code review instructions

- Start with the design document, then inspect `pkg/uidsl`, `pkg/web`, `pkg/app/server.go`, and `examples/kanban/scripts/app.js`.

### Technical details

Primary design doc:

```text
ttmp/2026/05/03/KANBAN-DSL--kanban-specific-dsl-for-goja-websites/design-doc/01-kanban-dsl-architecture-and-implementation-guide.md
```

## Step 2: Added examples and fluid builder API

After drafting the object-style API, the user asked for example apps and then requested a fluid builder API so Go can enforce the DSL more strongly. I expanded the document with multiple app examples and then updated the API recommendation to prefer a builder with typed sub-builders and `.build()` validation.

The most important change is that `kanban.board(id)` should return a mutable `BoardBuilder`, not immediately accept one huge options object. The builder collects columns, data hooks, render hooks, features, and action callbacks; `.build()` validates and freezes the board.

### Prompt Context

**User prompt (verbatim):** "show me a couple of example kanban apps i can create with thta DSL. Then add them to the doc."

**User prompt (verbatim):** "use a fluid builder api so that we can do some strong enforcement on the go side. Update the document"

**Assistant interpretation:** Expand the guide with concrete use cases, then reshape the proposed API around a chainable builder to support strong Go-side validation.

**Inferred user intent:** The user wants the DSL to feel ergonomic but also robust, with guardrails that catch mistakes before runtime interactions begin.

**Commit (code):** N/A — documentation/design step.

### What I did

- Added example Kanban apps:
  - trail planning / Field Notes board,
  - editorial publishing pipeline,
  - CRM sales pipeline,
  - personal habit board.
- Added builder-style versions of the trail, editorial, and sales examples.
- Added a detailed fluid builder API section.
- Added Go-side enforcement rules and type-state-like sub-builder design.
- Updated tasks to include builder implementation phases.

### Why

- A large object literal is too loose for a DSL that coordinates server callbacks and browser actions.
- Builder sub-objects let Go expose smaller method sets and produce better validation errors.

### What worked

- The builder API maps well onto Go structs: `BoardBuilder`, `ColumnListBuilder`, `ColumnBuilder`, `DataBuilder`, `FeatureBuilder`, `RenderBuilder`, and `ActionBuilder`.
- Feature methods can imply validation requirements, such as `dragDrop()` requiring `actions.cardMoved()` unless the board is read-only.

### What didn't work

- The object-literal examples still exist in the document as conceptual examples. They should be treated as legacy/low-level or rewritten during future cleanup.

### What I learned

- For this DSL, `.build()` is not cosmetic. It is the moment when the Go side can aggregate missing requirements, reject duplicates, and freeze callback registration.

### What was tricky to build

- JavaScript cannot enforce compile-time types, but Go can simulate type-state by returning different builder objects with different available methods.

### What warrants a second pair of eyes

- Whether `.mount(app, prefix)` should auto-build or require explicit `.build()`.
- Whether a low-level `fromSpec` escape hatch is worth exposing in v1.

### What should be done in the future

- Implement builder validation before implementing browser runtime details.
- Keep examples in builder style so interns copy the right API.

### Code review instructions

- Review the fluid builder section in the design doc first.
- Then review the implementation tasks in `tasks.md`.

### Technical details

The preferred final shape is:

```javascript
const board = kanban.board("trail-notes")
  .columns(cols => cols.column("todo").title("To Do").done())
  .data(data => data.cards(loadCards).id(card => String(card.id)).column(card => card.status))
  .features(features => features.search().preciseMove().dragDrop())
  .actions(actions => actions.cardMoved(onCardMoved))
  .build();
```

## Step 3: Implemented first `kanban.dsl` module shell and builder runtime

I implemented the first working version of `require("kanban.dsl")`. The implementation follows the updated design direction: importing the DSL gives app authors a fluid `kanban.board(id)` builder, and mounting a board lets the Go side inject and serve the generic browser runtime at `/<prefix>/client.js`.

### Prompt Context

**User prompt (verbatim):** "Implement"

**User prompt (verbatim):** "keep a diary, commit at appropriate intervals"

**User prompt (verbatim):** "importing the dsl also means that the go side can inject / serve custom frontend JS that the app itself doesn't need to write"

**Assistant interpretation:** Start with the reusable native module, keep chronological notes, and make sure the design principle about Go-owned frontend runtime is implemented rather than left as prose.

**Inferred user intent:** The app should not need to define `/app.js` for standard Kanban behaviors. Once the script imports `kanban.dsl` and calls `board.mount(app, prefix)`, the Go module should register the browser runtime route and action routes.

**Commit (code):** pending at time of writing this diary step.

### What I did

- Added `pkg/kanbanddsl`.
- Implemented `NewRegistrar()` and registered native module `kanban.dsl`.
- Wired `kanbanddsl.NewRegistrar()` into `pkg/app/server.go`.
- Added `kanban.board(id)` returning a mutable fluid builder.
- Added builder families:
  - `BoardBuilder`,
  - `ColumnListBuilder`,
  - `ColumnBuilder`,
  - `DataBuilder`,
  - `FeatureBuilder`,
  - `RenderBuilder`,
  - `ActionBuilder`.
- Added `.build()` validation with aggregated error messages.
- Added immutable `Board` objects with:
  - `.render(ctx)`,
  - `.mount(app, prefix)`,
  - `.dispatch(action, event)`,
  - `.clientScriptURL()`.
- Added server-rendered board/column/card output backed by `ui.dsl` node structs.
- Added `board.mount(app, prefix)` route registration:
  - `GET <prefix>/client.js`,
  - `GET <prefix>/<boardId>/fragment`,
  - `POST <prefix>/<boardId>/action/:action`.
- Added generic browser runtime source in Go via `ClientScript()`.
- Added runtime integration tests for `require("kanban.dsl")`, builder rendering, validation aggregation, dispatch event normalization, and client script availability.

### Why

This establishes the core abstraction before migrating the example app. The important architectural move is that the Go native module owns the browser runtime. App authors should write data hooks, render hooks, and callbacks; they should not need to write drag/drop, precise-move fetch calls, fragment replacement, or count updates.

### What worked

- The existing `express` app object is sufficient for `board.mount(app, prefix)` because Go can call its `get` and `post` methods with Go-backed function values.
- The existing `ui.dsl` node structs can be constructed directly by `kanban.dsl` and still rendered by `res.html(...)` / `ui.render(...)`.
- Runtime integration tests can build an engine with `kanbanddsl.NewRegistrar()` and `uidsl.NewRegistrar()` and then use normal JavaScript `require(...)`.

### What didn't work

- The first `cardMoved` dispatch normalization test failed because missing Goja object properties were not always caught by `goja.IsUndefined(...)` alone. The event normalizer now treats `nil`, `undefined`, and `null` as missing via a helper.

### Exact validation commands

```bash
go test ./pkg/kanbanddsl -count=1
go test ./... -count=1
```

Both pass after the missing-value fix.

### What was tricky

- `board.mount(...)` has to bridge in the opposite direction from typical Express use: instead of JavaScript calling Go handlers directly, Go calls JavaScript-facing `app.get` / `app.post` methods with Go-backed handler functions.
- The browser runtime must be generic and data-attribute driven. It currently looks for `data-kb-*` attributes instead of app-specific classes.

### What warrants a second pair of eyes

- Whether action route names should stay `/<prefix>/<boardId>/action/:action` or move to a more RPC-like shape.
- Whether `board.render(...)` should include the client script tag automatically only after mount or always expose it explicitly.
- Whether `.mount(...)` should register duplicate `client.js` routes for multiple boards or share them per prefix. The current implementation shares one client script route per prefix.

### What should be done next

- Commit this foundation.
- Migrate `examples/kanban/scripts/app.js` to use the builder and remove app-owned `/app.js` Kanban runtime code.
- Add an HTTP integration test proving `/_kanban/client.js` is served through the mounted Express app.

## Step 4: Migrated the example app to `kanban.dsl`

I migrated `examples/kanban/scripts/app.js` away from its hand-written Kanban client runtime. The app now imports `require("kanban.dsl")`, creates a `trail-notes` board with the fluid builder API, registers a server-side `cardMoved` callback, and calls `board.mount(app, "/_kanban")`.

### Prompt Context

**User prompt (verbatim):** "importing the dsl also means that the go side can inject / serve custom frontend JS that the app itself doesn't need to write"

**Assistant interpretation:** The existing example must demonstrate this property, not just the package internals.

**Inferred user intent:** The example app should stop defining its own `/app.js` for normal Kanban interactions. It should rely on the DSL-mounted script route.

**Commit (code):** pending at time of writing this diary step.

### What I did

- Added `const kanban = require("kanban.dsl")` to the example.
- Replaced manual board rendering functions with `kanban.board("trail-notes")...build()`.
- Kept domain code intact:
  - migrations,
  - seed data,
  - `listCards`,
  - `moveCard`,
  - new-card form handling,
  - Field Notes CSS.
- Moved card markup into `.render(render => render.card(...))`.
- Moved search form into `.render(render => render.toolbar(...))`.
- Moved movement behavior into `.actions(actions => actions.cardMoved(...))`.
- Removed the app-owned `/app.js` route and app-owned `clientScript()` function.
- Updated the example README to explain that `board.mount(app, "/_kanban")` serves the frontend runtime.
- Added an HTTP integration test that proves a mounted board serves `/_kanban/client.js` and handles `POST /_kanban/<board>/action/cardMoved` with refreshed HTML.

### Validation

```bash
node -c examples/kanban/scripts/app.js
go test ./... -count=1
```

I also ran a live server smoke test with a temporary SQLite database:

```bash
GOTOOLCHAIN=go1.26.2 go run ./cmd/goja-site serve \
  --db /tmp/goja-kanban-XXXX.db \
  --scripts examples/kanban/scripts \
  --addr 127.0.0.1:60125 \
  --dev

curl http://127.0.0.1:60125/
curl http://127.0.0.1:60125/_kanban/client.js
curl -X POST http://127.0.0.1:60125/_kanban/trail-notes/action/cardMoved \
  -H 'Content-Type: application/json' \
  -d '{"cardId":"1","to":{"columnId":"done","index":0}}'
```

The page included `/_kanban/client.js`, the client script contained the generic Kanban runtime markers, and the action endpoint returned JSON with refreshed server-rendered HTML.

### What didn't work

The first live action POST returned HTTP 500 with a nil pointer panic. I reproduced it with a small direct `board.dispatch(...)` program and found that `normalizeEvent` called `.String()` / `.ToInteger()` on missing Goja values while constructing the fallback `from` object. The fix was to use `firstString(...)` and `firstInt(...)`, both of which treat nil/undefined/null values as absent.

### What was tricky

The example still uses a normal server-rendered search form for progressive enhancement, but the mounted client script also sees `data-kb-search` and handles live filtering. This is the intended split: no custom browser JavaScript in the app, but accessible server behavior remains available.

### What should be done next

- Run a browser-level validation of live search and precise move form behavior.
- Consider styling `.kb-root` if the extra wrapper ever affects layout.

## Step 5: Browser-validated search, precise move, and drag/drop

I ran browser validation against the migrated Kanban example and specifically tested drag/drop behavior.

### Prompt Context

**User prompt (verbatim):** "make sure to test drag drop too"

**User prompt (verbatim):** "feel free to add logging if necessary

continue"

**Assistant interpretation:** Do not stop at unit tests or endpoint curl checks. Verify the browser runtime path, including the HTML5 drag/drop event handlers, and add debug logging if it helps confirm the path.

**Inferred user intent:** The key promise of `kanban.dsl` is that app authors do not write client-side Kanban JavaScript, so the DSL-owned browser runtime must actually handle real interactions.

**Commit (code):** pending at time of writing this diary step.

### What I changed

- Fixed `boardFor(element)` in the browser runtime so toolbar elements inside the `data-kb-root` wrapper can find the board even when they are not descendants of the `data-kb-board-id` element.
- Added opt-in client runtime debug logging controlled by:

```javascript
localStorage.setItem("gojaKanbanDebug", "1")
```

- The runtime now emits debug logs for:
  - `dragstart`,
  - `drop`,
  - `postAction`,
  - `postAction response`.

Logging is silent by default.

### Browser validation performed

I started the example with a fresh temporary SQLite database:

```bash
GOTOOLCHAIN=go1.26.2 go run ./cmd/goja-site serve \
  --db /tmp/goja-kanban-browser-XXXX.db \
  --scripts examples/kanban/scripts \
  --addr 127.0.0.1:60128 \
  --dev
```

Then I used Playwright to validate the following.

#### Live search

- Initial visible cards: `10`.
- Filled search with `weather`.
- Visible cards after filtering: `1`.
- Hidden cards after filtering: `9`.
- To Do count after filtering: `1`.
- Clearing the search restored `10` visible cards.

This verified that the generic `data-kb-search` path works after the `boardFor(...)` fix.

#### Precise move form

- Selected card `2`.
- Changed destination column to `done`.
- Changed destination index to `0`.
- Submitted the generated `data-kb-move-form`.
- Observed `POST /_kanban/trail-notes/action/cardMoved`.
- Response status: `200`.
- Response included refreshed HTML.
- Card `2` moved from `todo` to `done`.
- Done count became `5` in that validation run.

#### Drag/drop

Playwright's high-level `locator.dragTo(...)` did not trigger the HTML5 drag/drop path reliably in this environment, so I validated the browser runtime by dispatching real browser `DragEvent` objects with a `DataTransfer` object from inside the page. This exercised the same document-level handlers installed by `/_kanban/client.js`:

```javascript
const source = document.querySelector('[data-kb-card-id="1"]');
const list = document.querySelector('[data-kb-drop-column="done"]');
const dt = new DataTransfer();
const rect = list.getBoundingClientRect();
source.dispatchEvent(new DragEvent('dragstart', { bubbles: true, cancelable: true, dataTransfer: dt }));
list.dispatchEvent(new DragEvent('dragover', { bubbles: true, cancelable: true, dataTransfer: dt, clientX: rect.left + 20, clientY: rect.top + 20 }));
list.dispatchEvent(new DragEvent('drop', { bubbles: true, cancelable: true, dataTransfer: dt, clientX: rect.left + 20, clientY: rect.top + 20 }));
source.dispatchEvent(new DragEvent('dragend', { bubbles: true, cancelable: true, dataTransfer: dt }));
```

Observed result:

```json
{
  "eventResult": { "movedInDom": "done" },
  "response": { "status": 200, "ok": true, "hasHtml": true },
  "after": "done",
  "doneCount": "4"
}
```

Debug console confirmed the exact drag/drop path:

```text
[kanban.debug] dragstart {cardId: 1, columnId: todo, index: 0}
[kanban.debug] drop {cardId: 1, fromColumnId: todo, fromIndex: 0, toColumnId: done, toIndex: 0}
[kanban.debug] postAction {boardId: trail-notes, action: cardMoved, url: /_kanban/trail-notes/action/cardMoved, event: Object}
[kanban.debug] postAction response {action: cardMoved, status: 200, hasHtml: true, payload: Object}
[kanban] Moved card
```

### Validation commands

```bash
go test ./pkg/kanbanddsl -count=1
go test ./... -count=1
```

Both passed.

### What still deserves future attention

- Add a persistent Playwright test harness if this repo adopts browser automation in CI.
- Keep the lower-level synthetic `DragEvent` approach available because Playwright `dragTo(...)` can be unreliable for HTML5 drag/drop in headless/browser automation environments.

## Step 6: Fixed real browser drag start issue

The user reported that dragging a card header selected text and dragging the `...` button did nothing. The console also showed unrelated-looking `index.js` errors about Rangy and missing `document.head` / `document.body`.

### Prompt Context

**User prompt (verbatim):** "i can't drag and drop though, maybe because of this:

Rangy is not supported in this environment. Reason: No body element found index.js:53:169
Error: containerNode is not set index.js:189:253
TypeError: can't access property "appendChild", document.head is null index.js:155:11338"

**User prompt (verbatim):** "when i drag on a card header it just selects the text, and when i drag on the ... nothing seems to happen"

**Assistant interpretation:** The drag/drop path was still fragile in a real browser. The `index.js` errors did not point at `/_kanban/client.js`, but the text selection symptom suggested the card was not actually considered draggable by the browser.

**Inferred user intent:** Make drag/drop feel like a normal Kanban card drag, without the user needing to aim at a special invisible area.

### What I changed

- Changed rendered card markup from a bare `draggable` attribute to `draggable="true"`.
- Added a test assertion that rendered draggable cards include `draggable="true"`.
- Added generic runtime-injected CSS from `/_kanban/client.js`:
  - `cursor: grab` on draggable cards,
  - `cursor: grabbing` while active,
  - `user-select: none` for draggable cards and descendants,
  - fallback drag-over / dragging styles.
- The runtime style injection is defensive and appends to `document.head || document.body || document.documentElement` to avoid head/body timing assumptions.

### Why

`draggable` is an enumerated HTML attribute. A bare `draggable` attribute can be interpreted inconsistently for this use case; `draggable="true"` is the correct explicit form. The injected CSS prevents the common failure mode where dragging text inside the card selects text instead of making the card feel draggable.

The `...` element is a button. Dragging from an interactive control may still be browser-dependent, but dragging from the card body/header now starts the card drag. If we want the menu button itself to drag too, we should either make it not a real button until menus are implemented or add a dedicated drag handle.

### Validation

I restarted the server on the same requested port, `60128`, with the new build:

```bash
lsof-who -p 60128 -k
GOTOOLCHAIN=go1.26.2 go run ./cmd/goja-site serve \
  --db /tmp/goja-kanban-browser-XXXX.db \
  --scripts examples/kanban/scripts \
  --addr 127.0.0.1:60128 \
  --dev
```

Browser checks showed:

```json
{
  "draggableAttribute": "true",
  "draggableProperty": true,
  "cursor": "grab",
  "userSelect": "none",
  "hasRuntimeStyles": true
}
```

Then Playwright `dragTo(...)` succeeded against the real page:

```json
{
  "before": "todo",
  "after": "done",
  "doneCount": "4",
  "response": { "status": 200, "ok": true }
}
```

Debug console confirmed the real path:

```text
[kanban.debug] dragstart {cardId: 1, columnId: todo, index: 0}
[kanban.debug] drop {cardId: 1, fromColumnId: todo, fromIndex: 0, toColumnId: done, toIndex: 0}
[kanban.debug] postAction {boardId: trail-notes, action: cardMoved, url: /_kanban/trail-notes/action/cardMoved, event: Object}
[kanban.debug] postAction response {action: cardMoved, status: 200, hasHtml: true, payload: Object}
[kanban] Moved card
```

Go validation:

```bash
go test ./pkg/kanbanddsl -count=1
go test ./... -count=1
```

### Notes on the Rangy errors

Those errors mention `index.js`, not `/_kanban/client.js`, and they did not appear in my current browser console after restart. They are likely from an extension, injected script, or unrelated page helper. The Kanban runtime now avoids assuming `document.head` exists when injecting its own styles.

## Step 7: Propagated host sessions through Kanban render and action routes

The host now exposes `req.session`, and `kanban.dsl` propagates that session through mounted routes. This lets app authors scope board rows by cookie without writing cookie plumbing in app JavaScript.

### Prompt Context

**User prompt (verbatim):** "Ok, let's add go side support for user sessinos, so that the JS can scope kanban boards to a cookie, but almost doesn't have to see the session id except when querying the db."

**Assistant interpretation:** Mounted Kanban fragment and action routes should carry the host session into `ctx.session` and `event.session` automatically.

### What changed

- `GET /_kanban/<board>/fragment` now calls `board.Render` with `{ query, session }`.
- `POST /_kanban/<board>/action/:action` injects `session` into the event before dispatch.
- Action-triggered refresh renders also receive the same session.
- The Kanban example now queries and mutates `cards` by `session_id`.

### Why

The DSL should hide transport mechanics. The application still needs to use the session ID at the persistence boundary, but it should not generate cookies or manually parse them.

### Validation

- `go test ./pkg/kanbanddsl -count=1`
- `go test ./... -count=1`
- Live two-cookie-jar smoke test confirmed one session's new card is invisible to another session.
