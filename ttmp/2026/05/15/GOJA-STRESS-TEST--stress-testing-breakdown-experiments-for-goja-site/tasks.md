# Tasks

## Phase 0 — Stress ticket setup

- [x] Create GOJA-STRESS-TEST ticket workspace.
- [x] Write stress testing breakdown plan.
- [x] Add SQLite schema/import/report scripts in the ticket scripts folder.
- [x] Add quick stress sweep script (~2-3 minutes) before hour-scale experiment.
- [x] Add hour-scale stress sweep script for later use.

## Phase 1 — Quick stress validation

- [x] Run quick stress sweep.
- [x] Import quick stress results into SQLite.
- [x] Generate quick stress Markdown report with embedded SQL queries.
- [x] Record quick stress findings in diary and changelog.
- [x] Upload quick stress report to reMarkable.
- [x] Commit quick stress results.

## Phase 2 — Follow-up decisions

- [ ] Decide whether the hour-scale experiment is safe to run.
- [x] If needed, tune stress rates/scenarios based on quick sweep behavior.
- [ ] Optionally add explicit multi-site / many-VM stress scenarios.
- [x] Run targeted pprof on the first scenario that bends.

## Phase 3 — Kanban action knee search

- [x] Add targeted `kanban-action` knee-search script.
- [x] Run targeted `kanban-action` knee search from 60/s to 100/s.
- [x] Import knee-search results into SQLite.
- [x] Generate knee-search Markdown report with embedded SQL queries.
- [x] Run targeted pprof at the first bending rate.
- [x] Upload knee-search and pprof reports to reMarkable.
- [ ] Commit knee-search and pprof artifacts.

