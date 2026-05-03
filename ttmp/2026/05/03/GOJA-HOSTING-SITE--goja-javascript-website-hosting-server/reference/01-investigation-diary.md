---
Title: Investigation Diary
Ticket: GOJA-HOSTING-SITE
Status: active
Topics:
    - go
    - goja
    - javascript
    - web
    - glazed
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../corporate-headquarters/discord-bot/internal/jsdiscord/ui_module.go
      Note: Investigation evidence for UI DSL registrar design
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/engine/factory.go
      Note: Investigation evidence for runtime assembly decisions
    - Path: ../../../../../../../corporate-headquarters/jesus/pkg/engine/engine.go
      Note: Investigation evidence for historical JavaScript server behavior
ExternalSources: []
Summary: Chronological investigation record for the Goja JavaScript website hosting server design ticket.
LastUpdated: 2026-05-03T13:50:00-04:00
WhatFor: Record what was inspected and produced while designing the Goja hosting server.
WhenToUse: Read before resuming implementation so the next engineer understands evidence, decisions, and remaining risks.
---


# Diary

## Goal

This diary records the setup and investigation work for a new `docmgr` ticket about a small Go server that hosts JavaScript websites via `go-go-goja`, Glazed commands, SQLite, filesystem access, an Express-style route module, and an HTML UI DSL.

## Step 1: Created ticket and wrote the design guide

I created the `GOJA-HOSTING-SITE` ticket, added a primary design document and this investigation diary, then inspected the current `go-go-goja`, `discord-bot`, and `jesus` repositories to ground the design in existing code. The resulting design document is intended for a new intern: it explains the runtime architecture, command structure, module contracts, HTTP module, UI DSL, Kanban example, implementation phases, and tests.

The most important decision was to use `../corporate-headquarters/jesus` as historical reference only. The new implementation should use the evolved `../corporate-headquarters/go-go-goja/engine` factory and runtime module registrar APIs rather than copying Jesus' older runtime assembly approach.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket for building a little go server (using glazed framework for the commands) that takes a db and directory of js scripts using the ../corporate-headquarters/go-go-goja framework, with the database module and fs modules enabled, and by also registering a simplified express.js style HTTP server module, so that we can write little websites in JS and have them served.

For that, also create a UI DSL module that can generate HTML (see for example how such a UI DSL is used in the ../corporate-headquarters/discord-bot package). I think we already had something similar in ../corporate-headquarters/jesus, but we have evolved go-go-goja since.

For example, we should be able to bild a kanban application in JS entirely with the db and the express endpoints, returnning ui.dsl stuff that is rendered to HTML.

 Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the                                                                                      
  system needed to understand what it is, with prose paragraphs and                                                                                                                                         
bullet                                                                                                                                                                                                      
  point sand pseudocode and diagrams and api references and                                                                                                                                                 
file                                                                                                                                                                                                        
                                                                                                                                                                                                            
references.                                                                                                                                                                                                 
    It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a docmgr research/design ticket, produce an intern-friendly technical implementation guide for a Go + goja website hosting server, and upload the result to reMarkable.

**Inferred user intent:** The user wants a concrete, evidence-backed plan that an intern can follow to implement the server without needing to reverse-engineer go-go-goja, Jesus, or the Discord UI DSL from scratch.

**Commit (code):** N/A — documentation-only ticket setup and design work.

### What I did

- Ran `docmgr status --summary-only` to confirm the local docmgr root.
- Created ticket `GOJA-HOSTING-SITE` with title `Goja JavaScript Website Hosting Server`.
- Added a design document: `design-doc/01-goja-javascript-website-hosting-server-design-and-implementation-guide.md`.
- Added this diary: `reference/01-investigation-diary.md`.
- Inspected `../corporate-headquarters/go-go-goja` for:
  - native module interface and registry,
  - runtime factory,
  - runtime module registrars,
  - module middleware,
  - database module,
  - filesystem module.
- Inspected `../corporate-headquarters/discord-bot/internal/jsdiscord` for the existing `ui` module registrar and chainable builder pattern.
- Inspected `../corporate-headquarters/jesus` for historical Express-style request/response and dynamic route dispatch behavior.
- Wrote the primary design guide with diagrams, pseudocode, API sketches, file references, implementation phases, testing strategy, risks, and open questions.

### Why

- The design needed to be based on the current `go-go-goja` architecture, not on older assumptions from `jesus`.
- The intern needs file-level references to understand where patterns already exist.
- The user specifically asked for prose paragraphs, bullet points, pseudocode, diagrams, API references, and file references.

### What worked

- `docmgr ticket create-ticket` successfully created the ticket workspace.
- `docmgr doc add` created the primary design document and diary.
- Repository inspection found direct evidence for all major design claims:
  - `modules.NativeModule` in `go-go-goja/modules/common.go`.
  - runtime factory and registrars in `go-go-goja/engine/factory.go` and `engine/runtime_modules.go`.
  - `database` and `fs` module APIs in `go-go-goja/modules/database/database.go` and `modules/fs/fs.go`.
  - Discord UI module registrar and builders in `discord-bot/internal/jsdiscord/ui_module.go`, `ui_embed.go`, and `ui_message.go`.
  - Jesus Express-like request/response and routing in `jesus/pkg/engine/handlers.go`, `jesus/pkg/engine/engine.go`, and `jesus/pkg/web/router.go`.

### What didn't work

- No implementation was attempted in this step, so there were no Go test failures.
- One bulk line-number inspection command produced truncated output because the combined output exceeded the tool limit. I reran smaller targeted inspections for the key files.
- An `edit` attempt to replace `uidss.NewRegistrar()` failed because the old text occurred twice and the edit tool requires unique matches. I corrected the typo with a small Python replacement script.

Exact failed edit output:

```text
Found 2 occurrences of edits[0] in ttmp/2026/05/03/GOJA-HOSTING-SITE--goja-javascript-website-hosting-server/design-doc/01-goja-javascript-website-hosting-server-design-and-implementation-guide.md. Each oldText must be unique. Please provide more context to make it unique.
```

### What I learned

- `go-go-goja` now has a clean `engine.FactoryBuilder` with module middleware and runtime-scoped registrars, which is a better fit than Jesus' older direct `goja.New()` setup.
- The `database` module can be configured or preconfigured; for this server, preconfigured and non-reconfigurable is the safer design.
- The `fs` module exposes broad filesystem access, so the v1 design must be documented as trusted-script only unless path scoping is added.
- The Discord bot UI DSL is a strong pattern for the HTML DSL: export constructors from a runtime module, use builder proxies for higher-level structures, normalize JS values before rendering.

### What was tricky to build

- The main design tension is between familiarity and correctness. Express-like APIs are easy for JS developers, but full Express compatibility would create a large surface. The guide therefore specifies a deliberately small Express subset.
- Another tricky point is Goja concurrency. Net/http is concurrent, while Goja runtime access must be serialized through the runtime owner. The design explicitly routes HTTP handler execution through `Runtime.Owner.Call`.
- Database module composition needs care. If the default registry's `database` module is used directly, scripts may need to call `configure()`. The design recommends registering a preconfigured database module instance so the CLI `--db` flag remains authoritative.

### What warrants a second pair of eyes

- Confirm the exact current `go-go-goja` API for registering a custom preconfigured `database.DBModule` alongside `UseModuleMiddleware` without duplicate module IDs.
- Confirm whether the `database` alias `db` is registered in the current default registry or needs explicit registration.
- Review whether JS route handlers should support promises in v1, because async support affects response lifetime and event loop behavior.
- Review whether the filesystem module must be path-scoped before the tool is used outside trusted local scripts.

### What should be done in the future

- Implement Phase 1 through Phase 6 from the design guide.
- Add a dedicated `docs/javascript-api.md` once the actual module API is implemented and tested.
- Add a security note in the README before shipping any executable server.

### Code review instructions

- Start by reading the design guide at `ttmp/2026/05/03/GOJA-HOSTING-SITE--goja-javascript-website-hosting-server/design-doc/01-goja-javascript-website-hosting-server-design-and-implementation-guide.md`.
- Then inspect the referenced source files in:
  - `../corporate-headquarters/go-go-goja/modules/common.go`
  - `../corporate-headquarters/go-go-goja/engine/factory.go`
  - `../corporate-headquarters/go-go-goja/engine/runtime_modules.go`
  - `../corporate-headquarters/go-go-goja/modules/database/database.go`
  - `../corporate-headquarters/go-go-goja/modules/fs/fs.go`
  - `../corporate-headquarters/discord-bot/internal/jsdiscord/ui_module.go`
  - `../corporate-headquarters/jesus/pkg/engine/handlers.go`
- Validate documentation hygiene with `docmgr doctor --ticket GOJA-HOSTING-SITE --stale-after 30`.

### Technical details

Key commands run:

```bash
docmgr status --summary-only
docmgr ticket create-ticket --ticket GOJA-HOSTING-SITE --title "Goja JavaScript Website Hosting Server" --topics go,goja,javascript,web,glazed
docmgr doc add --ticket GOJA-HOSTING-SITE --doc-type design-doc --title "Goja JavaScript Website Hosting Server Design and Implementation Guide"
docmgr doc add --ticket GOJA-HOSTING-SITE --doc-type reference --title "Investigation Diary"
find ../corporate-headquarters/go-go-goja -maxdepth 3 -type f | head -80
find ../corporate-headquarters/discord-bot -maxdepth 4 -type f | head -120
find ../corporate-headquarters/jesus -maxdepth 4 -type f | head -120
rg -n "Register|NativeModule|database|fs|module|require\(" ../corporate-headquarters/go-go-goja/{engine,modules,testdata} ../corporate-headquarters/discord-bot/internal/jsdiscord ../corporate-headquarters/jesus/pkg -S
```
