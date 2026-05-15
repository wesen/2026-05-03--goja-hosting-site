---
Title: Quick Stress Sweep SQLite Report - stress-quick-20260515T145900Z
Ticket: GOJA-STRESS-TEST
Status: active
Topics:
    - benchmarking
    - stress-testing
    - sqlite
    - vegeta
    - prometheus
DocType: reference
Intent: historical
Owners: []
RelatedFiles:
    - Path: ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/archive/stress-benchmarks.sqlite
      Note: SQLite DB containing stress-quick-20260515T145900Z
    - Path: ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/scripts/run-stress-quick-sweep.sh
      Note: script used for the quick stress sweep
ExternalSources: []
Summary: SQLite-backed stress test report with embedded SQL queries for breakdown analysis.
LastUpdated: 2026-05-15T15:02:15Z
WhatFor: ""
WhenToUse: ""
---



# Quick Stress Sweep SQLite Report: stress-quick-20260515T145900Z

This report was generated from SQLite database `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/archive/stress-benchmarks.sqlite`.

Each section includes the exact SQL query used to generate the table.

## Breakdown criteria used in this report

The report flags a breakdown candidate when any grouped scenario/rate cell shows one of:

- `success_min < 0.99`,
- non-empty Vegeta errors,
- achieved throughput below 95% of offered rate,
- p95 above 100 ms for light scenarios (`null`, `render`, `db-read`),
- p95 above 250 ms for heavier scenarios (`db-write`, Kanban scenarios),
- max latency above 1000 ms.

## Matrix metadata

What was run, where the ignored raw artifacts live, and which commit produced the data.

```sql
SELECT matrix_id, created_at_utc, imported_at_utc, repo_commit, git_dirty,
       duration, warmup_duration, scenarios, rates, repeat_count, out_root
FROM benchmark_matrices
WHERE matrix_id = :matrix_id;
```

| matrix_id | created_at_utc | imported_at_utc | repo_commit | git_dirty | duration | warmup_duration | scenarios | rates | repeat_count | out_root |
|---|---|---|---|---|---|---|---|---|---|---|
| stress-quick-20260515T145900Z | 2026-05-15T15:02:05Z | 2026-05-15T15:02:05Z | 8691e5b8367058b176b463ea21620a5100972b21 | 1 | 10s | 3s | null,render,db-write,kanban-action | 50/s,100/s,200/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z |

## Offered vs achieved throughput and latency

Primary stress table: compares requested rate to achieved throughput, success, and tail latency.

```sql
SELECT
  scenario,
  rate_target,
  CAST(REPLACE(rate_target, '/s', '') AS REAL) AS offered_rps,
  COUNT(*) AS runs,
  SUM(requests) AS requests,
  ROUND(AVG(throughput), 2) AS throughput_avg,
  ROUND(AVG(throughput) / NULLIF(CAST(REPLACE(rate_target, '/s', '') AS REAL), 0), 3) AS throughput_ratio,
  ROUND(AVG(success_ratio), 4) AS success_avg,
  ROUND(AVG(p50_ms), 2) AS p50_avg_ms,
  ROUND(AVG(p95_ms), 2) AS p95_avg_ms,
  ROUND(AVG(p99_ms), 2) AS p99_avg_ms,
  ROUND(MAX(max_ms), 2) AS max_ms
FROM benchmark_runs
WHERE matrix_id = :matrix_id
GROUP BY scenario, rate_target
ORDER BY scenario, offered_rps;
```

| scenario | rate_target | offered_rps | runs | requests | throughput_avg | throughput_ratio | success_avg | p50_avg_ms | p95_avg_ms | p99_avg_ms | max_ms |
|---|---|---|---|---|---|---|---|---|---|---|---|
| db-write | 50/s | 50.0 | 1 | 500 | 50.09 | 1.002 | 1.0 | 2.11 | 4.01 | 7.14 | 12.18 |
| db-write | 100/s | 100.0 | 1 | 1000 | 100.08 | 1.001 | 1.0 | 1.73 | 3.58 | 6.06 | 13.94 |
| db-write | 200/s | 200.0 | 1 | 2000 | 200.07 | 1.0 | 1.0 | 1.55 | 3.07 | 5.13 | 10.22 |
| kanban-action | 50/s | 50.0 | 1 | 500 | 50.03 | 1.001 | 1.0 | 10.06 | 27.02 | 44.74 | 53.95 |
| kanban-action | 100/s | 100.0 | 1 | 1000 | 86.23 | 0.862 | 1.0 | 869.45 | 1527.67 | 1570.05 | 1607.18 |
| kanban-action | 200/s | 200.0 | 1 | 2000 | 96.04 | 0.48 | 1.0 | 5168.01 | 10347.13 | 10762.33 | 10829.1 |
| null | 50/s | 50.0 | 1 | 500 | 50.09 | 1.002 | 1.0 | 0.52 | 1.03 | 1.35 | 2.09 |
| null | 100/s | 100.0 | 1 | 1000 | 100.1 | 1.001 | 1.0 | 0.36 | 0.85 | 1.1 | 1.58 |
| null | 200/s | 200.0 | 1 | 2000 | 200.1 | 1.0 | 1.0 | 0.23 | 0.63 | 0.91 | 1.08 |
| render | 50/s | 50.0 | 1 | 500 | 50.09 | 1.002 | 1.0 | 0.99 | 2.96 | 3.48 | 3.95 |
| render | 100/s | 100.0 | 1 | 1000 | 100.09 | 1.001 | 1.0 | 0.68 | 2.75 | 3.23 | 4.94 |
| render | 200/s | 200.0 | 1 | 2000 | 200.07 | 1.0 | 1.0 | 0.65 | 2.38 | 3.21 | 10.8 |

## Breakdown candidates

Flags cells where success falls, throughput misses offered rate, p95 exceeds conservative stress thresholds, or max latency is very high.

```sql
WITH agg AS (
  SELECT
    scenario,
    rate_target,
    CAST(REPLACE(rate_target, '/s', '') AS REAL) AS offered_rps,
    COUNT(*) AS runs,
    MIN(success_ratio) AS success_min,
    AVG(success_ratio) AS success_avg,
    AVG(throughput) AS throughput_avg,
    AVG(throughput) / NULLIF(CAST(REPLACE(rate_target, '/s', '') AS REAL), 0) AS throughput_ratio,
    AVG(p95_ms) AS p95_avg_ms,
    MAX(max_ms) AS max_ms,
    MAX(errors_json) AS errors_json
  FROM benchmark_runs
  WHERE matrix_id = :matrix_id
  GROUP BY scenario, rate_target
)
SELECT
  scenario,
  rate_target,
  ROUND(success_min, 4) AS success_min,
  ROUND(throughput_ratio, 3) AS throughput_ratio,
  ROUND(p95_avg_ms, 2) AS p95_avg_ms,
  ROUND(max_ms, 2) AS max_ms,
  CASE
    WHEN success_min < 0.99 THEN 'success_loss'
    WHEN errors_json <> '[]' THEN 'vegeta_errors'
    WHEN throughput_ratio < 0.95 THEN 'throughput_shortfall'
    WHEN scenario IN ('null', 'render', 'db-read') AND p95_avg_ms > 100 THEN 'latency_threshold'
    WHEN scenario IN ('db-write', 'kanban-fragment', 'kanban-action', 'kanban-mixed') AND p95_avg_ms > 250 THEN 'latency_threshold'
    WHEN max_ms > 1000 THEN 'tail_outlier'
    ELSE 'ok'
  END AS breakdown_signal,
  errors_json
FROM agg
WHERE breakdown_signal <> 'ok'
ORDER BY scenario, offered_rps;
```

| scenario | rate_target | success_min | throughput_ratio | p95_avg_ms | max_ms | breakdown_signal | errors_json |
|---|---|---|---|---|---|---|---|
| kanban-action | 100/s | 1.0 | 0.862 | 1527.67 | 1607.18 | throughput_shortfall | [] |
| kanban-action | 200/s | 1.0 | 0.48 | 10347.13 | 10829.1 | throughput_shortfall | [] |

## Latency knee growth between adjacent rates

Looks for rate-to-rate p95 growth. Large growth factors identify where a scenario starts bending even before errors appear.

```sql
WITH agg AS (
  SELECT
    scenario,
    rate_target,
    CAST(REPLACE(rate_target, '/s', '') AS REAL) AS offered_rps,
    AVG(p95_ms) AS p95_avg_ms,
    AVG(p99_ms) AS p99_avg_ms,
    AVG(throughput) AS throughput_avg,
    AVG(success_ratio) AS success_avg
  FROM benchmark_runs
  WHERE matrix_id = :matrix_id
  GROUP BY scenario, rate_target
), ordered AS (
  SELECT
    *,
    LAG(p95_avg_ms) OVER (PARTITION BY scenario ORDER BY offered_rps) AS previous_p95_avg_ms,
    LAG(offered_rps) OVER (PARTITION BY scenario ORDER BY offered_rps) AS previous_offered_rps
  FROM agg
)
SELECT
  scenario,
  previous_offered_rps || '/s -> ' || rate_target AS rate_step,
  ROUND(previous_p95_avg_ms, 2) AS previous_p95_ms,
  ROUND(p95_avg_ms, 2) AS p95_ms,
  ROUND(p95_avg_ms / NULLIF(previous_p95_avg_ms, 0), 2) AS p95_growth_factor,
  ROUND(p99_avg_ms, 2) AS p99_ms,
  ROUND(throughput_avg, 2) AS throughput_avg,
  ROUND(success_avg, 4) AS success_avg
FROM ordered
WHERE previous_p95_avg_ms IS NOT NULL
ORDER BY p95_growth_factor DESC, scenario, offered_rps;
```

| scenario | rate_step | previous_p95_ms | p95_ms | p95_growth_factor | p99_ms | throughput_avg | success_avg |
|---|---|---|---|---|---|---|---|
| kanban-action | 50.0/s -> 100/s | 27.02 | 1527.67 | 56.53 | 1570.05 | 86.23 | 1.0 |
| kanban-action | 100.0/s -> 200/s | 1527.67 | 10347.13 | 6.77 | 10762.33 | 96.04 | 1.0 |
| render | 50.0/s -> 100/s | 2.96 | 2.75 | 0.93 | 3.23 | 100.09 | 1.0 |
| db-write | 50.0/s -> 100/s | 4.01 | 3.58 | 0.89 | 6.06 | 100.08 | 1.0 |
| db-write | 100.0/s -> 200/s | 3.58 | 3.07 | 0.86 | 5.13 | 200.07 | 1.0 |
| render | 100.0/s -> 200/s | 2.75 | 2.38 | 0.86 | 3.21 | 200.07 | 1.0 |
| null | 50.0/s -> 100/s | 1.03 | 0.85 | 0.83 | 1.1 | 100.1 | 1.0 |
| null | 100.0/s -> 200/s | 0.85 | 0.63 | 0.74 | 0.91 | 200.1 | 1.0 |

## Slowest individual runs by p95

The noisiest runs and likely candidates for targeted pprof reruns.

```sql
SELECT
  scenario,
  rate_target,
  run_number,
  requests,
  ROUND(success_ratio, 4) AS success_ratio,
  ROUND(throughput, 2) AS throughput,
  ROUND(p50_ms, 2) AS p50_ms,
  ROUND(p95_ms, 2) AS p95_ms,
  ROUND(p99_ms, 2) AS p99_ms,
  ROUND(max_ms, 2) AS max_ms,
  status_codes_json,
  errors_json
FROM benchmark_runs
WHERE matrix_id = :matrix_id
ORDER BY p95_ms DESC
LIMIT 25;
```

| scenario | rate_target | run_number | requests | success_ratio | throughput | p50_ms | p95_ms | p99_ms | max_ms | status_codes_json | errors_json |
|---|---|---|---|---|---|---|---|---|---|---|---|
| kanban-action | 200/s | 1 | 2000 | 1.0 | 96.04 | 5168.01 | 10347.13 | 10762.33 | 10829.1 | {"200": 2000} | [] |
| kanban-action | 100/s | 1 | 1000 | 1.0 | 86.23 | 869.45 | 1527.67 | 1570.05 | 1607.18 | {"200": 1000} | [] |
| kanban-action | 50/s | 1 | 500 | 1.0 | 50.03 | 10.06 | 27.02 | 44.74 | 53.95 | {"200": 500} | [] |
| db-write | 50/s | 1 | 500 | 1.0 | 50.09 | 2.11 | 4.01 | 7.14 | 12.18 | {"200": 500} | [] |
| db-write | 100/s | 1 | 1000 | 1.0 | 100.08 | 1.73 | 3.58 | 6.06 | 13.94 | {"200": 1000} | [] |
| db-write | 200/s | 1 | 2000 | 1.0 | 200.07 | 1.55 | 3.07 | 5.13 | 10.22 | {"200": 2000} | [] |
| render | 50/s | 1 | 500 | 1.0 | 50.09 | 0.99 | 2.96 | 3.48 | 3.95 | {"200": 500} | [] |
| render | 100/s | 1 | 1000 | 1.0 | 100.09 | 0.68 | 2.75 | 3.23 | 4.94 | {"200": 1000} | [] |
| render | 200/s | 1 | 2000 | 1.0 | 200.07 | 0.65 | 2.38 | 3.21 | 10.8 | {"200": 2000} | [] |
| null | 50/s | 1 | 500 | 1.0 | 50.09 | 0.52 | 1.03 | 1.35 | 2.09 | {"200": 500} | [] |
| null | 100/s | 1 | 1000 | 1.0 | 100.1 | 0.36 | 0.85 | 1.1 | 1.58 | {"200": 1000} | [] |
| null | 200/s | 1 | 2000 | 1.0 | 200.1 | 0.23 | 0.63 | 0.91 | 1.08 | {"200": 2000} | [] |

## Runs with non-100% success or Vegeta errors

Hard failures: anything here needs immediate investigation before interpreting latency numbers.

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
ORDER BY scenario, CAST(REPLACE(rate_target, '/s', '') AS REAL), run_number;
```

_No rows._

## HTTP route deltas

Prometheus route counters prove which route classes were actually hit at each stress rate.

```sql
SELECT
  scenario,
  rate_target,
  labels_json,
  SUM(delta_value) AS delta_sum
FROM benchmark_metric_deltas
WHERE matrix_id = :matrix_id
  AND metric_name = 'goja_site_http_requests_total'
GROUP BY scenario, rate_target, labels_json
ORDER BY scenario, CAST(REPLACE(rate_target, '/s', '') AS REAL), labels_json;
```

| scenario | rate_target | labels_json | delta_sum |
|---|---|---|---|
| db-write | 50/s | {"method": "POST", "route": "other", "site": "default", "status_class": "2xx"} | 500.0 |
| db-write | 100/s | {"method": "POST", "route": "other", "site": "default", "status_class": "2xx"} | 1000.0 |
| db-write | 200/s | {"method": "POST", "route": "other", "site": "default", "status_class": "2xx"} | 2000.0 |
| kanban-action | 50/s | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 500.0 |
| kanban-action | 100/s | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 1000.0 |
| kanban-action | 200/s | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 2000.0 |
| null | 50/s | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 500.0 |
| null | 100/s | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 1000.0 |
| null | 200/s | {"method": "GET", "route": "/", "site": "default", "status_class": "2xx"} | 2000.0 |
| render | 50/s | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 500.0 |
| render | 100/s | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 1000.0 |
| render | 200/s | {"method": "GET", "route": "other", "site": "default", "status_class": "2xx"} | 2000.0 |

## DB operation and error deltas

DB work performed and DB error evidence by scenario/rate.

```sql
SELECT
  scenario,
  rate_target,
  metric_name,
  labels_json,
  SUM(delta_value) AS delta_sum
FROM benchmark_metric_deltas
WHERE matrix_id = :matrix_id
  AND metric_name IN ('goja_site_db_operations_total', 'goja_site_db_errors_total')
GROUP BY scenario, rate_target, metric_name, labels_json
ORDER BY scenario, CAST(REPLACE(rate_target, '/s', '') AS REAL), metric_name, labels_json;
```

| scenario | rate_target | metric_name | labels_json | delta_sum |
|---|---|---|---|---|
| db-write | 50/s | goja_site_db_operations_total | {"db_policy": "simple", "operation": "exec", "site": "default", "sql_kind": "insert"} | 500.0 |
| db-write | 50/s | goja_site_db_operations_total | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 500.0 |
| db-write | 100/s | goja_site_db_operations_total | {"db_policy": "simple", "operation": "exec", "site": "default", "sql_kind": "insert"} | 1000.0 |
| db-write | 100/s | goja_site_db_operations_total | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 1000.0 |
| db-write | 200/s | goja_site_db_operations_total | {"db_policy": "simple", "operation": "exec", "site": "default", "sql_kind": "insert"} | 2000.0 |
| db-write | 200/s | goja_site_db_operations_total | {"db_policy": "simple", "operation": "query", "site": "default", "sql_kind": "select"} | 2000.0 |

## Kanban action/render deltas

Kanban work performed and action/fragment/render evidence by scenario/rate.

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
    'goja_site_kanban_render_duration_seconds_count',
    'goja_site_kanban_errors_total'
  )
GROUP BY scenario, rate_target, metric_name, labels_json
ORDER BY scenario, CAST(REPLACE(rate_target, '/s', '') AS REAL), metric_name, labels_json;
```

| scenario | rate_target | metric_name | labels_json | delta_sum |
|---|---|---|---|---|
| kanban-action | 50/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 500.0 |
| kanban-action | 50/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 500.0 |
| kanban-action | 100/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 1000.0 |
| kanban-action | 100/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 1000.0 |
| kanban-action | 200/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 2000.0 |
| kanban-action | 200/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 2000.0 |

## Stored artifact locations

Per-run raw artifacts. Use these directories for server logs, pprof files, raw Vegeta data, and metrics snapshots.

```sql
SELECT scenario, rate_target, run_number, out_dir
FROM benchmark_runs
WHERE matrix_id = :matrix_id
ORDER BY scenario, CAST(REPLACE(rate_target, '/s', '') AS REAL), run_number;
```

| scenario | rate_target | run_number | out_dir |
|---|---|---|---|
| db-write | 50/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/db-write/rate-50_s/run-1 |
| db-write | 100/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/db-write/rate-100_s/run-1 |
| db-write | 200/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/db-write/rate-200_s/run-1 |
| kanban-action | 50/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/kanban-action/rate-50_s/run-1 |
| kanban-action | 100/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/kanban-action/rate-100_s/run-1 |
| kanban-action | 200/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/kanban-action/rate-200_s/run-1 |
| null | 50/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/null/rate-50_s/run-1 |
| null | 100/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/null/rate-100_s/run-1 |
| null | 200/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/null/rate-200_s/run-1 |
| render | 50/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/render/rate-50_s/run-1 |
| render | 100/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/render/rate-100_s/run-1 |
| render | 200/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-quick-20260515T145900Z/render/rate-200_s/run-1 |
