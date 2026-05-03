# Tasks

## Phase 0 — Ticket planning and delivery hygiene

- [x] Create docmgr ticket workspace for `GOJA-HOSTING-SITE`.
- [x] Create primary design / implementation guide document.
- [x] Create investigation diary.
- [x] Inspect `go-go-goja` runtime, module, database, and fs patterns.
- [x] Inspect `discord-bot` UI DSL pattern.
- [x] Inspect `jesus` historical Express-style server behavior.
- [x] Write intern-friendly implementation guide with diagrams, pseudocode, APIs, phases, risks, and file references.
- [x] Validate ticket with `docmgr doctor`.
- [x] Upload document bundle to reMarkable.

## Phase 1 — Go module and Glazed CLI skeleton

- [x] Create `go.mod` for the new server repository.
- [x] Add local `replace` for `github.com/go-go-golems/go-go-goja` to use `../corporate-headquarters/go-go-goja`.
- [x] Create `cmd/goja-site/main.go` with a Cobra root command.
- [x] Wire Glazed logging section into the root command.
- [x] Implement a Glazed-defined `serve` command.
- [x] Add `--addr`, `--db`, `--scripts`, and `--dev` flags.
- [x] Make `serve --help` available through Glazed/Cobra command construction.

## Phase 2 — Runtime host and script loader

- [x] Add `pkg/app.Config`.
- [x] Add `pkg/app.Server` to own SQLite, go-go-goja runtime, route host, and HTTP server.
- [x] Open and ping SQLite from `--db`.
- [x] Build a go-go-goja runtime through `engine.NewBuilder()`.
- [x] Register a preconfigured non-reconfigurable `database` module.
- [x] Register a preconfigured non-reconfigurable `db` alias module.
- [x] Enable trusted-script support modules: `fs`, `path`, `time`, `timer`.
- [x] Load all `*.js` files from `--scripts` in deterministic lexical order.
- [x] Shut down HTTP server, runtime, and database on close.

## Phase 3 — Express-style HTTP module

- [x] Implement ordered route registry.
- [x] Implement `:param` route matching.
- [x] Register `require("express")` through a runtime module registrar.
- [x] Export `express.app()`.
- [x] Export `app.get`, `app.post`, `app.put`, `app.patch`, `app.delete`, and `app.all`.
- [x] Build JS request object with `method`, `url`, `path`, `query`, `params`, `headers`, `cookies`, `ip`, `body`, and `rawBody`.
- [x] Parse JSON request bodies.
- [x] Parse URL-encoded form request bodies.
- [x] Implement response methods: `status`, `set`, `type`, `json`, `send`, `html`, `redirect`, and `end`.
- [x] Dispatch HTTP handlers through `Runtime.Owner.Call` to avoid concurrent Goja access.
- [x] Add route registry unit tests.
- [x] Add Express module integration tests.

## Phase 4 — HTML UI DSL module

- [x] Create HTML node model: document, element, text, raw HTML, fragment.
- [x] Implement safe HTML renderer.
- [x] Escape text content and attribute values by default.
- [x] Support boolean attributes.
- [x] Support flexible `class` attribute values.
- [x] Register `require("ui.dsl")` through a runtime module registrar.
- [x] Register `require("ui")` alias for convenience.
- [x] Export common HTML element constructors.
- [x] Export `fragment`, `text`, `raw`, `render`, and `page`.
- [x] Integrate returned UI DSL nodes with `res.html` and handler return values.
- [x] Add UI DSL renderer tests.

## Phase 5 — Kanban website in JavaScript

- [x] Create `examples/kanban/scripts/app.js`.
- [x] Define SQLite schema migration in JavaScript.
- [x] Seed initial cards when the database is empty.
- [x] Build complete board HTML through `ui.dsl`.
- [x] Add `GET /` route.
- [x] Add CSS route `GET /style.css`.
- [x] Add JSON API route `GET /api/cards`.
- [x] Add create-card form endpoint `POST /cards`.
- [x] Add move-card endpoint `POST /cards/:id/move`.
- [x] Add example README with run command.

## Phase 6 — Validation, browser testing, and commits

- [x] Run `go mod tidy`.
- [x] Run `gofmt`.
- [x] Run `go test ./... -count=1`.
- [x] Start the Kanban server locally.
- [x] Verify homepage with `curl`.
- [x] Verify JSON API with `curl`.
- [x] Test the Kanban website with Playwright.
- [x] Fix any browser/UI issues found by Playwright.
- [x] Commit implementation at appropriate checkpoints.
- [x] Update diary with commands, failures, and validation evidence.
- [x] Run final `docmgr doctor`.

## Future hardening tasks

- [ ] Add path-scoped filesystem module or document a hard trusted-script boundary in the executable README.
- [ ] Add promise/async route handler support if needed.
- [ ] Add file-watch reload mode.
- [ ] Add TypeScript declarations for `express` and `ui.dsl` modules.
- [ ] Add production security review before running untrusted scripts.

## Phase 7 — Planned client interactivity redesign

- [x] Write design guide for search, precise movement, and UI DSL client behavior support.
- [ ] Add server-rendered search form and query handling.
- [ ] Add precise move form with destination column and target position.
- [ ] Implement `moveCard({ id, toStatus, toIndex })` with position normalization.
- [ ] Add `GET /api/cards?search=&status=&tag=`.
- [ ] Add JSON move endpoint `POST /api/cards/:id/move`.
- [ ] Add app-specific `/app.js` route for live search and drag/drop.
- [ ] Add declarative `data-*` behavior attributes to cards and columns.
- [ ] Add Playwright tests for search, drag/drop/move, reload persistence, and console cleanliness.
- [ ] Extract reusable `ui.clientScript`, `ui.clientState`, and `ui.behavior` helpers after app-specific behavior is proven.
