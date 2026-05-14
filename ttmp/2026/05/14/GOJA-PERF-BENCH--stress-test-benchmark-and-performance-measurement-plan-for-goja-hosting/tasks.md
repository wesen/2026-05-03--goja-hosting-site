# Tasks

## Phase 0 — Ticket planning and documentation baseline

- [x] Create GOJA-PERF-BENCH ticket workspace.
- [x] Inspect goja-site runtime, multi-site dispatch, database policy, and Kanban mount paths.
- [x] Capture baseline validation command output for the current repository.
- [x] Write intern-oriented stress test, benchmark, and performance measurement design/implementation guide.
- [x] Write dedicated production observability metrics/tracing analysis and implementation guide.
- [x] Write and maintain chronological investigation diary.
- [x] Relate key source files and operational scripts to ticket documents.
- [x] Add missing performance/observability vocabulary topics.
- [x] Validate ticket with `docmgr doctor --ticket GOJA-PERF-BENCH --stale-after 30`.
- [x] Upload ticket bundle to reMarkable.

## Phase 1 — Minimal production observability spine

Goal: add a private diagnostics listener and basic Prometheus metrics before serious load generation.

- [x] Add Prometheus client dependency.
- [x] Create `pkg/observability` package with config, registry, diagnostics server, and label helpers.
- [x] Register Go runtime and process collectors on a custom registry.
- [x] Add `--metrics-addr`, `--metrics-path`, and `--pprof` flags to `serve`.
- [x] Add `--metrics-addr`, `--metrics-path`, and `--pprof` flags to `serve-multi`.
- [x] Start diagnostics listener only when `--metrics-addr` is set.
- [x] Mount `/metrics` on the diagnostics listener.
- [x] Mount `/debug/pprof/*` only when `--pprof` is set.
- [x] Add basic HTTP metrics: requests total, request duration, response bytes, in-flight requests.
- [x] Add coarse low-cardinality route classifier.
- [x] Add multi-site metrics: configured host count, site up gauge, unknown host request counter, dispatch duration.
- [x] Add unit tests for diagnostics server, route labels, status class labels, and response recorder.
- [x] Add app integration tests proving metrics increment for single-site and multi-site requests.
- [x] Run `go test ./...`.
- [x] Commit Phase 1.

## Phase 2 — Load-generation MVP

Goal: create a repeatable load harness that can use the new `/metrics` endpoint from day one.

- [x] Choose primary initial load engine, preferably Vegeta CLI plus later Go-library embedding.
- [x] Add `bench/targets/` with null-route and multi-site target examples.
- [x] Add `bench/results/.gitignore` to avoid committing generated result artifacts.
- [x] Add `scripts/bench-vegeta.sh` that builds or accepts a `goja-site` binary, starts a server, waits for readiness, runs Vegeta, captures output, and cleans up.
- [x] Support single-site null route scenario.
- [x] Support single-site render route scenario.
- [x] Support DB read/write route scenario.
- [x] Support multi-site Host-header mix scenario.
- [x] Scrape `/metrics` before and after each run.
- [x] Record metadata: commit, dirty status, Go version, OS/arch, scenario, rate, duration, concurrency/worker settings, observability mode.
- [x] Write Markdown and JSON summaries under `bench/results/`.
- [x] Document how to install/use Vegeta and alternatives (`fortio`, `hey`, `bombardier`, `k6`).
- [x] Run a short local smoke load test.
- [x] Commit Phase 2.

## Phase 3 — Database and db.guard metrics

Goal: expose the database and guard bottlenecks that realistic goja-site apps are likely to hit.

- [x] Add SQL kind classifier with bounded labels (`select`, `insert`, `update`, `delete`, `pragma`, `other`, etc.).
- [x] Wrap `databasemod.QueryExecer` with query/exec duration metrics.
- [x] Add DB operation counters and error counters.
- [x] Add tests proving raw SQL is never exported as a metric label.
- [x] Add `db.guard` observer interface or equivalent decoupled metrics hook.
- [x] Expose guard check duration, check totals, cleanup attempts, hard/soft limit events, and DB size gauges.
- [x] Add tests for guarded write metrics.
- [x] Run `go test ./...`.
- [x] Commit Phase 3.

## Phase 4 — Kanban metrics

Goal: measure realistic application interaction paths separately from generic HTTP timing.

- [x] Add Kanban observer interface or metrics hook.
- [x] Measure fragment render duration.
- [x] Measure action total duration.
- [x] Measure action dispatch duration.
- [x] Measure action refresh render duration.
- [x] Measure rendered HTML bytes.
- [x] Count action errors by bounded error class.
- [x] Add tests for fragment/action metrics and refresh labels.
- [x] Run `go test ./...`.
- [x] Commit Phase 4.

## Phase 5 — pprof capture automation in load harness

Goal: make expensive bottleneck analysis reproducible during stress runs.

- [x] Extend load harness to capture CPU profile during a run when pprof is enabled.
- [x] Capture heap profile after run.
- [x] Capture goroutine profile after run.
- [x] Store pprof artifacts in result directory but keep them ignored by git.
- [x] Add report links/instructions for `go tool pprof`.
- [x] Run short profile capture smoke test.
- [ ] Commit Phase 5.

## Phase 6 — OpenTelemetry tracing

Goal: add sampled distributed tracing after metrics and load generation are stable.

- [x] Add OpenTelemetry dependencies.
- [x] Add tracing config and CLI flags/env handling.
- [x] Initialize tracer provider and OTLP exporter with clean shutdown.
- [x] Wrap HTTP handlers with OTel HTTP instrumentation.
- [ ] Add spans for multi-site dispatch, guard checks, Kanban actions, and renders.
- [x] Add spans for DB query/exec operations.
- [x] Add safe low-cardinality span attributes only.
- [x] Add sampling controls and tests/no-op defaults.
- [x] Add example OTel Collector to Jaeger config.
- [x] Run `go test ./...`.
- [ ] Commit Phase 6.


## Phase 6A — go-go-goja request context propagation design

Goal: design the upstream go-go-goja changes needed for native modules to access the active request context while JavaScript handlers execute.

- [x] Inspect go-go-goja runtimeowner, runtimebridge, gojahttp, and database module context behavior.
- [x] Write dedicated request-context propagation analysis and implementation guide.
- [x] Implement runtimebridge current-call context stack in go-go-goja.
- [x] Wrap runtimeowner invoke/invokePost with current-call context.
- [x] Add QueryExecerContext support to go-go-goja database module.
- [x] Update goja-site DB wrappers to QueryContext/ExecContext after go-go-goja change is available.
- [x] Add trace parentage tests proving DB spans are children of HTTP request spans.

## Phase 7 — Benchmark scenarios and production dashboards

Goal: turn raw metrics into usable engineering and operations artifacts.

- [ ] Add benchmark scripts/fixtures described by the first design guide.
- [ ] Add in-process `testing.B` benchmarks for startup, null route, render route, DB routes, guarded writes, and multi-site dispatch.
- [ ] Add observability overhead benchmark matrix: off, metrics, metrics+pprof, metrics+tracing sampled.
- [ ] Add starter Prometheus scrape config.
- [ ] Add starter Grafana dashboard JSON or documented dashboard panels.
- [ ] Add starter alert rules.
- [ ] Run benchmark smoke suite.
- [ ] Append measured baseline report to ticket.
- [ ] Upload refreshed bundle to reMarkable.
- [ ] Commit Phase 7.
