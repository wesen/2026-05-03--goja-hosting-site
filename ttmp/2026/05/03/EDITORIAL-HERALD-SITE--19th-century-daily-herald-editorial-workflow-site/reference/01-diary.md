---
Title: Diary
Ticket: EDITORIAL-HERALD-SITE
Status: active
Topics:
    - goja-site
    - javascript
    - kanban
    - sqlite
    - documentation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: sites/editorial/scripts/00_domain.js
      Note: Daily Herald domain constants and helpers (commit bbef48b)
    - Path: sites/editorial/scripts/01_styles.js
      Note: Broadsheet visual style for the reference-inspired UI (commit bbef48b)
    - Path: sites/editorial/scripts/02_repository.js
      Note: SQLite schema and persistence layer for staff
    - Path: sites/editorial/scripts/03_workflow.js
      Note: Seed data
    - Path: sites/editorial/scripts/04_views.js
      Note: Server-rendered shell
    - Path: sites/editorial/scripts/05_routes.js
      Note: HTTP routes
    - Path: ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/design-doc/01-daily-herald-editorial-workflow-architecture.md
      Note: Architecture plan derived from reference images
    - Path: ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/sources/images/01-full-editorial-workflow-reference.png
      Note: Primary visual reference imported with cp
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Capture the step-by-step implementation history for replacing the previous Editorial example with a richer 19th-century Daily Herald newsroom application, including reference images, architectural choices, failures, validation commands, and review guidance.

## Step 1: Import references and build the first Daily Herald application

The first milestone created the `EDITORIAL-HERALD-SITE` ticket, copied the user-provided reference images into the ticket workspace, and replaced the single-file Editorial example with a split-script Daily Herald application. The result is a server-rendered newsroom shell with a left desk sidebar, a main `kanban.dsl` workflow board, a right story dossier panel, SQLite-backed staff/story/checklist data, JSON APIs, and local validation on port `60133`.

This milestone intentionally keeps the board as only one component inside a broader application. The interesting behavior now comes from the combined domain model: stories have authors, desks, priorities, due dates, checklist progress, and detailed dossier content in addition to workflow status.

### Prompt Context

**User prompt (verbatim):** "Import the images into the ticket with `cp` for referencing later."

**Assistant interpretation:** Create a dedicated docmgr ticket for the editorial redesign, copy the supplied clipboard images into a stable ticket path, and use them as implementation references.

**Inferred user intent:** Preserve visual requirements and source material in versioned project documentation before implementing a redesign based on those images.

**Commit (code):** `bbef48b8027842f2b503a883b213d921204f9d1f` — "Redesign editorial site as Daily Herald workflow"

### What I did

- Created ticket `EDITORIAL-HERALD-SITE` with `docmgr ticket create-ticket`.
- Copied the three reference PNGs with `cp` into:
  - `ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/sources/images/01-full-editorial-workflow-reference.png`
  - `ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/sources/images/02-story-detail-panel-reference.png`
  - `ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/sources/images/03-board-only-reference.png`
- Ran `understand_image` on the full workflow reference to extract concrete UI and data requirements.
- Added design document:
  - `ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/design-doc/01-daily-herald-editorial-workflow-architecture.md`
- Removed the previous single file:
  - `sites/editorial/scripts/app.js`
- Added split scripts:
  - `sites/editorial/scripts/00_domain.js`
  - `sites/editorial/scripts/01_styles.js`
  - `sites/editorial/scripts/02_repository.js`
  - `sites/editorial/scripts/03_workflow.js`
  - `sites/editorial/scripts/04_views.js`
  - `sites/editorial/scripts/05_routes.js`
- Implemented SQLite tables `herald_staff`, `herald_stories`, and `herald_checklist`.
- Implemented seed data for a 19th-century newspaper workflow, including pitches, reporting, writing, editing, and ready-for-print stories.
- Implemented `kanban.dsl` board mounting and movement callback through `Herald.workflow.moveStory(event)`.
- Added JSON endpoints:
  - `GET /api/stories`
  - `GET /api/stories/:id`
  - `GET /api/assignments`
  - `GET /api/desks`
  - `GET /api/metrics`
- Served the site locally on `127.0.0.1:60133`.
- Committed the first milestone as `bbef48b8027842f2b503a883b213d921204f9d1f`.

### Why

- The user wanted the reference images imported into the ticket so the visual brief would survive beyond `/tmp` clipboard paths.
- The previous Editorial example did not demonstrate enough application depth; it needed to become a richer newsroom workflow rather than a visually restyled board.
- Split scripts make this larger JavaScript example reviewable and easier for future authors to copy.

### What worked

- `docmgr ticket create-ticket --ticket EDITORIAL-HERALD-SITE --title "19th century Daily Herald editorial workflow site" --topics goja-site,javascript,kanban,sqlite,documentation` created the expected workspace.
- `cp` imported all three reference images into `sources/images` successfully.
- `understand_image` identified the main shell pieces: sidebar, topbar, board, right dossier, desk filtering concepts, author metadata, tags, and checklist.
- `node -c sites/editorial/scripts/*.js` passed after formatting.
- `go test ./... -count=1` passed:
  - `pkg/app`
  - `pkg/dbguard`
  - `pkg/kanbanddsl`
  - `pkg/uidsl`
  - `pkg/web`
- Local API validation worked:
  - `curl -fsS http://127.0.0.1:60133/api/metrics | jq .`
  - returned `active: 8`, `ready: 2`, `highPriority: 4`, `checklistPercent: 23`.
- Main page validation worked after fixing the unsupported tag helper:
  - `curl -fsS http://127.0.0.1:60133/ | grep -Eo 'The Daily Herald|Editorial Workflow|City Council debates|Checklist' | head -10`

### What didn't work

- The first page render failed with this exact runtime error:

```text
JavaScript handler error: TypeError: Object has no member 'aside' at sidebar (sites/editorial/scripts/04_views.js:23:20(116))
```

- Root cause: `ui.dsl` does not currently expose an `aside()` helper, even though HTML has an `aside` element.
- Fix: replaced `ui.aside(...)` with `ui.section(...)` while preserving the CSS classes `sidebar` and `dossier`.
- The page now renders, but clicking story cards still uses full-page navigation via `href="/?story=<id>"`, causing a visible flash. That is the next behavior to improve with progressive enhancement.

### What I learned

- The `ui.dsl` element surface is intentionally narrower than all HTML tags. New examples should either use known helpers or the DSL should grow an explicit generic-tag escape hatch.
- The current app architecture can support a rich shell without changing Go code: all domain state, views, routes, and board callbacks live in site scripts.
- The right story dossier is an ideal place to demonstrate progressive enhancement because it can remain server-rendered and be swapped as a fragment.

### What was tricky to build

- The main sharp edge was shaping `kanban.dsl` card links so the card remains draggable but also navigates to a selected story for the dossier. In v1 this is a normal link inside the rendered card, which is simple and accessible but triggers a full page reload.
- The second sharp edge was SQLite session isolation. All seed, query, movement, and checklist functions use `Herald.util.sessionId(session)` so each cookie session gets its own staff/story/checklist rows.
- The unsupported `ui.aside` helper was discovered only at runtime because JavaScript syntax validation cannot detect missing members on the DSL object.

### What warrants a second pair of eyes

- Review the generated board markup around linked cards and drag/drop behavior. A card that contains a link may produce subtle click-vs-drag interactions in browsers.
- Review `Herald.workflow.seedIfEmpty` and `Herald.repo.countStories`; it seeds per session, which is correct for examples but can surprise reviewers expecting shared production data.
- Review `Herald.workflow.moveStory`; it normalizes destination and source status positions, but it does not yet enforce checklist completion before `ready`.

### What should be done in the future

- Add progressive enhancement to avoid full-page flashes when selecting stories or toggling checklist items.
- Consider adding an `ui.element(tagName, attrs, children...)` helper or additional semantic helpers such as `ui.aside` if this comes up repeatedly.
- Add a future optional movement rule that warns or blocks movement to `ready` when checklist items remain incomplete.

### Code review instructions

- Start with `sites/editorial/scripts/00_domain.js` for constants and domain helpers.
- Then read `sites/editorial/scripts/02_repository.js` for schema and SQL access.
- Then read `sites/editorial/scripts/03_workflow.js` for seeding, movement, metrics, and checklist behavior.
- Then read `sites/editorial/scripts/04_views.js` for the shell, board, cards, and dossier.
- Finally read `sites/editorial/scripts/05_routes.js` for HTTP routes and API shape.
- Validate with:

```bash
node -c sites/editorial/scripts/*.js
go test ./... -count=1
curl -fsS http://127.0.0.1:60133/ | grep 'The Daily Herald'
curl -fsS http://127.0.0.1:60133/api/desks | jq
curl -fsS http://127.0.0.1:60133/api/metrics | jq
```

### Technical details

Local server command used for validation:

```bash
GOTOOLCHAIN=go1.26.2 go run ./cmd/goja-site serve \
  --db ./tmp/editorial-herald/app.db \
  --scripts ./sites/editorial/scripts \
  --addr 127.0.0.1:60133 \
  --dev
```

Current local URL:

```text
http://127.0.0.1:60133/
```

Known next UI issue:

```text
http://localhost:60133/?story=33 reloads the entire page when story links are clicked.
```
