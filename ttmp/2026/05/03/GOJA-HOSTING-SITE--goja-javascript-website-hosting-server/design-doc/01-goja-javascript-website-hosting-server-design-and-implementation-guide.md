---
Title: Goja JavaScript Website Hosting Server Design and Implementation Guide
Ticket: GOJA-HOSTING-SITE
Status: active
Topics:
    - go
    - goja
    - javascript
    - web
    - glazed
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../corporate-headquarters/discord-bot/internal/jsdiscord/ui_module.go
      Note: Shows runtime-registered JS UI DSL module pattern
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/engine/factory.go
      Note: Defines factory composition
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/engine/runtime_modules.go
      Note: Defines RuntimeModuleRegistrar for Express and UI DSL modules
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/modules/common.go
      Note: Defines NativeModule and registry contract used by the proposed modules
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/modules/database/database.go
      Note: Provides database module API and preconfiguration options
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/modules/fs/fs.go
      Note: Provides filesystem module API to expose to trusted scripts
    - Path: ../../../../../../../corporate-headquarters/jesus/pkg/engine/handlers.go
      Note: Historical Express-style request and response reference
ExternalSources: []
Summary: Design for a Glazed-powered Go server that hosts JavaScript websites using go-go-goja, SQLite, fs, Express-style routing, and an HTML UI DSL.
LastUpdated: 2026-05-03T13:45:00-04:00
WhatFor: Onboard an intern to implement the Go + goja JavaScript website hosting server.
WhenToUse: Use before implementing the server, modules, DSL, tests, or demo Kanban application.
---


# Goja JavaScript Website Hosting Server Design and Implementation Guide

## Executive summary

We want a small Go server that lets developers build little websites in JavaScript while Go owns the process, command-line UX, database connection, filesystem access, HTTP listener, and safety boundaries. The application should be a new Go CLI/server built with the Glazed command framework. At runtime it should load a directory of JavaScript scripts, create one `go-go-goja` runtime, enable the database and filesystem modules, register a simplified Express.js-style HTTP module, register an HTML UI DSL module, execute the scripts, and serve the routes they register.

The target developer experience is intentionally simple:

```bash
goja-site serve \
  --db ./app.db \
  --scripts ./scripts \
  --addr :8080
```

```javascript
// scripts/app.js
const db = require("database");
const fs = require("fs");
const express = require("express");
const ui = require("ui.dsl");

const app = express.app();

app.get("/", (req, res) => {
  const cards = db.query("select * from cards order by position");
  return res.html(ui.page({ title: "Kanban" },
    ui.div({ class: "board" },
      ["todo", "doing", "done"].map(status =>
        ui.section({ class: "column" },
          ui.h2(status),
          cards.filter(c => c.status === status).map(card =>
            ui.article({ class: "card" },
              ui.h3(card.title),
              ui.p(card.description || "")
            )
          )
        )
      )
    )
  ));
});

app.post("/cards", (req, res) => {
  db.exec("insert into cards(title, status, position) values (?, ?, ?)",
    req.body.title, req.body.status || "todo", Date.now());
  res.redirect("/");
});
```

The key architectural decision is to build on the current `../corporate-headquarters/go-go-goja` factory system instead of copying the older `../corporate-headquarters/jesus` runtime. `jesus` is useful as a historical reference for Express-style request/response shapes and route dispatch, but it predates the evolved `go-go-goja` factory, module middleware, runtime module registrar, and native module patterns.

## Scope and non-goals

### In scope

- A Go CLI built with Glazed conventions.
- A `serve` command that accepts a database path, script directory, bind address, and development options.
- A runtime built with `github.com/go-go-golems/go-go-goja/engine`.
- Explicitly enabled modules:
  - `database`, preconfigured to the CLI-supplied SQLite database.
  - `fs`, for script-controlled local file access during development.
  - a new Express-style HTTP module, probably `require("express")` or `require("http.server")`.
  - a new HTML UI DSL module, `require("ui.dsl")`.
- A route registry that maps Go `net/http` requests to JavaScript handler functions.
- HTML rendering from structured JS values returned by `ui.dsl`.
- A Kanban example that demonstrates schema creation, CRUD endpoints, and server-rendered HTML.
- Tests at Go service level and goja runtime integration level.

### Out of scope for the first implementation

- Full Express compatibility.
- A package manager or npm compatibility layer.
- Browser-side JavaScript bundling.
- Multi-tenant sandboxing.
- Production-grade security for untrusted scripts.
- Hot module replacement. A simple `--reload` or restart-on-change can come later.
- WebSockets. Keep the initial HTTP module request/response based.

## Current-state evidence from related repositories

This section anchors the proposed design to concrete files. The intern should read these files before implementing anything.

### go-go-goja native module contract

`../corporate-headquarters/go-go-goja/modules/common.go` defines the core native module interface. A native module exposes `Name()`, `Doc()`, and `Loader(*goja.Runtime, *goja.Object)`. The registry can register module implementations and enable them into a `goja_nodejs/require.Registry`.

Important lines:

- `modules.NativeModule` requires `Name`, `Doc`, and `Loader` (`modules/common.go:28-32`).
- `Registry.Register` appends a module (`modules/common.go:48-52`).
- `Registry.Enable` calls `gojaRegistry.RegisterNativeModule(m.Name(), m.Loader)` (`modules/common.go:85-89`).
- `modules.Register` adds a module to the default global registry (`modules/common.go:92-98`).

This means simple global modules can be implemented as `modules.NativeModule` packages with an `init()` registration. Runtime-scoped modules that need access to a specific server instance should use `engine.RuntimeModuleRegistrar` instead.

### go-go-goja runtime factory and module selection

`../corporate-headquarters/go-go-goja/engine/factory.go` is the current runtime assembly API. It supports static modules, module middleware, runtime-scoped module registrars, and runtime initializers.

Important lines:

- `WithModules` appends static module registrations (`engine/factory.go:67-72`).
- `UseModuleMiddleware` is the preferred way to control default-registry modules (`engine/factory.go:74-87`).
- `WithRuntimeModuleRegistrars` registers modules that need per-runtime objects (`engine/factory.go:89-94`).
- During `Build`, middleware is evaluated and default registry module names are converted to specs (`engine/factory.go:134-147`).
- `NewRuntime` creates the VM, event loop, runtime owner, require registry, and runtime values map (`engine/factory.go:187-224`).
- Runtime module registrars receive `RuntimeModuleContext` before `require` is enabled (`engine/factory.go:235-248`).

`../corporate-headquarters/go-go-goja/engine/runtime_modules.go` defines that runtime registrar contract:

- `RuntimeModuleRegistrar` has `ID()` and `RegisterRuntimeModules(ctx, reg)` (`engine/runtime_modules.go:12-17`).
- The context exposes `Context`, `VM`, `Loop`, `Owner`, `AddCloser`, and `Values` (`engine/runtime_modules.go:19-27`).

Our Express and UI DSL modules should use runtime registrars when they need shared host state. The Express module definitely needs a host-side route registry. The UI DSL module can be a plain native module if it is pure, but registering it alongside Express as a runtime registrar keeps the host composition explicit.

### Module middleware and aliases

`../corporate-headquarters/go-go-goja/engine/module_middleware.go` lets the host choose exactly which default modules to expose.

Important lines:

- `MiddlewareOnly(names...)` replaces the module selection with the named modules (`engine/module_middleware.go:32-48`).
- `MiddlewareAdd(names...)` appends modules that exist in the default registry (`engine/module_middleware.go:68-86`).
- `allRegisteredModuleNames` reads from `modules.ListDefaultModules()` (`engine/module_middleware.go:110-119`).

For this application, prefer an explicit module selection:

```go
factory, err := engine.NewBuilder().
  UseModuleMiddleware(engine.MiddlewareOnly("database", "db", "fs", "path", "time", "timer")).
  WithRuntimeModuleRegistrars(
    web.NewExpressRegistrar(host),
    uidsl.NewRegistrar(),
  ).
  Build()
```

Note that database aliases matter. Existing tests mention `database` and `db` as aliases, so implementation should verify which names are actually registered in the current `go-go-goja` version before relying on both.

### Database module

`../corporate-headquarters/go-go-goja/modules/database/database.go` already provides the SQL module we need.

Important lines:

- The module supports options such as `WithName`, `WithPreconfiguredDB`, `WithCloseFn`, and `WithConfigureEnabled` (`modules/database/database.go:20-57`).
- `DBModule` defaults to name `database` and `allowConfigure: true` (`modules/database/database.go:70-81`).
- The TypeScript declaration advertises `configure`, `query`, `exec`, and `close` (`modules/database/database.go:91-124`).
- `Loader` exports `configure`, `query`, `exec`, and `close` (`modules/database/database.go:153-160`).
- `Configure` opens a SQL connection and stores it in the module (`modules/database/database.go:162-180`).
- `Query` returns `[]map[string]any` (`modules/database/database.go:195-240`).

For this server, Go should open the SQLite database and expose a preconfigured database module. Avoid requiring user scripts to call `configure()`. That keeps the CLI database path authoritative and prevents scripts from silently switching database files.

Recommended host behavior:

- Open `*sql.DB` from `--db`.
- Run migrations for server-owned metadata only if needed.
- Register `database.New(database.WithPreconfiguredDB(db), database.WithCloseFn(db.Close), database.WithConfigureEnabled(false))` if the current package API permits static registration with a custom module instance.
- If `DefaultRegistryModule("database")` always uses the global default module, add a custom `engine.ModuleSpec` for the preconfigured instance instead of using the default registry database module.

### Filesystem module

`../corporate-headquarters/go-go-goja/modules/fs/fs.go` provides asynchronous and synchronous filesystem helpers.

Important lines:

- The TypeScript declaration lists async helpers such as `readFile`, `writeFile`, `exists`, `mkdir`, `readdir`, `stat`, and sync helpers such as `readFileSync`, `writeFileSync`, `existsSync`, `mkdirSync`, `readdirSync`, `statSync` (`modules/fs/fs.go:27-64`).
- The docs call out promise-based and synchronous helpers (`modules/fs/fs.go:67-80`).
- The loader requires `runtimebridge` owner bindings for async methods (`modules/fs/fs.go:82-87`).
- The loader exports async and sync functions using `modules.SetExport` (`modules/fs/fs.go:89-184`).

Because `fs` currently does not appear path-scoped in the inspected file, treat it as a trusted-script development module. Do not advertise this server as safe for untrusted scripts until a path-scoped filesystem module exists.

### Discord bot UI DSL pattern

`../corporate-headquarters/discord-bot/internal/jsdiscord/ui_module.go` shows a successful JavaScript UI DSL exposed as `require("ui")`.

Important lines:

- `UIRegistrar` implements `engine.RuntimeModuleRegistrar` and registers `ui` with `reg.RegisterNativeModule("ui", uiLoader)` (`internal/jsdiscord/ui_module.go:12-24`).
- `uiLoader` exports builders such as `message`, `embed`, `button`, `select`, `form`, and helpers such as `row`, `pager`, `actions`, `confirm`, `card`, `ok`, and `error` (`internal/jsdiscord/ui_module.go:26-128`).
- The embed builder uses an ES6 Proxy to expose chainable methods and validates unknown methods (`internal/jsdiscord/ui_embed.go:21-98`).
- The message builder accumulates state and `.build()` returns a normalized Go value (`internal/jsdiscord/ui_message.go:11-108`).
- `extractEmbed` and row component extraction show how builders can accept either builder proxies or already-built Go values (`internal/jsdiscord/ui_message.go:110-196`).

For HTML, we should reuse the same shape: small constructors, chainable builders for complex components, `.build()` for normalized output, and clear type errors when a child value is invalid. The difference is the output target: Discord payload structs become an HTML AST (`Node`) rendered to safe HTML.

### Jesus as historical Express reference

`../corporate-headquarters/jesus` contains an older server that dynamically registers routes from JavaScript. It should be used as a reference, not copied wholesale.

Important lines:

- `Engine` stores handlers as `map[string]map[string]*HandlerInfo` and file handlers separately (`pkg/engine/engine.go:18-30`).
- `HandlerInfo` stores the JavaScript callable plus content type and options (`pkg/engine/engine.go:32-37`).
- `EvalJob` carries the handler, code, response writer, request, result channel, session ID, and source (`pkg/engine/engine.go:39-49`).
- `NewEngine` creates a Goja runtime, enables module registry, configures database, sets JS field name mapping, starts an event loop, sets bindings, and binds `const db = require('database')` (`pkg/engine/engine.go:58-115`).
- The default bootstrap script registers `app.get("/")`, `app.get("/health")`, and `app.post("/counter")` (`pkg/engine/engine.go:129-180`).
- `ExpressRequest` contains method, URL, path, query, headers, body, cookies, IP, protocol, hostname, and params (`pkg/engine/handlers.go:23-36`).
- `ExpressResponse` has status, headers, cookies, writer, engine, and sent state (`pkg/engine/handlers.go:38-46`).
- `res.send` auto-detects HTML, JSON, and text for string responses (`pkg/engine/handlers.go:70-134`).
- `res.json` sets `Content-Type: application/json` and JSON-encodes data (`pkg/engine/handlers.go:136-166`).
- `HandleDynamicRoute` finds a registered JS handler and submits an `EvalJob` (`pkg/web/router.go:9-49`).
- The Jesus JavaScript API reference documents the intended Express-like developer experience (`pkg/doc/docs/javascript-api-reference.md:56-96`).

The new implementation should keep the good parts: familiar route APIs, request/response objects, body parsing, redirects, and content-type handling. It should replace the older runtime construction with the current `go-go-goja/engine.Factory` and keep the HTTP module isolated as a native module/registrar.

## Proposed architecture

### High-level component diagram

```text
+-------------------+        +------------------------+
| goja-site CLI     |        | scripts directory      |
| Glazed commands   |        | app.js, routes/*.js    |
+---------+---------+        +-----------+------------+
          |                              |
          | serve --db --scripts --addr |
          v                              v
+-----------------------------------------------------+
| Go Host Process                                     |
|                                                     |
|  +-----------------+     +----------------------+   |
|  | SQLite *sql.DB  |---->| database JS module   |   |
|  +-----------------+     +----------------------+   |
|                                                     |
|  +-----------------+     +----------------------+   |
|  | net/http server |<----| Express route module |   |
|  +--------+--------+     +----------+-----------+   |
|           |                         |               |
|           v                         v               |
|  +-----------------+     +----------------------+   |
|  | Route registry  |---->| Goja runtime owner   |   |
|  +-----------------+     +----------+-----------+   |
|                                      |               |
|  +-----------------+     +----------v-----------+   |
|  | HTML renderer   |<----| ui.dsl JS module     |   |
|  +-----------------+     +----------------------+   |
+-----------------------------------------------------+
```

### Package layout

Recommended repository layout for the new server:

```text
.
├── cmd/goja-site/main.go
├── cmd/goja-site/cmd/root.go
├── cmd/goja-site/cmd/serve.go
├── pkg/app/config.go
├── pkg/app/server.go
├── pkg/app/script_loader.go
├── pkg/web/host.go
├── pkg/web/route_registry.go
├── pkg/web/express_module.go
├── pkg/web/request_response.go
├── pkg/web/body.go
├── pkg/web/static.go
├── pkg/uidsl/node.go
├── pkg/uidsl/render.go
├── pkg/uidsl/module.go
├── pkg/uidsl/builders.go
├── examples/kanban/scripts/app.js
├── examples/kanban/README.md
└── internal/testutil/runtime.go
```

### Responsibilities by package

#### `cmd/goja-site/cmd`

Owns command-line UX using Glazed.

- Root command:
  - sets application name and logging.
  - wires Glazed help if local docs are embedded.
- `serve` command:
  - parses flags.
  - calls `app.RunServe(ctx, settings)`.
  - emits startup rows or events for Glazed output.

#### `pkg/app`

Owns process-level orchestration.

- Opens SQLite database.
- Builds the go-go-goja factory.
- Creates the runtime.
- Loads scripts from directory in deterministic order.
- Starts and stops the HTTP server.
- Handles signals and graceful shutdown.

#### `pkg/web`

Owns HTTP hosting and Express-style module.

- Route registry.
- Request parsing.
- Response object semantics.
- Native module export for JavaScript route registration.
- `http.Handler` that dispatches Go requests into JS handlers.

#### `pkg/uidsl`

Owns HTML UI representation and rendering.

- Normalized HTML node model.
- Escaping and attribute rendering.
- JS module exports for element constructors.
- Optional chainable builder helpers for pages, forms, tables, and cards.

## Glazed CLI design

### Root command

Follow the Glazed command authoring pattern:

- Define command structs embedding `*cmds.CommandDescription`.
- Define settings structs with `glazed` tags.
- Use `fields.New(...)` for flags.
- Decode with `vals.DecodeSectionInto(schema.DefaultSlug, settings)`.
- Add logging to the root command.
- Build Cobra commands through Glazed CLI helpers.

### `serve` command settings

```go
type ServeSettings struct {
  Addr       string `glazed:"addr"`
  DBPath     string `glazed:"db"`
  ScriptsDir string `glazed:"scripts"`
  StaticDir  string `glazed:"static"`
  Dev        bool   `glazed:"dev"`
  Reload     bool   `glazed:"reload"`
}
```

Suggested flags:

- `--addr`, default `:8080`.
- `--db`, required or default `./app.db`.
- `--scripts`, required or default `./scripts`.
- `--static`, optional static asset directory.
- `--dev`, enables verbose logs and friendlier error pages.
- `--reload`, future option; can initially return a clear not-implemented error.

### Serve command pseudocode

```go
func (c *ServeCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
  s := &ServeSettings{}
  if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
    return err
  }

  srv, err := app.NewServer(app.Config{
    Addr: s.Addr,
    DBPath: s.DBPath,
    ScriptsDir: s.ScriptsDir,
    StaticDir: s.StaticDir,
    Dev: s.Dev,
  })
  if err != nil { return err }

  gp.AddRow(ctx, types.NewRow(
    types.MRP("addr", s.Addr),
    types.MRP("db", s.DBPath),
    types.MRP("scripts", s.ScriptsDir),
  ))

  return srv.Run(ctx)
}
```

## Runtime assembly

### Preferred assembly flow

```text
serve command
  -> app.NewServer(config)
    -> open SQLite DB
    -> create web.Host with route registry
    -> build go-go-goja factory
      -> enable database + fs + small support modules
      -> register express module with host pointer
      -> register ui.dsl module
    -> create runtime
    -> execute scripts in deterministic order
    -> create net/http server using host as handler
```

### Pseudocode

```go
func NewServer(cfg Config) (*Server, error) {
  db, err := sql.Open("sqlite3", cfg.DBPath)
  if err != nil { return nil, err }

  host := web.NewHost(web.HostOptions{Dev: cfg.Dev})

  dbModule := databasemod.New(
    databasemod.WithPreconfiguredDB(db),
    databasemod.WithCloseFn(db.Close),
    databasemod.WithConfigureEnabled(false),
  )

  factory, err := engine.NewBuilder().
    WithModules(engine.NativeModuleSpec{
      ModuleID: "database:app",
      ModuleName: dbModule.Name(),
      Loader: dbModule.Loader,
    }).
    UseModuleMiddleware(engine.MiddlewareOnly("fs", "path", "time", "timer")).
    WithRuntimeModuleRegistrars(
      web.NewExpressRegistrar(host),
      uidsl.NewRegistrar(),
    ).
    Build()
  if err != nil { return nil, err }

  rt, err := factory.NewRuntime(context.Background())
  if err != nil { return nil, err }

  server := &Server{cfg: cfg, db: db, host: host, runtime: rt}
  if err := server.LoadScripts(); err != nil { server.Close(); return nil, err }
  return server, nil
}
```

Important detail: the exact module composition depends on whether `WithModules(...)` plus `UseModuleMiddleware(...)` appends both explicit and selected default modules. The inspected `Build` method does append explicit modules first, then middleware-selected defaults when middleware exists (`engine/factory.go:126-147`). Therefore the above pattern should produce the custom database module plus default `fs`, `path`, `time`, and `timer` modules.

## Express-style HTTP module

### Design goal

The HTTP module should feel familiar to an Express user but be small enough to implement correctly.

Required first-version API:

```javascript
const express = require("express");
const app = express.app();

app.get(path, handler);
app.post(path, handler);
app.put(path, handler);
app.patch(path, handler);
app.delete(path, handler);
app.all(path, handler);

app.static(prefix, directory);       // optional phase 2
app.use(middleware);                 // optional phase 2
```

A route handler can respond in two styles:

```javascript
app.get("/json", (req, res) => {
  res.status(200).json({ ok: true });
});

app.get("/return", (req, res) => {
  return ui.div("Hello");
});
```

The host should support both styles because returned `ui.dsl` nodes make small websites pleasant, while explicit `res` methods are familiar for JSON, redirects, and status codes.

### Request object contract

```typescript
type Request = {
  method: string;
  url: string;
  path: string;
  query: Record<string, string | string[]>;
  params: Record<string, string>;
  headers: Record<string, string>;
  cookies: Record<string, string>;
  ip: string;
  body: any;
  rawBody: string | Uint8Array;
};
```

Implement this in Go by parsing:

- path and raw query from `r.URL`.
- headers from `r.Header`.
- cookies from `r.Cookies()`.
- JSON body when `Content-Type` contains `application/json`.
- form body when `application/x-www-form-urlencoded` or `multipart/form-data`.
- raw string/body fallback otherwise.

### Response object contract

```typescript
type Response = {
  status(code: number): Response;
  set(name: string, value: string): Response;
  type(contentType: string): Response;
  cookie(name: string, value: string, options?: CookieOptions): Response;
  send(value: any): void;
  html(nodeOrString: any): void;
  json(value: any): void;
  redirect(statusOrURL: number | string, maybeURL?: string): void;
  end(): void;
};
```

Response semantics:

- Status defaults to `200`.
- `res.status(code)` is chainable.
- `res.set(name, value)` stores headers until send.
- `res.json(value)` sets JSON content type and encodes JSON.
- `res.html(node)` renders a UI DSL node and sets HTML content type.
- `res.send(string)` should auto-detect HTML only if no content type was set, following the Jesus behavior (`pkg/engine/handlers.go:97-133`).
- Any second send should be a no-op or a clear error in development mode. Pick one and test it.

### Route registry

Use an ordered registry. Avoid the exact `map[path][method]` shape from Jesus if we need path parameters like `/cards/:id`; maps do not preserve route specificity or registration order.

```go
type Route struct {
  Method  string
  Pattern string
  Handler goja.Callable
  Options RouteOptions
}

type Registry struct {
  mu     sync.RWMutex
  routes []Route
}
```

Matching algorithm for v1:

1. Compare method (`GET`, `POST`, etc.), with `ALL` as wildcard.
2. Split pattern and request path by `/`.
3. Segment match:
   - exact segment matches exact segment.
   - `:name` matches one segment and stores a param.
   - `*` or `/*` can be phase 2.
4. Return the first registered match.

Pseudocode:

```go
func (r *Registry) Match(method, path string) (*Route, map[string]string, bool) {
  r.mu.RLock(); defer r.mu.RUnlock()
  for _, route := range r.routes {
    if route.Method != method && route.Method != "ALL" { continue }
    params, ok := matchPattern(route.Pattern, path)
    if ok { return &route, params, true }
  }
  return nil, nil, false
}
```

### Dispatch from Go to JavaScript

Goja runtimes are not safe for arbitrary concurrent access. The existing `go-go-goja` runtime owns a `runtimeowner.Runner` (`engine/factory.go:201-204`) and exposes it on `Runtime.Owner` (`engine/runtime.go:31-37`). The HTTP handler should call JavaScript through the owner rather than directly invoking callables from net/http goroutines.

Pseudocode:

```go
func (h *Host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  route, params, ok := h.routes.Match(r.Method, r.URL.Path)
  if !ok { http.NotFound(w, r); return }

  reqDTO, err := BuildRequest(r, params)
  if err != nil { http.Error(w, err.Error(), 400); return }

  response := NewResponse(w, h.renderer, h.dev)

  _, err = h.runtime.Owner.Call(r.Context(), "http-handler", func(ctx context.Context, vm *goja.Runtime) (any, error) {
    reqVal := vm.ToValue(reqDTO)
    resVal := vm.ToValue(response.JSAdapter(vm))
    result, err := route.Handler(goja.Undefined(), reqVal, resVal)
    if err != nil { return nil, err }
    if !response.Sent() && !goja.IsUndefined(result) && !goja.IsNull(result) {
      return nil, response.SendGojaValue(vm, result)
    }
    return nil, nil
  })
  if err != nil { h.writeError(w, err); return }
}
```

### Module registrar

```go
type ExpressRegistrar struct { host *Host }

func (r *ExpressRegistrar) ID() string { return "express-http" }

func (r *ExpressRegistrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
  r.host.SetRuntime(ctx.Owner, ctx.VM)
  reg.RegisterNativeModule("express", func(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)
    _ = exports.Set("app", func() goja.Value { return r.newAppObject(vm) })
  })
  return nil
}
```

The JS `app` object should be tiny:

```go
func (r *ExpressRegistrar) newAppObject(vm *goja.Runtime) goja.Value {
  obj := vm.NewObject()
  for _, method := range []string{"get", "post", "put", "patch", "delete", "all"} {
    m := method
    _ = obj.Set(m, func(pattern string, handler goja.Value) error {
      fn, ok := goja.AssertFunction(handler)
      if !ok { return fmt.Errorf("app.%s requires a function", m) }
      r.host.Register(strings.ToUpper(m), pattern, fn)
      return nil
    })
  }
  return obj
}
```

## UI DSL module

### Design goal

The UI DSL should produce safe HTML from JavaScript without string concatenation. It should support simple constructor calls and optional builders for higher-level components.

### Core data model

Use a small Go AST:

```go
type Node interface { node() }

type Element struct {
  Tag      string
  Attrs    map[string]AttrValue
  Children []Node
}

type Text struct { Value string }
type RawHTML struct { Value string } // restricted escape hatch
type Fragment struct { Children []Node }
type Doctype struct { Name string }
```

Rules:

- Text nodes are HTML-escaped.
- Attribute values are HTML-escaped.
- Boolean attributes render only the name when true.
- `style` accepts either string or map.
- `class` accepts string, array, or object map of truthy class names.
- `RawHTML` must be explicit: `ui.raw("<span>trusted</span>")`.

### JavaScript API

Keep the basic API small and predictable:

```javascript
const ui = require("ui.dsl");

ui.html(attrs?, ...children)
ui.head(...children)
ui.body(attrs?, ...children)
ui.title(text)
ui.meta(attrs)
ui.link(attrs)
ui.script(attrs?, ...children)
ui.style(text)
ui.div(attrs?, ...children)
ui.span(attrs?, ...children)
ui.h1(...children)
ui.h2(...children)
ui.p(...children)
ui.a(attrs?, ...children)
ui.form(attrs?, ...children)
ui.input(attrs)
ui.button(attrs?, ...children)
ui.ul(...children)
ui.li(...children)
ui.table(...children)
ui.thead(...children)
ui.tbody(...children)
ui.tr(...children)
ui.th(...children)
ui.td(...children)
ui.section(attrs?, ...children)
ui.article(attrs?, ...children)
ui.fragment(...children)
ui.text(value)
ui.raw(trustedHTML)
ui.render(node) // returns string
ui.page(options, ...children) // full HTML document
```

Flexible argument rules:

```javascript
ui.div("text")
ui.div({ class: "card" }, "text")
ui.div({ class: ["card", active && "active"] }, [child1, child2])
```

### Rendering example

```javascript
ui.render(ui.div({ class: "card", hidden: false },
  ui.h3("Fix bug"),
  ui.p("Escape <this> text")
));
```

Output:

```html
<div class="card"><h3>Fix bug</h3><p>Escape &lt;this&gt; text</p></div>
```

### Builder API for common components

Borrow the chainable-builder idea from Discord bot's `ui.embed()` and `ui.message()`.

```javascript
ui.card("Title")
  .class("kanban-card")
  .body(ui.p("Description"))
  .footer(ui.button({ type: "submit" }, "Save"))
  .build()
```

But do not make builders the only way to create HTML. Constructor functions are easier to teach and test.

### UI DSL module implementation pseudocode

```go
type Registrar struct{}
func (r *Registrar) ID() string { return "ui-dsl" }
func (r *Registrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
  reg.RegisterNativeModule("ui.dsl", Loader)
  return nil
}

func Loader(vm *goja.Runtime, moduleObj *goja.Object) {
  exports := moduleObj.Get("exports").(*goja.Object)
  for _, tag := range allowedTags {
    t := tag
    _ = exports.Set(t, func(call goja.FunctionCall) goja.Value {
      node := BuildElementFromCall(vm, t, call)
      return vm.ToValue(node)
    })
  }
  _ = exports.Set("text", func(v any) *Text { return &Text{Value: fmt.Sprint(v)} })
  _ = exports.Set("raw", func(s string) *RawHTML { return &RawHTML{Value: s} })
  _ = exports.Set("fragment", func(call goja.FunctionCall) goja.Value { ... })
  _ = exports.Set("render", func(v goja.Value) (string, error) { return RenderGojaValue(vm, v) })
  _ = exports.Set("page", func(call goja.FunctionCall) goja.Value { ... })
}
```

### Value normalization

The most important implementation detail is normalization. JavaScript handlers might return:

- a Go `*Element` exported from the module.
- an array of nodes.
- a string.
- a number or boolean.
- `null` or `undefined`.

Normalize everything before rendering:

```go
func Normalize(vm *goja.Runtime, v goja.Value) (Node, error) {
  if goja.IsUndefined(v) || goja.IsNull(v) { return Fragment{}, nil }
  switch x := v.Export().(type) {
  case Node: return x, nil
  case string: return Text{Value: x}, nil
  case []any: return normalizeArray(vm, x)
  case []Node: return Fragment{Children: x}, nil
  default: return Text{Value: fmt.Sprint(x)}, nil
  }
}
```

## Kanban example design

### Schema

```sql
CREATE TABLE IF NOT EXISTS cards (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  description TEXT,
  status TEXT NOT NULL DEFAULT 'todo',
  position INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### JavaScript application

```javascript
const db = require("database");
const express = require("express");
const ui = require("ui.dsl");
const app = express.app();

function migrate() {
  db.exec(`CREATE TABLE IF NOT EXISTS cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    position INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
  )`);
}

function cardView(card) {
  return ui.article({ class: "card" },
    ui.h3(card.title),
    card.description ? ui.p(card.description) : null,
    ui.form({ method: "post", action: `/cards/${card.id}/move` },
      ui.select({ name: "status" },
        ["todo", "doing", "done"].map(s =>
          ui.option({ value: s, selected: s === card.status }, s)
        )
      ),
      ui.button({ type: "submit" }, "Move")
    )
  );
}

function boardView() {
  const cards = db.query("select * from cards order by position, id");
  return ui.page({ title: "Kanban" },
    ui.link({ rel: "stylesheet", href: "/static/style.css" }),
    ui.main({ class: "board" },
      ["todo", "doing", "done"].map(status =>
        ui.section({ class: "column" },
          ui.h2(status),
          cards.filter(c => c.status === status).map(cardView)
        )
      )
    ),
    ui.form({ method: "post", action: "/cards" },
      ui.input({ name: "title", placeholder: "New card", required: true }),
      ui.button({ type: "submit" }, "Add")
    )
  );
}

migrate();

app.get("/", (req, res) => res.html(boardView()));

app.post("/cards", (req, res) => {
  db.exec("insert into cards(title, status, position) values (?, 'todo', ?)",
    req.body.title, Date.now());
  res.redirect("/");
});

app.post("/cards/:id/move", (req, res) => {
  db.exec("update cards set status = ?, updated_at = CURRENT_TIMESTAMP where id = ?",
    req.body.status, req.params.id);
  res.redirect("/");
});
```

## Implementation phases

### Phase 1: Project skeleton and Glazed CLI

Files:

- `cmd/goja-site/main.go`
- `cmd/goja-site/cmd/root.go`
- `cmd/goja-site/cmd/serve.go`
- `pkg/app/config.go`

Tasks:

1. Create module and `go.mod`.
2. Add dependencies: `go-go-goja`, `glazed`, `cobra`, `goja`, `goja_nodejs`, `mattn/go-sqlite3`, `zerolog`.
3. Implement root command with Glazed logging and help conventions.
4. Implement `serve` command settings and validation.
5. Add a smoke test that `goja-site serve --help` exits successfully.

Acceptance criteria:

- `go test ./...` passes.
- `go run ./cmd/goja-site serve --help` shows flags.

### Phase 2: Runtime host and script loader

Files:

- `pkg/app/server.go`
- `pkg/app/script_loader.go`
- `internal/testutil/runtime.go`

Tasks:

1. Open SQLite DB from config.
2. Build go-go-goja runtime factory.
3. Execute every `*.js` file under `--scripts` in deterministic lexical order.
4. Support `index.js` first if present, then the rest lexically, or document pure lexical behavior.
5. Close runtime, DB, and HTTP server cleanly.

Acceptance criteria:

- A test script can `require("database")` and `require("fs")`.
- A script execution error includes file path and line information where possible.

### Phase 3: Express-style route module

Files:

- `pkg/web/host.go`
- `pkg/web/route_registry.go`
- `pkg/web/express_module.go`
- `pkg/web/request_response.go`
- `pkg/web/body.go`

Tasks:

1. Implement route registration from JS.
2. Implement path matching with `:params`.
3. Implement request DTO parsing.
4. Implement response methods.
5. Dispatch via runtime owner, not direct concurrent VM access.
6. Add HTTP tests with `httptest`.

Acceptance criteria:

- JS can register `GET /hello` and return JSON.
- JS can register `GET /users/:id` and read `req.params.id`.
- JS can register `POST /echo` and read JSON/form bodies.
- Double send behavior is tested.

### Phase 4: UI DSL module and renderer

Files:

- `pkg/uidsl/node.go`
- `pkg/uidsl/render.go`
- `pkg/uidsl/module.go`
- `pkg/uidsl/builders.go`

Tasks:

1. Implement node model.
2. Implement safe renderer.
3. Implement tag constructors.
4. Implement `ui.page`, `ui.fragment`, `ui.text`, `ui.raw`, and `ui.render`.
5. Integrate `res.html` and return-value rendering.
6. Add escaping tests.

Acceptance criteria:

- `ui.div("<x>")` renders escaped text.
- `ui.raw("<x>")` renders raw HTML.
- Boolean attributes behave correctly.
- Arrays and nested arrays flatten into fragments.
- Returning a node from a handler sends HTML.

### Phase 5: Kanban example

Files:

- `examples/kanban/scripts/app.js`
- `examples/kanban/scripts/static.js` or static directory
- `examples/kanban/README.md`

Tasks:

1. Create schema migration in JS.
2. Implement board page.
3. Implement create card endpoint.
4. Implement move card endpoint.
5. Add simple CSS.
6. Add README with run command.

Acceptance criteria:

- `goja-site serve --db examples/kanban/kanban.db --scripts examples/kanban/scripts` starts.
- Browser can create and move cards.
- Data persists in SQLite.

### Phase 6: Documentation and API reference

Files:

- `docs/javascript-api.md`
- `docs/ui-dsl.md`
- `docs/server-operations.md`

Tasks:

1. Document Express subset.
2. Document UI DSL constructors.
3. Document database and filesystem availability.
4. Document security model: trusted scripts only.
5. Document examples and troubleshooting.

## Testing strategy

### Unit tests

- `pkg/web/route_registry_test.go`: path matching, method matching, params, ordering.
- `pkg/web/body_test.go`: JSON, form, empty body, invalid JSON.
- `pkg/web/request_response_test.go`: status, headers, JSON, HTML, redirects, double sends.
- `pkg/uidsl/render_test.go`: escaping, attrs, fragments, raw HTML, page rendering.

### Runtime integration tests

Use `engine.NewBuilder()` with the same registrars as production. Run JS strings that require modules and register routes.

```go
func TestJSRegistersHTMLRoute(t *testing.T) {
  host := web.NewHost(web.HostOptions{Dev: true})
  rt := newTestRuntime(t, host)

  _, err := rt.Owner.Call(context.Background(), "load", func(ctx context.Context, vm *goja.Runtime) (any, error) {
    _, err := vm.RunString(`
      const express = require("express");
      const ui = require("ui.dsl");
      const app = express.app();
      app.get("/", () => ui.h1("Hello"));
    `)
    return nil, err
  })
  require.NoError(t, err)

  rr := httptest.NewRecorder()
  host.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
  require.Equal(t, 200, rr.Code)
  require.Contains(t, rr.Body.String(), "<h1>Hello</h1>")
}
```

### End-to-end tests

- Start the server on `127.0.0.1:0`.
- Load Kanban scripts.
- Use `http.Client` to create a card.
- Fetch `/` and assert the card appears.
- Query SQLite directly to assert persistence.

### Validation commands

```bash
go test ./... -count=1
go test ./pkg/web ./pkg/uidsl -run Test -count=1
go run ./cmd/goja-site serve --db /tmp/goja-site.db --scripts examples/kanban/scripts --addr :8080
curl -i http://localhost:8080/
```

## Risks and mitigations

### Risk: unsafe filesystem/database access

The requested server enables `database` and `fs`. This is powerful and unsafe for untrusted scripts. Mitigation: document this as a trusted-script tool. Add path-scoped filesystem and database permissions in a future phase if needed.

### Risk: Goja concurrency bugs

Net/http handlers run concurrently; a single Goja runtime must be accessed serially. Mitigation: dispatch all JavaScript handler calls through `Runtime.Owner.Call` and add concurrent request tests.

### Risk: response object crosses VM boundaries incorrectly

If Go response adapters are exported to JS and used after the Go request finishes, scripts could attempt late writes. Mitigation: response adapter tracks `sent` and `closed`, and all async behavior is initially disallowed or awaited explicitly.

### Risk: UI DSL accidentally permits XSS

HTML builders must escape text and attributes by default. Mitigation: `raw()` is explicit, tested, and documented as trusted only.

### Risk: older Jesus patterns conflict with current go-go-goja patterns

Jesus manually creates Goja runtime and uses older bindings. Mitigation: use Jesus only for request/response semantics; use current `go-go-goja/engine` factory for runtime assembly.

## Open questions

1. Should the module be called `express`, `http`, or `http.server`? Recommendation: `express` for familiarity, but document that it is an Express subset.
2. Should `database` also be aliased to `db`? Recommendation: yes if no duplicate ID conflict occurs, because older examples and tests mention both names.
3. Should `fs` be scoped to the script directory? Recommendation: not in v1 unless quickly available; document trusted-only behavior.
4. Should scripts be loaded as CommonJS modules or executed as top-level files? Recommendation: execute entry scripts directly with `vm.RunScript` and allow `require()` for shared files through module roots.
5. Should route handlers support promises? Recommendation: yes eventually. For v1, test whether go-go-goja owner/eventloop can await promises safely; if not, keep handlers synchronous and document it.

## References

- `../corporate-headquarters/go-go-goja/modules/common.go`: native module interface and registry.
- `../corporate-headquarters/go-go-goja/engine/factory.go`: runtime factory, module middleware, runtime registrar execution.
- `../corporate-headquarters/go-go-goja/engine/runtime_modules.go`: runtime-scoped module registrar contract.
- `../corporate-headquarters/go-go-goja/engine/module_middleware.go`: `MiddlewareOnly`, `MiddlewareAdd`, and module selection.
- `../corporate-headquarters/go-go-goja/modules/database/database.go`: database module options and query/exec API.
- `../corporate-headquarters/go-go-goja/modules/fs/fs.go`: filesystem module API.
- `../corporate-headquarters/discord-bot/internal/jsdiscord/ui_module.go`: runtime module registrar pattern for a JS UI DSL.
- `../corporate-headquarters/discord-bot/internal/jsdiscord/ui_embed.go`: chainable builder via ES6 Proxy.
- `../corporate-headquarters/discord-bot/internal/jsdiscord/ui_message.go`: builder normalization and child extraction patterns.
- `../corporate-headquarters/jesus/pkg/engine/engine.go`: historical dynamic JavaScript server runtime.
- `../corporate-headquarters/jesus/pkg/engine/handlers.go`: historical Express request/response model.
- `../corporate-headquarters/jesus/pkg/web/router.go`: historical dynamic route dispatch.
- `../corporate-headquarters/jesus/pkg/doc/docs/javascript-api-reference.md`: historical JS API reference for Express-style server scripting.
