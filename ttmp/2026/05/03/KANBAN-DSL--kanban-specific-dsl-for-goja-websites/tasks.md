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

- [ ] Create `pkg/kanbanddsl` package.
- [ ] Implement runtime-scoped `kanban.dsl` registrar.
- [ ] Add registrar to `pkg/app/server.go`.
- [ ] Export `kanban.board(id)` returning a mutable `BoardBuilder`.
- [ ] Add runtime smoke test for `require("kanban.dsl")`.

## Phase 2 — Fluid builder API

- [ ] Implement `BoardBuilder` with `.title`, `.theme`, `.columns`, `.data`, `.features`, `.render`, `.actions`, and `.build`.
- [ ] Implement `ColumnListBuilder` and `ColumnBuilder` with `.column(id)`, `.title`, `.limit`, `.terminal`, `.attrs`, and `.done`.
- [ ] Implement `DataBuilder` with `.cards`, `.id`, `.column`, `.position`, and `.searchText`.
- [ ] Implement `FeatureBuilder` with `.search`, `.preciseMove`, `.dragDrop`, `.createCard`, `.cardMenu`, and `.readOnly`.
- [ ] Implement `RenderBuilder` with `.card`, `.toolbar`, `.columnHeader`, `.emptyColumn`, and `.boardShell`.
- [ ] Implement `ActionBuilder` with `.cardMoved`, `.cardCreated`, `.cardUpdated`, `.cardDeleted`, `.cardClicked`, `.cardMenuAction`, and `.custom`.
- [ ] Add `.build()` validation with aggregated errors.
- [ ] Freeze built boards and reject builder reuse after build.

## Phase 3 — Rendering and mounting

- [ ] Render board/columns/cards to `ui.dsl` nodes.
- [ ] Support custom `render.card` mixed with `ui.dsl` components.
- [ ] Implement `board.mount(app, prefix)`.
- [ ] Serve generic Kanban client runtime.
- [ ] Add board fragment endpoint.
- [ ] Add action dispatch endpoint.

## Phase 4 — Interactions and callbacks

- [ ] Implement browser runtime search behavior.
- [ ] Implement precise move controls.
- [ ] Implement drag/drop behavior.
- [ ] Define and validate action envelopes.
- [ ] Dispatch `cardMoved` to server-side Goja callback.
- [ ] Return refreshed server-rendered board fragments.

## Phase 5 — Example migration and tests

- [ ] Rewrite `examples/kanban/scripts/app.js` to use `kanban.dsl`.
- [ ] Add trail board example.
- [ ] Add editorial pipeline example.
- [ ] Add sales pipeline example.
- [ ] Add Go unit/runtime tests.
- [ ] Add Playwright tests for search, precise move, drag/drop, callback dispatch, and no console errors.
