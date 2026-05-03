---
Title: SQLite Size Guard and Cleanup Callback Design
Ticket: SESSION-DB-MAINT
Status: active
Topics:
    - go
    - goja
    - javascript
    - web
    - sqlite
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/modules/database/database.go
      Note: Upstream database module exposes QueryExecer wrapper point
    - Path: examples/kanban/scripts/app.js
      Note: Current session-scoped cards table is a cleanup target example
    - Path: go.mod
      Note: Shows local go-go-goja replace and dependency context
    - Path: pkg/app/server.go
      Note: |-
        Server currently opens SQLite and wires preconfigured database modules
        Creates Guard
    - Path: pkg/dbguard/guard.go
      Note: Implemented size checks
    - Path: pkg/dbguard/guard_test.go
      Note: Stats and no-callback over-limit tests
    - Path: pkg/dbguard/metered.go
      Note: Metered QueryExecer wrapper used to avoid changing upstream database module
    - Path: pkg/dbguard/registrar.go
      Note: db.guard native module API implementation
    - Path: pkg/dbguard/registrar_test.go
      Note: Runtime integration test for cleanup callback dispatch through database module
    - Path: pkg/web/session.go
      Note: Session identity enables per-session cleanup policies
ExternalSources: []
Summary: Design for monitoring SQLite database size and invoking server-side JavaScript cleanup callbacks without modifying the upstream go-go-goja database module unless necessary.
LastUpdated: 2026-05-03T16:45:00-04:00
WhatFor: Use when implementing database quota/cleanup behavior for goja-site sessions and app-owned data.
WhenToUse: Read before adding DB size limits, cleanup callbacks, session pruning, or replacing/wrapping the database module.
---



# SQLite Size Guard and Cleanup Callback Design

## Executive summary

The next storage problem for `goja-site` is database growth. Once sessions exist, each browser can create its own cards, tasks, logs, and generated app data. That is useful, but it means the single SQLite file can grow without bound. We need a Go-side size guard that can measure the SQLite database, detect when it exceeds a configured limit, and call a server-side JavaScript cleanup callback if the app registered one.

The preferred design is to add a local `db.guard` / storage guard module in `goja-site` and wrap the preconfigured `*sql.DB` with a metered `QueryExecer` before passing it into the existing upstream `go-go-goja` database module. This lets us intercept writes without modifying the original database module. The original module already accepts any `QueryExecer` through `databasemod.WithPreconfiguredDB(db)`, so we can provide a wrapper that implements the same interface.

The cleanup callback should be app-specific JavaScript. Go can know the database file size, WAL size, thresholds, and reason for the check. It cannot know which rows are safe to delete. The app knows its schema and can decide to delete old sessions, trim large task lists, remove old generated artifacts, checkpoint WAL files, or run `VACUUM` when appropriate.

The key design is:

```text
Go measures and triggers.
JavaScript decides what to delete.
SQLite confirms whether cleanup worked.
```

## Problem statement

The current database module gives JavaScript direct access to a preconfigured SQLite connection:

```javascript
const db = require("database");
db.query("SELECT * FROM cards WHERE session_id = ?", sessionId);
db.exec("INSERT INTO cards(...) VALUES (...)");
```

This is intentionally powerful. It lets app scripts define their own schema and behavior. But it also means the Go host does not understand the data model. If a database grows beyond 100 MB, the host cannot safely decide which rows to delete. It does not know whether `cards`, `tasks`, `events`, `imports`, or `logs` are important.

The cleanup decision therefore belongs in JavaScript. But the measurement and trigger logic belong in Go. Application scripts should not have to repeatedly `stat` the database file, remember to include `-wal` files, debounce cleanup calls, or ensure cleanup does not recursively call itself.

The desired app shape is something like:

```javascript
const db = require("database");
const guard = require("db.guard");

guard.configure({
  maxBytes: 50 * 1024 * 1024,
  cooldownMs: 30_000,
  checkEveryWrites: 10,
});

guard.onLimitExceeded(event => {
  const oversized = db.query(`
    SELECT session_id, COUNT(*) AS count
    FROM cards
    GROUP BY session_id
    HAVING count > 200
    ORDER BY count DESC
  `);

  oversized.forEach(row => {
    db.exec(`
      DELETE FROM cards
      WHERE session_id = ?
        AND id NOT IN (
          SELECT id FROM cards
          WHERE session_id = ?
          ORDER BY updated_at DESC
          LIMIT 200
        )
    `, row.session_id, row.session_id);
  });

  return { ok: true, deletedSessionsChecked: oversized.length };
});
```

The app sees a cleanup event and performs schema-aware cleanup. The guard stays generic.

## Requirements

The design should satisfy these requirements:

- Measure the real storage footprint of a SQLite database, including WAL and shared-memory files where relevant.
- Detect size after writes, not only at startup.
- Allow apps to register a JavaScript callback for cleanup.
- Avoid modifying the upstream `go-go-goja/modules/database` module if possible.
- Preserve the existing `require("database")` and `require("db")` app API if possible.
- Prevent cleanup recursion when the callback itself calls `db.exec(...)`.
- Avoid calling cleanup on every write once the database is over the limit; use cooldowns and/or write counters.
- Make the cleanup event rich enough for application decisions and logs.
- Keep the default behavior safe: if no callback is registered, report/record over-limit state but do not delete data.
- Make it possible to fail hard or keep accepting writes as a policy decision.

## SQLite size is not just one file

A SQLite database may have multiple files:

| File | Meaning |
|---|---|
| `app.db` | Main database file. |
| `app.db-wal` | Write-ahead log file when WAL mode is active. |
| `app.db-shm` | Shared memory file used with WAL. |

If we only check `app.db`, we may miss a large WAL file. For a practical quota, the guard should compute:

```text
totalBytes = size(app.db) + size(app.db-wal) + size(app.db-shm)
```

The guard should also expose SQLite internal page stats:

```sql
PRAGMA page_size;
PRAGMA page_count;
PRAGMA freelist_count;
```

These numbers explain whether the file is large because it contains live pages or because it has free pages that could potentially be reclaimed by `VACUUM` or incremental vacuum.

A useful stats object should include:

```javascript
{
  path: "/path/to/app.db",
  fileBytes: 12345678,
  walBytes: 1048576,
  shmBytes: 32768,
  totalBytes: 13418422,
  maxBytes: 10485760,
  overByBytes: 2937662,
  pageSize: 4096,
  pageCount: 3000,
  freelistCount: 500,
  estimatedLiveBytes: 10240000,
  checkedAt: "2026-05-03T20:45:00Z"
}
```

The app can use this to decide whether it should delete rows, checkpoint WAL, vacuum, or simply log.

## Avoiding changes to the upstream database module

The upstream `go-go-goja` database module already has the extension point we need. It accepts a `QueryExecer`:

```go
type QueryExecer interface {
    Query(query string, args ...any) (*sql.Rows, error)
    Exec(query string, args ...any) (sql.Result, error)
}
```

`goja-site` currently passes the raw `*sql.DB`:

```go
databaseModule := databasemod.New(
    databasemod.WithPreconfiguredDB(db),
    databasemod.WithConfigureEnabled(false),
)
```

Instead, we can pass a wrapper:

```go
rawDB, _ := sql.Open("sqlite3", cfg.DBPath)
sizeGuard := dbguard.NewGuard(dbguard.Options{Path: cfg.DBPath, MaxBytes: ...})
meteredDB := dbguard.NewMeteredDB(rawDB, sizeGuard)

databaseModule := databasemod.New(
    databasemod.WithPreconfiguredDB(meteredDB),
    databasemod.WithConfigureEnabled(false),
)
```

The wrapper implements the same interface:

```go
type MeteredDB struct {
    inner *sql.DB
    guard *Guard
}

func (m *MeteredDB) Query(query string, args ...any) (*sql.Rows, error) {
    return m.inner.Query(query, args...)
}

func (m *MeteredDB) Exec(query string, args ...any) (sql.Result, error) {
    result, err := m.inner.Exec(query, args...)
    if err == nil {
        m.guard.AfterExec(query)
    }
    return result, err
}
```

This design leaves the original database module unchanged. The database module still thinks it is calling a `QueryExecer`. The guard sees every write because all `db.exec(...)` calls go through `Exec`.

This is the cleanest path because it preserves the JavaScript API:

```javascript
const db = require("database");
db.exec(...);
db.query(...);
```

and adds a separate control module:

```javascript
const guard = require("db.guard");
guard.onLimitExceeded(...);
```

## Proposed architecture

```mermaid
flowchart TD
    JS[app.js] --> DBModule[require("database") upstream DB module]
    JS --> GuardModule[require("db.guard") local guard module]

    DBModule --> MeteredDB[MeteredDB implements QueryExecer]
    MeteredDB --> SQLite[(SQLite *sql.DB)]
    MeteredDB --> Guard[DB Size Guard]

    GuardModule --> Guard
    Guard --> Stats[Measure app.db + app.db-wal + app.db-shm]
    Guard --> Limit{totalBytes > maxBytes?}
    Limit -->|no| Done[Return]
    Limit -->|yes| Callback{JS callback registered?}
    Callback -->|no| Log[Record over-limit state]
    Callback -->|yes| Cleanup[Call Goja cleanup callback]
    Cleanup --> JS
    JS --> DBModule
```

The important detail is the shared guard object. The metered database wrapper uses it to trigger checks after writes. The `db.guard` module uses it to register callbacks, configure limits, and expose manual inspection.

## The JavaScript API

A first version of the guard module could expose:

```typescript
type DBGuard = {
  configure(options: GuardOptions): void;
  onLimitExceeded(callback: (event: LimitExceededEvent) => CleanupResult): void;
  stats(): DBStats;
  checkNow(reason?: string): CheckResult;
  isOverLimit(): boolean;
};
```

Configuration:

```typescript
type GuardOptions = {
  maxBytes?: number;
  softMaxBytes?: number;
  hardMaxBytes?: number;
  cooldownMs?: number;
  checkEveryWrites?: number;
  includeWal?: boolean;
  checkpointBeforeCheck?: boolean;
  failWritesOverHardLimit?: boolean;
};
```

Event shape:

```typescript
type LimitExceededEvent = {
  reason: "startup" | "afterExec" | "interval" | "manual";
  query?: string;
  stats: DBStats;
  thresholds: {
    softMaxBytes?: number;
    hardMaxBytes?: number;
    maxBytes: number;
  };
  overByBytes: number;
  cleanupAttempt: number;
  inCallback: true;
};
```

Result shape:

```typescript
type CleanupResult = {
  ok?: boolean;
  message?: string;
  deletedRows?: number;
  deletedSessions?: number;
  vacuum?: boolean;
  checkpoint?: boolean;
  data?: any;
};
```

Example app usage:

```javascript
const db = require("database");
const guard = require("db.guard");

guard.configure({
  maxBytes: 50 * 1024 * 1024,
  cooldownMs: 60_000,
  checkEveryWrites: 20,
  includeWal: true,
});

guard.onLimitExceeded(event => {
  console.log("database is over limit", event.stats.totalBytes, event.overByBytes);

  // Example policy: keep only the newest 200 cards per session.
  const sessions = db.query(`
    SELECT session_id, COUNT(*) AS count
    FROM cards
    GROUP BY session_id
    HAVING count > 200
  `);

  let deletedRows = 0;
  sessions.forEach(session => {
    const result = db.exec(`
      DELETE FROM cards
      WHERE session_id = ?
        AND id NOT IN (
          SELECT id FROM cards
          WHERE session_id = ?
          ORDER BY updated_at DESC, id DESC
          LIMIT 200
        )
    `, session.session_id, session.session_id);
    deletedRows += Number(result.rowsAffected || 0);
  });

  // If many rows were deleted, ask SQLite to reclaim pages.
  if (deletedRows > 0) {
    db.exec("PRAGMA wal_checkpoint(TRUNCATE)");
    db.exec("VACUUM");
  }

  return { ok: true, deletedRows, checkedSessions: sessions.length };
});
```

## Callback timing

There are four useful check moments:

| Moment | Reason | Notes |
|---|---|---|
| Startup | `startup` | Lets the app clean up an already-large database. |
| After writes | `afterExec` | Catches growth caused by app mutations. |
| Manual call | `manual` | Useful from admin endpoints or tests. |
| Interval | `interval` | Useful for WAL growth or external writes. |

The first implementation should support startup, after-exec, and manual checks. Interval checks can be added later if needed.

After-exec checks should be throttled. Measuring file sizes is cheap, but not free. More importantly, cleanup callbacks can be expensive. Recommended defaults:

```go
CheckEveryWrites: 10
Cooldown: 30 * time.Second
```

The guard should track:

```go
type Guard struct {
    writeCount int64
    lastCleanup time.Time
    inCleanup bool
}
```

The rough algorithm is:

```text
AfterExec(query):
  increment write count
  if disabled: return
  if inCleanup: return
  if writeCount % checkEveryWrites != 0: return
  if now - lastCleanup < cooldown: return
  stats = measure()
  if stats.totalBytes <= maxBytes: return
  if no callback: record over-limit and return
  inCleanup = true
  call callback(event)
  inCleanup = false
  lastCleanup = now
  measure again and record result
```

The `inCleanup` guard is critical because cleanup callbacks will call `db.exec(...)`. Without this flag, cleanup deletes could recursively trigger cleanup again.

## Calling JavaScript from the guard

The callback is registered from JavaScript through `db.guard`:

```javascript
guard.onLimitExceeded(event => { ... });
```

The guard module must store a `goja.Callable` and the runtime it belongs to:

```go
type Guard struct {
    vm *goja.Runtime
    callback goja.Callable
}
```

A synchronous after-exec callback is possible because `MeteredDB.Exec` is invoked from the Go function backing `db.exec(...)`, which is already running inside the Goja runtime owner. Calling the registered callback directly is acceptable as long as the callback belongs to the same runtime and recursion is guarded.

Pseudo-implementation:

```go
func (g *Guard) maybeCallCallback(reason string, query string, stats Stats) {
    if g.callback == nil || g.inCleanup {
        return
    }
    g.inCleanup = true
    defer func() { g.inCleanup = false }()

    event := map[string]any{
        "reason": reason,
        "query": query,
        "stats": stats.Map(),
        "overByBytes": stats.TotalBytes - g.maxBytes,
    }
    result, err := g.callback(goja.Undefined(), g.vm.ToValue(event))
    if err != nil {
        g.lastError = err.Error()
        return
    }
    g.lastResult = result.Export()
}
```

If a future interval checker runs in a Go goroutine, it must not call the Goja callback directly. It must enter through the runtime owner:

```go
owner.Call(ctx, "db-guard-cleanup", func(ctx context.Context, vm *goja.Runtime) (any, error) {
    return g.callCallbackInRuntime(vm, event)
})
```

That distinction matters. Synchronous callbacks triggered inside `db.exec` are already in the runtime. Background callbacks are not.

## Soft limit and hard limit

The guard should eventually distinguish a soft limit from a hard limit.

A soft limit means:

```text
The DB is too large. Try cleanup if a callback is registered, but do not fail the original write.
```

A hard limit means:

```text
The DB is dangerously large. Optionally fail writes if cleanup cannot reduce size.
```

Recommended v1 behavior:

- Implement `maxBytes` as a soft limit.
- Log and expose over-limit state when no callback exists.
- Do not fail writes by default.
- Add `failWritesOverHardLimit` later, after the cleanup path is proven.

Failing writes is tempting, but it changes application semantics. A user creating a card could get a 500 because cleanup failed. That may be correct for some apps, but the default should be conservative.

## Cleanup policy examples

The guard should not prescribe cleanup policy, but the documentation should give app authors patterns.

### Keep newest N cards per session

```sql
DELETE FROM cards
WHERE session_id = ?
  AND id NOT IN (
    SELECT id FROM cards
    WHERE session_id = ?
    ORDER BY updated_at DESC, id DESC
    LIMIT 200
  );
```

This policy is simple and works for task/card apps.

### Delete inactive sessions

If the app maintains a session metadata table:

```sql
CREATE TABLE IF NOT EXISTS app_sessions (
  session_id TEXT PRIMARY KEY,
  first_seen TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_seen TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

then cleanup can remove old sessions:

```sql
DELETE FROM cards
WHERE session_id IN (
  SELECT session_id
  FROM app_sessions
  WHERE last_seen < datetime('now', '-30 days')
);

DELETE FROM app_sessions
WHERE last_seen < datetime('now', '-30 days');
```

The current Go session layer does not maintain this table. That is intentional. Apps that need age-based cleanup can add it.

### Delete largest sessions first

```sql
SELECT session_id, COUNT(*) AS cards
FROM cards
GROUP BY session_id
ORDER BY cards DESC
LIMIT 10;
```

Then trim those sessions first. This is useful when a few sessions dominate storage.

### Reclaim SQLite file space

Deletes do not always shrink the database file immediately. After significant cleanup, the callback may need:

```sql
PRAGMA wal_checkpoint(TRUNCATE);
VACUUM;
```

`VACUUM` can be expensive and may require temporary disk space. It should not run after every small cleanup. A callback can decide to run it only when many rows were deleted or when `freelist_count` is high.

## Measuring after cleanup

After the callback returns, the guard should measure again and store both before and after stats:

```javascript
{
  reason: "afterExec",
  before: { totalBytes: 73400320, ... },
  callbackResult: { ok: true, deletedRows: 1200 },
  after: { totalBytes: 48234496, ... },
  reducedByBytes: 25165824,
  stillOverLimit: false
}
```

This lets `guard.stats()` or logs answer the important question: did cleanup actually work?

## `db.guard` module API sketch

The Go module registrar could look like:

```go
type GuardRegistrar struct {
    guard *dbguard.Guard
}

func (r *GuardRegistrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *require.Registry) error {
    r.guard.SetRuntime(ctx.Owner, ctx.Runtime)
    reg.RegisterNativeModule("db.guard", r.loader)
    return nil
}
```

The loader exports:

```go
func (r *GuardRegistrar) loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)
    _ = exports.Set("configure", func(v goja.Value) error { ... })
    _ = exports.Set("onLimitExceeded", func(fn goja.Value) error { ... })
    _ = exports.Set("stats", func() map[string]any { return r.guard.Stats().Map() })
    _ = exports.Set("checkNow", func(reason string) map[string]any { return r.guard.CheckNow(reason).Map() })
    _ = exports.Set("isOverLimit", func() bool { return r.guard.IsOverLimit() })
}
```

From JavaScript:

```javascript
const guard = require("db.guard");

guard.configure({ maxBytes: 50 * 1024 * 1024 });

guard.onLimitExceeded(event => {
  // app-specific cleanup
  return { ok: true };
});
```

This keeps `database` focused on SQL and `db.guard` focused on storage policy.

## If wrapping is not enough: local database module

The wrapper approach should be tried first. It avoids modifying the original database module and preserves the existing JavaScript API. However, there are cases where a local database module may be better:

- We want callback invocation after the JavaScript-facing `db.exec` result map is constructed, not inside the lower-level `QueryExecer.Exec` call.
- We want to expose richer write metadata such as `rowsAffected`, `lastInsertId`, and query duration in the cleanup event.
- We want transaction-aware cleanup behavior.
- We want per-call options such as `db.exec({ sql, args, skipGuard: true })`.
- We want the guard and database module to share one JavaScript-facing configuration namespace.

A local module could copy the small upstream database module into `pkg/appdb` or `pkg/sitedb`, then add metering directly around its exported `Exec` method:

```go
func (m *Module) Exec(query string, args ...any) (map[string]any, error) {
    result, err := m.innerExec(query, args...)
    if err == nil && result["success"] == true {
        m.guard.AfterExec(query, result)
    }
    return result, err
}
```

This is more code and creates drift from `go-go-goja/modules/database`, so it should be a fallback. The project should prefer composition through `QueryExecer` until there is a clear reason not to.

## Can this be done entirely in JavaScript?

A pure JavaScript solution is possible:

```javascript
function guardedExec(sql, ...args) {
  const result = db.exec(sql, ...args);
  maybeCheckSizeAndCleanup();
  return result;
}
```

But this does not meet the goal. Every app would have to remember to use `guardedExec` instead of `db.exec`. Existing modules such as `kanban.dsl` callbacks may call normal app functions that call `db.exec`. A Go-side wrapper catches all writes through the configured database module without changing app habits.

A pure JavaScript solution also cannot easily and portably measure `app.db-wal` unless `fs` is enabled and the app knows the database path. The Go host already knows the DB path.

## Session-aware cleanup design

The session implementation gives us session IDs, but it does not maintain a session registry table. That is a good default. Cleanup policy should be app-owned. For apps that want cleanup by session age or size, the recommended pattern is to add an app table:

```sql
CREATE TABLE IF NOT EXISTS app_sessions (
  session_id TEXT PRIMARY KEY,
  first_seen TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_seen TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Then, in normal routes:

```javascript
function touchSession(session) {
  db.exec(`
    INSERT INTO app_sessions(session_id, first_seen, last_seen)
    VALUES (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    ON CONFLICT(session_id) DO UPDATE SET last_seen = CURRENT_TIMESTAMP
  `, session.id);
}

app.get("/", (req, res) => {
  touchSession(req.session);
  res.html(boardPage(req));
});
```

Cleanup callbacks can then reason about old sessions:

```javascript
guard.onLimitExceeded(event => {
  const old = db.query(`
    SELECT session_id
    FROM app_sessions
    WHERE last_seen < datetime('now', '-14 days')
  `);

  old.forEach(s => {
    db.exec("DELETE FROM cards WHERE session_id = ?", s.session_id);
    db.exec("DELETE FROM app_sessions WHERE session_id = ?", s.session_id);
  });

  return { ok: true, deletedSessions: old.length };
});
```

This is probably the right level of abstraction. Go gives each browser an ID and tells the app when the database is too large. The app decides which sessions are old, too large, or safe to trim.

## Implementation plan

### Phase 1: Stats-only guard

Add `pkg/dbguard` with:

- `Stats` type,
- file size measurement for `db`, `db-wal`, and `db-shm`,
- SQLite pragma stats,
- `Stats().Map()` for JavaScript export,
- tests using a temporary SQLite DB.

Expose `require("db.guard").stats()` and `checkNow()` without automatic callbacks first.

### Phase 2: JavaScript callback registration

Add runtime registrar support:

- `guard.onLimitExceeded(fn)`,
- `guard.configure(options)`,
- callback storage per runtime,
- manual `checkNow("manual")` invokes callback if over limit.

Add tests that register a callback and force a low `maxBytes`.

### Phase 3: Metered QueryExecer wrapper

Wrap `*sql.DB` before passing it to the upstream database module:

```go
metered := dbguard.NewMeteredDB(db, guard)
databasemod.WithPreconfiguredDB(metered)
```

Trigger checks after writes, with:

- write counter,
- cooldown,
- recursion guard,
- no callback when no limit configured.

### Phase 4: Example cleanup policy

Update the Kanban example with an optional cleanup policy:

- configure `db.guard` with a small demo limit only if an environment flag or app setting is enabled,
- cleanup old/largest session rows,
- optionally run `PRAGMA wal_checkpoint(TRUNCATE)` and `VACUUM`.

The example should not surprise users by deleting data under normal settings. It should be documented and easy to enable for demonstration.

### Phase 5: Hard-limit policy

Only after soft cleanup is proven, add optional hard limit behavior:

```javascript
guard.configure({
  hardMaxBytes: 100 * 1024 * 1024,
  failWritesOverHardLimit: true,
});
```

This should be opt-in because it can change user-facing app behavior.

## Testing strategy

### Unit tests

- Stats include main DB, WAL, and SHM files when present.
- Missing WAL/SHM files count as zero.
- `maxBytes` comparison computes `overByBytes` correctly.
- Cooldown prevents repeated callbacks.
- `inCleanup` prevents recursion.

### Runtime integration tests

- `require("db.guard")` loads.
- `guard.configure(...)` changes thresholds.
- `guard.onLimitExceeded(fn)` stores callback.
- `guard.checkNow("manual")` calls callback when over limit.
- Callback can call `db.exec(...)` without recursive cleanup.

### End-to-end tests

- Create a temporary SQLite DB.
- Configure a tiny max size.
- Insert rows through `require("database").exec(...)`.
- Verify cleanup callback fires.
- Verify callback deletes rows.
- Verify final stats are recorded.

### Live app validation

For the Kanban example:

1. Start with a tiny quota.
2. Create many cards in one session.
3. Confirm cleanup callback sees the over-limit event.
4. Confirm it can delete rows from the largest/oldest sessions.
5. Confirm another session remains usable.

## Risks and mitigations

### Risk: cleanup callback recursively triggers itself

Mitigation: `inCleanup` guard. Writes during cleanup should still execute, but they should not trigger a nested cleanup callback.

### Risk: cleanup runs too often

Mitigation: `checkEveryWrites` and `cooldownMs`.

### Risk: cleanup deletes wrong data

Mitigation: keep cleanup app-owned. Go should not delete rows automatically.

### Risk: `VACUUM` is expensive

Mitigation: make `VACUUM` a callback decision. The guard can report `freelist_count`; the app decides whether reclaiming disk space is worth the cost.

### Risk: background checks call Goja unsafely

Mitigation: only call JS directly when already inside Goja execution. Any goroutine/interval-based checks must use the runtime owner.

### Risk: measuring only file size misses WAL growth

Mitigation: include `db-wal` and `db-shm` in total size.

## Recommended v1 design

The recommended first implementation is:

- Add `pkg/dbguard`.
- Create a shared `Guard` in `app.NewServer` after opening SQLite.
- Wrap `*sql.DB` in `MeteredDB` and pass it to the existing upstream `database` and `db` modules.
- Register a new local native module `db.guard` for configuration, stats, manual checks, and callback registration.
- Implement soft-limit cleanup only.
- Keep the original database module unchanged.
- Document app-owned cleanup policies with session-aware examples.

This gives us the behavior we want without forking or editing `go-go-goja/modules/database`.

## Open questions

- Should the guard be enabled by default with a large limit, or disabled until configured by JavaScript?
- Should the CLI expose `--db-max-bytes`, or should this live entirely in app scripts through `db.guard.configure(...)`?
- Should cleanup callbacks be allowed to fail the original write, or should they only log/report in v1?
- Should the session implementation grow a standard `app_sessions` helper table, or should cleanup examples remain app-owned?
- Should the guard run `PRAGMA wal_checkpoint(PASSIVE)` before measuring, or should it report WAL growth exactly as-is?

## Final recommendation

Do not modify the upstream database module yet. Use the extension point it already provides: `WithPreconfiguredDB` accepts a `QueryExecer`, so `goja-site` can pass a metered wrapper. Add a separate `db.guard` module to register cleanup callbacks and expose stats. This keeps the existing JavaScript database API stable while adding storage policy around it.

If the wrapper later proves insufficient, write a local app database module in `pkg/sitedb` or `pkg/appdb` and explicitly document why the fork is necessary. But the first implementation should prefer composition over replacement.
