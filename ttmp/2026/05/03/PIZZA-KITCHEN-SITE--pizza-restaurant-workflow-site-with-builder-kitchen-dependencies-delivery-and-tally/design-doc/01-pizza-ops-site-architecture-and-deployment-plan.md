---
Title: Pizza Ops Site Architecture and Deployment Plan
Slug: pizza-ops-site-architecture-and-deployment-plan
Short: Design for a richer pizza restaurant goja-site example with a pizza builder, dependent kitchen tasks, delivery workflow, and paid/tips tally.
Topics:
- goja-site
- javascript
- kanban
- sqlite
- documentation
DocType: design-doc
Ticket: PIZZA-KITCHEN-SITE
Status: active
Intent: long-term
Created: 2026-05-03
Updated: 2026-05-03
---

# Pizza Ops Site Architecture and Deployment Plan

## Executive Summary

Pizza Ops is a richer example site for `goja-site`. It exists to demonstrate that these hosted JavaScript sites are not merely restyled Kanban boards. The site combines a normal server-rendered pizza builder form, SQLite-backed order state, a dependency-aware kitchen task board, a delivery status board, and a business tally for paid orders and tips.

The key teaching point is that `ui.dsl` and `kanban.dsl` can be mixed in the same application. The pizza builder is ordinary `ui.dsl`; it creates domain records. The kitchen and delivery boards are `kanban.dsl`; they expose domain transitions as server-side JavaScript callbacks. The app owns the rules: task dependencies, delivery gating, payment, revenue, and tips.

## Problem Statement

The existing production examples show useful styling differences, but they still share the same basic shape: one board, one table, cards moving through columns. That under-sells the architecture. A new developer should see an example where JavaScript coordinates multiple views over one domain model.

Pizza Ops should answer these questions:

- Can a page contain ordinary forms and Kanban boards together? Yes.
- Can one submitted form create many downstream cards? Yes.
- Can server-side callbacks reject invalid moves? Yes.
- Can one board affect another board? Yes; kitchen completion advances delivery.
- Can the app keep business metrics outside the board itself? Yes; paid orders update sales and tips.

## Proposed Solution

Add a new multi-site app at:

```text
sites/pizza
```

Deploy it at:

```text
https://pizza.kanban.yolo.scapegoat.dev/
```

The site has four script layers loaded in lexical order:

```text
00_domain.js   shared menu data, columns, task templates, utility functions
01_styles.js   readable CSS string for the page
02_repository.js  low-level SQLite schema and query helpers
03_workflow.js    order creation, dependency checks, automatic delivery transitions, move rules, tally
04_views.js       ui.dsl rendering and Kanban board builders
05_routes.js      Express routes and board mounting
```

Each file attaches its portion to `globalThis.Pizza`. This keeps the example readable without requiring local CommonJS file resolution. It also makes the load order explicit, which matches the current `goja-site` script loader.

## Domain Model

The site stores two related entities.

### `pizza_orders`

An order is the customer-facing pizza and delivery workflow item.

Important fields:

- `session_id`: isolates demo data by browser session.
- `customer`, `address`: order identity and delivery destination.
- `size`, `crust`, `sauce`, `toppings`, `notes`: pizza builder choices.
- `delivery_status`: current delivery board column.
- `total_cents`: calculated menu total.
- `tip_cents`: captured payment tip.
- `paid`: whether the order has been paid.
- `position`: ordering inside the delivery Kanban column.

### `pizza_tasks`

A task is one kitchen step for one order.

Important fields:

- `order_id`: parent order.
- `code`: stable dependency key such as `stretch` or `bake`.
- `title`, `description`, `station`: card rendering data.
- `deps`: comma-separated dependency task codes.
- `status`: current kitchen board column.
- `position`: ordering inside the kitchen Kanban column.

The initial dependency chain is:

```text
stretch -> sauce -> cheese -> toppings -> bake -> box
```

## Interaction Flow

### 1. Build a pizza

The customer fills out a `ui.dsl` form. Submitting `POST /orders` inserts one `pizza_orders` row and six `pizza_tasks` rows.

The first task starts as `ready`; all dependent tasks start as `blocked`.

### 2. Work kitchen tasks

The kitchen board exposes four columns:

```text
Blocked | Ready | Working | Done
```

The `actions.cardMoved` callback calls `Pizza.store.moveKitchenTask(event)`. That function checks dependencies before allowing movement into `ready`, `working`, or `done`.

If a cook tries to move `bake` before `toppings` is done, the callback returns:

```javascript
{ ok: false, error: "Dependencies first: toppings" }
```

The generic Kanban runtime shows the failure and leaves the source of truth unchanged.

### 3. Advance delivery state

The delivery board exposes six columns:

```text
Waiting | Cooking | Quality | Out | Delivered | Paid
```

Kitchen activity affects delivery automatically:

- once any kitchen task is `working` or `done`, the order moves to `cooking` if it was still waiting;
- once all kitchen tasks are `done`, the order moves to `quality` if it was still before quality.

The delivery board also rejects moving an order into `quality`, `out`, `delivered`, or `paid` until all kitchen tasks are done.

### 4. Capture payment and tips

Payment is not just a drag/drop action. The order card includes a payment form. Submitting it captures a tip, marks the order paid, moves the delivery status to `paid`, and updates the tally.

The tally is derived from paid orders:

- total orders,
- delivered orders,
- paid orders,
- paid sales,
- tips.

## Design Decisions

### Use multiple script files

The pizza app is intentionally split into named layers. Example code should teach structure, not hide logic in dense callbacks. The loader already walks scripts recursively and sorts by filename, so numbered files are a simple convention.

### Keep domain rules in store functions

Kanban callbacks should be thin. They delegate to `Pizza.store.moveKitchenTask` and `Pizza.store.moveDeliveryOrder`. This makes business rules testable through HTTP/API smoke tests and easy for readers to find.

### Use `ui.dsl` for the builder

The pizza builder is not a Kanban feature. It is an ordinary server-rendered form. This demonstrates that `kanban.dsl` is a specialized component within a larger page, not the whole application framework.

### Use two boards over one domain

The kitchen board and delivery board read different projections of the same SQLite state. This demonstrates why server-side JavaScript callbacks are useful: one action can update multiple views without writing browser-specific application JavaScript.

### Payment uses a form instead of drag/drop

Payment has additional data: the tip. A card move cannot capture that well. Keeping payment as a form shows that ordinary HTML interactions can live inside Kanban cards.

## Deployment Plan

1. Add `sites/pizza/scripts` and `sites/pizza/README.md`.
2. Add pizza to `deploy/sites.yaml` and `deploy/sites.local.yaml`.
3. Add pizza to the K3s ConfigMap in the GitOps repo.
4. Add `pizza.kanban.yolo.scapegoat.dev` to the K3s Ingress rules and TLS hosts.
5. Validate locally with `serve-multi` and Host header routing.
6. Commit and push app changes; the app workflow publishes a new GHCR image and opens a GitOps PR.
7. Commit/push K3s manifest changes or update the automated PR branch as needed.
8. Merge the GitOps PR and validate the public URL.

DNS does not need a new Terraform record because `*.kanban.yolo.scapegoat.dev` already points to the cluster ingress IP.

## Validation Checklist

Local checks:

```bash
node -c sites/pizza/scripts/*.js
go test ./... -count=1
go run ./cmd/goja-site serve-multi --config deploy/sites.local.yaml
curl -H 'Host: pizza.kanban.yolo.scapegoat.dev' http://127.0.0.1:60131/
curl -H 'Host: pizza.kanban.yolo.scapegoat.dev' http://127.0.0.1:60131/api/tally
```

Behavior checks:

- The pizza builder creates an order and task chain.
- Blocked kitchen tasks cannot move to working/done before dependencies are done.
- Completing all kitchen tasks moves the order to quality.
- Delivery cannot move to later stages before kitchen completion.
- Payment captures tips and updates the tally.
- `HEAD /` works through the existing GET fallback.

## Alternatives Considered

### One dense `app.js`

Rejected. The user explicitly wants examples to be readable, not golfed. A single dense file also makes it hard to teach how the domain model, rendering, and routes relate.

### Separate browser JavaScript app

Rejected for this example. The purpose is to demonstrate server-rendered `goja-site` capabilities and the Go-owned Kanban runtime. Adding a custom browser SPA would obscure the core architecture.

### One board with mixed kitchen and delivery cards

Rejected. Two boards make it clearer that different projections can share one domain model and that one callback can affect another view.
