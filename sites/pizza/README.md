# Pizza Ops Example Site

Pizza Ops is a richer `goja-site` example than a plain Kanban board. It shows how a trusted JavaScript site can combine ordinary server-rendered UI, SQLite domain state, and two coordinated Kanban boards.

The app has four visible pieces:

1. **Pizza builder** — a `ui.dsl` form where a customer chooses size, crust, sauce, toppings, address, and notes.
2. **Kitchen dependency board** — each order becomes a sequence of kitchen tasks. A task can only move forward when its dependencies are done.
3. **Delivery board** — each order also appears as a delivery workflow: waiting, cooking, quality check, out for delivery, delivered, paid.
4. **Tally** — paid orders contribute to sales and tip totals.

## Run locally

From the repository root:

```bash
go run ./cmd/goja-site serve \
  --db ./tmp/pizza/app.db \
  --scripts ./sites/pizza/scripts \
  --addr :8080 \
  --dev
```

Open <http://localhost:8080/>.

For multi-site testing:

```bash
go run ./cmd/goja-site serve-multi --config deploy/sites.local.yaml
curl -H 'Host: pizza.kanban.yolo.scapegoat.dev' http://127.0.0.1:60131/
```

## Script layout

The scripts are intentionally split for readability. `goja-site` loads all `.js` files in lexical order, so each file attaches its part to `globalThis.Pizza`.

```text
sites/pizza/scripts/
  00_domain.js   menu data, columns, task template, small utility functions
  01_styles.js   readable CSS for the page and boards
  02_repository.js  low-level SQLite schema and query helpers
  03_workflow.js    order creation, dependency checks, automatic delivery transitions, move rules, tally
  04_views.js       ui.dsl page rendering and Kanban board builders
  05_routes.js      Express-style routes and board mounting
```

This is the preferred style for larger examples: domain logic and rendering are named, formatted, and easy to inspect.

## Dependency model

Every pizza order creates this task chain:

```text
stretch -> sauce -> cheese -> toppings -> bake -> box
```

A kitchen task starts as `blocked` unless it has no dependencies. Whenever tasks change, the store refreshes the order and unlocks tasks whose dependencies have reached `done`.

The kitchen board rejects invalid moves. For example, `bake` cannot move to `working` or `done` until `toppings` is done.

## Delivery model

The delivery board tracks the whole order. It can move through early states manually, but it cannot move into `quality`, `out`, `delivered`, or `paid` until all kitchen tasks are done.

Payment is intentionally a form on the order card, not a drag/drop shortcut. The payment form captures a tip, marks the order paid, moves it to the `paid` column, and updates the tally.

## APIs

The example exposes JSON endpoints for validation and demos:

```text
GET /api/tally
GET /api/orders
GET /api/tasks
```

The page itself is still server-rendered HTML. The Kanban browser behavior comes from `kanban.dsl`'s Go-owned runtime mounted under `/_kanban`.
