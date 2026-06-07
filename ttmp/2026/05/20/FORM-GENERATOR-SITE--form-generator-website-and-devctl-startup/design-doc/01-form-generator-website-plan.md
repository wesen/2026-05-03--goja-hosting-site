---
Title: Form generator website plan
Ticket: FORM-GENERATOR-SITE
Status: active
Topics:
    - goja-site
    - javascript
    - web
    - documentation
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: .devctl.yaml
      Note: Local profile defaults described by the design
    - Path: devctl/goja-site-plugin.py
      Note: Startup implementation described by the design
    - Path: sites/forms/scripts/app.js
      Note: Form generator app described by the design
ExternalSources: []
Summary: Plan and current design for the Form Studio form generator site and its devctl startup workflow.
LastUpdated: 2026-05-20T16:24:00-04:00
WhatFor: Explains how the form generator website is structured and how devctl starts it locally.
WhenToUse: Use when reviewing, extending, or operating the Form Studio app.
---


# Form generator website plan

## Executive Summary

Form Studio is a goja-site-hosted JavaScript website for building small forms, previewing them, and collecting responses in SQLite. The repository now has a `devctl` startup path that runs this app locally with one command:

```bash
devctl up --force
```

The app is served at `http://127.0.0.1:60131/`, uses `sites/forms/scripts/app.js`, and stores local state in `tmp/form-generator.db`.

## Problem Statement

The hosting repo needs a repeatable local startup workflow before building more application examples. It also needs a concrete, useful website that demonstrates the host's JavaScript APIs beyond Kanban demos. A form generator is a good fit because it exercises routing, SQLite persistence, HTML UI generation, POST body handling, and JSON APIs without requiring new Go modules.

## Proposed Solution

### Devctl startup

Add a repo-local devctl plugin with three protocol-v2 operations:

- `config.mutate`: publishes default ports, script path, database path, base URL, and metrics URL.
- `validate.run`: checks for `go`, core Go files, and the form script directory.
- `launch.plan`: asks devctl to supervise `go run ./cmd/goja-site serve ...`.

The launch command is:

```bash
go run ./cmd/goja-site serve \
  --addr 127.0.0.1:60131 \
  --db tmp/form-generator.db \
  --scripts sites/forms/scripts \
  --dev \
  --metrics-addr 127.0.0.1:60132 \
  --service-name goja-site-devctl
```

### Form generator application

The Form Studio JavaScript app uses:

- `database` for SQLite tables and queries.
- `express` for routes and POST handlers.
- `ui.dsl` for server-rendered HTML.

It creates three tables:

- `forms`: form title, description, slug, timestamps.
- `form_fields`: form schema fields with label, generated name, type, placeholder, options, required flag, and position.
- `form_responses`: submitted response JSON payloads.

The app includes:

- `GET /`: dashboard with existing forms and a create-form panel.
- `POST /forms`: create a form and redirect to its builder.
- `GET /forms/:id`: builder page with field editor, field list, live preview, and responses.
- `POST /forms/:id/fields`: add a field.
- `GET /f/:slug`: public respondent view.
- `POST /f/:slug/responses`: store a response.
- `GET /api/forms`: list forms.
- `GET /api/forms/:id`: return a form with fields and responses.

## Design Decisions

- **Single-site devctl target first:** `serve-multi` currently fails because an existing demo script references a missing `kanban.dsl.preciseMove` feature. The devctl workflow therefore starts the new form generator directly with `goja-site serve`.
- **SQLite JSON responses:** Response payloads are stored as JSON text. This keeps the first version flexible while schema editing is still simple.
- **Server-rendered UI:** `ui.dsl` avoids frontend build tooling and validates the hosting site's current strengths.
- **Generated slugs and field names:** Titles become URL slugs; labels become response keys. This gives predictable URLs and JSON payloads without asking users for internal identifiers.
- **Ignored local runtime data:** `.devctl/`, `tmp/`, and local SQLite files are ignored so logs and databases are not committed.

## Alternatives Considered

- **Start all existing demo sites with `serve-multi`:** Rejected for the initial devctl path because the trail demo currently fails on a stale Kanban DSL API call.
- **Build a separate React frontend:** Rejected for now because the goal is to prove the goja-site hosting path quickly and avoid frontend build dependencies.
- **Use one table per generated form:** Rejected for the first version because JSON response rows keep schema changes simple.

## Implementation Plan

- [x] Create docmgr ticket, diary, design doc, and tasks.
- [x] Add `.devctl.yaml` and `devctl/goja-site-plugin.py`.
- [x] Add `sites/forms/scripts/app.js` with dashboard, builder, public form, response storage, and JSON APIs.
- [x] Validate `devctl plan` and `devctl up --force`.
- [ ] Add edit/delete/reorder controls for fields.
- [ ] Add export endpoints for responses.
- [ ] Add automated browser smoke tests.

## Open Questions

- Should Form Studio eventually be added to `deploy/sites.local.yaml` once the older multi-site demos are compatible again?
- Should response validation reject missing required fields server-side, or is browser-level required validation sufficient for the first iteration?
- Should generated field names be globally stable after label edits when edit support is added?

## References

- Diary: `ttmp/2026/05/20/FORM-GENERATOR-SITE--form-generator-website-and-devctl-startup/reference/01-diary.md`
- Devctl plugin: `devctl/goja-site-plugin.py`
- Form Studio app: `sites/forms/scripts/app.js`
