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


## 2026-05-03

Deployed Pizza Ops to pizza.kanban.yolo.scapegoat.dev on image sha-7c9289f and validated DNS, TLS, Argo CD health, public page, and APIs.

### Related Files

- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/goja-kanban/configmap.yaml — Pizza site added to multi-site config
- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/goja-kanban/ingress.yaml — Pizza host added to public ingress and TLS
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/PIZZA-KITCHEN-SITE--pizza-restaurant-workflow-site-with-builder-kitchen-dependencies-delivery-and-tally/reference/01-diary.md — Production deployment evidence

