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

- [ ] Add precise move render mode to `FeatureSpec` and builder API.
- [ ] Implement `preciseMove("none")` while preserving eager default.
- [ ] Add tests for eager and no-precise rendering.
- [ ] Add optimized benchmark fixture.
- [ ] Run before/after single-VM benchmark.
- [ ] Run before/after multi-VM benchmark.
- [ ] Capture pprof for baseline and optimized degraded cells.

## Phase 2 — Broader architecture experiments

- [ ] Design action refresh modes: full, none, columns, patches, events.
- [ ] Prototype partial HTML patch response mode.
- [ ] Evaluate optional client-side site code API.
- [ ] Decide whether same-host VM pooling is worth designing after render-cost reductions.
