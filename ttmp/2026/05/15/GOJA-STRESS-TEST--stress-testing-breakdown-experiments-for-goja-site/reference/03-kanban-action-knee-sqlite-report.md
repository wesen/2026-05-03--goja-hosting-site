---
Title: Kanban Action Knee Stress Report - stress-kanban-action-knee-20260515T150544Z
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
      Note: SQLite DB containing kanban-action knee matrix
ExternalSources: []
Summary: SQLite-backed stress test report with embedded SQL queries for breakdown analysis.
LastUpdated: 2026-05-15T15:15:04Z
WhatFor: ""
WhenToUse: ""
---


# Kanban Action Knee Stress Report: stress-kanban-action-knee-20260515T150544Z

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
| stress-kanban-action-knee-20260515T150544Z | 2026-05-15T15:15:04Z | 2026-05-15T15:15:04Z | a214bf3c132f29ba3d3080abec4de9f999068595 | 1 | 30s | 5s | kanban-action | 60/s,70/s,80/s,90/s,100/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z |

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
| kanban-action | 60/s | 60.0 | 3 | 5400 | 60.02 | 1.0 | 1.0 | 12.99 | 88.63 | 140.97 | 285.69 |
| kanban-action | 70/s | 70.0 | 3 | 6300 | 69.96 | 0.999 | 1.0 | 11.86 | 57.32 | 114.19 | 199.61 |
| kanban-action | 80/s | 80.0 | 3 | 7200 | 79.7 | 0.996 | 1.0 | 81.68 | 616.79 | 670.21 | 1034.3 |
| kanban-action | 90/s | 90.0 | 3 | 8100 | 88.16 | 0.98 | 1.0 | 308.79 | 838.29 | 940.24 | 1265.01 |
| kanban-action | 100/s | 100.0 | 3 | 9000 | 95.99 | 0.96 | 1.0 | 1480.67 | 1746.78 | 1789.84 | 4040.89 |

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
| kanban-action | 80/s | 1.0 | 0.996 | 616.79 | 1034.3 | latency_threshold | [] |
| kanban-action | 90/s | 1.0 | 0.98 | 838.29 | 1265.01 | latency_threshold | [] |
| kanban-action | 100/s | 1.0 | 0.96 | 1746.78 | 4040.89 | latency_threshold | [] |

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
| kanban-action | 70.0/s -> 80/s | 57.32 | 616.79 | 10.76 | 670.21 | 79.7 | 1.0 |
| kanban-action | 90.0/s -> 100/s | 838.29 | 1746.78 | 2.08 | 1789.84 | 95.99 | 1.0 |
| kanban-action | 80.0/s -> 90/s | 616.79 | 838.29 | 1.36 | 940.24 | 88.16 | 1.0 |
| kanban-action | 60.0/s -> 70/s | 88.63 | 57.32 | 0.65 | 114.19 | 69.96 | 1.0 |

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
| kanban-action | 100/s | 3 | 3000 | 1.0 | 88.2 | 3625.57 | 3984.21 | 4022.12 | 4040.89 | {"200": 3000} | [] |
| kanban-action | 90/s | 1 | 2700 | 1.0 | 87.01 | 482.89 | 1174.97 | 1243.39 | 1265.01 | {"200": 2700} | [] |
| kanban-action | 80/s | 1 | 2400 | 1.0 | 79.09 | 114.29 | 926.2 | 1018.38 | 1034.3 | {"200": 2400} | [] |
| kanban-action | 90/s | 3 | 2700 | 1.0 | 87.64 | 57.45 | 762.22 | 865.3 | 876.06 | {"200": 2700} | [] |
| kanban-action | 100/s | 2 | 3000 | 1.0 | 99.89 | 479.49 | 677.75 | 738.73 | 772.04 | {"200": 3000} | [] |
| kanban-action | 80/s | 2 | 2400 | 1.0 | 80.01 | 103.71 | 652.12 | 678.59 | 714.54 | {"200": 2400} | [] |
| kanban-action | 100/s | 1 | 3000 | 1.0 | 99.89 | 336.95 | 578.4 | 608.67 | 623.7 | {"200": 3000} | [] |
| kanban-action | 90/s | 2 | 2700 | 1.0 | 89.84 | 386.03 | 577.67 | 712.03 | 740.79 | {"200": 2700} | [] |
| kanban-action | 80/s | 3 | 2400 | 1.0 | 79.99 | 27.04 | 272.04 | 313.67 | 342.38 | {"200": 2400} | [] |
| kanban-action | 60/s | 3 | 1800 | 1.0 | 60.02 | 14.59 | 180.08 | 252.73 | 285.69 | {"200": 1800} | [] |
| kanban-action | 70/s | 3 | 2100 | 1.0 | 69.86 | 10.91 | 66.32 | 104.93 | 148.97 | {"200": 2100} | [] |
| kanban-action | 70/s | 2 | 2100 | 1.0 | 70.01 | 13.52 | 61.75 | 165.9 | 199.61 | {"200": 2100} | [] |
| kanban-action | 60/s | 1 | 1800 | 1.0 | 60.02 | 12.54 | 44.56 | 107.18 | 143.04 | {"200": 1800} | [] |
| kanban-action | 70/s | 1 | 2100 | 1.0 | 70.01 | 11.15 | 43.89 | 71.74 | 93.59 | {"200": 2100} | [] |
| kanban-action | 60/s | 2 | 1800 | 1.0 | 60.02 | 11.85 | 41.25 | 63.0 | 75.85 | {"200": 1800} | [] |

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
| kanban-action | 60/s | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 5400.0 |
| kanban-action | 70/s | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 6300.0 |
| kanban-action | 80/s | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 7200.0 |
| kanban-action | 90/s | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 8100.0 |
| kanban-action | 100/s | {"method": "POST", "route": "/_kanban/:board/action/:action", "site": "default", "status_class": "2xx"} | 9000.0 |

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

_No rows._

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
| kanban-action | 60/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 5400.0 |
| kanban-action | 60/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 5400.0 |
| kanban-action | 70/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 6300.0 |
| kanban-action | 70/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 6300.0 |
| kanban-action | 80/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 7200.0 |
| kanban-action | 80/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 7200.0 |
| kanban-action | 90/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 8100.0 |
| kanban-action | 90/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 8100.0 |
| kanban-action | 100/s | goja_site_kanban_action_duration_seconds_count | {"action": "cardMoved", "board": "bench", "refresh": "true", "site": "default"} | 9000.0 |
| kanban-action | 100/s | goja_site_kanban_render_duration_seconds_count | {"board": "bench", "reason": "action_refresh", "site": "default"} | 9000.0 |

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
| kanban-action | 60/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-60_s/run-1 |
| kanban-action | 60/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-60_s/run-2 |
| kanban-action | 60/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-60_s/run-3 |
| kanban-action | 70/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-70_s/run-1 |
| kanban-action | 70/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-70_s/run-2 |
| kanban-action | 70/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-70_s/run-3 |
| kanban-action | 80/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-80_s/run-1 |
| kanban-action | 80/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-80_s/run-2 |
| kanban-action | 80/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-80_s/run-3 |
| kanban-action | 90/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-90_s/run-1 |
| kanban-action | 90/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-90_s/run-2 |
| kanban-action | 90/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-90_s/run-3 |
| kanban-action | 100/s | 1 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-100_s/run-1 |
| kanban-action | 100/s | 2 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-100_s/run-2 |
| kanban-action | 100/s | 3 | /home/manuel/code/wesen/2026-05-03--goja-hosting-site/bench/results/stress-kanban-action-knee-20260515T150544Z/kanban-action/rate-100_s/run-3 |
