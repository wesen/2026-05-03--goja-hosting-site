---
Title: Investigation Diary
Ticket: GOJA-PERF-BENCH
Status: active
Topics:
    - goja
    - performance
    - benchmarking
    - stress-testing
    - observability
DocType: reference
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
    - Path: ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/design-doc/02-goja-site-production-observability-metrics-and-tracing-guide.md
      Note: New observability guide added in Step 5
ExternalSources: []
Summary: Chronological diary for the GOJA-PERF-BENCH documentation and benchmark design investigation.
LastUpdated: 2026-05-14T13:37:52.761406421-04:00
WhatFor: Use this to understand how the benchmark design document was produced and how to continue implementation.
WhenToUse: When reviewing the performance benchmark plan, validating commands, or picking up the next implementation phase.
---









# Diary

## Goal

This diary records the creation of the `GOJA-PERF-BENCH` docmgr ticket and the investigation behind the stress-test, benchmark, and performance measurement design for `goja-site`.

## Step 1: Create the ticket workspace

I created a new docmgr ticket dedicated to stress testing and benchmarking the complete `goja-site` hosting concept. The ticket is intentionally scoped as a measurement and implementation-guide ticket, not as an optimization ticket, because the codebase needs baselines before performance changes are proposed.

The generated workspace created the normal ticket files and separate documents for the primary design guide and the investigation diary.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to stress test and benchmark and measure performance of the whole goja hosting concept. 

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.

Keep a detailed diary."

**Assistant interpretation:** Create a structured ticket with a deep intern-facing design and implementation guide for benchmarking `goja-site`, keep an investigation diary, validate the documentation, and upload the final bundle to reMarkable.

**Inferred user intent:** The user wants a durable planning artifact that lets a new engineer implement a performance benchmark suite without needing repeated oral context.

**Commit (code):** N/A — documentation-only work in this step.

### What I did

- Ran `pwd && git status --short && docmgr status --summary-only`.
- Confirmed the repository root is `/home/manuel/code/wesen/2026-05-03--goja-hosting-site`.
- Observed an existing untracked `tmp/` directory before starting.
- Created ticket `GOJA-PERF-BENCH` with topics `goja,performance,benchmarking,stress-testing,observability`.
- Created the primary design document.
- Created this investigation diary document.

Commands run:

```text
docmgr ticket create-ticket --ticket GOJA-PERF-BENCH --title "Stress test benchmark and performance measurement plan for goja hosting" --topics goja,performance,benchmarking,stress-testing,observability

docmgr doc add --ticket GOJA-PERF-BENCH --doc-type design-doc --title "Goja Hosting Stress Test Benchmark and Performance Guide"

docmgr doc add --ticket GOJA-PERF-BENCH --doc-type reference --title "Investigation Diary"
```

### Why

- The user explicitly requested a new docmgr ticket.
- A separate diary makes the investigation reviewable and continuation-friendly.
- The ticket topics make the document discoverable through docmgr search and validation.

### What worked

- The ticket was created under `ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting`.
- `docmgr doc add` created both the design-doc and reference documents with frontmatter.

### What didn't work

- N/A for this step.

### What I learned

- The repository already uses docmgr heavily; `docmgr status --summary-only` reported 7 existing tickets and 17 docs before this ticket was added.
- The worktree already had untracked `tmp/`, so later status output must not treat that as a new artifact from this documentation task.

### What was tricky to build

The main subtlety was choosing a ticket id that is broad enough for a full benchmark program but specific enough to be searchable. `GOJA-PERF-BENCH` was chosen because it names the subsystem (`GOJA`), outcome (`PERF`), and artifact class (`BENCH`).

### What warrants a second pair of eyes

- Whether `GOJA-PERF-BENCH` should become the long-term canonical ticket for all performance work, or whether future implementation should split into smaller tickets.

### What should be done in the future

- Implement the harness described by the design guide in smaller follow-up commits/tickets.

### Code review instructions

- Start with the ticket index and this diary.
- Verify that generated ticket paths and topics are correct.
- No code behavior changed in this step.

### Technical details

Ticket path:

```text
ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting
```

## Step 2: Inspect the runtime architecture and existing validation surface

I inspected the code paths that define the hosting concept: single-site server construction, multi-site dispatch, config normalization, database module wiring, guarded database execution, Kanban route mounting, CLI entrypoints, and existing tests/scripts. This gave the design guide concrete file references instead of speculative architecture descriptions.

The most important conclusion is that one `Server` maps to one Goja runtime and one SQLite handle, while `MultiServer` maps normalized Host headers to separate `Server` instances. That means performance work needs both per-request latency benchmarks and per-site startup/memory benchmarks.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Gather enough implementation evidence to write an intern-facing benchmark design grounded in current source files.

**Inferred user intent:** The guide should teach the system accurately, including how requests flow and where benchmark probes should be placed.

**Commit (code):** N/A — investigation-only work in this step.

### What I did

- Listed ticket directories and repository files.
- Searched for relevant terms across `cmd`, `pkg`, `examples`, `sites`, `deploy`, docs, and `go.mod`.
- Read key source files:
  - `pkg/app/server.go`
  - `pkg/app/multi_server.go`
  - `pkg/app/multi_config.go`
  - `pkg/app/database.go`
  - `pkg/dbguard/metered.go`
  - `pkg/kanbanddsl/mount.go`
  - `pkg/app/multi_server_test.go`
  - `pkg/app/database_test.go`
  - `cmd/goja-site/serve.go`
  - `cmd/goja-site/main.go`
  - `deploy/sites.local.yaml`
  - `scripts/playwright-kanban-smoke.sh`
- Captured line-numbered excerpts with `nl -ba` for evidence-backed references.

Representative commands:

```text
find ttmp/2026/05 -maxdepth 3 -type d | sort

rg -n "goja|require|http|Serve|DB|sqlite|handler|middleware|multi|site|script|kanban|benchmark|pprof" cmd pkg examples sites deploy go.mod README.md -S || true

nl -ba pkg/app/server.go | sed -n '22,158p'
nl -ba pkg/app/multi_server.go | sed -n '13,103p'
nl -ba pkg/app/multi_config.go | sed -n '14,121p'
nl -ba pkg/app/database.go | sed -n '16,82p'
nl -ba pkg/kanbanddsl/mount.go | sed -n '9,98p'
nl -ba go.mod | sed -n '1,80p'
```

### Why

- The requested guide needed file references, API references, diagrams, and pseudocode.
- Performance recommendations should be anchored to actual code paths.
- The benchmark plan must distinguish startup, route dispatch, Goja execution, rendering, SQLite, guarded writes, and multi-site host dispatch.

### What worked

- The architecture boundaries were clear in the code:
  - `pkg/app/server.go` owns one site.
  - `pkg/app/multi_server.go` dispatches by Host header.
  - `pkg/app/database.go` selects database policies and module names.
  - `pkg/kanbanddsl/mount.go` shows the realistic Kanban fragment/action request paths.
- Existing tests already provide patterns for temporary script directories and `httptest` requests.
- `scripts/playwright-kanban-smoke.sh` already shows robust shell cleanup and process-management patterns for external tests.

### What didn't work

- No existing benchmark files were found, so the guide had to propose a new harness layout rather than extend a current benchmark suite.
- The repository does not currently expose pprof or metrics flags, so those are proposed as optional future instrumentation.

### What I learned

- `NewServer` loads scripts at startup through `s.runtime.Owner.Call` and `vm.RunScript`.
- `NewMultiServer` constructs one `Server` per configured site, so multi-site startup and memory should scale roughly with site count plus shared process overhead.
- `db.guard` currently wraps writes, not reads, through `MeteredDB.Exec`.
- Kanban action routes can combine JavaScript dispatch, database work, UI DSL render, and JSON response generation in a single request.

### What was tricky to build

The tricky part was separating observed behavior from inferred benchmark implications. For example, the code proves that multi-site mode owns a `map[string]*Server` and creates one server per site. It does not by itself prove the memory-per-site slope. The guide therefore states that startup/memory slope must be measured rather than assuming a value.

### What warrants a second pair of eyes

- Whether the guide's proposed scenario matrix covers the real production bottlenecks the team cares about most.
- Whether pprof should be implemented in the CLI package, the app package, or an external wrapper.

### What should be done in the future

- Add benchmark scripts and `testing.B` benchmarks exactly where the guide proposes.
- Capture real numbers for null route, UI render, DB read/write, Kanban action, and multi-site scenarios.

### Code review instructions

- Review the line references in the design guide against the current files.
- Pay special attention to `pkg/app/server.go`, `pkg/app/multi_server.go`, and `pkg/kanbanddsl/mount.go`, because these shape the request-flow diagrams and measurement matrix.
- Validate with `go test ./...`.

### Technical details

Important observed source anchors:

- `pkg/app/server.go:22-28`: `Server` fields.
- `pkg/app/server.go:31-87`: `NewServer` startup path.
- `pkg/app/server.go:141-158`: script loading path.
- `pkg/app/multi_server.go:57-69`: multi-site request dispatch path.
- `pkg/app/database.go:18-53`: database module wiring.
- `pkg/kanbanddsl/mount.go:45-88`: Kanban action route behavior.

## Step 3: Run baseline tests and write the deliverables

I ran the repository test suite before finalizing the guide, then wrote the design document, updated the task list, and replaced the generated placeholder diary with this chronological record. The design guide is intentionally implementation-oriented: it explains the current system, shows request-flow diagrams, defines benchmark scenarios, gives pseudocode, proposes file layout, and lists acceptance criteria for each phase.

The guide does not claim measured performance numbers yet. It defines the harness needed to produce them, because the current repository lacks a benchmark suite.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Produce the actual ticket deliverables and keep validation evidence.

**Inferred user intent:** The final ticket should be useful to an intern without additional context from the assistant session.

**Commit (code):** N/A — documentation-only work in this step.

### What I did

- Ran the full Go test suite.
- Wrote `tasks.md` with completed documentation tasks and future implementation tasks.
- Wrote `design-doc/01-goja-hosting-stress-test-benchmark-and-performance-guide.md`.
- Wrote `reference/01-investigation-diary.md`.

Validation command and output:

```text
go test ./...
?   	github.com/go-go-golems/goja-site/cmd/goja-site	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/app	0.051s
ok  	github.com/go-go-golems/goja-site/pkg/dbguard	0.043s
?   	github.com/go-go-golems/goja-site/pkg/doc	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/kanbanddsl	0.014s
```

### Why

- Running tests ensures the repository was healthy while the documentation was produced.
- Updating tasks makes the ticket actionable after the planning deliverable is complete.
- The design guide needs to be long-form because it is explicitly for a new intern.

### What worked

- `go test ./...` passed.
- The design guide now contains:
  - executive summary,
  - problem statement and scope,
  - new-intern system orientation,
  - ASCII diagrams,
  - evidence-backed current-state analysis,
  - benchmark matrix,
  - scenario scripts,
  - Go benchmark pseudocode,
  - CLI/API references,
  - phased implementation plan,
  - validation strategy,
  - risks and open questions,
  - file references.

### What didn't work

- No actual stress tests were run because the ticket asks for a design/implementation guide and the repository does not yet contain the harness. The guide clearly marks real baseline measurements as a follow-up task.

### What I learned

- The best first implementation step is not pprof or a load generator; it is deterministic benchmark fixtures and small `testing.B` benchmarks.
- The existing Playwright script is a useful model for process cleanup but should not become the main performance driver.

### What was tricky to build

The document had to be detailed enough for an intern while avoiding fake precision. I avoided inventing latency numbers, throughput numbers, or memory slopes. Instead, I specified exactly how those numbers should be produced and what metadata must be recorded.

### What warrants a second pair of eyes

- Whether the proposed initial benchmark set is too broad for one intern sprint.
- Whether the chosen default external load tool should be standardized before implementation begins.
- Whether diagnostics flags should expose pprof through manually mounted handlers or the default `net/http/pprof` registration pattern.

### What should be done in the future

- Implement `bench/scripts` fixtures.
- Add `pkg/app/*_bench_test.go` and `pkg/dbguard/*_bench_test.go`.
- Add external benchmark scripts and result summarization.
- Run on production-like hardware and append measured baseline reports to this ticket.

### Code review instructions

- Start with the primary design doc:
  - `ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/design-doc/01-goja-hosting-stress-test-benchmark-and-performance-guide.md`
- Then read this diary:
  - `ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/reference/01-investigation-diary.md`
- Validate repository health with:

```text
go test ./...
```

### Technical details

Future implementation should preserve this distinction:

```text
Correctness smoke:
  go test ./...
  scripts/playwright-kanban-smoke.sh

Fast performance feedback:
  go test ./pkg/app -run '^$' -bench . -benchmem
  go test ./pkg/dbguard -run '^$' -bench . -benchmem

Realistic load feedback:
  compiled goja-site binary + external load tool + pprof artifacts
```

## Step 4: Validate docmgr metadata and upload the bundle to reMarkable

I validated the new ticket with `docmgr doctor`, fixed vocabulary warnings by adding the missing topic slugs, and uploaded a bundled PDF to reMarkable. The uploaded bundle includes the ticket index, tasks, changelog, primary design guide, and diary.

This step turns the local ticket deliverables into the requested reMarkable reading artifact.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Validate the completed docmgr ticket and deliver the documentation bundle to reMarkable.

**Inferred user intent:** The user wants the guide available both in the repository ticket and on the reading device.

**Commit (code):** N/A — documentation and upload-only work in this step.

### What I did

- Ran `docmgr doctor --ticket GOJA-PERF-BENCH --stale-after 30`.
- Saw warnings for unknown topics: `benchmarking`, `observability`, `performance`, and `stress-testing`.
- Added those topics to `ttmp/vocabulary.yaml` using `docmgr vocab add`.
- Re-ran `docmgr doctor --ticket GOJA-PERF-BENCH --stale-after 30`; it passed.
- Ran a reMarkable dry-run bundle upload.
- Ran the real reMarkable bundle upload.

Important command output:

```text
## Doctor Report (1 findings)

### GOJA-PERF-BENCH

- ✅ All checks passed
```

Dry-run upload output:

```text
DRY: bundle name=GOJA PERF BENCH Guide
DRY: remote-dir=/ai/2026/05/14/GOJA-PERF-BENCH
DRY: pandoc <bundle> -> <tmp>/GOJA_PERF_BENCH_Guide.pdf
DRY: upload GOJA_PERF_BENCH_Guide.pdf -> /ai/2026/05/14/GOJA-PERF-BENCH
```

Real upload output:

```text
OK: uploaded GOJA_PERF_BENCH_Guide.pdf -> /ai/2026/05/14/GOJA-PERF-BENCH
```

### Why

- `docmgr doctor` catches stale or invalid metadata before the ticket becomes a long-lived reference.
- The vocabulary additions make the new performance topics reusable by future tickets.
- The upload fulfills the user's request to send the bundle to reMarkable.

### What worked

- The docmgr warning was straightforward to fix with vocabulary additions.
- The reMarkable dry-run showed the expected input files and destination.
- The reMarkable upload succeeded.

### What didn't work

- The first `docmgr doctor` run did not pass because the ticket introduced new topics that were not in the vocabulary yet.

### What I learned

- New topic slugs should be added before doctor validation when creating a ticket outside the existing vocabulary.
- The upload bundle name `GOJA PERF BENCH Guide` is sanitized by the upload tool to `GOJA_PERF_BENCH_Guide.pdf`.

### What was tricky to build

The only tricky part was preserving validation evidence after the upload. This diary step records the upload output so future readers can see where the reMarkable artifact went without needing a separate cloud listing.

### What warrants a second pair of eyes

- Whether the new vocabulary descriptions are sufficiently general for future performance and observability tickets.

### What should be done in the future

- If the guide changes materially, re-upload the bundle with `--force` so reMarkable has the latest copy.

### Code review instructions

- Check `ttmp/vocabulary.yaml` for the four new topic entries.
- Re-run:

```text
docmgr doctor --ticket GOJA-PERF-BENCH --stale-after 30
```

- Confirm the upload path in the final handoff:

```text
/ai/2026/05/14/GOJA-PERF-BENCH/GOJA_PERF_BENCH_Guide.pdf
```

### Technical details

The uploaded bundle contains:

- ticket index,
- task list,
- changelog,
- primary benchmark guide,
- investigation diary.

## Step 5: Add a dedicated production observability guide

After the user asked whether Prometheus metrics and Jaeger-style tracing would make sense, I added a second design document to the same ticket. This guide treats metrics and tracing as production observability features that also improve benchmark analysis, rather than as benchmark-only instrumentation.

The guide recommends Prometheus metrics first, pprof diagnostics on a private listener, and OpenTelemetry tracing later with OTLP export so Jaeger can be used through an OpenTelemetry Collector without hard-coding Jaeger APIs into the application.

### Prompt Context

**User prompt (verbatim):** "Ok, create a dedicate analysis / implementation guide in the ticket as well, like the first one, and upload to remarkable when done. continue keeping a diary."

**Assistant interpretation:** Add a second long-form intern-oriented ticket document focused on production observability with Prometheus metrics and OpenTelemetry/Jaeger tracing, update the diary, validate, and re-upload the bundle to reMarkable.

**Inferred user intent:** The user wants observability to become a first-class part of the goja-site performance plan and wants the updated ticket available on reMarkable.

**Commit (code):** N/A — documentation-only work in this step.

### What I did

- Searched the repository for existing observability, Prometheus, OpenTelemetry, Jaeger, pprof, metrics, tracing, and logging references.
- Confirmed there is currently Glazed logging setup but no existing Prometheus or OTel instrumentation.
- Created a new design doc:
  - `design-doc/02-goja-site-production-observability-metrics-and-tracing-guide.md`
- Captured additional line-numbered source anchors for CLI flags, multi-site CLI, root logging, and `db.guard` internals.
- Updated `tasks.md` with observability follow-up tasks.

Commands run:

```text
rg -n "prometheus|otel|opentelemetry|jaeger|pprof|metrics|expvar|trace|slog|log" cmd pkg go.mod deploy .github -S || true

docmgr doc add --ticket GOJA-PERF-BENCH --doc-type design-doc --title "Goja Site Production Observability Metrics and Tracing Guide"

nl -ba cmd/goja-site/serve.go | sed -n '18,84p'
nl -ba cmd/goja-site/serve_multi.go | sed -n '18,70p'
nl -ba cmd/goja-site/main.go | sed -n '17,68p'
nl -ba pkg/dbguard/guard.go | sed -n '15,75p;170,205p;240,310p'
```

### Why

- Prometheus metrics are useful for both stress tests and live multi-site production monitoring.
- Tracing should be designed early enough that metrics labels, route classification, and DB/Kanban instrumentation boundaries do not conflict with future spans.
- The first benchmark guide mentioned observability, but the subject is large enough to warrant its own implementation guide.

### What worked

- The repository search showed no existing Prometheus or OpenTelemetry integration, so the new guide could propose a clean package boundary under `pkg/observability`.
- Existing code has good instrumentation boundaries:
  - `pkg/app/server.go` for single-site lifecycle and HTTP handler wrapping.
  - `pkg/app/multi_server.go` for host dispatch and unknown host metrics.
  - `pkg/app/database.go` and `pkg/dbguard/metered.go` for SQL timings.
  - `pkg/dbguard/guard.go` for guard size/limit metrics.
  - `pkg/kanbanddsl/mount.go` for fragment/action/render timings.

### What didn't work

- There is no existing metrics abstraction or tracing provider to extend, so implementation will require new dependencies and a new package.
- The current route host does not obviously expose registered route patterns from outside, so the first metrics implementation may need a coarse route classifier until route patterns are available.

### What I learned

- Glazed logging is already initialized at the Cobra root, but that does not solve metrics/tracing.
- `serve` and `serve-multi` need parallel diagnostics flags because both commands can run production servers.
- `db.guard` already measures file, WAL, SHM, page size, page count, and freelist count, which maps naturally to Prometheus gauges.

### What was tricky to build

The main design constraint is cardinality. The guide had to be explicit that raw URL paths, raw Host headers, SQL text, session IDs, request bodies, and arbitrary JavaScript error strings must not become labels or span attributes. The implementation should start with controlled route classes and SQL kinds, then improve only if route patterns can be exposed safely.

### What warrants a second pair of eyes

- Whether `pkg/observability` should own all Prometheus objects, or whether small observer interfaces should be added to `dbguard` and `kanbanddsl` to keep those packages decoupled.
- Whether pprof should share the metrics diagnostics listener or use a separate `--pprof-addr`.
- Which histogram bucket boundaries best fit `goja-site` workloads.

### What should be done in the future

- Implement the observability guide in phases, starting with metrics scaffolding and a private diagnostics listener.
- Add benchmark runs with observability disabled, metrics enabled, tracing sampled, and tracing at 100% for short controlled runs.
- Add example Prometheus scrape config, OTel Collector config, and starter Grafana dashboard.

### Code review instructions

- Review the new document:
  - `ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/design-doc/02-goja-site-production-observability-metrics-and-tracing-guide.md`
- Check the proposed metric names and labels for cardinality risks.
- Re-run docmgr validation after related-file metadata is updated:

```text
docmgr doctor --ticket GOJA-PERF-BENCH --stale-after 30
```

### Technical details

The observability guide proposes:

```text
Prometheus first:
  --metrics-addr 127.0.0.1:19090
  /metrics on a private diagnostics listener

pprof second:
  --pprof only on diagnostics listener

Tracing third:
  OpenTelemetry OTLP export
  Jaeger through OpenTelemetry Collector

Benchmark overhead dimension:
  observability=off
  metrics=on
  metrics+tracing_sample_1pct=on
  metrics+tracing_sample_100pct=short-run-only
```

## Step 6: Validate and re-upload the expanded bundle

I related the new observability guide to the source files it references, updated the ticket changelog, validated the ticket again, and re-uploaded the reMarkable bundle with the new guide included. The upload overwrote the previous bundle with `--force` because the bundle name and destination were intentionally kept stable.

This makes the reMarkable artifact a current complete packet: index, tasks, changelog, benchmark guide, observability guide, and diary.

### Prompt Context

**User prompt (verbatim):** (same as Step 5)

**Assistant interpretation:** Finish the ticket bookkeeping, validate the new document, and upload the updated bundle.

**Inferred user intent:** The user wants the same quality and delivery path as the first guide, including diary continuity.

**Commit (code):** N/A — documentation and upload-only work in this step.

### What I did

- Related the new observability guide to source files with `docmgr doc relate`.
- Related this diary to the new observability guide.
- Updated the ticket changelog.
- Ran `docmgr doctor --ticket GOJA-PERF-BENCH --stale-after 30`.
- Ran a reMarkable dry-run bundle upload that included the new `02-...observability...` document.
- Ran the real reMarkable upload with `--force`.

Validation output:

```text
## Doctor Report (1 findings)

### GOJA-PERF-BENCH

- ✅ All checks passed
```

Dry-run upload included:

```text
DRY: include .../design-doc/02-goja-site-production-observability-metrics-and-tracing-guide.md (title="02-goja-site-production-observability-metrics-and-tracing-guide")
DRY: upload GOJA_PERF_BENCH_Guide.pdf -> /ai/2026/05/14/GOJA-PERF-BENCH
```

Real upload output:

```text
OK: uploaded GOJA_PERF_BENCH_Guide.pdf -> /ai/2026/05/14/GOJA-PERF-BENCH
```

### Why

- Related-file metadata makes the new guide discoverable from source file relationships.
- Validation catches frontmatter and topic issues before handoff.
- Re-uploading with the same stable bundle name keeps the reMarkable destination simple.

### What worked

- `docmgr doctor` passed after the new document and relationships were added.
- The reMarkable dry-run confirmed the second design document was included.
- The `--force` upload succeeded.

### What didn't work

- N/A for this step.

### What I learned

- The stable bundle path can be reused for updated ticket packets as long as `--force` is used intentionally.
- The new observability guide did not introduce new vocabulary warnings because the earlier vocabulary additions already covered `observability` and `performance`.

### What was tricky to build

The main operational detail was remembering that the first uploaded PDF would otherwise remain stale. The second upload used the same name and destination with `--force` so the reMarkable copy reflects the newly added observability guide.

### What warrants a second pair of eyes

- Confirm that overwriting the existing reMarkable PDF is the preferred behavior for this ticket, since it can replace annotations on the old copy.

### What should be done in the future

- If future ticket docs are added, either upload a new dated bundle name or use `--force` again intentionally.

### Code review instructions

- Re-run:

```text
docmgr doctor --ticket GOJA-PERF-BENCH --stale-after 30
```

- Confirm the reMarkable handoff path:

```text
/ai/2026/05/14/GOJA-PERF-BENCH/GOJA_PERF_BENCH_Guide.pdf
```

### Technical details

The current uploaded bundle contains:

- ticket index,
- task list,
- changelog,
- `01-goja-hosting-stress-test-benchmark-and-performance-guide.md`,
- `02-goja-site-production-observability-metrics-and-tracing-guide.md`,
- `01-investigation-diary.md`.

## Step 7: Split the ticket into implementation phases

I converted the ticket task list from a short checklist into a phased implementation plan. The phases intentionally start with a minimal observability spine, then add load generation, then deepen database, guard, Kanban, pprof, tracing, and dashboard work.

This plan answers the sequencing question from the previous discussion: do a thin metrics foundation first, add load generation early, and then deepen observability and benchmarking together.

### Prompt Context

**User prompt (verbatim):** "alright, add detailed takss split into phases to the ticket, and then work on them one by one, keeping a detailed diary, and committing at appropriate intervals."

**Assistant interpretation:** Update the ticket tasks with detailed phases, then begin implementing those phases sequentially, maintaining diary entries and making focused commits.

**Inferred user intent:** The user wants the planning ticket to become an executable implementation backlog and wants the repository history to show reviewable checkpoints.

**Commit (code):** pending — documentation task split will be committed before code changes.

### What I did

- Rewrote `tasks.md` into eight phases:
  - Phase 0: ticket planning and documentation baseline.
  - Phase 1: minimal production observability spine.
  - Phase 2: load-generation MVP.
  - Phase 3: database and `db.guard` metrics.
  - Phase 4: Kanban metrics.
  - Phase 5: pprof capture automation in the load harness.
  - Phase 6: OpenTelemetry tracing.
  - Phase 7: benchmark scenarios and production dashboards.
- Kept already completed documentation work checked off.
- Left implementation tasks unchecked so progress can be tracked phase by phase.

### Why

- A phased checklist reduces ambiguity and keeps code review scope manageable.
- Committing the documentation plan before code work creates a clean baseline.
- The split reflects the decision to avoid both extremes: not all observability before benchmarking, and not blind load testing without metrics.

### What worked

- The task list now names concrete files, commands, validation steps, and commit points.
- The first implementation phase is small enough to start immediately: Prometheus registry, diagnostics listener, flags, basic HTTP/multi-site metrics, and tests.

### What didn't work

- N/A for this step.

### What I learned

- The observability and benchmarking work should be treated as one epic but several focused commits/PRs.
- Phase 1 should be production-quality enough to keep, but not so broad that it blocks load-generation work.

### What was tricky to build

The main trick was ordering. If the task list puts all observability before all benchmarking, it delays feedback. If it puts load generation first, early benchmarks lack internal measurements. The chosen ordering starts with a thin observability spine, then adds load generation, then deepens instrumentation where benchmark scenarios need it.

### What warrants a second pair of eyes

- Whether Phase 1 is still too large for one commit, especially if HTTP and multi-site metrics touch several app boundaries.
- Whether Phase 2 should start with Vegeta CLI only or immediately add a Go scenario runner.

### What should be done in the future

- Work through Phase 1 first.
- Commit the documentation baseline before implementation changes.
- Commit each phase or coherent sub-phase separately.

### Code review instructions

- Review `tasks.md` first to confirm phase ordering.
- Confirm that checked items reflect completed documentation only, not implementation that has not happened yet.

### Technical details

The immediate next implementation target is Phase 1:

```text
Prometheus dependency
pkg/observability package
--metrics-addr / --metrics-path / --pprof flags
private /metrics listener
basic HTTP and multi-site metrics
unit/integration tests
go test ./...
```

## Step 8: Implement Phase 1 minimal observability spine

I implemented the first code phase: a private Prometheus diagnostics listener, optional pprof mounting, basic HTTP metrics, multi-site dispatch metrics, low-cardinality route/status label helpers, and tests. This gives future load-generation work a metrics surface from the beginning instead of adding measurements after the fact.

The implementation intentionally keeps diagnostics off unless `--metrics-addr` is set. If `--pprof` is requested without a diagnostics address, startup returns an error so pprof cannot accidentally land on the public app listener.

### Prompt Context

**User prompt (verbatim):** (same as Step 7)

**Assistant interpretation:** Start working through the phased task list and commit coherent implementation chunks.

**Inferred user intent:** The user wants the plan converted into working code with reviewable commits and diary evidence.

**Commit (code):** 6657f2504ac07194c20c02b6fd934829513f4cc8 — "feat: add prometheus diagnostics foundation"

### What I did

- Added Prometheus client dependency.
- Added `pkg/observability` with:
  - config defaults,
  - isolated Prometheus registry,
  - Go/process collectors,
  - diagnostics server,
  - manual pprof mounting,
  - HTTP metrics middleware,
  - response status/byte recorder,
  - route/status/site label helpers,
  - multi-site metrics.
- Extended `app.Config` with `SiteName` and `Observability`.
- Extended `app.MultiConfig` with non-YAML/non-JSON observability plumbing.
- Wrapped `Server.Handler()` with HTTP metrics when observability is configured.
- Added multi-site metrics in `NewMultiServer` and `MultiServer.ServeHTTP`.
- Added `--metrics-addr`, `--metrics-path`, and `--pprof` flags to `serve` and `serve-multi`.
- Added unit tests for labels, status class, response recorder, and invalid pprof config.
- Added app integration tests proving HTTP and multi-site metrics increment.

Files changed:

```text
cmd/goja-site/serve.go
cmd/goja-site/serve_multi.go
go.mod
go.sum
pkg/app/config.go
pkg/app/multi_config.go
pkg/app/multi_server.go
pkg/app/server.go
pkg/app/observability_test.go
pkg/observability/config.go
pkg/observability/diagnostics.go
pkg/observability/http.go
pkg/observability/labels.go
pkg/observability/labels_test.go
pkg/observability/multi.go
pkg/observability/registry.go
```

### Why

- Serious benchmarks need internal metrics to explain latency and throughput results.
- Production multi-site hosting needs per-site and unknown-host visibility.
- A private diagnostics listener is safer than exposing `/metrics` or pprof on the public app listener.

### What worked

- The implementation kept existing behavior disabled by default.
- `go test ./...` passed after dependencies and tests were fixed.
- The tests verify that route labels remain coarse and that unknown hosts use `host_class="unknown"` rather than raw host labels.

Final validation output:

```text
go test ./...
?   	github.com/go-go-golems/goja-site/cmd/goja-site	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/app	0.033s
ok  	github.com/go-go-golems/goja-site/pkg/dbguard	(cached)
?   	github.com/go-go-golems/goja-site/pkg/doc	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/kanbanddsl	(cached)
ok  	github.com/go-go-golems/goja-site/pkg/observability	0.003s
```

### What didn't work

The first `go test ./...` failed after `go get github.com/prometheus/client_golang@latest` because transitive Prometheus packages were missing from `go.sum`.

Exact failure pattern:

```text
missing go.sum entry for module providing package github.com/beorn7/perks/quantile (imported by github.com/prometheus/client_golang/prometheus); to add:
	go get github.com/prometheus/client_golang/prometheus@v1.23.2
```

Fix command:

```text
go get github.com/prometheus/client_golang/prometheus@v1.23.2 github.com/prometheus/client_golang/prometheus/promhttp@v1.23.2 && go mod tidy
```

### What I learned

- `github.com/prometheus/client_golang` needs its `prometheus` and `promhttp` subpackages resolved explicitly for this module's `go.sum` when introduced this way.
- A custom registry avoids global Prometheus registration conflicts in tests.
- The existing `Server.Handler()` boundary is sufficient for first-pass HTTP metrics without modifying `gojahttp.Host`.

### What was tricky to build

The tricky part was adding metrics without changing default serving behavior or creating cardinality hazards. The implementation uses a coarse route classifier and labels unknown hosts as `unknown`, not as the raw `Host` header.

Another subtle point is diagnostics startup: pprof is useful, but only safe when it is explicitly mounted on a private diagnostics listener. The code rejects `--pprof` without `--metrics-addr` to enforce that invariant.

### What warrants a second pair of eyes

- Whether `MultiServer` dispatch duration should include downstream site handling or only the map lookup. The current implementation includes downstream handling for successful dispatches.
- Whether `Server.Handler()` should cache the wrapped handler instead of constructing a wrapper each call.
- Whether response recorder should implement optional interfaces such as `http.Flusher` or `http.Hijacker` before WebSocket or streaming use cases are added.

### What should be done in the future

- Add DB and `db.guard` metrics in Phase 3.
- Add Kanban metrics in Phase 4.
- Consider exposing exact route patterns later if `gojahttp.Host` can safely provide them.

### Code review instructions

- Start with `pkg/observability` to review metric names, labels, and diagnostics behavior.
- Then inspect `pkg/app/server.go` and `pkg/app/multi_server.go` for integration points.
- Review CLI flag wiring in `cmd/goja-site/serve.go` and `cmd/goja-site/serve_multi.go`.
- Validate with:

```text
go test ./...
```

### Technical details

New initial metrics include:

```text
goja_site_http_requests_total{site,method,route,status_class}
goja_site_http_request_duration_seconds_bucket{site,method,route}
goja_site_http_response_bytes_bucket{site,method,route}
goja_site_http_in_flight_requests{site}
goja_site_hosts_configured{mode}
goja_site_site_up{site}
goja_site_unknown_host_requests_total{host_class}
goja_site_multi_dispatch_duration_seconds_bucket{result}
```
