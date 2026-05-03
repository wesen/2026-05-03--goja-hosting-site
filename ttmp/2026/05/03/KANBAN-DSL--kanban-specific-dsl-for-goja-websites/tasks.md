# Tasks

## Phase 0 — Ticket and design

- [x] Create `KANBAN-DSL` docmgr ticket.
- [x] Create Kanban DSL architecture/design guide.
- [x] Add concrete example app designs.
- [x] Update design to prefer a fluid builder API with Go-side validation.
- [x] Add builder-style examples.
- [x] Validate ticket with `docmgr doctor`.
- [x] Upload updated bundle to reMarkable.

## Phase 1 — Module shell

- [x] Create `pkg/kanbanddsl` package.
- [x] Implement runtime-scoped `kanban.dsl` registrar.
- [x] Add registrar to `pkg/app/server.go`.
- [x] Export `kanban.board(id)` returning a mutable `BoardBuilder`.
- [x] Add runtime smoke test for `require("kanban.dsl")`.

## Phase 2 — Fluid builder API

- [x] Implement `BoardBuilder` with `.title`, `.theme`, `.columns`, `.data`, `.features`, `.render`, `.actions`, and `.build`.
- [x] Implement `ColumnListBuilder` and `ColumnBuilder` with `.column(id)`, `.title`, `.limit`, `.terminal`, `.attrs`, and `.done`.
- [x] Implement `DataBuilder` with `.cards`, `.id`, `.column`, `.position`, and `.searchText`.
- [x] Implement `FeatureBuilder` with `.search`, `.preciseMove`, `.dragDrop`, `.createCard`, `.cardMenu`, and `.readOnly`.
- [x] Implement `RenderBuilder` with `.card`, `.toolbar`, `.columnHeader`, `.emptyColumn`, and `.boardShell`.
- [x] Implement `ActionBuilder` with `.cardMoved`, `.cardCreated`, `.cardUpdated`, `.cardDeleted`, `.cardClicked`, `.cardMenuAction`, and `.custom`.
- [x] Add `.build()` validation with aggregated errors.
- [x] Freeze built boards and reject builder reuse after build.

## Phase 3 — Rendering and mounting

- [x] Render board/columns/cards to `ui.dsl` nodes.
- [x] Support custom `render.card` mixed with `ui.dsl` components.
- [x] Implement `board.mount(app, prefix)`.
- [x] Serve generic Kanban client runtime.
- [x] Add board fragment endpoint.
- [x] Add action dispatch endpoint.

## Phase 4 — Interactions and callbacks

- [x] Implement browser runtime search behavior.
- [x] Implement precise move controls.
- [x] Implement drag/drop behavior.
- [x] Define and validate action envelopes.
- [x] Dispatch `cardMoved` to server-side Goja callback.
- [x] Return refreshed server-rendered board fragments.

## Phase 5 — Example migration and tests

- [x] Rewrite `examples/kanban/scripts/app.js` to use `kanban.dsl`.
- [ ] Add trail board example.
- [ ] Add editorial pipeline example.
- [ ] Add sales pipeline example.
- [x] Add Go unit/runtime tests.
- [ ] Add Playwright tests for search, precise move, drag/drop, callback dispatch, and no console errors.
