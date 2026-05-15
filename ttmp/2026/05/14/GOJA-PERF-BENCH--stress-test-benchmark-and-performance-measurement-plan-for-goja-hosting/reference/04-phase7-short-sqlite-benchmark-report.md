---
Title: Benchmark Matrix SQLite Report - phase7-short-20260515T125010Z
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
      Note: SQLite database containing the short matrix rows
    - Path: ttmp/2026/05/14/GOJA-PERF-BENCH--stress-test-benchmark-and-performance-measurement-plan-for-goja-hosting/scripts/run-phase7-short-matrix.sh
      Note: retrace script used for the short matrix run
ExternalSources: []
Summary: SQLite-backed benchmark matrix report generated from stored Vegeta and metrics-delta results.
LastUpdated: 2026-05-15T14:04:27Z
WhatFor: ""
WhenToUse: ""
---



# Benchmark Matrix SQLite Report: phase7-short-20260515T125010Z

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
| phase7-short-20260515T125010Z | 2026-05-15T14:04:26Z | 2026-05-15T14:04:26Z | d111109733062a5c7a40c8059c483b53385871f8 | 1 | 60s | 10s | null,render,db-read,db-write,kanban-fragment,kanban-action,kanban-mixed | 5/s,10/s,25/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z |

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
| db-read | 10/s | 3 | 1800 | 1.0 | 2.84 | 3.2 | 3.42 | 4.12 | 10.02 |
| db-read | 25/s | 3 | 4500 | 1.0 | 3.51 | 3.57 | 3.64 | 4.14 | 25.02 |
| db-read | 5/s | 3 | 900 | 1.0 | 2.71 | 3.17 | 3.49 | 4.08 | 5.02 |
| db-write | 10/s | 3 | 1800 | 1.0 | 5.17 | 5.24 | 5.32 | 5.75 | 10.02 |
| db-write | 25/s | 3 | 4500 | 1.0 | 4.71 | 4.87 | 5.09 | 7.39 | 25.02 |
| db-write | 5/s | 3 | 900 | 1.0 | 5.57 | 5.96 | 6.65 | 8.54 | 5.02 |
| kanban-action | 10/s | 3 | 1800 | 1.0 | 23.0 | 26.71 | 31.65 | 42.03 | 10.02 |
| kanban-action | 25/s | 3 | 4500 | 1.0 | 24.09 | 35.24 | 51.73 | 59.5 | 25.01 |
| kanban-action | 5/s | 3 | 900 | 1.0 | 23.48 | 29.64 | 35.49 | 43.05 | 5.02 |
| kanban-fragment | 10/s | 3 | 1800 | 1.0 | 17.81 | 19.81 | 21.06 | 32.89 | 10.02 |
| kanban-fragment | 25/s | 3 | 4500 | 1.0 | 20.12 | 20.61 | 21.47 | 32.16 | 25.01 |
| kanban-fragment | 5/s | 3 | 900 | 1.0 | 17.12 | 20.71 | 23.82 | 33.59 | 5.02 |
| kanban-mixed | 10/s | 3 | 1800 | 1.0 | 21.81 | 22.89 | 25.0 | 34.6 | 10.02 |
| kanban-mixed | 25/s | 3 | 4500 | 1.0 | 19.54 | 21.29 | 22.52 | 32.81 | 25.01 |
| kanban-mixed | 5/s | 3 | 900 | 1.0 | 19.93 | 24.52 | 28.58 | 35.19 | 5.02 |
| null | 10/s | 3 | 1800 | 1.0 | 1.06 | 1.07 | 1.09 | 1.45 | 10.02 |
| null | 25/s | 3 | 4500 | 1.0 | 0.85 | 0.91 | 0.99 | 1.34 | 25.02 |
| null | 5/s | 3 | 900 | 1.0 | 1.07 | 1.18 | 1.24 | 1.67 | 5.02 |
| render | 10/s | 3 | 1800 | 1.0 | 3.28 | 3.39 | 3.47 | 3.93 | 10.02 |
| render | 25/s | 3 | 4500 | 1.0 | 3.2 | 3.34 | 3.45 | 3.77 | 25.02 |
| render | 5/s | 3 | 900 | 1.0 | 2.79 | 2.91 | 3.06 | 3.57 | 5.02 |

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
| kanban-action | 25/s | 3 | 1500 | 1.0 | 13.0 | 51.73 | 98.07 | 130.76 | {"200": 1500} | [] |
| kanban-action | 5/s | 2 | 300 | 1.0 | 10.38 | 35.49 | 44.18 | 55.24 | {"200": 300} | [] |
| kanban-action | 10/s | 1 | 600 | 1.0 | 10.63 | 31.65 | 59.73 | 105.39 | {"200": 600} | [] |
| kanban-action | 5/s | 1 | 300 | 1.0 | 10.21 | 29.95 | 53.35 | 66.17 | {"200": 300} | [] |
| kanban-action | 25/s | 2 | 1500 | 1.0 | 9.7 | 29.89 | 45.03 | 80.73 | {"200": 1500} | [] |
| kanban-mixed | 5/s | 3 | 300 | 1.0 | 9.43 | 28.58 | 36.86 | 38.96 | {"200": 300} | [] |
| kanban-action | 10/s | 3 | 600 | 1.0 | 9.79 | 25.48 | 35.03 | 43.41 | {"200": 600} | [] |
| kanban-mixed | 5/s | 1 | 300 | 1.0 | 8.57 | 25.04 | 36.32 | 37.16 | {"200": 300} | [] |
| kanban-mixed | 10/s | 1 | 600 | 1.0 | 9.54 | 25.0 | 36.57 | 39.94 | {"200": 600} | [] |
| kanban-action | 25/s | 1 | 1500 | 1.0 | 8.47 | 24.09 | 35.4 | 43.28 | {"200": 1500} | [] |
| kanban-fragment | 5/s | 1 | 300 | 1.0 | 8.25 | 23.82 | 37.83 | 44.86 | {"200": 300} | [] |
| kanban-action | 5/s | 3 | 300 | 1.0 | 8.98 | 23.48 | 31.61 | 38.51 | {"200": 300} | [] |
| kanban-action | 10/s | 2 | 600 | 1.0 | 9.2 | 23.0 | 31.32 | 37.76 | {"200": 600} | [] |
| kanban-mixed | 25/s | 1 | 1500 | 1.0 | 7.27 | 22.52 | 32.34 | 39.21 | {"200": 1500} | [] |
| kanban-mixed | 10/s | 3 | 600 | 1.0 | 8.43 | 21.87 | 35.77 | 42.42 | {"200": 600} | [] |
| kanban-mixed | 10/s | 2 | 600 | 1.0 | 7.86 | 21.81 | 31.47 | 37.15 | {"200": 600} | [] |
| kanban-mixed | 25/s | 2 | 1500 | 1.0 | 7.81 | 21.8 | 35.59 | 43.44 | {"200": 1500} | [] |
| kanban-fragment | 25/s | 3 | 1500 | 1.0 | 7.94 | 21.47 | 32.09 | 51.78 | {"200": 1500} | [] |
| kanban-fragment | 5/s | 2 | 300 | 1.0 | 7.08 | 21.18 | 35.24 | 36.96 | {"200": 300} | [] |
| kanban-fragment | 10/s | 1 | 600 | 1.0 | 8.28 | 21.06 | 35.55 | 42.95 | {"200": 600} | [] |

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
| db-read | 10/s | goja_site_http_requests_total | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 1800.0 |
| db-read | 25/s | goja_site_http_requests_total | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 4500.0 |
| db-read | 5/s | goja_site_http_requests_total | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 900.0 |
| db-write | 10/s | goja_site_http_requests_total | {"method": "POST", "route": "other", "site": "default", "status_class": "2xx"} | 1800.0 |
| db-write | 25/s | goja_site_http_requests_total | {"method": "POST", "route": "other", "site": "default", "status_class": "2xx"} | 4500.0 |
| db-write | 5/s | goja_site_http_requests_total | {"method": "POST", "route": "other", "site": "default", "status_class": "2xx"} | 900.0 |
| kanban-action | 10/s | goja_site_http_requests_total | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 1800.0 |
| kanban-action | 25/s | goja_site_http_requests_total | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 4500.0 |
| kanban-action | 5/s | goja_site_http_requests_total | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 900.0 |
| kanban-fragment | 10/s | goja_site_http_requests_total | {"method": "GET", "route": "/_kanban/:board/fragment", "site": "default", "status_class": "2xx"} | 1800.0 |
| kanban-fragment | 25/s | goja_site_http_requests_total | {"method": "GET", "route": "/_kanban/:board/fragment", "site": "default", "status_class": "2xx"} | 4500.0 |
| kanban-fragment | 5/s | goja_site_http_requests_total | {"method": "GET", "route": "/_kanban/:board/fragment", "site": "default", "status_class": "2xx"} | 900.0 |
| kanban-mixed | 10/s | goja_site_http_requests_total | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 165.0 |
| kanban-mixed | 10/s | goja_site_http_requests_total | {"method": "GET", "route": "/_kanban/:board/fragment", "site": "default", "status_class": "2xx"} | 1308.0 |
| kanban-mixed | 10/s | goja_site_http_requests_total | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 327.0 |
| kanban-mixed | 25/s | goja_site_http_requests_total | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 411.0 |
| kanban-mixed | 25/s | goja_site_http_requests_total | {"method": "GET", "route": "/_kanban/:board/fragment", "site": "default", "status_class": "2xx"} | 3270.0 |
| kanban-mixed | 25/s | goja_site_http_requests_total | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 819.0 |
| kanban-mixed | 5/s | goja_site_http_requests_total | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 84.0 |
| kanban-mixed | 5/s | goja_site_http_requests_total | {"method": "GET", "route": "/_kanban/:board/fragment", "site": "default", "status_class": "2xx"} | 651.0 |
| kanban-mixed | 5/s | goja_site_http_requests_total | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 165.0 |
| null | 10/s | goja_site_http_requests_total | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 1800.0 |
| null | 25/s | goja_site_http_requests_total | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 4500.0 |
| null | 5/s | goja_site_http_requests_total | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 900.0 |
| render | 10/s | goja_site_http_requests_total | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 1800.0 |
| render | 25/s | goja_site_http_requests_total | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 4500.0 |
| render | 5/s | goja_site_http_requests_total | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 900.0 |

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
| db-read | 10/s | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 18000.0 |
| db-read | 25/s | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 45000.0 |
| db-read | 5/s | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 9000.0 |
| db-write | 10/s | {"db_policy": "simple", "operation": "exec", "site": "default", "sql_kind": "insert"} | 1800.0 |
| db-write | 10/s | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 1800.0 |
| db-write | 25/s | {"db_policy": "simple", "operation": "exec", "site": "default", "sql_kind": "insert"} | 4500.0 |
| db-write | 25/s | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 4500.0 |
| db-write | 5/s | {"db_policy": "simple", "operation": "exec", "site": "default", "sql_kind": "insert"} | 900.0 |
| db-write | 5/s | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 900.0 |

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
| kanban-action | 10/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 1800.0 |
| kanban-action | 10/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 1800.0 |
| kanban-action | 25/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 4500.0 |
| kanban-action | 25/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 4500.0 |
| kanban-action | 5/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 900.0 |
| kanban-action | 5/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 900.0 |
| kanban-fragment | 10/s | goja_site_kanban_fragment_duration_seconds_count | {"board": "bench", "site": "default"} | 1800.0 |
| kanban-fragment | 10/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "fragment", "site": "default"} | 1800.0 |
| kanban-fragment | 25/s | goja_site_kanban_fragment_duration_seconds_count | {"board": "bench", "site": "default"} | 4500.0 |
| kanban-fragment | 25/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "fragment", "site": "default"} | 4500.0 |
| kanban-fragment | 5/s | goja_site_kanban_fragment_duration_seconds_count | {"board": "bench", "site": "default"} | 900.0 |
| kanban-fragment | 5/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "fragment", "site": "default"} | 900.0 |
| kanban-mixed | 10/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 327.0 |
| kanban-mixed | 10/s | goja_site_kanban_fragment_duration_seconds_count | {"board": "bench", "site": "default"} | 1308.0 |
| kanban-mixed | 10/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 327.0 |
| kanban-mixed | 10/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "fragment", "site": "default"} | 1308.0 |
| kanban-mixed | 25/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 819.0 |
| kanban-mixed | 25/s | goja_site_kanban_fragment_duration_seconds_count | {"board": "bench", "site": "default"} | 3270.0 |
| kanban-mixed | 25/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 819.0 |
| kanban-mixed | 25/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "fragment", "site": "default"} | 3270.0 |
| kanban-mixed | 5/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 165.0 |
| kanban-mixed | 5/s | goja_site_kanban_fragment_duration_seconds_count | {"board": "bench", "site": "default"} | 651.0 |
| kanban-mixed | 5/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 165.0 |
| kanban-mixed | 5/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "fragment", "site": "default"} | 651.0 |

## Stored artifact locations

```sql
SELECT scenario, rate_target, run_number, out_dir
FROM benchmark_runs
WHERE matrix_id = :matrix_id
ORDER BY scenario, rate_target, run_number;
```

| scenario | rate_target | run_number | out_dir |
|---|---|---|---|
| db-read | 10/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-read/rate-10_s/run-1 |
| db-read | 10/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-read/rate-10_s/run-2 |
| db-read | 10/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-read/rate-10_s/run-3 |
| db-read | 25/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-read/rate-25_s/run-1 |
| db-read | 25/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-read/rate-25_s/run-2 |
| db-read | 25/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-read/rate-25_s/run-3 |
| db-read | 5/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-read/rate-5_s/run-1 |
| db-read | 5/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-read/rate-5_s/run-2 |
| db-read | 5/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-read/rate-5_s/run-3 |
| db-write | 10/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-write/rate-10_s/run-1 |
| db-write | 10/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-write/rate-10_s/run-2 |
| db-write | 10/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-write/rate-10_s/run-3 |
| db-write | 25/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-write/rate-25_s/run-1 |
| db-write | 25/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-write/rate-25_s/run-2 |
| db-write | 25/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-write/rate-25_s/run-3 |
| db-write | 5/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-write/rate-5_s/run-1 |
| db-write | 5/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-write/rate-5_s/run-2 |
| db-write | 5/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/db-write/rate-5_s/run-3 |
| kanban-action | 10/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-action/rate-10_s/run-1 |
| kanban-action | 10/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-action/rate-10_s/run-2 |
| kanban-action | 10/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-action/rate-10_s/run-3 |
| kanban-action | 25/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-action/rate-25_s/run-1 |
| kanban-action | 25/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-action/rate-25_s/run-2 |
| kanban-action | 25/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-action/rate-25_s/run-3 |
| kanban-action | 5/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-action/rate-5_s/run-1 |
| kanban-action | 5/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-action/rate-5_s/run-2 |
| kanban-action | 5/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-action/rate-5_s/run-3 |
| kanban-fragment | 10/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-fragment/rate-10_s/run-1 |
| kanban-fragment | 10/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-fragment/rate-10_s/run-2 |
| kanban-fragment | 10/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-fragment/rate-10_s/run-3 |
| kanban-fragment | 25/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-fragment/rate-25_s/run-1 |
| kanban-fragment | 25/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-fragment/rate-25_s/run-2 |
| kanban-fragment | 25/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-fragment/rate-25_s/run-3 |
| kanban-fragment | 5/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-fragment/rate-5_s/run-1 |
| kanban-fragment | 5/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-fragment/rate-5_s/run-2 |
| kanban-fragment | 5/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-fragment/rate-5_s/run-3 |
| kanban-mixed | 10/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-mixed/rate-10_s/run-1 |
| kanban-mixed | 10/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-mixed/rate-10_s/run-2 |
| kanban-mixed | 10/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-mixed/rate-10_s/run-3 |
| kanban-mixed | 25/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-mixed/rate-25_s/run-1 |
| kanban-mixed | 25/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-mixed/rate-25_s/run-2 |
| kanban-mixed | 25/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-mixed/rate-25_s/run-3 |
| kanban-mixed | 5/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-mixed/rate-5_s/run-1 |
| kanban-mixed | 5/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-mixed/rate-5_s/run-2 |
| kanban-mixed | 5/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/kanban-mixed/rate-5_s/run-3 |
| null | 10/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/null/rate-10_s/run-1 |
| null | 10/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/null/rate-10_s/run-2 |
| null | 10/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/null/rate-10_s/run-3 |
| null | 25/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/null/rate-25_s/run-1 |
| null | 25/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/null/rate-25_s/run-2 |
| null | 25/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/null/rate-25_s/run-3 |
| null | 5/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/null/rate-5_s/run-1 |
| null | 5/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/null/rate-5_s/run-2 |
| null | 5/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/null/rate-5_s/run-3 |
| render | 10/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/render/rate-10_s/run-1 |
| render | 10/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/render/rate-10_s/run-2 |
| render | 10/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/render/rate-10_s/run-3 |
| render | 25/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/render/rate-25_s/run-1 |
| render | 25/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/render/rate-25_s/run-2 |
| render | 25/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/render/rate-25_s/run-3 |
| render | 5/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/render/rate-5_s/run-1 |
| render | 5/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/render/rate-5_s/run-2 |
| render | 5/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/phase7-short-20260515T125010Z/render/rate-5_s/run-3 |
