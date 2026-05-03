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

