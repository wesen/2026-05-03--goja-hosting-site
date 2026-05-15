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
