# Tasks

## Research and documentation

- [x] Create `KANBAN-K3S-ARGOCD` docmgr ticket.
- [x] Inspect current `goja-site` single-site server architecture.
- [x] Inspect Hetzner K3s Argo CD app patterns.
- [x] Compare `docs-yolo` and `hair-booking` deployment approaches.
- [x] Determine whether existing DNS wildcard covers `XXX.kanban.yolo.scapegoat.dev`.
- [x] Write intern-friendly design and implementation guide.
- [x] Keep investigation diary.
- [x] Upload ticket bundle to reMarkable.

## Local goja-site implementation plan

- [ ] Add `sites/<name>/scripts` production layout.
- [ ] Add `deploy/sites.yaml` multi-site config.
- [ ] Add `MultiConfig` and `SiteConfig` types.
- [ ] Add YAML/JSON config loader.
- [ ] Add `app.MultiServer` with host-based routing.
- [ ] Expose `Server.Handler()` or `Server.ServeHTTP` for reuse by host router.
- [ ] Add `goja-site serve-multi --config <path>` CLI command.
- [ ] Add Go-owned `/healthz` and `/readyz` endpoints in multi-site server.
- [ ] Add tests for duplicate hosts, unknown hosts, DB isolation, and health checks.
- [ ] Update production site scripts to use image paths such as `/app/sites/<name>/assets`.

## Packaging plan

- [ ] Add Dockerfile for CGO SQLite build and Debian runtime image.
- [ ] Add GitHub Actions workflow to test, build, and publish GHCR image.
- [ ] Decide whether GHCR package is public or private.
- [ ] If private, copy the `hair-booking` Vault/VSO image pull secret pattern.
- [ ] Pin deployments to immutable `sha-<commit>` image tags.

## DNS and TLS plan

- [ ] Confirm current K3s public IP.
- [ ] Add DNS record `*.kanban.yolo.scapegoat.dev A <k3s-ip>`.
- [ ] Validate site hostnames with `dig`.
- [ ] Use explicit Ingress hosts for first deployment.
- [ ] Revisit DNS-01 wildcard certificate support only if dynamic site creation is needed.

## GitOps / Argo CD plan

- [ ] Add `gitops/kustomize/goja-kanban` package in the K3s repo.
- [ ] Add namespace, PVC, ConfigMap, Deployment, Service, and Ingress manifests.
- [ ] Use `replicas: 1` and `strategy: Recreate` for SQLite safety.
- [ ] Add `gitops/applications/goja-kanban.yaml`.
- [ ] Run `kubectl kustomize gitops/kustomize/goja-kanban`.
- [ ] Commit and push GitOps changes.
- [ ] One-time apply Argo CD Application with `kubectl apply -f gitops/applications/goja-kanban.yaml`.
- [ ] Validate Argo CD sync, pod health, PVC, ingress, cert-manager certificate, and public URLs.
