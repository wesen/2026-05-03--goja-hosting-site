# Tasks

## Phase 0 — Ticket and harness setup

- [x] Create GOJA-MULTI-VM-STRESS ticket workspace.
- [x] Write multi-VM serve-multi stress plan.
- [x] Add single-run multi-VM Vegeta harness script.
- [x] Add quick sweep script.
- [x] Add Markdown summary renderer.
- [x] Validate script syntax and docmgr hygiene.
- [x] Commit initial ticket setup.

## Phase 1 — Quick multi-VM validation sweep

- [x] Run quick multi-VM sweep.
- [x] Generate quick sweep Markdown report.
- [x] Record results in diary/changelog.
- [x] Commit quick sweep results.

## Phase 2 — Follow-up stress experiments

- [ ] Decide whether to add pprof runs for multi-VM cases.
- [ ] Add a longer even-hot run if quick sweep is healthy.
- [ ] Add a one-hot idle-VM overhead run with larger VM counts.
- [ ] Add kanban-action only after choosing safe rates below the known single-VM knee.
