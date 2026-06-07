---
Title: Diary
Ticket: FORM-GENERATOR-SITE
Status: active
Topics:
    - goja-site
    - javascript
    - web
    - documentation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: .devctl.yaml
      Note: Declares the local devctl profile and goja-site plugin
    - Path: .gitignore
      Note: Keeps devctl logs and local SQLite runtime data out of commits
    - Path: devctl/goja-site-plugin.py
      Note: Implements devctl protocol v2 startup for Form Studio
    - Path: sites/forms/scripts/app.js
      Note: Initial Form Studio form generator website
ExternalSources: []
Summary: Chronological diary for the Form Studio form generator site and devctl startup path.
LastUpdated: 2026-05-20T16:24:00-04:00
WhatFor: Records implementation steps, failures, validation commands, and review instructions for FORM-GENERATOR-SITE.
WhenToUse: Use when resuming work on the form generator website or debugging the devctl startup workflow.
---


# Diary

## Goal

Track the work needed to make this repository easy to start with devctl and to build the first Form Studio form generator website on top of goja-site.

## Step 1: Ticket setup, devctl startup, and initial Form Studio app

This step created the FORM-GENERATOR-SITE docmgr ticket, added a repo-local devctl plugin, and built a first usable JavaScript form generator site. The intended local path is now `devctl up`, which supervises `goja-site serve` on `127.0.0.1:60131` and serves scripts from `sites/forms/scripts` with SQLite state in `tmp/form-generator.db`.

The first devctl attempt targeted the existing multi-site demo config, but startup failed because an older demo script expected a missing `kanban.dsl` feature (`preciseMove`). I switched the devctl plan to a single Form Studio site so the startup path validates the new work directly instead of depending on unrelated demo-site compatibility.

### Prompt Context

**User prompt (verbatim):** "Create a docmgr to get this up and running and build a form generator website. 

First, Create a devctl way of starting this hosting site. 

Keep a diary, commit at appropriate intervals"

**Assistant interpretation:** Create ticket documentation, add devctl-based local startup, implement the first form generator website, keep a detailed diary, and commit focused milestones.

**Inferred user intent:** Make the goja-site hosting repository operational through a repeatable local workflow, then use it to host a useful form-builder application.

**Commit (code):** 93e367cd2922293ea90cb9ca0caba6c1b7cea5c0 — "Add devctl form generator site"

### What I did

- Created docmgr ticket `FORM-GENERATOR-SITE` with a design doc, diary, and three implementation tasks.
- Read the installed `devctl` user, scripting, and plugin-authoring help before implementing the plugin.
- Added `.devctl.yaml` with a `local` profile and a repo-local plugin at `devctl/goja-site-plugin.py`.
- Implemented plugin ops: `config.mutate`, `validate.run`, and `launch.plan` using protocol v2 NDJSON on stdout and logs on stderr.
- Added `.gitignore` entries for `.devctl/`, `tmp/`, and local SQLite files.
- Added `sites/forms/scripts/app.js`, a Form Studio app that can create forms, add fields, render public forms, store responses as JSON, and expose JSON APIs.
- Validated the plugin and app with:
  - `python3 -m py_compile devctl/goja-site-plugin.py`
  - `devctl plugins list`
  - `devctl plan`
  - `devctl up --force`
  - `curl -fsS http://127.0.0.1:60131/ | grep -q 'Form Studio'`
  - `curl -fsS http://127.0.0.1:60131/api/forms | grep -q 'Website intake'`
  - `curl -fsS -X POST http://127.0.0.1:60131/forms -d 'title=Bug+Report&description=Capture+bugs' -o /tmp/form-create.html -w '%{http_code} %{redirect_url}\n'`
  - `devctl status --tail-lines 5`
  - `devctl down`

### Why

- `devctl` gives the repo a single copy/pasteable local startup command and makes logs/status/down behavior consistent.
- A dedicated form-generator script isolates the new website from older demo sites and provides a clear target for future app work.
- Keeping local runtime data under ignored paths prevents generated logs and SQLite files from polluting commits.

### What worked

- The plugin handshake, plan, and launch plan validated successfully after switching to the single Form Studio service.
- `devctl up --force` reached a healthy service on `http://127.0.0.1:60131/`.
- The home page and forms API returned expected Form Studio content.
- Creating a new form via `POST /forms` returned `302 http://127.0.0.1:60131/forms/2`.

### What didn't work

- First attempt: `devctl up --force` against `goja-site serve-multi --config deploy/sites.local.yaml` failed with a health timeout.
- The service stderr showed the actual application error:
  - `Error: create site trail (trail.kanban.yolo.scapegoat.dev): execute script /home/manuel/code/wesen/2026-05-03--goja-hosting-site/sites/trail/scripts/app.js: TypeError: Object has no member 'preciseMove' at /home/manuel/code/wesen/2026-05-03--goja-hosting-site/sites/trail/scripts/app.js:228:17(8)`
- There were also duplicate-column migration log lines from the old trail demo database, but those were not the fatal error.

### What I learned

- The existing multi-site local config currently depends on at least one stale demo script, so it is not the right health target for the new form-generator work.
- `goja-site serve` is sufficient for a first focused devctl path and avoids Host-header/base-domain complications.
- The app can use the existing `database`, `express`, and `ui.dsl` modules to build a useful CRUD-like form builder without Go changes.

### What was tricky to build

- The devctl plugin must keep stdout as strict NDJSON; all diagnostics therefore stay on stderr and the service itself is delegated to devctl supervision.
- The first launch plan technically described the hosting server, but it bound success to unrelated multi-site demo compatibility. The symptom was a health timeout, and the logs revealed a stale JavaScript API call. The fix was to narrow the launch plan to the new form generator script directory and single-site `serve` command.
- The form renderer needs to map schema field types into `ui.dsl` nodes while keeping submitted response payloads predictable. The current implementation stores one JSON object per response keyed by generated field names.

### What warrants a second pair of eyes

- Review `sites/forms/scripts/app.js` for route coverage, field-name collisions, validation strength, and response JSON handling.
- Review `devctl/goja-site-plugin.py` for command quoting; current defaults are safe repo-relative constants, but future user-overridden values may need shell quoting if they include spaces.
- Confirm whether local devctl should eventually run a multi-site config again after stale demo scripts are updated.

### What should be done in the future

- Add edit/delete/reorder controls for form fields.
- Add CSV/JSON response export endpoints.
- Add richer validation rules and required-field enforcement on submit.
- Consider a Playwright smoke test for creating a form, adding a field, and submitting a response.

### Code review instructions

- Start with `.devctl.yaml` and `devctl/goja-site-plugin.py` to verify the startup workflow.
- Then review `sites/forms/scripts/app.js` for the form schema, route handlers, and UI rendering functions.
- Validate with `python3 -m py_compile devctl/goja-site-plugin.py && devctl plan && devctl up --force`, then curl `/` and `/api/forms`, and finally run `devctl down`.

### Technical details

- Local server: `http://127.0.0.1:60131/`
- Metrics endpoint: `http://127.0.0.1:60132/metrics`
- Script directory: `sites/forms/scripts`
- SQLite database: `tmp/form-generator.db`
- Main APIs:
  - `GET /api/forms`
  - `GET /api/forms/:id`
  - `GET /f/:slug`
  - `POST /f/:slug/responses`
