---
Title: Kanban render optimization design diary
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - goja
    - kanban
    - performance
    - optimization
DocType: reference
Intent: historical
Summary: "Chronological diary for the Kanban render optimization design package."
---

# Kanban render optimization design diary

## Step 1: Create optimization design ticket and write intern-ready guides

Created ticket `GOJA-KANBAN-RENDER-OPT` to hold the design work that follows from the single-VM and multi-VM stress results.

### Prompt context

The user asked for a detailed analysis/design/implementation guide for a new intern, with prose, bullets, pseudocode, diagrams, API references, and file references. The user also asked for a separate holistic architectural optimization guide that explains deeper changes to the way hosted sites run JavaScript inside Goja sandboxes and interact with the browser/display protocol. Both guides should use a textbook writing style and avoid analogies.

### Evidence used

The guides are based on:

- single-VM `kanban-action` knee around 80/s,
- single-VM pprof showing `uidsl.renderNode`, `renderAttrs`, `preciseMoveForm`, allocation, and GC,
- multi-VM `kanban-fragment` saturation around 200-400/s depending on VM count,
- multi-VM pprof showing the same render/allocation hotspots under four active VMs,
- source files in `pkg/kanbanddsl`, especially `render.go`, `mount.go`, `client_runtime.go`, `builder.go`, and `types.go`,
- benchmark fixture `bench/scripts/kanban-board/app.js`.

### Documents written

```text
design/01-kanban-render-optimization-implementation-guide.md
design/02-holistic-goja-hosting-performance-architecture-guide.md
```

The first guide focuses on the immediate optimization: make precise move rendering optional/lazy while preserving compatibility.

The second guide discusses broader architectural options: reducing generated markup, caching, partial HTML patches, semantic event responses, optional client-side site code, and VM architecture choices.

### Current recommendation

Do the smallest measured optimization first: add a backward-compatible precise-move render mode, keep eager rendering as the default, add an optimized benchmark fixture, and compare baseline vs optimized under the existing single-VM and multi-VM harnesses.

## Step 2: Upload design guides to reMarkable

Uploaded the two design guides as a bundled PDF.

### Command

```text
remarquee upload bundle design/01-kanban-render-optimization-implementation-guide.md design/02-holistic-goja-hosting-performance-architecture-guide.md --name "GOJA KANBAN RENDER OPT Design Guides" --remote-dir "/ai/2026/05/15/GOJA-KANBAN-RENDER-OPT" --toc-depth 2 --non-interactive
```

### Result

```text
OK: uploaded GOJA KANBAN RENDER OPT Design Guides.pdf -> /ai/2026/05/15/GOJA-KANBAN-RENDER-OPT
```

## Step 3: Remove eager precise move forms and add frontend accessibility actions

The user clarified that backward compatibility is not required and that the DSL should remain opinionated and tight. Based on that, I changed direction from configurable `preciseMove` modes to removing the eager precise move form path from the core DSL.

### What changed

Implementation files changed:

```text
pkg/kanbanddsl/types.go
pkg/kanbanddsl/builder.go
pkg/kanbanddsl/render.go
pkg/kanbanddsl/client_runtime.go
pkg/kanbanddsl/builder_test.go
pkg/kanbanddsl/mount_test.go
bench/scripts/kanban-board/app.js
```

Concrete changes:

- Removed `FeatureSpec.PreciseMove`.
- Removed `features.preciseMove()` from the builder API.
- Removed server-rendered per-card `preciseMoveForm` generation.
- Updated the benchmark fixture to use `.features(features => features.search().dragDrop())`.
- Added a small default `Actions` button per movable card instead of a full form.
- Added card accessibility attributes: `role="listitem"`, `tabindex="0"`, and an `aria-label` with card, column, and position.
- Added list semantics to the card list container.
- Added frontend action menu behavior in `client_runtime.go`:
  - click `Actions` to open a menu,
  - keyboard Enter/Space on focused card opens the menu,
  - Escape closes and restores focus,
  - ArrowUp/ArrowDown navigate menu items,
  - menu supports Move up, Move down, Move to top, Move to bottom, and Move to another column,
  - results are announced through an `aria-live` region,
  - focus returns to the moved card after the server response refreshes HTML.

### Why this direction

The previous design would have added modes such as `preciseMove("none")`, `preciseMove("button")`, and `preciseMove("lazy")`. That would preserve compatibility, but it would make the DSL expose render strategy choices. The updated direction makes the DSL more opinionated: movement is a semantic Kanban action, and accessibility belongs primarily in the client runtime rather than in hundreds of server-rendered forms.

### Validation

Ran:

```text
go test ./pkg/kanbanddsl
go test ./...
```

Both passed.

## Step 4: Run post-simplification benchmarks

I added and ran a focused post-simplification benchmark script.

### Script

```text
scripts/02-run-post-simplification-benchmarks.sh
```

### Command

```text
ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/scripts/02-run-post-simplification-benchmarks.sh
```

### Result root

```text
bench/results/kanban-simplified-20260515T173954Z
```

### Results

```text
single kanban-fragment 200/s:
  throughput 200.03/s
  p95 9.887 ms
  mean response bytes 60,831

single kanban-action 100/s:
  throughput 100.06/s
  p95 8.165 ms
  mean response bytes 77,035

multi 4VM kanban-fragment 400/s:
  throughput 399.76/s
  p95 7.760 ms
  mean response bytes 60,831
```

### Interpretation

This is a large improvement. The previous multi-VM 4VM `kanban-fragment` 400/s cell had p95 around 5.86 seconds and throughput around 246.53/s. After removing eager precise move forms, the same shape reached the offered 400/s with p95 under 8 ms.

The result confirms the pprof diagnosis: eager precise move forms were creating a large amount of repeated UI DSL node/attribute/rendering work and large responses.

### Remaining work

The performance change is validated by Go tests and load tests. The accessibility behavior still needs browser-level testing for keyboard navigation, focus restoration, and live-region announcements.

## Step 5: Add and run Playwright accessibility validation

Added a dedicated Playwright accessibility smoke test for the new frontend card action menu.

### Script

```text
scripts/03-run-kanban-accessibility-playwright.sh
```

The script starts `goja-site` against `bench/scripts/kanban-board`, installs Playwright in a temporary directory, opens the benchmark board, and verifies:

- card 1 has `role="listitem"`, `tabindex="0"`, and an informative `aria-label`,
- keyboard focus on the card plus Enter opens the action menu,
- the menu has `role="menu"`,
- ArrowDown and Escape work inside the menu,
- clicking `Move to Done` posts the action,
- the refreshed DOM shows card 1 in the `done` column,
- focus is restored to card 1,
- the live region announces `Moved card 1`,
- the browser console has no warnings, errors, or page errors.

I also updated the existing `scripts/playwright-kanban-smoke.sh` and `examples/kanban/scripts/app.js` so the older example site uses the new compact action menu rather than removed `preciseMove()` forms.

### Validation commands

```text
ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/scripts/03-run-kanban-accessibility-playwright.sh
scripts/playwright-kanban-smoke.sh
go test ./...
```

### Results

```text
kanban accessibility playwright smoke passed: http://127.0.0.1:19220
kanban playwright smoke passed: http://127.0.0.1:19111
go test ./... passed
```

## Step 6: Write anti-overfit benchmark plan

The user asked for a detailed document listing the benchmark classes we should add so we do not overfit to the Kanban fixture, plus a second section prioritizing benchmark construction and execution order by ease, instrumentation effort, and value.

### Document

```text
design/03-anti-overfit-benchmark-plan.md
```

### Contents

The plan covers:

- Kanban size scaling,
- UI shape benchmarks,
- interaction benchmarks,
- database-heavy benchmarks,
- mixed multi-site/multi-VM benchmarks,
- startup and reload benchmarks,
- memory and soak benchmarks,
- browser/frontend benchmarks,
- error and pathological benchmarks,
- observability overhead benchmarks.

It then prioritizes implementation order, starting with formalizing the post-simplification regression benchmarks, then adding Kanban size scaling, UI shape renderer fixtures, DB-heavy fixtures, mixed multi-site runs, browser timings, observability overhead, startup/reload, soak, and error-path tests.

### Recommended immediate deliverable

The proposed first anti-overfit matrix is:

```text
kanban-fragment-10
kanban-fragment-120
kanban-fragment-500
render-flat-1000
render-attrs-1000
db-read-100
db-write-batch-10

rates: 25/s, 100/s, 250/s
repeat: 3
```

## Step 7: Build and run Anti-overfit matrix v1

Built the first anti-overfit benchmark fixtures and runner.

### Fixtures and harness changes

- Added `bench/scripts/render-shapes/app.js` with `render-flat-1000` and `render-attrs-1000` routes.
- Added generated Kanban fixtures under `bench/scripts/kanban-board-10` and `bench/scripts/kanban-board-500`.
- Extended `scripts/bench-vegeta.sh` with scenarios for Kanban sizes, render shapes, `db-read-100`, and `db-write-batch-10`.
- Added `scripts/04-run-anti-overfit-matrix.sh` and `scripts/05-render-anti-overfit-report.py` to this ticket.

### Run

```text
MATRIX_ID=anti-overfit-v1-20260515T182902Z \
  ttmp/2026/05/15/GOJA-KANBAN-RENDER-OPT--kanban-render-performance-optimization-design-for-goja-site/scripts/04-run-anti-overfit-matrix.sh
```

Shape: seven scenarios, three rates, three repeats, for 63 measured runs. Each run used `15s` measured duration and `3s` warmup.

### Report

```text
reference/03-anti-overfit-benchmark-report.md
archive/anti-overfit-v1-20260515T182902Z/01-matrix-summary.json
archive/anti-overfit-v1-20260515T182902Z/01-matrix-summary.md
```

### Main findings

- `kanban-fragment-10` stayed cheap through `250/s`: p95 avg `2.14 ms`.
- `kanban-fragment` stayed healthy at `100/s`: p95 avg `9.43 ms`; at `250/s` it showed queueing but held near full throughput.
- `kanban-fragment-500` was healthy at `25/s`, but saturated by `100/s`.
- `render-attrs-1000` is the strongest next render bottleneck: it saturated at `100/s` and timed out at `250/s`.
- `db-write-batch-10` queued already at `100/s`, indicating write-heavy SQLite behavior is a separate benchmark/optimization line.

## Step 8: Upload anti-overfit benchmark bundle to reMarkable

Uploaded the anti-overfit benchmark plan and Anti-overfit matrix v1 report as one PDF bundle.

```text
/ai/2026/05/15/GOJA-KANBAN-RENDER-OPT/GOJA KANBAN RENDER OPT Anti Overfit Benchmarks.pdf
```

## Step 9: Profile both anti-overfit follow-up candidates

The user asked, "why not both," so I profiled both candidate cells from the anti-overfit matrix.

### Captures

```text
archive/pprof-anti-overfit-20260515T191339Z/render-attrs-1000-100_s
archive/pprof-anti-overfit-20260515T191339Z/kanban-fragment-500-100_s
```

Both used `30s` measured duration, `5s` warmup, and `10s` CPU profile capture. Raw `vegeta.bin` files were removed after generating JSON/text/profile artifacts.

### Report

```text
reference/04-anti-overfit-follow-up-pprof-report.md
```

### Finding

`render-attrs-1000` reproduced severe queueing at `100/s` and points strongly at Goja object export, UI DSL element conversion, attribute rendering, allocation, and GC. `kanban-fragment-500` did not reproduce the short matrix's severe 100/s queueing in this longer profile run; it held `100/s` with p95 around `66 ms`, making it a useful regression workload but a weaker immediate optimization target.

Recommendation: optimize generic UI DSL object/attribute conversion first, then keep both workloads in the validation set.

## Step 10: Upload follow-up pprof report to reMarkable

Uploaded the comparative follow-up pprof report.

```text
/ai/2026/05/15/GOJA-KANBAN-RENDER-OPT/GOJA KANBAN RENDER OPT Follow Up Pprof.pdf
```

## Step 11: Investigate render-attrs-1000 bottleneck in detail

Created a source-level investigation for the `render-attrs-1000` performance issue.

### Document

```text
design/04-render-attrs-1000-performance-investigation.md
```

### Core diagnosis

The bottleneck is a UI DSL bridge problem, not a Kanban-specific problem. The current `ui.dsl` path uses generic Goja `Value.Export()` in the hot path for attribute detection and extraction. In `elementFromCall`, attrs are effectively exported twice: first through `isAttrs(args[0])`, then again through `args[0].Export()` to obtain the map. The pprof line listings show this as two large cumulative branches under `elementFromCall`.

The recommended first implementation slice is to replace `isAttrs` plus second `Export()` with one `attrsFromValue` path, then add focused microbenchmarks and compatibility tests around attrs and child disambiguation.

## Step 12: Implement first render-attrs optimization slice

Implemented the conservative first slice in `../go-go-golems/go-go-goja`: avoid double `Export()` for `ui.dsl` attrs detection/extraction.

### Commit

```text
2bc72d5f04f2ea3c47b2394484d523b4c26bbf1c
perf: avoid double export for ui attrs
```

### Changes

- Added attr compatibility tests.
- Added child-vs-attrs disambiguation tests.
- Added `ui.dsl` microbenchmarks for attr element calls and 1000-node render pages.
- Replaced `isAttrs` plus second `Export()` with `attrsFromValue`, which exports once and returns the attrs map.
- Avoided allocating an empty attrs map for elements without attrs.

### Validation

`go-go-goja` pre-commit passed `golangci-lint`, `go generate ./...`, and `go test ./...`.

### Macro result

The follow-up `render-attrs-1000` 100/s pprof run did not improve end-to-end throughput/latency. It reduced alloc-space shape (`Object.Export` and `elementFromCall` roughly halved), but the path still saturated badly. This means the next useful slice should be a render-ready attr representation rather than more small changes around `attrsFromValue`.

Report:

```text
reference/05-render-attrs-first-optimization-report.md
```
