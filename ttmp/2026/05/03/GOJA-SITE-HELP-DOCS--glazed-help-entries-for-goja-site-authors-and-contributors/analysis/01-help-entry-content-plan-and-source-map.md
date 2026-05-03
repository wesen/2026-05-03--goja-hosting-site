---
Title: Help Entry Content Plan and Source Map
Ticket: GOJA-SITE-HELP-DOCS
Status: active
Topics:
    - documentation
    - glazed
    - goja-site
    - developer-guide
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/goja-site/main.go
      Note: CLI root loads embedded help system
    - Path: pkg/doc/developer-guide/developer-guide.md
      Note: Contributor architecture guide help entry
    - Path: pkg/doc/doc.go
      Note: Embedded Glazed help package planned by the analysis
    - Path: pkg/doc/reference/js-api-reference.md
      Note: JavaScript module API reference help entry
    - Path: pkg/doc/topics/user-guide.md
      Note: Site author user guide help entry
    - Path: pkg/doc/tutorials/getting-started.md
      Note: Getting-started tutorial help entry
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Help Entry Content Plan and Source Map

This document plans the first built-in Glazed help set for `goja-site`. The goal is not to dump every implementation detail into CLI help. The goal is to give new developers a coherent path: first run a site, then build their own page, then consult the JavaScript API surface, then understand the Go architecture well enough to contribute.

The style should be textbook-like: foundational before procedural, concrete examples before abstract claims, and enough explanation of design choices that readers can extend the system without cargo-culting the examples. Each help entry should answer three questions: what mental model should I have, what code can I write now, and where do I look when it breaks?

## Glazed help conventions to follow

The entries should use standard Glazed help frontmatter. The authoritative local references are:

```bash
glaze help writing-help-entries
glaze help how-to-write-good-documentation-pages
```

Important conventions pulled from those references:

- Use exact frontmatter field names: `Title`, `Slug`, `Short`, `Topics`, `Commands`, `Flags`, `IsTopLevel`, `IsTemplate`, `ShowPerDefault`, and `SectionType`.
- Do not add a top-level Markdown `#` heading inside the document body; Glazed renders the title from frontmatter.
- Use `Tutorial` for the step-by-step entry, `GeneralTopic` for conceptual guides, and `GeneralTopic`/reference-style prose for the API reference.
- Each entry should include a troubleshooting table and `See Also` links.
- Slugs should be short and stable because users type them as `goja-site help <slug>`.

## Proposed help entry set

| Entry | Slug | SectionType | Primary reader | Purpose |
| --- | --- | --- | --- | --- |
| Getting Started with goja-site | `getting-started` | `Tutorial` | New user evaluating the tool | Run the example, understand the first route, add persistence, and learn the session idea. |
| User Guide: Writing Your Own goja-site Pages | `user-guide` | `GeneralTopic` | Site author | Explain the application lifecycle, routes, UI rendering, Kanban boards, static assets, and single-site/multi-site shape. |
| JavaScript API Reference | `js-api-reference` | `GeneralTopic` | Site author needing exact API names | Reference the exposed modules and object shapes: `express`, request/response, `ui.dsl`, `kanban.dsl`, `database`/`db`, `db.guard`, sessions, and utility modules. |
| Developer Guide: goja-site Internals | `developer-guide` | `GeneralTopic` | Go contributor | Explain CLI setup, server startup, module registration, request dispatch, UI rendering, Kanban dispatch, DB guard, and multi-site isolation. |

The four entries should be top-level because they form the initial learning map for the project. Later entries can become deeper examples, such as `kanban-dsl-tutorial`, `db-guard-guide`, `multi-site-deployment`, or `writing-native-modules`.

## Entry 1: Getting Started

The getting-started entry should be a runnable tutorial. It should not begin with architecture. It should begin with the command that produces a visible page:

```bash
go run ./cmd/goja-site serve \
  --db examples/kanban/kanban.db \
  --scripts examples/kanban/scripts \
  --addr :8080 \
  --dev
```

It should teach the smallest useful mental model:

1. `goja-site` starts a Go HTTP server.
2. The server creates a Goja JavaScript runtime.
3. Scripts run at startup and register routes.
4. Route handlers receive `req`/`res` objects.
5. `res.html(...)` renders `ui.dsl` nodes.
6. SQLite is available through `require("database")` / `require("db")`.
7. Session identity comes from `req.session.id`.

Source files and docs to cite or reflect:

- `examples/kanban/README.md` for the validated example command and explanation of mounted Kanban routes.
- `cmd/goja-site/serve.go` for the actual flags and default values.
- `pkg/web/request_response.go` for request/response fields.
- `pkg/uidsl/module.go` for the minimal route rendering example.
- `pkg/web/session.go` for session DTO semantics.

The tutorial should include a tiny route example and a tiny SQLite example. It should close with common beginner failures: wrong scripts directory, route not registered, missing schema, missing session filtering, and forgetting `board.mount(...)`.

## Entry 2: User Guide

The user guide should be the practical guide for people creating their own pages. It should assume the reader can already run the example and now wants to build a site from scratch.

Topics to include:

- Site lifecycle: scripts run once at startup; handlers run per request.
- Route registration with `express.app()`.
- Request and response usage.
- Server-rendered UI through `ui.dsl`.
- Persistent state with SQLite.
- Safe use of request values as SQL parameters.
- Session-scoped tables using `req.session.id`, `ctx.session.id`, and `event.session.id`.
- Static assets through `app.static(prefix, dir)`.
- Kanban boards through the board builder.
- Multi-site serving through `serve-multi` and a YAML config.
- Production expectations: one DB per site, single-writer SQLite deployment shape, and host-based routing.

Source files and docs to cite or reflect:

- `sites/trail/scripts/app.js`, `sites/editorial/scripts/app.js`, and `sites/crm/scripts/app.js` for diverse production examples.
- `deploy/sites.local.yaml` and `deploy/sites.yaml` for multi-site config shape.
- `pkg/app/multi_config.go` for YAML fields and validation rules.
- `pkg/kanbanddsl/builder.go` for builder shape.
- `ttmp/.../KANBAN-K3S-ARGOCD.../design-doc/01-multi-site-goja-kanban-k3s-argo-cd-deployment-guide.md` for production deployment concepts.

The user guide should be explanatory rather than exhaustive. The API reference owns exhaustive lists; the user guide owns how the pieces fit together.

## Entry 3: JavaScript API Reference

The API reference should be organized by module and object shape. It should favor stable contracts over implementation trivia.

Modules and APIs to document:

### `require("express")`

Source: `pkg/web/express_module.go`.

- `express.app()`
- `app.get(pattern, handler)`
- `app.post(pattern, handler)`
- `app.put(pattern, handler)`
- `app.patch(pattern, handler)`
- `app.delete(pattern, handler)`
- `app.all(pattern, handler)`
- `app.static(prefix, dir)`

Also mention HEAD fallback, implemented in `pkg/web/host.go`, because it affects externally visible HTTP behavior.

### Request/response objects

Source: `pkg/web/request_response.go`.

Request fields:

- `method`
- `url`
- `path`
- `query`
- `params`
- `headers`
- `cookies`
- `session`
- `ip`
- `body`
- `rawBody`

Response methods:

- `status(code)`
- `set(name, value)`
- `type(value)`
- `json(value)`
- `send(value)`
- `html(node)`
- `redirect(url)` / `redirect(status, url)`
- `end()`

### `require("ui.dsl")`

Source: `pkg/uidsl/module.go` and `pkg/uidsl/render.go`.

Document tag function convention: optional attrs object plus children. List the supported tags exported by the module. Document helpers:

- `fragment(...children)`
- `text(value)`
- `raw(html)`
- `render(node)`
- `page(attrs, ...children)`

Warn that `raw` should never receive untrusted request data.

### `require("kanban.dsl")`

Source: `pkg/kanbanddsl/builder.go`, `pkg/kanbanddsl/mount.go`, `pkg/kanbanddsl/dispatch.go`, and `pkg/kanbanddsl/client_runtime.go`.

Document builder methods:

- board-level methods: `title`, `description`, `theme`, `className`, `attrs`, `columns`, `data`, `features`, `render`, `actions`, `build`, `mount`.
- column methods: `column`, `title`, `description`, `limit`, `terminal`, `className`, `attrs`, `done`.
- data callbacks: `cards`, `id`, `column`, `position`, `searchText`.
- feature methods: `search`, `preciseMove`, `dragDrop`, `createCard`, `cardMenu`, `readOnly`.
- render callbacks: `card`, `columnHeader`, `toolbar`, `emptyColumn`, `boardShell`.
- action callbacks: `cardMoved`, `cardCreated`, `cardUpdated`, `cardDeleted`, `cardClicked`, `cardMenuAction`, `custom`.

Explain mounted routes:

- `GET /_kanban/client.js`
- `GET /_kanban/<boardId>/fragment`
- `POST /_kanban/<boardId>/action/:action`

### `require("database")` and `require("db")`

Source: upstream `go-go-goja/modules/database` plus local integration in `pkg/app/server.go`.

The help entry should not over-promise details that belong to upstream Glazed/go-go-goja docs, but it should show the project’s common usage: `db.exec(...)` and `db.query(...)` with placeholders.

### `require("db.guard")`

Source: `pkg/dbguard/registrar.go`, `pkg/dbguard/types.go`, `pkg/dbguard/guard.go`.

Document:

- `configure(options)`
- `onLimitExceeded(fn)`
- `stats()`
- `checkNow(reason)`
- `isOverLimit()`
- `lastResult()`

Options:

- `maxBytes`
- `softMaxBytes`
- `hardMaxBytes`
- `cooldownMs`
- `checkEveryWrites`
- `includeWal`
- `failWritesOverHardLimit`

### Sessions and trusted utility modules

Session source: `pkg/web/session.go` and `pkg/web/request_response.go`.

Trusted utility modules are registered by the app server from go-go-goja. The reference should mention them without turning this into a filesystem API manual.

## Entry 4: Developer Guide

The developer guide should map public behavior to Go packages. It is for a contributor who will change code, not merely write site scripts.

Architecture topics:

- CLI command construction through Glazed and Cobra.
- Help system integration through `pkg/doc` and `help_cmd.SetupCobraRootCommand(...)`.
- Single-site server startup in `pkg/app/server.go`.
- Multi-site host dispatch in `pkg/app/multi_server.go`.
- Config normalization in `pkg/app/multi_config.go`.
- Route registration and dispatch in `pkg/web`.
- Request/response DTO boundary in `pkg/web/request_response.go`.
- Session creation and cookie behavior in `pkg/web/session.go`.
- UI node model in `pkg/uidsl`.
- Kanban builder/runtime/mount/action split in `pkg/kanbanddsl`.
- SQLite size guard in `pkg/dbguard`.

The guide should include a contribution map:

| Change | Package |
| --- | --- |
| CLI behavior | `cmd/goja-site` |
| Native JS modules | registrar package + `pkg/app/server.go` registration |
| Route behavior | `pkg/web` |
| UI rendering | `pkg/uidsl` |
| Kanban behavior | `pkg/kanbanddsl` |
| DB guard behavior | `pkg/dbguard` |
| Multi-site config/dispatch | `pkg/app` |

It should also describe the key design constraints from prior work:

- Preserve app-owned domain state.
- Keep sessions opaque and host-owned.
- Keep common Kanban browser interactions Go-owned and generic.
- Prefer composition/wrapping over upstream forks.
- Keep SQLite cleanup/maintenance possible even when hard limits reject growth writes.
- Keep multi-site isolation strong: one runtime and one SQLite DB per site.

## Help implementation plan

Implementation should add:

```text
pkg/doc/doc.go
pkg/doc/tutorials/getting-started.md
pkg/doc/topics/user-guide.md
pkg/doc/reference/js-api-reference.md
pkg/doc/developer-guide/developer-guide.md
```

`cmd/goja-site/main.go` should then load the docs:

```go
helpSystem := help.NewHelpSystem()
if err := doc.AddDocToHelpSystem(helpSystem); err != nil { ... }
help_cmd.SetupCobraRootCommand(helpSystem, root)
```

This follows the same embedding pattern used by Glazed itself in `pkg/doc/doc.go` and `cmd/glaze/main.go`.

## Validation checklist

- `go test ./... -count=1`
- `go run ./cmd/goja-site help getting-started`
- `go run ./cmd/goja-site help user-guide`
- `go run ./cmd/goja-site help js-api-reference`
- `go run ./cmd/goja-site help developer-guide`
- `go run ./cmd/goja-site help` to confirm discoverability.
- `docmgr doctor --ticket GOJA-SITE-HELP-DOCS --stale-after 30`

## Future help entries

This first set should remain focused. Later tickets can add deeper entries once users ask for them:

- `kanban-dsl-tutorial`: build a board from a plain table.
- `db-guard-guide`: soft cleanup callbacks and hard-limit behavior.
- `multi-site-deployment`: config, DNS, K3s, Argo CD, and PVC runbook.
- `sessions-guide`: practical patterns for private-by-session demo apps.
- `native-module-development`: adding a new Go-backed `require(...)` module.
