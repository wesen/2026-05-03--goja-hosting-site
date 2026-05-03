# Changelog

## 2026-05-03

- Initial workspace created


## 2026-05-03

Implemented local Pizza Ops example with readable split scripts, stacked boards, no precise-move controls, dependency-aware kitchen workflow, delivery board, tally, and Done-column grouping by pizza order.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/deploy/sites.local.yaml — Local config includes pizza
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/deploy/sites.yaml — Production config includes pizza
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/sites/pizza/README.md — Pizza app README
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/sites/pizza/scripts — Pizza Ops split JavaScript app


## 2026-05-03

Refactored Pizza Ops into repository/workflow/view/route layers and validated automatic cooking/quality delivery transitions from kitchen task state.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/sites/pizza/scripts/02_repository.js — New persistence layer
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/sites/pizza/scripts/03_workflow.js — New workflow layer and automatic delivery transitions
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/sites/pizza/scripts/04_views.js — Views now use workflow facade

