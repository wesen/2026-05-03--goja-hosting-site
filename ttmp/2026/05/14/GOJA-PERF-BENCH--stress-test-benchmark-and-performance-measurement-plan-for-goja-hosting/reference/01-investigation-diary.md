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

## Step 9: Implement Phase 2 Vegeta load-generation MVP

I implemented the first load-generation layer around Vegeta. This adds benchmark fixture sites, target examples, documentation, and a wrapper script that can build `goja-site`, start single-site or multi-site scenarios, enable the metrics listener, scrape metrics before and after the load run, and write a result directory.

The script is intentionally shell-based for the first iteration. It gives immediate value while leaving room for a future Go scenario runner that embeds Vegeta as a library once workflows become more stateful.

### Prompt Context

**User prompt (verbatim):** (same as Step 7)

**Assistant interpretation:** Continue through the task phases after Phase 1, starting with the load-generation MVP.

**Inferred user intent:** The user wants practical tools to generate load, preferably using Go-native tooling, and wants progress committed in focused chunks.

**Commit (code):** 48743a178f37978ffa3dde04841241207b3ea3ae — "feat: add vegeta benchmark harness"

### What I did

- Added benchmark fixture scripts:
  - `bench/scripts/null-route/app.js`
  - `bench/scripts/render-route/app.js`
  - `bench/scripts/db-read-write/app.js`
- Added target examples:
  - `bench/targets/null.txt`
  - `bench/targets/render.txt`
  - `bench/targets/db-read.txt`
  - `bench/targets/db-write.txt`
  - `bench/targets/payloads/write-one.json`
- Added `bench/results/.gitignore` so generated benchmark artifacts are not committed.
- Added `bench/README.md` documenting Vegeta and alternatives: Fortio, hey, bombardier, and k6.
- Added `scripts/bench-vegeta.sh` with scenarios:
  - `null`
  - `render`
  - `db-read`
  - `db-write`
  - `multi`
- The wrapper supports:
  - building a temporary `goja-site` binary or using `--binary`,
  - app and metrics ports,
  - output directory selection,
  - temporary DB cleanup,
  - metrics scraping before and after,
  - raw Vegeta binary output,
  - JSON report,
  - text report,
  - Markdown summary,
  - target capture.

### Why

- Phase 1 made metrics available; Phase 2 needed a repeatable way to put traffic through those metrics.
- Vegeta is a good first tool because it is Go-based, supports constant-rate testing, can use Host headers for multi-site tests, and can later be embedded as a library.

### What worked

- `go test ./...` still passed after adding the benchmark fixtures and script.
- `scripts/bench-vegeta.sh --help` printed the expected usage.
- The script is careful to keep generated outputs under `bench/results/`, which is ignored except for `.gitignore`.

Validation output:

```text
go test ./...
?   	github.com/go-go-golems/goja-site/cmd/goja-site	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/app	(cached)
ok  	github.com/go-go-golems/goja-site/pkg/dbguard	(cached)
?   	github.com/go-go-golems/goja-site/pkg/doc	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/kanbanddsl	(cached)
ok  	github.com/go-go-golems/goja-site/pkg/observability	(cached)
```

### What didn't work

A live Vegeta smoke run was skipped because `vegeta` is not currently installed in this environment.

Exact command/output:

```text
if command -v vegeta >/dev/null 2>&1; then scripts/bench-vegeta.sh --scenario null --duration 2s --rate 5/s; else echo 'vegeta not installed; skipping smoke load'; fi
vegeta not installed; skipping smoke load
```

### What I learned

- The load harness can be useful before a custom Go scenario runner exists.
- Multi-site load can be represented with Vegeta target files containing per-target Host headers.
- The script should keep raw target files in each result directory so runs are reproducible.

### What was tricky to build

The tricky part was keeping the wrapper generic without turning it into the full future Go scenario runner. The current script supports the important first scenarios and captures metrics, but avoids complex per-user state, response extraction, or weighted workflows. Those should move to a Go runner later if needed.

### What warrants a second pair of eyes

- Whether the target file format with `@payload.json` is acceptable for all installed Vegeta versions.
- Whether the multi-site scenario should use synthetic benchmark scripts or the real `sites/trail`, `sites/crm`, `sites/editorial`, and `sites/pizza` apps by default.
- Whether the wrapper should auto-install Vegeta or only document installation.

### What should be done in the future

- Install Vegeta and run the short smoke scenario.
- Add pprof capture to the wrapper in Phase 5.
- Add a Go-based scenario runner if shell + target files become too limiting.

### Code review instructions

- Start with `scripts/bench-vegeta.sh`.
- Review cleanup behavior and ensure generated files stay under ignored result directories.
- Review `bench/README.md` for the recommended tool guidance.
- Run after installing Vegeta:

```text
go install github.com/tsenart/vegeta/v12@latest
scripts/bench-vegeta.sh --scenario null --duration 5s --rate 10/s
```

### Technical details

The load harness writes these result files:

```text
summary.md
server.log
targets.txt
vegeta.bin
vegeta.json
vegeta.txt
metrics-before.prom
metrics-after.prom
```

## Step 10: Run Vegeta smoke test and fix readiness race

After Vegeta was installed, I ran the Phase 2 smoke test. The first run exposed a useful harness bug: port `18080` was already serving something, so the readiness curl succeeded against the wrong listener while the newly started `goja-site` process exited with `bind: address already in use`. The harness still ran Vegeta and metrics scraping failed because the diagnostics listener from the failed process was not available.

I fixed the readiness check to verify that the child server process is still alive before and after the curl readiness probe. I then reran the smoke test on alternate ports, and the load run completed with non-empty metrics snapshots.

### Prompt Context

**User prompt (verbatim):** "installed vegeta"

**Assistant interpretation:** Now that the load tool is available, run the smoke test that was previously skipped, fix issues if they appear, and record results.

**Inferred user intent:** The user wants confirmation that the newly added Vegeta harness actually works locally.

**Commit (code):** pending — script fix and diary update will be committed after validation.

### What I did

- Verified Vegeta is installed:

```text
/home/manuel/go/bin/vegeta
Version:
Commit:
Runtime: go1.26.2 linux/amd64
Date:
```

- Ran the default smoke:

```text
scripts/bench-vegeta.sh --scenario null --duration 2s --rate 5/s
```

- Found that app port `18080` was already occupied and the harness readiness check could be fooled by an existing listener.
- Updated `scripts/bench-vegeta.sh` so `wait_ready` checks `kill -0 "$SERVER_PID"` before curling and again after a successful curl.
- Reran the smoke on alternate ports:

```text
scripts/bench-vegeta.sh --scenario null --duration 2s --rate 5/s --port 18180 --metrics-port 19190
```

### Why

- The Phase 2 task list still had a pending short smoke test.
- A load harness must fail fast if the server it started exits, especially when a previous process or unrelated listener is already bound to the requested port.

### What worked

The rerun succeeded:

```text
Requests      [total, rate, throughput]         10, 5.55, 5.55
Duration      [total, attack, wait]             1.801s, 1.801s, 597.318µs
Latencies     [min, mean, 50, 90, 95, 99, max]  367.446µs, 718.41µs, 668.721µs, 1.172ms, 1.491ms, 1.491ms, 1.491ms
Bytes In      [total, mean]                     20, 2.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:10
Error Set:
benchmark complete: /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/20260514T183706Z-null
```

Metrics snapshots were non-empty:

```text
11529 bench/results/20260514T183706Z-null/metrics-before.prom
11550 bench/results/20260514T183706Z-null/metrics-after.prom
23079 total
```

The final metrics included the expected counters/gauges:

```text
goja_site_hosts_configured{mode="single"} 1
goja_site_http_requests_total{method="GET",route="/",site="default",status_class="2xx"} 11
goja_site_site_up{site="default"} 1
```

### What didn't work

The first smoke run failed as a true harness correctness test even though it produced a Vegeta report. The server log showed:

```text
Error: listen tcp 127.0.0.1:18080: bind: address already in use
```

Metrics scraping also failed in that run:

```text
curl: (7) Failed to connect to 127.0.0.1 port 19090 after 0 ms: Couldn't connect to server
```

### What I learned

- Readiness checks must verify the intended child process, not just any successful HTTP response on the target port.
- The load harness should probably get a future `--strict-metrics` option so failed metrics scraping can fail the run rather than only leaving empty files.
- The first successful smoke had 11 HTTP requests in metrics because readiness probing contributes one request before the 10 Vegeta requests.

### What was tricky to build

The subtle bug was caused by a port collision. A simple curl readiness probe cannot distinguish between the server started by the harness and an unrelated process already listening on that port. Checking the child PID before and after the curl catches the common bind-failure case.

### What warrants a second pair of eyes

- Whether the harness should proactively check that app and metrics ports are free before starting.
- Whether failed metrics scraping should be fatal by default now that metrics are part of the benchmark contract.

### What should be done in the future

- Add optional strict mode for metrics scrape failures.
- Consider defaulting to less common ports or choosing random free ports.
- Continue with Phase 3 database and `db.guard` metrics.

### Code review instructions

- Review the `wait_ready` change in `scripts/bench-vegeta.sh`.
- Reproduce with:

```text
scripts/bench-vegeta.sh --scenario null --duration 2s --rate 5/s --port 18180 --metrics-port 19190
```

### Technical details

Successful smoke result directory:

```text
bench/results/20260514T183706Z-null
```

This directory is ignored by git through `bench/results/.gitignore`.

## Step 11: Implement Phase 3 database and db.guard metrics

I implemented Phase 3 by adding database operation metrics and guarded database metrics. JavaScript `database`/`db` calls are now wrapped at the Go `QueryExecer` boundary, so both module names and both database policies can be measured without changing site scripts.

The `db.guard` integration uses an observer interface in `pkg/dbguard` so guard internals do not import Prometheus directly. The Prometheus implementation lives in `pkg/observability` and is attached from `pkg/app/database.go` when the guarded policy is used and observability is configured.

### Prompt Context

**User prompt (verbatim):** "go ahead, continue keeping a detailed diary."

**Assistant interpretation:** Continue implementing the phased backlog, starting with Phase 3, and keep detailed diary notes.

**Inferred user intent:** The user wants steady progress beyond the Vegeta smoke test, with each phase validated and documented.

**Commit (code):** 2e55df47cb1a7ca811e658ed0eeb226ecee23a82 — "feat: add database guard metrics"

### What I did

- Added `pkg/observability/sql.go`:
  - DB operation counter.
  - DB operation duration histogram.
  - DB error counter.
  - bounded SQL kind classifier.
  - bounded error class classifier.
  - `InstrumentQueryExecer` wrapper.
- Added `pkg/observability/guard.go`:
  - guard check counters and duration histograms.
  - guard limit exceeded counter.
  - cleanup attempt counter.
  - DB size gauges.
  - DB limit gauges.
  - writes-since-check gauge.
- Extended `pkg/observability/registry.go` with DB and guard metric groups.
- Added `dbguard.Observer` and `Guard.SetObserver`.
- Instrumented `dbguard.Guard` paths:
  - `Stats`.
  - `BeforeExec`.
  - `ErrorAfterExec`.
  - `AfterExec` skipped paths.
  - `CheckNow` under-limit/over-limit/cleanup paths.
  - hard-limit errors.
  - measured DB stats.
- Wrapped the selected database policy in `pkg/app/database.go` with `observability.InstrumentQueryExecer`.
- Attached `observability.NewGuardObserver` when guarded policy is used.
- Added tests for SQL kind labels and DB/guard metrics.
- Ran `go test ./...`.
- Ran a short `db-read` Vegeta smoke on alternate ports and confirmed DB metrics appear in the metrics snapshot.

### Why

- HTTP metrics show that a route is slow, but DB metrics explain whether SQLite reads/writes are the source.
- Guarded policy is the default and can add work around writes, so guard checks must be observable before serious stress testing.
- The metrics use SQL kind labels instead of raw SQL to avoid leaking data and exploding cardinality.

### What worked

`go test ./...` passed:

```text
?   	github.com/go-go-golems/goja-site/cmd/goja-site	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/app	0.065s
ok  	github.com/go-go-golems/goja-site/pkg/dbguard	0.038s
?   	github.com/go-go-golems/goja-site/pkg/doc	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/kanbanddsl	(cached)
ok  	github.com/go-go-golems/goja-site/pkg/observability	0.007s
```

The `db-read` Vegeta smoke succeeded:

```text
scripts/bench-vegeta.sh --scenario db-read --duration 2s --rate 5/s --port 18181 --metrics-port 19191
Requests      [total, rate, throughput]         10, 5.55, 5.55
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:10
benchmark complete: /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/20260514T184834Z-db-read
```

The metrics snapshot showed DB operation metrics:

```text
goja_site_db_operations_total{db_policy="simple",operation="exec",site="default",sql_kind="create"} 2
goja_site_db_operations_total{db_policy="simple",operation="exec",site="default",sql_kind="insert"} 100
goja_site_db_operations_total{db_policy="simple",operation="query",site="default",sql_kind="select"} 101
```

### What didn't work

- N/A for this step. The implementation compiled and tests passed after the first full test run.

### What I learned

- The cleanest DB instrumentation point is after policy selection and before constructing the JavaScript `database` and `db` modules.
- Guard metrics need to observe both fast skipped paths and actual measurement/cleanup paths, otherwise write-heavy workloads with no configured limits would appear invisible.
- The benchmark fixture inserts seed rows during script load, so DB metrics include startup script SQL as well as request-time SQL. That is useful but should be called out in benchmark reports.

### What was tricky to build

The tricky part was keeping `dbguard` decoupled from Prometheus while still exposing enough internal events. The observer interface solves this, but it means future guard changes must remember to call observer methods for new paths.

Another subtle point is result labeling. The guard has many skipped paths (`no limit configured`, throttle, cooldown, cleanup already running). These were normalized into bounded result labels such as `skipped_no_limit` and `skipped_throttle` rather than exporting arbitrary text.

### What warrants a second pair of eyes

- Whether `goja_site_db_operation_duration_seconds` should remain one metric with an `operation` label or split into separate `query_duration` and `exec_duration` metrics.
- Whether guard observer calls while holding the guard mutex are acceptable; Prometheus calls should be quick, but this is still worth reviewing.
- Whether startup script SQL should be labeled separately from request-time SQL in a future phase.

### What should be done in the future

- Add Kanban fragment/action/render metrics in Phase 4.
- Consider a request context or operation label that distinguishes script load from HTTP request execution.
- Add a guarded-policy scenario to the Vegeta harness so guard metrics can be exercised externally without relying only on unit/integration tests.

### Code review instructions

- Start with `pkg/observability/sql.go` and `pkg/observability/guard.go`.
- Then review `pkg/dbguard/guard.go` observer hooks.
- Then review `pkg/app/database.go` to verify wrapping order.
- Validate with:

```text
go test ./...
scripts/bench-vegeta.sh --scenario db-read --duration 2s --rate 5/s --port 18181 --metrics-port 19191
```

### Technical details

New DB metrics:

```text
goja_site_db_operations_total{site,db_policy,operation,sql_kind}
goja_site_db_operation_duration_seconds_bucket{site,db_policy,operation,sql_kind}
goja_site_db_errors_total{site,db_policy,operation,sql_kind,error_class}
```

New guard metrics:

```text
goja_site_db_guard_checks_total{site,phase,result}
goja_site_db_guard_check_duration_seconds_bucket{site,phase,result}
goja_site_db_guard_limit_exceeded_total{site,kind,hard}
goja_site_db_guard_cleanup_attempts_total{site,result}
goja_site_db_size_bytes{site,component}
goja_site_db_limit_bytes{site,limit_type}
goja_site_db_guard_writes_since_check{site}
```

## Step 12: Implement Phase 4 Kanban metrics

I implemented Phase 4 by adding Kanban-specific metrics around the mounted board routes. This gives visibility into the realistic interaction path that combines JavaScript callbacks, board data loading, UI DSL rendering, and optional refreshed HTML in action responses.

The implementation adds an observer interface to `pkg/kanbanddsl`, then provides a Prometheus-backed observer in `pkg/observability`. The app server wires the observer into the Kanban DSL registrar when observability is enabled for a site.

### Prompt Context

**User prompt (verbatim):** "continue, keep a diary, commit at appropriate intervals"

**Assistant interpretation:** Continue with the next implementation phase, keep diary evidence, validate, and commit at a coherent checkpoint.

**Inferred user intent:** The user wants ongoing progress through the phased backlog with reviewable commits and clear notes.

**Commit (code):** a08e5848944e1c4e9eed7a78f712f865662ed679 — "feat: add kanban metrics"

### What I did

- Added `pkg/kanbanddsl/observer.go` with a Kanban observer interface.
- Extended `kanbanddsl.Runtime` to carry an optional observer.
- Changed the Kanban registrar to accept an optional observer while keeping `kanbanddsl.NewRegistrar()` backwards compatible.
- Added `pkg/observability/kanban.go` with Prometheus metrics for:
  - fragment route duration,
  - action route duration,
  - action dispatch duration,
  - render duration,
  - rendered HTML bytes,
  - bounded error counters.
- Extended the observability registry with Kanban metrics.
- Wired `observability.NewKanbanObserver` into `app.NewServer`.
- Instrumented `Board.Mount` fragment and action handlers.
- Added app integration coverage that exercises a mounted board fragment and `cardMoved` action, then verifies metric histogram counts.
- Ran `go test ./...`.

### Why

- Kanban action routes are likely to be among the most representative goja-site workloads.
- A single Kanban action can include JavaScript dispatch, application state mutation, board reload, UI node creation, Go rendering, and JSON response writing.
- Generic HTTP latency cannot distinguish dispatch time from refresh render time.

### What worked

`go test ./...` passed:

```text
?   	github.com/go-go-golems/goja-site/cmd/goja-site	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/app	0.063s
ok  	github.com/go-go-golems/goja-site/pkg/dbguard	(cached)
?   	github.com/go-go-golems/goja-site/pkg/doc	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/kanbanddsl	(cached)
ok  	github.com/go-go-golems/goja-site/pkg/observability	(cached)
```

The new integration test verifies:

```text
goja_site_kanban_fragment_duration_seconds
goja_site_kanban_action_duration_seconds{refresh="true"}
goja_site_kanban_dispatch_duration_seconds
goja_site_kanban_render_duration_seconds{reason="action_refresh"}
```

### What didn't work

- N/A for this step. The instrumentation and tests passed after formatting.

### What I learned

- The Kanban registrar was a clean place to inject site-specific observers because each `Server` builds its own Goja runtime and module registrar set.
- Fragment routes can observe render duration before response rendering, but they do not currently know final HTML byte length because the response helper handles HTML rendering internally.
- Action refresh routes do explicitly call `uidsl.RenderAny`, so they can record rendered HTML bytes.

### What was tricky to build

The tricky part was preserving backwards compatibility. Existing tests and code call `kanbanddsl.NewRegistrar()` without arguments, so the registrar now accepts optional observers with a variadic parameter.

The other subtle point was avoiding duplicate or misleading action observations on error paths. The action handler now records dispatch duration, action duration, and render duration where possible, including error classes through bounded labels.

### What warrants a second pair of eyes

- Whether fragment render should also record rendered HTML bytes by changing the response path to explicitly call `uidsl.RenderAny`.
- Whether action names should be further normalized if custom trusted actions become numerous.
- Whether the observer should recover from panics defensively, or whether Prometheus observer implementations are safe enough.

### What should be done in the future

- Add a Kanban scenario to the Vegeta harness or a future Go scenario runner so these metrics can be exercised outside tests.
- Consider measuring full-page `board.render(...)` calls, not just mounted fragment/action routes.
- Continue to Phase 5 pprof capture automation.

### Code review instructions

- Start with `pkg/kanbanddsl/observer.go`, `pkg/kanbanddsl/registrar.go`, and `pkg/kanbanddsl/mount.go`.
- Then review `pkg/observability/kanban.go` for metric names and labels.
- Finally review `pkg/app/observability_test.go` for integration coverage.
- Validate with:

```text
go test ./...
```

### Technical details

New Kanban metrics:

```text
goja_site_kanban_fragment_duration_seconds{site,board}
goja_site_kanban_action_duration_seconds{site,board,action,refresh}
goja_site_kanban_dispatch_duration_seconds{site,board,action}
goja_site_kanban_render_duration_seconds{site,board,reason}
goja_site_kanban_rendered_html_bytes{site,board,reason}
goja_site_kanban_errors_total{site,board,action,phase,error_class}
```

## Step 13: Implement Phase 5 pprof capture automation

I implemented Phase 5 by extending the Vegeta harness with optional pprof capture. The harness can now start `goja-site` with pprof enabled on the private diagnostics listener, capture a CPU profile during the load run, and capture heap, goroutine, and allocs profiles after the run.

This phase makes performance debugging more reproducible: a benchmark result directory can now contain the external load report, Prometheus snapshots, server logs, and Go runtime profiles for offline analysis with `go tool pprof`.

### Prompt Context

**User prompt (verbatim):** "go ahead"

**Assistant interpretation:** Continue to the next planned phase, which is pprof capture automation in the load harness.

**Inferred user intent:** The user wants the benchmark/observability work to keep moving without additional prompting, while preserving diary and commits.

**Commit (code):** pending at diary draft time — this step will be updated after the Phase 5 commit.

### What I did

- Added `--pprof` to `scripts/bench-vegeta.sh`.
- Added `--pprof-seconds N` to control CPU profile duration.
- When `--pprof` is set, the harness now:
  - passes `--pprof` to `goja-site serve` or `serve-multi`,
  - starts CPU profile capture from `/debug/pprof/profile?seconds=N` while Vegeta runs,
  - waits for CPU profile capture to finish,
  - captures `/debug/pprof/heap` after the run,
  - captures `/debug/pprof/goroutine?debug=1` after the run,
  - captures `/debug/pprof/allocs` after the run,
  - includes pprof metadata and artifact names in `summary.md`.
- Kept pprof artifacts under `bench/results/...`, which is ignored by git.
- Ran shell syntax validation.
- Ran `go test ./...`.
- Ran a short pprof-enabled smoke test.

### Why

- Metrics identify which subsystem is slow, but pprof helps explain why CPU, allocation, heap, or goroutine behavior looks wrong.
- Capturing profiles inside the benchmark result directory keeps evidence tied to a specific run.
- pprof should remain opt-in because it adds overhead and can expose internal details.

### What worked

Validation passed:

```text
bash -n scripts/bench-vegeta.sh
go test ./...
```

The pprof smoke run succeeded:

```text
scripts/bench-vegeta.sh --scenario null --duration 2s --rate 5/s --port 18182 --metrics-port 19192 --pprof --pprof-seconds 1
Requests      [total, rate, throughput]         10, 5.55, 5.55
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:10
benchmark complete: /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/20260514T195501Z-null
```

The result directory contains pprof artifacts:

```text
allocs.pprof
cpu.pprof
goroutine.txt
heap.pprof
metrics-after.prom
metrics-before.prom
server.log
summary.md
targets.txt
vegeta.bin
vegeta.json
vegeta.txt
```

The profile files are valid gzip-compressed pprof data:

```text
allocs.pprof: gzip compressed data
cpu.pprof: gzip compressed data
heap.pprof: gzip compressed data
goroutine.txt: ASCII text
```

### What didn't work

- N/A for this step after fixing the shell argument handling before the smoke run.

### What I learned

- Bash `${VAR:+--pprof}` is not safe when the variable contains `0`, because non-empty strings still expand. The harness now builds an explicit `PPROF_ARGS` array only when `CAPTURE_PPROF == 1`.
- Short CPU profiles on tiny smoke workloads are small but still valid. Real analysis should use longer runs.

### What was tricky to build

The tricky detail was coordinating CPU profile capture with the load run. The CPU profile endpoint blocks for its requested duration, so the harness starts that curl in the background before Vegeta and waits for it after Vegeta completes.

The second tricky detail was preserving the safety invariant from Phase 1: pprof is still only exposed through the diagnostics listener, because the harness only passes `--pprof` together with `--metrics-addr`.

### What warrants a second pair of eyes

- Whether failed profile capture should fail the whole benchmark run instead of being best-effort.
- Whether CPU profile duration should default to the Vegeta duration instead of a separate `--pprof-seconds` default.
- Whether to add block/mutex profiles after enabling their runtime sampling rates.

### What should be done in the future

- Add `go tool pprof` convenience commands to generated summaries.
- Consider strict mode for required metrics/profile artifacts.
- Continue to Phase 6 OpenTelemetry tracing.

### Code review instructions

- Review `scripts/bench-vegeta.sh`, especially `PPROF_ARGS`, profile capture timing, and cleanup.
- Reproduce with:

```text
scripts/bench-vegeta.sh --scenario null --duration 2s --rate 5/s --port 18182 --metrics-port 19192 --pprof --pprof-seconds 1
```

- Inspect:

```text
file bench/results/<run>/cpu.pprof bench/results/<run>/heap.pprof
```

### Technical details

New pprof artifacts:

```text
cpu.pprof       # /debug/pprof/profile?seconds=N
heap.pprof      # /debug/pprof/heap
allocs.pprof    # /debug/pprof/allocs
goroutine.txt   # /debug/pprof/goroutine?debug=1
```

## Step 14: Start Phase 6 OpenTelemetry tracing

I implemented the first Phase 6 tracing slice: OpenTelemetry dependencies, CLI flags, tracer provider setup, OTLP HTTP export, HTTP handler tracing, database operation spans, benchmark harness flags, and an example collector config for forwarding traces to a Jaeger-compatible backend.

This is intentionally a first tracing slice rather than the final tracing story. HTTP and DB spans are now in place. Multi-site dispatch, guard checks, and Kanban spans remain as explicit follow-up tasks because they need careful context propagation and low-cardinality attribute design.

### Prompt Context

**User prompt (verbatim):** "continue"

**Assistant interpretation:** Continue from Phase 5 into Phase 6 OpenTelemetry tracing and commit a coherent first tracing slice.

**Inferred user intent:** The user wants ongoing implementation progress with diary entries and commits, not another planning-only response.

**Commit (code):** pending at diary draft time — this step will be updated after the tracing slice commit.

### What I did

- Added OpenTelemetry dependencies:
  - `go.opentelemetry.io/otel`
  - `go.opentelemetry.io/otel/sdk`
  - `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`
  - `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`
- Added `pkg/observability/tracing.go`:
  - `TracingConfig`.
  - `InitTracing`.
  - OTLP HTTP exporter setup.
  - parent-based trace-id-ratio sampler.
  - service name resource attribute.
  - W3C trace context/baggage propagation.
  - HTTP handler wrapping through `otelhttp`.
- Extended `Observability` with a tracer and tracing-enabled flag.
- Wrapped `Server.Handler()` with OTel HTTP instrumentation when tracing is enabled.
- Added DB query/exec spans in `pkg/observability/sql.go` with bounded attributes:
  - `goja_site.site`
  - `goja_site.db_policy`
  - `goja_site.sql_kind`
- Added tracing flags to `serve` and `serve-multi`:
  - `--otel-enabled`
  - `--otel-endpoint`
  - `--otel-sample-ratio`
  - `--service-name`
- Added corresponding Vegeta harness flags:
  - `--otel`
  - `--otel-endpoint`
  - `--otel-sample-ratio`
- Added `deploy/observability/otel-collector-jaeger.example.yaml`.
- Added a disabled-tracing no-op test.
- Ran tests and a short OTel-enabled Vegeta smoke.

### Why

- HTTP spans provide top-level request timing and route classification.
- DB spans add useful lower-level timing for database operations.
- OTLP keeps the app backend-neutral; Jaeger can be used through an OpenTelemetry Collector.
- Tracing must be opt-in and sampled because it has more overhead than metrics.

### What worked

`go test ./...` passed:

```text
?   	github.com/go-go-golems/goja-site/cmd/goja-site	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/app	(cached)
ok  	github.com/go-go-golems/goja-site/pkg/dbguard	(cached)
?   	github.com/go-go-golems/goja-site/pkg/doc	[no test files]
ok  	github.com/go-go-golems/goja-site/pkg/kanbanddsl	(cached)
ok  	github.com/go-go-golems/goja-site/pkg/observability	(cached)
```

The harness help now exposes OTel flags:

```text
--otel                enable OpenTelemetry tracing for the goja-site process
--otel-endpoint URL   OTLP HTTP traces endpoint (default: http://127.0.0.1:4318/v1/traces)
--otel-sample-ratio N trace sample ratio between 0 and 1 (default: 0.01)
```

The OTel-enabled smoke succeeded at the HTTP level:

```text
scripts/bench-vegeta.sh --scenario null --duration 1s --rate 2/s --port 18183 --metrics-port 19193 --otel --otel-sample-ratio 1
Requests      [total, rate, throughput]         2, 4.00, 3.99
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:2
benchmark complete: /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/20260514T202056Z-null
```

### What didn't work

The OTel smoke did not have a collector listening locally, so shutdown/export logged the expected connection failure:

```text
2026/05/14 16:21:01 traces export: Post "http://127.0.0.1:4318/v1/traces": dial tcp 127.0.0.1:4318: connect: connection refused
```

This is acceptable for the smoke because it validates that tracing can be enabled without breaking request handling. A real trace delivery test needs an OpenTelemetry Collector or test exporter.

The first `go test ./...` after adding tracing code failed because the OTel dependencies were not retained in `go.mod`/`go.sum` until the imports existed. Running explicit `go get ... && go mod tidy` after adding imports fixed it.

### What I learned

- Adding OTel before imports are present can be undone by `go mod tidy`; add code imports first or run `go get` again after code is present.
- `otelhttp` provides a clean top-level request span without modifying `gojahttp.Host`.
- DB spans currently start from `context.Background()` because the JavaScript database module interface does not pass request context into `QueryExecer`. This means DB spans are useful timing events but are not yet children of request spans.

### What was tricky to build

The main tricky issue is context propagation across the Goja boundary. HTTP requests have context, but JavaScript route handlers call database methods through a `QueryExecer` interface that does not carry context. A future improvement could store request context in the runtime owner call path or add context-aware database module support.

The second tricky issue is exporter behavior when no collector is present. The app should continue serving requests, but shutdown/export can log connection failures. The smoke confirmed request handling remains healthy.

### What warrants a second pair of eyes

- Whether DB spans without request parentage are useful enough, or whether they should wait for context propagation.
- Whether default `--otel-endpoint` should be empty and rely on `OTEL_EXPORTER_OTLP_ENDPOINT` instead of defaulting to localhost.
- Whether CLI flags should mirror standard `OTEL_*` environment variables more completely.

### What should be done in the future

- Add an integration test with an in-memory or local OTel collector/test exporter.
- Add spans for multi-site dispatch, guard checks, Kanban dispatch, and Kanban render once context propagation is clearer.
- Consider adding context-aware DB instrumentation in go-go-goja's database module.

### Code review instructions

- Start with `pkg/observability/tracing.go`.
- Review CLI flag wiring in `cmd/goja-site/serve.go` and `cmd/goja-site/serve_multi.go`.
- Review DB span attributes in `pkg/observability/sql.go` and confirm raw SQL is not attached.
- Validate with:

```text
go test ./...
scripts/bench-vegeta.sh --scenario null --duration 1s --rate 2/s --port 18183 --metrics-port 19193 --otel --otel-sample-ratio 1
```

### Technical details

New tracing flags:

```text
--otel-enabled
--otel-endpoint http://127.0.0.1:4318/v1/traces
--otel-sample-ratio 0.01
--service-name goja-site
```

New harness flags:

```text
--otel
--otel-endpoint URL
--otel-sample-ratio N
```

## Step 15: Analyze go-go-goja context propagation for JS native modules

I investigated how `go-go-goja` currently handles `context.Context` across runtime owner calls, HTTP route execution, runtime bridge bindings, and database module calls. I then wrote a dedicated design and implementation guide for doing context propagation properly in `go-go-goja` itself rather than patching around it only in `goja-site`.

The key finding is positive: `gojahttp.Host` already passes `r.Context()` into `runtimeowner.Runner.Call` when invoking JavaScript HTTP handlers. The missing piece is a standard way for native modules invoked from JavaScript to retrieve the current owner-call/request context.

### Prompt Context

**User prompt (verbatim):** "how does go-go-goja handles context in its calls. I want to do the context through JS right, even if it means updating go-go-goja itself. 

Create a separate analysis and implementation guide for that topic and upload to the remarkable tablet when done."

**Assistant interpretation:** Inspect go-go-goja internals, explain how context currently flows, design the right upstream changes for JS/native-module context propagation, store the guide in the ticket, and upload the refreshed bundle to reMarkable.

**Inferred user intent:** The user wants correct trace parentage and cancellation semantics across HTTP → Goja JavaScript → native modules → database, even if that requires changing the shared go-go-goja library.

**Commit (code):** N/A — analysis/design-document step only.

### What I did

- Inspected go-go-goja v0.4.16 in the module cache.
- Read key source files:
  - `pkg/runtimeowner/runner.go`
  - `pkg/runtimeowner/types.go`
  - `pkg/gojahttp/host.go`
  - `pkg/gojahttp/request_response.go`
  - `pkg/runtimebridge/runtimebridge.go`
  - `engine/factory.go`
  - `modules/database/database.go`
- Captured line-numbered evidence with `nl -ba` for the design guide.
- Created a new ticket design doc:
  - `design-doc/03-goja-request-context-propagation-across-javascript-and-native-modules.md`
- Updated the ticket task list with Phase 6A for upstream go-go-goja context propagation.

### Why

- The current `goja-site` DB spans use `context.Background()` because the JS database module does not receive request context.
- Adding more spans without solving context propagation would create disconnected traces.
- The right abstraction belongs in go-go-goja because all native modules can benefit, not just `goja-site`'s database wrapper.

### What worked

- `runtimeowner.Runner` already has context-aware `Call` and `Post` APIs.
- `gojahttp.Host` already invokes route handlers with `h.owner.Call(r.Context(), "http-handler", ...)`.
- `runtimebridge` already stores per-VM runtime lifecycle bindings, making it a natural package to extend with current-call context.
- The database module has a clear compatibility path: add `QueryExecerContext`, keep `QueryExecer`, and export VM-closing JS function closures from `Loader`.

### What didn't work

- Native modules currently cannot retrieve the current request context from `*goja.Runtime`.
- `modules/database.QueryExecer` has only `Query` and `Exec`, not `QueryContext` and `ExecContext`.
- `DBModule.Query` and `Exec` currently do not receive `*goja.Runtime`, so the loader must export closures that close over `vm` if they need `runtimebridge.CurrentContext(vm)`.

### What I learned

- The context is not lost at the HTTP boundary; it is lost at the native module boundary.
- Owner context in `runtimeowner` currently solves reentrant owner scheduling, not general current request-context lookup.
- The clean design is to add a current-call context stack to `runtimebridge` and wrap `runtimeowner.invoke` / `invokePost`.

### What was tricky to build

The tricky distinction is between runtime lifecycle context and request/call context. `runtimebridge.Bindings.Context` is runtime-scoped and lasts until runtime shutdown. It is not the same as the per-request context that carries cancellation, deadlines, and OpenTelemetry parent spans.

Another subtlety is JavaScript ergonomics. Passing `req.context` explicitly through JS would work mechanically but is the wrong API. Site authors should not be responsible for threading Go context through every native module call.

### What warrants a second pair of eyes

- Whether `runtimebridge` is the best home for current-call context, or whether a new package should own it.
- Whether the current-context stack should be mutex-protected even though Goja owner calls are serialized.
- Whether async timer callbacks should inherit request context or runtime lifecycle context by default.

### What should be done in the future

- Implement the guide in go-go-goja:
  - runtimebridge current-call context stack,
  - runtimeowner wrapping,
  - database `QueryContext`/`ExecContext` support.
- Then update goja-site wrappers to use context-aware DB calls and add trace parentage tests.

### Code review instructions

- Start with the new guide:
  - `ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/design-doc/03-goja-request-context-propagation-across-javascript-and-native-modules.md`
- Compare the claims against go-go-goja v0.4.16 source files named in the References section.
- Pay special attention to `runtimeowner.Runner.Call`, `gojahttp.Host.ServeHTTP`, `runtimebridge.Bindings`, and `modules/database.DBModule.Loader`.

### Technical details

Recommended go-go-goja API sketch from the guide:

```text
runtimebridge.CurrentContext(vm *goja.Runtime) context.Context
runtimebridge.WithCallContext(vm *goja.Runtime, ctx context.Context, fn func() (any, error)) (any, error)
runtimebridge.WithCallContextVoid(vm *goja.Runtime, ctx context.Context, fn func() error) error
```

Recommended database extension:

```text
QueryExecerContext.QueryContext(ctx, query, args...)
QueryExecerContext.ExecContext(ctx, query, args...)
DBModule.QueryContext(ctx, query, args...)
DBModule.ExecContext(ctx, query, args...)
```

## Step 16: Upload context propagation guide bundle to reMarkable

I uploaded the refreshed GOJA-PERF-BENCH bundle to reMarkable after adding the third design document. The bundle now includes the original benchmark guide, the production observability guide, the new go-go-goja request context propagation guide, tasks, changelog, index, and diary.

The upload used the same stable PDF name and destination as previous bundle uploads, with `--force`, so the reMarkable copy is current.

### Prompt Context

**User prompt (verbatim):** (same as Step 15)

**Assistant interpretation:** Deliver the new context propagation guide to the reMarkable tablet after storing it in the ticket.

**Inferred user intent:** The user wants the analysis available for reading/review on device, not only in the repo.

**Commit (code):** 07c38f2b8ae3d87bcdf18ba96568d715b2b3159f — "docs: design goja context propagation"

### What I did

- Ran a dry-run upload including the new third design doc.
- Ran the real upload with `--force`.

Dry-run evidence:

```text
DRY: include .../design-doc/03-goja-request-context-propagation-across-javascript-and-native-modules.md (title="03-goja-request-context-propagation-across-javascript-and-native-modules")
DRY: upload GOJA_PERF_BENCH_Guide.pdf -> /ai/2026/05/14/GOJA-PERF-BENCH
```

Real upload evidence:

```text
OK: uploaded GOJA_PERF_BENCH_Guide.pdf -> /ai/2026/05/14/GOJA-PERF-BENCH
```

### Why

- The user explicitly requested upload to the reMarkable tablet.
- Keeping the same bundle name makes the ticket packet easy to locate.

### What worked

- The dry run confirmed the new third design doc was included.
- The forced upload succeeded.

### What didn't work

- N/A for this step.

### What I learned

- The stable bundle workflow remains effective for iterative ticket packets, but it can overwrite annotations on the prior PDF.

### What was tricky to build

The only operational caveat is the intentional use of `--force`; this keeps one canonical bundle on reMarkable but may replace annotated copies.

### What warrants a second pair of eyes

- Confirm whether future uploads should use a new bundle name instead of overwriting the canonical `GOJA_PERF_BENCH_Guide.pdf`.

### What should be done in the future

- Implement the go-go-goja context propagation guide in the go-go-goja repo or with a local `replace` workflow.
- Re-upload after implementation notes/results are added.

### Code review instructions

- Confirm the reMarkable destination:

```text
/ai/2026/05/14/GOJA-PERF-BENCH/GOJA_PERF_BENCH_Guide.pdf
```

### Technical details

The uploaded bundle now includes:

```text
index.md
tasks.md
changelog.md
01-goja-hosting-stress-test-benchmark-and-performance-guide.md
02-goja-site-production-observability-metrics-and-tracing-guide.md
03-goja-request-context-propagation-across-javascript-and-native-modules.md
01-investigation-diary.md
```

## Step 17: Implement go-go-goja current request context propagation and wire goja-site DB spans

I implemented the Phase 6A context propagation design in the local `go-go-goja` checkout at `../go-go-golems/go-go-goja`, then wired `goja-site` to use that local checkout through a temporary `replace` directive. The implementation now lets JavaScript native modules retrieve the active owner-call context from `runtimebridge.CurrentContext(vm)`, and the database module uses that context for `QueryContext`/`ExecContext` when available.

### Prompt Context

**User prompt (verbatim):** "The go-go-goja code is in ../go-go-golems/go-go-goja/ btw . Let's implement this, you can use replace in go.mod for now to use the local development go-go-goja.

Keep a detailed diary as you go, commit at appropriate intervals."

**Assistant interpretation:** Implement the context propagation design across the upstream local `go-go-goja` repo and this `goja-site` repo, validate with tests, document each step, and commit focused changes.

**Inferred user intent:** The user wants the previously designed context propagation mechanism moved from design to working code so `goja-site` DB spans become children of HTTP request spans.

**Commit (code):** `4b7aa26c86eae3692551bd05f32713fa5dd12880` in `../go-go-golems/go-go-goja` — "feat: propagate owner call context to native modules".

### What I did

In `../go-go-golems/go-go-goja`:

- Added `runtimebridge.CurrentContext(vm)`.
- Added `runtimebridge.WithCallContext(...)` and `WithCallContextVoid(...)`.
- Added a per-VM current-call context stack with push/pop/peek behavior.
- Updated `runtimebridge.Delete(vm)` to also delete current-call context state.
- Broke the prior `runtimebridge -> runtimeowner` import dependency by replacing `Bindings.Owner runtimeowner.Runner` with a local `runtimebridge.OwnerRunner` interface that exposes only `Post(...)`.
- Added an `engine.runtimebridgeOwner` adapter so the engine can still store a runtime owner in `runtimebridge.Bindings` without an import cycle.
- Wrapped `runtimeowner.runner.invoke` and `invokePost` with `runtimebridge.WithCallContext` / `WithCallContextVoid`.
- Added `modules/database.QueryExecerContext`.
- Added `DBModule.QueryContext` and `DBModule.ExecContext`.
- Changed the database module loader to export closures that close over `vm`, call `runtimebridge.CurrentContext(vm)`, and then call the context-aware DB methods.
- Kept `DBModule.Query` and `Exec` as compatibility wrappers around `context.Background()`.
- Added tests for runtimebridge fallback/nesting/panic cleanup.
- Added a runtimeowner test proving `Runner.Call(ctx, ...)` makes the same context visible through `runtimebridge.CurrentContext(vm)`.
- Added database module tests proving JS `db.exec(...)` receives the owner-call context and legacy `QueryExecer` fallback still works.

In `goja-site`:

- Added a temporary local module replacement:

```text
replace github.com/go-go-golems/go-go-goja => ../go-go-golems/go-go-goja
```

- Added `QueryContext`/`ExecContext` to `simpleDB` using `*sql.DB.QueryContext` and `ExecContext`.
- Added `QueryContext`/`ExecContext` to `dbguard.MeteredDB`, preserving guard checks while using context-aware SQL execution.
- Added `QueryContext`/`ExecContext` to `observability.InstrumentedQueryExecer`.
- Changed DB span creation to start from the incoming context rather than `context.Background()`.
- Added context-aware fallback helpers for legacy `databasemod.QueryExecer` implementations.
- Added `TestServerDBSpansAreChildrenOfHTTPSpan`, using an in-memory OpenTelemetry exporter to prove a JS route's `db.query(...)` span has the HTTP request span as parent.

### Why

The previous Phase 6 tracing slice emitted DB spans, but those spans were disconnected because the DB instrumentation had to use `context.Background()`. This step fixes the actual propagation gap:

```text
HTTP request context -> gojahttp.Host -> runtimeowner.Call -> JavaScript handler -> database module -> goja-site DB instrumentation -> sql.DB QueryContext/ExecContext
```

### What worked

- The design direction was correct: `gojahttp.Host` was already passing `r.Context()` into `runtimeowner.Call`, so the missing implementation was at the owner/native-module boundary.
- A per-VM stack in `runtimebridge` gives native modules a simple API: `runtimebridge.CurrentContext(vm)`.
- Exporting database JS functions as loader closures solved the problem that `DBModule.Query`/`Exec` did not otherwise receive `*goja.Runtime`.
- The local `replace` directive let `goja-site` consume the development version immediately.
- The in-memory trace exporter test now proves parent-child trace structure instead of relying only on visual Jaeger inspection.

### What didn't work

My first implementation attempt made `runtimeowner` import `runtimebridge` while `runtimebridge` still imported `runtimeowner` for the `Bindings.Owner` field. That created an import-cycle risk. I fixed it by decoupling `runtimebridge` from `runtimeowner`:

```text
runtimebridge.OwnerRunner interface { Post(...) error }
engine.runtimebridgeOwner adapter wraps runtimeowner.Runner
```

The first trace parentage test also failed because `otelhttp.NewHandler` uses the global OpenTelemetry tracer provider, while the DB instrumentation used the explicit `obs.Tracer`. The failure showed only root DB spans and no HTTP span. I fixed the test by temporarily setting the global provider with `otel.SetTracerProvider(provider)` and restoring the previous provider in cleanup.

### Validation

In `../go-go-golems/go-go-goja`:

```text
go test ./pkg/runtimebridge ./pkg/runtimeowner ./modules/database
```

Result: passed.

```text
go test ./...
```

Result: passed.

The `go-go-goja` pre-commit hook also ran:

```text
golangci-lint run -v
go generate ./...
go test ./...
```

Result: passed, then commit `4b7aa26c86eae3692551bd05f32713fa5dd12880` was created.

In `goja-site`:

```text
go test ./pkg/app -run 'TestServerDBSpansAreChildrenOfHTTPSpan' -count=1
```

Result: passed after setting the global test trace provider.

```text
go test ./...
```

Result: passed.

### What I learned

- The trace gap was exactly where expected: request context was present at `Runner.Call`, but native modules could not access it before this change.
- `otelhttp` and explicit DB tracers can accidentally use different providers in tests; for parentage tests, both sides must use the same provider.
- Avoiding an import cycle required making `runtimebridge` own only the minimal owner scheduling interface it needs.

### What was tricky to build

The subtle part was keeping `runtimebridge` as the package native modules import while also making `runtimeowner` call into it. That required removing the direct `runtimebridge` dependency on `runtimeowner` and adding a small engine adapter.

The second tricky part was maintaining backwards compatibility. The database module still accepts old `QueryExecer` implementations, and the new context-aware path is optional via `QueryExecerContext`.

### What warrants a second pair of eyes

- The public API change from `runtimebridge.Bindings.Owner runtimeowner.Runner` to `runtimebridge.OwnerRunner` is source-compatible for typical module uses that only call `Post`, but consumers that expected the full `runtimeowner.Runner` from `bindings.Owner` would need to adapt.
- The context stack uses a mutex even though owner calls are serialized; this is intentionally defensive but could be reviewed for simplicity.
- `dbguard` now executes SQL with request context, but guard measurement internals still use non-context `QueryRow`; adding `QueryRowContext` can be a follow-up.

### What should be done in the future

- Add guard check spans now that DB execution context is propagated.
- Add Kanban action/dispatch/render spans using the same propagated request context pattern.
- Decide whether to keep the `replace` directive until a tagged `go-go-goja` release exists, or switch to a pseudo-version once pushed.
- Consider adding a go-go-goja HTTP-level integration test that proves `gojahttp.Host` + database module gives DB calls the HTTP request context directly.

### Code review instructions

Review in two repos.

First, `../go-go-golems/go-go-goja` commit:

```text
4b7aa26c86eae3692551bd05f32713fa5dd12880 feat: propagate owner call context to native modules
```

Focus files:

```text
pkg/runtimebridge/runtimebridge.go
pkg/runtimeowner/runner.go
engine/factory.go
modules/database/database.go
pkg/runtimebridge/runtimebridge_test.go
pkg/runtimeowner/runner_test.go
modules/database/database_test.go
```

Then review `goja-site` local changes:

```text
go.mod
pkg/app/database.go
pkg/dbguard/metered.go
pkg/observability/sql.go
pkg/app/observability_test.go
```

Pay special attention to `TestServerDBSpansAreChildrenOfHTTPSpan`; it is the proof that the feature works end-to-end from HTTP through JavaScript into DB instrumentation.

### Technical details

The new working context flow is:

```text
otelhttp starts HTTP span in r.Context()
gojahttp.Host passes r.Context() to owner.Call(...)
runtimeowner.invoke calls runtimebridge.WithCallContext(vm, ctx, ...)
JS handler calls db.query/db.exec
modules/database loader closure calls runtimebridge.CurrentContext(vm)
DBModule.QueryContext/ExecContext calls QueryExecerContext if available
go-site InstrumentedQueryExecer starts DB span from incoming ctx
sql.DB QueryContext/ExecContext receives the request context
```

### Step 17 follow-up: commits and final validation

After implementing and validating the code, I committed the goja-site changes separately from the ticket documentation:

```text
5b2aca3e47f3f4b2b0379d9ffa3388e526213828 feat: parent database spans to request traces
8ace3000483edbe99de63d7218115d248e17f90f docs: record context propagation implementation
```

I then reran final validation from the goja-site repo:

```text
go test ./...
docmgr doctor --ticket GOJA-PERF-BENCH --stale-after 30
```

Both passed. The only remaining untracked item in the goja-site worktree is the pre-existing unrelated `tmp/` directory.

## Step 18: Build Phase 7 benchmark matrix runner and validate real scenarios/actions

I moved from instrumentation work into proper benchmark methodology. The user explicitly said we do not need to instrument smaller parts right now, so I focused on repeatable scenario/rate runs, valid fixture sites, valid Kanban action payloads, metrics snapshots, summaries, and reproducible scripts.

### Prompt Context

**User prompt (verbatim):** "ok, do it. Keep a detailed diary. Save any script you write in the scripts/ folder of the ticket so that we can retrace your steps as well. Commit at appropriate intervals. Make sure you are testing valid sites and actions."

**Assistant interpretation:** Implement Phase 7 benchmark harness improvements, add valid benchmark fixtures and matrix scripts, store reproducibility scripts under the ticket's `scripts/` directory, run smoke validation, and commit in focused chunks.

**Inferred user intent:** The user wants benchmarking to become repeatable and auditable, with special attention to not benchmarking broken routes or invalid Kanban actions.

**Commit (code):** `5475ea75cdf10bb523056b3abac1cc2199dfe6fd` — "feat: add benchmark matrix runner".

### What I did

Added benchmark matrix and runner artifacts:

- `bench/scenarios.yaml`
- `scripts/bench-matrix.sh`
- `bench/scripts/kanban-board/app.js`
- `bench/targets/kanban-page.txt`
- `bench/targets/kanban-fragment.txt`
- `bench/targets/kanban-action.txt`
- `bench/targets/kanban-mixed.txt`
- `bench/targets/payloads/kanban-move.json`

Enhanced `scripts/bench-vegeta.sh`:

- Added `--warmup-duration`.
- Added Kanban scenarios:
  - `kanban-page`,
  - `kanban-fragment`,
  - `kanban-action`,
  - `kanban-mixed`.
- Added `metadata.json` per run.
- Added `metrics-delta.txt` per run.
- Copies the exact generated `targets.txt` into each result directory.

Added observability deployment starters:

- `deploy/observability/prometheus.example.yaml`
- `deploy/observability/grafana-dashboard-goja-site-benchmark.json`

Added ticket scripts for retracing the benchmark steps:

- `ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/scripts/run-phase7-smoke-matrix.sh`
- `ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/scripts/run-phase7-short-matrix.sh`

Added a smoke report:

- `reference/02-phase7-smoke-benchmark-report.md`

### Why

The project already has enough instrumentation for benchmark interpretation. The more important gap was reproducibility: defined scenarios, valid routes/actions, consistent result directories, matrix summaries, and scripts that can be rerun later.

### What worked

- The new Kanban fixture mounted a real 120-card board at `/_kanban/bench`.
- The `kanban-action` scenario successfully posted a valid `cardMoved` payload to:

```text
POST /_kanban/bench/action/cardMoved
```

- The `kanban-mixed` scenario was verified to hit both fragment and action paths.
- The matrix runner successfully produced:
  - `runs.tsv`,
  - `matrix-summary.json`,
  - `matrix-summary.md`,
  - one full artifact directory per run.

### What didn't work

The first full matrix smoke run produced valid per-scenario run results, but summary generation failed with:

```text
TypeError: str.format() got multiple values for keyword argument 'errors'
```

The bug was caused by passing `errors=...` into `format(...)` while also expanding a row dict that already had an `errors` key. I fixed it by renaming the explicit formatter fields to `status_json` and `errors_json`.

After that fix, the smoke matrix completed and wrote `matrix-summary.md` successfully.

### Validation

Unit/integration validation:

```text
go test ./...
```

Result: passed.

Dedicated valid Kanban action benchmark smoke:

```text
scripts/bench-vegeta.sh --scenario kanban-action --duration 2s --rate 2/s --port 18210 --metrics-port 19210
```

Result: 4 requests, 4 HTTP 200s, 100% success.

Full smoke matrix:

```text
scripts/bench-matrix.sh --scenarios null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed --rates 2/s --duration 2s --warmup-duration 0s --repeat 1 --start-port 18240 --start-metrics-port 19240 --out-root bench/results/phase7-smoke-manual2
```

Result: all seven scenarios returned 100% success with only HTTP 200 status codes.

Kanban mixed action verification after target ordering fix:

```text
scripts/bench-vegeta.sh --scenario kanban-mixed --duration 2s --rate 2/s --port 18260 --metrics-port 19260
```

Metrics delta evidence:

```text
goja_site_kanban_action_duration_seconds_count{action="cardMoved",board="bench",refresh="true",site="default"} 1
goja_site_kanban_fragment_duration_seconds_count{board="bench",site="default"} 2
goja_site_kanban_render_duration_seconds_count{board="bench",reason="action_refresh",site="default"} 1
goja_site_kanban_render_duration_seconds_count{board="bench",reason="fragment",site="default"} 2
```

This proves that the tiny mixed smoke hit a valid action endpoint, not just fragment reads.

### What I learned

- Very short mixed Vegeta runs may only consume the first few targets, so target ordering matters for smoke validation.
- A mixed target file should put at least one valid action near the top so low-request smoke runs exercise it.
- For real benchmark matrices, per-run target snapshots and metadata are as important as raw latency summaries because they make results auditable.

### What was tricky to build

The main subtlety was making the mixed Kanban workload both representative and smoke-friendly. A weighted target list that puts all action targets at the bottom is fine for longer runs, but it can fail to validate actions in short smoke runs. I reordered the mixed targets to include a `cardMoved` POST near the top.

### What warrants a second pair of eyes

- Whether the short matrix defaults should use rates `5/s,10/s,25/s` for all scenarios, or lower defaults for Kanban action/write-heavy scenarios.
- Whether the starter Grafana dashboard should use datasource UID `prometheus` or a variable-based datasource for portability.
- Whether `bench/scenarios.yaml` should become machine-parsed by the matrix runner later.

### What should be done in the future

- Run the short matrix script and write a true baseline report.
- Run saturation rates for the scenario that matters most.
- Use `--pprof` only on targeted reruns where p95 or errors show a knee.
- Add alert-rule examples after a few real baselines establish sensible thresholds.

### Code review instructions

Review these implementation files first:

```text
scripts/bench-vegeta.sh
scripts/bench-matrix.sh
bench/scenarios.yaml
bench/scripts/kanban-board/app.js
bench/targets/kanban-*.txt
```

Then review the retrace scripts:

```text
ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/scripts/run-phase7-smoke-matrix.sh
ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/scripts/run-phase7-short-matrix.sh
```

Finally, review the report:

```text
ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/reference/02-phase7-smoke-benchmark-report.md
```

### Technical details

Each `scripts/bench-vegeta.sh` result directory now contains:

```text
metadata.json
metrics-before.prom
metrics-after.prom
metrics-delta.txt
server.log
summary.md
targets.txt
vegeta.bin
vegeta.json
vegeta.txt
```

Each matrix result root contains:

```text
runs.tsv
matrix-summary.json
matrix-summary.md
<scenario>/rate-<rate>/run-<n>/...
```
