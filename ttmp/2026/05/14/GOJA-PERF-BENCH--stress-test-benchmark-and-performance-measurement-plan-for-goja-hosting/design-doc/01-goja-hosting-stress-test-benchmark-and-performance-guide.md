---
Title: Goja Hosting Stress Test Benchmark and Performance Guide
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
    - Path: pkg/app/database.go
      Note: Database module and policy wiring for simple versus guarded benchmark scenarios
    - Path: pkg/app/multi_config.go
      Note: Multi-site config schema and normalization needed for generated benchmark configs
    - Path: pkg/app/multi_server.go
      Note: Multi-site Host-header dispatch and per-site Server ownership for scaling benchmarks
    - Path: pkg/app/server.go
      Note: Single-site runtime
    - Path: pkg/dbguard/metered.go
      Note: Guarded SQLite write wrapper whose overhead should be benchmarked
    - Path: pkg/kanbanddsl/mount.go
      Note: Kanban fragment and action request paths for realistic interaction benchmarks
    - Path: scripts/playwright-kanban-smoke.sh
      Note: Existing end-to-end process lifecycle and cleanup pattern for external benchmark scripts
ExternalSources: []
Summary: Design and implementation guide for benchmarking the complete goja-site hosting concept.
LastUpdated: 2026-05-14T13:37:52.680314329-04:00
WhatFor: Use this to onboard an intern who will build, run, and maintain stress tests for goja-site.
WhenToUse: Before adding benchmark harnesses, pprof instrumentation, load-test scripts, or performance gates.
---








# Goja Hosting Stress Test Benchmark and Performance Guide

## Executive Summary

`goja-site` hosts small trusted JavaScript websites inside a Go process. The current system creates one Goja runtime per site, exposes Go-owned modules to JavaScript, opens a SQLite database per site, loads JavaScript files at startup, and dispatches HTTP requests through an Express-style router. The multi-site mode creates several independent `Server` instances and dispatches by normalized `Host` header.

This document is a design and implementation guide for a benchmark and stress-test program that measures the whole hosting concept rather than one isolated function. The goal is to answer practical questions:

- How much traffic can a single `goja-site serve` instance handle for typical pages?
- How much extra overhead comes from multi-site host dispatch?
- Which layer dominates latency: Go HTTP routing, Goja JavaScript execution, UI DSL rendering, SQLite, the Kanban DSL, or browser-side refresh behavior?
- Does throughput degrade linearly as the number of hosted sites grows?
- Do memory, goroutine count, file descriptors, database size checks, or Goja runtime queues grow unexpectedly under load?
- Which metrics should become regression gates before the hosting model is used more broadly?

The recommended approach is a layered benchmark suite:

1. **In-process Go benchmarks** for fast, repeatable micro and subsystem measurements.
2. **Black-box HTTP load tests** against `goja-site serve` and `serve-multi` to measure realistic end-to-end behavior.
3. **Stateful stress scenarios** that create, move, search, and render Kanban-style records through JavaScript and SQLite.
4. **Observability hooks** for CPU profiles, heap profiles, goroutine profiles, database counters, and per-route latency histograms.
5. **Result normalization** so benchmark reports include commit hash, Go version, OS, CPU model, site count, database policy, request mix, and warm-up rules.

The intern implementing this should first add measurement scaffolding without changing product behavior. Only after stable baseline numbers exist should the team optimize code paths.

## Problem Statement and Scope

### Problem

The current repository has functionality tests and a browser smoke test, but it does not yet contain a repeatable performance program. There is no baseline for request latency, startup time, memory footprint per site, SQLite overhead, `db.guard` overhead, Goja execution overhead, or Kanban action refresh cost. Without those numbers, performance changes are anecdotal and regressions are hard to detect.

### Scope

This ticket covers the design of a complete performance measurement effort for the existing `goja-site` concept:

- Single-site server startup and HTTP request handling.
- Multi-site server startup and Host-header dispatch.
- Goja runtime creation and script loading.
- JavaScript route callbacks invoked by the Go-owned Express-style host.
- HTML rendering through `ui.dsl` and `uidsl.RenderAny`.
- SQLite reads and writes through the `database`/`db` modules.
- Guarded database policy and `db.guard` checks.
- Kanban DSL fragment/action routes and optional HTML refresh.
- Browser smoke coverage as a correctness companion, not as the main benchmark engine.

### Out of scope for the first implementation

- Distributed load testing across many machines.
- Browser rendering performance across multiple real devices.
- Production Kubernetes autoscaling design.
- Security hardening of trusted JavaScript execution.
- Rewriting the host architecture before measurements exist.

## System Orientation for a New Intern

### What `goja-site` is

`goja-site` is a Go CLI and hosting runtime for trusted server-side JavaScript websites. A site author writes `.js` files that `require("express")`, `require("ui.dsl")`, `require("database")`, and optionally `require("kanban.dsl")`. The Go process loads those scripts once at startup. The scripts register route handlers. Later HTTP requests call those handlers.

This is not a Node.js server. Go owns the process, HTTP listener, SQLite handle, JavaScript runtime lifecycle, module registry, and response rendering. JavaScript owns application-specific routes, SQL queries, view functions, Kanban callbacks, and domain logic.

### Key repository entrypoints

- `cmd/goja-site/main.go`: creates the Cobra root command, attaches help docs, registers `serve`, `serve-multi`, and JavaScript verb commands.
- `cmd/goja-site/serve.go`: CLI command for one site.
- `cmd/goja-site/serve_multi.go`: CLI command for many sites under one listener.
- `pkg/app/server.go`: owns one hosted site runtime.
- `pkg/app/multi_server.go`: owns the Host-header dispatch map for many sites.
- `pkg/app/multi_config.go`: validates and normalizes multi-site YAML/JSON config.
- `pkg/app/database.go`: connects Go's SQLite handle to JavaScript modules.
- `pkg/dbguard/*`: guarded database policy, size checks, and callback support.
- `pkg/kanbanddsl/*`: Go-owned Kanban DSL, rendering, actions, and browser runtime.
- `sites/*/scripts`: realistic JavaScript applications used for load scenarios.
- `scripts/playwright-kanban-smoke.sh`: existing end-to-end correctness smoke test.

### Runtime architecture diagram

```text
             one process: goja-site

  +-------------------------------------------------------+
  | CLI command                                           |
  |   serve or serve-multi                                |
  +-------------------------+-----------------------------+
                            |
                            v
  +-------------------------------------------------------+
  | net/http.Server                                       |
  |  - ReadHeaderTimeout                                  |
  |  - Handler: Server host or MultiServer dispatcher     |
  +-------------------------+-----------------------------+
                            |
             single site    |     multi-site mode
                            v
  +-------------------------------------------------------+
  | MultiServer                                           |
  |  - normalize Host                                    |
  |  - map host -> *Server                               |
  |  - /healthz and /readyz handled outside sites         |
  +-------------------------+-----------------------------+
                            |
                            v
  +-------------------------------------------------------+
  | Server                                                |
  |  - *sql.DB                                           |
  |  - *engine.Runtime / Goja runtime owner              |
  |  - gojahttp.Host route registry                      |
  |  - module registrars: express, ui.dsl, kanban.dsl    |
  |  - module specs: database and db alias               |
  +-------------------------+-----------------------------+
                            |
                            v
  +-------------------------------------------------------+
  | JavaScript site scripts                               |
  |  - require modules                                    |
  |  - create schema                                      |
  |  - register routes                                    |
  |  - render HTML / JSON                                 |
  |  - run SQL                                            |
  +-------------------------------------------------------+
```

### Request path diagram

```text
HTTP client
  |
  | GET / or POST /_kanban/<board>/action/:action
  v
net/http.Server
  |
  | single site: Handler = gojahttp.Host
  | multi site: Handler = MultiServer, then site.ServeHTTP
  v
gojahttp.Host route match
  |
  | calls JavaScript handler through runtime owner
  v
JavaScript callback
  |
  +--> db.query/db.exec -> Go database module -> *sql.DB -> SQLite file
  |
  +--> ui.dsl nodes -> Go renderer -> HTML string
  |
  +--> kanban.dsl Dispatch/Render -> callbacks + UI DSL renderer
  v
Response writer
```

## Current-State Analysis with Evidence

### Single-site server lifecycle

`pkg/app/server.go` defines `Server` as the unit that owns the database, runtime, route host, and optional HTTP server. The struct fields are declared at `pkg/app/server.go:22-28`. `NewServer` applies defaults for address, database path, and script directory at `pkg/app/server.go:31-40`, normalizes database policy at `pkg/app/server.go:41-43`, opens and pings SQLite at `pkg/app/server.go:45-55`, creates a `gojahttp.Host` at `pkg/app/server.go:57`, builds module configuration at `pkg/app/server.go:58-64`, builds the Goja runtime factory at `pkg/app/server.go:66-74`, creates one runtime at `pkg/app/server.go:76-81`, then loads scripts at `pkg/app/server.go:83-87`.

Performance implication: startup time includes filesystem directory creation, SQLite open/ping, module registration, runtime construction, and script execution. Request latency includes `gojahttp.Host` dispatch plus JavaScript callback execution. The code does not currently expose per-phase timings, so the first benchmark implementation should add measurements around these lifecycle steps without changing semantics.

### Script loading

`LoadScripts` resolves JavaScript files, reads each file, and calls `vm.RunScript` through `s.runtime.Owner.Call` at `pkg/app/server.go:141-158`. Scripts are loaded once per `Server`. For multi-site, this cost repeats once per site.

Performance implication: startup benchmarks must separate:

- script discovery time,
- file read time,
- Goja compile/evaluate time,
- application initialization SQL time,
- route registration time.

### HTTP serving and shutdown

Single-site serving constructs an `http.Server` with `Handler: s.host` and `ReadHeaderTimeout: 5 * time.Second` at `pkg/app/server.go:99-100`. The server runs `ListenAndServe` in a goroutine and shuts down on context cancellation at `pkg/app/server.go:101-118`.

Performance implication: black-box load tests should use real `go run ./cmd/goja-site serve` or a compiled binary, not only `httptest`, because the CLI path and listener path are part of the product. However, in-process benchmarks are still valuable for regression speed.

### Multi-site lifecycle and dispatch

`pkg/app/multi_server.go` defines `MultiServer` with a map from host to `*Server` at `pkg/app/multi_server.go:13-16`. `NewMultiServer` normalizes config and creates one `Server` per site at `pkg/app/multi_server.go:19-31`. Request handling special-cases `/healthz` and `/readyz`, normalizes `r.Host`, looks up the site in a map, returns 404 for unknown hosts, then calls `site.ServeHTTP` at `pkg/app/multi_server.go:57-69`.

Performance implication: Host dispatch itself should be cheap, but total memory and startup cost grow with the number of sites because each site has its own Goja runtime and SQLite handle. Benchmarks should report per-site incremental memory after `N = 1, 4, 16, 64` sites.

### Multi-site configuration

`pkg/app/multi_config.go` describes the config schema at `pkg/app/multi_config.go:14-33`. `LoadMultiConfig` reads YAML or JSON and normalizes it at `pkg/app/multi_config.go:36-54`. `Normalize` supplies defaults, validates site names, derives hosts and DB paths, rejects duplicates, and normalizes database policies at `pkg/app/multi_config.go:57-108`. Host normalization lowercases, trims dots, and strips a normal `host:port` suffix at `pkg/app/multi_config.go:115-121`.

Performance implication: multi-site benchmarks need reproducible generated configs. The benchmark harness should create temporary configs with deterministic site names, script directories, and DB paths rather than relying only on checked-in `deploy/sites.local.yaml`.

### Database module and policy surface

`pkg/app/database.go` builds the JavaScript database modules. For `DBPolicySimple`, it wraps `*sql.DB` with `simpleDB`; for `DBPolicyGuarded`, it creates a `dbguard.Guard`, wraps the database with `dbguard.NewMeteredDB`, and registers the `db.guard` module at `pkg/app/database.go:18-35`. It exposes both `database` and `db` module names at `pkg/app/database.go:37-53`. `simpleDB.Query` rejects writes unless allowed at `pkg/app/database.go:61-69`, and `simpleDB.Exec` rejects writes unless allowed at `pkg/app/database.go:71-78`.

Performance implication: the benchmark suite must compare `guarded`, `simple --readonly`, and `simple --allow-writes`. Guarded mode is the default and includes database size checks after writes through `MeteredDB.Exec`.

### Guarded database execution

`pkg/dbguard/metered.go` forwards `Query` directly and wraps `Exec` with `BeforeExec`, the real SQL execution, `AfterExec`, and `ErrorAfterExec`. The current file shows this at `pkg/dbguard/metered.go:16-33`.

Performance implication: write-heavy scenarios need a guard-overhead benchmark. A benchmark should measure plain SQLite, simple policy, and guarded policy under the same SQL mix.

### Kanban mount and action routes

`pkg/kanbanddsl/mount.go` registers a client runtime route, a fragment refresh route, and a POST action route. `Board.Mount` validates the express app and normalizes the prefix at `pkg/kanbanddsl/mount.go:11-16`. It registers `GET <prefix>/client.js` once per prefix at `pkg/kanbanddsl/mount.go:23-33`. It registers the fragment route and calls `b.Render` at `pkg/kanbanddsl/mount.go:34-44`. It registers the action route, extracts `:action`, adds the session to the action body, dispatches to JavaScript callbacks, and optionally re-renders HTML into the JSON response at `pkg/kanbanddsl/mount.go:45-88`.

Performance implication: Kanban actions are important stress cases because one POST may do all of the following: JavaScript dispatch, SQLite write, JavaScript read, UI node creation, Go UI render, and JSON response. Benchmarks should measure action paths with `refresh=true` and `refresh=false` separately.

### Existing correctness coverage

The current tests include multi-site host routing and database isolation in `pkg/app/multi_server_test.go`, including a JavaScript route that writes to SQLite and returns visit counts. This is an excellent seed for in-process benchmarks because it already creates temporary script directories and `httptest` requests. The repository also has a Playwright smoke script at `scripts/playwright-kanban-smoke.sh`; it starts `goja-site`, waits for readiness with curl, runs browser automation, creates a card, searches, moves the card, and checks for console errors.

Current validation command run during this investigation:

```text
go test ./...
?   	github.com/go-go-golems/goja-site/cmd/goja-site	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/app	0.051s
ok  	github.com/go-go-golems/goja-site/pkg/dbguard	0.043s
?   	github.com/go-go-golems/goja-site/pkg/doc	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/kanbanddsl	0.014s
```

## What Must Be Measured

### Measurement matrix

| Area | Why it matters | Primary metric | Secondary metrics |
|---|---|---|---|
| Startup single-site | Affects deployment, restarts, local dev | time to ready | allocations, loaded script count, DB open time |
| Startup multi-site | Tests concept scalability | time to ready vs site count | RSS, heap, goroutines, DB handles |
| Simple GET route | Baseline host overhead | requests/sec, p50/p95/p99 latency | CPU profile, allocs/op |
| Rendered HTML route | UI DSL overhead | latency by HTML size | bytes rendered, allocs/op |
| DB read route | SQLite query path | latency per query count | DB time, rows read |
| DB write route | write path and guard | writes/sec, p99 latency | DB size check time, lock waits |
| Kanban fragment | server-rendered board cost | latency by card count | render time, HTML size |
| Kanban action refresh | most realistic interaction | action latency | dispatch time, re-render time, JSON size |
| Multi-host mix | one process hosting many sites | aggregate throughput | fairness per host, tail latency |
| Soak test | resource leaks | stable RSS/goroutines over time | GC pause, heap objects, DB file size |

### Required benchmark scenarios

#### Scenario A: Null route

A JavaScript route returns a constant string. This isolates HTTP dispatch and Goja callback overhead.

```javascript
const express = require("express");
const app = express.app();
app.get("/", (req, res) => res.type("text/plain").send("ok"));
```

Expected use: baseline all other routes against this.

#### Scenario B: UI render route

A JavaScript route creates a fixed-size `ui.dsl` tree and sends HTML. Vary node count: 10, 100, 1000.

```javascript
const express = require("express");
const ui = require("ui.dsl");
const app = express.app();

function page(n) {
  const rows = [];
  for (let i = 0; i < n; i++) {
    rows.push(ui.li({ class: "row" }, "row-" + i));
  }
  return ui.html(ui.body(ui.main(ui.h1("bench"), ui.ul(rows))));
}

app.get("/render", (req, res) => res.html(page(Number(req.query.n || 100))));
```

Expected use: measure UI DSL and render path independently of database work.

#### Scenario C: SQLite read route

A route selects rows by indexed session or key. Vary row count and query count per request.

```javascript
const db = require("database");
const express = require("express");
const app = express.app();

db.exec("CREATE TABLE IF NOT EXISTS items(id INTEGER PRIMARY KEY, k TEXT, v TEXT)");
db.exec("CREATE INDEX IF NOT EXISTS idx_items_k ON items(k)");

app.get("/read", (req, res) => {
  const n = Number(req.query.n || 10);
  let total = 0;
  for (let i = 0; i < n; i++) {
    total += db.query("SELECT COUNT(*) AS c FROM items WHERE k = ?", "hot")[0].c;
  }
  res.json({ ok: true, total });
});
```

Expected use: measure database module conversion overhead and SQLite read overhead.

#### Scenario D: SQLite write route

A route inserts records. Run under guarded and simple policies.

```javascript
app.post("/write", (req, res) => {
  const n = Number(req.body.n || 1);
  for (let i = 0; i < n; i++) {
    db.exec("INSERT INTO items(k, v) VALUES (?, ?)", "hot", "value-" + Date.now() + "-" + i);
  }
  res.json({ ok: true, inserted: n });
});
```

Expected use: quantify `db.guard` write overhead and SQLite locking behavior.

#### Scenario E: Kanban action route

Use a minimal board with `move` and `create` actions, then benchmark:

- `GET /_kanban/<board>/fragment`
- `POST /_kanban/<board>/action/move` with `refresh=false`
- `POST /_kanban/<board>/action/move` with default refresh
- `GET /` full page render containing board HTML

Expected use: represent the real product concept where JavaScript, SQLite, DSL rendering, and JSON response assembly happen together.

#### Scenario F: Multi-site host mix

Generate `N` sites with identical scripts and separate DB files. Run one load generator that sends Host headers across all sites.

```text
for request_index in 0..total_requests:
  host = hosts[request_index % len(hosts)]
  path = weighted_choice(["/", "/render?n=100", "/read?n=3", "/write"])
  send HTTP request with Host: host
```

Expected use: measure fairness, aggregate throughput, and per-site isolation.

## Proposed Benchmark Architecture

### File layout

Add the following files in phases:

```text
bench/
  scripts/
    null-route/app.js
    render-route/app.js
    db-read-write/app.js
    kanban-minimal/app.js
  configs/
    README.md
  results/
    .gitignore
pkg/app/
  server_bench_test.go
  multi_server_bench_test.go
pkg/dbguard/
  metered_bench_test.go
scripts/
  bench-http.sh
  bench-wrk.lua
  bench-hey.sh
  bench-report.py
```

The benchmark code should use temporary DB files and generated configs. Do not commit large benchmark outputs. Commit only small sample reports if needed.

### In-process Go benchmarks

In-process benchmarks should use `testing.B` and `httptest`. They are fast enough for local iteration and CI smoke benchmarking. They should not replace external load tests because they skip kernel socket behavior and the CLI path.

Pseudocode:

```go
func BenchmarkServerNullRoute(b *testing.B) {
    root := b.TempDir()
    scripts := writeBenchScript(b, root, nullRouteJS)
    srv, err := NewServer(Config{
        DBPath: filepath.Join(root, "app.db"),
        ScriptDirs: []string{scripts},
        DBPolicy: DBPolicySimple,
        AllowWrites: true,
    })
    if err != nil { b.Fatal(err) }
    defer srv.Close(context.Background())

    req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rr := httptest.NewRecorder()
        srv.ServeHTTP(rr, req)
        if rr.Code != http.StatusOK { b.Fatalf("status=%d", rr.Code) }
    }
}
```

Recommended commands:

```bash
go test ./pkg/app -run '^$' -bench 'BenchmarkServer' -benchmem -count=10 \
  | tee bench/results/app-server-$(git rev-parse --short HEAD).txt

go test ./pkg/dbguard -run '^$' -bench . -benchmem -count=10 \
  | tee bench/results/dbguard-$(git rev-parse --short HEAD).txt
```

### Startup benchmarks

Use Go benchmarks for `NewServer` and `NewMultiServer`, but be careful: startup benchmarks can be dominated by filesystem and SQLite initialization. That is useful, but must be labeled.

Pseudocode:

```go
func BenchmarkNewMultiServerSites(b *testing.B) {
    for _, n := range []int{1, 4, 16, 64} {
        b.Run(fmt.Sprintf("sites=%d", n), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                root := b.TempDir()
                cfg := generatedMultiConfig(root, n)
                srv, err := NewMultiServer(cfg)
                if err != nil { b.Fatal(err) }
                _ = srv.Close(context.Background())
            }
        })
    }
}
```

Record:

- elapsed startup time,
- heap allocated after startup,
- goroutine count after startup,
- number of open DB files,
- scripts loaded per site.

### External HTTP load tests

Use a compiled binary for serious runs:

```bash
go build -o ./tmp/goja-site-bench ./cmd/goja-site
```

Then run scenarios through either `hey`, `wrk`, `vegeta`, or `oha`. The exact tool matters less than repeatability, but use one primary tool in CI and document its version.

Example with `hey`:

```bash
./tmp/goja-site-bench serve \
  --db ./tmp/bench/null.db \
  --scripts ./bench/scripts/null-route \
  --db-policy simple \
  --allow-writes \
  --addr :18080 > ./tmp/bench/server.log 2>&1 &
SERVER_PID=$!

hey -z 30s -c 64 http://127.0.0.1:18080/ \
  | tee bench/results/null-route-hey.txt

kill "$SERVER_PID"
```

Example with Host headers for multi-site:

```bash
hey -z 60s -c 128 \
  -H 'Host: site-000.bench.example.test' \
  http://127.0.0.1:18080/render?n=100
```

For mixed hosts, use a small Go or Python load generator rather than one `hey` command, because `hey` uses one header value per run.

### Pprof and runtime metrics

Add optional profiling under a flag, not always-on public endpoints. Recommended CLI flags:

```text
--pprof-addr string       optional address for net/http/pprof, disabled by default
--metrics-addr string     optional address for expvar/Prometheus-style metrics, disabled by default
--trace-startup           log startup phase durations
```

Implementation sketch:

```go
// in cmd settings
PprofAddr string `glazed:"pprof-addr"`

// after server creation, before Run
if settings.PprofAddr != "" {
    go func() {
        mux := http.NewServeMux()
        // import net/http/pprof for registration or mount handlers manually
        log.Printf("pprof listening on %s", settings.PprofAddr)
        _ = http.ListenAndServe(settings.PprofAddr, mux)
    }()
}
```

Recommended pprof capture commands:

```bash
go tool pprof -http=:0 http://127.0.0.1:18081/debug/pprof/profile?seconds=30
go tool pprof -http=:0 http://127.0.0.1:18081/debug/pprof/heap
go tool pprof -http=:0 http://127.0.0.1:18081/debug/pprof/goroutine
```

Important: do not expose pprof on the public site listener in production. Bind it to `127.0.0.1` or a private diagnostics listener.

## Metrics and Report Format

Every benchmark result should include enough metadata to be comparable later:

```yaml
benchmark_id: goja-null-route-http
commit: abc1234
branch: main
dirty_worktree: false
go_version: go1.26.2
os: linux
arch: amd64
cpu_model: "..."
ram_gb: 32
scenario: null-route
mode: serve
site_count: 1
db_policy: simple-allow-writes
script_set: bench/scripts/null-route
warmup_seconds: 5
duration_seconds: 60
concurrency: 64
load_tool: hey 0.1.4
results:
  requests_per_second: 0
  latency_ms_p50: 0
  latency_ms_p95: 0
  latency_ms_p99: 0
  errors: 0
  bytes_per_response_p50: 0
runtime:
  rss_mb_start: 0
  rss_mb_end: 0
  heap_mb_end: 0
  goroutines_end: 0
  gc_pause_p99_ms: 0
artifacts:
  stdout: bench/results/...
  cpu_profile: bench/results/...
  heap_profile: bench/results/...
```

The first implementation may write Markdown plus raw tool output. A later implementation can emit JSON and generate charts.

## API and CLI References for the Intern

### Existing CLI: single-site serve

Source: `cmd/goja-site/serve.go`.

Current flags include:

- `--addr`: HTTP bind address.
- `--db`: SQLite database path.
- `--scripts`: repeatable script directory list.
- `--db-policy`: `guarded` or `simple`.
- `--readonly`: disable writes for simple policy.
- `--allow-writes`: allow writes for simple policy.
- `--dev`: show detailed development errors in HTTP responses.

The command creates `app.NewServer` and runs it until SIGINT or SIGTERM.

### Existing CLI: multi-site serve

Source: `cmd/goja-site/serve_multi.go`.

Current flag:

- `--config`: YAML or JSON multi-site config path.

The command loads config, creates `app.NewMultiServer`, prints host summary, and runs one listener.

### Existing JavaScript modules

The help reference in `pkg/doc/reference/js-api-reference.md` documents the modules available to site scripts:

- `require("express")`: register HTTP routes and static file serving.
- `require("ui.dsl")`: construct renderable HTML nodes.
- `require("kanban.dsl")`: build and mount interactive Kanban boards.
- `require("database")` and `require("db")`: query and execute SQLite statements.
- `require("db.guard")`: configure guarded policy and inspect database size status.
- Trusted utility modules from go-go-goja such as `fs`, `path`, `time`, `timer`, and `yaml`.

### Existing test helpers to reuse

`pkg/app/multi_server_test.go` contains `writeSiteScript(t, dir, body string) string`, which creates a script directory and writes `app.js`. Consider moving or duplicating a benchmark-specific helper for `testing.B`:

```go
func writeBenchScript(tb testing.TB, dir, body string) string {
    tb.Helper()
    scripts := filepath.Join(dir, "scripts")
    if err := os.MkdirAll(scripts, 0o755); err != nil { tb.Fatal(err) }
    if err := os.WriteFile(filepath.Join(scripts, "app.js"), []byte(body), 0o644); err != nil { tb.Fatal(err) }
    return scripts
}
```

## Implementation Plan

### Phase 1: Baseline benchmark fixtures

Add deterministic benchmark scripts under `bench/scripts`. Start with null route, render route, DB read/write, and minimal Kanban. Keep them intentionally small and documented.

Acceptance criteria:

- `go test ./...` still passes.
- `go run ./cmd/goja-site serve --scripts bench/scripts/null-route ...` starts and serves `/`.
- Each script has a short README explaining the scenario.

### Phase 2: In-process benchmarks

Add `pkg/app/server_bench_test.go` and `pkg/app/multi_server_bench_test.go`.

Benchmarks to implement first:

- `BenchmarkNewServerNullRoute`
- `BenchmarkServerNullRouteGET`
- `BenchmarkServerRenderGET/n=10,100,1000`
- `BenchmarkServerDBRead/q=1,10,100`
- `BenchmarkServerDBWrite/policy=simple,guarded`
- `BenchmarkNewMultiServer/sites=1,4,16`
- `BenchmarkMultiServerHostDispatch/sites=1,4,16`

Acceptance criteria:

```bash
go test ./pkg/app -run '^$' -bench . -benchmem
```

prints stable benchmark rows and no benchmark writes outside temporary directories.

### Phase 3: DB guard microbenchmarks

Add `pkg/dbguard/metered_bench_test.go` to measure `MeteredDB.Exec` overhead against plain `*sql.DB` and `simpleDB` where practical.

Acceptance criteria:

```bash
go test ./pkg/dbguard -run '^$' -bench . -benchmem
```

reports insert and update costs with and without guard checks.

### Phase 4: External load scripts

Add shell scripts in `scripts/` that build the binary, start it, wait for readiness, run load, collect logs, and clean up. Model the cleanup style after `scripts/playwright-kanban-smoke.sh`, which already handles temporary DB files, log paths, process cleanup, and failure log printing.

Acceptance criteria:

```bash
scripts/bench-http.sh --scenario null --duration 30s --concurrency 64
scripts/bench-http.sh --scenario kanban --duration 30s --concurrency 32
scripts/bench-http.sh --scenario multi --sites 4 --duration 60s --concurrency 128
```

Each command writes a report and raw output under `bench/results/` or a configurable output directory.

### Phase 5: Optional pprof and metrics listener

Add disabled-by-default diagnostics flags. Keep diagnostics listener separate from the public site listener. Capture CPU and heap profiles during external load tests.

Acceptance criteria:

- `goja-site serve --pprof-addr 127.0.0.1:18081 ...` exposes pprof only on that address.
- Existing CLI behavior is unchanged when the flag is absent.
- Benchmark scripts can capture CPU and heap profiles if the flag is supplied.

### Phase 6: Result summarizer

Add a small report generator. Python is acceptable if it only parses committed raw output formats and emits Markdown/JSON. A Go command is better if the project wants single-language tooling.

Acceptance criteria:

- Reports include metadata listed in this document.
- Reports are small enough for code review.
- Raw large artifacts are ignored by git.

### Phase 7: CI smoke benchmark and nightly benchmark

Do not run long stress tests on every PR. Add:

- A fast PR job: in-process benchmarks with `-benchtime=100ms` only to catch panics and gross regressions.
- A scheduled nightly job: longer `-count=10` Go benchmarks and external HTTP benchmark on a pinned runner if available.

Acceptance criteria:

- PR job is under a few minutes.
- Nightly artifacts preserve raw output and generated summaries.
- Regression thresholds are initially advisory until several baselines exist.

## Validation Strategy

### Correctness gates before performance runs

Always run:

```bash
go test ./...
```

If benchmarking Kanban/browser behavior, also run:

```bash
scripts/playwright-kanban-smoke.sh
```

### Benchmark repeatability rules

- Build once, then benchmark the compiled binary.
- Record `git status --short`; mark dirty worktrees clearly.
- Run a warm-up period before measuring.
- Prefer fixed CPU governor/performance mode for serious numbers.
- Avoid running heavy desktop/browser tasks during local measurements.
- Keep temporary DBs on the same filesystem across comparisons.
- Run each benchmark multiple times and compare distributions, not single best runs.

### Regression gates after baselines exist

Possible first gates:

- Null route p95 latency should not regress by more than 15% against the rolling baseline.
- Multi-site startup time per site should not regress by more than 20%.
- Guarded write route p95 should not regress by more than 20%.
- Heap after 10-minute soak should not grow monotonically beyond a fixed leak threshold.

Treat these gates as advisory until the team has enough history to distinguish noise from real regressions.

## Risks, Alternatives, and Open Questions

### Risks

- **Benchmarking the wrong thing:** In-process benchmarks can hide socket, kernel, and CLI overhead. Mitigation: always pair them with black-box HTTP load tests.
- **Noisy local results:** Developer laptops are not stable benchmark machines. Mitigation: preserve metadata and use repeated runs.
- **Optimizing before measuring:** The system has several possible bottlenecks. Mitigation: add pprof capture before making performance changes.
- **Exposing diagnostics publicly:** pprof endpoints can leak sensitive information. Mitigation: disabled by default and bind to private addresses.
- **Database file growth changing results:** Write-heavy tests can slow down as DB files grow. Mitigation: reset DBs per run and separately design soak tests for growth behavior.
- **JavaScript scenario drift:** Benchmark scripts can become unrealistic. Mitigation: include both synthetic and real-site scenarios from `sites/trail`, `sites/crm`, `sites/editorial`, and `sites/pizza`.

### Alternatives considered

1. **Only use `go test -bench`.** Rejected as the only strategy because it misses real HTTP listener behavior and load generator pressure.
2. **Only use browser automation.** Rejected because browser tests are slower, noisier, and primarily correctness-oriented.
3. **Immediately add Prometheus everywhere.** Deferred because the first need is stable measurements; a small pprof/expvar listener is enough initially.
4. **Benchmark production sites only.** Rejected because synthetic scenarios are needed to isolate layers and explain regressions.

### Open questions

- Should benchmark artifacts live in this repository, a separate performance repository, or CI artifact storage only?
- Which external load tool should be standardized: `hey`, `wrk`, `vegeta`, or `oha`?
- Which deployment shape is the main target: one large multi-site process or many single-site processes?
- What production-like hardware should define baseline performance?
- Should `gojahttp.Host` expose per-route metrics, or should route timing be wrapped at the `Server` boundary?
- Should SQLite pragmas be standardized for performance tests, or should tests use default settings to match current behavior?

## Intern Checklist

Use this checklist when implementing the benchmark harness:

1. Read `pkg/app/server.go`, `pkg/app/multi_server.go`, `pkg/app/database.go`, and `pkg/kanbanddsl/mount.go`.
2. Run `go test ./...` and save the output in the diary.
3. Add benchmark scripts under `bench/scripts`.
4. Add one in-process benchmark at a time; verify it uses `b.TempDir()`.
5. Add external benchmark shell script with cleanup traps modeled after `scripts/playwright-kanban-smoke.sh`.
6. Add optional pprof listener behind a disabled-by-default flag.
7. Run short benchmarks locally and record metadata.
8. Write down suspicious results before changing code.
9. Only optimize after CPU/heap profiles identify a bottleneck.
10. Update this ticket with measured baseline numbers and attach reports.

## References

- `cmd/goja-site/main.go`: CLI root, help wiring, command registration.
- `cmd/goja-site/serve.go`: single-site CLI flags and `app.NewServer` invocation.
- `cmd/goja-site/serve_multi.go`: multi-site CLI flags and `app.NewMultiServer` invocation.
- `pkg/app/server.go`: single-site server, Goja runtime, module registration, script loading, HTTP serving.
- `pkg/app/multi_server.go`: Host-header dispatch and health endpoints.
- `pkg/app/multi_config.go`: multi-site config schema and normalization.
- `pkg/app/database.go`: `database` and `db` module creation, simple and guarded policies.
- `pkg/dbguard/metered.go`: guarded execution wrapper around SQLite writes.
- `pkg/kanbanddsl/mount.go`: Kanban client, fragment, and action routes.
- `pkg/doc/reference/js-api-reference.md`: JavaScript API reference for site authors.
- `pkg/doc/topics/user-guide.md`: conceptual guide for site lifecycle and multi-site mode.
- `scripts/playwright-kanban-smoke.sh`: existing process lifecycle and end-to-end smoke test pattern.
- `deploy/sites.local.yaml`: checked-in multi-site local config with trail, editorial, crm, and pizza sites.
