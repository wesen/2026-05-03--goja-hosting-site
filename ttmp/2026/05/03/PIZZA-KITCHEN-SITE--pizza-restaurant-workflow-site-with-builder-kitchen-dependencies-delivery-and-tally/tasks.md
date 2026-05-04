# Tasks

## TODO

- [ ] Add tasks here



## Implementation

- [x] Create readable split-script Pizza Ops app.
- [x] Add pizza builder form with `ui.dsl`.
- [x] Add kitchen Kanban board with dependency gating.
- [x] Add delivery Kanban board with kitchen-completion gating.
- [x] Add paid/tips tally and JSON APIs.
- [x] Add local and multi-site config entries for pizza.
- [x] Remove generated precise-move controls from pizza cards.
- [x] Stack kitchen and delivery boards vertically.
- [x] Collapse Done kitchen tasks by pizza order into summary cards.
- [x] Validate locally with single-site and multi-site servers.
- [x] Deploy to `pizza.kanban.yolo.scapegoat.dev`.

- [x] Split persistence into `Pizza.repo` and workflow rules into `Pizza.store`.
- [x] Automatically move orders to cooking when kitchen work starts.
- [x] Automatically move orders to quality when all kitchen tasks are done.
- [x] Validate automatic delivery transitions locally with a stable cookie jar.
- [x] Publish image `sha-7c9289f`.
- [x] Add Pizza site to K3s ConfigMap and Ingress.
- [x] Validate DNS, TLS certificate, Argo CD health, and public Pizza URL.