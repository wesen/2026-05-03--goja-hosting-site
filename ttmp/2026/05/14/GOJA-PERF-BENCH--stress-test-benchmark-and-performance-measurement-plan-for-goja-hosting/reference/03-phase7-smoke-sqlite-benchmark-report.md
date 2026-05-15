---
Title: Benchmark Matrix SQLite Report - phase7-smoke-20260515T123252Z
Ticket: GOJA-PERF-BENCH
Status: active
Topics:
    - benchmarking
    - sqlite
    - vegeta
    - prometheus
DocType: reference
Intent: historical
Owners: []
RelatedFiles:
    - Path: ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/archive/phase7-benchmarks.sqlite
      Note: SQLite database storing benchmark matrix results
    - Path: ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/scripts/import-benchmark-matrix-to-sqlite.py
      Note: SQLite importer for bench-matrix output
    - Path: ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/scripts/render-benchmark-report-from-sqlite.py
      Note: Markdown report renderer embedding SQL queries
ExternalSources: []
Summary: SQLite-backed benchmark matrix report generated from stored Vegeta and metrics-delta results.
LastUpdated: 2026-05-15T12:33:51Z
WhatFor: ""
WhenToUse: ""
---




# Benchmark Matrix SQLite Report: phase7-smoke-20260515T123252Z

This report was generated from SQLite database `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/archive/phase7-benchmarks.sqlite`.

Every section includes the SQL query used to generate it, followed by the query result.

## Matrix metadata

```sql
SELECT matrix_id, created_at_utc, imported_at_utc, repo_commit, git_dirty,
       duration, warmup_duration, scenarios, rates, repeat_count, out_root
FROM benchmark_matrices
WHERE matrix_id = :matrix_id;
```

| matrix_id | created_at_utc | imported_at_utc | repo_commit | git_dirty | duration | warmup_duration | scenarios | rates | repeat_count | out_root |
|---|---|---|---|---|---|---|---|---|---|---|
| phase7-smoke-20260515T123252Z | 2026-05-15T12:33:51Z | 2026-05-15T12:33:51Z | c8ed0c83b391546e07f9c18c356e26db0d0fbab6 | 1 | 2s | 0s | null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed | 2/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-smoke-20260515T123252Z |

## Scenario/rate aggregate latency and success

```sql
SELECT
  scenario,
  rate_target,
  COUNT(*) AS runs,
  SUM(requests) AS requests,
  ROUND(AVG(success_ratio), 4) AS success_avg,
  ROUND(MIN(p95_ms), 2) AS p95_min_ms,
  ROUND(AVG(p95_ms), 2) AS p95_avg_ms,
  ROUND(MAX(p95_ms), 2) AS p95_max_ms,
  ROUND(AVG(p99_ms), 2) AS p99_avg_ms,
  ROUND(AVG(throughput), 2) AS throughput_avg
FROM benchmark_runs
WHERE matrix_id = :matrix_id
GROUP BY scenario, rate_target
ORDER BY scenario, rate_target;
```

| scenario | rate_target | runs | requests | success_avg | p95_min_ms | p95_avg_ms | p95_max_ms | p99_avg_ms | throughput_avg |
|---|---|---|---|---|---|---|---|---|---|
| db-read | 2/s | 1 | 4 | 1.0 | 1.94 | 1.94 | 1.94 | 1.94 | 2.66 |
| db-write | 2/s | 1 | 4 | 1.0 | 11.96 | 11.96 | 11.96 | 11.96 | 2.66 |
| kanban-action | 2/s | 1 | 4 | 1.0 | 35.41 | 35.41 | 35.41 | 35.41 | 2.64 |
| kanban-fragment | 2/s | 1 | 4 | 1.0 | 42.27 | 42.27 | 42.27 | 42.27 | 2.58 |
| kanban-mixed | 2/s | 1 | 4 | 1.0 | 17.93 | 17.93 | 17.93 | 17.93 | 2.65 |
| null | 2/s | 1 | 4 | 1.0 | 0.74 | 0.74 | 0.74 | 0.74 | 2.67 |
| render | 2/s | 1 | 4 | 1.0 | 1.98 | 1.98 | 1.98 | 1.98 | 2.67 |

## Slowest individual runs by p95

```sql
SELECT
  scenario,
  rate_target,
  run_number,
  requests,
  ROUND(success_ratio, 4) AS success_ratio,
  ROUND(p50_ms, 2) AS p50_ms,
  ROUND(p95_ms, 2) AS p95_ms,
  ROUND(p99_ms, 2) AS p99_ms,
  ROUND(max_ms, 2) AS max_ms,
  status_codes_json,
  errors_json
FROM benchmark_runs
WHERE matrix_id = :matrix_id
ORDER BY p95_ms DESC
LIMIT 20;
```

| scenario | rate_target | run_number | requests | success_ratio | p50_ms | p95_ms | p99_ms | max_ms | status_codes_json | errors_json |
|---|---|---|---|---|---|---|---|---|---|---|
| kanban-fragment | 2/s | 1 | 4 | 1.0 | 20.36 | 42.27 | 42.27 | 42.27 | {"200": 4} | [] |
| kanban-action | 2/s | 1 | 4 | 1.0 | 22.56 | 35.41 | 35.41 | 35.41 | {"200": 4} | [] |
| kanban-mixed | 2/s | 1 | 4 | 1.0 | 11.85 | 17.93 | 17.93 | 17.93 | {"200": 4} | [] |
| db-write | 2/s | 1 | 4 | 1.0 | 7.97 | 11.96 | 11.96 | 11.96 | {"200": 4} | [] |
| render | 2/s | 1 | 4 | 1.0 | 1.41 | 1.98 | 1.98 | 1.98 | {"200": 4} | [] |
| db-read | 2/s | 1 | 4 | 1.0 | 1.71 | 1.94 | 1.94 | 1.94 | {"200": 4} | [] |
| null | 2/s | 1 | 4 | 1.0 | 0.59 | 0.74 | 0.74 | 0.74 | {"200": 4} | [] |

## Runs with non-100% success or errors

```sql
SELECT
  scenario,
  rate_target,
  run_number,
  requests,
  ROUND(success_ratio, 4) AS success_ratio,
  status_codes_json,
  errors_json,
  out_dir
FROM benchmark_runs
WHERE matrix_id = :matrix_id
  AND (success_ratio < 1 OR errors_json <> '[]')
ORDER BY scenario, rate_target, run_number;
```

_No rows._

## HTTP request deltas captured from Prometheus metrics

```sql
SELECT
  scenario,
  rate_target,
  metric_name,
  labels_json,
  SUM(delta_value) AS delta_sum
FROM benchmark_metric_deltas
WHERE matrix_id = :matrix_id
  AND metric_name = 'goja_site_http_requests_total'
GROUP BY scenario, rate_target, metric_name, labels_json
ORDER BY scenario, rate_target, labels_json;
```

| scenario | rate_target | metric_name | labels_json | delta_sum |
|---|---|---|---|---|
| db-read | 2/s | goja_site_http_requests_total | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 4.0 |
| db-write | 2/s | goja_site_http_requests_total | {"method": "POST", "route": "other", "site": "default", "status_class": "2xx"} | 4.0 |
| kanban-action | 2/s | goja_site_http_requests_total | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 4.0 |
| kanban-fragment | 2/s | goja_site_http_requests_total | {"method": "GET", "route": "/_kanban/:board/fragment", "site": "default", "status_class": "2xx"} | 4.0 |
| kanban-mixed | 2/s | goja_site_http_requests_total | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 1.0 |
| kanban-mixed | 2/s | goja_site_http_requests_total | {"method": "GET", "route": "/_kanban/:board/fragment", "site": "default", "status_class": "2xx"} | 2.0 |
| kanban-mixed | 2/s | goja_site_http_requests_total | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 1.0 |
| null | 2/s | goja_site_http_requests_total | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 4.0 |
| render | 2/s | goja_site_http_requests_total | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 4.0 |

## DB operation deltas captured from Prometheus metrics

```sql
SELECT
  scenario,
  rate_target,
  labels_json,
  SUM(delta_value) AS delta_sum
FROM benchmark_metric_deltas
WHERE matrix_id = :matrix_id
  AND metric_name = 'goja_site_db_operations_total'
GROUP BY scenario, rate_target, labels_json
ORDER BY scenario, rate_target, labels_json;
```

| scenario | rate_target | labels_json | delta_sum |
|---|---|---|---|
| db-read | 2/s | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 40.0 |
| db-write | 2/s | {"db_policy": "simple", "operation": "exec", "site": "default", "sql_kind": "insert"} | 4.0 |
| db-write | 2/s | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 4.0 |

## Kanban action and fragment metric deltas

```sql
SELECT
  scenario,
  rate_target,
  metric_name,
  labels_json,
  SUM(delta_value) AS delta_sum
FROM benchmark_metric_deltas
WHERE matrix_id = :matrix_id
  AND metric_name IN (
    'goja_site_kanban_action_duration_seconds_count',
    'goja_site_kanban_fragment_duration_seconds_count',
    'goja_site_kanban_render_duration_seconds_count'
  )
GROUP BY scenario, rate_target, metric_name, labels_json
ORDER BY scenario, rate_target, metric_name, labels_json;
```

| scenario | rate_target | metric_name | labels_json | delta_sum |
|---|---|---|---|---|
| kanban-action | 2/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 4.0 |
| kanban-action | 2/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 4.0 |
| kanban-fragment | 2/s | goja_site_kanban_fragment_duration_seconds_count | {"board": "bench", "site": "default"} | 4.0 |
| kanban-fragment | 2/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "fragment", "site": "default"} | 4.0 |
| kanban-mixed | 2/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 1.0 |
| kanban-mixed | 2/s | goja_site_kanban_fragment_duration_seconds_count | {"board": "bench", "site": "default"} | 2.0 |
| kanban-mixed | 2/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 1.0 |
| kanban-mixed | 2/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "fragment", "site": "default"} | 2.0 |

## Stored artifact locations

```sql
SELECT scenario, rate_target, run_number, out_dir
FROM benchmark_runs
WHERE matrix_id = :matrix_id
ORDER BY scenario, rate_target, run_number;
```

| scenario | rate_target | run_number | out_dir |
|---|---|---|---|
| db-read | 2/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-smoke-20260515T123252Z/db-read/rate-2_s/run-1 |
| db-write | 2/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-smoke-20260515T123252Z/db-write/rate-2_s/run-1 |
| kanban-action | 2/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-smoke-20260515T123252Z/kanban-action/rate-2_s/run-1 |
| kanban-fragment | 2/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-smoke-20260515T123252Z/kanban-fragment/rate-2_s/run-1 |
| kanban-mixed | 2/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-smoke-20260515T123252Z/kanban-mixed/rate-2_s/run-1 |
| null | 2/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-smoke-20260515T123252Z/null/rate-2_s/run-1 |
| render | 2/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-smoke-20260515T123252Z/render/rate-2_s/run-1 |
