---
Title: Multi-VM serve-multi stress diary
Ticket: GOJA-MULTI-VM-STRESS
Status: active
Topics:
    - benchmarking
    - stress-testing
    - goja
    - observability
DocType: reference
Intent: historical
Summary: "Chronological diary for multi-VM serve-multi stress testing work."
---

# Multi-VM serve-multi stress diary

## Step 1: Ticket setup and harness design

Created ticket `GOJA-MULTI-VM-STRESS` to isolate multi-VM `serve-multi` stress work from the earlier single-site/single-VM benchmark and stress tickets.

### Prompt context

The user asked to create a `serve-multi` setup, start stress testing the system, create a new docmgr ticket, keep a diary, and save all scripts in the ticket `scripts/` folder.

### Design decision

Use `serve-multi` as the first multi-VM mechanism. A single `serve-multi` process creates one `Server` per configured site, and each `Server` creates its own Goja runtime/VM. Multiple VMs running the same script are modeled as multiple sites with the same `scripts:` directory and different Host headers.

This is not a VM pool behind one logical host. It is a multi-site/multi-VM stress model.

### Scripts added

```text
scripts/01-run-multi-vm-vegeta.sh
scripts/02-run-multi-vm-quick-sweep.sh
scripts/03-render-multi-vm-summary.py
```

The single-run script generates:

- a temporary `serve-multi` YAML config,
- one site per requested VM,
- Host-header Vegeta targets,
- before/after Prometheus metric snapshots,
- metric deltas,
- Vegeta JSON/text output,
- metadata for the run,
- optional pprof artifacts.

The quick sweep script builds `goja-site` once and runs a small matrix:

```text
null even-hot:             1,2,4,8 VMs at 200/s
null one-hot:              2,4,8 VMs at 200/s
kanban-fragment even-hot:  1,2,4,8 VMs at 50/s
```

### Validation pending

Next steps are shell/Python validation, `docmgr doctor`, commit setup, then run the quick sweep.

## Step 2: Run quick multi-VM validation sweep

Ran the first `serve-multi` multi-VM stress validation sweep.

### Command

```text
ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/scripts/02-run-multi-vm-quick-sweep.sh
```

### Matrix ID

```text
multi-vm-quick-20260515T163935Z
```

### Shape

```text
null even-hot:             1,2,4,8 VMs at 200/s
null one-hot:              2,4,8 VMs at 200/s
kanban-fragment even-hot:  1,2,4,8 VMs at 50/s
```

Every run used 3s warmup and 10s measured load.

### Result root

```text
bench/results/multi-vm-quick-20260515T163935Z
```

### Report

```text
reference/02-multi-vm-quick-sweep-report.md
```

### Findings

All 11 runs completed with 100% success, HTTP 200 only, and no Vegeta error sets.

`null` at 200/s stayed below 1 ms p95 for 1, 2, 4, and 8 VMs in even-hot distribution. The one-hot idle-VM check also stayed below 1 ms p95 with 2, 4, and 8 loaded VMs.

`kanban-fragment` at 50/s stayed healthy across 1, 2, 4, and 8 VMs. p95 ranged from 16.95 ms to 20.31 ms.

### Interpretation

This validates the `serve-multi` generated config approach, Host-header target generation, metrics capture, and result rendering. It does not yet find a multi-VM saturation point. The next useful experiment is a higher-rate sweep for `null` and `kanban-fragment`, plus a carefully bounded `kanban-action` sweep below or around the known single-VM knee.

## Step 3: Add higher-rate saturation sweep script

After the quick validation sweep passed, I added a higher-rate saturation sweep to look for an inflection point under `serve-multi`.

### Script

```text
scripts/04-run-multi-vm-saturation-sweep.sh
```

### Default shape

```text
null even-hot:
  vm_count: 1,2,4,8
  rates:    400/s,800/s,1200/s,2000/s

kanban-fragment even-hot:
  vm_count: 1,2,4,8
  rates:    100/s,200/s,400/s
```

Each run defaults to 3s warmup and 10s measured duration. This is intentionally still a bounded sweep, but it is high enough to find whether the cheap route or Kanban fragment route bends in the tested range.

I intentionally left `kanban-action` out of this first multi-VM saturation script because the single-VM ticket already showed an action-refresh knee around 80/s. `kanban-fragment` is safer for first multi-VM saturation because it stresses real render work without action refresh state mutation.

## Step 4: Run higher-rate multi-VM saturation sweep

Ran the higher-rate `serve-multi` saturation sweep.

### Command

```text
ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/scripts/04-run-multi-vm-saturation-sweep.sh
```

### Matrix ID

```text
multi-vm-saturation-20260515T164805Z
```

### Shape

```text
null even-hot:
  vm_count: 1,2,4,8
  rates:    400/s,800/s,1200/s,2000/s

kanban-fragment even-hot:
  vm_count: 1,2,4,8
  rates:    100/s,200/s,400/s
```

Each run used 3s warmup and 10s measured load.

### Result root

```text
bench/results/multi-vm-saturation-20260515T164805Z
```

### Report

```text
reference/03-multi-vm-saturation-sweep-report.md
```

### Result summary

All 28 runs returned HTTP 200 only, 100% Vegeta success, and no Vegeta error sets.

The minimal `null` route did not saturate through 2000/s total offered rate for 1, 2, 4, or 8 VMs. p95 stayed below 0.6 ms in every `null` cell.

The `kanban-fragment` route did show a clear inflection point. At 100/s total, all VM counts stayed usable. At 200/s total, 1 VM saturated badly, 2 VMs showed strong queueing, 4 VMs stayed much healthier, and 8 VMs showed high p95 despite keeping throughput near target. At 400/s total, every VM count saturated heavily.

Important cells:

```text
1 VM, 100/s: p95 58.91 ms, throughput 100.04/s
1 VM, 200/s: p95 5901.66 ms, throughput 124.64/s
2 VM, 200/s: p95 720.15 ms, throughput 183.80/s
4 VM, 200/s: p95 179.41 ms, throughput 199.87/s
8 VM, 200/s: p95 564.94 ms, throughput 198.41/s
4 VM, 400/s: p95 5861.77 ms, throughput 246.53/s
8 VM, 400/s: p95 5559.25 ms, throughput 249.44/s
```

### Interpretation

The multi-VM model improves the Kanban fragment capacity compared with one VM, but it does not scale linearly. The 400/s rows cluster around roughly 240-250 achieved requests per second with multi-second p95 latency. That suggests a process-level or shared-resource ceiling for this fixture, not just per-VM owner-loop serialization.

This fixture returns about 246 KB per request. At high rates, the system is paying for Goja route execution, UI DSL rendering, allocation/GC, and response writing. The next diagnostic step should be pprof on a degraded multi-VM cell, probably `kanban-fragment` 4 VMs or 8 VMs at 400/s.

## Step 5: Fix multi-VM pprof timing

The first pprof attempt for `kanban-fragment` 4 VMs at 400/s produced a CPU profile with zero samples. The reason was script timing: `01-run-multi-vm-vegeta.sh` captured `/debug/pprof/profile` after the Vegeta attack had completed, so the process was mostly idle while the CPU profile was collected.

I fixed the harness so CPU profiling starts just before the measured Vegeta attack and overlaps the load window. Heap, allocs, and goroutine snapshots still run after the attack.

The unusable zero-sample profile directory was left in the archive for evidence, but the raw `vegeta.bin` was removed to avoid keeping a ~940 MiB artifact.
