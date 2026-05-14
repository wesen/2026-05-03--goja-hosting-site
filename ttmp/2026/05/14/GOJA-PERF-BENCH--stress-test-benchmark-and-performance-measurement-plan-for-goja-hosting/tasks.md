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

- [ ] Add Prometheus client dependency.
- [ ] Create `pkg/observability` package with config, registry, diagnostics server, and label helpers.
- [ ] Register Go runtime and process collectors on a custom registry.
- [ ] Add `--metrics-addr`, `--metrics-path`, and `--pprof` flags to `serve`.
- [ ] Add `--metrics-addr`, `--metrics-path`, and `--pprof` flags to `serve-multi`.
- [ ] Start diagnostics listener only when `--metrics-addr` is set.
- [ ] Mount `/metrics` on the diagnostics listener.
- [ ] Mount `/debug/pprof/*` only when `--pprof` is set.
- [ ] Add basic HTTP metrics: requests total, request duration, response bytes, in-flight requests.
- [ ] Add coarse low-cardinality route classifier.
- [ ] Add multi-site metrics: configured host count, site up gauge, unknown host request counter, dispatch duration.
- [ ] Add unit tests for diagnostics server, route labels, status class labels, and response recorder.
- [ ] Add app integration tests proving metrics increment for single-site and multi-site requests.
- [ ] Run `go test ./...`.
- [ ] Commit Phase 1.

## Phase 2 — Load-generation MVP

Goal: create a repeatable load harness that can use the new `/metrics` endpoint from day one.

- [ ] Choose primary initial load engine, preferably Vegeta CLI plus later Go-library embedding.
- [ ] Add `bench/targets/` with null-route and multi-site target examples.
- [ ] Add `bench/results/.gitignore` to avoid committing generated result artifacts.
- [ ] Add `scripts/bench-vegeta.sh` that builds or accepts a `goja-site` binary, starts a server, waits for readiness, runs Vegeta, captures output, and cleans up.
- [ ] Support single-site null route scenario.
- [ ] Support single-site render route scenario.
- [ ] Support DB read/write route scenario.
- [ ] Support multi-site Host-header mix scenario.
- [ ] Scrape `/metrics` before and after each run.
- [ ] Record metadata: commit, dirty status, Go version, OS/arch, scenario, rate, duration, concurrency/worker settings, observability mode.
- [ ] Write Markdown and JSON summaries under `bench/results/`.
- [ ] Document how to install/use Vegeta and alternatives (`fortio`, `hey`, `bombardier`, `k6`).
- [ ] Run a short local smoke load test.
- [ ] Commit Phase 2.

## Phase 3 — Database and db.guard metrics

Goal: expose the database and guard bottlenecks that realistic goja-site apps are likely to hit.

- [ ] Add SQL kind classifier with bounded labels (`select`, `insert`, `update`, `delete`, `pragma`, `other`, etc.).
- [ ] Wrap `databasemod.QueryExecer` with query/exec duration metrics.
- [ ] Add DB operation counters and error counters.
- [ ] Add tests proving raw SQL is never exported as a metric label.
- [ ] Add `db.guard` observer interface or equivalent decoupled metrics hook.
- [ ] Expose guard check duration, check totals, cleanup attempts, hard/soft limit events, and DB size gauges.
- [ ] Add tests for guarded write metrics.
- [ ] Run `go test ./...`.
- [ ] Commit Phase 3.

## Phase 4 — Kanban metrics

Goal: measure realistic application interaction paths separately from generic HTTP timing.

- [ ] Add Kanban observer interface or metrics hook.
- [ ] Measure fragment render duration.
- [ ] Measure action total duration.
- [ ] Measure action dispatch duration.
- [ ] Measure action refresh render duration.
- [ ] Measure rendered HTML bytes.
- [ ] Count action errors by bounded error class.
- [ ] Add tests for fragment/action metrics and refresh labels.
- [ ] Run `go test ./...`.
- [ ] Commit Phase 4.

## Phase 5 — pprof capture automation in load harness

Goal: make expensive bottleneck analysis reproducible during stress runs.

- [ ] Extend load harness to capture CPU profile during a run when pprof is enabled.
- [ ] Capture heap profile after run.
- [ ] Capture goroutine profile after run.
- [ ] Store pprof artifacts in result directory but keep them ignored by git.
- [ ] Add report links/instructions for `go tool pprof`.
- [ ] Run short profile capture smoke test.
- [ ] Commit Phase 5.

## Phase 6 — OpenTelemetry tracing

Goal: add sampled distributed tracing after metrics and load generation are stable.

- [ ] Add OpenTelemetry dependencies.
- [ ] Add tracing config and CLI flags/env handling.
- [ ] Initialize tracer provider and OTLP exporter with clean shutdown.
- [ ] Wrap HTTP handlers with OTel HTTP instrumentation.
- [ ] Add spans for multi-site dispatch, DB query/exec, guard checks, Kanban actions, and renders.
- [ ] Add safe low-cardinality span attributes only.
- [ ] Add sampling controls and tests/no-op defaults.
- [ ] Add example OTel Collector to Jaeger config.
- [ ] Run `go test ./...`.
- [ ] Commit Phase 6.

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
