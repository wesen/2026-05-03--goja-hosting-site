# Changelog

## 2026-05-03

- Initial workspace created


## 2026-05-03

Created session implementation and SQLite size guard design documents with source file relationships.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/SESSION-DB-MAINT--session-handling-and-sqlite-size-maintenance/design-doc/01-session-cookie-implementation-guide.md — Session implementation guide
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/SESSION-DB-MAINT--session-handling-and-sqlite-size-maintenance/design-doc/02-sqlite-size-guard-and-cleanup-callback-design.md — SQLite size guard and cleanup callback design


## 2026-05-03

Uploaded SESSION-DB-MAINT ticket bundle to reMarkable at /ai/2026/05/03/SESSION-DB-MAINT.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/SESSION-DB-MAINT--session-handling-and-sqlite-size-maintenance/tasks.md — Marked reMarkable upload complete


## 2026-05-03

Prepared SESSION-DB-MAINT ticket documentation for commit after doctor validation and reMarkable upload.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/SESSION-DB-MAINT--session-handling-and-sqlite-size-maintenance/tasks.md — Marked documentation commit task complete


## 2026-05-03

Implemented pkg/dbguard with metered database wrapper, db.guard module, SQLite stats, soft-limit cleanup callbacks, tests, and server wiring.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/app/server.go — Wraps SQLite with MeteredDB and registers db.guard
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/dbguard — New DB size guard package and db.guard module
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/SESSION-DB-MAINT--session-handling-and-sqlite-size-maintenance/reference/01-investigation-diary.md — Recorded db.guard implementation and validation


## 2026-05-03

Uploaded updated SESSION-DB-MAINT implementation bundle to reMarkable after db.guard implementation.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/SESSION-DB-MAINT--session-handling-and-sqlite-size-maintenance/changelog.md — Recorded implementation bundle upload


## 2026-05-03

Added hard-limit enforcement policy and implementation tasks to the SQLite size guard design.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/SESSION-DB-MAINT--session-handling-and-sqlite-size-maintenance/design-doc/02-sqlite-size-guard-and-cleanup-callback-design.md — Hard-limit enforcement addendum
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/SESSION-DB-MAINT--session-handling-and-sqlite-size-maintenance/tasks.md — Hard-limit implementation tasks


## 2026-05-03

Implemented opt-in hard-limit write enforcement for db.guard with SQL classification and tests.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/dbguard/guard.go — Adds hard-limit preflight and post-exec enforcement
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/dbguard/hardlimit_test.go — Hard-limit enforcement tests
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/dbguard/metered.go — Rejects growth writes when hard limit policy requires it
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/dbguard/sqlkind.go — Classifies SQL statements for hard-limit policy

