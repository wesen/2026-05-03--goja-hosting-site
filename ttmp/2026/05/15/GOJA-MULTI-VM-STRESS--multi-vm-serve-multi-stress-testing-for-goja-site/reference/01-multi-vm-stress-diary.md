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
