# Changelog

## 2026-05-15

- Initial workspace created


## 2026-05-15

Created multi-VM serve-multi stress plan and initial harness scripts.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/scripts/02-run-multi-vm-quick-sweep.sh — quick validation sweep script


## 2026-05-15

Ran quick multi-VM serve-multi validation sweep; all 11 runs succeeded with HTTP 200 only.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/reference/02-multi-vm-quick-sweep-report.md — quick multi-VM stress report


## 2026-05-15

Added higher-rate multi-VM saturation sweep script for null and kanban-fragment even-hot runs.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/scripts/04-run-multi-vm-saturation-sweep.sh — saturation sweep script


## 2026-05-15

Ran higher-rate multi-VM saturation sweep; null stayed healthy through 2000/s, kanban-fragment saturated around 200-400/s depending on VM count.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/reference/03-multi-vm-saturation-sweep-report.md — saturation sweep report


## 2026-05-15

Fixed multi-VM pprof capture so CPU profiling overlaps the measured Vegeta attack.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/15/GOJA-MULTI-VM-STRESS--multi-vm-serve-multi-stress-testing-for-goja-site/scripts/01-run-multi-vm-vegeta.sh — pprof timing fix

