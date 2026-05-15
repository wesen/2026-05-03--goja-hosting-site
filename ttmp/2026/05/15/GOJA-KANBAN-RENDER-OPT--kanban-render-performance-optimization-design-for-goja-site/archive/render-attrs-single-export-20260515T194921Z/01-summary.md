---
Title: render-attrs-1000 single-export validation summary
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - benchmarking
    - profiling
DocType: reference
Intent: historical
Summary: "Raw Vegeta summary for render-attrs-1000 after ui.dsl single-export optimization."
---

# goja-site Vegeta benchmark

- Scenario: render-attrs-1000
- Duration: 30s
- Warmup duration: 5s
- Rate: 100/s
- Base URL: http://127.0.0.1:28082
- Metrics URL: http://127.0.0.1:29092/metrics
- pprof capture: 1
- pprof seconds: 10
- OpenTelemetry enabled: 0
- OpenTelemetry endpoint: http://127.0.0.1:4318/v1/traces
- OpenTelemetry sample ratio: 0.01
- Commit: 37b3ca80827d267f1ae568df724183c6266be667
- Dirty worktree: true
- Go version: go version go1.26.2 linux/amd64
- Vegeta: Version: 

## Report

```text
Requests      [total, rate, throughput]         3000, 100.03, 59.70
Duration      [total, attack, wait]             50.247s, 29.991s, 20.257s
Latencies     [min, mean, 50, 90, 95, 99, max]  12.775ms, 10.171s, 9.853s, 18.043s, 19.337s, 20.114s, 20.257s
Bytes In      [total, mean]                     750120000, 250040.00
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
