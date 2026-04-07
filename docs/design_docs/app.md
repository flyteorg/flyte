# App Service Design Doc

**Status:** Draft
**Author:** Kevin Su
**Date:** 2026-04-06

---

## 1. Overview

This document specifies the design for implementing the **App Service** in OSS Flyte v2. The App Service enables users to deploy and manage long-running applications (serving endpoints, LLM inference, APIs, dashboards) alongside Flyte workflow executions.

### Goals

- Deploy long-running containerized apps (FastAPI, Flask, vLLM, Streamlit, etc.) via the Flyte control plane
- Autoscale apps based on request rate or concurrency
- Provide ingress (public URLs) for deployed apps
- Manage app lifecycle: create, update, stop
- Expose replica-level observability (logs, status, pod info)
- Fit naturally into the existing v2 architecture (buf connect, GORM, SetupContext pattern)

### Non-Goals (v1 scope)

- Multi-cluster scheduling and lease-based assignment (single-cluster first)
- VPC/CNAME custom domain routing
- Connector app config map injection
- Artifact-triggered app deployments

---

## 2. Architecture Context — Flyte v2 Today

Flyte v2 uses a modular service architecture:

```
┌─────────────────────────────┐      ┌─────────────────────────────────┐
│  Control Plane               │      │  Data Plane                      │
│                              │      │                                  │
│  ┌─────────────────────┐    │      │  ┌─────────────────────────┐    │
│  │ RunService           │    │      │  │ ActionsService           │    │
│  │ TaskService          │    │      │  │ ├── ActionsClient (K8s)  │    │
│  │ TriggerSvc           │    │      │  │ ├── Enqueue (TaskAction) │    │
│  │                      │    │      │  │ ├── Watch  (TaskAction)  │    │
│  │ AppService ──connect─┼────┼─────>│  │ ├── Abort  (TaskAction)  │    │
│  │ (passthrough)        │    │      │  │                          │    │
│  │                      │    │      │  │ InternalAppService (NEW) │    │
│  └──────────────────────┘    │      │  │ ├── AppK8sClient (K8s)   │    │
│                              │      │  │ │   Deploy (KService)    │    │
│  No DB for apps              │      │  │ │   GetStatus (KService) │    │
│                              │      │  │ │   Stop (del KService)  │    │
│                              │      │  │ └── Replicas (Pods)      │    │
│                              │      │  └────────────┬─────────────┘    │
│                              │      │               │                  │
│                              │      │               ▼                  │
│                              │      │  ┌──────────────────────────┐   │
│                              │      │  │ Kubernetes                │   │
│                              │      │  │ ┌──────────────┐         │   │
│                              │      │  │ │ TaskAction CRs│         │   │
│                              │      │  │ ├──────────────┤         │   │
│                              │      │  │ │ KService CRs │         │   │
│                              │      │  │ │ (apps)       │         │   │
│                              │      │  │ └──────────────┘         │   │
│                              │      │  └──────────────────────────┘   │
└─────────────────────────────┘      └─────────────────────────────────┘
```

**Key design simplification:**
- **No database for apps** — the KService CRD is the single source of truth for both app spec and status. No `apps` table, no repository layer, no migrations.
- **Control plane / data plane split** — AppService (control plane) forwards requests directly to InternalAppService (data plane) via connect. InternalAppService has direct K8s access to manage KService CRDs.
- **No status watcher/sync loop** — when AppService needs app status, it calls InternalAppService, which reads directly from the KService CRD via AppK8sClient.
- **No delete operation** — apps can only be stopped. Stopping deletes the KService CRD. A subsequent Create can re-deploy the app.

**Key patterns to follow:**
- **SetupContext**: Services register handlers on `sc.Mux`, background workers via `sc.AddWorker`
- **buf connect**: gRPC handlers generated from proto definitions

**Current state:** `runs/service/app_service.go` is a stub returning `CodeUnimplemented` for `Create` and empty responses for other RPCs.

---

## 3. Data Model

### 3.1 Protobuf (already defined in `flyteidl2/app/`)

The proto definitions are already complete. Key messages:

| Proto | Purpose |
|-------|---------|
| `App` | Top-level: metadata + spec + status |
| `Spec` | Container/K8sPod payload, autoscaling, ingress, desired_state, timeouts |
| `Status` | DeploymentStatus enum, replicas, ingress URLs, conditions, lease_expiration |
| `Identifier` | project, domain, name |
| `Replica` | Individual pod/instance info |
| `AutoscalingConfig` | min/max replicas, scaling metric (RPS or concurrency) |
| `IngressConfig` | private flag, subdomain, cname |

**DeploymentStatus (derived from KService CRD at read time):**
```
KService not found        → STOPPED
LatestCreated != Ready    → DEPLOYING
Ready=True                → ACTIVE
Ready=False with error    → FAILED
```

### 3.2 No Database — KService CRD as Single Source of Truth

There is **no `apps` table** in the database. The KService CRD in Kubernetes is the single source of truth for both app spec and status.

**Design decisions:**
- **No DB model, no repository layer, no migrations** — all app state lives in the KService CRD
- **Spec is stored as annotations/labels on the KService** — the app spec (container image, autoscaling config, etc.) is encoded into the KService manifest itself
- **Status is derived from KService conditions** — no need to persist or sync status
- **Listing apps** uses K8s label selectors (`flyte.org/app-managed=true`) to find all managed KServices
- **Identifier lookups** use KService naming convention: `{project}-{domain}-{name}`
- This eliminates the need for optimistic locking, migrations, and DB-level change detection

---

## 5. Service Layer

### 5.1 Architecture: InternalAppService ↔ AppK8sClient

```
┌───────────────────────────┐         ┌───────────────────────────────┐
│  Control Plane             │         │  Data Plane                    │
│                            │         │                                │
│  RunService                │         │  ActionsService (tasks)        │
│  TaskService               │         │  ├── ActionsClient (K8s)       │
│  AppService ───connect────┼────────>│                                │
│  (passthrough)            │         │  InternalAppService (NEW)      │
│                            │         │  ├── AppK8sClient (K8s)        │
│                            │         │  │                             │
│  No DB for apps            │         │  │ Create: KService            │
│                            │         │  │ Update: KService            │
│                            │         │  │ Get:    KService            │
│                            │         │  │ Stop:   del KService        │
│                            │         │  │ List:   KServices           │
│                            │         │  │                             │
│                            │         │  └── Kubernetes (KService CRs) │
└───────────────────────────┘         └───────────────────────────────┘
```

**Key simplification:**
- **No database operations for apps** — all state lives in KService CRDs
- No status watcher syncing CRD status back to DB
- No `UpdateStatus` RPC or background reconciliation loop
- No `AppNotifier` pub/sub — status is always read fresh from the CRD
- **Control plane / data plane split** — AppService is a thin passthrough in the control plane; InternalAppService in the data plane has K8s access

### 5.2 InternalAppService Implementation

New file in `actions/service/internal_app_service.go`:

```go
type InternalAppService struct {
    appconnect.UnimplementedAppServiceHandler
    k8sClient AppK8sClientInterface  // direct K8s access for CRD operations
    cfg       *AppConfig
}
```

### 5.3 RPC Implementations

#### Create

```
User → Create(CreateRequest{app})
  1. Validate spec (container or pod payload required)
  2. Generate public host from subdomain pattern: "{name}-{project}-{domain}.{base_domain}"
  3. k8sClient.Deploy(app)                     ← create KService CRD
  4. Return created app (status=PENDING, from CRD initial state)
```

#### Get

```
RunService → InternalAppService.Get(GetRequest{app_id | ingress})
  1. k8sClient.Get(appID or host)              ← get KService CRD (spec + status)
  2. Map KService → App proto (spec + status)
  3. Return app
```

#### Update

```
User → Update(UpdateRequest{app, reason})
  1. k8sClient.Deploy(updated_app)             ← update KService CRD
  2. Return updated app
```

#### Stop

```
User → Stop(StopRequest{app_id})
  1. k8sClient.Stop(appID)                     ← delete KService CRD
  2. Return success
```

> A subsequent Create with the same identifier can re-deploy the app
> by creating a new KService CRD.

#### List

```
User → List(ListRequest{filter_by, limit, token})
  1. k8sClient.List(filters)                   ← list KServices with label selector
  2. Map each KService → App proto (spec + status)
  3. Return apps + next_page_token
```

> List uses label selectors (`flyte.org/app-managed=true`, project/domain labels)
> to find all managed KServices. Pagination uses K8s continue tokens.

#### Watch (server-streaming)

```
User → Watch(WatchRequest{target})
  1. Set up K8s watch on KService objects matching target scope
  2. Send initial state (current apps matching target)
  3. Loop:
     a. Wait for KService event from K8s watch
     b. Map KService → App proto
     c. Send WatchResponse
  4. Close watch on context cancel
```

> Watch is backed by K8s watch on KService CRDs, not an in-memory pub/sub.
> No AppNotifier needed.


---

## 6. K8s Integration

### 6.1 Approach: Knative Service

Apps are deployed as **Knative Services** (KServices), which provide:
- Autoscaling (scale-to-zero, scale-from-zero)
- Revision management (blue-green deploys)
- Traffic routing
- Request timeout enforcement

**Dependency:** Knative Serving must be installed in the cluster.

### 6.2 AppK8sClient — KService Lifecycle Manager

New file: `actions/k8s/app_client.go`

```go
type AppK8sClient struct {
    k8sClient  client.WithWatch    // same as ActionsClient
    k8sCache   cache.Cache
    namespace  string
}

// Interface
type AppK8sClientInterface interface {
    // Deploy creates or updates a KService for the given app
    Deploy(ctx context.Context, app *flyteapp.App) error

    // Stop deletes the KService for the given app
    Stop(ctx context.Context, appID *flyteapp.Identifier) error

    // GetStatus reads the current status from the KService CRD
    GetStatus(ctx context.Context, appID *flyteapp.Identifier) (*flyteapp.Status, error)

    // GetReplicas lists pods for an app
    GetReplicas(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Replica, error)

    // DeleteReplica deletes a specific pod
    DeleteReplica(ctx context.Context, replicaID *flyteapp.ReplicaIdentifier) error
}
```

**Key simplification vs. previous design:**
- No `StartWatching()` / `StopWatching()` — no background watcher syncing status back to DB
- No `appClient` callback to InternalAppService — status is pulled on demand, not pushed
- `GetStatus()` is the new method: reads KService CRD and maps to App Status proto

### 6.3 GetStatus(): KService CRD → App Status

```
InternalAppService.Get()
    │ calls k8sClient.GetStatus(appID)
    ▼
AppK8sClient.GetStatus(appID):
    │
    ├── 1. Get KService by name "{project}-{domain}-{name}"
    │      in configured namespace
    │
    ├── 2. If not found → return Status{phase: STOPPED}
    │
    ├── 3. Map KService status → App Status:
    │      - LatestCreated != LatestReady → DEPLOYING
    │      - Ready=True                   → ACTIVE
    │      - Ready=False with error       → FAILED
    │
    ├── 4. Read replica count from KService status
    │
    ├── 5. Read ingress URL from KService status.url
    │
    └── Return Status proto
```

### 6.4 Deploy(): App Proto → KService

```
InternalAppService.Create() / Update()
    │ calls k8sClient.Deploy(app)
    ▼
AppK8sClient.Deploy(app):
    │
    ├── 1. Build PodSpec from App.Spec
    │      Container/K8sPod → core.TaskTemplate
    │      → flytek8s.ToK8sPodSpec() → corev1.PodSpec
    │      Post-process:
    │        ├── Add container ports from App.Spec
    │        ├── Add tolerations (flyte.org/node-role=worker)
    │        ├── Add identity labels (project/domain/name)
    │        └── Add env vars (FLYTE_APP_NAME, etc.)
    │
    ├── 2. Build KService manifest
    │      Name: "{project}-{domain}-{name}"
    │      Namespace: configurable (default: "flyte-apps")
    │      Labels:
    │        flyte.org/app-managed: "true"
    │      Annotations:
    │        flyte.org/app-id: JSON(identifier)
    │        flyte.org/spec-sha: SHA256(spec)  // change detection
    │      Template Annotations (autoscaling):
    │        autoscaling.knative.dev/min-scale: spec.autoscaling.replicas.min
    │        autoscaling.knative.dev/max-scale: spec.autoscaling.replicas.max
    │        autoscaling.knative.dev/target: scaling_metric.target_value
    │        autoscaling.knative.dev/metric: "rps" | "concurrency"
    │        autoscaling.knative.dev/window: scaledown_period
    │      Spec:
    │        PodSpec: (from step 1)
    │        TimeoutSeconds: spec.timeouts.request_timeout (default 300s, max 3600s)
    │
    ├── 3. Create or Update in K8s
    │      Try k8sClient.Create(kservice)
    │      If AlreadyExists → compare spec-sha → Update if changed
    │      Retry 3x on conflict
    │
    └── Done
```

### 6.5 Stop(): Delete KService CRD

```
User calls Stop(app_id)
    │
    ▼
InternalAppService.Stop()
    │ calls k8sClient.Stop(appID)
    ▼
AppK8sClient.Stop()
    │ Deletes KService from cluster
    ▼
Done — subsequent Get() will see STOPPED (KService not found)
```

### 6.6 Replica Server (Phase 4 — optional)

New file: `actions/k8s/app_replica.go`

**What it is:** A pod-level observability and management layer that lets users
inspect individual pods backing an app, without requiring direct `kubectl` access.

**RPCs:**
- `ListReplica` — show all pods for an app with per-pod status (running, pending, crash-looping)
- `GetReplica` — get status of a specific pod
- `DeleteReplica` — force-kill a specific pod (manual escape hatch)
- `WatchReplicas` — stream pod status changes (for UI dashboards)

**Priority:** Low. The core app lifecycle (Create → Deploy → ACTIVE) works without it.
Users can fall back to `kubectl` for pod-level debugging. Defer to Phase 4.

```go
func (c *AppK8sClient) GetReplicas(ctx, appID) ([]*flyteapp.Replica, error) {
    // List pods with label selector: project + domain + name
    // Convert pod status → Replica proto (reuse flytek8s.DemystifyPending)
}

func (c *AppK8sClient) DeleteReplica(ctx, replicaID) error {
    // Force-kill a specific pod; Knative auto-replaces it
}
```

### 6.7 Registration in ActionsService Setup

Modified: `actions/setup.go`

```go
func Setup(ctx context.Context, sc *app.SetupContext) error {
    // ... existing TaskAction setup ...

    // NEW: App K8s client (KService lifecycle)
    if cfg.Apps.Enabled {
        appK8sClient := actionsk8s.NewAppK8sClient(
            sc.K8sClient,
            sc.K8sCache,
            cfg.Apps.Namespace,  // default: "flyte-apps"
        )
        // No background watcher needed — status is read on demand

        // Create InternalAppService with K8s client (no DB dependency)
        internalAppSvc := service.NewInternalAppService(appK8sClient, &cfg.Apps)
    }
}
```

---

## 7. Integration into Existing v2 Codebase

### 7.1 Files to Create

```
actions/
├── k8s/
│   ├── app_client.go           # AppK8sClient — KService CRUD + status read
│   └── app_replica.go          # Replica queries (pod list/delete)
├── service/
│   └── internal_app_service.go # InternalAppService — K8s-only app management
```

### 7.2 Files to Modify

```
runs/
├── service/
│   └── app_service.go          # Replace stub with passthrough to InternalAppService
├── setup.go                    # Wire AppService with connect client to InternalAppService

actions/
├── setup.go                    # Register AppK8sClient + InternalAppService
├── config/config.go            # Add Apps config section
```

**No changes to repository/, migrations/, or executor/** — apps have no DB operations.

### 7.3 Setup Registration (runs/setup.go)

Replace the current stub:

```go
// Before (stub):
appSvc := service.NewInternalAppService()

// After (real):
// AppService is a passthrough — connects to InternalAppService via connect client
appSvc := service.NewAppService(internalAppServiceClient)
```

### 7.4 ActionsService Setup (actions/setup.go)

```go
// Add to existing Setup():
if cfg.Apps.Enabled {
    appK8sClient := actionsk8s.NewAppK8sClient(sc.K8sClient, sc.K8sCache, cfg.Apps.Namespace)
    internalAppSvc := service.NewInternalAppService(appK8sClient, &service.AppConfig{
        PublicURLPattern: cfg.Apps.PublicURLPattern,
    })
    // Register InternalAppService on the mux
}
```

---

## 8. Configuration

### 8.1 ActionsService Config

Add to `actions/config/config.go`:

```go
type AppConfig struct {
    // Enabled controls whether the app controller is started.
    Enabled bool `json:"enabled" pflag:",Enable app deployment controller"`

    // Namespace is the K8s namespace for KServices.
    Namespace string `json:"namespace" pflag:",Namespace for app KServices" default:"flyte-apps"`

    // DefaultRequestTimeout for apps that don't specify one.
    DefaultRequestTimeout time.Duration `json:"defaultRequestTimeout" default:"5m"`

    // MaxRequestTimeout is the hard cap on request timeout.
    MaxRequestTimeout time.Duration `json:"maxRequestTimeout" default:"1h"`

    // PublicURLPattern is a Go template for generating public URLs.
    // Variables: {{.Name}}, {{.Project}}, {{.Domain}}
    // Example: "https://{{.Name}}-{{.Project}}.apps.flyte.example.com"
    PublicURLPattern string `json:"publicUrlPattern" pflag:",URL pattern for app ingress"`
}
```

---

## 9. End-to-End Dataflow

### 9.1 Happy Path: Create and Deploy

```
SDK/CLI          AppService (ctrl)    InternalAppService (data)   AppK8sClient
   │                    │                       │                       │
   │  Create(app_spec)  │                       │                       │
   │───────────────────>│  Create(app_spec)     │                       │
   │                    │──────────────────────>│                       │
   │                    │                       │ Validate spec         │
   │                    │                       │ Deploy(app)           │
   │                    │                       │──────────────────────>│
   │                    │                       │                       │ Build PodSpec
   │                    │                       │                       │ Create KService
   │                    │  CreateResponse(app)  │                       │
   │  CreateResponse    │<──────────────────────│                       │
   │<───────────────────│                       │                       │
   │                    │                       │   (Knative deploys)   │
   │                    │                       │                       │
   │  Get(app_id)       │                       │                       │
   │───────────────────>│  Get(app_id)          │                       │
   │                    │──────────────────────>│                       │
   │                    │                       │ Get(appID)            │
   │                    │                       │──────────────────────>│
   │                    │                       │                       │ Read KService CRD
   │                    │                       │  App{spec + ACTIVE}   │
   │                    │                       │<──────────────────────│
   │                    │  App{spec + ACTIVE}   │                       │
   │  App{spec+ACTIVE}  │<──────────────────────│                       │
   │<───────────────────│                       │                       │
   │                    │                       │                       │
   │  App is live at:   │                       │                       │
   │  https://myapp-proj.apps.flyte.example.com                        │
```

### 9.2 Update (Spec Change)

```
SDK/CLI          AppService (ctrl)    InternalAppService (data)   AppK8sClient
   │                    │                       │                       │
   │  Update(new_spec)  │                       │                       │
   │───────────────────>│  Update(new_spec)     │                       │
   │                    │──────────────────────>│                       │
   │                    │                       │ Deploy(updated_app)    │
   │                    │                       │──────────────────────>│
   │                    │                       │                       │ Compare spec-sha
   │                    │                       │                       │ Update KService
   │                    │                       │                       │ (Knative creates
   │                    │                       │                       │  new Revision)
   │                    │  UpdateResponse(app)  │                       │
   │  UpdateResponse    │<──────────────────────│                       │
   │<───────────────────│                       │                       │
```

### 9.3 Stop

```
SDK/CLI          AppService (ctrl)    InternalAppService (data)   AppK8sClient
   │                    │                       │                       │
   │  Stop(app_id)      │                       │                       │
   │───────────────────>│  Stop(app_id)         │                       │
   │                    │──────────────────────>│                       │
   │                    │                       │ Stop(appID)           │
   │                    │                       │──────────────────────>│
   │                    │                       │                       │ Delete KService
   │                    │  StopResponse         │                       │
   │  StopResponse      │<──────────────────────│                       │
   │<───────────────────│                       │                       │
   │                    │                       │                       │
   │  Get(app_id)       │                       │                       │
   │───────────────────>│  Get(app_id)          │                       │
   │                    │──────────────────────>│                       │
   │                    │                       │ Get → KService not found
   │                    │  App not found (STOPPED)                      │
   │  Not found         │<──────────────────────│                       │
   │<───────────────────│                       │                       │
```

---

## 10. Phased Implementation Plan

### Phase 1: K8s Integration + Service Layer

1. Create `AppK8sClient` in `actions/k8s/app_client.go` (Deploy, Stop, Get, List, GetStatus)
2. Implement `InternalAppService` RPCs: Create, Get, Update, Stop, List
3. Wire `AppService` (passthrough) in `runs/setup.go`
4. Wire `InternalAppService` + `AppK8sClient` in `actions/setup.go`
5. **Unit tests** for service (mock AppK8sClient)
6. **Integration tests** with envtest + Knative CRDs

### Phase 2: Streaming RPCs

1. Implement `Watch` RPC (backed by K8s watch on KService CRDs)
3. **End-to-end tests** with Knative in kind/minikube

### Phase 3: Polish

1. Log streaming (`AppLogsService.TailLogs` → K8s pod logs)
2. Replica Server (pod-level observability)
3. Ingress configuration (Knative domain mapping or Istio VirtualService)
4. Helm chart updates (`charts/flyte-binary/`, `charts/flyte-sandbox/`)
5. SDK integration (Python `union.app()` decorator)
6. Documentation

---

## 11. Testing Strategy

| Layer | Test Type | Tool |
|-------|-----------|------|
| InternalAppService | Unit | Mock AppK8sClient (mockery) |
| AppK8sClient | Integration | envtest (controller-runtime) + Knative CRDs |
| E2E | Integration | kind + Knative + flyte-sandbox |

---

## 12. Open Questions

1. **Knative dependency** — Should we support a non-Knative fallback (raw Deployment + HPA + Ingress) for clusters without Knative? This adds complexity but broadens adoption.

2. **Namespace strategy** — One namespace per app vs. shared namespace (`flyte-apps`)? Another option is `project-domain` namespaces for isolation.

3. **Ingress controller** — Knative typically uses Kourier or Istio. Should we prescribe one, or be agnostic?

4. **Auth for app endpoints** — How should app endpoints be authenticated? Knative doesn't provide auth out of the box. Options: Istio auth policy, app-level API keys, or proxy sidecar.

5. **Scale-to-zero** — Knative supports scale-to-zero. Should we expose `min_replicas: 0` as a valid option, or require at least 1?

6. **Resource quotas** — Should apps share the same resource pool as workflow executions, or have a separate quota?

7. **List performance** — For `List` with many apps, K8s list with label selectors should be efficient, but at very large scale, should we add an informer cache?

---

## 13. Appendix: Key Reference Files

### Flyte v2 OSS

| Component | Path |
|-----------|------|
| Proto definitions | `flyteidl2/app/app_definition.proto` |
| Proto payloads | `flyteidl2/app/app_payload.proto` |
| Proto service | `flyteidl2/app/app_service.proto` |
| Proto replicas | `flyteidl2/app/replica_definition.proto` |
| Proto logs | `flyteidl2/app/app_logs_service.proto` |
| App stub (to replace) | `runs/service/app_service.go` |
| Setup — runs (to modify) | `flyte2/runs/setup.go` |
| Setup — actions (to modify) | `flyte2/actions/setup.go` |
| ActionsService (pattern ref) | `flyte2/actions/service/actions_service.go` |
| ActionsClient (pattern ref) | `flyte2/actions/k8s/` |
| App framework | `flyte2/app/context.go` |
| Manager (unified binary) | `flyte2/manager/` |
