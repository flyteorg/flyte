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
- Manage app lifecycle: create, update, stop, delete
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
┌──────────────────────────────┐      ┌──────────────────────────────────┐
│  Control Plane                │      │  Data Plane                       │
│                               │      │                                   │
│  ┌──────────────────────┐    │      │  ┌──────────────────────────┐    │
│  │ RunService            │    │      │  │ ActionsService            │    │
│  │ TaskService           │    │      │  │ ├── ActionsClient (K8s)   │    │
│  │ TriggerSvc            │    │      │  │ ├── Enqueue (TaskAction)  │    │
│  │                       │    │      │  │ ├── Watch  (TaskAction)   │    │
│  │ AppService ──connect──┼────┼─────>│  │ ├── Abort  (TaskAction)   │    │
│  │ ├── In-memory cache   │    │      │  │                           │    │
│  │ └── TTL-based eviction│    │      │  │ InternalAppService (NEW)  │    │
│  │                       │    │      │  │ ├── AppK8sClient (K8s)    │    │
│  └───────────────────────┘    │      │  │ │   Deploy (KService)     │    │
│                               │      │  │ │   GetStatus (KService)  │    │
│  No DB for apps               │      │  │ │   Stop (scale-to-zero)  │    │
│                               │      │  │ │   Delete (del KService) │    │
│                               │      │  │ └── Replicas (Pods)       │    │
│                               │      │  └────────────┬──────────────┘    │
│                               │      │               │                   │
│                               │      │               ▼                   │
│                               │      │  ┌───────────────────────────┐   │
│                               │      │  │ Kubernetes                 │   │
│                               │      │  │ ┌───────────────┐         │   │
│                               │      │  │ │ TaskAction CRs │         │   │
│                               │      │  │ ├───────────────┤         │   │
│                               │      │  │ │ KService CRs  │         │   │
│                               │      │  │ │ (apps)        │         │   │
│                               │      │  │ └───────────────┘         │   │
│                               │      │  └───────────────────────────┘   │
└──────────────────────────────┘      └──────────────────────────────────┘
```

**Key design simplification:**
- **No database for apps** — the KService CRD is the single source of truth for both app spec and status. No `apps` table, no repository layer, no migrations.
- **Control plane / data plane split** — AppService (control plane) forwards requests to InternalAppService (data plane) via connect. AppService maintains an in-memory cache to avoid redundant calls.
- **In-memory cache in AppService** — Get/List responses are cached with a short TTL. Mutating operations (Create, Update, Stop, Delete) invalidate the relevant cache entries. This reduces cross-plane RPC calls without introducing stale-state risk.
- **No status watcher/sync loop** — when a cache miss occurs, AppService calls InternalAppService, which reads directly from the KService CRD via AppK8sClient.
- **Stop vs Delete** — Stop scales the app to zero by setting `max-scale=0` on the KService (app can be resumed). Delete removes the KService CRD entirely.

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
KService not found            → NOT_FOUND (deleted)
max-scale=0                   → STOPPED
LatestCreated != LatestReady  → DEPLOYING
Ready=True                    → ACTIVE
Ready=False with error        → FAILED
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
┌───────────────────────────────────┐  ┌───────────────────────────────────┐
│  Control Plane                    │  │  Data Plane                       │
│                                   │  │                                   │
│  RunService                       │  │  ActionsService (tasks)           │
│  TaskService                      │  │  ├── ActionsClient (K8s)          │
│  AppService ────────connect──────┼─>│                                   │
│  ├── In-memory cache (TTL)       │  │  InternalAppService (NEW)         │
│  │   cache hit → return          │  │  ├── AppK8sClient (K8s)           │
│  │   cache miss → call internal  │  │  │                                │
│  │   mutate → invalidate cache   │  │  │  Create: KService              │
│  │                                │  │  │  Update: KService              │
│  │                                │  │  │  Get:    KService              │
│  No DB for apps                   │  │  │  Stop:   scale-to-zero         │
│                                   │  │  │  Delete: del KService          │
│                                   │  │  │  List:   KServices             │
│                                   │  │  │                                │
│                                   │  │  └── Kubernetes (KService CRs)    │
└───────────────────────────────────┘  └───────────────────────────────────┘
```

**Key simplification:**
- **No database operations for apps** — all state lives in KService CRDs
- **In-memory cache in AppService** — avoids redundant connect calls to InternalAppService for repeated Get/List requests. Cache is invalidated on Create, Update, Stop, and Delete.
- No status watcher syncing CRD status back to DB
- No `UpdateStatus` RPC or background reconciliation loop
- No `AppNotifier` pub/sub — status is always read fresh from the CRD (on cache miss)
- **Control plane / data plane split** — AppService caches and forwards in the control plane; InternalAppService in the data plane has K8s access

### 5.2 AppService Cache (Control Plane)

AppService maintains an in-memory TTL cache to reduce cross-plane RPC calls to InternalAppService.

```go
type AppService struct {
    appconnect.UnimplementedAppServiceHandler
    internalClient appconnect.AppServiceClient  // connect client to InternalAppService (data plane pod)
    cache          *AppCache
}

type AppCache struct {
    mu    sync.RWMutex
    items map[string]*cacheEntry  // key: "project/domain/name"
    ttl   time.Duration           // default: 30s
}
```

**Cache behavior:**
- **Get**: Check cache first. On hit, return cached App. On miss, call InternalAppService, cache the result.
- **List**: Not cached (results vary by filter/pagination). Always forwarded to InternalAppService.
- **Create, Update, Stop, Delete**: Forward to InternalAppService, then invalidate the cache entry for that app.
- **TTL eviction**: Cached entries expire after a configurable TTL (default 30s). This bounds staleness for status changes driven by Knative (e.g., scaling events).

> The cache is per-instance (not shared across replicas). This is acceptable because
> app status changes are infrequent relative to read frequency, and the TTL ensures
> eventual consistency.

### 5.3 InternalAppService Implementation

New file in `app/internal/service/internal_app_service.go`:

```go
type InternalAppService struct {
    appconnect.UnimplementedAppServiceHandler
    k8sClient AppK8sClientInterface  // direct K8s access for CRD operations
    cfg       *AppConfig
}
```

### 5.4 RPC Implementations (InternalAppService)

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
  1. k8sClient.Stop(appID)                     ← set max-scale=0 on KService
  2. Return success
```

> Stop scales the app to zero replicas by patching `max-scale=0` on the
> KService. The KService CRD remains — the app can be resumed by updating
> it with a new spec or by restoring the original scaling config.

#### Delete

```
User → Delete(DeleteRequest{app_id})
  1. k8sClient.Delete(appID)                   ← delete KService CRD
  2. Return success
```

> Delete removes the KService CRD entirely. The app is gone and must be
> re-created from scratch.

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

New file: `app/internal/k8s/app_client.go`

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

    // Stop scales the KService to zero (sets max-scale=0)
    Stop(ctx context.Context, appID *flyteapp.Identifier) error

    // Delete removes the KService CRD entirely
    Delete(ctx context.Context, appID *flyteapp.Identifier) error

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

### 6.5 Stop(): Scale to Zero

```
User calls Stop(app_id)
    │
    ▼
InternalAppService.Stop()
    │ calls k8sClient.Stop(appID)
    ▼
AppK8sClient.Stop()
    │ Patches KService: set annotation autoscaling.knative.dev/max-scale=0
    ▼
Done — Knative scales to zero pods. KService CRD remains.
       Subsequent Get() returns status=STOPPED (0 replicas).
       App can be resumed via Update().
```

### 6.6 Delete(): Remove KService CRD

```
User calls Delete(app_id)
    │
    ▼
InternalAppService.Delete()
    │ calls k8sClient.Delete(appID)
    ▼
AppK8sClient.Delete()
    │ Deletes KService from cluster
    ▼
Done — subsequent Get() returns not found.
```

### 6.7 Replica Server (Phase 4 — optional)

New file: `app/internal/k8s/app_replica.go`

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

### 6.8 Registration

- **Control plane pod:** `app/setup.go` registers AppService (see section 7.3)
- **Data plane pod:** `internalapp/setup.go` registers InternalAppService with AppK8sClient (see section 7.4)

Both are called from the manager. No modifications to `actions/setup.go` are needed.

---

## 7. Integration into Existing v2 Codebase

### 7.1 New Top-Level Directory: `app/`

A new `app/` top-level directory (replacing the existing framework files, alongside `actions/`, `runs/`, `manager/`) to house all app service code:

```
app/                              # NEW top-level directory
├── setup.go                       # Control plane setup (AppService + connect client)
├── config/
│   └── config.go                  # AppConfig (TTL, InternalAppService URL)
├── service/
│   └── app_service.go             # AppService (control plane, cache + connect client)
├── internal/                      # Data plane (separate pod in production)
│   ├── setup.go                   # Data plane setup (InternalAppService + K8s client)
│   ├── config/
│   │   └── config.go              # InternalAppConfig (namespace, timeouts, URL pattern)
│   ├── service/
│   │   └── internal_app_service.go # InternalAppService (K8s-only app management)
│   └── k8s/
│       ├── app_client.go          # AppK8sClient — KService CRUD + status read
│       └── app_replica.go         # Replica queries (pod list/delete)
```

### 7.2 Files to Modify

```
manager/
│   └── ...                     # Wire app/ service into the unified binary
```

**No changes to actions/, runs/, repository/, migrations/, or executor/**.

### 7.3 Control Plane Setup (app/setup.go)

```go
func Setup(ctx context.Context, sc *app.SetupContext) error {
    cfg := config.GetConfig()

    // Connect client to InternalAppService (running in data plane pod)
    internalClient := appconnect.NewAppServiceClient(
        sc.HTTPClient,
        cfg.InternalAppServiceURL,
    )

    // AppService with cache
    appSvc := service.NewAppService(internalClient, cfg)
    appconnect.RegisterAppServiceHandler(sc.Mux, appSvc)
    return nil
}
```

### 7.4 Data Plane Setup (internalapp/setup.go)

```go
func Setup(ctx context.Context, sc *app.SetupContext) error {
    cfg := config.GetConfig()

    // K8s client for KService lifecycle
    appK8sClient := k8s.NewAppK8sClient(sc.K8sClient, sc.K8sCache, cfg.Namespace)

    // InternalAppService with direct K8s access
    internalAppSvc := service.NewInternalAppService(appK8sClient, cfg)
    appconnect.RegisterInternalAppServiceHandler(sc.Mux, internalAppSvc)
    return nil
}
```

Both are wired into the manager. In production, the control plane pod runs `apps.Setup()` and the data plane pod runs `internalapp.Setup()`. In dev/sandbox mode, both can run in the same process.

---

## 8. Configuration

### 8.1 Apps Config

New file: `app/config/config.go`:

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

    // CacheTTL is the TTL for the AppService in-memory cache.
    // Controls how long Get responses are cached in the control plane.
    CacheTTL time.Duration `json:"cacheTTL" default:"30s"`
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

### 9.3 Stop (Scale to Zero)

```
SDK/CLI          AppService (ctrl)    InternalAppService (data)   AppK8sClient
   │                    │                       │                       │
   │  Stop(app_id)      │                       │                       │
   │───────────────────>│  Stop(app_id)         │                       │
   │                    │──────────────────────>│                       │
   │                    │                       │ Stop(appID)           │
   │                    │                       │──────────────────────>│
   │                    │                       │                       │ Patch max-scale=0
   │                    │  StopResponse         │                       │
   │  StopResponse      │<──────────────────────│                       │
   │<───────────────────│                       │                       │
   │                    │                       │                       │
   │  Get(app_id)       │                       │                       │
   │───────────────────>│  Get(app_id)          │                       │
   │                    │──────────────────────>│                       │
   │                    │                       │ Get(appID)            │
   │                    │                       │──────────────────────>│
   │                    │                       │                       │ Read KService CRD
   │                    │                       │  App{spec + STOPPED}  │
   │                    │  App{spec + STOPPED}  │<──────────────────────│
   │  App{spec+STOPPED} │<──────────────────────│                       │
   │<───────────────────│                       │                       │
```

### 9.4 Delete

```
SDK/CLI          AppService (ctrl)    InternalAppService (data)   AppK8sClient
   │                    │                       │                       │
   │  Delete(app_id)    │                       │                       │
   │───────────────────>│  Delete(app_id)       │                       │
   │                    │──────────────────────>│                       │
   │                    │                       │ Delete(appID)         │
   │                    │                       │──────────────────────>│
   │                    │                       │                       │ Delete KService
   │                    │  DeleteResponse       │                       │
   │  DeleteResponse    │<──────────────────────│                       │
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

1. **Ingress controller** — Knative typically uses Kourier or Istio. Should we prescribe one, or be agnostic?

2. **Auth for app endpoints** — How should app endpoints be authenticated? Knative doesn't provide auth out of the box. Options: Istio auth policy, app-level API keys, or proxy sidecar.
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
| App stub (to remove) | `runs/service/app_service.go` |
| Apps service — control plane (NEW) | `app/` |
| InternalApp service — data plane (NEW) | `internalapp/` |
| ActionsService (pattern ref) | `actions/service/actions_service.go` |
| ActionsClient (pattern ref) | `actions/k8s/` |
| App framework (to be removed) | `app/context.go` |
| Manager (to modify) | `manager/` |
