---
Title: Pprof run summary - kanban-fragment-500-100_s
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - benchmarking
    - profiling
DocType: reference
Intent: historical
Summary: "Raw Vegeta summary for kanban-fragment-500-100_s pprof run."
---

# goja-site Vegeta benchmark

- Scenario: kanban-fragment-500
- Duration: 30s
- Warmup duration: 5s
- Rate: 100/s
- Base URL: http://127.0.0.1:26081
- Metrics URL: http://127.0.0.1:27091/metrics
- pprof capture: 1
- pprof seconds: 10
- OpenTelemetry enabled: 0
- OpenTelemetry endpoint: http://127.0.0.1:4318/v1/traces
- OpenTelemetry sample ratio: 0.01
- Commit: 4013051f64f3d76e22ead76ebd4075fb41a1206a
- Dirty worktree: true
- Go version: go version go1.26.2 linux/amd64
- Vegeta: Version: 

## Report

```text
Requests      [total, rate, throughput]         3000, 100.03, 100.01
Duration      [total, attack, wait]             29.998s, 29.99s, 7.548ms
Latencies     [min, mean, 50, 90, 95, 99, max]  5.168ms, 24.551ms, 17.385ms, 55.656ms, 66.472ms, 82.266ms, 107.746ms
Bytes In      [total, mean]                     754380000, 251460.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:3000  
Error Set:
```

## Artifacts

- Raw results: vegeta.bin
- JSON report: vegeta.json
- Text report: vegeta.txt
- Run metadata: metadata.json
- Metrics before: metrics-before.prom
- Metrics after: metrics-after.prom
- Metrics delta: metrics-delta.txt
- Server log: server.log
- Targets: targets.txt
- CPU profile: cpu.pprof (when --pprof is set)
- Heap profile: heap.pprof (when --pprof is set)
- Goroutine profile: goroutine.txt (when --pprof is set)
