# Tasks

## Phase 0 — Ticket setup and delivery

- [x] Create `SESSION-DB-MAINT` ticket workspace.
- [x] Add session implementation guide document.
- [x] Add SQLite size guard / cleanup callback design document.
- [x] Add investigation diary.
- [x] Relate documents to source files.
- [x] Run `docmgr doctor --ticket SESSION-DB-MAINT --stale-after 30`.
- [x] Upload ticket bundle to reMarkable.
- [x] Commit ticket documentation.

## Phase 1 — Session implementation documentation

- [x] Explain `pkg/web/session.go` session manager.
- [x] Explain cookie attributes and ID generation.
- [x] Explain `req.session` request DTO integration.
- [x] Explain `ctx.session` and `event.session` propagation through `kanban.dsl`.
- [x] Explain session-scoped Kanban rows with `session_id`.
- [x] Include validation evidence and implementation file references.

## Phase 2 — SQLite size guard design

- [x] Analyze whether the upstream database module needs to be modified.
- [x] Identify `QueryExecer` as the preferred wrapper point.
- [x] Design `MeteredDB` wrapper around `*sql.DB`.
- [x] Design `db.guard` module API.
- [x] Design JavaScript cleanup callback event/result shapes.
- [x] Include SQLite file/WAL/SHM size measurement strategy.
- [x] Include session-aware cleanup policy examples.
- [x] Include phased implementation plan and risks.

## Future implementation tasks

- [ ] Create `pkg/dbguard` stats package.
- [ ] Measure main DB, WAL, and SHM file sizes.
- [ ] Add SQLite page stats via `PRAGMA page_size`, `page_count`, and `freelist_count`.
- [ ] Add `db.guard` native module registrar.
- [ ] Add `guard.configure`, `guard.stats`, `guard.checkNow`, and `guard.onLimitExceeded`.
- [ ] Wrap `*sql.DB` with `MeteredDB` before passing it to `databasemod.WithPreconfiguredDB`.
- [ ] Trigger soft-limit checks after writes with cooldown and recursion guard.
- [ ] Add runtime integration tests for cleanup callback dispatch.
- [ ] Add Kanban example cleanup policy behind an explicit demo/config flag.
