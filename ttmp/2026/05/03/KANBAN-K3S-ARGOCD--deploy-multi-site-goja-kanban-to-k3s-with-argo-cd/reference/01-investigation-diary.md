---
Title: Investigation Diary
Ticket: KANBAN-K3S-ARGOCD
Status: active
Topics:
    - deployment
    - kubernetes
    - argocd
    - kanban
    - dns
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Chronological notes for researching multi-site goja-site deployment to K3s/Argo CD."
LastUpdated: 2026-05-03T21:45:00-04:00
WhatFor: "Resume or review the KANBAN-K3S-ARGOCD deployment investigation."
WhenToUse: "Before implementing multi-site hosting, DNS changes, or GitOps deployment manifests."
---

# Investigation Diary

## Goal

Research what is needed to deploy `goja-site` to the Hetzner K3s cluster with Argo CD as a multi-site Kanban host.

The requested public shape is:

```text
XXX.kanban.yolo.scapegoat.dev
```

Each site should come from a different scripts folder and get its own SQLite database.

## Step 1: Created ticket workspace

Created docmgr ticket:

```text
KANBAN-K3S-ARGOCD — Deploy Multi Site Goja Kanban to K3s with Argo CD
```

Important docs:

- `design-doc/01-multi-site-goja-kanban-k3s-argo-cd-deployment-guide.md`
- `reference/01-investigation-diary.md`

## Step 2: Inspected current local goja-site shape

Files inspected:

- `cmd/goja-site/serve.go`
- `pkg/app/config.go`
- `pkg/app/server.go`
- `examples/kanban/scripts/app.js`

Findings:

- The current CLI supports one site per process:

```bash
goja-site serve --db ./app.db --scripts ./scripts --addr :8080
```

- `app.Config` has one DB path and one scripts dir.
- `app.NewServer` already creates a complete isolated app instance:
  - SQLite DB,
  - Goja runtime,
  - Express app host,
  - `ui.dsl`,
  - `kanban.dsl`,
  - `db.guard`.

Conclusion:

- Multi-site support should create one existing-style app instance per site.
- The new code should add an outer host router instead of mixing all sites into one runtime.

## Step 3: Inspected K3s GitOps patterns

Files inspected in `/home/manuel/code/wesen/2026-03-27--hetzner-k3s`:

- `gitops/applications/docs-yolo.yaml`
- `gitops/applications/hair-booking.yaml`
- `gitops/kustomize/docs-yolo/*`
- `gitops/kustomize/hair-booking/*`
- `docs/app-packaging-and-gitops-pr-standard.md`
- `docs/public-repo-ghcr-argocd-deployment-playbook.md`
- `docs/hetzner-k3s-server-setup.md`

Findings:

- `docs-yolo` is the best public lightweight app template:
  - Deployment,
  - Service,
  - Ingress,
  - PVC,
  - Argo CD Application.
- `hair-booking` is useful for private GHCR and persistent-data rollout lessons:
  - image pull secret pattern if image is private,
  - single-writer PVC behavior,
  - public `yolo.scapegoat.dev` Ingress,
  - `letsencrypt-prod` cert-manager issuer.
- New Argo CD Application files are not automatically live; they need one initial `kubectl apply`.

## Step 4: DNS conclusion

The existing documented DNS record is:

```text
*.yolo.scapegoat.dev -> 91.98.46.169
```

This is not enough for:

```text
trail.kanban.yolo.scapegoat.dev
```

Reason: DNS wildcards cover one label only.

Required new DNS record:

```text
*.kanban.yolo.scapegoat.dev A 91.98.46.169
```

TLS note:

- wildcard DNS does not automatically mean wildcard TLS,
- first deployment should use explicit Ingress hosts for known sites,
- wildcard TLS would require DNS-01 ACME support.

## Step 5: Wrote design guide

Wrote a long-form intern-friendly implementation guide covering:

- local code changes,
- multi-site config shape,
- host routing,
- DB isolation,
- static assets,
- Docker packaging,
- GHCR workflow,
- DNS requirements,
- Ingress/TLS design,
- Kustomize manifests,
- Argo CD bootstrap,
- validation commands,
- alternatives and open questions.

## Current recommendation

Implement in this order:

1. `serve-multi` locally.
2. Dockerfile and GHCR publish workflow.
3. DNS wildcard `*.kanban.yolo.scapegoat.dev`.
4. GitOps `goja-kanban` Kustomize package.
5. One-time Argo CD Application apply.
6. Public URL validation.

## Step 6: Corrected DNS ownership to Terraform repo

The user clarified that DNS updates should not be done manually in a DNS UI. They should happen through the adjacent Terraform repository:

```text
/home/manuel/code/wesen/terraform
```

I inspected:

- `/home/manuel/code/wesen/terraform/dns/README.md`
- `/home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf`

Finding:

- the DigitalOcean `scapegoat.dev` zone is managed in `dns/zones/scapegoat-dev/envs/prod`,
- records are declared in `local.base_records`,
- existing K3s records include:
  - `k3s` -> `91.98.46.169`,
  - `*.yolo` -> `91.98.46.169`.

Therefore the Kanban nested wildcard should be implemented as a Terraform record similar to:

```hcl
wildcard_kanban_yolo_a = {
  type  = "A"
  name  = "*.kanban.yolo"
  value = "91.98.46.169"
  ttl   = 3600
}
```

The implementation task list was updated so DNS work happens through Terraform, not manual DNS edits.

## Step 7: Implemented local multi-site server and packaging scaffold

I implemented the local app-side foundation for the deployment plan.

### Code changes

- Added `pkg/app/multi_config.go`:
  - `MultiConfig`,
  - `SiteConfig`,
  - YAML/JSON config loading,
  - host and DB path derivation,
  - duplicate host/name validation.
- Added `pkg/app/multi_server.go`:
  - one outer HTTP server,
  - host-header dispatch,
  - `/healthz` and `/readyz`,
  - one isolated `app.Server` per site.
- Updated `pkg/app/server.go`:
  - exposed `Handler()` and `ServeHTTP(...)`.
- Added `cmd/goja-site/serve_multi.go` and wired it into `cmd/goja-site/main.go`.

### Site and packaging changes

- Added `sites/trail/scripts/app.js` based on the existing Field Notes Kanban app.
- Added placeholder sites:
  - `sites/editorial/scripts/app.js`,
  - `sites/crm/scripts/app.js`.
- Added:
  - `deploy/sites.yaml`,
  - `deploy/sites.local.yaml`,
  - `Dockerfile`,
  - `.dockerignore`,
  - `.github/workflows/publish-image.yaml`.

### Validation

Ran:

```bash
go test ./... -count=1
GOTOOLCHAIN=go1.26.2 go run ./cmd/goja-site serve-multi --config deploy/sites.local.yaml
curl -H 'Host: trail.kanban.yolo.scapegoat.dev' http://127.0.0.1:60131/
curl -H 'Host: editorial.kanban.yolo.scapegoat.dev' http://127.0.0.1:60131/
curl -H 'Host: crm.kanban.yolo.scapegoat.dev' http://127.0.0.1:60131/
curl http://127.0.0.1:60131/healthz
```

All host routes responded. The Trail site produced the expected Field Notes board; the other two placeholders responded with their titles.

### Notes

The Trail app still logs expected duplicate-column errors during migration because its JavaScript migration deliberately catches older demo DB alteration failures. This is pre-existing behavior from the database module logging exec errors before JS catches them.
