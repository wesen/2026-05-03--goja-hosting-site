---
Title: "JavaScript API Reference"
Slug: "js-api-reference"
Short: "Reference for the JavaScript modules exposed by goja-site: express, ui.dsl, kanban.dsl, database/db, db.guard, sessions, and trusted utility modules."
Topics:
- goja-site
- javascript
- api-reference
- ui-dsl
- kanban-dsl
- sqlite
Commands:
- goja-site
- serve
- serve-multi
Flags: []
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This reference describes the API that site scripts can `require(...)`. It is written from the point of view of a trusted application author: you are allowed to write server-side JavaScript that reads files, talks to SQLite, registers HTTP routes, and renders HTML. The trust boundary is not between Go and JavaScript. The trust boundary is between your server-side code and the users sending HTTP requests.

The APIs are deliberately small. A small API is easier to remember, easier to secure, and easier to teach. When you need a new capability, prefer adding a focused native module rather than turning the route layer into a full web framework.

## `require("express")`

The Express-style module registers HTTP handlers on the current site host.

```javascript
const express = require("express");
const app = express.app();
```

| Function | Purpose |
| --- | --- |
| `app.get(pattern, handler)` | Register a GET route. |
| `app.post(pattern, handler)` | Register a POST route. |
| `app.put(pattern, handler)` | Register a PUT route. |
| `app.patch(pattern, handler)` | Register a PATCH route. |
| `app.delete(pattern, handler)` | Register a DELETE route. |
| `app.all(pattern, handler)` | Register a route for any method. |
| `app.static(prefix, dir)` | Serve files from `dir` under URL `prefix`. |

Route parameters use colon segments:

```javascript
app.get("/cards/:id", (req, res) => {
  res.json({ id: req.params.id });
});
```

If no explicit `HEAD` route exists, a matching `GET` route is used and the body is suppressed. This lets uptime checks and browsers behave as expected without requiring every page to register duplicate HEAD routes.

## Request object

Handlers receive `(req, res)`. The request is a plain JavaScript object.

| Field | Meaning |
| --- | --- |
| `req.method` | HTTP method. |
| `req.url` | Full request URL path and query. |
| `req.path` | URL path without query. |
| `req.query` | Query parameters. Single values are strings; repeated values are arrays. |
| `req.params` | Route parameters captured from the route pattern. |
| `req.headers` | Request headers as lower/normalized Go header names joined into strings. |
| `req.cookies` | Cookie name/value map. |
| `req.session` | Opaque session DTO with `id`, `isNew`, and `cookieName`. |
| `req.ip` | Remote IP as seen by the Go server. |
| `req.body` | Parsed body for supported form/JSON requests. |
| `req.rawBody` | Raw body string. |

The session object is intentionally boring:

```javascript
{
  id: "opaque-random-token",
  isNew: false,
  cookieName: "goja_site_session"
}
```

Use `req.session.id` as a key in your own tables when data should be browser/session scoped.

## Response object

The response object is a small writer API. Once a response is sent, later writes are ignored.

| Function | Purpose |
| --- | --- |
| `res.status(code)` | Set status and return `res` for chaining. |
| `res.set(name, value)` | Set a response header and return `res`. |
| `res.type(value)` | Set `Content-Type` and return `res`. |
| `res.json(value)` | Send JSON. |
| `res.send(value)` | Send text, HTML-looking strings, or JSON for non-strings. |
| `res.html(node)` | Render a `ui.dsl` node or document and send HTML. |
| `res.redirect(url)` | Redirect with status 302. |
| `res.redirect(status, url)` | Redirect with an explicit status. |
| `res.end()` | Send headers/status with no body. |

## `require("ui.dsl")`

The UI DSL exposes HTML tag functions plus helpers.

```javascript
const ui = require("ui.dsl");
```

A tag function accepts an optional attributes object followed by children:

```javascript
ui.a({ href: "/cards/1", class: "card-link" }, "Open card")
ui.div(ui.h1("Title"), ui.p("Body"))
```

Supported tags include document/head tags, common text and layout tags, forms, tables, and inline code tags: `html`, `head`, `body`, `title`, `meta`, `link`, `script`, `style`, `main`, `img`, `br`, `hr`, `time`, `svg`, `path`, `rect`, `line`, `polyline`, `circle`, `div`, `span`, `h1`, `h2`, `h3`, `h4`, `p`, `a`, `form`, `input`, `button`, `select`, `option`, `ul`, `ol`, `li`, `table`, `thead`, `tbody`, `tr`, `th`, `td`, `section`, `article`, `header`, `footer`, `nav`, `label`, `textarea`, `strong`, `em`, `small`, `pre`, and `code`.

Helpers:

| Helper | Purpose |
| --- | --- |
| `ui.fragment(...children)` | Return children without a wrapper element. |
| `ui.text(value)` | Convert a value to a text node. |
| `ui.raw(html)` | Insert raw HTML. Use sparingly and never with untrusted input. |
| `ui.render(node)` | Render a node to an HTML string. |
| `ui.page(attrs, ...children)` | Build a full HTML document. `attrs.title` sets the page title. |

## `require("kanban.dsl")`

The Kanban DSL builds server-rendered boards with Go-owned browser interactions.

```javascript
const kanban = require("kanban.dsl");
```

The top-level entry is:

```javascript
kanban.board(id)
```

A board builder supports:

| Builder method | Purpose |
| --- | --- |
| `.title(text)` | Set the board title. |
| `.description(text)` | Set descriptive text. |
| `.theme(name)` | Store a theme name for render/use by CSS. |
| `.className(name)` | Add a board-level class. |
| `.attrs(object)` | Add board-level attributes. |
| `.columns(fn)` | Define columns with the column builder. |
| `.data(fn)` | Define card data callbacks. |
| `.features(fn)` | Enable search, drag/drop, precise move, and other features. |
| `.render(fn)` | Customize toolbar, card, column, empty-state, and shell rendering. |
| `.actions(fn)` | Register server-side callbacks. |
| `.build()` | Validate and register the board. |
| `.mount(app, prefix)` | Build, register, and mount board routes in one step. |

Column builder:

```javascript
.columns(cols => cols
  .column("todo").title("Todo").description("Ready to start").limit(10).done()
  .column("done").title("Done").terminal(true).done())
```

Data callbacks:

| Callback | Required | Purpose |
| --- | --- | --- |
| `data.cards(ctx)` | yes | Return all cards for the current render context. |
| `data.id(card)` | yes | Return stable card ID. |
| `data.column(card)` | yes | Return column ID for card. |
| `data.position(card)` | no | Return numeric sort position. |
| `data.searchText(card)` | no | Return searchable text for client/server filtering. |

Feature builder:

```javascript
.features(features => features
  .search({ mode: "client" })
  .preciseMove()
  .dragDrop())
```

Actions include `cardMoved`, `cardCreated`, `cardUpdated`, `cardDeleted`, `cardClicked`, `cardMenuAction`, and `custom(name, fn)`. Common boards only need `cardMoved`.

The mounted board registers:

| Route | Purpose |
| --- | --- |
| `GET /_kanban/client.js` | Generic Go-owned browser runtime. |
| `GET /_kanban/<boardId>/fragment` | Server-rendered board fragment refresh. |
| `POST /_kanban/<boardId>/action/:action` | Action dispatch into registered callbacks. |

## `require("database")` and `require("db")`

Both database modules are backed by the configured SQLite connection. Use them for schema setup and ordinary queries. The examples use:

```javascript
db.exec("CREATE TABLE IF NOT EXISTS cards (id INTEGER PRIMARY KEY, title TEXT NOT NULL)");
db.exec("INSERT INTO cards(title) VALUES (?)", "Write docs");
const rows = db.query("SELECT * FROM cards ORDER BY id DESC");
```

The important rule is to bind values as arguments instead of interpolating request data into SQL strings.

## `require("db.guard")`

The DB guard measures SQLite database size and lets scripts react when limits are crossed.

| Function | Purpose |
| --- | --- |
| `guard.configure(options)` | Configure soft/hard limits and checking behavior. |
| `guard.onLimitExceeded(fn)` | Register cleanup callback. |
| `guard.stats()` | Return current file/page/WAL statistics. |
| `guard.checkNow(reason)` | Force a limit check and return the result. |
| `guard.isOverLimit()` | Return whether the last/current state is over limit. |
| `guard.lastResult()` | Return the last check result. |

Options include `maxBytes`, `softMaxBytes`, `hardMaxBytes`, `cooldownMs`, `checkEveryWrites`, `includeWal`, and `failWritesOverHardLimit`.

## Trusted utility modules

The host also registers trusted utility modules from go-go-goja, including filesystem/path/time/timer-style modules used by server-side scripts. These are intended for trusted site code. Do not expose arbitrary file operations through HTTP routes unless you have built your own authorization layer.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `require("kanban.dsl")` fails. | The script is not running under goja-site or the module registration failed. | Run through `goja-site serve` or inspect server startup errors. |
| `res.html(...)` prints an error-looking string. | A child passed to `ui.dsl` could not be normalized. | Pass strings, UI nodes, fragments, or arrays of nodes; avoid passing arbitrary objects as children. |
| A board builder fails at `.build()`. | Required columns/data callbacks are missing, a column ID is duplicated, or movement features lack `actions.cardMoved`. | Read the validation error; it lists each missing or invalid part. |
| Hard-limit database writes fail. | `db.guard` is configured with `failWritesOverHardLimit`. | Run cleanup SQL, VACUUM where appropriate, or raise the hard limit. |

## See Also

- `getting-started` shows the first runnable site.
- `user-guide` explains how these APIs fit together in a complete application.
- `developer-guide` maps these JavaScript APIs to the Go packages that implement them.
