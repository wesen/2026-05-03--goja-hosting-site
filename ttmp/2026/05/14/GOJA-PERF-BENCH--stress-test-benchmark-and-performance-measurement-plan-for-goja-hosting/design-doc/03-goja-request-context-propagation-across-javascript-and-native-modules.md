---
Title: Goja Request Context Propagation Across JavaScript and Native Modules
Ticket: GOJA-PERF-BENCH
Status: active
Topics:
    - goja
    - performance
    - benchmarking
    - stress-testing
    - observability
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/modules/database/database.go
      Note: Current non-context-aware database module boundary
    - Path: ../../../../../../../../../go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/gojahttp/host.go
      Note: HTTP request context enters JavaScript handler through Owner.Call
    - Path: ../../../../../../../../../go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimebridge/runtimebridge.go
      Note: Existing per-VM runtime bindings to extend with current-call context
    - Path: ../../../../../../../../../go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimeowner/runner.go
      Note: Owner Call/Post scheduling and owner context marking
    - Path: go.mod
      Note: temporary local replace for development go-go-goja checkout
    - Path: pkg/app/database.go
      Note: simpleDB now implements context-aware database execution
    - Path: pkg/app/observability_test.go
      Note: end-to-end trace parentage regression test
    - Path: pkg/dbguard/metered.go
      Note: guarded DB wrapper now forwards request context into sql.DB
    - Path: pkg/observability/sql.go
      Note: goja-site DB spans now start from propagated request context
ExternalSources: []
Summary: Analysis and implementation guide for propagating request context through go-go-goja owner calls, JavaScript handlers, native modules, database calls, and OpenTelemetry spans.
LastUpdated: 2026-05-14T16:45:00-04:00
WhatFor: Use this before changing go-go-goja or goja-site to make DB, guard, Kanban, and native-module spans children of the originating HTTP request.
WhenToUse: When implementing Phase 6 tracing context propagation, context-aware database operations, or native module access to current request context.
---










# Goja Request Context Propagation Across JavaScript and Native Modules

## Executive Summary

`go-go-goja` already has an important context model: `runtimeowner.Runner.Call(ctx, op, fn)` accepts a `context.Context`, schedules `fn` onto the Goja owner/event-loop goroutine, and invokes `fn(ctx, vm)` with an owner-marked context. `gojahttp.Host` already passes `r.Context()` into this call when it invokes JavaScript HTTP handlers. That means the originating HTTP request context already reaches the Go closure that calls the JavaScript handler.

However, native modules invoked from JavaScript do not currently have a standard way to retrieve that current call context. The database module exposes JavaScript functions as methods like `Query(query string, args ...any)` and calls a `QueryExecer` interface that has no `context.Context` parameter. Therefore, `goja-site` can create DB spans today, but it must use `context.Background()`, so those DB spans are not children of the HTTP request span.

The right fix is to update `go-go-goja` itself with a runtime-scoped current-call context mechanism and context-aware native module interfaces. The core idea is:

1. The runtime owner marks the active context while executing a Goja owner call.
2. Native modules can safely read the current context for the current owner-goroutine call.
3. The database module grows `QueryContext`/`ExecContext` support while preserving the existing `Query`/`Exec` interface.
4. `gojahttp.Host` continues passing `r.Context()` into `Owner.Call`, so JavaScript route handlers automatically run under the request context.
5. `goja-site` DB, guard, Kanban, and render spans become children of the HTTP request span without requiring site authors to pass `req.context` manually.

Do **not** expose Go `context.Context` directly into JavaScript as a normal user-facing API. JavaScript authors should keep writing:

```javascript
db.query("SELECT * FROM cards WHERE session_id = ?", req.session.id)
```

The Go native module should discover the current Go context implicitly and use it for tracing, cancellation, and deadlines.

## Problem Statement

The current tracing slice in `goja-site` can create:

- HTTP spans through `otelhttp`.
- Database spans around `QueryExecer.Query` and `QueryExecer.Exec`.

But the DB spans start from `context.Background()` because the database module interface has no request context. This yields disconnected traces:

```text
Trace A:
  HTTP POST /_kanban/:board/action/:action

Trace B:
  goja-site.db.exec

Trace C:
  goja-site.db.query
```

What we want is one coherent trace:

```text
Trace A:
  HTTP POST /_kanban/:board/action/:action
    ├─ goja-site.js.handler
    │   ├─ goja-site.kanban.action
    │   │   ├─ goja-site.kanban.dispatch
    │   │   │   ├─ goja-site.db.exec
    │   │   │   └─ goja-site.db.query
    │   │   └─ goja-site.kanban.render action_refresh
    │   └─ response write
```

The missing piece is a standard way for Go native modules called from JavaScript to obtain the request context that was active when the JavaScript handler was invoked.

## Current-State Evidence

### `runtimeowner.Runner` already carries context into owner calls

`runtimeowner.Runner` is defined with context-aware methods. `CallFunc` receives `context.Context`, and `Runner.Call` accepts a context at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimeowner/types.go:14-24`:

```go
type CallFunc func(context.Context, *goja.Runtime) (any, error)

type Runner interface {
    Call(ctx context.Context, op string, fn CallFunc) (any, error)
    Post(ctx context.Context, op string, fn PostFunc) error
    Shutdown(ctx context.Context) error
    IsClosed() bool
}
```

`Runner.Call` normalizes the context, applies optional `MaxWait`, checks whether it is already on the owner goroutine, otherwise schedules work on the event loop, wraps the context with owner metadata, and invokes the function with that context. The key execution path is at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimeowner/runner.go:62-105`.

Important observed behavior:

- The caller's context is preserved.
- The scheduled closure receives `ownerCtx := r.withOwnerContext(ctx)`.
- Cancellation is checked before invocation.
- The original caller also waits on either `ctx.Done()` or `resultCh`.

### Owner context is currently about reentrancy, not request-context lookup

`runner.withOwnerContext` stores an internal marker containing the runner and goroutine id at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimeowner/runner.go:198-202`. `isOwnerContext` checks this marker at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimeowner/runner.go:205-210`.

This is used to avoid requeueing nested owner calls when already on the owner goroutine. It is not a public way for arbitrary native modules to retrieve the current request context.

### `gojahttp.Host` already passes the HTTP request context into JavaScript handler execution

`gojahttp.Host.ServeHTTP` matches the route, builds the request DTO, creates the response wrapper, and invokes the JavaScript route through `h.owner.Call(r.Context(), "http-handler", ...)` at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/gojahttp/host.go:79-89`.

This is the most important positive finding: request context already enters the Goja owner call.

The problem is not the HTTP boundary. The problem is native module access from inside the JavaScript call.

### JavaScript request DTO does not include a Go context handle

`RequestDTO` contains method, URL, path, query, params, headers, cookies, session, IP, body, and raw body at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/gojahttp/request_response.go:16-28`. Its `Map` exports those fields to JavaScript at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/gojahttp/request_response.go:30-43`.

This is good. Go contexts should not be exported as ordinary JavaScript values. Site authors should not need to thread a hidden context token through every call.

### `runtimebridge` stores runtime lifecycle context and owner, but not current call context

`runtimebridge.Bindings` contains `Context`, `Loop`, and `Owner` at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimebridge/runtimebridge.go:12-18`. The engine stores these bindings per VM at runtime creation, as shown in `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/engine/factory.go:218-222`.

This bridge is useful for native modules that need runtime-scoped lifecycle context and owner-thread scheduling. It does **not** currently provide the current HTTP request context or active owner-call context.

### The database module is not context-aware

The database module interface is currently:

```go
type QueryExecer interface {
    Query(query string, args ...any) (*sql.Rows, error)
    Exec(query string, args ...any) (sql.Result, error)
}
```

This is at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/modules/database/database.go:15-18`.

`DBModule.Query` calls `m.queryExecer.Query(query, flattenArgs(args)...)` at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/modules/database/database.go:195-205`. `DBModule.Exec` calls `m.queryExecer.Exec(query, flattenArgs(args)...)` at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/modules/database/database.go:250-260`.

There is no context parameter in the JS-exposed methods, the database module, or the `QueryExecer` interface. This is why `goja-site`'s current DB tracing wrapper cannot parent DB spans to HTTP spans.

## Design Goals

1. **Correct trace parentage:** DB, guard, Kanban, and native-module spans should be children of the originating HTTP request span.
2. **No JavaScript ergonomics regression:** Site authors should not manually pass `req.context` into every native module call.
3. **Context cancellation:** Native module operations should be able to respect request cancellation and deadlines.
4. **Owner-thread safety:** The design must not violate Goja's single-thread ownership constraints.
5. **Backwards compatibility:** Existing modules and `QueryExecer` implementations should continue compiling where possible.
6. **Low-cardinality observability:** Context propagation must not smuggle user/session/request identifiers into labels or span attributes by default.
7. **Reusable primitive:** The mechanism should help all native modules, not just `database`.

## Recommended Architecture

### Add current-call context storage to `runtimebridge`

`runtimebridge` already maps a `*goja.Runtime` to runtime-scoped bindings. Extend it with a per-runtime current-call context stack.

Conceptual API:

```go
package runtimebridge

func WithCallContext(vm *goja.Runtime, ctx context.Context, fn func() (any, error)) (any, error)
func WithCallContextVoid(vm *goja.Runtime, ctx context.Context, fn func() error) error
func CurrentContext(vm *goja.Runtime) context.Context
```

Behavior:

- If `ctx` is nil, use the runtime lifecycle context from `Bindings.Context` or `context.Background()`.
- `WithCallContext` pushes `ctx` for the VM before invoking `fn` and pops it afterward.
- `CurrentContext(vm)` returns the top-of-stack context for the VM if one exists.
- If no call context exists, return `Bindings.Context` if available; otherwise `context.Background()`.
- The implementation should be safe for nested owner calls.

Because each Goja VM is owner-goroutine serialized, a per-VM stack is sufficient if all reads/writes occur on the owner goroutine. To be defensive and race-testable, protect the stack with a mutex.

Sketch:

```go
type callContextStack struct {
    mu    sync.Mutex
    stack []context.Context
}

var callContextsByVM sync.Map // map[*goja.Runtime]*callContextStack

func WithCallContext(vm *goja.Runtime, ctx context.Context, fn func() (any, error)) (any, error) {
    if vm == nil || fn == nil {
        return nil, errors.New("runtimebridge: nil vm or fn")
    }
    if ctx == nil {
        ctx = Context(vm)
    }
    st := getCallContextStack(vm)
    st.push(ctx)
    defer st.pop()
    return fn()
}

func CurrentContext(vm *goja.Runtime) context.Context {
    if vm == nil {
        return context.Background()
    }
    if st, ok := lookupStack(vm); ok {
        if ctx, ok := st.peek(); ok && ctx != nil {
            return ctx
        }
    }
    if bindings, ok := Lookup(vm); ok && bindings.Context != nil {
        return bindings.Context
    }
    return context.Background()
}
```

### Wrap owner invocations in `runtimeowner.Runner.invoke`

The best place to set current context is `runtimeowner.runner.invoke`, because every owner `Call` reaches it. It currently calls `fn(ctx, r.vm)` at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimeowner/runner.go:161-179`.

Proposed change:

```go
func (r *runner) invoke(ctx context.Context, op string, fn CallFunc) (any, error) {
    invokeFn := func() (any, error) {
        if !r.opts.RecoverPanics {
            return fn(ctx, r.vm)
        }
        // existing panic recovery around fn(ctx, r.vm)
    }
    return runtimebridge.WithCallContext(r.vm, ctx, invokeFn)
}
```

And for `Post`:

```go
func (r *runner) invokePost(ctx context.Context, op string, fn PostFunc) {
    _ = runtimebridge.WithCallContextVoid(r.vm, ctx, func() error {
        fn(ctx, r.vm)
        return nil
    })
}
```

This means native modules invoked synchronously from JavaScript can call:

```go
ctx := runtimebridge.CurrentContext(vm)
```

and receive the request context when running under `gojahttp.Host`.

### Add context-aware database interfaces while keeping compatibility

Extend `modules/database` with optional context-aware interfaces:

```go
type QueryExecerContext interface {
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
```

Do not remove the current `QueryExecer` immediately.

Update `DBModule.Query`:

```go
func (m *DBModule) Query(query string, args ...any) ([]map[string]any, error) {
    ctx := runtimebridge.CurrentContext(vm?) // see loader binding below
    rows, err := queryContext(ctx, m.queryExecer, query, flattenArgs(args)...)
    ...
}
```

The challenge is that `DBModule.Query` currently does not receive `*goja.Runtime`. There are two good ways to solve that.

## Database Module Implementation Options

### Option A: Bind VM at loader time and export closures instead of methods

Currently `Loader` exports method values:

```go
modules.SetExport(exports, m.Name(), "query", m.Query)
modules.SetExport(exports, m.Name(), "exec", m.Exec)
```

Change the loader to export closures that close over `vm`:

```go
func (m *DBModule) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)
    modules.SetExport(exports, m.Name(), "query", func(query string, args ...any) ([]map[string]any, error) {
        return m.QueryContext(runtimebridge.CurrentContext(vm), query, args...)
    })
    modules.SetExport(exports, m.Name(), "exec", func(query string, args ...any) (map[string]any, error) {
        return m.ExecContext(runtimebridge.CurrentContext(vm), query, args...)
    })
}
```

Add Go methods:

```go
func (m *DBModule) QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error)
func (m *DBModule) ExecContext(ctx context.Context, query string, args ...any) (map[string]any, error)
```

Keep `Query` and `Exec` as compatibility wrappers:

```go
func (m *DBModule) Query(query string, args ...any) ([]map[string]any, error) {
    return m.QueryContext(context.Background(), query, args...)
}
```

This is the recommended option.

### Option B: Store VM on `DBModule`

At loader time, set `m.vm = vm`, then `m.Query` can call `runtimebridge.CurrentContext(m.vm)`.

This is not ideal because a single `DBModule` value may be registered across multiple runtimes. Storing a VM on the module can create races or incorrect runtime association if modules are reused.

### Option C: Expose explicit JS `req.context`

JavaScript would call:

```javascript
db.query(req.context, "SELECT ...")
```

Reject this option. It leaks Go implementation details into JavaScript and makes site authors responsible for context propagation.

## Query/Exec Adapter Design

Add helper functions:

```go
func queryRows(ctx context.Context, qe QueryExecer, query string, args ...any) (*sql.Rows, error) {
    if ctx == nil {
        ctx = context.Background()
    }
    if qec, ok := qe.(QueryExecerContext); ok {
        return qec.QueryContext(ctx, query, args...)
    }
    return qe.Query(query, args...)
}

func execResult(ctx context.Context, qe QueryExecer, query string, args ...any) (sql.Result, error) {
    if ctx == nil {
        ctx = context.Background()
    }
    if qec, ok := qe.(QueryExecerContext); ok {
        return qec.ExecContext(ctx, query, args...)
    }
    return qe.Exec(query, args...)
}
```

This preserves existing implementations and allows `*sql.DB` to be wrapped in a context-aware adapter.

Note: `*sql.DB` has `QueryContext`/`ExecContext`, but it does not satisfy the variadic custom `QueryExecerContext` exactly unless the method signatures match. They do match if defined as above.

## goja-site Integration After go-go-goja Update

Once `go-go-goja/modules/database` supports context-aware execution, update `goja-site` DB wrappers:

```go
type simpleDB struct {
    db          *sql.DB
    allowWrites bool
}

func (s *simpleDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    if !s.allowWrites && !isReadOnlySQL(query) { ... }
    return s.db.QueryContext(ctx, query, args...)
}

func (s *simpleDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
    if !s.allowWrites { ... }
    return s.db.ExecContext(ctx, query, args...)
}
```

For `dbguard.MeteredDB`:

```go
func (m *MeteredDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    return m.inner.QueryContext(ctx, query, args...)
}

func (m *MeteredDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
    if err := m.guard.BeforeExecContext(ctx, query); err != nil { ... }
    result, err := m.inner.ExecContext(ctx, query, args...)
    ...
}
```

For observability wrapper:

```go
func (i *InstrumentedQueryExecer) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    ctx, span := i.tracer.Start(ctx, "goja-site.db.query", ...)
    rows, err := i.inner.QueryContext(ctx, query, args...)
    ...
}
```

This makes DB spans children of the HTTP span automatically.

## Context Flow Diagram

Desired flow:

```text
net/http request
  r.Context() contains OTel HTTP span
      |
      v
gojahttp.Host.ServeHTTP
  h.owner.Call(r.Context(), "http-handler", fn)
      |
      v
runtimeowner.Runner.Call
  schedules onto owner loop
  wraps ctx with owner marker
      |
      v
runtimeowner.invoke
  runtimebridge.WithCallContext(vm, ctx, fn)
      |
      v
JavaScript handler executes
  app.get("/", (req, res) => { db.query(...) })
      |
      v
database module closure
  ctx := runtimebridge.CurrentContext(vm)
  m.QueryContext(ctx, sql, args...)
      |
      v
go-site InstrumentedQueryExecer.QueryContext
  tracer.Start(ctx, "goja-site.db.query")
      |
      v
*sql.DB.QueryContext(ctx, ...)
```

Resulting trace:

```text
HTTP GET /
  └─ goja-site.db.query
```

For Kanban:

```text
HTTP POST /_kanban/:board/action/cardMoved
  ├─ goja-site.kanban.action
  │   ├─ goja-site.kanban.dispatch
  │   │   └─ goja-site.db.exec
  │   └─ goja-site.kanban.render
  │       └─ goja-site.db.query
```

## Cancellation Behavior

Correct context propagation also improves cancellation. If a client disconnects or a request deadline expires:

1. `r.Context()` is canceled.
2. `runtimeowner.Call` observes cancellation while waiting or before invocation.
3. Native modules retrieve the same canceled context.
4. `sql.DB.QueryContext`/`ExecContext` can abort where the driver supports it.
5. OTel spans record errors/cancellation.

This is a production reliability improvement, not only a tracing improvement.

## Implementation Plan

### Phase A: go-go-goja runtimebridge current context

Files:

```text
pkg/runtimebridge/runtimebridge.go
pkg/runtimeowner/runner.go
pkg/runtimeowner/runner_test.go
```

Tasks:

1. Add current-call context stack keyed by `*goja.Runtime`.
2. Add `WithCallContext`, `WithCallContextVoid`, and `CurrentContext`.
3. Wrap `runner.invoke` and `runner.invokePost` with `runtimebridge.WithCallContext`.
4. Add tests:
   - `CurrentContext` returns background/lifecycle context outside calls.
   - `CurrentContext` inside `Runner.Call` returns caller context values.
   - Nested owner calls restore the previous context after return.
   - Canceled context is visible inside native call.
   - Race test with serialized owner calls.

Test sketch:

```go
type ctxKey string

ctx := context.WithValue(context.Background(), ctxKey("request"), "abc")
_, err := r.Call(ctx, "test.current-context", func(callCtx context.Context, vm *goja.Runtime) (any, error) {
    got := runtimebridge.CurrentContext(vm).Value(ctxKey("request"))
    if got != "abc" { t.Fatalf(...) }
    return nil, nil
})
```

### Phase B: database module context support in go-go-goja

Files:

```text
modules/database/database.go
modules/database/database_test.go
```

Tasks:

1. Add `QueryExecerContext` interface.
2. Add `QueryContext` and `ExecContext` methods to `DBModule`.
3. Export JS `query`/`exec` as closures that call `runtimebridge.CurrentContext(vm)`.
4. Keep `Query` and `Exec` wrappers for compatibility.
5. Add tests:
   - context-aware fake DB receives context value.
   - legacy `QueryExecer` still works.
   - canceled context reaches fake DB.
   - no JS API change required.

Fake DB sketch:

```go
type contextRecordingDB struct { got context.Context }
func (db *contextRecordingDB) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
    db.got = ctx
    return emptyRows(), nil
}
```

If returning real `*sql.Rows` is cumbersome in tests, test `ExecContext` first because fake `sql.Result` is easy.

### Phase C: goja-site context-aware DB wrappers

Files:

```text
pkg/app/database.go
pkg/dbguard/metered.go
pkg/observability/sql.go
```

Tasks:

1. Add `QueryContext`/`ExecContext` to `simpleDB`.
2. Add `QueryContext`/`ExecContext` to `dbguard.MeteredDB`.
3. Add `QueryContext`/`ExecContext` to `observability.InstrumentedQueryExecer`.
4. Start DB spans with incoming context instead of `context.Background()`.
5. Keep existing `Query`/`Exec` wrappers for compatibility.
6. Add tests that DB spans are parented under HTTP spans using an in-memory trace exporter.

### Phase D: Kanban and guard span context

Once current context is available from `runtimebridge.CurrentContext(vm)`:

- `kanbanddsl.Board.Mount` can start spans using current context.
- `dbguard` can receive context through `BeforeExecContext`/`AfterExecContext`/`CheckNowContext` or through the DB wrapper.

Do not add more `context.Background()` spans for these paths if a request context can now be retrieved.

### Phase E: Update goja-site dependency

After go-go-goja changes are released or replaced locally:

1. Update `go.mod` to the new version or a local `replace` during development.
2. Remove temporary standalone DB span context fallback where possible.
3. Verify traces with a local OTel Collector.
4. Run the benchmark harness with `--otel --otel-sample-ratio 1`.

## API Sketches

### runtimebridge

```go
package runtimebridge

func CurrentContext(vm *goja.Runtime) context.Context

func WithCallContext(vm *goja.Runtime, ctx context.Context, fn func() (any, error)) (any, error)

func WithCallContextVoid(vm *goja.Runtime, ctx context.Context, fn func() error) error
```

### database module

```go
type QueryExecerContext interface {
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (m *DBModule) QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error)
func (m *DBModule) ExecContext(ctx context.Context, query string, args ...any) (map[string]any, error)
```

### goja-site wrappers

```go
type ContextQueryExecer interface {
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
```

## Alternatives Considered

### Pass context explicitly through JavaScript

Example:

```javascript
db.query(req.context, "SELECT ...")
```

Rejected. This pollutes the JavaScript API, is easy to forget, and exposes a Go implementation detail. It also makes existing scripts harder to migrate.

### Store context globally in goja-site only

Rejected as a long-term solution. The problem is generic to all `go-go-goja` native modules. Solving it in `goja-site` alone creates a special case and does not help modules like `fs`, `exec`, future HTTP clients, or user-authored native modules.

### Use goroutine-local storage

Rejected. Go has no safe official goroutine-local API. The existing runtime owner marker uses goroutine id only to validate owner reentry, not to carry arbitrary request context through the system.

### Only rely on metrics, not traces

Rejected. Metrics are sufficient for aggregate latency analysis, but they cannot explain one specific slow request. Correct trace parentage is still valuable.

## Risks and Mitigations

### Risk: context stack leaks after panic

Mitigation: use `defer pop()` inside `WithCallContext`, and ensure `runner.invoke` still recovers panics as it does today.

### Risk: context visible across concurrent requests

Mitigation: Goja owner calls are serialized. Use a stack protected by mutex and tests under `-race`. Do not read current context from arbitrary goroutines unless explicitly documented to fall back to lifecycle context.

### Risk: async promise continuations lose request context

`gojahttp.Host.awaitAndFinishPromise` calls `h.owner.Call(ctx, "http-handler.promise-state", ...)` with the original request context at `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/gojahttp/host.go:117-145`, so promise state polling can preserve request context. Other asynchronous module callbacks need review.

### Risk: long-running background timers inherit request context incorrectly

If a JS route schedules a timer, the timer callback should probably run with the runtime lifecycle context or an explicitly captured context depending on module semantics. Do not blindly retain request context for background work after the request completes.

### Risk: DB spans still disconnected if wrappers do not implement context interfaces

Mitigation: add tests in goja-site ensuring `InstrumentedQueryExecer.QueryContext` is used when called from JS under HTTP request context.

## Validation Strategy

### go-go-goja tests

```bash
go test ./pkg/runtimebridge ./pkg/runtimeowner ./modules/database ./pkg/gojahttp -race
```

Test cases:

- current context available inside owner call,
- current context nested stack restore,
- database module receives current request context,
- legacy DB implementation still works,
- HTTP handler DB call receives `r.Context()` value.

### goja-site tests

```bash
go test ./...
```

Add an in-memory trace exporter test:

1. Start test server with tracing provider using in-memory exporter.
2. Request a route that calls `db.query`.
3. Assert trace contains one HTTP span and one DB child span with same trace id and parent span id.

### Manual trace smoke

Start local collector, then run:

```bash
scripts/bench-vegeta.sh \
  --scenario db-read \
  --duration 5s \
  --rate 5/s \
  --port 18184 \
  --metrics-port 19194 \
  --otel \
  --otel-sample-ratio 1
```

Expected trace tree:

```text
GET /read
  └─ goja-site.db.query
```

## Implementation Checklist

1. Implement `runtimebridge.CurrentContext` and call-context stack.
2. Wrap `runtimeowner.runner.invoke` and `invokePost`.
3. Add runtimeowner/runtimebridge tests.
4. Add `QueryExecerContext` to database module.
5. Export DB JS functions as VM-closing closures.
6. Add database context tests.
7. Release or replace `go-go-goja` in `goja-site`.
8. Add context-aware methods to `simpleDB`, `MeteredDB`, and `InstrumentedQueryExecer`.
9. Add in-memory trace parentage test in `goja-site`.
10. Add Kanban and guard spans using propagated context.
11. Run benchmarks with tracing sampled and tracing at 100% short-run.
12. Update GOJA-PERF-BENCH diary and reMarkable bundle.

## References

- `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimeowner/types.go`: context-aware `Runner` interface.
- `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimeowner/runner.go`: owner call scheduling, owner context marker, cancellation behavior.
- `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/gojahttp/host.go`: HTTP request context passed to JavaScript handler through `Owner.Call`.
- `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/gojahttp/request_response.go`: JavaScript request DTO shape.
- `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/pkg/runtimebridge/runtimebridge.go`: current runtime lifecycle bindings per VM.
- `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/engine/factory.go`: runtimebridge bindings registered during runtime creation.
- `/home/manuel/go/pkg/mod/github.com/go-go-golems/go-go-goja@v0.4.16/modules/database/database.go`: non-context-aware database module and `QueryExecer` interface.
- `pkg/observability/tracing.go`: current `goja-site` OTel setup.
- `pkg/observability/sql.go`: current standalone DB spans that should become request-parented after context propagation.
