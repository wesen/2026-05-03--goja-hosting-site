# Changelog

## 2026-05-03

- Initial workspace created


## 2026-05-03

Created multi-site Goja Kanban K3s/Argo CD deployment research guide, including local code changes, DNS requirement, packaging plan, and GitOps manifests.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/design-doc/01-multi-site-goja-kanban-k3s-argo-cd-deployment-guide.md — Main intern-friendly design and implementation guide
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/reference/01-investigation-diary.md — Chronological research notes
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/tasks.md — Implementation task checklist


## 2026-05-03

Uploaded KANBAN-K3S-ARGOCD multi-site deployment guide bundle to reMarkable.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/changelog.md — Recorded reMarkable upload
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/tasks.md — Marked reMarkable upload complete


## 2026-05-03

Corrected DNS plan to use the Terraform-managed DigitalOcean scapegoat.dev zone in ../terraform.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/design-doc/01-multi-site-goja-kanban-k3s-argo-cd-deployment-guide.md — Updated DNS implementation instructions
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/tasks.md — Expanded DNS Terraform tasks
- /home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf — Terraform location for the new *.kanban.yolo DNS record


## 2026-05-03

Implemented local serve-multi host router, multi-site config, packaged site layout, Dockerfile, and GHCR workflow scaffold.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/Dockerfile — Container packaging scaffold
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/cmd/goja-site/serve_multi.go — serve-multi CLI command
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/deploy/sites.yaml — Production multi-site config
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/app/multi_config.go — Multi-site config loader and validation
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/app/multi_server.go — Host-routed multi-site HTTP server


## 2026-05-03

Applied Terraform DNS wildcard for *.kanban.yolo.scapegoat.dev and validated resolution.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/reference/01-investigation-diary.md — Recorded Terraform plan/apply and DNS validation
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/tasks.md — Marked DNS Terraform tasks complete
- /home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf — Added wildcard_kanban_yolo_a


## 2026-05-03

Added K3s GitOps manifests for goja-kanban Argo CD deployment and validated kustomize rendering.

### Related Files

- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/goja-kanban.yaml — New Argo CD Application
- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/goja-kanban — Kustomize package for namespace
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/reference/01-investigation-diary.md — Recorded GitOps implementation


## 2026-05-03

Fixed GHCR workflow and Dockerfile to satisfy the local go-go-goja replace path in CI/container builds.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/.github/workflows/publish-image.yaml — Clones go-go-goja before tests
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/Dockerfile — Clones go-go-goja before go mod download


## 2026-05-03

Resolved Argo CD PVC sync-wave deadlock and validated goja-kanban as Synced/Healthy with public site responses.

### Related Files

- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/goja-kanban/deployment.yaml — Deployment moved to PVC sync wave
- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/goja-kanban/service.yaml — Service moved to same sync wave
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/reference/01-investigation-diary.md — Recorded deployment hang root cause and validation


## 2026-05-03

Updated Hetzner K3s runbooks to document the local-path PVC WaitForFirstConsumer sync-wave trap.

### Related Files

- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/app-packaging-and-gitops-pr-standard.md — Added single-writer PVC app category and same-wave rule
- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/operator-troubleshooting-faq.md — Added troubleshooting entry for Pending PVC with no Deployment
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/reference/01-investigation-diary.md — Recorded runbook update


## 2026-05-03

Implemented HEAD fallback and replaced placeholder editorial/CRM sites with real Kanban apps.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/web/host.go — HEAD falls back to GET with body suppression
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/web/host_integration_test.go — HEAD fallback integration test
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/sites/crm/scripts/app.js — Real CRM Kanban app
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/sites/editorial/scripts/app.js — Real editorial Kanban app


## 2026-05-03

Published sha-5fdc211 image, updated K3s GitOps deployment, and validated production GET/HEAD for trail, editorial, and CRM sites.

### Related Files

- /home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/goja-kanban/deployment.yaml — Updated production image to sha-5fdc211
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/reference/01-investigation-diary.md — Recorded image workflow and production validation
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/tasks.md — Marked production deploy validation tasks complete


## 2026-05-03

Added automatic GitOps PR workflow wiring using infra-tooling and Vault-backed GitHub Actions OIDC.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/.github/workflows/publish-image.yaml — Reusable publish and GitOps PR workflow
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/deploy/gitops-targets.json — goja-kanban-prod GitOps target
- /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf — Vault role and policy for goja hosting GitOps PR token


## 2026-05-03

Validated automatic app publish to K3s GitOps PR workflow; GitHub Actions opened K3s PR #70.

### Related Files

- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/.github/workflows/publish-image.yaml — Reusable workflow opened GitOps PR
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/deploy/gitops-targets.json — Target used for K3s PR
- /home/manuel/code/wesen/2026-05-03--goja-hosting-site/ttmp/2026/05/03/KANBAN-K3S-ARGOCD--deploy-multi-site-goja-kanban-to-k3s-with-argo-cd/reference/01-investigation-diary.md — Recorded PR #70 validation

