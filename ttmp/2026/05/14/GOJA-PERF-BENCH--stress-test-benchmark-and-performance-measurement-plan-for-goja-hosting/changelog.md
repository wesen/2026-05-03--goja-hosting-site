# Changelog

## 2026-05-14

- Initial workspace created


## 2026-05-14

Created intern-oriented goja-site stress test and benchmark design guide plus investigation diary.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/design-doc/01-goja-hosting-stress-test-benchmark-and-performance-guide.md — Primary deliverable


## 2026-05-14

Captured current repository validation with go test ./... passing.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/reference/01-investigation-diary.md — Validation evidence and chronological notes


## 2026-05-14

Validated docmgr metadata, added performance vocabulary topics, and uploaded the bundle to reMarkable.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/reference/01-investigation-diary.md — Upload and validation evidence


## 2026-05-14

Added dedicated production observability metrics and tracing implementation guide.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/design-doc/02-goja-site-production-observability-metrics-and-tracing-guide.md — Second design deliverable


## 2026-05-14

Recorded validation and reMarkable upload evidence for the observability guide update.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/reference/01-investigation-diary.md — Step 6 validation and upload evidence


## 2026-05-14

Expanded tasks into phased implementation backlog for observability, load generation, metrics, tracing, and dashboards.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/tasks.md — Phased task backlog


## 2026-05-14

Implemented Phase 1 observability spine with Prometheus diagnostics listener, HTTP metrics, multi-site metrics, and tests.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/cmd/goja-site/serve.go — Single-site diagnostics flags
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/app/server.go — HTTP metrics wrapper integration
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/observability/registry.go — Prometheus registry and collectors for Phase 1


## 2026-05-14

Committed Phase 1 observability spine as 6657f2504ac07194c20c02b6fd934829513f4cc8.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/observability/registry.go — Phase 1 committed implementation
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/reference/01-investigation-diary.md — Diary Step 8 updated with code commit hash


## 2026-05-14

Implemented Phase 2 load-generation MVP around Vegeta with benchmark fixtures, targets, metrics scraping, and report output.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/README.md — Load tool documentation
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/scripts/bench-vegeta.sh — Vegeta load harness


## 2026-05-14

Committed Phase 2 Vegeta load harness as 48743a178f37978ffa3dde04841241207b3ea3ae; live smoke pending Vegeta installation.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/scripts/bench-vegeta.sh — Phase 2 committed load harness
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/reference/01-investigation-diary.md — Diary Step 9 updated with code commit hash


## 2026-05-14

Ran Vegeta null-route smoke test, fixed readiness race on port collision, and confirmed metrics snapshots are captured.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/scripts/bench-vegeta.sh — Readiness check fix after Vegeta smoke
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/reference/01-investigation-diary.md — Step 10 smoke evidence


## 2026-05-14

Implemented Phase 3 database and db.guard metrics with bounded SQL labels and observer-based guard instrumentation.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/app/database.go — DB metric wrapper integration
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/dbguard/guard.go — Guard observer hooks
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/observability/sql.go — DB QueryExecer instrumentation


## 2026-05-14

Committed Phase 3 database and db.guard metrics as 2e55df47cb1a7ca811e658ed0eeb226ecee23a82.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/observability/sql.go — Phase 3 committed DB metrics
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/reference/01-investigation-diary.md — Diary Step 11 updated with code commit hash


## 2026-05-14

Implemented Phase 4 Kanban metrics for mounted fragments, actions, dispatch, refresh render, HTML bytes, and bounded errors.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/app/observability_test.go — Kanban metrics integration test
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/kanbanddsl/mount.go — Kanban route instrumentation
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/observability/kanban.go — Prometheus Kanban metrics


## 2026-05-14

Committed Phase 4 Kanban metrics as a08e5848944e1c4e9eed7a78f712f865662ed679.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/observability/kanban.go — Phase 4 committed Kanban metrics
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/reference/01-investigation-diary.md — Diary Step 12 updated with code commit hash

