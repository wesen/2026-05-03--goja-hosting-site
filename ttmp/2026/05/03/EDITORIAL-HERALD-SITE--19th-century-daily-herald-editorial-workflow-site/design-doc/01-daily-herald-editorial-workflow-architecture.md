---
Title: Daily Herald Editorial Workflow Architecture
Slug: daily-herald-editorial-workflow-architecture
Short: Architecture and implementation plan for replacing the Editorial example with a 19th-century newspaper workflow app.
Topics:
- goja-site
- javascript
- kanban
- sqlite
- documentation
DocType: design-doc
Ticket: EDITORIAL-HERALD-SITE
Status: active
Intent: long-term
Created: 2026-05-03
Updated: 2026-05-03
---

# Daily Herald Editorial Workflow Architecture

## Executive Summary

The current Editorial example is still too close to a generic Kanban board. The new Daily Herald site should feel like a complete 19th-century newsroom application: a left navigation desk directory, a broadsheet-style workflow board, a right story dossier panel, detailed assignments and checklist data, and useful JSON endpoints for stories, staff, assignments, and desk summaries.

The implementation should follow the same readability pattern used for Pizza Ops: split the site into small numbered scripts with clear responsibilities. The app remains server-rendered with `ui.dsl`; `kanban.dsl` owns drag/drop and board refresh; SQLite stores stories, staff, checklist items, and assignments.

## Reference Images

The ticket imports the requested reference images for implementation review:

```text
sources/images/01-full-editorial-workflow-reference.png
sources/images/02-story-detail-panel-reference.png
sources/images/03-board-only-reference.png
```

Key UI features extracted from the images:

- left sidebar with newspaper building mark, Dashboard/Workflow/Assignments/Calendar/Archive/People, desk sections, editor profile, and established date;
- top bar with `The Daily Herald`, `Editorial Room`, global search, notification icon, and editor portrait;
- main Editorial Workflow area with Board/Table tabs, Filter/Sort controls, and New Story action;
- five workflow columns: Pitches, Reporting, Writing, Editing, Ready for Print;
- cards with title, desk/topic, byline, due date, avatar, and completion/check mark for print-ready items;
- right side story dossier with status, title, desk, author/role, due date, description, tags, checklist, and Open full view action.

## Problem Statement

A good example site should teach that `goja-site` can coordinate several server-rendered surfaces over one domain model. The current Editorial site mostly demonstrates styling and cards. It does not yet demonstrate:

- a domain-specific application shell;
- detailed assignment metadata;
- side-panel story details;
- staff and desk concepts;
- checklist/progress data;
- useful non-board routes and APIs;
- a clear architecture for larger JavaScript sites.

## Proposed Solution

Replace `sites/editorial/scripts/app.js` with a split-script Daily Herald application:

```text
sites/editorial/scripts/
  00_domain.js      workflow columns, desk definitions, staff seed data, utility functions
  01_styles.js      sepia broadsheet / 19th-century newspaper UI
  02_repository.js  SQLite schema and low-level queries/mutations
  03_workflow.js    story creation, seeding, movement, detail selection, checklist/tally APIs
  04_views.js       ui.dsl shell, sidebar, topbar, dossier panel, kanban board builders
  05_routes.js      HTTP routes, board mounting, API endpoints
```

This preserves lexical loading while avoiding a single large script.

## Data Model

### `herald_staff`

Staff rows represent byline and editor identity.

Fields:

- `id`
- `session_id`
- `name`
- `role`
- `initials`
- `portrait_mark`

### `herald_stories`

Story rows drive both the board and detail panel.

Fields:

- `id`
- `session_id`
- `title`
- `dek`
- `description`
- `workflow_status`: `pitches`, `reporting`, `writing`, `editing`, `ready`
- `position`
- `desk`: City Hall, Society, Business, Education, Technology, Environment
- `author_id`
- `priority`: High, Normal, Low
- `due_date`
- `tags`: comma-separated tags for v1
- `source_notes`
- `word_target`
- `word_count`
- `created_at`, `updated_at`

### `herald_checklist`

Checklist rows provide concrete assignment work beyond moving a card.

Fields:

- `id`
- `session_id`
- `story_id`
- `label`
- `done`
- `position`

Checklist examples:

- Interview Alderman Pierce
- Review housing report
- Verify budget figures
- Confirm copy desk headline
- Typeset first proof

## UI Architecture

### Shell

The shell should use a three-column app layout:

```text
left sidebar | main workflow | right story dossier
```

The sidebar and dossier are always rendered by the server. The active story comes from `?story=<id>`, falling back to the first writing/story row if absent.

### Kanban Board

The board is still `kanban.dsl`, but it is embedded into a broader page.

Columns:

```text
Pitches | Reporting | Writing | Editing | Ready for Print
```

Card rendering includes:

- story title;
- desk;
- byline;
- due date;
- portrait mark/avatar;
- checklist progress;
- print-ready check mark when status is `ready`.

The board should not use generated precise move controls. Drag/drop is enough for the visual example.

### Dossier Panel

The dossier panel renders detailed story state:

- status label;
- title;
- desk;
- author row with portrait, name, role;
- due date;
- description;
- tag chips;
- checklist with checked/unchecked boxes;
- assignment facts: priority, word count, target, source notes;
- Open full view button.

### Interesting Features Beyond Kanban

The implementation should include:

- `GET /?story=<id>` server-selected detail panel;
- `POST /stories` to create a story;
- `POST /stories/:id/checklist/:itemId/toggle` to toggle checklist state;
- `GET /api/stories` for board data;
- `GET /api/stories/:id` for dossier JSON;
- `GET /api/assignments` for checklist/staff assignment view;
- `GET /api/desks` for desk counts and status summary;
- top-level metrics: active stories, due today/soon, ready for print, checklist completion.

## Movement Rules

Moving cards updates `workflow_status` and normalizes positions. If a story reaches `ready`, it should become visually marked as print-ready. Checklist items are not hard blockers in v1; instead, the dossier makes incomplete work visible. A future iteration could reject movement to Ready until required checklist items are complete.

## Design Decisions

### Server-selected dossier instead of custom browser state

The first version should use ordinary links/forms and query parameters to select a story. This keeps the example aligned with server-rendered `goja-site` and avoids app-specific browser JavaScript.

### Split scripts by responsibility

The Pizza example showed that dense one-file examples become difficult to maintain. The Daily Herald app should start with a clean split.

### Keep `kanban.dsl` as a component

The app shell, sidebars, detail panel, and APIs are ordinary `ui.dsl` routes. The Kanban board is only one component in the newsroom application.

## Implementation Plan

1. Replace `sites/editorial/scripts/app.js` with split scripts.
2. Add schema and seed data for staff, stories, and checklist rows.
3. Build the shell and broadsheet visual style.
4. Build the Kanban board with drag/drop only.
5. Build the dossier panel and checklist toggle forms.
6. Add JSON APIs.
7. Validate locally with single-site and multi-site serving.
8. Push app changes to publish a new image.
9. Let the existing K3s GitOps automation deploy the new image to `editorial.kanban.yolo.scapegoat.dev`.

## Validation Checklist

```bash
node -c sites/editorial/scripts/*.js
go test ./... -count=1
GOTOOLCHAIN=go1.26.2 go run ./cmd/goja-site serve --db ./tmp/editorial-herald/app.db --scripts ./sites/editorial/scripts --addr 127.0.0.1:60133 --dev
curl -fsS http://127.0.0.1:60133/ | grep 'The Daily Herald'
curl -fsS http://127.0.0.1:60133/api/stories | jq length
curl -fsS http://127.0.0.1:60133/api/desks | jq
```

Production validation after deployment:

```bash
curl -k -fsSI https://editorial.kanban.yolo.scapegoat.dev/ | head -1
curl -k -fsS https://editorial.kanban.yolo.scapegoat.dev/ | grep 'The Daily Herald'
curl -k -fsS https://editorial.kanban.yolo.scapegoat.dev/api/desks | jq
```
