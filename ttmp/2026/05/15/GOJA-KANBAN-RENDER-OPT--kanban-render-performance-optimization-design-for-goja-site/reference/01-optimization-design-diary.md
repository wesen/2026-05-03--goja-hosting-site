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
