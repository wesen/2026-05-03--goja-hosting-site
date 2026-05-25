# Tasks

## Phase 0 — Design package

- [x] Create GOJA-KANBAN-RENDER-OPT ticket workspace.
- [x] Write immediate Kanban render optimization implementation guide.
- [x] Write holistic Goja hosting performance architecture guide.
- [x] Keep a chronological design diary.
- [x] Add a script stub for future before/after benchmark runs.
- [x] Validate docmgr hygiene.
- [x] Upload guide bundle to reMarkable.
- [x] Commit ticket docs.

## Phase 1 — First implementation slice

- [x] Remove precise move form support from `FeatureSpec` and builder API.
- [x] Remove eager server-rendered precise move forms instead of adding render modes.
- [x] Update tests for compact accessible card action controls.
- [x] Update Kanban benchmark fixture to use the simplified DSL.
- [x] Run post-simplification single-VM benchmark.
- [x] Run post-simplification multi-VM benchmark.
- [ ] Capture pprof for baseline and optimized degraded cells.

## Phase 2 — Broader architecture experiments

- [ ] Design action refresh modes: full, none, columns, patches, events.
- [ ] Prototype partial HTML patch response mode.
- [ ] Evaluate optional client-side site code API.
- [ ] Decide whether same-host VM pooling is worth designing after render-cost reductions.

## Phase 1A — Accessibility follow-up

- [x] Add frontend action menu, keyboard navigation, live region, and focus restoration.
- [x] Add browser/Playwright accessibility interaction tests.
- [x] Validate keyboard-only workflow with Playwright.

## Phase 3 — Anti-overfit benchmark planning

- [x] Write detailed anti-overfit benchmark plan.
- [x] Build Anti-overfit matrix v1 fixtures and runner.
- [x] Run Anti-overfit matrix v1 and write report.
- [ ] Decide which optimization class to pursue next from broader benchmark evidence.

- [x] Upload Anti-overfit matrix v1 report to reMarkable.

## Phase 4 — Anti-overfit follow-up pprof

- [x] Capture pprof for `render-attrs-1000` at `100/s`.
- [x] Capture pprof for `kanban-fragment-500` at `100/s`.
- [x] Write comparative follow-up pprof report.
- [x] Decide whether next implementation target is generic UI attr rendering, large Kanban/full-tree rendering, or action response protocol.

## Phase 5 — render-attrs-1000 investigation and optimization

- [x] Write detailed `render-attrs-1000` performance investigation.
- [x] Add focused `ui.dsl` microbenchmarks in `go-go-goja`.
- [x] Add `ui.dsl` attribute compatibility tests before optimization.
- [x] Implement first optimization slice: avoid double `Export()` in attrs detection/extraction.
- [x] Re-run `render-attrs-1000` pprof after first optimization slice.

- [x] Cut over `Element.Attrs` to render-ready `[]uidsl.Attr` without compatibility map path.

## Phase 6 — Goja export boundary research

- [x] Study `/home/manuel/code/others/goja` internals for `Value.Export`, object export, array export, key enumeration, and object construction.
- [x] Write intern-ready UI DSL export-boundary internals and next optimization design.
- [x] Upload export-boundary internals/design guide to reMarkable.
- [ ] Later prototype `ui.attrs(...)` Go-owned attrs wrapper.
- [ ] Later prototype flat-pair constructor / Go-side builder variants.
- [ ] Later consider Goja internals API only after microbenchmarks justify it.
