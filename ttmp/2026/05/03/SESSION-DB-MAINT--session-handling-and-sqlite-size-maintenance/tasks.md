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


## Phase 3 — Hard-limit enforcement

- [x] Add hard-limit enforcement policy to the design document.
- [x] Add SQL statement classification for growth vs cleanup/maintenance statements.
- [x] Add `Guard.BeforeExec(query)` hard-limit preflight.
- [x] Add post-exec hard-limit error handling after cleanup attempts.
- [x] Allow cleanup and maintenance statements while over the hard limit.
- [x] Always allow writes while the cleanup callback is running.
- [x] Return explicit hard-limit errors through `db.exec(...)` when fail mode is enabled.
- [x] Add tests for rejected growth writes and allowed cleanup writes.

## Future implementation tasks

- [x] Create `pkg/dbguard` stats package.
- [x] Measure main DB, WAL, and SHM file sizes.
- [x] Add SQLite page stats via `PRAGMA page_size`, `page_count`, and `freelist_count`.
- [x] Add `db.guard` native module registrar.
- [x] Add `guard.configure`, `guard.stats`, `guard.checkNow`, and `guard.onLimitExceeded`.
- [x] Wrap `*sql.DB` with `MeteredDB` before passing it to `databasemod.WithPreconfiguredDB`.
- [x] Trigger soft-limit checks after writes with cooldown and recursion guard.
- [x] Add runtime integration tests for cleanup callback dispatch.
- [ ] Add Kanban example cleanup policy behind an explicit demo/config flag.
