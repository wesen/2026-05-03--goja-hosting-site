---
Title: Pizza Ops Implementation Diary
Slug: pizza-ops-implementation-diary
Short: Chronological implementation and validation notes for the Pizza Ops example site.
Topics:
- goja-site
- javascript
- kanban
- sqlite
DocType: reference
Ticket: PIZZA-KITCHEN-SITE
Status: active
Intent: long-term
Created: 2026-05-03
Updated: 2026-05-03
---

# Pizza Ops Implementation Diary

## Step 1: Architecture and initial local implementation

Created ticket `PIZZA-KITCHEN-SITE` for a richer example site that is more than a restyled Kanban board. The design uses a normal `ui.dsl` pizza builder form, a dependency-aware kitchen Kanban board, a delivery Kanban board, and a tally for paid orders and tips.

The app was split into readable scripts instead of one dense `app.js`:

```text
sites/pizza/scripts/00_domain.js
sites/pizza/scripts/01_styles.js
sites/pizza/scripts/02_store.js
sites/pizza/scripts/03_views.js
sites/pizza/scripts/04_routes.js
```

## Step 2: Local Pizza Ops refinement and done-column grouping

After local review, the Pizza Ops example was adjusted for readability and vertical space:

- disabled `preciseMove()` on the pizza kitchen and delivery boards, removing the generated `Move` button plus destination/index dropdowns from every card;
- kept drag/drop enabled so the Go-owned Kanban runtime still drives movement;
- stacked Kitchen above Delivery instead of placing the two boards side by side;
- restyled toward the more legible Trail Notes look while keeping a monochrome retro UI;
- split the app into readable numbered scripts under `sites/pizza/scripts`;
- collapsed completed kitchen tasks by pizza order in the Done column.

The Done column now shows one summary card per order instead of one card per completed task. Raw task data remains available through `GET /api/tasks`; the grouping only affects the kitchen board projection. The grouped cards use synthetic IDs such as `done-order-5`, and the server rejects attempts to drag those summary cards as a safety guard.

Validation:

```bash
node -c sites/pizza/scripts/*.js
go test ./... -count=1
GOTOOLCHAIN=go1.26.2 go run ./cmd/goja-site serve --db ./tmp/pizza-single/app.db --scripts ./sites/pizza/scripts --addr 127.0.0.1:60132 --dev
GOTOOLCHAIN=go1.26.2 go run ./cmd/goja-site serve-multi --config deploy/sites.local.yaml
```

Local checks:

```text
GET  /                      -> Pizza Ops page renders
HEAD / with pizza Host      -> HTTP/1.1 200 OK
GET  /api/tally             -> {"delivered":0,"orders":2,"paid":0,"revenueCents":0,"tipCents":0}
Rendered HTML               -> no generated "Move" controls
Rendered HTML               -> contains "Completed kitchen work" grouped Done cards
```
