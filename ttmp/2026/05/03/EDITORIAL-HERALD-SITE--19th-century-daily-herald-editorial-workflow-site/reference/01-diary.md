---
Title: Diary
Ticket: EDITORIAL-HERALD-SITE
Status: active
Topics:
    - goja-site
    - javascript
    - kanban
    - sqlite
    - documentation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: sites/editorial/scripts/00_domain.js
      Note: Daily Herald domain constants and helpers (commit bbef48b)
    - Path: sites/editorial/scripts/01_styles.js
      Note: |-
        Broadsheet visual style for the reference-inspired UI (commit bbef48b)
        Dossier busy-state styling for fragment requests (commit 8c6630a)
        Styles for active filters and story table (commit e820e72)
        Removed top search styling and polished Kanban search/full story styles (commit 7264fdb)
    - Path: sites/editorial/scripts/02_repository.js
      Note: |-
        SQLite schema and persistence layer for staff
        Persistence helpers for adding checklist items and updating tags (commit 7264fdb)
    - Path: sites/editorial/scripts/03_workflow.js
      Note: |-
        Seed data
        Tag and desk filtering plus filtered active-story selection (commit e820e72)
        Workflow helpers for tag and checklist mutations (commit 7264fdb)
    - Path: sites/editorial/scripts/04_views.js
      Note: |-
        Server-rendered shell
        Story link and checklist form data attributes for progressive enhancement (commit 8c6630a)
        Functional tag links
        Working tag/checklist forms and full story HTML page (commit 7264fdb)
    - Path: sites/editorial/scripts/05_routes.js
      Note: |-
        HTTP routes
        Panel fragment route and enhanced checklist response path (commit 8c6630a)
        Filtered stories and assignments APIs (commit e820e72)
        Tag/checklist POST routes and HTML story route (commit 7264fdb)
    - Path: sites/editorial/scripts/06_client.js
      Note: |-
        Progressive enhancement script for no-flash dossier fragment replacement (commit 8c6630a)
        URL-encoded progressive form submissions for panel updates (commit 7264fdb)
    - Path: ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/design-doc/01-daily-herald-editorial-workflow-architecture.md
      Note: Architecture plan derived from reference images
    - Path: ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/sources/images/01-full-editorial-workflow-reference.png
      Note: Primary visual reference imported with cp
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---





# Diary

## Goal

Capture the step-by-step implementation history for replacing the previous Editorial example with a richer 19th-century Daily Herald newsroom application, including reference images, architectural choices, failures, validation commands, and review guidance.

## Step 1: Import references and build the first Daily Herald application

The first milestone created the `EDITORIAL-HERALD-SITE` ticket, copied the user-provided reference images into the ticket workspace, and replaced the single-file Editorial example with a split-script Daily Herald application. The result is a server-rendered newsroom shell with a left desk sidebar, a main `kanban.dsl` workflow board, a right story dossier panel, SQLite-backed staff/story/checklist data, JSON APIs, and local validation on port `60133`.

This milestone intentionally keeps the board as only one component inside a broader application. The interesting behavior now comes from the combined domain model: stories have authors, desks, priorities, due dates, checklist progress, and detailed dossier content in addition to workflow status.

### Prompt Context

**User prompt (verbatim):** "Import the images into the ticket with `cp` for referencing later."

**Assistant interpretation:** Create a dedicated docmgr ticket for the editorial redesign, copy the supplied clipboard images into a stable ticket path, and use them as implementation references.

**Inferred user intent:** Preserve visual requirements and source material in versioned project documentation before implementing a redesign based on those images.

**Commit (code):** `bbef48b8027842f2b503a883b213d921204f9d1f` — "Redesign editorial site as Daily Herald workflow"

### What I did

- Created ticket `EDITORIAL-HERALD-SITE` with `docmgr ticket create-ticket`.
- Copied the three reference PNGs with `cp` into:
  - `ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/sources/images/01-full-editorial-workflow-reference.png`
  - `ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/sources/images/02-story-detail-panel-reference.png`
  - `ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/sources/images/03-board-only-reference.png`
- Ran `understand_image` on the full workflow reference to extract concrete UI and data requirements.
- Added design document:
  - `ttmp/2026/05/03/EDITORIAL-HERALD-SITE--19th-century-daily-herald-editorial-workflow-site/design-doc/01-daily-herald-editorial-workflow-architecture.md`
- Removed the previous single file:
  - `sites/editorial/scripts/app.js`
- Added split scripts:
  - `sites/editorial/scripts/00_domain.js`
  - `sites/editorial/scripts/01_styles.js`
  - `sites/editorial/scripts/02_repository.js`
  - `sites/editorial/scripts/03_workflow.js`
  - `sites/editorial/scripts/04_views.js`
  - `sites/editorial/scripts/05_routes.js`
- Implemented SQLite tables `herald_staff`, `herald_stories`, and `herald_checklist`.
- Implemented seed data for a 19th-century newspaper workflow, including pitches, reporting, writing, editing, and ready-for-print stories.
- Implemented `kanban.dsl` board mounting and movement callback through `Herald.workflow.moveStory(event)`.
- Added JSON endpoints:
  - `GET /api/stories`
  - `GET /api/stories/:id`
  - `GET /api/assignments`
  - `GET /api/desks`
  - `GET /api/metrics`
- Served the site locally on `127.0.0.1:60133`.
- Committed the first milestone as `bbef48b8027842f2b503a883b213d921204f9d1f`.

### Why

- The user wanted the reference images imported into the ticket so the visual brief would survive beyond `/tmp` clipboard paths.
- The previous Editorial example did not demonstrate enough application depth; it needed to become a richer newsroom workflow rather than a visually restyled board.
- Split scripts make this larger JavaScript example reviewable and easier for future authors to copy.

### What worked

- `docmgr ticket create-ticket --ticket EDITORIAL-HERALD-SITE --title "19th century Daily Herald editorial workflow site" --topics goja-site,javascript,kanban,sqlite,documentation` created the expected workspace.
- `cp` imported all three reference images into `sources/images` successfully.
- `understand_image` identified the main shell pieces: sidebar, topbar, board, right dossier, desk filtering concepts, author metadata, tags, and checklist.
- `node -c sites/editorial/scripts/*.js` passed after formatting.
- `go test ./... -count=1` passed:
  - `pkg/app`
  - `pkg/dbguard`
  - `pkg/kanbanddsl`
  - `pkg/uidsl`
  - `pkg/web`
- Local API validation worked:
  - `curl -fsS http://127.0.0.1:60133/api/metrics | jq .`
  - returned `active: 8`, `ready: 2`, `highPriority: 4`, `checklistPercent: 23`.
- Main page validation worked after fixing the unsupported tag helper:
  - `curl -fsS http://127.0.0.1:60133/ | grep -Eo 'The Daily Herald|Editorial Workflow|City Council debates|Checklist' | head -10`

### What didn't work

- The first page render failed with this exact runtime error:

```text
JavaScript handler error: TypeError: Object has no member 'aside' at sidebar (sites/editorial/scripts/04_views.js:23:20(116))
```

- Root cause: `ui.dsl` does not currently expose an `aside()` helper, even though HTML has an `aside` element.
- Fix: replaced `ui.aside(...)` with `ui.section(...)` while preserving the CSS classes `sidebar` and `dossier`.
- The page now renders, but clicking story cards still uses full-page navigation via `href="/?story=<id>"`, causing a visible flash. That is the next behavior to improve with progressive enhancement.

### What I learned

- The `ui.dsl` element surface is intentionally narrower than all HTML tags. New examples should either use known helpers or the DSL should grow an explicit generic-tag escape hatch.
- The current app architecture can support a rich shell without changing Go code: all domain state, views, routes, and board callbacks live in site scripts.
- The right story dossier is an ideal place to demonstrate progressive enhancement because it can remain server-rendered and be swapped as a fragment.

### What was tricky to build

- The main sharp edge was shaping `kanban.dsl` card links so the card remains draggable but also navigates to a selected story for the dossier. In v1 this is a normal link inside the rendered card, which is simple and accessible but triggers a full page reload.
- The second sharp edge was SQLite session isolation. All seed, query, movement, and checklist functions use `Herald.util.sessionId(session)` so each cookie session gets its own staff/story/checklist rows.
- The unsupported `ui.aside` helper was discovered only at runtime because JavaScript syntax validation cannot detect missing members on the DSL object.

### What warrants a second pair of eyes

- Review the generated board markup around linked cards and drag/drop behavior. A card that contains a link may produce subtle click-vs-drag interactions in browsers.
- Review `Herald.workflow.seedIfEmpty` and `Herald.repo.countStories`; it seeds per session, which is correct for examples but can surprise reviewers expecting shared production data.
- Review `Herald.workflow.moveStory`; it normalizes destination and source status positions, but it does not yet enforce checklist completion before `ready`.

### What should be done in the future

- Add progressive enhancement to avoid full-page flashes when selecting stories or toggling checklist items.
- Consider adding an `ui.element(tagName, attrs, children...)` helper or additional semantic helpers such as `ui.aside` if this comes up repeatedly.
- Add a future optional movement rule that warns or blocks movement to `ready` when checklist items remain incomplete.

### Code review instructions

- Start with `sites/editorial/scripts/00_domain.js` for constants and domain helpers.
- Then read `sites/editorial/scripts/02_repository.js` for schema and SQL access.
- Then read `sites/editorial/scripts/03_workflow.js` for seeding, movement, metrics, and checklist behavior.
- Then read `sites/editorial/scripts/04_views.js` for the shell, board, cards, and dossier.
- Finally read `sites/editorial/scripts/05_routes.js` for HTTP routes and API shape.
- Validate with:

```bash
node -c sites/editorial/scripts/*.js
go test ./... -count=1
curl -fsS http://127.0.0.1:60133/ | grep 'The Daily Herald'
curl -fsS http://127.0.0.1:60133/api/desks | jq
curl -fsS http://127.0.0.1:60133/api/metrics | jq
```

### Technical details

Local server command used for validation:

```bash
GOTOOLCHAIN=go1.26.2 go run ./cmd/goja-site serve \
  --db ./tmp/editorial-herald/app.db \
  --scripts ./sites/editorial/scripts \
  --addr 127.0.0.1:60133 \
  --dev
```

Current local URL:

```text
http://127.0.0.1:60133/
```

Known next UI issue:

```text
http://localhost:60133/?story=33 reloads the entire page when story links are clicked.
```

## Step 2: Add progressive dossier updates to avoid full-page flashes

The second milestone added progressive enhancement for story selection and checklist toggles. The app still renders ordinary links and forms for accessibility and fallback, but browsers with JavaScript now fetch a server-rendered dossier fragment, replace only `.dossier`, and update the URL with `history.pushState`.

This preserves the server-rendered architecture while removing the visual flash caused by full page reloads. It also gives the Daily Herald app a practical pattern for rich interactions without introducing a large browser framework.

### Prompt Context

**User prompt (verbatim):** "how can we avoid a flash when clicking on things on http://localhost:60133/?story=33, since it seems to reload the whole page?"

**Assistant interpretation:** Explain and then implement a way to avoid full-page reloads for story-card and dossier interactions.

**Inferred user intent:** Keep the application feeling responsive and polished while preserving the server-rendered Goja model.

**Commit (code):** `8c6630a8d2dd405c8be896f99a9031e06273007d` — "Add progressive dossier updates to Editorial Herald"

### What I did

- Added `sites/editorial/scripts/06_client.js` defining `Herald.clientScript`.
- Served the browser enhancement at `GET /assets/herald.js`.
- Added `GET /stories/:id/panel`, which returns only `Herald.views.dossier(story)` as HTML.
- Added a tiny request helper in routes:
  - `wantsPanel(req)` checks `X-Herald-Panel: 1`.
  - `renderPanel(req, res, storyId)` renders the selected dossier fragment.
- Updated story-card links in `Herald.views.storyCard(story)` with:
  - `data-herald-story-link="true"`
  - `data-herald-panel-url="/stories/<id>/panel"`
- Updated checklist forms in `Herald.views.checkItem(story, item)` with:
  - `data-herald-panel-form="true"`
- Added the browser script to the page with:
  - `ui.script({ src: "/assets/herald.js", defer: true })`
- Added an opacity transition for `.dossier[aria-busy="true"]` while fragment requests are in flight.
- Validated with Playwright that selecting a story and toggling checklist items do not trigger `beforeunload`.

### Why

- Full-page navigation caused a visible flash when selecting stories like `/?story=33`.
- The right dossier is a self-contained server-rendered fragment, so replacing only that region is a low-risk progressive enhancement.
- Keeping normal `href` and `form` behavior means the app still works if JavaScript fails.

### What worked

- `GET /assets/herald.js` returns the enhancement script after restarting the local server.
- `GET /stories/1/panel` returns standalone dossier HTML.
- `POST /stories/:id/checklist/:itemId/toggle` returns panel HTML when called with `X-Herald-Panel: 1`.
- Playwright story-selection validation returned:

```json
{
  "url": "http://127.0.0.1:60133/?story=36",
  "beforeUnload": 0,
  "dossierTitle": "City Council debates new housing plan"
}
```

- Playwright checklist-toggle validation returned:

```json
{
  "url": "http://127.0.0.1:60133/?story=36",
  "beforeUnload": 0,
  "before": 1,
  "after": 0,
  "title": "City Council debates new housing plan"
}
```

- Validation commands passed:

```bash
node -c sites/editorial/scripts/*.js
go test ./... -count=1
```

### What didn't work

- Before restarting the local server, the new routes returned 404:

```text
curl: (22) The requested URL returned error: 404
```

- Root cause: the running server had been started before `06_client.js` and the new routes existed, so the process needed to be restarted to load the new script file.
- Fix: stopped the process on port `60133`, removed `tmp/editorial-herald` while the process was stopped, and restarted the local server.

### What I learned

- `--dev` is useful, but adding an entirely new script file still warrants a clean restart when validating route availability.
- The route-level fragment approach is enough to remove the flash without modifying `kanban.dsl`.
- Checklist state can be updated in the panel without page reload; the board card progress is not yet refreshed, which is acceptable for the first enhancement but worth improving.

### What was tricky to build

- The enhancement needed to avoid interfering with normal browser affordances. The click handler skips modified clicks (`meta`, `ctrl`, `shift`, `alt`) so users can still open links in new tabs.
- The panel replacement uses a `<template>` element and replaces only the first element in the returned fragment. This expects `Herald.views.dossier(story)` to render exactly one root element with class `dossier`.
- The checklist fallback path still redirects, while enhanced requests return HTML. That split is controlled by `X-Herald-Panel: 1` and should be kept consistent if more panel forms are added.

### What warrants a second pair of eyes

- Review whether checklist toggles should also refresh the Kanban board fragment so card progress updates immediately.
- Review whether `popstate` should support returning to the no-story URL by clearing or resetting the dossier; currently it only fetches a panel if `story` is present.
- Review whether this pattern should be generalized into `kanban.dsl` as a first-class detail-panel feature.

### What should be done in the future

- Add optional board-fragment refresh after checklist toggles if stale card progress feels confusing.
- Consider moving this progressive panel pattern into a reusable helper for example sites.
- Consider making `kanban.dsl` support card-selected panels declaratively.

### Code review instructions

- Start with `sites/editorial/scripts/06_client.js` to understand the progressive enhancement contract.
- Then read the new helper routes in `sites/editorial/scripts/05_routes.js`.
- Then review the data attributes added in `sites/editorial/scripts/04_views.js`.
- Validate no-reload behavior manually by opening `http://127.0.0.1:60133/?story=33`, clicking another story, and watching that only the right dossier changes.
- Automated/browser validation can use Playwright-style checks for `beforeunload === 0` after story clicks and checklist toggles.

### Technical details

Server route contract:

```text
GET  /assets/herald.js
GET  /stories/:id/panel
POST /stories/:id/checklist/:itemId/toggle
     Header: X-Herald-Panel: 1 -> returns dossier HTML
     No header                  -> redirects to /?story=:id
```

Browser behavior:

```text
click [data-herald-story-link]
→ prevent full navigation
→ fetch /stories/:id/panel
→ replace .dossier
→ history.pushState({}, "", /?story=:id)

submit [data-herald-panel-form]
→ prevent full form submission
→ POST with FormData and X-Herald-Panel: 1
→ replace .dossier
```

## Step 3: Make tags and table view functional

The third milestone turned the visual placeholders for tags and the Board/Table tabs into working application behavior. Tag chips now link to filtered board/table views, desk sidebar entries filter stories by desk, and the Table tab renders the same filtered story set as a real table rather than acting as a decorative tab.

This milestone also removed the redundant top-right Filter, Sort, and New Story toolbar from the workflow header. The board already provides its own search affordance, and the story form remains directly below the header where it is easier to understand as a server-backed action.

### Prompt Context

**User prompt (verbatim):** "then, make tags work, table view work, remove filter / sort / new story since they are already in the kanban."

**Assistant interpretation:** Wire tag chips and the table tab to real filtered server-rendered behavior, and remove redundant header controls.

**Inferred user intent:** Make the UI honest: visible controls should do something, and redundant placeholder controls should disappear.

**Commit (code):** `e820e72` — "Make Editorial Herald tags and table view functional"

### What I did

- Updated `Herald.workflow.listStories(session, query)` to accept query filters.
- Added `Herald.workflow.matchesQuery(story, query)` with support for:
  - `?tag=<tag>` exact tag matching against comma-separated story tags;
  - `?desk=<desk>` exact desk matching against story desk IDs.
- Updated `Herald.workflow.activeStory(session, query)` so filtered pages select the first matching story for the dossier when no explicit `story` query parameter is present.
- Updated `Herald.board` data loading to pass `ctx.query` into `listStories`.
- Made sidebar desk entries link to `/?desk=<desk>`.
- Made dossier tag chips link to `/?tag=<tag>`.
- Added an active filter chip with a clear link.
- Replaced decorative Board/Table spans with real links that preserve existing filters.
- Added `Herald.views.tableView(req)` to render a proper story table with headline, desk, author, status, due date, checklist progress, and tags.
- Updated `GET /api/stories` and `GET /api/assignments` to honor query filters.
- Removed the top-right workflow header controls:
  - `Filter⌄`
  - `Sort: Priority⌄`
  - `+ New Story`

### Why

- Tag chips and the Table tab looked interactive but did not previously change the application state.
- The redundant top-right controls made the UI feel like a mockup rather than a working app.
- Query-driven filters keep the site server-rendered, bookmarkable, and easy to inspect with `curl`.

### What worked

- Local filtered board validation:

```bash
curl -fsS 'http://127.0.0.1:60133/?tag=Society' \
  | grep -Eo 'Tag: Society|The rise of women|Summer fashion|City Council debates' \
  | sort | uniq -c
```

After restarting the server, this showed the Society filter and Society stories without the unrelated City Council dossier content.

- Local table validation:

```bash
curl -fsS 'http://127.0.0.1:60133/?view=table&tag=Society' \
  | grep -Eo 'The rise of women|Summer fashion|City Council debates' \
  | sort | uniq -c
```

This showed the filtered table content.

- API filter validation:

```bash
curl -fsS 'http://127.0.0.1:60133/api/stories?tag=Society' | jq '[.[].title]'
```

Returned:

```json
[
  "The rise of women in medicine",
  "Summer fashion arrives in the city"
]
```

- Browser validation with Playwright on `/?view=table&tag=Society` returned:

```json
{
  "tableRows": 2,
  "hasTopFilter": 0,
  "hasTopSort": 0,
  "hasNewStoryButton": 0,
  "activeFilter": "Tag: Society×",
  "firstDossier": "The rise of women in medicine"
}
```

- Validation commands passed:

```bash
node -c sites/editorial/scripts/*.js
go test ./... -count=1
```

### What didn't work

- The first filtered-page `curl` check still showed `City Council debates` because the running local process had not reloaded the edited `activeStory` behavior.
- Fix: restarted the local server on port `60133` after stopping the old process, then repeated the checks.

### What I learned

- Filtered pages need both the main view and the dossier to honor the same query semantics, otherwise the page feels internally inconsistent.
- The table view is easiest to keep correct when it reads from the same `Herald.workflow.listStories(session, query)` method as the board.
- Query-string filters are a good fit for this app because they work for links, tables, APIs, and future deployment smoke tests.

### What was tricky to build

- Preserving query parameters across Board/Table tabs required a small `urlWithQuery(query, changes)` helper. Without that, changing views would clear the active tag or desk filter.
- `activeStory` could not simply use `repo.firstStory` anymore. When a tag or desk filter is active, the right dossier should show the first matching story unless the URL explicitly selects `?story=<id>`.
- Tag matching is exact and case-insensitive against split tags. This is clearer than substring matching, but it means tag spelling in seed data matters.

### What warrants a second pair of eyes

- Review whether active explicit `?story=<id>` should be forced to match the active filter or whether explicit story selection should continue to win.
- Review whether the story creation form should also inherit the current tag/desk filter or remain global.
- Review whether API filtering should support combined `tag` + `desk` filters in production examples; the implementation already combines them with AND semantics.

### What should be done in the future

- Add visual empty states for filters that return zero rows.
- Consider making tag and desk filters progressive/no-flash as a future enhancement if full-page filter changes feel too heavy.
- Consider adding tag counts to the sidebar or dossier.

### Code review instructions

- Start in `sites/editorial/scripts/03_workflow.js` with `listStories`, `matchesQuery`, and `activeStory`.
- Then review `sites/editorial/scripts/04_views.js` for `tabs`, `activeFilter`, `urlWithQuery`, `tableView`, and tag links.
- Check `sites/editorial/scripts/05_routes.js` to confirm filtered APIs use `req.query`.
- Validate manually:

```bash
curl -fsS 'http://127.0.0.1:60133/?tag=Society' | grep 'Tag: Society'
curl -fsS 'http://127.0.0.1:60133/?view=table&tag=Society' | grep 'story-table'
curl -fsS 'http://127.0.0.1:60133/api/stories?tag=Society' | jq '[.[].title]'
```

### Technical details

New query contract:

```text
/?tag=Society              -> board filtered to stories with exact tag Society
/?desk=city                -> board filtered to City Hall desk stories
/?view=table               -> table view instead of board
/?view=table&tag=Society   -> filtered table view
```

The filter function combines filters with AND semantics:

```text
tag filter must match one split story tag exactly, case-insensitively
desk filter must match story.desk exactly
```

## Step 4: Polish dossier actions, full story view, and search styling

The fourth milestone turned the remaining decorative dossier controls into working forms. The tag `+` control is now an inline add-tag form, `+ Add item` is now an inline checklist-item form, and both use the existing progressive panel update path so they update the dossier without a full page flash.

This milestone also removed the top search field from the header, restyled the Kanban search field so it matches the newspaper UI, and changed `Open full view` from a JSON endpoint into a proper server-rendered story page.

### Prompt Context

**User prompt (verbatim):** "tags + does not work, +Add item does not work. Open full view only shows the json. the top search can go. The search cards is badly styled."

**Assistant interpretation:** Replace non-functional visual placeholders with working server-backed interactions, make the full view human-readable HTML, remove the global top search, and improve the Kanban search input styling.

**Inferred user intent:** Bring the Editorial Herald UI closer to production quality by removing mock controls and making every visible control behave as expected.

**Commit (code):** `7264fdb` — "Polish Editorial Herald dossier actions"

### What I did

- Added repository helpers in `sites/editorial/scripts/02_repository.js`:
  - `addChecklistItem(sid, storyId, label)`
  - `setStoryTags(sid, storyId, tags)`
- Added workflow helpers in `sites/editorial/scripts/03_workflow.js`:
  - `addChecklistItem(session, storyId, label)`
  - `addTag(session, storyId, tag)`
  - `removeTag(session, storyId, tag)` for future use
- Replaced the decorative `+` tag chip in `sites/editorial/scripts/04_views.js` with `addTagForm(story)`.
- Replaced decorative `+ Add item` with `addChecklistForm(story)`.
- Added `fullStoryPage(story)` to render a newspaper-style full story view.
- Changed `GET /stories/:id` in `sites/editorial/scripts/05_routes.js` to render HTML instead of JSON.
- Kept JSON detail data available at `GET /api/stories/:id`.
- Added `POST /stories/:id/tags`.
- Added `POST /stories/:id/checklist`.
- Updated `sites/editorial/scripts/06_client.js` so progressive panel forms submit as `application/x-www-form-urlencoded` instead of `FormData` multipart. This makes parsed form fields reliable for add-tag and add-checklist actions.
- Removed the topbar search input from `topbar()`.
- Styled the `kanban.dsl` default search toolbar:
  - `.kb-toolbar`
  - `.kb-toolbar input[type=search]`
  - `.kb-toolbar input[type=search]::placeholder`

### Why

- The `+` tag and `+ Add item` controls looked actionable but were static text, which made the app feel unfinished.
- A link labeled `Open full view` should not return raw JSON. JSON belongs under `/api/...`, while `/stories/:id` should be a page.
- The top search duplicated the Kanban card search and did not yet have behavior, so removing it simplified the UI.
- The default Kanban search was visually inconsistent with the broadsheet theme.

### What worked

- `node -c sites/editorial/scripts/*.js` passed.
- `go test ./... -count=1` passed.
- Browser validation confirmed:

```json
{
  "topSearch": 0,
  "boardSearchBox": {
    "width": 420.0000305175781,
    "height": 43.993072509765625
  },
  "computed": {
    "height": "43.9931px",
    "minHeight": "44px",
    "borderRadius": "999px",
    "font": "16px Georgia, \"Times New Roman\", serif"
  },
  "hasScoop": 1,
  "hasItem": 1,
  "fullStatus": 200,
  "fullIsJson": false,
  "fullHasBrief": true
}
```

- The full story page now contains `Assignment brief` and is not JSON.
- The add-tag and add-checklist flows update the dossier through the progressive enhancement path.

### What didn't work

- Before restarting the local server, the Kanban search styling still appeared unstyled in Playwright because the old script/CSS string was still loaded in the running process.
- The first CSS selector used quoted attribute selectors inside a string rendered by `ui.style`; the rendered HTML escaped quotes in CSS as `&#34;`, so the browser did not apply the selector:

```css
.kb-toolbar input[type=&#34;search&#34;]
```

- Fix: changed the CSS selector to avoid quotes:

```css
.kb-toolbar input[type=search]
```

- After restarting the server, Playwright confirmed the search box computed styles were applied.

### What I learned

- CSS strings rendered through `ui.style` are HTML-escaped. Attribute selectors should avoid quoted values when possible, or `ui.style` should eventually treat CSS content differently.
- `FormData` requests can be inconvenient with the current body parser path. Sending `URLSearchParams` with `application/x-www-form-urlencoded` is simpler and reliable for progressive-enhancement forms.
- Keeping `/stories/:id` as HTML and `/api/stories/:id` as JSON makes the route model much clearer.

### What was tricky to build

- The progressive form helper needed to preserve the `X-Herald-Panel: 1` header while also setting `Content-Type`. Because `fetchPanel` merges options shallowly, the form submit path now supplies both headers explicitly.
- Full-story validation with raw `curl /stories/1` can show `Story not found` if it uses a fresh cookie session that seeded stories with different IDs. Browser validation through an actual page link is a better check because the link belongs to the active session.
- Styling the Kanban search required understanding that `ui.style` escaped quotes and that the old running process needed a restart before Playwright could observe the new CSS string.

### What warrants a second pair of eyes

- Review whether the add-tag form should reject punctuation or normalize tag capitalization.
- Review whether the add-checklist form should refresh the Kanban card progress immediately. The dossier updates now; the board card progress may remain stale until the board refreshes.
- Review whether `removeTag` should be surfaced as an explicit UI action or removed until needed.

### What should be done in the future

- Add delete/remove controls for tags and checklist items if the app should support editing mistakes.
- Add a board fragment refresh after checklist mutations if stale progress on cards is noticeable.
- Consider changing `ui.style` rendering so CSS content is not over-escaped.

### Code review instructions

- Start with `sites/editorial/scripts/04_views.js` and review `addTagForm`, `addChecklistForm`, and `fullStoryPage`.
- Review `sites/editorial/scripts/05_routes.js` for the new POST routes and changed `/stories/:id` HTML route.
- Review `sites/editorial/scripts/06_client.js` to confirm enhanced forms submit as URL-encoded data.
- Review `sites/editorial/scripts/01_styles.js` for the top-search removal aftermath and Kanban search styling.
- Validate locally:

```bash
node -c sites/editorial/scripts/*.js
go test ./... -count=1
```

Then test in the browser:

```text
http://localhost:60133/
```

- Add a tag in the dossier and confirm it appears without a full-page reload.
- Add a checklist item and confirm it appears without a full-page reload.
- Click `Open full view ↗` and confirm it renders HTML rather than JSON.
- Confirm the top search is gone and the remaining `Search cards...` input is styled.

### Technical details

New routes:

```text
POST /stories/:id/tags       body: tag=<name>
POST /stories/:id/checklist  body: label=<checklist item>
GET  /stories/:id            HTML full story view
GET  /api/stories/:id        JSON story detail
```

Progressive form submission now sends:

```http
X-Herald-Panel: 1
Content-Type: application/x-www-form-urlencoded;charset=UTF-8
```

## Step 5: Deploy Daily Herald Editorial to production

The fifth milestone deployed the completed Daily Herald editorial redesign to the production `goja-kanban` K3s application. The app repository push published image `ghcr.io/wesen/2026-05-03--goja-hosting-site:sha-a6379c7`, the automated workflow opened the K3s GitOps pull request, and that pull request was merged to let Argo CD reconcile the new image.

Production validation confirmed that the editorial host serves the redesigned Daily Herald UI, the new HTML full-story view, the filtered stories API, and the polished search/add controls. Argo CD returned to `Synced Healthy` with the expected immutable image tag.

### Prompt Context

**User prompt (verbatim):** "alright good enough, deploy."

**Assistant interpretation:** Push the current app commits, let the image/GitOps automation run, merge the generated K3s PR, wait for Argo CD, and validate the public Editorial host.

**Inferred user intent:** Ship the Daily Herald redesign to the public `editorial.kanban.yolo.scapegoat.dev` production site.

**Commit (code):** `a6379c7` — "Diary: record Editorial Herald dossier polish" was the pushed app commit and image source.

### What I did

- Verified local git status before deployment. Only `tmp/` was untracked from the local SQLite server.
- Pushed app repo `main`:

```bash
git push origin main
```

- Monitored GitHub Actions run:
  - Run: `25295444096`
  - URL: `https://github.com/wesen/2026-05-03--goja-hosting-site/actions/runs/25295444096`
  - Result: `completed/success`
  - Jobs:
    - `release / publish`
    - `release / Open GitOps PR`
- Confirmed generated K3s PR:
  - PR: `https://github.com/wesen/2026-03-27--hetzner-k3s/pull/74`
  - Title: `Deploy goja-kanban-prod using ghcr.io/wesen/2026-05-03--goja-hosting-site:sha-a6379c7`
  - Head: `automation/2026-05-03--goja-hosting-site-goja-kanban-prod-sha-a6379c7`
  - Mergeable: `MERGEABLE`
- Merged PR #74:
  - Merge commit: `cecc102ce156bd791421bb16044e600265038d08`
- Waited for Argo CD / Kubernetes rollout:
  - image changed from `sha-7c9289f` to `sha-a6379c7`
  - deployment reached `ready=1/1`
  - Argo CD returned to `Synced Healthy`

### Why

- The user explicitly approved the redesigned Editorial site for deployment.
- The repository already has a validated GHCR publish + GitOps PR workflow, so the deployment should use that path rather than manual cluster edits.
- Argo CD should reconcile an immutable image tag from the K3s GitOps repo.

### What worked

- App workflow succeeded:

```text
25295444096 publish-image completed success a6379c7
```

- K3s PR was opened automatically:

```text
https://github.com/wesen/2026-03-27--hetzner-k3s/pull/74
```

- K3s PR merged successfully:

```text
mergeCommit: cecc102ce156bd791421bb16044e600265038d08
```

- Argo CD rollout completed:

```text
app=Synced Healthy ready=1/1 image=ghcr.io/wesen/2026-05-03--goja-hosting-site:sha-a6379c7
```

- Production smoke tests passed:

```bash
curl -k -fsSI https://editorial.kanban.yolo.scapegoat.dev/ | head -1
# HTTP/2 200
```

The production page contained:

```text
The Daily Herald
Editorial Workflow
Search cards
New tag
New checklist item
Open full view
```

and did not contain the removed top-search placeholder `Search stories`.

- Production full story view returned HTML content:

```text
Back to Editorial Room
Assignment brief
```

- Production filtered API worked:

```bash
curl -k -fsS 'https://editorial.kanban.yolo.scapegoat.dev/api/stories?tag=Society' | jq '[.[].title]'
```

returned:

```json
[
  "The rise of women in medicine",
  "Summer fashion arrives in the city"
]
```

### What didn't work

- No deployment blockers occurred.
- The known non-blocking local state remains: `tmp/` is untracked because the local dev server uses `tmp/editorial-herald/app.db`.

### What I learned

- The automatic publish-image to K3s PR flow remains healthy after the Editorial redesign.
- Argo CD took roughly a minute after PR merge to observe the GitOps change and roll the Deployment.
- Production validation should use a cookie jar for routes like `/stories/:id` because story IDs are session-scoped seed data.

### What was tricky to build

- There is a short delay where Argo CD reports `Synced Healthy` for the old image before observing the merged GitOps change. Polling both the Argo app state and the Deployment image avoids mistaking the old healthy state for the completed deployment.
- Full-story validation needs to fetch a story ID from the same session that rendered the page. Directly requesting `/stories/1` in a fresh session may not find a story because IDs depend on existing DB rows and session-scoped seeding.

### What warrants a second pair of eyes

- Review the production UI manually to ensure the 19th-century layout looks acceptable at real browser widths.
- Review whether the public production database now contains multiple session-seeded story sets from smoke testing. This is expected for the example app but may matter if demo data should stay very small.

### What should be done in the future

- If production smoke tests should avoid growing the SQLite database, add a dedicated read-only demo session or cleanup path.
- Optionally deploy a follow-up that refreshes board card checklist progress after dossier mutations.

### Code review instructions

- Deployment source commit: `a6379c7` in `wesen/2026-05-03--goja-hosting-site`.
- Deployed image: `ghcr.io/wesen/2026-05-03--goja-hosting-site:sha-a6379c7`.
- GitOps PR: `https://github.com/wesen/2026-03-27--hetzner-k3s/pull/74`.
- Validate production:

```bash
curl -k -fsSI https://editorial.kanban.yolo.scapegoat.dev/ | head -1
curl -k -fsS https://editorial.kanban.yolo.scapegoat.dev/ | grep 'The Daily Herald'
curl -k -fsS 'https://editorial.kanban.yolo.scapegoat.dev/api/stories?tag=Society' | jq '[.[].title]'
```

### Technical details

Final rollout evidence:

```text
goja-kanban Argo CD: Synced Healthy
Deployment image: ghcr.io/wesen/2026-05-03--goja-hosting-site:sha-a6379c7
Deployment ready: 1/1
Pod: goja-kanban-55c54f85f9-knnt7 Running
```
