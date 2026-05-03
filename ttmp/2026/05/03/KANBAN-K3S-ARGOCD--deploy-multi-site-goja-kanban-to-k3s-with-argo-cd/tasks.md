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

- [x] Add `sites/<name>/scripts` production layout.
- [x] Add `deploy/sites.yaml` multi-site config.
- [x] Add `MultiConfig` and `SiteConfig` types.
- [x] Add YAML/JSON config loader.
- [x] Add `app.MultiServer` with host-based routing.
- [x] Expose `Server.Handler()` or `Server.ServeHTTP` for reuse by host router.
- [x] Add `goja-site serve-multi --config <path>` CLI command.
- [x] Add Go-owned `/healthz` and `/readyz` endpoints in multi-site server.
- [x] Add tests for duplicate hosts, unknown hosts, DB isolation, and health checks.
- [x] Update production site scripts to use image paths such as `/app/sites/<name>/assets`.

## Packaging plan

- [x] Add Dockerfile for CGO SQLite build and Debian runtime image.
- [x] Add GitHub Actions workflow to test, build, and publish GHCR image.
- [x] Decide whether GHCR package is public or private. Initial manifests assume public GHCR; add image pull secret later if package remains private.
- [ ] If private, copy the `hair-booking` Vault/VSO image pull secret pattern.
- [x] Pin deployments to immutable `sha-<commit>` image tags.

## DNS and TLS plan

- [x] Locate the Terraform-owned DigitalOcean DNS zone in `../terraform`.
- [x] Confirm existing `*.yolo.scapegoat.dev` record is managed in Terraform.
- [x] Confirm current K3s public IP.
- [x] Add Terraform record `*.kanban.yolo.scapegoat.dev A <k3s-ip>` in `../terraform/dns/zones/scapegoat-dev/envs/prod/main.tf`.
- [x] Run `terraform -chdir=dns/zones/scapegoat-dev/envs/prod fmt`.
- [x] Run `terraform -chdir=dns/zones/scapegoat-dev/envs/prod plan`.
- [x] Apply the Terraform DNS change, if approved.
- [x] Validate site hostnames with `dig`.
- [ ] Use explicit Ingress hosts for first deployment.
- [ ] Revisit DNS-01 wildcard certificate support only if dynamic site creation is needed.

## GitOps / Argo CD plan

- [x] Add `gitops/kustomize/goja-kanban` package in the K3s repo.
- [x] Add namespace, PVC, ConfigMap, Deployment, Service, and Ingress manifests.
- [x] Use `replicas: 1` and `strategy: Recreate` for SQLite safety.
- [x] Add `gitops/applications/goja-kanban.yaml`.
- [x] Run `kubectl kustomize gitops/kustomize/goja-kanban`.
- [x] Commit GitOps changes.
- [x] Push GitOps changes.
- [x] One-time apply Argo CD Application with `kubectl apply -f gitops/applications/goja-kanban.yaml`.
- [x] Validate Argo CD sync, pod health, PVC, ingress, cert-manager certificate, and public URLs.


## Production app content

- [x] Implement HEAD fallback for dynamic Goja routes.
- [x] Add a real Editorial Pipeline Kanban app.
- [x] Add a real CRM Pipeline Kanban app.
- [x] Validate GET and HEAD locally for trail, editorial, and CRM hosts.
- [ ] Publish updated image through GitHub Actions.
- [ ] Update K3s GitOps image tag to the published SHA.
- [ ] Validate updated production GET and HEAD behavior.
