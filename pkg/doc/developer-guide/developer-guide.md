---
Title: "Developer Guide: goja-site Internals"
Slug: "developer-guide"
Short: "Understand the Go architecture behind goja-site so you can contribute safely."
Topics:
- goja-site
- developer-guide
- architecture
- goja
- glazed
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

goja-site is small enough to understand in one sitting, but it crosses several boundaries: a Glazed/Cobra CLI starts a Go HTTP server; the server creates Goja runtimes; native modules expose Go services to JavaScript; JavaScript registers routes; requests cross back into Goja callbacks; responses are rendered by Go. The project is a useful study in keeping those boundaries explicit.

The central design principle is composition. The host does not try to become a complete framework. It composes a runtime, route host, SQLite connection, session manager, UI renderer, Kanban runtime, and database guard. Each piece has a narrow job, and most features are added by registering another native module or wrapping an existing interface.

## CLI entry point

The binary starts in `cmd/goja-site/main.go`. It builds a Cobra root command, adds Glazed logging flags, creates Glazed command descriptions for `serve` and `serve-multi`, and converts them into Cobra commands.

```text
cmd/goja-site/main.go
  ├─ newServeCommand()       -> cmd/goja-site/serve.go
  └─ newServeMultiCommand()  -> cmd/goja-site/serve_multi.go
```

The help pages you are reading are embedded through `pkg/doc` and loaded into a Glazed `HelpSystem`. The root command then calls `help_cmd.SetupCobraRootCommand(...)`, which adds the rich `goja-site help ...` command tree.

## Single-site server

`pkg/app/server.go` owns the single-site path. Conceptually, startup does this:

```text
Config
  -> open SQLite database
  -> create route Host
  -> create Goja runtime
  -> register native modules
  -> load scripts from ScriptsDir
  -> expose net/http Handler
```

The server registers modules such as `express`, `ui.dsl`, `kanban.dsl`, `database`, `db`, and `db.guard`. The JavaScript files then run and register routes against the host. After startup, ordinary HTTP requests are dispatched through the host to the registered Goja callbacks.

This order matters. Routes are not discovered dynamically per request. Scripts declare the site during startup, which keeps request dispatch simple and makes startup errors visible early.

## Route host and request dispatch

`pkg/web` implements the Express-style surface. `express_module.go` exposes `express.app()` and route registration. `host.go` stores route entries, matches methods and path patterns, serves static directories, manages session DTOs, and calls JavaScript handlers.

`request_response.go` translates between Go's `net/http` objects and JavaScript-friendly DTOs:

- `NewRequestDTO(...)` parses query strings, route params, cookies, headers, body, remote IP, and session data.
- `Response.JSObject(...)` exposes `status`, `set`, `type`, `json`, `send`, `html`, `redirect`, and `end`.
- `res.html(...)` delegates rendering to the configured UI renderer.

A subtle but important behavior lives here: `HEAD` can fall back to `GET` when no explicit `HEAD` route exists. The fallback uses a response writer that preserves headers and status while discarding the body.

## UI DSL

`pkg/uidsl` implements `require("ui.dsl")`. The module exports tag functions, `fragment`, `text`, `raw`, `render`, and `page`. Tag calls produce Go node structs rather than strings. Rendering happens later, when the response asks for HTML.

This design has two practical benefits:

- JavaScript page functions remain ordinary functions that return data structures.
- Go can centralize escaping, document rendering, and normalization rules.

When extending the UI DSL, prefer adding a tag or helper that fits the existing node model. Avoid adding helpers that secretly write to the response or depend on request-global state.

## Kanban DSL

`pkg/kanbanddsl` is the largest domain-specific module. The builder in `builder.go` gathers a `BoardConfig` through fluent JavaScript calls. `Build()` validates the configuration before registering the board. Validation is part of the developer experience: a missing `data.id(...)` or a drag/drop board without `actions.cardMoved(...)` fails early with a readable error.

The mounted board exposes three kinds of routes:

```text
/_kanban/client.js
/_kanban/<boardId>/fragment
/_kanban/<boardId>/action/:action
```

The browser runtime is generic and data-attribute driven. It does not know the application schema. It reports user actions to Go; Go dispatches them to the JavaScript callback; the callback mutates application state; the server renders a fresh fragment. This is the core loop:

```text
browser drag/drop
  -> POST action envelope
  -> Go dispatch
  -> JS actions.cardMoved(event)
  -> SQLite mutation
  -> server-rendered fragment
  -> browser replaces board fragment
```

When contributing to this package, keep the division intact. App-specific rules belong in callbacks. Generic interaction mechanics belong in the runtime and mount layer.

## Database guard

`pkg/dbguard` wraps the SQLite query executor rather than modifying the upstream database module. This was a deliberate architectural decision. The guard measures database files, WAL files, page counts, and freelist counts; it can call JavaScript cleanup callbacks; and it can optionally reject growth writes when a hard limit is exceeded.

The important contribution rule is to preserve cleanup escape hatches. When a database is over the hard limit, cleanup and maintenance SQL such as `DELETE`, `DROP`, `VACUUM`, `PRAGMA`, `ANALYZE`, and `REINDEX` must remain possible. A guard that blocks cleanup is worse than no guard.

## Multi-site server

`pkg/app/multi_config.go` and `pkg/app/multi_server.go` implement host-based multi-site serving. A multi-site server creates one isolated site instance per configured site. Isolation means separate runtime, route host, database path, Kanban registry, and database guard.

The outer server only chooses the site from the normalized Host header. It does not share route tables between sites. This is why the K3s deployment can serve `trail`, `editorial`, and `crm` from one pod while keeping one SQLite DB per site.

## Adding features safely

A good contribution usually follows one of these patterns:

| Feature kind | Preferred place |
| --- | --- |
| New JavaScript-visible capability | A focused native module registrar. |
| New route behavior | `pkg/web`, with integration tests. |
| New HTML node behavior | `pkg/uidsl`, with render tests. |
| New Kanban interaction | `pkg/kanbanddsl`, with builder/mount/runtime tests. |
| New storage policy | `pkg/dbguard`, with guard tests and live smoke tests where needed. |
| New CLI behavior | A Glazed command in `cmd/goja-site`. |

The project already has tests for route dispatch, sessions, multi-site isolation, Kanban builder/mount behavior, UI rendering, and database guard limits. Add tests near the boundary you change.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| A module is unavailable in JavaScript. | Its registrar was not added to the runtime setup. | Inspect `pkg/app/server.go` and confirm the registrar is registered before scripts load. |
| A test passes alone but fails in a package run. | Global runtime or board state is being reused unexpectedly. | Prefer per-test runtimes and unique board IDs; reset temp DB paths. |
| A route change breaks HEAD checks. | The fallback writer or method matching changed. | Run `go test ./pkg/web -count=1` and add an integration case. |
| A Kanban runtime change works in one app but not another. | The runtime assumed an app-specific DOM shape. | Use `data-kb-*` attributes and validate against trail, editorial, and CRM examples. |

## See Also

- `getting-started` and `user-guide` explain the public model that internals must preserve.
- `js-api-reference` is the contract that site scripts rely on.
- Source files to read first: `cmd/goja-site/main.go`, `pkg/app/server.go`, `pkg/web/host.go`, `pkg/uidsl/module.go`, `pkg/kanbanddsl/builder.go`, and `pkg/dbguard/guard.go`.
