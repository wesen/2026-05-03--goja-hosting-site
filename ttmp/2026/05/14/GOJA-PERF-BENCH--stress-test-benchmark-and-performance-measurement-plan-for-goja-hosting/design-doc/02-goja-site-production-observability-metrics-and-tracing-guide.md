---
Title: Goja Site Production Observability Metrics and Tracing Guide
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
    - Path: cmd/goja-site/main.go
      Note: CLI root and Glazed logging setup that observability flags must coexist with
    - Path: cmd/goja-site/serve.go
      Note: Single-site command flags and run path where diagnostics and tracing config should be wired
    - Path: cmd/goja-site/serve_multi.go
      Note: Multi-site command flags and run path where process-wide diagnostics should be wired
    - Path: pkg/app/database.go
      Note: Database QueryExecer construction point for SQL instrumentation
    - Path: pkg/app/multi_server.go
      Note: Host dispatch boundary for per-site metrics and unknown-host counters
    - Path: pkg/app/server.go
      Note: Single-site lifecycle and HTTP handler boundary for metrics and tracing
    - Path: pkg/dbguard/guard.go
      Note: Guard stats and limit measurement internals for DB size metrics
    - Path: pkg/dbguard/metered.go
      Note: Guarded exec boundary for write and guard overhead metrics
    - Path: pkg/kanbanddsl/mount.go
      Note: Kanban fragment/action/render path for domain-specific metrics and spans
ExternalSources: []
Summary: Analysis and implementation guide for Prometheus metrics, pprof diagnostics, and OpenTelemetry tracing in goja-site.
LastUpdated: 2026-05-14T14:20:00-04:00
WhatFor: Use this to implement production-grade observability for goja-site while also supporting benchmark analysis.
WhenToUse: Before adding metrics, tracing, diagnostics listeners, benchmark instrumentation, or production monitoring dashboards.
---










# Goja Site Production Observability Metrics and Tracing Guide

## Executive Summary

Yes: `goja-site` should have first-class production observability, not just ad hoc benchmark logging. The hosting model has several layers that can each become the bottleneck: Host-header dispatch, the Go HTTP server, Goja runtime calls, JavaScript route handlers, UI DSL rendering, SQLite queries and writes, `db.guard` checks, and Kanban DSL action refreshes. A Prometheus metrics endpoint is the right first production monitoring surface because it is cheap, familiar, pull-based, and useful for both stress testing and live operation.

Tracing is also useful, but it should be added through OpenTelemetry rather than Jaeger-specific APIs. Jaeger can be one backend behind an OpenTelemetry Collector. The right sequence is:

1. **Prometheus metrics** on a separate diagnostics listener.
2. **pprof diagnostics** on the same private diagnostics listener or a sibling listener.
3. **OpenTelemetry tracing** with sampling and OTLP export.
4. **Benchmark overhead measurements** comparing metrics/tracing disabled, metrics enabled, and tracing sampled.
5. **Production dashboards and alerts** once the metric names stabilize.

The design goal is to make `goja-site` operable as a multi-site production process while preserving benchmark correctness. Instrumentation must be low-cardinality, disabled or privately bound by default where sensitive, and measured for overhead. Do not label metrics with raw URLs, raw Host headers from unknown traffic, SQL text, session IDs, user IDs, request bodies, or JavaScript error strings.

## Problem Statement and Scope

### Problem

The current benchmark guide defines how to measure `goja-site`, but a benchmark harness alone does not make the system operable in production. Production hosting needs answers to questions such as:

- Which hosted site is slow?
- Which route or Kanban action is slow?
- Is the latency coming from JavaScript, UI rendering, SQLite, `db.guard`, or Host dispatch?
- Are Goja handler errors increasing?
- Are unknown Host-header requests increasing?
- Is memory growing per site?
- Are database files near guard limits?
- Does instrumentation itself change benchmark results?

Without structured metrics and traces, operators must infer these answers from logs, local reproduction, or one-off profiling.

### Scope

This guide covers a production-grade observability implementation for the current Go codebase:

- Prometheus metrics endpoint.
- Standard Go runtime and process collectors.
- HTTP request counters, duration histograms, response sizes, and in-flight gauges.
- Multi-site host dispatch metrics.
- Startup and script-load metrics.
- Goja handler metrics.
- SQLite query/exec metrics.
- `db.guard` metrics.
- Kanban fragment/action/render metrics.
- Optional pprof diagnostics listener.
- OpenTelemetry tracing with OTLP export and Jaeger-compatible deployment through a collector.
- Dashboard and alerting recommendations.
- Benchmark dimensions for instrumentation overhead.

### Out of scope for the first implementation

- Full distributed Kubernetes observability stack implementation.
- Log aggregation system design.
- User-level analytics.
- Raw SQL tracing.
- Per-session or per-user labels.
- Public exposure of diagnostics endpoints.

## Current Observability State

### Existing CLI and logging hooks

The CLI root uses Glazed logging setup. `cmd/goja-site/main.go:17-29` creates the Cobra root command and wires `logging.InitLoggerFromCobra` plus `logging.AddLoggingSectionToRootCommand`. This is useful for logging configuration, but it is not a metrics or tracing system.

Single-site `serve` currently exposes flags for bind address, database path, script directories, database policy, read-only/write mode, and dev error behavior. The settings struct is at `cmd/goja-site/serve.go:20-28`, and the flags are created at `cmd/goja-site/serve.go:44-52`. `serve` constructs `app.NewServer` and runs it at `cmd/goja-site/serve.go:75-82`.

Multi-site `serve-multi` currently exposes only `--config`. The settings struct is at `cmd/goja-site/serve_multi.go:20-22`, the config flag is at `cmd/goja-site/serve_multi.go:37-39`, and the command loads config, creates `app.NewMultiServer`, and runs it at `cmd/goja-site/serve_multi.go:52-67`.

Observation: diagnostics flags should be added to both commands, but the actual observability implementation should live below the CLI so tests and future embedding can reuse it.

### Existing server boundaries

`pkg/app/server.go` defines one hosted site. A `Server` owns `cfg`, `db`, `runtime`, `host`, and `httpSrv` at `pkg/app/server.go:22-28`. `NewServer` opens SQLite, creates the route host, builds runtime modules, creates the Goja runtime, sets the runtime on the host, and loads scripts at `pkg/app/server.go:31-87`. `Run` constructs `http.Server` with `Handler: s.host` at `pkg/app/server.go:99-100`.

Observation: because `Run` directly uses `s.host` as the handler, a request-level metrics wrapper should either wrap `s.host` in `Server.Run` or `Server.Handler`, or be installed inside the host/router if route pattern data is needed.

### Existing multi-site boundary

`pkg/app/multi_server.go` defines a `MultiServer` with `cfg`, `sites map[string]*Server`, and `httpSrv` at `pkg/app/multi_server.go:13-16`. `NewMultiServer` creates one `Server` per configured site at `pkg/app/multi_server.go:19-31`. `ServeHTTP` handles `/healthz` and `/readyz`, normalizes `r.Host`, looks up the site, returns 404 for unknown hosts, and dispatches to `site.ServeHTTP` at `pkg/app/multi_server.go:57-69`.

Observation: this is the right place to count unknown hosts, configured sites, dispatch duration, and selected site names. It is also where raw unknown Host headers must be handled carefully to avoid cardinality explosions.

### Existing database boundaries

`pkg/app/database.go` chooses between simple and guarded database policies. `buildDatabaseRuntimeConfig` constructs either `simpleDB` or `dbguard.NewMeteredDB`, then exposes `database` and `db` module names at `pkg/app/database.go:18-53`. `simpleDB.Query` and `simpleDB.Exec` live at `pkg/app/database.go:61-78`.

`pkg/dbguard/metered.go` wraps guarded writes. `MeteredDB.Query` currently forwards directly to the inner DB, and `MeteredDB.Exec` runs guard checks before and after SQL execution at `pkg/dbguard/metered.go:16-33`.

Observation: database instrumentation should wrap the `databasemod.QueryExecer` interface, not duplicate logic in every policy. Guard-specific metrics should live in `dbguard` because only the guard knows check phases and limit state.

### Existing guard internals

`pkg/dbguard/guard.go` stores guard options, Goja callback state, write counters, cleanup state, last stats, and last result at `pkg/dbguard/guard.go:15-27`. The default guard config sets cooldown, write-check cadence, and WAL inclusion at `pkg/dbguard/guard.go:29-31`. File and SQLite size stats are measured at `pkg/dbguard/guard.go:240-258`, including `PRAGMA page_size`, `page_count`, and `freelist_count`.

Observation: guard metrics should expose database size, live-byte estimates, configured limits, check counts, skipped checks, cleanup attempts, cleanup callback errors, and hard-limit blocks.

### Existing Kanban boundaries

`pkg/kanbanddsl/mount.go` registers the Kanban client script, fragment route, and action route. The fragment route calls `b.Render` at `pkg/kanbanddsl/mount.go:34-44`. The action route extracts the action, attaches the session to the body, calls `b.Dispatch`, optionally re-renders HTML with `uidsl.RenderAny`, and returns JSON at `pkg/kanbanddsl/mount.go:45-88`.

Observation: Kanban instrumentation should measure dispatch duration separately from refresh rendering duration. Otherwise a slow action callback and a slow HTML refresh look the same.

## Proposed Architecture

### High-level diagram

```text
                  +-------------------------------+
                  | goja-site process             |
                  |                               |
public traffic -->| app listener                  |
                  |  :8080                        |
                  |  Server / MultiServer         |
                  |    metrics wrappers           |
                  |    trace spans                 |
                  |                               |
Prometheus -----> | diagnostics listener          |
                  |  127.0.0.1:19090              |
operator -------> |  /metrics                     |
                  |  /debug/pprof/* optional      |
                  |                               |
OTel spans -----> | OTLP exporter                 |
                  +---------------+---------------+
                                  |
                                  v
                         OpenTelemetry Collector
                                  |
                    +-------------+-------------+
                    |                           |
                    v                           v
                  Jaeger                     Tempo/etc.
```

### Package layout

Add a small observability package under `pkg/observability` and keep product packages independent from Prometheus/OpenTelemetry details where practical.

Recommended layout:

```text
pkg/observability/
  config.go          # CLI-facing config structs and defaults
  registry.go        # Prometheus registry setup and standard collectors
  diagnostics.go     # diagnostics HTTP server with /metrics and optional pprof
  http.go            # HTTP middleware and response recorder
  sql.go             # QueryExecer wrapper and SQL kind labeling helpers
  startup.go         # startup/script-load metric helpers
  tracing.go         # OpenTelemetry provider setup, no-op fallback
  labels.go          # route/site/host sanitization helpers
  noop.go            # no-op implementation for tests/disabled mode if needed
```

Then integrate it from:

```text
cmd/goja-site/serve.go
cmd/goja-site/serve_multi.go
pkg/app/config.go
pkg/app/server.go
pkg/app/multi_server.go
pkg/app/database.go
pkg/dbguard/guard.go
pkg/dbguard/metered.go
pkg/kanbanddsl/mount.go
```

### Dependencies

Prometheus dependencies:

```bash
go get github.com/prometheus/client_golang/prometheus \
       github.com/prometheus/client_golang/prometheus/promhttp \
       github.com/prometheus/client_golang/prometheus/collectors
```

OpenTelemetry dependencies when tracing is added:

```bash
go get go.opentelemetry.io/otel \
       go.opentelemetry.io/otel/sdk \
       go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp \
       go.opentelemetry.io/otel/propagation \
       go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
```

Avoid Jaeger-specific exporters in application code unless there is a hard requirement. The modern pattern is app -> OTLP -> OpenTelemetry Collector -> Jaeger/Tempo/vendor.

## Configuration and CLI Design

### New config fields

Add an observability config to `pkg/app/config.go` or a new `pkg/observability.Config` referenced by app config.

API sketch:

```go
package observability

type Config struct {
    MetricsAddr       string
    MetricsPath       string
    EnableMetrics     bool
    EnablePprof       bool
    PprofPathPrefix   string

    TracingEnabled    bool
    ServiceName       string
    ServiceVersion    string
    OTLPEndpoint      string
    TraceSampleRatio  float64

    // Guardrails.
    UnsafeHostLabels  bool // default false; do not expose raw unknown hosts.
    RouteLabelMode    string // coarse, pattern, none
}
```

Default behavior:

```text
metrics disabled unless --metrics-addr is set
metrics path defaults to /metrics
pprof disabled unless --pprof is set
pprof only binds diagnostics listener, never public listener
tracing disabled unless --otel-enabled or OTEL_* env config is present
sample ratio defaults to 0.01 when tracing is enabled without explicit sampler
```

### New single-site flags

Add to `serveSettings` in `cmd/goja-site/serve.go`:

```go
MetricsAddr      string  `glazed:"metrics-addr"`
MetricsPath      string  `glazed:"metrics-path"`
Pprof            bool    `glazed:"pprof"`
OtelEnabled      bool    `glazed:"otel-enabled"`
OtelEndpoint     string  `glazed:"otel-endpoint"`
OtelSampleRatio  float64 `glazed:"otel-sample-ratio"`
ServiceName      string  `glazed:"service-name"`
```

Add flags:

```text
--metrics-addr        Diagnostics bind address for Prometheus metrics, disabled when empty.
--metrics-path        Metrics path, default /metrics.
--pprof               Serve /debug/pprof on the diagnostics listener.
--otel-enabled        Enable OpenTelemetry tracing.
--otel-endpoint       OTLP HTTP endpoint, default from OTEL_EXPORTER_OTLP_ENDPOINT.
--otel-sample-ratio   Trace sample ratio, default 0.01 when tracing is enabled.
--service-name        Service name for metrics/traces, default goja-site.
```

### New multi-site flags

Add the same observability flags to `serve-multi`. The diagnostics listener is process-wide, not per site.

The multi-site command should also expose a process-level metric showing configured sites:

```text
goja_site_hosts_configured 4
```

### Environment variables

Support standard OpenTelemetry environment variables where possible:

```text
OTEL_SERVICE_NAME=goja-site
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_TRACES_SAMPLER=parentbased_traceidratio
OTEL_TRACES_SAMPLER_ARG=0.01
```

Prometheus metrics are pull-based, so the important production config is the bind address and scrape path.

## Prometheus Metric Design

### Naming conventions

Use one prefix:

```text
goja_site_
```

Use Prometheus base units:

- seconds for durations,
- bytes for sizes,
- totals for counters.

Use histograms for latency, gauges for current state, and counters for events.

### Label rules

Allowed low-cardinality labels:

```text
site
mode
method
route
status_class
sql_kind
db_policy
phase
result
board
action
refresh
error_class
```

Avoid these labels:

```text
raw_path
raw_query
session_id
user_id
request_body
sql_text
raw_error_message
raw_unknown_host
file_path
```

If a label can grow without bound, do not add it.

### HTTP metrics

```text
goja_site_http_requests_total{site,method,route,status_class}
goja_site_http_request_duration_seconds_bucket{site,method,route}
goja_site_http_response_bytes_bucket{site,method,route}
goja_site_http_in_flight_requests{site}
```

Implementation notes:

- `site` is the configured site name when known.
- For single-site mode, use `site="default"` unless a site name is added to config.
- `route` should be a pattern or coarse route class, not raw path.
- `status_class` should be `2xx`, `3xx`, `4xx`, `5xx`, or `unknown`.

Initial route classifier if route patterns are not available:

```go
func CoarseRoute(path string) string {
    switch {
    case path == "/":
        return "/"
    case path == "/healthz" || path == "/readyz":
        return path
    case path == "/_kanban/client.js":
        return "/_kanban/client.js"
    case strings.HasPrefix(path, "/_kanban/") && strings.Contains(path, "/fragment"):
        return "/_kanban/:board/fragment"
    case strings.HasPrefix(path, "/_kanban/") && strings.Contains(path, "/action/"):
        return "/_kanban/:board/action/:action"
    case strings.HasPrefix(path, "/assets/"):
        return "/assets/*"
    default:
        return "other"
    }
}
```

### Multi-site metrics

```text
goja_site_hosts_configured
goja_site_site_up{site}
goja_site_unknown_host_requests_total{host_class}
goja_site_multi_dispatch_duration_seconds_bucket{result}
```

Use `host_class="unknown"`, not the raw host, by default. If there is a trusted internal environment where raw host labels are useful, hide it behind `--unsafe-host-labels` and document that it can break Prometheus cardinality.

### Startup metrics

```text
goja_site_startup_duration_seconds{phase}
goja_site_script_load_duration_seconds_bucket{site,script_class}
goja_site_script_load_errors_total{site,script_class,error_class}
goja_site_runtime_created_total{site}
goja_site_runtime_closed_total{site}
```

Phases:

```text
normalize_config
open_db
ping_db
build_db_runtime
build_goja_factory
create_goja_runtime
load_scripts
total
```

Do not label by full script path. Prefer basename or `script_class`. Script filenames in this repo are controlled and low-cardinality, but absolute paths are still noisy and environment-specific.

### Goja handler metrics

```text
goja_site_js_handler_duration_seconds_bucket{site,route}
goja_site_js_handler_errors_total{site,route,error_class}
goja_site_runtime_owner_call_duration_seconds_bucket{site,operation}
goja_site_runtime_owner_call_errors_total{site,operation,error_class}
```

Possible `operation` values:

```text
load_script
http_handler
kanban_fragment
kanban_action
kanban_render
```

If `gojahttp.Host` does not expose handler boundaries, first implement only outer HTTP metrics plus domain-specific Kanban and DB metrics. Do not overfit by patching internals without evidence.

### Database metrics

```text
goja_site_db_query_duration_seconds_bucket{site,db_policy,sql_kind}
goja_site_db_exec_duration_seconds_bucket{site,db_policy,sql_kind}
goja_site_db_errors_total{site,db_policy,operation,sql_kind,error_class}
goja_site_db_operations_total{site,db_policy,operation,sql_kind}
```

`operation` values:

```text
query
exec
```

`sql_kind` values should be coarse and derived from existing parsing where possible:

```text
select
insert
update
delete
replace
create
alter
drop
pragma
with
explain
vacuum
other
```

No raw SQL text.

### DB guard metrics

```text
goja_site_db_guard_checks_total{site,phase,result}
goja_site_db_guard_check_duration_seconds_bucket{site,phase,result}
goja_site_db_guard_cleanup_attempts_total{site,result}
goja_site_db_guard_limit_exceeded_total{site,kind,hard}
goja_site_db_size_bytes{site,component}
goja_site_db_live_bytes{site}
goja_site_db_limit_bytes{site,limit_type}
goja_site_db_guard_writes_since_check{site}
```

`component` values:

```text
db
wal
shm
total
```

`limit_type` values:

```text
soft
hard
max
```

`phase` values:

```text
before_exec
after_exec
manual_stats
check_now
cleanup_callback
```

### Kanban metrics

```text
goja_site_kanban_fragment_duration_seconds_bucket{site,board}
goja_site_kanban_action_duration_seconds_bucket{site,board,action,refresh}
goja_site_kanban_dispatch_duration_seconds_bucket{site,board,action}
goja_site_kanban_render_duration_seconds_bucket{site,board,reason}
goja_site_kanban_rendered_html_bytes_bucket{site,board,reason}
goja_site_kanban_action_errors_total{site,board,action,error_class}
```

`reason` values:

```text
fragment
action_refresh
full_page
```

The board label is acceptable only if board IDs are declared by trusted scripts and bounded. If arbitrary user-defined board IDs become possible, replace `board` with `board_class`.

### Runtime and process collectors

Use Prometheus standard collectors:

- Go collector: GC, goroutines, heap, threads.
- Process collector: CPU, RSS, file descriptors where supported.

With `prometheus/client_golang`, these are registered on a custom registry:

```go
registry := prometheus.NewRegistry()
registry.MustRegister(collectors.NewGoCollector())
registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
```

## Implementation Details

### Diagnostics server

Create a diagnostics server that can serve `/metrics` and optionally `/debug/pprof`.

Pseudocode:

```go
package observability

type DiagnosticsServer struct {
    srv *http.Server
}

func StartDiagnostics(ctx context.Context, cfg Config, reg *prometheus.Registry) (*DiagnosticsServer, error) {
    if cfg.MetricsAddr == "" {
        return nil, nil
    }

    mux := http.NewServeMux()
    metricsPath := cfg.MetricsPath
    if metricsPath == "" {
        metricsPath = "/metrics"
    }
    mux.Handle(metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

    if cfg.EnablePprof {
        MountPprof(mux, "/debug/pprof")
    }

    srv := &http.Server{
        Addr:              cfg.MetricsAddr,
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,
    }

    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = srv.Shutdown(shutdownCtx)
    }()

    go func() {
        err := srv.ListenAndServe()
        if err != nil && !errors.Is(err, http.ErrServerClosed) {
            // log error; do not panic after startup
        }
    }()

    return &DiagnosticsServer{srv: srv}, nil
}
```

Mounting pprof manually avoids accidentally exposing it on the default global mux:

```go
func MountPprof(mux *http.ServeMux, prefix string) {
    mux.HandleFunc(prefix+"/", pprof.Index)
    mux.HandleFunc(prefix+"/cmdline", pprof.Cmdline)
    mux.HandleFunc(prefix+"/profile", pprof.Profile)
    mux.HandleFunc(prefix+"/symbol", pprof.Symbol)
    mux.HandleFunc(prefix+"/trace", pprof.Trace)
    mux.Handle(prefix+"/goroutine", pprof.Handler("goroutine"))
    mux.Handle(prefix+"/heap", pprof.Handler("heap"))
    mux.Handle(prefix+"/threadcreate", pprof.Handler("threadcreate"))
    mux.Handle(prefix+"/block", pprof.Handler("block"))
    mux.Handle(prefix+"/mutex", pprof.Handler("mutex"))
}
```

### HTTP middleware

Create a response recorder:

```go
type statusRecorder struct {
    http.ResponseWriter
    status int
    bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
    if r.status == 0 {
        r.status = status
        r.ResponseWriter.WriteHeader(status)
    }
}

func (r *statusRecorder) Write(p []byte) (int, error) {
    if r.status == 0 {
        r.status = http.StatusOK
    }
    n, err := r.ResponseWriter.Write(p)
    r.bytes += n
    return n, err
}
```

Middleware sketch:

```go
func (m *HTTPMetrics) Wrap(site string, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        route := CoarseRoute(r.URL.Path)
        labels := prometheus.Labels{
            "site": site,
            "method": r.Method,
            "route": route,
        }
        m.InFlight.WithLabelValues(site).Inc()
        start := time.Now()
        rec := &statusRecorder{ResponseWriter: w}
        defer func() {
            m.InFlight.WithLabelValues(site).Dec()
            status := rec.status
            if status == 0 { status = http.StatusOK }
            statusClass := StatusClass(status)
            m.Requests.WithLabelValues(site, r.Method, route, statusClass).Inc()
            m.Duration.WithLabelValues(site, r.Method, route).Observe(time.Since(start).Seconds())
            m.ResponseBytes.WithLabelValues(site, r.Method, route).Observe(float64(rec.bytes))
        }()
        next.ServeHTTP(rec, r)
    })
}
```

Integration options:

1. In `Server.Run`, set `Handler: obs.HTTP.Wrap(site, s.host)`.
2. In `Server.Handler`, return wrapped handler.
3. In `MultiServer.ServeHTTP`, wrap per-site dispatch manually.

Preferred first implementation:

- Single-site: `Server` has a `siteName` and `observability` field; `Handler()` returns wrapped `s.host`.
- Multi-site: `MultiServer` records dispatch metrics, then calls each site's already-wrapped handler with the correct site label.

### Database wrapper

Wrap the `databasemod.QueryExecer` created in `buildDatabaseRuntimeConfig`.

Pseudocode:

```go
type InstrumentedQueryExecer struct {
    inner    databasemod.QueryExecer
    site     string
    policy   string
    metrics  *DBMetrics
}

func (i *InstrumentedQueryExecer) Query(query string, args ...any) (*sql.Rows, error) {
    kind := SQLKindLabel(query)
    start := time.Now()
    rows, err := i.inner.Query(query, args...)
    i.metrics.ObserveQuery(i.site, i.policy, kind, time.Since(start), err)
    return rows, err
}

func (i *InstrumentedQueryExecer) Exec(query string, args ...any) (sql.Result, error) {
    kind := SQLKindLabel(query)
    start := time.Now()
    result, err := i.inner.Exec(query, args...)
    i.metrics.ObserveExec(i.site, i.policy, kind, time.Since(start), err)
    return result, err
}
```

This avoids modifying the JavaScript database module and keeps all SQL instrumentation in Go.

### Guard instrumentation

Add an optional observer interface to `pkg/dbguard` to avoid importing Prometheus into guard core if you want to keep it decoupled.

```go
type Observer interface {
    ObserveCheck(phase string, result string, duration time.Duration)
    ObserveLimitExceeded(kind SQLKind, hard bool)
    SetDBSize(component string, bytes int64)
    SetDBLimit(limitType string, bytes int64)
    ObserveCleanup(result string, duration time.Duration)
}
```

Add to `Guard`:

```go
type Guard struct {
    // existing fields
    observer Observer
}

func (g *Guard) SetObserver(o Observer) {
    g.mu.Lock()
    defer g.mu.Unlock()
    g.observer = o
}
```

Call the observer from check paths after measuring duration and stats. Keep observer calls non-blocking and panic-safe:

```go
func safeObserve(fn func()) {
    defer func() { _ = recover() }()
    fn()
}
```

### Kanban instrumentation

Avoid importing Prometheus directly into `pkg/kanbanddsl` if possible. Define an observer interface:

```go
type Observer interface {
    ObserveFragment(board string, duration time.Duration, err error)
    ObserveAction(board, action string, refresh bool, duration time.Duration, err error)
    ObserveDispatch(board, action string, duration time.Duration, err error)
    ObserveRender(board, reason string, duration time.Duration, htmlBytes int, err error)
}
```

Then the app-level registrar can pass an observer into `kanbanddsl.NewRegistrar(...)` if the package API supports it, or a later change can add that API. First implementation can instrument around the code in `Board.Mount` directly.

Action route pseudocode:

```go
startAction := time.Now()
startDispatch := time.Now()
result, err := b.Dispatch(action, bodyObj)
observer.ObserveDispatch(b.cfg.ID, action, time.Since(startDispatch), err)
if err != nil { panic(b.vm.NewGoError(err)) }

refresh := shouldRefresh(out["refresh"])
if refresh {
    startRender := time.Now()
    node, err := b.Render(...)
    html, err := uidsl.RenderAny(...)
    observer.ObserveRender(b.cfg.ID, "action_refresh", time.Since(startRender), len(html), err)
}
observer.ObserveAction(b.cfg.ID, action, refresh, time.Since(startAction), nil)
```

### Tracing with OpenTelemetry

Use OpenTelemetry with no-op fallback when disabled. Do not put Jaeger-specific concepts in core code.

Initialization pseudocode:

```go
func InitTracing(ctx context.Context, cfg Config) (shutdown func(context.Context) error, tracer trace.Tracer, err error) {
    if !cfg.TracingEnabled {
        return func(context.Context) error { return nil }, otel.Tracer("goja-site/noop"), nil
    }

    exporter, err := otlptracehttp.New(ctx,
        otlptracehttp.WithEndpointURL(cfg.OTLPEndpoint),
    )
    if err != nil { return nil, nil, err }

    sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceSampleRatio))
    provider := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithSampler(sampler),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName(cfg.ServiceName),
        )),
    )
    otel.SetTracerProvider(provider)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{}, propagation.Baggage{},
    ))
    return provider.Shutdown, provider.Tracer("github.com/go-go-golems/goja-site"), nil
}
```

HTTP handler wrapping with `otelhttp`:

```go
handler := otelhttp.NewHandler(
    observedHandler,
    "goja-site.http.request",
    otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
        return r.Method + " " + CoarseRoute(r.URL.Path)
    }),
)
```

Span hierarchy:

```text
goja-site.http.request
  ├─ goja-site.multi.dispatch
  ├─ goja-site.js.handler
  │   ├─ goja-site.db.query
  │   ├─ goja-site.db.exec
  │   ├─ goja-site.db.guard.check
  │   ├─ goja-site.kanban.dispatch
  │   └─ goja-site.kanban.render
  └─ response write
```

Trace attributes:

```text
goja_site.site
goja_site.route
goja_site.db_policy
goja_site.sql_kind
goja_site.kanban.board
goja_site.kanban.action
goja_site.kanban.refresh
http.request.method
http.route
http.response.status_code
```

Do not attach request bodies, session cookies, raw SQL, or arbitrary JavaScript values to spans.

## Benchmark Integration

The benchmark ticket should treat observability as a benchmark dimension.

Add these benchmark modes:

```text
observability=off
metrics=on
metrics+pprof=on
metrics+tracing_sample_1pct=on
metrics+tracing_sample_100pct=on-short-run-only
```

Benchmark questions:

- How much overhead does metrics wrapping add to null-route throughput?
- How much overhead does DB instrumentation add to read/write routes?
- How much overhead does Kanban action instrumentation add?
- How much overhead does 1% tracing add?
- What is the cost of 100% tracing during short debugging runs?

Example commands after implementation:

```bash
scripts/bench-http.sh --scenario null --duration 60s --concurrency 64 --observability off
scripts/bench-http.sh --scenario null --duration 60s --concurrency 64 --metrics-addr 127.0.0.1:19090
scripts/bench-http.sh --scenario kanban --duration 60s --concurrency 32 --metrics-addr 127.0.0.1:19090 --otel-sample-ratio 0.01
```

Each benchmark report should include:

```yaml
observability:
  metrics_enabled: true
  pprof_enabled: false
  tracing_enabled: true
  trace_sample_ratio: 0.01
  metrics_scraped: true
```

## Production Deployment Notes

### Prometheus scrape config

Example:

```yaml
scrape_configs:
  - job_name: goja-site
    metrics_path: /metrics
    static_configs:
      - targets:
          - goja-site:19090
```

For Kubernetes, expose the diagnostics port only inside the cluster and use a `ServiceMonitor` or equivalent.

### OpenTelemetry Collector to Jaeger

Example conceptual collector config:

```yaml
receivers:
  otlp:
    protocols:
      http:
      grpc:

processors:
  batch:

exporters:
  jaeger:
    endpoint: jaeger-collector:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [jaeger]
```

In newer collector deployments, prefer the exporter recommended by the installed collector version. The application should only know about OTLP.

### Dashboard starter panels

Initial Grafana dashboard panels:

- Request rate by site.
- p50/p95/p99 request latency by site and route.
- Error rate by site and status class.
- Unknown host requests.
- Go heap, goroutines, GC pause.
- DB query/exec latency by site and SQL kind.
- DB errors by site and operation.
- DB size and guard limit utilization.
- Kanban action latency by board/action.
- Runtime startup/script-load duration.

### Alert starter rules

Examples:

```text
High5xxRate:
  5xx request rate > 1% for 10 minutes

HighRequestLatency:
  p95 request latency > 1s for 10 minutes

DBNearHardLimit:
  db total bytes / hard limit > 0.9 for 15 minutes

UnknownHostSpike:
  unknown host request rate > expected threshold

GoroutineLeakSuspected:
  goroutines increasing continuously over 30 minutes
```

Tune thresholds after production baselines exist.

## Implementation Plan

### Phase 1: Metrics scaffolding and diagnostics listener

Files:

- `pkg/observability/config.go`
- `pkg/observability/registry.go`
- `pkg/observability/diagnostics.go`
- `cmd/goja-site/serve.go`
- `cmd/goja-site/serve_multi.go`

Tasks:

1. Add Prometheus dependency.
2. Add config and defaults.
3. Add CLI flags for `--metrics-addr`, `--metrics-path`, and `--pprof`.
4. Start diagnostics listener only when `--metrics-addr` is set.
5. Register Go and process collectors.
6. Add tests that `/metrics` is not started by default and is served when configured.

Acceptance command:

```bash
go test ./...
go run ./cmd/goja-site serve --metrics-addr 127.0.0.1:19090 --scripts examples/kanban/scripts --db /tmp/goja-observe.db
curl -fsS http://127.0.0.1:19090/metrics | grep '^go_'
```

### Phase 2: HTTP and multi-site metrics

Files:

- `pkg/observability/http.go`
- `pkg/observability/labels.go`
- `pkg/app/server.go`
- `pkg/app/multi_server.go`

Tasks:

1. Add HTTP response recorder.
2. Add route classifier.
3. Wrap single-site handlers.
4. Record multi-site dispatch metrics and unknown host counter.
5. Add tests using `httptest` and a custom Prometheus registry.

Acceptance command:

```bash
go test ./pkg/app ./pkg/observability
```

Manual check:

```bash
curl -H 'Host: unknown.example.test' http://127.0.0.1:8080/
curl -fsS http://127.0.0.1:19090/metrics | grep goja_site_unknown_host_requests_total
```

### Phase 3: Startup and script-load metrics

Files:

- `pkg/observability/startup.go`
- `pkg/app/server.go`
- `pkg/app/multi_server.go`

Tasks:

1. Time `NewServer` phases.
2. Time each script load.
3. Count runtime created/closed.
4. Ensure script labels are bounded.

Acceptance criteria:

- Metrics expose startup phases after server creation.
- Failed script loads increment error counters.
- No full absolute script paths appear in labels.

### Phase 4: Database and guard metrics

Files:

- `pkg/observability/sql.go`
- `pkg/app/database.go`
- `pkg/dbguard/guard.go`
- `pkg/dbguard/metered.go`

Tasks:

1. Wrap QueryExecer with timing metrics.
2. Add SQL kind label helper.
3. Add guard observer interface.
4. Record guard checks, DB size gauges, and limit gauges.
5. Add tests for label values and no raw SQL exposure.

Acceptance command:

```bash
go test ./pkg/app ./pkg/dbguard ./pkg/observability
```

### Phase 5: Kanban metrics

Files:

- `pkg/kanbanddsl/mount.go`
- `pkg/kanbanddsl/registrar.go`
- `pkg/observability/kanban.go` if needed

Tasks:

1. Add Kanban observer interface.
2. Record fragment duration.
3. Record action total duration.
4. Record dispatch duration.
5. Record refresh render duration and rendered HTML bytes.
6. Add tests for action labels and refresh labels.

Acceptance criteria:

- Kanban metrics appear after fragment and action requests.
- Action and render timings are separate.
- Board/action labels are bounded by trusted script definitions.

### Phase 6: pprof diagnostics

Files:

- `pkg/observability/diagnostics.go`

Tasks:

1. Mount pprof handlers manually when `--pprof` is set.
2. Confirm pprof is not exposed without flag.
3. Document security warning.

Acceptance command:

```bash
go run ./cmd/goja-site serve --metrics-addr 127.0.0.1:19090 --pprof ...
curl -fsS http://127.0.0.1:19090/debug/pprof/goroutine?debug=1 | head
```

### Phase 7: OpenTelemetry tracing

Files:

- `pkg/observability/tracing.go`
- `pkg/app/server.go`
- `pkg/app/multi_server.go`
- `pkg/app/database.go`
- `pkg/dbguard/guard.go`
- `pkg/kanbanddsl/mount.go`

Tasks:

1. Add OTel dependencies.
2. Add tracing config and CLI flags.
3. Initialize tracer provider in CLI run path.
4. Wrap HTTP handlers with `otelhttp`.
5. Add manual spans around DB, guard, and Kanban operations.
6. Add shutdown flushing on process exit.
7. Add tests with in-memory or no-op tracer provider where practical.

Acceptance criteria:

- Tracing disabled by default.
- OTLP endpoint works when configured.
- Span attributes are low-cardinality and do not include sensitive data.
- Sampling works.

### Phase 8: Dashboards, alerts, and benchmark overhead report

Files:

```text
deploy/observability/prometheus-scrape.example.yaml
deploy/observability/otel-collector-jaeger.example.yaml
deploy/observability/grafana-dashboard-goja-site.json
bench/results/observability-overhead-baseline.md
```

Tasks:

1. Add example scrape config.
2. Add example OTel collector config.
3. Add starter Grafana dashboard.
4. Run benchmark overhead matrix.
5. Document overhead results in the ticket.

## Testing Strategy

### Unit tests

- Label sanitization tests.
- Route classifier tests.
- Status recorder tests.
- SQL kind classifier tests.
- Metrics registry tests with isolated registries.
- Diagnostics server tests using random local ports or `httptest` where possible.

### Integration tests

- Start `Server` with metrics enabled and confirm metrics change after requests.
- Start `MultiServer` with two sites and confirm per-site labels.
- Request unknown host and confirm unknown host metric increments without raw host label.
- Execute DB read/write and confirm query/exec metrics.
- Execute Kanban action and confirm action/render metrics.

### Performance tests

- Run benchmark suite with metrics disabled.
- Run benchmark suite with metrics enabled.
- Run benchmark suite with metrics plus 1% tracing.
- Run short benchmark with 100% tracing for worst-case overhead.

### Security tests

- Confirm `/metrics` and `/debug/pprof` are not served on the public app listener.
- Confirm pprof is disabled unless explicitly enabled.
- Confirm raw SQL and session IDs are not in metric labels or span attributes.

## Risks and Mitigations

### High-cardinality metrics

Risk: raw paths, hosts, SQL text, or errors make Prometheus expensive or unusable.

Mitigation: route classifier, SQL kind classifier, status class labels, error class labels, and no raw unknown host labels by default.

### Instrumentation overhead

Risk: metrics and tracing reduce throughput or distort benchmark results.

Mitigation: benchmark with observability off/on and with tracing sample ratios. Use no-op observers when disabled.

### Sensitive diagnostics exposure

Risk: pprof and metrics expose internal state.

Mitigation: diagnostics listener disabled by default, bind to localhost/private addresses, never mount pprof on public app listener.

### Package coupling

Risk: Prometheus imports spread through domain packages.

Mitigation: use observer interfaces for `dbguard` and `kanbanddsl`; keep Prometheus construction in `pkg/observability`.

### Trace volume explosion

Risk: tracing every request and DB operation can generate too much data.

Mitigation: default sample ratio, parent-based sampling, batch exporter, and short 100% tracing windows only for debugging.

## Alternatives Considered

### Metrics only for benchmarks

Rejected. The same metrics that explain stress tests are useful in production. Implementing benchmark-only metrics would create duplicate instrumentation later.

### Logs instead of metrics

Rejected as the primary system. Logs are useful for discrete events and debugging, but they are weaker for latency histograms, rates, dashboards, and alerts.

### Jaeger-specific tracing APIs

Rejected. OpenTelemetry keeps the application backend-neutral and works with Jaeger, Tempo, Honeycomb, Datadog, and other systems through OTLP/collectors.

### Public `/metrics` on the app listener

Rejected as the default. It is operationally convenient but unsafe for production unless protected by ingress/auth. A separate diagnostics listener is a safer default.

### Label by raw route path

Rejected. Raw paths become high-cardinality as soon as IDs or search terms enter URLs.

## Open Questions

- Should single-site mode have a configurable `--site-name`, or should it always use `site="default"`?
- Can `gojahttp.Host` expose route patterns so HTTP metrics can use exact registered patterns rather than coarse route classes?
- Should `dbguard` own Prometheus metrics directly or stay decoupled through an observer interface?
- Which histogram buckets fit the expected workload: web-style buckets, sub-millisecond buckets for null routes, or both?
- Should tracing be initialized entirely from standard `OTEL_*` environment variables or mirrored with explicit Glazed flags?
- Should production deployment use Prometheus directly or an OpenTelemetry Collector for metrics too?

## Implementation Checklist for an Intern

1. Read the first benchmark guide in this ticket.
2. Read these files:
   - `cmd/goja-site/serve.go`
   - `cmd/goja-site/serve_multi.go`
   - `pkg/app/server.go`
   - `pkg/app/multi_server.go`
   - `pkg/app/database.go`
   - `pkg/dbguard/metered.go`
   - `pkg/dbguard/guard.go`
   - `pkg/kanbanddsl/mount.go`
3. Add `pkg/observability` with config, registry, diagnostics server, and label helpers.
4. Add metrics flags to `serve` and `serve-multi`.
5. Add HTTP metrics wrapper with low-cardinality route labels.
6. Add multi-site dispatch and unknown-host metrics.
7. Add database metrics with SQL kind labels only.
8. Add guard metrics through an observer interface.
9. Add Kanban metrics through an observer interface.
10. Add pprof only on the diagnostics listener.
11. Add OTel tracing only after metrics are stable.
12. Run unit and integration tests.
13. Run benchmark overhead comparisons.
14. Update this ticket with measured overhead numbers and screenshots/links to dashboards if available.

## References

- `cmd/goja-site/main.go`: CLI root and Glazed logging setup.
- `cmd/goja-site/serve.go`: single-site flags and run path.
- `cmd/goja-site/serve_multi.go`: multi-site flags and run path.
- `pkg/app/config.go`: app config location for observability config extension.
- `pkg/app/server.go`: single-site runtime and HTTP serving boundary.
- `pkg/app/multi_server.go`: multi-site dispatch boundary and unknown-host handling.
- `pkg/app/database.go`: database policy and module wrapping boundary.
- `pkg/dbguard/metered.go`: guarded DB exec boundary.
- `pkg/dbguard/guard.go`: guard stats, limits, callbacks, and measurement logic.
- `pkg/kanbanddsl/mount.go`: Kanban fragment/action/render boundaries.
- `ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/design-doc/01-goja-hosting-stress-test-benchmark-and-performance-guide.md`: benchmark plan that this observability guide extends.
