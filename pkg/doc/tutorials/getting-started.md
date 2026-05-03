---
Title: "Getting Started with goja-site"
Slug: "getting-started"
Short: "Build and run your first trusted JavaScript website with goja-site."
Topics:
- goja-site
- getting-started
- javascript
- sqlite
Commands:
- goja-site
- serve
Flags:
- addr
- db
- scripts
- dev
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

A `goja-site` program is a small web application written in trusted server-side JavaScript. The Go process supplies the boring but important machinery: an HTTP listener, route dispatch, a SQLite connection, an HTML rendering DSL, a Kanban board DSL, static file serving, and an opaque session cookie. Your JavaScript supplies the application-specific choices: routes, schema, queries, HTML structure, and callbacks.

The fastest way to learn the model is to run one site, then read the script that registered it. This is different from learning a browser framework. There is no build step, no client bundle that you own, and no API server separate from the page renderer. The JavaScript script is loaded by Go, it registers routes, and those routes return HTML or JSON.

## 1. Run the example

Start with the Kanban example because it exercises most of the system in one place: routing, SQLite, server-rendered HTML, sessions, and the Go-owned Kanban runtime.

```bash
go run ./cmd/goja-site serve \
  --db examples/kanban/kanban.db \
  --scripts examples/kanban/scripts \
  --addr :8080 \
  --dev
```

Then open:

```text
http://localhost:8080/
```

The `--scripts` directory is loaded into one Goja runtime. The `--db` path becomes the SQLite database exposed to JavaScript through `require("database")` and `require("db")`. The `--dev` flag asks the server to show detailed route errors in HTTP responses, which is useful while learning and unsafe as a production habit.

## 2. Read the route shape

A minimal site imports the Express-style router and the HTML DSL, creates an app, and registers a route.

```javascript
const express = require("express");
const ui = require("ui.dsl");

const app = express.app();

app.get("/", (req, res) => {
  res.html(ui.page(
    { title: "Hello goja-site" },
    ui.main(
      ui.h1("Hello goja-site"),
      ui.p("This page was rendered by trusted server-side JavaScript.")
    )
  ));
});
```

The important idea is that `res.html(...)` accepts a `ui.dsl` node, not a string template. The Go host renders the node tree into HTML. That gives you the convenience of JavaScript functions and ordinary control flow without mixing string concatenation into every page.

## 3. Add persistence

Most useful sites need state. The built-in database module exposes SQL functions backed by the configured SQLite file.

```javascript
const db = require("database");

db.exec(`CREATE TABLE IF NOT EXISTS notes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
)`);

app.post("/notes", (req, res) => {
  db.exec("INSERT INTO notes(title) VALUES (?)", String(req.body.title || "Untitled"));
  res.redirect("/");
});
```

Use placeholders for values. The script is trusted code, but request data is still untrusted input. Treat `req.body`, `req.query`, and route parameters as data, never as SQL text.

## 4. Understand sessions

Dynamic routes receive an opaque session object at `req.session`. The session ID is generated and signed by the Go host and stored in the `goja_site_session` cookie. Your app does not create the session table; it decides how to use the session ID in its own tables.

```javascript
function sessionId(session) {
  return String(session && session.id || "default");
}

const rows = db.query(
  "SELECT * FROM cards WHERE session_id = ? ORDER BY position",
  sessionId(req.session)
);
```

This pattern keeps the host simple. Go owns identity continuity; the JavaScript application owns domain state.

## 5. What to read next

After this first run, read the user guide before writing a full site. It explains the site lifecycle, single-site versus multi-site serving, static assets, session-scoped data, and production deployment shape.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| The server starts but `/` returns 404. | The script did not register a matching route, or the script directory is wrong. | Check `--scripts`, run with `--dev`, and confirm that your script calls `app.get("/", ...)`. |
| A route panics with a SQL error. | The table has not been created, a migration failed, or a query has the wrong placeholders. | Put `CREATE TABLE IF NOT EXISTS` at script startup and keep placeholder count aligned with arguments. |
| Browser state appears shared unexpectedly. | The app did not include `req.session.id` in its table queries. | Add a `session_id` column and include it in every read/write that should be session-scoped. |
| A Kanban board renders but drag/drop does nothing. | The board was rendered but not mounted, or the client script route is missing. | Call `board.mount(app, "/_kanban")` and include the board returned by `board.render(...)`. |

## See Also

- `user-guide` explains how to design complete sites.
- `js-api-reference` lists the JavaScript modules exposed to site scripts.
- `developer-guide` explains how the Go host loads scripts and dispatches requests.
