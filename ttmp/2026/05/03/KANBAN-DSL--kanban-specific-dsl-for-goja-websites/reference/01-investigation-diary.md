---
Title: Investigation Diary
Ticket: KANBAN-DSL
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
RelatedFiles: []
ExternalSources: []
Summary: "Chronological investigation and design diary for the kanban.dsl ticket."
LastUpdated: 2026-05-03T15:30:00-04:00
WhatFor: "Record why kanban.dsl is designed as a fluid builder with server-side callbacks and a browser runtime."
WhenToUse: "Read before implementing pkg/kanbanddsl or migrating the Kanban example."
---

# Diary

## Goal

This diary records the design work for a Kanban-specific DSL that runs in server-side Goja JavaScript, renders flexible boards with `ui.dsl`, ships a generic browser runtime, and dispatches browser actions such as `cardMoved` back into server-side callbacks.

## Step 1: Created KANBAN-DSL ticket and architecture guide

I created a new docmgr ticket for `kanban.dsl` and wrote a long-form architecture and implementation guide. The design starts from the current `goja-site` implementation: `ui.dsl` renders safe HTML, `express` handles routes, and the Kanban example currently hand-writes search, movement, JSON endpoints, and client JavaScript.

The guide proposes a runtime-scoped native Go module, `require("kanban.dsl")`, that owns generic board mechanics while letting app authors keep control over data access, card rendering, and server-side callbacks.

### Prompt Context

**User prompt (verbatim):** "let's invent a kanban specific "kanban.dsl" module to avoid having to write client side JS in the app itself, that however leaves us a lot of flexibility in designing various kanban boards, and gives us the freedom to mix it with ui.dsl components too.

Create a new docmgr ticket for a kanban specific DSL that for example allows action callback (such as cardMoved, etc...) to be written in the server side goja JS, and the frontend manages to know how to call it.

This is quite special, so think hard and brainstorm wide and do some really intelligent work here.

 Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the                                                                                      
  system needed to understand what it is, with prose paragraphs and                                                                                                                                         
bullet                                                                                                                                                                                                      
  point sand pseudocode and diagrams and api references and                                                                                                                                                 
file                                                                                                                                                                                                        
                                                                                                                                                                                                            
references.                                                                                                                                                                                                 
    It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Design a special-purpose Kanban DSL module that removes hand-written client JS from apps while preserving flexible server-side rendering and callbacks.

**Inferred user intent:** The user wants a serious architecture that bridges browser Kanban interactions to server-side Goja callbacks in a reusable, ergonomic, and enforceable way.

**Commit (code):** N/A — documentation/design step.

### What I did

- Created ticket `KANBAN-DSL`.
- Created design doc `01-kanban-dsl-architecture-and-implementation-guide.md`.
- Inspected current implementation files for evidence and references.
- Wrote a proposed `kanban.dsl` architecture with rendering, mounting, callback protocol, browser runtime, phases, risks, and file references.

### Why

- The current Kanban example has too much hand-written app-specific browser code.
- The desired abstraction needs careful protocol design: browser events must map safely to server-side Goja callbacks.

### What worked

- The current `ui.dsl` and `express` surfaces are enough to build on.
- The server already serializes Goja request dispatch through the runtime owner.

### What didn't work

- No implementation was attempted in this design step.

### What I learned

- The key abstraction is not a fixed Kanban schema. It is a board runtime with data access hooks, render hooks, and callback hooks.

### What was tricky to build

- The browser must not know app-specific data structures, yet it must send enough information for precise movement. The event envelope therefore includes card ID, from/to columns and indices, visible card IDs, and query state.

### What warrants a second pair of eyes

- Callback protocol shape.
- Fragment refresh strategy.
- Whether mounting should call existing Express app methods or integrate deeper into `web.Host`.

### What should be done in the future

- Implement the module in phases from `tasks.md`.

### Code review instructions

- Start with the design document, then inspect `pkg/uidsl`, `pkg/web`, `pkg/app/server.go`, and `examples/kanban/scripts/app.js`.

### Technical details

Primary design doc:

```text
ttmp/2026/05/03/KANBAN-DSL--kanban-specific-dsl-for-goja-websites/design-doc/01-kanban-dsl-architecture-and-implementation-guide.md
```

## Step 2: Added examples and fluid builder API

After drafting the object-style API, the user asked for example apps and then requested a fluid builder API so Go can enforce the DSL more strongly. I expanded the document with multiple app examples and then updated the API recommendation to prefer a builder with typed sub-builders and `.build()` validation.

The most important change is that `kanban.board(id)` should return a mutable `BoardBuilder`, not immediately accept one huge options object. The builder collects columns, data hooks, render hooks, features, and action callbacks; `.build()` validates and freezes the board.

### Prompt Context

**User prompt (verbatim):** "show me a couple of example kanban apps i can create with thta DSL. Then add them to the doc."

**User prompt (verbatim):** "use a fluid builder api so that we can do some strong enforcement on the go side. Update the document"

**Assistant interpretation:** Expand the guide with concrete use cases, then reshape the proposed API around a chainable builder to support strong Go-side validation.

**Inferred user intent:** The user wants the DSL to feel ergonomic but also robust, with guardrails that catch mistakes before runtime interactions begin.

**Commit (code):** N/A — documentation/design step.

### What I did

- Added example Kanban apps:
  - trail planning / Field Notes board,
  - editorial publishing pipeline,
  - CRM sales pipeline,
  - personal habit board.
- Added builder-style versions of the trail, editorial, and sales examples.
- Added a detailed fluid builder API section.
- Added Go-side enforcement rules and type-state-like sub-builder design.
- Updated tasks to include builder implementation phases.

### Why

- A large object literal is too loose for a DSL that coordinates server callbacks and browser actions.
- Builder sub-objects let Go expose smaller method sets and produce better validation errors.

### What worked

- The builder API maps well onto Go structs: `BoardBuilder`, `ColumnListBuilder`, `ColumnBuilder`, `DataBuilder`, `FeatureBuilder`, `RenderBuilder`, and `ActionBuilder`.
- Feature methods can imply validation requirements, such as `dragDrop()` requiring `actions.cardMoved()` unless the board is read-only.

### What didn't work

- The object-literal examples still exist in the document as conceptual examples. They should be treated as legacy/low-level or rewritten during future cleanup.

### What I learned

- For this DSL, `.build()` is not cosmetic. It is the moment when the Go side can aggregate missing requirements, reject duplicates, and freeze callback registration.

### What was tricky to build

- JavaScript cannot enforce compile-time types, but Go can simulate type-state by returning different builder objects with different available methods.

### What warrants a second pair of eyes

- Whether `.mount(app, prefix)` should auto-build or require explicit `.build()`.
- Whether a low-level `fromSpec` escape hatch is worth exposing in v1.

### What should be done in the future

- Implement builder validation before implementing browser runtime details.
- Keep examples in builder style so interns copy the right API.

### Code review instructions

- Review the fluid builder section in the design doc first.
- Then review the implementation tasks in `tasks.md`.

### Technical details

The preferred final shape is:

```javascript
const board = kanban.board("trail-notes")
  .columns(cols => cols.column("todo").title("To Do").done())
  .data(data => data.cards(loadCards).id(card => String(card.id)).column(card => card.status))
  .features(features => features.search().preciseMove().dragDrop())
  .actions(actions => actions.cardMoved(onCardMoved))
  .build();
```
