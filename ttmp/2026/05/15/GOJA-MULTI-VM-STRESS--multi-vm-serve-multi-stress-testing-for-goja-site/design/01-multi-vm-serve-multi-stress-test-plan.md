---
Title: Multi-VM serve-multi stress test plan
Ticket: GOJA-MULTI-VM-STRESS
Status: active
Topics:
    - benchmarking
    - stress-testing
    - goja
    - observability
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/app/multi_config.go
      Note: multi-site config format used by generated stress configs
    - Path: pkg/app/multi_server.go
      Note: serve-multi dispatch and per-site Server creation
    - Path: ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/scripts/01-run-multi-vm-vegeta.sh
      Note: single-run multi-VM stress harness
ExternalSources: []
Summary: Plan and runbook for stressing multiple Goja VMs via serve-multi site configs and Host-header load generation.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Multi-VM serve-multi stress test plan

This ticket measures `goja-site serve-multi` as a multi-VM host. In the current architecture, `serve-multi` creates one `Server` per configured site. Each `Server` owns its own Goja runtime and VM, route host, database path, and loaded JavaScript scripts. Therefore a config with `N` sites is a practical way to start `N` Goja VMs inside one process.

## What this ticket measures

The first measurements are not a transparent VM pool behind one logical host. They are Host-header dispatched sites inside one process:

```text
site-001.multi-vm.bench.test -> Server -> Goja VM 1
site-002.multi-vm.bench.test -> Server -> Goja VM 2
site-003.multi-vm.bench.test -> Server -> Goja VM 3
```

Vegeta selects the target VM by sending the corresponding `Host:` header.

## Experiment classes

### even-hot

Every configured VM receives traffic. This measures process-wide throughput across multiple hot VMs.

### one-hot

Only the first VM receives traffic. The remaining VMs are loaded and idle. This measures whether idle VMs impose memory, GC, dispatch, or process overhead on one hot site.

### skewed

The first VM receives most target entries, and every other VM receives one target entry. This approximates unequal site popularity.

## Initial quick sweep

The first quick sweep intentionally avoids the known expensive `kanban-action` path. It uses:

```text
null even-hot:             vm_count 1,2,4,8 at 200/s
null one-hot:              vm_count 2,4,8 at 200/s
kanban-fragment even-hot:  vm_count 1,2,4,8 at 50/s
```

Each run uses:

```text
3s warmup
10s measured
```

This is a validation sweep. It should answer whether the generated configs, Host headers, metrics, and result summaries work before longer stress experiments.

## Scripts

```text
scripts/01-run-multi-vm-vegeta.sh
scripts/02-run-multi-vm-quick-sweep.sh
scripts/03-render-multi-vm-summary.py
```

All generated raw run directories live under `bench/results/...` by default and are ignored by Git. Durable reports and diaries live in this ticket.

## Metrics to inspect

The existing observability spine already exposes the relevant signals:

```text
goja_site_hosts_configured
goja_site_site_up
goja_site_multi_dispatch_duration_seconds
goja_site_unknown_host_requests_total
goja_site_http_requests_total{site=...}
goja_site_http_request_duration_seconds{site=...}
goja_site_kanban_fragment_duration_seconds{site=...}
process_resident_memory_bytes
go_memstats_heap_alloc_bytes
go_goroutines
```

## Interpretation rules

If `even-hot` throughput degrades as VM count rises, inspect CPU, GC, and per-site route latency. If `one-hot` throughput degrades as idle VM count rises, inspect memory and GC pressure. If unknown-host counters rise, the target file or config has a Host-header mismatch.

Do not interpret these results as a VM pool result. A VM pool behind one host would require new routing and session-affinity code.
