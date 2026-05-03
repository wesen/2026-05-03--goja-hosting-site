# Changelog

## 2026-05-03

- Initial workspace created


## 2026-05-03

Created evidence-backed intern implementation guide for the Goja JavaScript website hosting server and UI DSL.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/GOJA-HOSTING-SITE--goja-javascript-website-hosting-server/design-doc/01-goja-javascript-website-hosting-server-design-and-implementation-guide.md — Primary design guide created for the ticket


## 2026-05-03

Recorded investigation diary with commands, evidence, and follow-up risks.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/GOJA-HOSTING-SITE--goja-javascript-website-hosting-server/reference/01-investigation-diary.md — Diary created for continuation and review


## 2026-05-03

Validated ticket hygiene with docmgr doctor after adding required topic vocabulary.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/vocabulary.yaml — Added ticket topic vocabulary for doctor validation


## 2026-05-03

Uploaded GOJA-HOSTING-SITE design bundle to reMarkable at /ai/2026/05/03/GOJA-HOSTING-SITE.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/GOJA-HOSTING-SITE--goja-javascript-website-hosting-server/design-doc/01-goja-javascript-website-hosting-server-design-and-implementation-guide.md — Included in uploaded reMarkable bundle


## 2026-05-03

Implemented Goja hosting server, UI DSL, Express-style module, Kanban example, and Playwright browser validation (commit 41cbc8e plus follow-up favicon fix).

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/examples/kanban/scripts/app.js — Kanban JavaScript website and Playwright-tested UI


## 2026-05-03

Updated diary with implementation, failures, fixes, Go test, curl, and Playwright validation evidence.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/GOJA-HOSTING-SITE--goja-javascript-website-hosting-server/reference/01-investigation-diary.md — Implementation diary updated after browser validation


## 2026-05-03

Restyled Kanban to Field Notes visual design, added static asset serving, and extended UI DSL tag coverage.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/examples/kanban/scripts/app.js — Field Notes Kanban redesign


## 2026-05-03

Added client interactivity and UI DSL redesign guide covering search, precise card movement, JSON APIs, and browser-side behavior helpers.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/GOJA-HOSTING-SITE--goja-javascript-website-hosting-server/design-doc/02-client-interactivity-and-ui-dsl-redesign-guide.md — New design guide for client-side interactions and UI DSL redesign


## 2026-05-03

Updated tasks and diary with Phase 7 client interactivity redesign plan.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/GOJA-HOSTING-SITE--goja-javascript-website-hosting-server/tasks.md — Added Phase 7 task plan


## 2026-05-03

Uploaded client interactivity guide bundle to reMarkable at /ai/2026/05/03/GOJA-HOSTING-SITE.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/GOJA-HOSTING-SITE--goja-javascript-website-hosting-server/design-doc/02-client-interactivity-and-ui-dsl-redesign-guide.md — Included in reMarkable client interactivity guide bundle


## 2026-05-03

Implemented Phase 7 Kanban search, precise movement, JSON APIs, and app-specific browser JavaScript.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/examples/kanban/scripts/app.js — Phase 7 implementation surface


## 2026-05-03

Added Go-side opaque cookie sessions, exposed req.session to JavaScript, propagated sessions through kanban.dsl, and scoped the Kanban example by session_id.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/examples/kanban/scripts/app.js — Cards are scoped by session_id
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/kanbanddsl/mount.go — Mounted Kanban routes pass session into ctx and event
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/web/request_response.go — RequestDTO now exposes req.session
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/web/session.go — New cookie-backed session manager
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/GOJA-HOSTING-SITE--goja-javascript-website-hosting-server/reference/01-investigation-diary.md — Recorded session implementation and validation

