# Changelog

## 2026-05-03

- Initial workspace created


## 2026-05-03

Added concrete example applications showing trail planning, editorial, sales CRM, and habit boards built with kanban.dsl.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-DSL--kanban-specific-dsl-for-goja-websites/design-doc/01-kanban-dsl-architecture-and-implementation-guide.md — Expanded with example apps


## 2026-05-03

Updated kanban.dsl guide to prefer a fluid builder API with strong Go-side validation and type-state-like sub-builders.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-DSL--kanban-specific-dsl-for-goja-websites/design-doc/01-kanban-dsl-architecture-and-implementation-guide.md — Fluid builder API design added


## 2026-05-03

Added builder-style Kanban app examples for trail planning, editorial, and sales boards.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-DSL--kanban-specific-dsl-for-goja-websites/design-doc/01-kanban-dsl-architecture-and-implementation-guide.md — Builder-style examples added


## 2026-05-03

Updated tasks and diary for the fluid builder API design.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-DSL--kanban-specific-dsl-for-goja-websites/tasks.md — Builder implementation phases added


## 2026-05-03

Uploaded updated KANBAN-DSL Architecture Guide bundle to reMarkable at /ai/2026/05/03/KANBAN-DSL.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-DSL--kanban-specific-dsl-for-goja-websites/tasks.md — Marked reMarkable upload complete


## 2026-05-03

Implemented initial pkg/kanbanddsl native module with fluid builder, Go-owned client runtime, mount routes, dispatch, rendering, and tests.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/app/server.go — Registered kanban.dsl runtime module
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/kanbanddsl — Initial kanban.dsl implementation
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-DSL--kanban-specific-dsl-for-goja-websites/reference/01-investigation-diary.md — Recorded implementation step


## 2026-05-03

Migrated the Kanban example to use kanban.dsl and removed app-owned browser Kanban runtime.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/examples/kanban/README.md — Documented DSL-owned frontend runtime
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/examples/kanban/scripts/app.js — Now uses fluid kanban.dsl builder and board.mount
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/kanbanddsl/mount_test.go — Added mounted client/action HTTP integration coverage


## 2026-05-03

Updated task checklist to mark implemented kanban.dsl builder, mount, runtime, callback, and example migration work.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-DSL--kanban-specific-dsl-for-goja-websites/tasks.md — Marked implementation tasks complete


## 2026-05-03

Browser-validated live search, precise move, and drag/drop; added opt-in kanban client debug logging.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/kanbanddsl/client_runtime.go — Added debug logging and fixed board lookup for toolbar search
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-DSL--kanban-specific-dsl-for-goja-websites/reference/01-investigation-diary.md — Recorded browser validation results


## 2026-05-03

Fixed browser drag start by rendering explicit draggable true and injecting runtime drag CSS; verified Playwright dragTo now moves cards.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/kanbanddsl/builder_test.go — Assert explicit draggable output
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/kanbanddsl/client_runtime.go — Inject runtime drag styles and avoid document.head-only style insertion
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/kanbanddsl/render.go — Render explicit draggable true for draggable cards


## 2026-05-03

Propagated host sessions through mounted Kanban render/action routes and session-scoped the example board.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/examples/kanban/scripts/app.js — Uses session_id for scoped card data
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/kanbanddsl/mount.go — Passes session into render contexts and action events

