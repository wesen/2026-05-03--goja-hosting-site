# Goja Kanban Example

Run the example website:

```bash
go run ./cmd/goja-site serve \
  --db examples/kanban/kanban.db \
  --scripts examples/kanban/scripts \
  --addr :8080 \
  --dev
```

Open <http://localhost:8080/>.

The page is registered from server-side JavaScript, persisted in SQLite through
`require("database")`, routed through `require("express")`, rendered as HTML
through `require("ui.dsl")`, and now uses `require("kanban.dsl")` for standard
Kanban behavior.

The app still owns the domain-specific pieces:

- the SQLite schema and queries,
- the session-scoped card queries using `req.session.id` / `ctx.session.id`,
- the Field Notes card rendering,
- the `cardMoved` server-side callback,
- the page chrome around the board.

The app no longer serves its own browser Kanban runtime. Calling
`board.mount(app, "/_kanban")` registers the generic DSL-owned frontend script at
`/_kanban/client.js` plus board fragment/action routes such as:

- `GET /_kanban/client.js`
- `GET /_kanban/trail-notes/fragment`
- `POST /_kanban/trail-notes/action/cardMoved`

That script provides live search, precise move form submission, drag/drop wiring,
action dispatch, and server-rendered fragment replacement without app-specific
client-side JavaScript.

The Go host also issues an opaque `goja_site_session` cookie for dynamic routes.
JavaScript receives it as `req.session.id`; mounted Kanban action callbacks and
render contexts receive it as `event.session.id` / `ctx.session.id`. The example
stores `session_id` on each card so different browsers get separate demo boards
while the app only has to mention the session ID where it queries or mutates the
database.
