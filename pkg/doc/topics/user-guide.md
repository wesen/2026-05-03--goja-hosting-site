---
Title: "User Guide: Writing Your Own goja-site Pages"
Slug: "user-guide"
Short: "Design complete goja-site applications with routes, SQLite data, sessions, UI nodes, Kanban boards, and static assets."
Topics:
- goja-site
- user-guide
- javascript
- sites
- kanban
Commands:
- goja-site
- serve
- serve-multi
Flags:
- scripts
- db
- config
- dev
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

A goja-site application is easiest to reason about as three layers that run in one process. The outer layer is Go: it owns the HTTP server, the SQLite connection, request parsing, response writing, sessions, and module registration. The middle layer is your trusted JavaScript: it declares schema, routes, queries, page functions, and callbacks. The inner layer is the browser: it receives HTML, CSS, forms, and the generic Kanban runtime when a board asks for it.

This guide is for site authors. You do not need to understand the Go internals before creating useful pages. You do need to understand where state lives, how routes are registered, and how to keep your JavaScript honest about user input.

## The site lifecycle

When `goja-site serve` starts, it creates one runtime and loads every script in the script directory. Those scripts run immediately. This startup phase is where you create tables, define helper functions, configure the database guard, and register routes.

```javascript
const db = require("database");
const express = require("express");
const ui = require("ui.dsl");

const app = express.app();

db.exec("CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, title TEXT NOT NULL)");

app.get("/", (req, res) => {
  const notes = db.query("SELECT * FROM notes ORDER BY id DESC");
  res.html(page(notes));
});
```

Requests do not reload the script. They call the handlers that were registered during startup. That means module-level variables persist for the life of the process, and SQLite is the right place for durable data.

## Routes and responses

The `express` module is intentionally small. It gives you the route methods needed for compact server-rendered apps: `get`, `post`, `put`, `patch`, `delete`, `all`, and `static`.

```javascript
app.get("/cards/:id", (req, res) => {
  const card = db.query("SELECT * FROM cards WHERE id = ?", req.params.id)[0];
  if (!card) return res.status(404).send("not found");
  res.html(cardPage(card));
});

app.post("/cards", (req, res) => {
  db.exec("INSERT INTO cards(title) VALUES (?)", String(req.body.title || "Untitled"));
  res.redirect("/");
});
```

The request object contains `method`, `url`, `path`, `query`, `params`, `headers`, `cookies`, `session`, `ip`, `body`, and `rawBody`. The response object can set status and headers, send JSON, send text, render HTML nodes, redirect, or end the response.

## Rendering pages with `ui.dsl`

The UI DSL turns JavaScript function calls into an HTML tree. A call may start with an attributes object, followed by children. Children can be strings, nodes, fragments, arrays, or values that become text.

```javascript
function page(notes) {
  return ui.page(
    { title: "Notes" },
    ui.style("body { font-family: system-ui; max-width: 760px; margin: 3rem auto; }") ,
    ui.main(
      ui.h1("Notes"),
      ui.form({ method: "post", action: "/cards" },
        ui.input({ name: "title", placeholder: "New note" }),
        ui.button({ type: "submit" }, "Add")
      ),
      ui.ul(notes.map(note => ui.li(note.title)))
    )
  );
}
```

The `ui.page(...)` helper separates head-like tags such as `style`, `link`, `meta`, and `title` from body content. This keeps simple pages readable without requiring you to assemble a full document by hand.

## Building Kanban pages

The Kanban DSL exists because drag/drop boards have a lot of repeated mechanics. Site authors should describe columns, data, rendering, and callbacks; the Go-owned runtime should handle common browser behavior.

```javascript
const kanban = require("kanban.dsl");

const board = kanban.board("work")
  .title("Work Board")
  .columns(cols => cols
    .column("todo").title("Todo").done()
    .column("done").title("Done").terminal(true).done())
  .data(data => data
    .cards(ctx => listCards(ctx.session, ctx.query))
    .id(card => String(card.id))
    .column(card => card.status)
    .position(card => Number(card.position || 0))
    .searchText(card => `${card.title} ${card.description}`.toLowerCase()))
  .features(features => features.search({ mode: "client" }).preciseMove().dragDrop())
  .render(render => render.card(card => ui.div(ui.strong(card.title), ui.p(card.description))))
  .actions(actions => actions.cardMoved(event => {
    moveCard({ session: event.session, id: event.cardId, toStatus: event.to.columnId, toIndex: event.to.index });
    return { ok: true, refresh: true };
  }))
  .build();

board.mount(app, "/_kanban");
```

The board receives render contexts with `ctx.session` and action events with `event.session`. Use those values when querying the database. The browser runtime sends action envelopes, not arbitrary code; your JavaScript callback remains the place where domain rules are enforced.

## Static assets

Use `app.static(prefix, dir)` for images, CSS, and other files that should be served directly from disk.

```javascript
app.static("/assets", "sites/trail/assets");
```

Then reference assets from `ui.dsl` nodes:

```javascript
ui.img({ src: "/assets/trail-map.png", alt: "Trail map" })
```

## Single-site and multi-site serving

Use `serve` while developing one site:

```bash
goja-site serve --scripts sites/trail/scripts --db data/trail.db --addr :8080 --dev
```

Use `serve-multi` when one process should host multiple isolated sites by Host header:

```yaml
addr: :8080
dataDir: /data/sites
baseDomain: kanban.yolo.scapegoat.dev
sites:
  - name: trail
    scriptsDir: sites/trail/scripts
  - name: editorial
    scriptsDir: sites/editorial/scripts
```

Each site gets its own runtime, route host, SQLite database, session-aware request handling, `ui.dsl`, `kanban.dsl`, and `db.guard` instance. This is the deployment shape used for the production Kanban examples.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| Form submissions produce empty bodies. | The request content type is missing or unsupported. | Use normal HTML forms or JSON requests; inspect `req.rawBody` while debugging. |
| A multi-site request returns an unknown host error. | The Host header does not match any normalized site host. | Check `baseDomain`, explicit `host`, ingress hosts, and local curl `-H 'Host: ...'`. |
| Existing production data misses new columns. | The script schema changed after the SQLite file already existed. | Add additive migrations with `ALTER TABLE ... ADD COLUMN` and ignore duplicate-column errors. |
| A page gets slow as data grows. | The script queries too much or filters in JavaScript. | Add SQL indexes and push filtering into SQL before rendering. |

## See Also

- `getting-started` walks through the first local run.
- `js-api-reference` documents module functions and callback shapes.
- `developer-guide` explains the Go internals behind the user-facing model.
