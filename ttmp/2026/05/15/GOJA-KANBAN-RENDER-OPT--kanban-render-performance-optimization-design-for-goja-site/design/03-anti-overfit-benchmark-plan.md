---
Title: Anti-overfit benchmark plan for goja-site performance work
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - goja
    - kanban
    - performance
    - benchmarking
    - optimization
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: bench/scenarios.yaml
      Note: canonical benchmark scenario documentation to expand
    - Path: scripts/bench-matrix.sh
      Note: matrix runner for scenario/rate/repeat sweeps
    - Path: scripts/bench-vegeta.sh
      Note: single-run benchmark harness to extend with anti-overfit scenarios
    - Path: ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/scripts/01-run-multi-vm-vegeta.sh
      Note: multi-VM benchmark harness for mixed-site tests
ExternalSources: []
Summary: Detailed benchmark plan to prevent overfitting goja-site optimization work to one Kanban fixture, with prioritized build/run order.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Anti-overfit benchmark plan for goja-site performance work

The Kanban simplification produced a large improvement. That result is real, but it came from one synthetic board shape: 120 cards, three columns, server-rendered UI DSL, and a particular interaction model. The next benchmarking work should prevent overfitting to that one fixture. The goal is to characterize how `goja-site` behaves across workload classes: small sites, large render trees, database-heavy sites, many VMs, long-running processes, browser interaction, errors, and observability modes.

A good benchmark suite should answer two questions at the same time:

1. Did a change improve the known bottleneck?
2. Did the change preserve or improve behavior for other plausible hosted-site shapes?

The benchmark set below is organized by behavior class, then followed by a prioritized implementation/run order.

## 1. Benchmark points to cover

### 1.1 Size-scaling benchmarks

Size-scaling benchmarks vary the size of the same logical UI. They answer whether the renderer degrades gradually or hits a knee when the number of nodes, attributes, strings, or response bytes crosses a threshold.

Recommended Kanban size dimensions:

```text
cards:   10, 50, 120, 500, 1000
columns: 3, 6, 12
```

Keep the card renderer constant while changing the data size. This isolates data volume from component shape.

Questions to answer:

- Is render cost proportional to card count?
- Does column count create independent overhead?
- At what size does response writing dominate rendering?
- Does GC cost grow smoothly or jump at specific sizes?
- Are p95 and p99 stable at medium rates?

Suggested scenarios:

```text
kanban-fragment-10
kanban-fragment-120
kanban-fragment-500
kanban-fragment-1000
kanban-action-10
kanban-action-120
kanban-action-500
```

Metrics to compare:

- throughput ratio,
- p50/p95/p99/max,
- response bytes,
- `goja_site_kanban_rendered_html_bytes`,
- `goja_site_kanban_render_duration_seconds`,
- heap allocation and GC metrics,
- pprof for one healthy and one degraded large case.

### 1.2 UI shape benchmarks

The current Kanban fixture stresses repeated card markup. Other hosted sites may stress different parts of `ui.dsl`. Shape benchmarks should isolate rendering patterns.

Recommended shapes:

| Scenario | Description | What it stresses |
|---|---|---|
| `render-flat-1000` | One parent with 1000 sibling child nodes. | Loop overhead, buffer writes, text rendering. |
| `render-deep-100` | A 100-level nested tree. | Recursive rendering depth and stack behavior. |
| `render-attrs-1000` | Many nodes with many attributes. | `renderAttrs`, map iteration, escaping, sorting if present. |
| `render-text-large` | Fewer nodes with large text bodies. | string escaping and response bytes. |
| `render-conditional-null` | Many conditional null/undefined children. | normalization and skipped-node behavior. |
| `render-table-100x20` | Table-like shape with rows and cells. | Wide structured output and repeated attrs. |

Questions to answer:

- Is `uidsl.renderAttrs` still a hotspot after Kanban simplification?
- Are deep trees significantly worse than wide trees?
- Do many attributes cost more than many text nodes?
- Does string escaping appear in pprof for large text content?
- Are there low-level renderer optimizations that help many shapes, not just Kanban?

### 1.3 Interaction benchmarks

Interaction benchmarks separate initial render from incremental behavior. This matters because an interactive hosted site may render once and then perform many small actions.

Recommended interactions:

```text
initial page render
fragment refresh
action with full refresh
action with no refresh
action with semantic event response
search/filter
card action menu open/close
drag/drop action
```

The current action path still returns refreshed HTML when the action callback returns `refresh: true`. The next architecture experiment may add semantic event responses. Benchmarks should compare both.

Questions to answer:

- How much does full refresh cost after precise form removal?
- How much would semantic event responses reduce action latency and response bytes?
- Does the frontend action menu introduce measurable browser overhead?
- Does drag/drop have different server cost than action-menu movement?

Server-side benchmark examples:

```text
kanban-action-refresh-120
kanban-action-no-refresh-120
kanban-action-events-120
```

Browser-side benchmark examples:

```text
open action menu on 120-card board
open action menu on 1000-card board
move card via keyboard action menu
move card via drag/drop
```

### 1.4 Database-heavy benchmarks

The render optimization may reveal database bottlenecks that were previously hidden. Database-heavy benchmarks should distinguish query cost, result size, write cost, transaction behavior, and guard overhead.

Recommended scenarios:

| Scenario | Description |
|---|---|
| `db-read-1` | Read one row. |
| `db-read-100` | Read 100 rows. |
| `db-read-1000` | Read 1000 rows. |
| `db-write-1` | Insert one row per request. |
| `db-write-batch-10` | Insert 10 rows per request in one transaction. |
| `db-write-batch-100` | Insert 100 rows per request in one transaction. |
| `db-guarded-write` | Same writes under guarded DB policy. |
| `db-render-table-1000` | Query 1000 rows and render them as HTML. |

Questions to answer:

- When rendering is cheap, does SQLite become the next bottleneck?
- How expensive is `db.guard` under write-heavy traffic?
- Does batching improve throughput?
- Does rendering query results dominate query execution?
- Do request-parented DB spans remain correct under load?

Instrumentation to inspect:

```text
goja_site_db_operations_total
goja_site_db_operation_duration_seconds
goja_site_db_errors_total
goja_site_db_guard_* metrics
OpenTelemetry DB spans
SQLite file size and WAL behavior
```

### 1.5 Multi-site and multi-VM benchmarks

The `serve-multi` work showed that multiple Goja VMs can run in parallel, but it only tested identical site types. The next multi-VM benchmarks should use mixed site shapes and larger VM counts.

Recommended dimensions:

```text
vm_count: 1, 4, 8, 16, 32, 64
traffic:  one-hot, even-hot, skewed, mixed-hot
site_mix: null, render, db, kanban
```

Traffic distributions:

| Distribution | Meaning | Question |
|---|---|---|
| `one-hot` | One hot site, many idle loaded VMs. | Do idle VMs add memory/GC overhead? |
| `even-hot` | All sites receive traffic evenly. | Does throughput scale with VM count? |
| `skewed` | One site receives most traffic, others receive a trickle. | Does one hot site affect tail latency for others? |
| `mixed-hot` | Different site types all receive traffic. | Does an expensive Kanban site degrade cheap sites? |

Questions to answer:

- Does memory from many idle VMs affect the hot site?
- Does one expensive site increase p95 for cheap sites?
- Does per-site metrics labeling remain useful and bounded?
- Does GC become process-wide enough to affect unrelated VMs?
- What VM count is safe for one process before memory or startup time becomes a problem?

### 1.6 Startup and reload benchmarks

The existing stress tests mostly use already-started VMs. Production hosting also needs startup and reload data.

Recommended measurements:

```text
process start -> listener ready
process start -> first successful request
script load time for one site
script load time for N sites
serve-multi config with 1, 4, 16, 64 sites
reload one site
reload all sites
```

Questions to answer:

- Is startup time linear in site count?
- Does script complexity dominate startup?
- How expensive is creating many Goja runtimes?
- Can a site be reloaded without disrupting unrelated sites?
- Does first request perform extra work compared with warm requests?

Metrics and artifacts:

- wall-clock startup time,
- first request latency,
- script count and script byte size,
- VM count,
- memory after startup,
- logs and readiness timestamps.

### 1.7 Memory and soak benchmarks

Short tests find obvious saturation but miss slow leaks and drift. Soak tests should run at safe rates below the knee.

Recommended soaks:

```text
30 min kanban-fragment at safe rate
30 min kanban-action at safe rate
2h mixed site workload
many sessions over time
site reload loop
write-heavy DB loop
```

Track:

```text
process_resident_memory_bytes
go_memstats_heap_alloc_bytes
go_memstats_heap_objects
go_gc_duration_seconds
go_goroutines
go_threads
SQLite file size
p95 drift
error counts
response bytes
```

Questions to answer:

- Does heap stabilize after warmup?
- Do goroutines leak?
- Does latency drift upward over time?
- Does SQLite or WAL size grow unexpectedly?
- Does GC frequency increase throughout the run?

### 1.8 Browser and frontend benchmarks

The accessibility work moved interaction mechanics to the frontend. Server benchmarks should be paired with browser checks so we do not simply move costs from server to browser.

Recommended browser measurements:

```text
large board DOM load time
time to first keyboard focus
action menu open latency
keyboard menu navigation latency
DOM update after action response
full root replacement time
semantic event update time, once implemented
```

Questions to answer:

- Is the action menu fast on 500-card and 1000-card boards?
- Does full root replacement cause visible browser delay?
- Does focus restoration remain reliable after large DOM replacements?
- Does semantic event application outperform full HTML replacement?
- Are there console warnings/errors under repeated interactions?

Automation approach:

- Playwright for correctness and coarse timing,
- browser `performance.now()` for local interaction timings,
- console and pageerror capture,
- optional trace viewer artifacts for slow cases.

### 1.9 Error and pathological benchmarks

Error paths should be cheap and safe. They must also preserve bounded observability labels.

Recommended scenarios:

| Scenario | Purpose |
|---|---|
| unknown host flood | Verify unknown-host handling and bounded labels. |
| 404 route flood | Verify cheap not-found path. |
| invalid action payload | Verify action validation and error path. |
| action callback throws | Verify JS exception handling. |
| DB locked errors | Verify DB error classes and request behavior. |
| huge request body | Verify limits and memory behavior. |
| slow JS action callback | Verify queueing and cancellation behavior. |

Questions to answer:

- Are error paths cheaper than successful expensive paths?
- Do metrics labels remain bounded?
- Do traces avoid raw error strings and request bodies?
- Can one bad site degrade another site?
- Are failures visible without creating high-cardinality telemetry?

### 1.10 Observability overhead benchmarks

Instrumentation must be measured as part of the product. The observability spine should be cheap enough for production metrics, and tracing should be sampled intentionally.

Recommended modes:

```text
metrics off
metrics on
metrics + pprof endpoints enabled but idle
metrics + tracing sampled 1%
metrics + tracing sampled 10%
metrics + tracing sampled 100%
```

Questions to answer:

- What is the cost of Prometheus instrumentation?
- Does enabling the diagnostics listener change throughput?
- What trace sampling rate is acceptable?
- Is 100% tracing only suitable for short diagnostics?
- Do pprof endpoints have no cost when idle?

Benchmark the same scenarios under each mode:

```text
null
render-flat-1000
db-read-100
kanban-fragment-120
kanban-action-120
```

## 2. Prioritized build and run order

This section orders the benchmark work by implementation effort, instrumentation effort, and value. The intent is to build useful coverage quickly before adding more complex experiments.

### Priority 1: Keep and formalize post-simplification Kanban regression benchmarks

Build effort: low.  
Instrumentation effort: already done.  
Value: very high.

These benchmarks protect the performance win already achieved.

Run regularly:

```text
single kanban-fragment 200/s
single kanban-action 100/s
multi 4VM kanban-fragment 400/s
```

Why first:

- They are already scripted.
- They caught the largest known bottleneck.
- They provide a fast regression guard against accidentally reintroducing large per-card markup.

Next implementation step:

- Turn `scripts/02-run-post-simplification-benchmarks.sh` into a stable regression script or Makefile target.
- Add threshold checks for p95, throughput ratio, and response bytes.

### Priority 2: Kanban size-scaling matrix

Build effort: low-medium.  
Instrumentation effort: low.  
Value: very high.

Create parameterized Kanban fixtures or generate fixture directories for:

```text
10, 120, 500, 1000 cards
3 and 6 columns
```

Why second:

- It directly checks whether the current optimization generalizes beyond the 120-card benchmark.
- It is easy to build by adapting `bench/scripts/kanban-board/app.js`.
- It informs whether the next step should be renderer optimization, response protocol changes, or browser-side updates.

Run order:

```text
kanban-fragment-10 at 200/s
kanban-fragment-120 at 200/s
kanban-fragment-500 at 100/s
kanban-fragment-1000 at 50/s
```

Then add saturation rates around the first degraded size.

### Priority 3: UI shape renderer micro-matrix

Build effort: medium.  
Instrumentation effort: low.  
Value: high.

Add fixture directories for non-Kanban UI shapes:

```text
render-flat-1000
render-deep-100
render-attrs-1000
render-text-large
render-table-100x20
```

Why third:

- It prevents optimizing only Kanban.
- It identifies renderer-level work that helps all hosted sites.
- It may reveal whether `renderAttrs`, string escaping, or node recursion remains a general bottleneck.

Run order:

```text
25/s, 100/s, 250/s
repeat 3
```

Capture pprof only for the worst shape.

### Priority 4: Database-heavy matrix

Build effort: medium.  
Instrumentation effort: already mostly done.  
Value: high.

Extend the DB fixture to cover:

```text
db-read-1
db-read-100
db-read-1000
db-write-batch-10
db-write-batch-100
db-render-table-1000
guarded-db-write
```

Why fourth:

- Rendering is now much cheaper for Kanban; DB may become visible in realistic sites.
- DB spans and metrics already exist.
- Guard overhead needs real load data.

Run after UI shape benchmarks because DB fixture work is slightly more involved.

### Priority 5: Mixed multi-site matrix

Build effort: medium-high.  
Instrumentation effort: mostly done.  
Value: high.

Use `serve-multi` with heterogeneous sites:

```text
site-001 null
site-002 render-flat-1000
site-003 db-read-100
site-004 kanban-fragment-120
```

Run distributions:

```text
even-hot
skewed
one expensive site + cheap sites
```

Why fifth:

- It answers production-hosting questions better than identical-site tests.
- It checks whether expensive sites affect cheap sites through CPU, GC, or response writing.
- It validates per-site metrics under mixed workloads.

### Priority 6: Browser interaction benchmarks

Build effort: medium-high.  
Instrumentation effort: medium.  
Value: medium-high.

Extend the Playwright accessibility smoke into a timing benchmark:

```text
action menu open latency on 120 cards
action menu open latency on 500 cards
action menu open latency on 1000 cards
keyboard move latency
focus restoration timing
full root replacement timing
```

Why sixth:

- The frontend now owns more interaction behavior.
- We need to avoid moving bottlenecks from server to browser.
- Browser timing is noisier and harder to automate reliably, so it comes after server-side coverage.

### Priority 7: Observability overhead matrix

Build effort: medium.  
Instrumentation effort: medium.  
Value: medium-high.

Run a small scenario set under different observability modes:

```text
metrics off
metrics on
metrics + idle pprof
metrics + tracing 1%
metrics + tracing 100%
```

Why seventh:

- Important for production decisions.
- Less urgent than functional workload coverage because metrics have already been cheap enough in current tests.
- Requires careful control to avoid noisy conclusions.

### Priority 8: Startup and reload benchmarks

Build effort: medium-high.  
Instrumentation effort: medium.  
Value: medium.

Measure:

```text
single site startup
serve-multi 4 sites startup
serve-multi 16 sites startup
serve-multi 64 sites startup
first request latency
script reload time when reload exists
```

Why eighth:

- Startup matters for hosting operations.
- It is less urgent than request-path performance because the current bottlenecks were request-path costs.
- Reload semantics may need more product design before benchmarking is meaningful.

### Priority 9: Memory and soak benchmarks

Build effort: medium.  
Instrumentation effort: medium-high.  
Value: high but slower feedback.

Run only after the shorter matrices establish safe rates.

Suggested first soak:

```text
kanban-fragment-500 at safe rate for 30 minutes
mixed multi-site at safe rate for 30 minutes
```

Why ninth:

- Soaks are valuable but time-consuming.
- They should use rates derived from earlier saturation tests.
- They are better for leak detection than for initial optimization discovery.

### Priority 10: Error/pathological matrix

Build effort: medium.  
Instrumentation effort: low-medium.  
Value: medium.

Add after normal workloads are covered:

```text
unknown host flood
invalid action payload
action callback throws
DB locked
huge request body
slow JS callback
```

Why last:

- These are important hardening tests.
- They are less likely to inform the immediate performance architecture unless they reveal isolation problems.
- They should be built once normal benchmark infrastructure is stable.

## 3. Recommended immediate next benchmark deliverable

The next benchmark package should be small and high value:

```text
Anti-overfit matrix v1:
  kanban-fragment-10
  kanban-fragment-120
  kanban-fragment-500
  render-flat-1000
  render-attrs-1000
  db-read-100
  db-write-batch-10

rates:
  25/s, 100/s, 250/s

repeat:
  3
```

This gives coverage across:

- Kanban size scaling,
- generic UI rendering,
- attribute-heavy rendering,
- database reads,
- database writes.

It is small enough to build without a large framework change and broad enough to detect overfitting to one Kanban fixture.

## 4. Rule for accepting future optimizations

A future optimization should be accepted as generally valuable only when it improves or preserves behavior across at least two workload dimensions.

Examples of acceptable evidence:

```text
reduces Kanban p95 and reduces render-attrs p95
reduces response bytes and reduces browser DOM update time
improves single-VM action and multi-VM mixed-site throughput
reduces CPU pprof hotspot and does not increase DB-heavy latency
```

Avoid accepting optimizations that only improve one synthetic fixture while making other shapes worse. The benchmark suite should make that visible.
