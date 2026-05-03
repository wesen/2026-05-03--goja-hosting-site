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
sites/pizza/scripts/02_repository.js
sites/pizza/scripts/03_workflow.js
sites/pizza/scripts/04_views.js
sites/pizza/scripts/05_routes.js
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

## Step 3: Repository/workflow split and automatic delivery transitions

The pizza app was refactored so the growing workflow logic has clearer boundaries:

```text
02_repository.js  owns SQLite schema and low-level queries/mutations
03_workflow.js    owns business rules, dependency refresh, Kanban moves, payment, and tally
04_views.js       owns rendering and board builders
05_routes.js      owns HTTP routes and board mounting
```

This separates persistence from domain transitions. `Pizza.repo` does not decide whether an order should be cooking or quality; it only reads and writes rows. `Pizza.store` is the workflow facade used by views and routes.

The automatic delivery transitions were tightened:

- When any task for a pizza leaves the initial ready/not-started flow and becomes `working` or `done`, the order is moved to `cooking` automatically if it is still waiting.
- When all tasks for a pizza are `done`, the order is moved to `quality` automatically if it is still before quality.
- `listOrders()` and `listTasks()` both refresh order transitions before returning data, so JSON APIs and rendered boards agree.

Validation used a cookie jar to keep one session stable:

```text
Move Ken/stretch -> working  => Ken delivery_status becomes cooking
Move all Ken tasks -> done   => Ken delivery_status becomes quality
Rendered page Done column    => two "Completed kitchen work" grouped cards
```
