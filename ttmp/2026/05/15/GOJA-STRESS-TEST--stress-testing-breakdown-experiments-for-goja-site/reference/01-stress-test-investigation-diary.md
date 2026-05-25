---
Title: Stress Test Investigation Diary
Ticket: GOJA-STRESS-TEST
Status: active
Topics:
    - benchmarking
    - stress-testing
    - sqlite
    - observability
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Chronological diary for GOJA-STRESS-TEST implementation and stress benchmark runs."
LastUpdated: 2026-05-15T15:00:00Z
WhatFor: "Use this to reconstruct what was changed, run, validated, and uploaded for the stress testing ticket."
WhenToUse: "During review or when continuing stress testing from the quick sweep to hour-scale experiments."
---

# Stress Test Investigation Diary

## Step 1: Create stress testing ticket and plan quick/hour sweep workflow

I created a new docmgr ticket for stress testing breakdown experiments and wrote the initial plan plus retraceable scripts.

### Prompt Context

**User prompt (verbatim):** "Ok, create a new ticket to do the stress testing, and add script to at first do a reasonably short test of like 2-3 minutes just to make sure things work, before embarking on an hour long experiment. Then write a detailed report in the new ticket and upload it to the remarkable tablet."

**Assistant interpretation:** Create a separate ticket from `GOJA-PERF-BENCH`, add scripts in that ticket for a quick validation sweep and an hour-scale stress sweep, run the quick sweep, write a detailed SQL-backed report, and upload it to reMarkable.

**Inferred user intent:** Keep stress-testing work separate and auditable, with a short safety pass before long experiments.

**Commit (code):** pending at time of this entry.

### What I did

Created ticket:

```text
GOJA-STRESS-TEST
```

Ticket path:

```text
ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site
```

Created/updated docs:

```text
design-doc/01-goja-site-stress-testing-breakdown-plan.md
reference/01-stress-test-investigation-diary.md
reference/02-quick-stress-sweep-sqlite-report.md
```

Added scripts:

```text
scripts/sqlite-benchmark-schema.sql
scripts/import-benchmark-matrix-to-sqlite.py
scripts/render-stress-report-from-sqlite.py
scripts/run-stress-quick-sweep.sh
scripts/run-stress-hour-sweep.sh
```

### Why

The short baseline in `GOJA-PERF-BENCH` only went to 25/s and did not identify breakdown. This new ticket isolates stress testing and defines how to detect breakdown using SQL-backed evidence.

### What worked

- `docmgr ticket create-ticket` created the ticket workspace.
- The existing SQLite importer pattern could be reused and kept self-contained in the new ticket scripts folder.
- The stress report renderer adds stress-specific queries for throughput shortfall, latency knees, and breakdown candidates.

### What should happen next

Run:

```text
ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/scripts/run-stress-quick-sweep.sh
```

Then inspect and upload:

```text
reference/02-quick-stress-sweep-sqlite-report.md
```

### Code review instructions

Start with:

```text
design-doc/01-goja-site-stress-testing-breakdown-plan.md
scripts/run-stress-quick-sweep.sh
scripts/run-stress-hour-sweep.sh
scripts/render-stress-report-from-sqlite.py
```

Confirm that the quick sweep is short and that the hour sweep is not accidentally triggered by the quick script.

## Step 2: Run the quick stress sweep and identify the first breakdown signal

I ran the quick stress sweep. This was the intended 2-3 minute validation pass before any hour-scale experiment. The run completed the benchmark matrix and imported results into SQLite. Report rendering initially failed because the newly committed renderer script did not have the executable bit; I fixed the permissions and rendered the report manually from the imported SQLite data.

### Prompt Context

**User prompt (verbatim):** "ok, let's first run a single quick sterss test to see how hings behave."

**Assistant interpretation:** Run the quick stress script once, inspect how scenarios behave, generate the SQLite-backed report, and report the observed breakdown behavior.

**Inferred user intent:** Validate the stress setup with a short experiment and learn whether the proposed rates already show bending before running a long experiment.

**Commit (code before run):** `8691e5b8367058b176b463ea21620a5100972b21` — "docs: plan goja-site stress testing".

### What I ran

Command:

```text
bash ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/scripts/run-stress-quick-sweep.sh
```

Matrix ID:

```text
stress-quick-20260515T145900Z
```

Matrix shape:

```text
4 scenarios × 3 rates × 1 repeat = 12 runs
```

Scenarios:

```text
null
render
db-write
kanban-action
```

Rates:

```text
50/s
100/s
200/s
```

Per run:

```text
3s warmup
10s measured
```

Result root:

```text
bench/results/stress-quick-20260515T145900Z
```

SQLite DB:

```text
archive/stress-benchmarks.sqlite
```

Generated report:

```text
reference/02-quick-stress-sweep-sqlite-report.md
```

### What worked

- All 12 Vegeta runs completed.
- Every request returned HTTP 200.
- There were no Vegeta error sets.
- The SQLite import succeeded.
- The generated report embeds the SQL query used for every result table.
- The quick sweep found a clear breakdown candidate before the hour-scale run.

### What didn't work

The script reached the SQLite import step successfully, but failed to execute the report renderer:

```text
Permission denied: render-stress-report-from-sqlite.py
```

Cause: the renderer and run scripts were committed without executable file mode.

Fix:

```text
chmod +x render-stress-report-from-sqlite.py run-stress-quick-sweep.sh run-stress-hour-sweep.sh
```

Then I rendered the report manually:

```text
render-stress-report-from-sqlite.py --matrix-id stress-quick-20260515T145900Z --db archive/stress-benchmarks.sqlite --out reference/02-quick-stress-sweep-sqlite-report.md --title "Quick Stress Sweep SQLite Report"
```

### Headline results

Light scenarios did not bend at 200/s:

```text
null 200/s:      throughput ratio 1.000, p95 0.63 ms, max 1.08 ms
render 200/s:    throughput ratio 1.000, p95 2.38 ms, max 10.80 ms
db-write 200/s:  throughput ratio 1.000, p95 3.07 ms, max 10.22 ms
```

Kanban action bent hard between 50/s and 100/s:

```text
kanban-action 50/s:   throughput ratio 1.001, p95 27.02 ms, max 53.95 ms
kanban-action 100/s:  throughput ratio 0.862, p95 1527.67 ms, max 1607.18 ms
kanban-action 200/s:  throughput ratio 0.480, p95 10347.13 ms, max 10829.10 ms
```

The report's breakdown candidate query flags:

```text
kanban-action 100/s: throughput_shortfall
kanban-action 200/s: throughput_shortfall
```

The latency knee query shows:

```text
kanban-action 50/s -> 100/s: p95 growth factor 56.53
kanban-action 100/s -> 200/s: p95 growth factor 6.77
```

### Interpretation

The quick stress test did exactly what it was supposed to do: it found that `kanban-action` saturates well before 100/s on this machine/configuration, while `null`, `render`, and `db-write` remain healthy through 200/s for a 10-second measured run.

The breakdown mode is not HTTP failure. It is queueing/saturation:

```text
success remains 100%
errors remain empty
throughput falls below offered rate
latency explodes
```

That is consistent with a single hot Goja VM / owner event-loop path becoming saturated by the heavy Kanban action workload, which includes action dispatch plus full board refresh HTML generation.

### What I learned

- The hour-scale sweep should probably not include `kanban-action` at rates as high as 400/s or 800/s until we narrow the knee, because even 200/s created ~10-second p95 latency.
- The next useful stress test is not the broad hour sweep. It is a targeted `kanban-action` knee search around 50-100/s.
- Candidate rates for the targeted run:

```text
60/s,70/s,80/s,90/s,100/s
```

with repeats, and possibly pprof on the first clearly bending rate.

### What warrants a second pair of eyes

- Whether the breakdown thresholds should classify `kanban-action 100/s` as `latency_threshold` before `throughput_shortfall`; currently the CASE expression reports `throughput_shortfall` first.
- Whether `kanban-action` should be changed to return smaller refresh payloads or partial updates before running higher-rate stress.
- Whether stress scripts should have a guardrail that aborts remaining higher rates when p95 exceeds a large threshold.

### What should be done in the future

- Upload the quick stress report to reMarkable.
- Commit the SQLite DB/report and executable-bit fix.
- Add a targeted knee script or run an override matrix:

```text
BENCH_SCENARIOS=kanban-action BENCH_RATES=60/s,70/s,80/s,90/s,100/s BENCH_REPEAT=3 ...
```

- Rerun the first bending rate with pprof.

### Code review instructions

Review the generated report first:

```text
reference/02-quick-stress-sweep-sqlite-report.md
```

Then verify the raw result count:

```text
sqlite3 archive/stress-benchmarks.sqlite "select matrix_id, count(*) from benchmark_runs group by matrix_id;"
```

Expected:

```text
stress-quick-20260515T145900Z|12
```

## Step 3: Upload quick stress report to reMarkable

I uploaded the quick stress sweep report to reMarkable.

### Prompt Context

**User prompt:** The original GOJA-STRESS-TEST request included uploading the detailed report to reMarkable after running the quick stress test.

**Assistant interpretation:** Upload the generated SQL-backed report PDF, not the raw SQLite DB.

**Commit (code/results):** pending at time of this entry.

### What I did

Dry run:

```text
DRY: bundle name=GOJA STRESS TEST Quick Sweep Report
DRY: remote-dir=/ai/2026/05/15/GOJA-STRESS-TEST
DRY: upload GOJA_STRESS_TEST_Quick_Sweep_Report.pdf -> /ai/2026/05/15/GOJA-STRESS-TEST
```

Real upload:

```text
OK: uploaded GOJA_STRESS_TEST_Quick_Sweep_Report.pdf -> /ai/2026/05/15/GOJA-STRESS-TEST
```

Remote path:

```text
/ai/2026/05/15/GOJA-STRESS-TEST/GOJA_STRESS_TEST_Quick_Sweep_Report.pdf
```

### What worked

The Markdown report uploaded cleanly as a PDF bundle. The report includes the SQL queries used to generate each table.

### What should be done in the future

Before running the hour-scale sweep, add a targeted knee-search script around `kanban-action` 60-100/s or run the existing quick script with environment overrides.

## Step 4: Add targeted Kanban action knee-search script

After the quick stress sweep showed that `kanban-action` bends between 50/s and 100/s, I added a targeted knee-search script instead of immediately running the broad hour-scale sweep.

### Prompt Context

**User prompt (verbatim):** "go ahead"

**Assistant interpretation:** Continue with the recommended next stress step: targeted `kanban-action` knee search around 60-100/s.

**Inferred user intent:** Narrow down where the first observed breakdown begins before spending time on an hour-scale experiment.

**Commit (code):** pending at time of this entry.

### What I added

```text
scripts/run-stress-kanban-action-knee.sh
```

Default shape:

```text
scenario: kanban-action
rates:    60/s,70/s,80/s,90/s,100/s
repeat:   3
duration: 30s measured
warmup:   5s
```

The script:

1. builds `goja-site` once,
2. runs `scripts/bench-matrix.sh`,
3. imports results into `archive/stress-benchmarks.sqlite`,
4. renders `reference/03-kanban-action-knee-sqlite-report.md` with embedded SQL queries.

### Validation

I ran:

```text
bash -n run-stress-kanban-action-knee.sh
python3 -m py_compile render-stress-report-from-sqlite.py import-benchmark-matrix-to-sqlite.py
```

Both passed.

## Step 5: Run targeted Kanban action knee search and pprof at 80/s

I ran the targeted `kanban-action` knee-search matrix and then captured a pprof diagnostic run at 80/s, which the knee report identifies as the first clear latency-threshold breach.

### Prompt Context

**User prompt (verbatim):** "go ahead"

**Assistant interpretation:** Continue from the quick stress finding by narrowing the `kanban-action` knee and profiling the first bending rate.

**Inferred user intent:** Avoid the broad hour-scale sweep until the observed first breakdown mode is better understood.

**Commit (code before run):** `a214bf3c132f29ba3d3080abec4de9f999068595` — "chore: add kanban action knee stress script".

### What I ran for the knee search

Command:

```text
ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/scripts/run-stress-kanban-action-knee.sh
```

Matrix ID:

```text
stress-kanban-action-knee-20260515T150544Z
```

Shape:

```text
1 scenario × 5 rates × 3 repeats = 15 runs
```

Rates:

```text
60/s
70/s
80/s
90/s
100/s
```

Per run:

```text
5s warmup
30s measured
```

Report:

```text
reference/03-kanban-action-knee-sqlite-report.md
```

### Knee-search results

Aggregate results:

```text
60/s:  throughput ratio 1.000, p95 avg 88.63 ms,  p99 avg 140.97 ms, max 285.69 ms
70/s:  throughput ratio 0.999, p95 avg 57.32 ms,  p99 avg 114.19 ms, max 199.61 ms
80/s:  throughput ratio 0.996, p95 avg 616.79 ms, p99 avg 670.21 ms, max 1034.30 ms
90/s:  throughput ratio 0.980, p95 avg 838.29 ms, p99 avg 940.24 ms, max 1265.01 ms
100/s: throughput ratio 0.960, p95 avg 1746.78 ms, p99 avg 1789.84 ms, max 4040.89 ms
```

Breakdown candidate query flags:

```text
kanban-action 80/s:  latency_threshold
kanban-action 90/s:  latency_threshold
kanban-action 100/s: latency_threshold
```

Latency knee query shows:

```text
70/s -> 80/s: p95 growth factor 10.76
```

### Interpretation

The first clear knee is around 80/s. Up to 70/s, the system keeps up with offered throughput and p95 stays under ~100 ms on average. At 80/s, p95 jumps above 600 ms and max latency crosses 1 second, even though success remains 100%.

This confirms the breakdown mode is queueing/tail latency before HTTP failures.

### What I ran for pprof

Command:

```text
scripts/bench-vegeta.sh --scenario kanban-action --duration 30s --warmup-duration 5s --rate 80/s --port 18700 --metrics-port 19700 --out-dir ttmp/2026/05/15/GOJA-STRESS-TEST--stress-testing-breakdown-experiments-for-goja-site/archive/pprof-kanban-action-80rps-20260515T151521Z --pprof --pprof-seconds 20
```

Artifacts:

```text
archive/pprof-kanban-action-80rps-20260515T151521Z
```

I removed the raw `vegeta.bin` before committing because it was about 891 MiB due to large response bodies. The useful pprof artifacts and text summaries remain.

### pprof results

The pprof diagnostic run itself completed successfully:

```text
2400 requests
80/s target
80.00 throughput
100% success
p95 178.339 ms
p99 221.485 ms
max 251.72 ms
```

The CPU profile points to UI DSL rendering, precise move form generation, JSON encoding, and allocation/GC:

```text
github.com/go-go-golems/go-go-goja/modules/uidsl.renderNode       5.79s cumulative, 22.42%
github.com/go-go-golems/goja-site/pkg/kanbanddsl.(*Board).preciseMoveForm 5.27s cumulative, 20.41%
github.com/go-go-golems/go-go-goja/modules/uidsl.renderAttrs      4.62s cumulative, 17.89%
encoding/json.appendString                                        2.06s cumulative, 7.98%
runtime.gcDrain                                                   6.85s cumulative, 26.53%
runtime.mallocgc                                                  5.03s cumulative, 19.48%
```

### What worked

- The targeted knee search gave a clearer breakpoint than the broad quick sweep.
- The pprof run produced CPU/heap/allocs/goroutine artifacts in the ticket archive.
- The raw `vegeta.bin` was removed to avoid committing an 891 MiB artifact.

### What didn't work

The pprof run at 80/s was healthier than the worst 80/s knee-search repeats. It still showed elevated tail latency, but not a multi-second queue. That means the pprof is useful for identifying expensive code paths, but a future pprof at 90/s or 100/s may better capture the consistently degraded state.

### What I learned

- The knee is not a sharp deterministic cutoff; there is high variance around 80-100/s.
- The first consistent report-level threshold breach is 80/s.
- The expensive path is full board refresh HTML generation plus large JSON response construction.
- `preciseMoveForm` appears prominently because every card renders move controls/options.

### What warrants a second pair of eyes

- Whether `preciseMoveForm` should be lazy/optional in stress fixtures and production UI.
- Whether Kanban action responses should return patches instead of the whole board HTML.
- Whether pprof should be repeated at 90/s to capture a more consistently degraded run.

### What should be done in the future

- Upload the knee report and pprof report to reMarkable.
- Commit SQLite DB updates, knee report, and pprof artifacts.
- Consider a second pprof run at 90/s if optimization work begins.
- Do not run the broad hour sweep until the `kanban-action` high-rate range is adjusted or intentionally excluded.

### Code review instructions

Review:

```text
reference/03-kanban-action-knee-sqlite-report.md
reference/04-kanban-action-pprof-report.md
archive/pprof-kanban-action-80rps-20260515T151521Z/cpu-top.txt
archive/pprof-kanban-action-80rps-20260515T151521Z/heap-top.txt
```

Also inspect the SQLite rows:

```text
sqlite3 archive/stress-benchmarks.sqlite "select matrix_id, count(*) from benchmark_runs group by matrix_id order by matrix_id;"
```

## Step 6: Upload knee-search and pprof reports to reMarkable

Uploaded the targeted reports as one bundled PDF.

### Command

```text
remarquee upload bundle reference/03-kanban-action-knee-sqlite-report.md reference/04-kanban-action-pprof-report.md --name "GOJA STRESS TEST Kanban Knee and Pprof" --remote-dir "/ai/2026/05/15/GOJA-STRESS-TEST" --toc-depth 2 --non-interactive
```

### Result

```text
OK: uploaded GOJA_STRESS_TEST_Kanban_Knee_and_Pprof.pdf -> /ai/2026/05/15/GOJA-STRESS-TEST
```
