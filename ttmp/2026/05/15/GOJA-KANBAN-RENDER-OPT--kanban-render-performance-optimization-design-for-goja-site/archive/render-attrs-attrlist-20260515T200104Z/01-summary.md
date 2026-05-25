---
Title: render-attrs-1000 attr-list validation summary
Ticket: GOJA-KANBAN-RENDER-OPT
Status: active
Topics:
    - benchmarking
    - profiling
DocType: reference
Intent: historical
Summary: "Raw Vegeta summary for render-attrs-1000 after ui.dsl Attr list cutover."
---

# goja-site Vegeta benchmark

- Scenario: render-attrs-1000
- Duration: 30s
- Warmup duration: 5s
- Rate: 100/s
- Base URL: http://127.0.0.1:28083
- Metrics URL: http://127.0.0.1:29093/metrics
- pprof capture: 1
- pprof seconds: 10
- OpenTelemetry enabled: 0
- OpenTelemetry endpoint: http://127.0.0.1:4318/v1/traces
- OpenTelemetry sample ratio: 0.01
- Commit: 27eb3293acec1ca87740bda64074b2c9b35b75a3
- Dirty worktree: true
- Go version: go version go1.26.2 linux/amd64
- Vegeta: Version: 

## Report

```text
Requests      [total, rate, throughput]         3000, 100.03, 71.79
Duration      [total, attack, wait]             41.789s, 29.991s, 11.799s
Latencies     [min, mean, 50, 90, 95, 99, max]  8.274ms, 5.59s, 5.271s, 10.758s, 11.222s, 11.714s, 11.799s
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
