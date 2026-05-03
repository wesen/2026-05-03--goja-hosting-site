---
Title: Multi Site Goja Kanban K3s Argo CD Deployment Guide
Ticket: KANBAN-K3S-ARGOCD
Status: active
Topics:
    - deployment
    - kubernetes
    - argocd
    - kanban
    - dns
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../2026-03-27--hetzner-k3s/docs/app-packaging-and-gitops-pr-standard.md
      Note: Packaging and GitOps PR standard
    - Path: ../../../../../../../2026-03-27--hetzner-k3s/docs/hetzner-k3s-server-setup.md
      Note: DNS wildcard and letsencrypt-prod cluster issuer reference
    - Path: ../../../../../../../2026-03-27--hetzner-k3s/docs/public-repo-ghcr-argocd-deployment-playbook.md
      Note: GHCR to Argo CD deployment playbook
    - Path: ../../../../../../../2026-03-27--hetzner-k3s/gitops/applications/docs-yolo.yaml
      Note: Reference Argo CD Application for lightweight public app
    - Path: ../../../../../../../2026-03-27--hetzner-k3s/gitops/applications/hair-booking.yaml
      Note: Reference Argo CD app with namespace creation
    - Path: ../../../../../../../2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml
      Note: Reference Deployment/PVC mounted app pattern
    - Path: ../../../../../../../2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/ingress.yaml
      Note: Reference Traefik/cert-manager Ingress for yolo.scapegoat.dev
    - Path: ../../../../../../../2026-03-27--hetzner-k3s/gitops/kustomize/hair-booking/deployment.yaml
      Note: Reference persistent app rollout and GHCR image pattern
    - Path: cmd/goja-site/serve.go
      Note: Current single-site serve command to extend with serve-multi
    - Path: examples/kanban/scripts/app.js
      Note: Existing Kanban app to adapt into packaged site scripts
    - Path: pkg/app/config.go
      Note: Current single-site Config shape; needs MultiConfig/SiteConfig
    - Path: pkg/app/server.go
      Note: Current Server constructs one DB
ExternalSources: []
Summary: Design for deploying one goja-site image to K3s/Argo CD as a multi-site Kanban host with one SQLite DB per site and subdomains under *.kanban.yolo.scapegoat.dev.
LastUpdated: 2026-05-03T21:45:00-04:00
WhatFor: Use when implementing or reviewing the multi-site goja-site deployment to the Hetzner K3s cluster.
WhenToUse: Before changing local goja-site code, DNS, GitHub Actions packaging, or the Argo CD GitOps repo for hosted Kanban sites.
---


# Multi Site Goja Kanban K3s Argo CD Deployment Guide

## Executive Summary

We want to deploy the `goja-site` Kanban hosting server to the Hetzner K3s cluster that lives in:

- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s`

The desired public shape is:

```text
https://trail.kanban.yolo.scapegoat.dev
https://editorial.kanban.yolo.scapegoat.dev
https://crm.kanban.yolo.scapegoat.dev
...
```

Each hostname should run a separate trusted Goja JavaScript app from a site-specific scripts directory, and each site should get its own SQLite database file. The deployment should be GitOps-managed by Argo CD, following the same patterns that were recently used for `docs-yolo` and `hair-booking`.

The recommended design is:

- add a local multi-site mode to `goja-site`, probably as `goja-site serve-multi`,
- store site definitions in a small YAML/JSON config,
- create one Goja runtime, route host, and SQLite DB per site,
- use one outer HTTP server that dispatches by request `Host`,
- package the binary and example/site scripts into an OCI image pushed to GHCR,
- deploy one Kubernetes `Deployment` with `replicas: 1`,
- mount one persistent volume at `/data`,
- store DBs as `/data/sites/<site>/app.db`,
- route all declared hostnames through one Kubernetes `Ingress`,
- create a new Argo CD `Application` under `gitops/applications/goja-kanban.yaml`,
- add a Kustomize package under `gitops/kustomize/goja-kanban`,
- commit/push the app repo and GitOps repo changes, then one-time apply the Argo CD `Application` object.

Important DNS finding: the existing DNS wildcard documented in the K3s repo is:

```text
*.yolo.scapegoat.dev -> 91.98.46.169
```

That wildcard matches `kanban.yolo.scapegoat.dev`, but it does **not** match `trail.kanban.yolo.scapegoat.dev`, because DNS wildcards only cover one label. Therefore the desired `XXX.kanban.yolo.scapegoat.dev` shape needs one additional DNS wildcard record:

```text
*.kanban.yolo.scapegoat.dev -> 91.98.46.169
```

TLS has a second subtlety: a wildcard DNS record is not the same as a wildcard TLS certificate. The simplest first deployment should generate explicit Ingress hosts for the known site list and let cert-manager issue a normal multi-SAN certificate through HTTP-01. A future dynamic/self-service site system would need either DNS-01 wildcard certificate support or an operator/controller that materializes new Ingress hosts.

## Problem Statement

The current `goja-site` binary serves exactly one site per process. The existing CLI shape is:

```bash
goja-site serve \
  --db examples/kanban/kanban.db \
  --scripts examples/kanban/scripts \
  --addr :8080
```

The current server config is similarly single-site:

```go
type Config struct {
    Addr       string
    DBPath     string
    ScriptsDir string
    Dev        bool
}
```

That is good for local development, but it does not directly express the production shape requested here:

- many sites,
- many script directories,
- one SQLite DB per site,
- one subdomain per site,
- one Kubernetes deployment managed by Argo CD.

A naive deployment could create one Kubernetes `Deployment` per site. That would work, but it would be noisy and repetitive for small trusted Kanban apps. Since `goja-site` already treats each app as a Goja runtime plus route host plus DB, the cleaner local change is to make the Go process able to host multiple independent app instances behind one HTTP listener.

The intern implementing this should understand four separate layers:

1. **Local application code**: how `goja-site` starts runtimes and serves requests.
2. **Container packaging**: how the binary and scripts get into a GHCR image.
3. **DNS and TLS**: how hostnames resolve and how cert-manager obtains certificates.
4. **GitOps / Argo CD**: how Kubernetes desired state is committed and reconciled.

## Current State Research

### Local `goja-site` server

Relevant files in this repository:

- `cmd/goja-site/serve.go`
- `pkg/app/config.go`
- `pkg/app/server.go`
- `pkg/web/host.go`
- `pkg/web/route_registry.go`
- `pkg/uidsl/*`
- `pkg/kanbanddsl/*`
- `pkg/dbguard/*`
- `examples/kanban/scripts/app.js`

The current server startup path is:

```text
CLI serve command
  -> app.NewServer(Config{Addr, DBPath, ScriptsDir, Dev})
    -> open one SQLite DB
    -> create one web.Host
    -> create one go-go-goja runtime
    -> register database/db, express, ui.dsl, kanban.dsl, db.guard
    -> load all .js files from ScriptsDir
    -> return Server
  -> Server.Run(ctx)
    -> http.ListenAndServe(Addr, host)
```

The important implementation detail is that `app.NewServer` already creates an isolated app world:

- one SQLite handle,
- one metered `db.guard` wrapper,
- one Goja runtime,
- one Express-style route registry,
- one `ui.dsl` renderer,
- one `kanban.dsl` module registration.

So multi-site support should not mix sites inside one runtime. It should create one existing `app.Server`-like instance per site and route by hostname.

### K3s / GitOps repo state

Relevant repo:

- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s`

Relevant Argo CD app examples:

- `gitops/applications/docs-yolo.yaml`
- `gitops/applications/hair-booking.yaml`

Relevant Kustomize examples:

- `gitops/kustomize/docs-yolo/*`
- `gitops/kustomize/hair-booking/*`

`docs-yolo` is the closest lightweight public app pattern:

- one namespace,
- one deployment,
- one service,
- one ingress,
- one PVC,
- public GHCR image,
- Argo CD `Application` points at `gitops/kustomize/docs-yolo`.

`hair-booking` is relevant for private-image and persistent-data operational details:

- uses `ghcr.io/wesen/hair-booking:sha-...`,
- has a data PVC,
- uses an image pull secret through Vault/VSO,
- sets `replicas: 1`,
- uses rollout settings that avoid two pods fighting over one RWO volume,
- has a normal Traefik Ingress with `cert-manager.io/cluster-issuer: letsencrypt-prod`.

The cluster docs explicitly say:

```text
k3s.scapegoat.dev -> 91.98.46.169
*.yolo.scapegoat.dev -> 91.98.46.169
```

That existing wildcard is good for first-level app names like:

```text
docs.yolo.scapegoat.dev
hair-booking.yolo.scapegoat.dev
```

It is not sufficient for second-level app names like:

```text
trail.kanban.yolo.scapegoat.dev
```

because `*.yolo.scapegoat.dev` only matches one label before `yolo`.

## Proposed Solution

### Target architecture

```mermaid
flowchart TD
    Browser[Browser] --> DNS[DNS: trail.kanban.yolo.scapegoat.dev]
    DNS --> Traefik[Traefik Ingress on K3s]
    Traefik --> Svc[Kubernetes Service goja-kanban]
    Svc --> Pod[goja-kanban Pod]

    Pod --> Router[Host Router]
    Router --> Trail[Site instance: trail]
    Router --> Editorial[Site instance: editorial]
    Router --> CRM[Site instance: crm]

    Trail --> TrailRT[Goja runtime]
    Trail --> TrailDB[/data/sites/trail/app.db]
    Trail --> TrailScripts[/app/sites/trail/scripts]

    Editorial --> EditorialRT[Goja runtime]
    Editorial --> EditorialDB[/data/sites/editorial/app.db]
    Editorial --> EditorialScripts[/app/sites/editorial/scripts]

    CRM --> CRMRT[Goja runtime]
    CRM --> CRMDB[/data/sites/crm/app.db]
    CRM --> CRMScripts[/app/sites/crm/scripts]
```

Each site remains a normal `goja-site` app internally. The only new part is the outer host router.

### Site directory layout

Use a site layout that is explicit and easy to package:

```text
goja-site repo/
  sites/
    trail/
      scripts/
        app.js
      assets/
        trail-map.png
    editorial/
      scripts/
        app.js
    crm/
      scripts/
        app.js
  deploy/
    sites.yaml
```

Example config:

```yaml
addr: ":8080"
dataDir: "/data/sites"
baseDomain: "kanban.yolo.scapegoat.dev"
dev: false
sites:
  - name: trail
    host: trail.kanban.yolo.scapegoat.dev
    scriptsDir: /app/sites/trail/scripts
    dbPath: /data/sites/trail/app.db
  - name: editorial
    host: editorial.kanban.yolo.scapegoat.dev
    scriptsDir: /app/sites/editorial/scripts
    dbPath: /data/sites/editorial/app.db
  - name: crm
    host: crm.kanban.yolo.scapegoat.dev
    scriptsDir: /app/sites/crm/scripts
    dbPath: /data/sites/crm/app.db
```

The config should allow `dbPath` to be omitted. If omitted, compute it as:

```text
<dataDir>/<site.name>/app.db
```

The config should allow `host` to be omitted. If omitted, compute it as:

```text
<site.name>.<baseDomain>
```

This gives concise production config while still allowing local overrides.

### Local code changes needed

Add multi-site app types:

```go
type MultiConfig struct {
    Addr       string
    DataDir    string
    BaseDomain string
    Dev        bool
    Sites      []SiteConfig
}

type SiteConfig struct {
    Name       string
    Host       string
    ScriptsDir string
    DBPath     string
}
```

Add a new server type:

```go
type MultiServer struct {
    cfg     MultiConfig
    sites   map[string]*Server // keyed by normalized host
    httpSrv *http.Server
}
```

The constructor should reuse the current single-site constructor:

```go
func NewMultiServer(cfg MultiConfig) (*MultiServer, error) {
    normalize cfg
    for each site in cfg.Sites:
        host := normalizeHost(site.Host or site.Name + "." + cfg.BaseDomain)
        dbPath := site.DBPath or filepath.Join(cfg.DataDir, site.Name, "app.db")
        srv, err := NewServer(Config{
            Addr: "", // not used by inner server
            DBPath: dbPath,
            ScriptsDir: site.ScriptsDir,
            Dev: cfg.Dev,
        })
        if err != nil { close already-created sites; return err }
        sites[host] = srv
    return &MultiServer{cfg: cfg, sites: sites}, nil
}
```

The outer `ServeHTTP` should route by hostname:

```go
func (m *MultiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    host := normalizeRequestHost(r.Host)
    site := m.sites[host]
    if site == nil {
        http.Error(w, "unknown goja-site host", http.StatusNotFound)
        return
    }
    site.ServeHTTP(w, r)
}
```

This requires exposing the inner site handler. The cleanest local refactor is:

```go
func (s *Server) Handler() http.Handler {
    return s.host
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.host.ServeHTTP(w, r)
}
```

`Server.Run` can stay exactly as it is for single-site development.

Add CLI command:

```bash
goja-site serve-multi --config /etc/goja-site/sites.yaml
```

The command should:

- read YAML or JSON config,
- validate duplicate hosts,
- validate duplicate site names,
- create DB directories,
- fail fast if any scripts directory is missing,
- log the host-to-site map at startup.

Pseudocode:

```go
func (c *serveMultiCommand) Run(ctx context.Context, vals *values.Values) error {
    cfgPath := vals.GetString("config")
    cfg := readMultiConfig(cfgPath)
    srv, err := app.NewMultiServer(cfg)
    if err != nil { return err }
    defer srv.Close(context.Background())
    return srv.Run(ctx)
}
```

### Static assets in multi-site mode

The current example uses:

```js
app.static("/assets", "examples/kanban/assets");
```

That path is workstation-specific and should not be used in the production image. Each packaged site should use an image path, for example:

```js
app.static("/assets", "/app/sites/trail/assets");
```

For reusable apps, provide a small site-local config module or environment variable later. For the first deployment, explicit absolute paths in each site script are acceptable because the scripts are packaged together with the image.

### Health checks

The current example app does not appear to define a Kubernetes-friendly health endpoint. Add either:

1. A Go-owned outer health endpoint in `MultiServer`, recommended:

```text
GET /healthz
```

This should answer `200 OK` before host routing if the process is up and all site instances loaded.

2. Or require each script app to define `/healthz`.

Use the Go-owned endpoint so Kubernetes probes do not depend on a particular site hostname.

Pseudocode:

```go
func (m *MultiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok\n"))
        return
    }
    // then route by host
}
```

### SQLite and persistence rules

Use one PVC and one DB file per site:

```text
/data/sites/trail/app.db
/data/sites/editorial/app.db
/data/sites/crm/app.db
```

Why one PVC instead of one PVC per site?

- simpler Kustomize package,
- simple backup target,
- low operational overhead,
- good enough for small Kanban apps.

Why one DB per site instead of one shared DB?

- hard isolation,
- easy backup/restore per site,
- `db.guard` limits apply per site,
- no accidental cross-site SQL queries,
- no need to add a `site_id` column to every app schema.

Kubernetes rules:

- `replicas: 1`,
- PVC access mode `ReadWriteOnce`,
- deployment strategy should avoid two pods mounting/writing the same SQLite DBs during rollout.

Recommended strategy copied from the lessons in `hair-booking`:

```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 0
    maxUnavailable: 1
```

`Recreate` is also acceptable and even simpler for SQLite:

```yaml
strategy:
  type: Recreate
```

For this deployment, prefer `Recreate` unless zero-downtime rollout is more important than SQLite safety. This is a small Kanban host; safety and clarity matter more.

## DNS Design

### Existing DNS is not enough for nested subdomains

Existing documented record:

```text
*.yolo.scapegoat.dev -> 91.98.46.169
```

Requested hostnames:

```text
XXX.kanban.yolo.scapegoat.dev
```

DNS wildcard matching is label-based:

- `*.yolo.scapegoat.dev` matches `docs.yolo.scapegoat.dev`,
- `*.yolo.scapegoat.dev` matches `kanban.yolo.scapegoat.dev`,
- `*.yolo.scapegoat.dev` does **not** match `trail.kanban.yolo.scapegoat.dev`.

Needed DNS record:

```text
*.kanban.yolo.scapegoat.dev A 91.98.46.169
```

This record must be added through the Terraform DNS repository, not manually in the DigitalOcean UI. The DNS owner is:

```text
/home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf
```

The existing `*.yolo.scapegoat.dev` record is represented there as a `local.base_records` entry named `wildcard_yolo_a`. Add a sibling record such as:

```hcl
wildcard_kanban_yolo_a = {
  type  = "A"
  name  = "*.kanban.yolo"
  value = "91.98.46.169"
  ttl   = 3600
}
```

Then run:

```bash
cd /home/manuel/code/wesen/terraform
terraform -chdir=dns/zones/scapegoat-dev/envs/prod fmt
terraform -chdir=dns/zones/scapegoat-dev/envs/prod plan
# apply only after review/approval
terraform -chdir=dns/zones/scapegoat-dev/envs/prod apply
```

If IPv6 is used later, add the corresponding AAAA record too.

### DNS validation commands

Before deploying Ingress, validate:

```bash
dig +short trail.kanban.yolo.scapegoat.dev A
dig +short editorial.kanban.yolo.scapegoat.dev A
dig +short crm.kanban.yolo.scapegoat.dev A
```

Expected output should include the K3s public IP:

```text
91.98.46.169
```

If DNS does not resolve, cert-manager HTTP-01 validation will fail.

## TLS and Ingress Design

### First deployment: explicit hosts

Use explicit known site hosts in the Ingress:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: goja-kanban
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: traefik
  tls:
    - hosts:
        - trail.kanban.yolo.scapegoat.dev
        - editorial.kanban.yolo.scapegoat.dev
        - crm.kanban.yolo.scapegoat.dev
      secretName: goja-kanban-tls
  rules:
    - host: trail.kanban.yolo.scapegoat.dev
      http: &httpRules
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: goja-kanban
                port:
                  number: 80
    - host: editorial.kanban.yolo.scapegoat.dev
      http: *httpRules
    - host: crm.kanban.yolo.scapegoat.dev
      http: *httpRules
```

This allows cert-manager to request a normal SAN certificate. It does not require DNS-01 wildcard certificate support.

### Future deployment: wildcard TLS

A wildcard Ingress host like this is attractive:

```yaml
rules:
  - host: "*.kanban.yolo.scapegoat.dev"
```

But the TLS certificate for:

```text
*.kanban.yolo.scapegoat.dev
```

requires ACME DNS-01 validation. HTTP-01 cannot prove ownership of wildcard names. Only use this if the cluster's cert-manager issuer is extended with DNS provider credentials for `scapegoat.dev`.

## Container Packaging

### Add a Dockerfile to the app repo

This repository currently has no Dockerfile. Add one that builds the Go binary and packages the sites.

Example:

```dockerfile
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go test ./...
RUN CGO_ENABLED=1 go build -o /out/goja-site ./cmd/goja-site

FROM debian:bookworm-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/goja-site /usr/local/bin/goja-site
COPY sites /app/sites
COPY deploy/sites.yaml /etc/goja-site/sites.yaml
RUN useradd -r -u 10001 -g root goja-site \
 && mkdir -p /data/sites \
 && chown -R 10001:0 /data
USER 10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/goja-site"]
CMD ["serve-multi", "--config", "/etc/goja-site/sites.yaml"]
```

SQLite through `github.com/mattn/go-sqlite3` uses CGO, so do not use a pure static `scratch` image unless you deliberately handle CGO/static linking. `debian:bookworm-slim` keeps this simple.

### GitHub Actions image publishing

Follow the pattern documented in the K3s repo:

- `docs/app-packaging-and-gitops-pr-standard.md`
- `docs/public-repo-ghcr-argocd-deployment-playbook.md`

Add:

```text
.github/workflows/publish-image.yaml
```

Minimum workflow behavior:

- on pull request: run `go test ./...` and build image without pushing,
- on push to `main`: run tests, build, push to GHCR,
- tags:
  - `sha-<git-sha>`
  - `main`
  - maybe `latest`.

Image name:

```text
ghcr.io/wesen/2026-05-03--goja-hosting-site:sha-<commit>
```

If the GitHub package is public, no image pull secret is needed. If it is private, copy the `hair-booking` VSO image pull secret pattern.

## Kubernetes / Argo CD Design

Create a new GitOps package:

```text
/home/manuel/code/wesen/2026-03-27--hetzner-k3s/
  gitops/
    applications/
      goja-kanban.yaml
    kustomize/
      goja-kanban/
        namespace.yaml
        pvc.yaml
        configmap.yaml
        deployment.yaml
        service.yaml
        ingress.yaml
        kustomization.yaml
```

### `gitops/applications/goja-kanban.yaml`

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: goja-kanban
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  destination:
    server: https://kubernetes.default.svc
    namespace: goja-kanban
  source:
    repoURL: https://github.com/wesen/2026-03-27--hetzner-k3s.git
    targetRevision: main
    path: gitops/kustomize/goja-kanban
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - ServerSideApply=true
```

### `namespace.yaml`

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: goja-kanban
  labels:
    app.kubernetes.io/name: goja-kanban
```

### `pvc.yaml`

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: goja-kanban-data
  annotations:
    argocd.argoproj.io/sync-wave: "1"
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: local-path
  resources:
    requests:
      storage: 10Gi
```

### `configmap.yaml`

Use this if you do **not** bake `/etc/goja-site/sites.yaml` into the image, or if you want GitOps to own the site list independently from the image:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: goja-kanban-sites
  annotations:
    argocd.argoproj.io/sync-wave: "1"
data:
  sites.yaml: |
    addr: ":8080"
    dataDir: "/data/sites"
    baseDomain: "kanban.yolo.scapegoat.dev"
    dev: false
    sites:
      - name: trail
        scriptsDir: /app/sites/trail/scripts
      - name: editorial
        scriptsDir: /app/sites/editorial/scripts
      - name: crm
        scriptsDir: /app/sites/crm/scripts
```

Recommendation: GitOps should own the site list, while the image owns the actual script contents. This means adding/removing a site requires both an image change and a GitOps config change, which is acceptable for this first phase.

### `deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goja-kanban
  annotations:
    argocd.argoproj.io/sync-wave: "2"
  labels:
    app.kubernetes.io/name: goja-kanban
    app.kubernetes.io/component: app
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app.kubernetes.io/name: goja-kanban
      app.kubernetes.io/component: app
  template:
    metadata:
      labels:
        app.kubernetes.io/name: goja-kanban
        app.kubernetes.io/component: app
    spec:
      enableServiceLinks: false
      securityContext:
        fsGroup: 0
      containers:
        - name: goja-kanban
          image: ghcr.io/wesen/2026-05-03--goja-hosting-site:sha-REPLACE_ME
          imagePullPolicy: IfNotPresent
          args:
            - serve-multi
            - --config
            - /etc/goja-site/sites.yaml
          ports:
            - containerPort: 8080
              name: http
          readinessProbe:
            httpGet:
              path: /readyz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 15
            periodSeconds: 20
          volumeMounts:
            - name: data
              mountPath: /data
            - name: site-config
              mountPath: /etc/goja-site
              readOnly: true
          resources:
            requests:
              cpu: 50m
              memory: 128Mi
            limits:
              memory: 512Mi
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: goja-kanban-data
        - name: site-config
          configMap:
            name: goja-kanban-sites
```

### `service.yaml`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: goja-kanban
  annotations:
    argocd.argoproj.io/sync-wave: "2"
  labels:
    app.kubernetes.io/name: goja-kanban
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: goja-kanban
    app.kubernetes.io/component: app
  ports:
    - name: http
      port: 80
      targetPort: http
```

### `ingress.yaml`

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: goja-kanban
  annotations:
    argocd.argoproj.io/sync-wave: "3"
    cert-manager.io/cluster-issuer: letsencrypt-prod
  labels:
    app.kubernetes.io/name: goja-kanban
spec:
  ingressClassName: traefik
  tls:
    - hosts:
        - trail.kanban.yolo.scapegoat.dev
        - editorial.kanban.yolo.scapegoat.dev
        - crm.kanban.yolo.scapegoat.dev
      secretName: goja-kanban-tls
  rules:
    - host: trail.kanban.yolo.scapegoat.dev
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: goja-kanban
                port:
                  number: 80
    - host: editorial.kanban.yolo.scapegoat.dev
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: goja-kanban
                port:
                  number: 80
    - host: crm.kanban.yolo.scapegoat.dev
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: goja-kanban
                port:
                  number: 80
```

### `kustomization.yaml`

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: goja-kanban
resources:
  - namespace.yaml
  - pvc.yaml
  - configmap.yaml
  - deployment.yaml
  - service.yaml
  - ingress.yaml
labels:
  - pairs:
      app.kubernetes.io/name: goja-kanban
      app.kubernetes.io/part-of: goja-site
```

## Deployment Flow

### Phase 1: local app repo implementation

1. Add `sites/` layout.
2. Copy/adapt current `examples/kanban/scripts/app.js` into one or more sites.
3. Change static paths from workstation paths to image paths.
4. Add multi-site config file under `deploy/sites.yaml`.
5. Add `app.MultiServer` and host routing.
6. Add `goja-site serve-multi --config ...`.
7. Add health endpoints.
8. Add tests:
   - duplicate host rejected,
   - unknown host returns 404,
   - each host sees different DB data,
   - `/healthz` works without Host-specific route,
   - single-site `serve` still works.
9. Add Dockerfile.
10. Add GitHub Actions publish workflow.

Validation commands:

```bash
go test ./... -count=1
go run ./cmd/goja-site serve-multi --config deploy/sites.local.yaml
curl -H 'Host: trail.kanban.yolo.scapegoat.dev' http://127.0.0.1:8080/
curl -H 'Host: editorial.kanban.yolo.scapegoat.dev' http://127.0.0.1:8080/
curl http://127.0.0.1:8080/healthz
```

### Phase 2: DNS through Terraform

Do not add this record manually. Edit the DigitalOcean zone Terraform in:

```text
/home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf
```

Add to `local.base_records`:

```hcl
wildcard_kanban_yolo_a = {
  type  = "A"
  name  = "*.kanban.yolo"
  value = "91.98.46.169"
  ttl   = 3600
}
```

Then run `terraform fmt`, `terraform plan`, and only then `terraform apply` if approved.

Validate:

```bash
dig +short trail.kanban.yolo.scapegoat.dev A
dig +short editorial.kanban.yolo.scapegoat.dev A
```

### Phase 3: GitHub image publication

After the Dockerfile and workflow exist:

```bash
git push origin main
```

Then check the package:

```bash
gh run list --workflow publish-image.yaml
gh run view --log
```

The image tag to use in GitOps should be:

```text
ghcr.io/wesen/2026-05-03--goja-hosting-site:sha-<commit>
```

### Phase 4: GitOps repo changes

In `/home/manuel/code/wesen/2026-03-27--hetzner-k3s`:

```bash
mkdir -p gitops/kustomize/goja-kanban
$EDITOR gitops/kustomize/goja-kanban/*.yaml
$EDITOR gitops/applications/goja-kanban.yaml
kubectl kustomize gitops/kustomize/goja-kanban
```

Commit and push:

```bash
git status --short
git add gitops/applications/goja-kanban.yaml gitops/kustomize/goja-kanban
git commit -m "Add goja kanban Argo CD deployment"
git push origin main
```

### Phase 5: one-time Argo CD bootstrap

The K3s docs note that new `gitops/applications/*.yaml` files are not automatically live unless something applies them. This repo does not currently use an app-of-apps layer that auto-creates every new Application.

Run once:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
export KUBECONFIG=$PWD/kubeconfig-91.98.46.169.yaml
kubectl apply -f gitops/applications/goja-kanban.yaml
kubectl -n argocd annotate application goja-kanban argocd.argoproj.io/refresh=hard --overwrite
```

Then watch:

```bash
kubectl -n argocd get application goja-kanban
kubectl -n goja-kanban get pods,svc,ingress,pvc
kubectl -n goja-kanban describe certificate goja-kanban-tls
```

### Phase 6: public validation

```bash
curl -I https://trail.kanban.yolo.scapegoat.dev/
curl -I https://editorial.kanban.yolo.scapegoat.dev/
curl -fsS https://trail.kanban.yolo.scapegoat.dev/healthz
```

Certificate check:

```bash
openssl s_client -connect trail.kanban.yolo.scapegoat.dev:443 \
  -servername trail.kanban.yolo.scapegoat.dev </dev/null 2>/dev/null | \
  openssl x509 -noout -subject -issuer -ext subjectAltName
```

Expected issuer should be Let's Encrypt, not `TRAEFIK DEFAULT CERT`.

## Design Decisions

### Decision 1: one process, many site instances

Use one process with many isolated app instances rather than one process per site.

Rationale:

- keeps Kubernetes manifests small,
- shares one binary and one pod,
- still isolates runtimes and DBs,
- matches the requested “multiple sites in the scripts folder” mental model.

### Decision 2: one DB file per site

Use separate SQLite files instead of a shared DB with `site_id` columns.

Rationale:

- easier backup and deletion,
- easier size guard limits,
- fewer app schema requirements,
- no accidental cross-site queries.

### Decision 3: explicit Ingress hosts first

Do not start with wildcard TLS.

Rationale:

- existing cert-manager issuer is known as `letsencrypt-prod`,
- existing app ingresses use HTTP-01 style host certs,
- wildcard TLS requires DNS-01 provider credentials,
- explicit hosts are enough for the first known site list.

### Decision 4: add DNS wildcard under `kanban`

Add `*.kanban.yolo.scapegoat.dev`.

Rationale:

- existing `*.yolo.scapegoat.dev` does not cover nested site hostnames,
- adding the nested wildcard keeps future site additions DNS-free,
- cert-manager still needs explicit TLS hosts unless DNS-01 wildcard TLS is added.

### Decision 5: use a PVC and `replicas: 1`

SQLite files live on a local-path PVC and should not be written by multiple pods.

Rationale:

- simple persistence,
- easy backup target,
- safe for single-node K3s,
- follows `docs-yolo` PVC and `hair-booking` single-writer lessons.

## Alternatives Considered

### Alternative A: one Deployment per site

This is operationally straightforward:

```text
trail Deployment -> trail DB PVC -> trail Ingress
editorial Deployment -> editorial DB PVC -> editorial Ingress
crm Deployment -> crm DB PVC -> crm Ingress
```

Rejected for first implementation because it creates a lot of YAML for small trusted apps and does not use the natural multi-site host-router shape.

Keep this option in mind if sites later need different resource limits, different images, different secrets, or independent rollout schedules.

### Alternative B: one runtime and one DB for all sites

This would use a single Goja runtime and route based on `Host` inside JavaScript.

Rejected because it weakens isolation and forces every app schema to remember site scoping. We already learned from session scoping that implicit scoping can be easy to forget. Separate DB files are cleaner.

### Alternative C: wildcard Ingress and wildcard certificate immediately

Attractive for dynamic sites, but likely requires cert-manager DNS-01 work. The current cluster patterns use explicit hosts with `cert-manager.io/cluster-issuer: letsencrypt-prod`. Do not expand certificate infrastructure unless dynamic site creation is required.

### Alternative D: scripts as ConfigMaps

Kubernetes ConfigMaps could hold app scripts.

Rejected for primary packaging because:

- scripts may include assets,
- ConfigMaps have size limits,
- app repo image should be the immutable application artifact,
- reviewing JS changes in the app repo is cleaner.

ConfigMaps are still useful for the site list.

## Intern Implementation Checklist

### Local repo checklist

- [ ] Add `sites/<site>/scripts` layout.
- [ ] Add `deploy/sites.yaml` and `deploy/sites.local.yaml`.
- [ ] Add `MultiConfig` and `SiteConfig`.
- [ ] Add config loader for YAML/JSON.
- [ ] Add `MultiServer` host router.
- [ ] Expose `Server.Handler()` or `Server.ServeHTTP`.
- [ ] Add `serve-multi` CLI command.
- [ ] Add `/healthz` and `/readyz` at the outer server level.
- [ ] Add tests for routing, DB isolation, and health.
- [ ] Add Dockerfile.
- [ ] Add GitHub Actions image publish workflow.

### DNS checklist

- [ ] Confirm K3s public IP is still `91.98.46.169`.
- [ ] Add `*.kanban.yolo.scapegoat.dev A 91.98.46.169` through `/home/manuel/code/wesen/terraform/dns/zones/scapegoat-dev/envs/prod/main.tf`.
- [ ] Validate with `dig`.

### GitOps checklist

- [ ] Add `gitops/kustomize/goja-kanban` package.
- [ ] Add `gitops/applications/goja-kanban.yaml`.
- [ ] Use `cert-manager.io/cluster-issuer: letsencrypt-prod`.
- [ ] Use `replicas: 1` and `strategy: Recreate`.
- [ ] Use one PVC at `/data`.
- [ ] Pin GHCR image to `sha-<commit>`.
- [ ] Commit and push GitOps changes.
- [ ] One-time apply the Argo CD Application.
- [ ] Validate Argo sync, pod health, ingress, certificate, and public URLs.

## File References

Current app repo:

- `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/cmd/goja-site/serve.go`
- `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/app/config.go`
- `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/app/server.go`
- `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/web/host.go`
- `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/kanbanddsl/*`
- `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/pkg/dbguard/*`
- `/home/manuel/code/wesen/2026-05-03--goja-hosting-site/examples/kanban/scripts/app.js`

K3s / GitOps repo:

- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/docs-yolo.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/hair-booking.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/deployment.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/pvc.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/docs-yolo/ingress.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/hair-booking/deployment.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/hair-booking/ingress.yaml`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/app-packaging-and-gitops-pr-standard.md`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/public-repo-ghcr-argocd-deployment-playbook.md`
- `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/hetzner-k3s-server-setup.md`

## Open Questions

1. Should the first public site list be `trail`, `editorial`, and `crm`, or should production start with only the existing Field Notes board?
2. Should the app repo image be public on GHCR, or should we copy the `hair-booking` private image pull secret pattern?
3. Should site scripts be baked into the image only, or should the site list and scripts eventually be loaded from a Git-backed or object-storage-backed source?
4. Do we want auth before public deployment? The current Kanban example is session-isolated but not authenticated.
5. Do we want a backup CronJob for `/data/sites` immediately, or is existing node/PVC backup strategy enough for phase one?

## References

- K3s repo: `/home/manuel/code/wesen/2026-03-27--hetzner-k3s`
- App repo: `/home/manuel/code/wesen/2026-05-03--goja-hosting-site`
- Existing DNS note: `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/hetzner-k3s-server-setup.md`
- GitOps packaging standard: `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/app-packaging-and-gitops-pr-standard.md`
- Public GHCR deployment playbook: `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/public-repo-ghcr-argocd-deployment-playbook.md`
