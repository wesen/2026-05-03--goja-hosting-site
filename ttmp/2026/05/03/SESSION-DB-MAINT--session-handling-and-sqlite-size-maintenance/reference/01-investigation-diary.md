---
Title: Investigation Diary
Ticket: SESSION-DB-MAINT
Status: active
Topics:
    - go
    - goja
    - javascript
    - web
    - sqlite
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Chronological diary for the session handling and SQLite size maintenance design ticket."
LastUpdated: 2026-05-03T16:45:00-04:00
WhatFor: "Record why this ticket was created, what documents were written, and what implementation directions were chosen."
WhenToUse: "Read before implementing db.guard, changing session semantics, or designing cleanup callbacks."
---

# Investigation Diary

## Goal

This diary records work for the `SESSION-DB-MAINT` ticket. The ticket has two purposes:

1. Document the Go-side cookie session implementation that now exists in `goja-site`.
2. Design a SQLite size guard that can monitor database growth and call a server-side JavaScript cleanup callback when the database grows past a configured limit.

The unifying theme is session-scoped storage. Once each browser has its own session ID, apps can keep per-session rows in SQLite. That creates a new maintenance problem: old sessions and large sessions can accumulate rows. The host should provide enough infrastructure to measure and trigger cleanup without deciding app-specific deletion policy.

## Step 1: Created the ticket and wrote the two core documents

### Prompt Context

**User prompt (verbatim):** "Let's create another ticket for the session, where we a) have an implementation document describing how sessions are handled

b) a second document where I want to design out the following: I want the db module to track the size of the sqlite db so that it doesn't grow past a certain size, and if it does, call a JS callback in order to do cleanup (if the callback is set). How can we do that without having to modify the original db module (but we might have to, in which case we should write our own db native module in our app). That way, once the DB grows in size, the frontend side can go over the different sessions and cleanup tasks that are too big or so or cleanup older sessions.

Upload the bundle of the ticket to remarkable when done"

**Assistant interpretation:** Create a new docmgr ticket with two long-form design documents and upload the resulting bundle to reMarkable. The first document should describe the implemented session mechanics. The second should reason carefully about how to add database size monitoring and cleanup callbacks, preferably without modifying the upstream database module.

**Inferred user intent:** The user wants durable design documentation before another implementation phase, especially because DB quota and cleanup behavior touches multiple layers: Go host, Goja runtime ownership, the existing database native module, sessions, and application-specific cleanup logic.

### What I did

- Created ticket `SESSION-DB-MAINT` with title `Session Handling and SQLite Size Maintenance`.
- Added design document `01-session-cookie-implementation-guide.md`.
- Added design document `02-sqlite-size-guard-and-cleanup-callback-design.md`.
- Added this investigation diary.
- Inspected the current upstream database module in `../corporate-headquarters/go-go-goja/modules/database/database.go`.
- Confirmed that the upstream database module already accepts a `QueryExecer` through `WithPreconfiguredDB`, which gives us a clean wrapper point.

### Key findings

The session implementation is already small and clear:

- `pkg/web/session.go` creates and validates opaque session IDs.
- `pkg/web/host.go` attaches sessions before dynamic route dispatch.
- `pkg/web/request_response.go` exposes `req.session` to JavaScript.
- `pkg/kanbanddsl/mount.go` passes sessions into `ctx.session` and `event.session`.
- `examples/kanban/scripts/app.js` scopes rows with `session_id`.

The database size guard can probably avoid modifying the upstream database module because the module accepts:

```go
type QueryExecer interface {
    Query(query string, args ...any) (*sql.Rows, error)
    Exec(query string, args ...any) (sql.Result, error)
}
```

That means `goja-site` can pass a metered wrapper instead of a raw `*sql.DB`:

```go
meteredDB := dbguard.NewMeteredDB(rawDB, guard)
databasemod.WithPreconfiguredDB(meteredDB)
```

The wrapper can intercept every `db.exec(...)` call while preserving the existing JavaScript API.

### Important design decision

The recommended storage guard design is split into two pieces:

- A `MeteredDB` wrapper around `*sql.DB` catches writes and measures size.
- A `db.guard` module lets JavaScript configure limits, inspect stats, and register `onLimitExceeded(event)` cleanup callbacks.

This preserves `require("database")` for normal SQL and adds `require("db.guard")` for quota policy.

### What should be implemented later

A future implementation should proceed in phases:

1. Implement stats-only measurement for `app.db`, `app.db-wal`, and `app.db-shm`.
2. Add `db.guard.stats()` and `db.guard.checkNow()`.
3. Add JavaScript callback registration.
4. Wrap the DB with a metered `QueryExecer` and trigger checks after writes.
5. Add session-aware cleanup examples for the Kanban app.

### Validation / hygiene

At the time of ticket creation, this was documentation work only. The relevant implementation was already validated in the earlier session work with:

```bash
node -c examples/kanban/scripts/app.js
go test ./... -count=1
docmgr doctor --ticket GOJA-HOSTING-SITE --stale-after 30
docmgr doctor --ticket KANBAN-DSL --stale-after 30
```

This ticket still needs its own `docmgr doctor` and reMarkable upload after the documents are complete.

## Step 2: Implemented `db.guard` and metered SQLite writes

I implemented the first version of the SQLite size guard that the design document proposed. The implementation avoids changing the upstream `go-go-goja/modules/database` module by wrapping the preconfigured `*sql.DB` in a local `MeteredDB` that implements the same `QueryExecer` interface expected by `databasemod.WithPreconfiguredDB`.

### What changed

- Added `pkg/dbguard` with:
  - `Guard`,
  - `MeteredDB`,
  - `Stats`,
  - `CheckResult`,
  - `db.guard` runtime registrar.
- `Guard.Stats()` measures:
  - main SQLite DB file,
  - `-wal` file,
  - `-shm` file,
  - `PRAGMA page_size`,
  - `PRAGMA page_count`,
  - `PRAGMA freelist_count`.
- `MeteredDB.Exec(...)` triggers `guard.AfterExec(...)` after successful writes.
- `db.guard` exposes:
  - `configure(options)`,
  - `onLimitExceeded(fn)`,
  - `stats()`,
  - `checkNow(reason)`,
  - `isOverLimit()`,
  - `lastResult()`.
- `pkg/app/server.go` now creates one guard, wraps SQLite with `MeteredDB`, passes the wrapper to both `database` and `db`, and registers `db.guard`.

### Callback behavior

The cleanup callback is synchronous in v1. This is safe because it is called from `db.exec(...)`, which is already executing inside the Goja runtime. A recursion guard prevents cleanup writes from recursively triggering cleanup callbacks.

### Validation

Commands:

```bash
go test ./pkg/dbguard -count=1
go test ./... -count=1
```

Live smoke test:

- Started a temporary `goja-site` app on `127.0.0.1:60130`.
- The app required `database` and `db.guard`.
- Configured `maxBytes: 1`, `checkEveryWrites: 1`, and a cleanup callback.
- Posted to `/add`, which inserted a row through `db.exec(...)`.
- Observed callback dispatch through the normal database module path.

Evidence:

```text
{'cleanupCalls': 1, 'callbackCalled': True, 'triggered': True, 'totalBytes': 8192}
```

### Remaining work

The implementation currently supports the soft-limit path. The future hard-limit behavior (`failWritesOverHardLimit`) is represented in the option/result types but is not enforced yet. The Kanban example also does not yet register a real cleanup policy; that should be added behind an explicit demo/config flag.
