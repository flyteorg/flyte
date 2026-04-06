# App Service Design Doc

**Status:** Draft
**Author:** Kevin
**Date:** 2026-04-02
**Updated:** 2026-04-06

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
┌──────────────────────────────────────────────────────────────────────┐
│  Manager (unified binary)                                             │
│                                                                       │
│  ┌───────────────────┐  ┌─────────────────────────┐  ┌────────────┐ │
│  │ RunService         │  │ ActionsService           │  │ Executor   │ │
│  │ (port 8090)       │  │ (K8s watch + bridge)     │  │ (operator) │ │
│  │                   │  │                          │  │            │ │
│  │ TaskService       │  │ ActionsClient            │  │ Reconcile  │ │
│  │ TriggerSvc        │  │ ├── Enqueue (TaskAction) │  │ TaskAction │ │
│  │                   │  │ ├── Watch  (TaskAction)  │  │ CRs        │ │
│  │                   │  │ ├── Abort  (TaskAction)  │  │            │ │
│  │ → AppService ─────┼──┤                          │  │            │ │
│  │   (connect call)  │  │ AppService (NEW)         │  │            │ │
│  │                   │  │ ├── AppRepo (DB: spec)   │  │            │ │
│  │                   │  │ ├── AppK8sClient (K8s)   │  │            │ │
│  │                   │  │ │   Deploy (KService)    │  │            │ │
│  │                   │  │ │   GetStatus (KService) │  │            │ │
│  │                   │  │ │   Stop (del KService)  │  │            │ │
│  │                   │  │ └── Replicas (Pods)      │  │            │ │
│  └──────────┬────────┘  └────────────┬─────────────┘  └──────┬─────┘ │
│             │                        │                       │       │
│             ▼                        ▼                       ▼       │
│      ┌────────────┐         ┌──────────────────────────────────┐     │
│      │ PostgreSQL/ │         │ Kubernetes                       │     │
│      │ SQLite      │         │ ┌──────────────┐ ┌────────────┐ │     │
│      │ (apps table │         │ │ TaskAction CRs│ │ KService   │ │     │
│      │  spec only) │         │ │ (tasks)       │ │ CRs (apps) │ │     │
│      └────────────┘         │ └──────────────┘ └────────────┘ │     │
│                              └──────────────────────────────────┘     │
└──────────────────────────────────────────────────────────────────────┘
```

**Key design simplification:**
- **DB stores spec only** — no status column. The KService CRD is the single source of truth for app status.
- **No status watcher/sync loop** — when RunService needs app status, it calls AppService (via connect, co-located in the same pod), which reads directly from the KService CRD via AppK8sClient.
- **No delete operation** — apps can only be stopped. Stopping deletes the KService CRD. The spec remains in DB for potential restart.

**Key patterns to follow:**
- **SetupContext**: Services register handlers on `sc.Mux`, background workers via `sc.AddWorker`
- **buf connect**: gRPC handlers generated from proto definitions
- **GORM + gormigrate**: Database models with migration framework
- **Repository pattern**: `interfaces/` → `impl/` → `models/` → `transformers/`

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

> Status is never stored in DB. It is read from the KService CRD on every Get request.

### 3.2 Database Model

New file: `runs/repository/models/app.go`

```go
package models

import (
    "time"
)

type App struct {
    ID uint `gorm:"primaryKey"`

    // App Identifier (unique composite key)
    Project string `gorm:"not null;uniqueIndex:idx_apps_identifier,priority:1"`
    Domain  string `gorm:"not null;uniqueIndex:idx_apps_identifier,priority:2"`
    Name    string `gorm:"not null;uniqueIndex:idx_apps_identifier,priority:3"`

    // Serialized protobuf (flyteidl2.app.Spec)
    Spec []byte `gorm:"type:bytea;not null"`

    // For optimistic locking and change detection
    Revision   uint64 `gorm:"not null;default:1"`
    SpecDigest []byte `gorm:"type:bytea"`

    // Denormalized for ingress lookups
    PublicHost string `gorm:"uniqueIndex:idx_apps_public_host"`

    // Metadata
    CreatedBy string
    CreatedAt time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_apps_created"`
    UpdatedAt time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (App) TableName() string { return "apps" }
```

**Design decisions:**
- **No `Status` column** — status is read from the KService CRD at query time
- **No `Phase`, `DesiredState`, `CurrentReplicas` columns** — these are all derived from the CRD
- **No `Org` column** — identifier is project + domain + name
- Spec stored as serialized protobuf bytes (same pattern as `Action.ActionSpec`)
- `SpecDigest` enables change detection without deserializing the full spec
- `Revision` for optimistic concurrency control
- `PublicHost` unique index for ingress-based lookups

### 3.3 Migration

New migration added to `runs/migrations/migrations.go`:

```go
{
    ID: "20260402_apps_init",
    Migrate: func(tx *gorm.DB) error {
        return tx.AutoMigrate(&models.App{})
    },
}
```

---

## 4. Repository Layer

### 4.1 Interface

New file: `runs/repository/interfaces/app.go`

```go
package interfaces

import (
    "context"
    "github.com/flyteorg/flyte/v2/runs/repository/models"
)

type AppRepo interface {
    Create(ctx context.Context, app *models.App) error
    Get(ctx context.Context, project, domain, name string) (*models.App, error)
    GetByHost(ctx context.Context, host string) (*models.App, error)
    Update(ctx context.Context, oldRevision uint64, app *models.App) (*models.App, error)
    List(ctx context.Context, filters ListFilters, limit int, token string) ([]*models.App, string, error)
}
```

### 4.2 Add to Repository interface

```go
// runs/repository/interfaces/repository.go
type Repository interface {
    ActionRepo() ActionRepo
    TaskRepo() TaskRepo
    AppRepo() AppRepo  // NEW
}
```

### 4.3 Implementation

New file: `runs/repository/impl/app.go`

Key behaviors:
- **Create**: INSERT with initial revision=1
- **Update**: `UPDATE ... WHERE revision = ? AND spec_digest = ?` (optimistic lock)
- **Get**: Lookup by composite key or by `PublicHost`
- **List**: Filter by project; paginate by `updated_at` cursor

No `Delete`, `UpdateStatus`, or `ListByPhase` — these are no longer needed.

---

## 5. Service Layer

### 5.1 Architecture: AppService ↔ AppK8sClient

```
┌─────────────────────────────────────────────────────────────────┐
│  Manager (unified binary)                                        │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │ Manager (RunService + ActionsService in same pod)            │ │
│  │                                                              │ │
│  │  RunService ──connect──> AppService (NEW)                    │ │
│  │                          ├── AppRepo (DB: spec)              │ │
│  │                          ├── AppK8sClient (K8s)              │ │
│  │                          │                                   │ │
│  │                          │ Create: DB + KService             │ │
│  │                          │ Update: DB + KService             │ │
│  │                          │ Get:    DB + KService             │ │
│  │                          │ Stop:   del KService              │ │
│  │                          │ List:   DB + KServices            │ │
│  │                                                              │ │
│  │  ActionsService (tasks)                                      │ │
│  │  ├── ActionsClient (K8s)                                     │ │
│  └──────────────────────────────────────────────────────────────┘ │
│             │                                   │                │
│             ▼                                   ▼                │
│      ┌────────────┐                    ┌──────────────┐         │
│      │ PostgreSQL/ │                    │ Kubernetes    │         │
│      │ SQLite      │                    │ (KService CRs│         │
│      │ (spec only) │                    │  + Pods)     │         │
│      └────────────┘                    └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

**Key simplification vs. previous design:**
- No status watcher syncing CRD status back to DB
- No `UpdateStatus` RPC or background reconciliation loop
- No `AppNotifier` pub/sub — status is always read fresh from the CRD
- DB is only written on Create and Update (spec changes)

### 5.2 AppService Implementation

Replace the stub in `runs/service/app_service.go`:

```go
type AppService struct {
    appconnect.UnimplementedAppServiceHandler
    repo      interfaces.AppRepo
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
  3. Compute spec_digest = SHA256(spec)
  4. Serialize spec to protobuf bytes
  5. repo.Create(app)                          ← write spec to DB
  6. k8sClient.Deploy(app)                     ← create KService CRD
  7. Return created app (status=PENDING, from CRD absence/initial state)
```

#### Get

```
RunService → AppService.Get(GetRequest{app_id | ingress})
  1. If app_id: repo.Get(project, domain, name)          ← get spec from DB
  2. If ingress: repo.GetByHost(host)                    ← get spec from DB
  3. k8sClient.GetStatus(appID)                          ← get status from KService CRD
  4. Combine spec (DB) + status (CRD) into App proto
  5. Return app
```

#### Update

```
User → Update(UpdateRequest{app, reason})
  1. Compute new spec_digest
  2. If spec changed (digest differs):
     a. Increment revision
     b. repo.Update(oldRevision, app)          ← update spec in DB
     c. k8sClient.Deploy(updated_app)          ← update KService CRD
  3. Return updated app
```

#### Stop

```
User → Stop(StopRequest{app_id})
  1. k8sClient.Stop(appID)                     ← delete KService CRD
  2. No DB update needed
  3. Return success
```

> The spec remains in DB. A subsequent Create or Update with the same identifier
> can re-deploy the app by creating a new KService CRD.

#### List

```
User → List(ListRequest{filter_by, limit, token})
  1. repo.List(filters, limit, token)          ← get specs from DB
  2. For each app: k8sClient.GetStatus(appID)  ← get status from CRD
  3. Return apps (spec + status) + next_page_token
```

> For List, status lookups can be batched by listing all KServices with
> label `flyte.org/app-managed=true` and matching by app identifier.

#### Watch (server-streaming)

```
User → Watch(WatchRequest{target})
  1. Set up K8s watch on KService objects matching target scope
  2. Send initial state (current apps matching target)
  3. Loop:
     a. Wait for KService event from K8s watch
     b. Look up spec from DB
     c. Combine spec + status from KService event
     d. Send WatchResponse
  4. Close watch on context cancel
```

> Watch is backed by K8s watch on KService CRDs, not an in-memory pub/sub.
> No AppNotifier needed.

#### Lease (server-streaming) — optional, for future multi-cluster

In single-cluster mode, the Lease RPC is not needed.

```go
func (s *AppService) Lease(...) error {
    return connect.NewError(connect.CodeUnimplemented,
        errors.New("Lease is not needed in single-cluster mode"))
}
```

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
- No `appClient` callback to AppService — status is pulled on demand, not pushed
- `GetStatus()` is the new method: reads KService CRD and maps to App Status proto

### 6.3 GetStatus(): KService CRD → App Status

```
AppService.Get()
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
AppService.Create() / Update()
    │ persists spec to DB
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
AppService.Stop()
    │ calls k8sClient.Stop(appID)
    ▼
AppK8sClient.Stop()
    │ Deletes KService from cluster
    │ (no DB update)
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

        // Make appK8sClient available to AppService via SetupContext
    }
}
```

---

## 7. Integration into Existing v2 Codebase

### 7.1 Files to Create

```
runs/
├── repository/
│   ├── interfaces/
│   │   └── app.go              # AppRepo interface
│   ├── impl/
│   │   └── app.go              # GORM implementation
│   ├── models/
│   │   └── app.go              # App DB model (spec only, no status)
│   └── transformers/
│       └── app.go              # Proto ↔ Model conversion

actions/
├── k8s/
│   ├── app_client.go           # AppK8sClient — KService CRUD + status read
│   └── app_replica.go          # Replica queries (pod list/delete)
```

### 7.2 Files to Modify

```
runs/
├── repository/
│   ├── interfaces/
│   │   └── repository.go       # Add AppRepo() to Repository interface
│   └── repository.go           # Wire AppRepo in NewRepository()
├── migrations/
│   └── migrations.go           # Add App model + migration
├── service/
│   └── app_service.go          # Replace stub with full implementation
├── setup.go                    # Wire real AppService with repo + appK8sClient

actions/
├── setup.go                    # Register AppK8sClient
├── config/config.go            # Add Apps config section
```

**No changes to executor/** — the executor only handles TaskAction CRs.

### 7.3 Setup Registration (runs/setup.go)

Replace the current stub:

```go
// Before (stub):
appSvc := service.NewAppService()

// After (real):
appRepo := impl.NewAppRepo(sc.DB)
appSvc := service.NewAppService(appRepo, appK8sClient, &service.AppConfig{
    PublicURLPattern: cfg.Apps.PublicURLPattern,
})
```

### 7.4 ActionsService Setup (actions/setup.go)

```go
// Add to existing Setup():
if cfg.Apps.Enabled {
    appK8sClient := actionsk8s.NewAppK8sClient(sc.K8sClient, sc.K8sCache, cfg.Apps.Namespace)
    // No background watcher — expose appK8sClient for AppService to use
}
```

---

## 8. Configuration

### 8.1 Runs Service Config

Add to `runs/config/config.go`:

```go
type AppsConfig struct {
    // PublicURLPattern is a Go template for generating public URLs.
    // Variables: {{.Name}}, {{.Project}}, {{.Domain}}
    // Example: "https://{{.Name}}-{{.Project}}.apps.flyte.example.com"
    PublicURLPattern string `json:"publicUrlPattern" pflag:",URL pattern for app ingress"`
}
```

### 8.2 ActionsService Config

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
}
```

---

## 9. End-to-End Dataflow

### 9.1 Happy Path: Create and Deploy

```
SDK/CLI              RunService             AppService             AppK8sClient
   │                    │                       │                       │
   │  Create(app_spec)  │                       │                       │
   │───────────────────>│  Create(app_spec)     │                       │
   │                    │──────────────────────>│                       │
   │                    │                       │ Validate spec         │
   │                    │                       │ Generate public host  │
   │                    │                       │ INSERT spec to DB     │
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
   │                    │                       │ Get spec from DB      │
   │                    │                       │ GetStatus(appID)      │
   │                    │                       │──────────────────────>│
   │                    │                       │                       │ Read KService CRD
   │                    │                       │  Status{ACTIVE}       │
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
SDK/CLI              RunService             AppService             AppK8sClient
   │                    │                       │                       │
   │  Update(new_spec)  │                       │                       │
   │───────────────────>│  Update(new_spec)     │                       │
   │                    │──────────────────────>│                       │
   │                    │                       │ Compute new spec_digest│
   │                    │                       │ Increment revision     │
   │                    │                       │ UPDATE spec in DB      │
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
SDK/CLI              RunService             AppService             AppK8sClient
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
   │                    │                       │ Get spec from DB      │
   │                    │                       │ GetStatus → not found │
   │                    │  App{spec + STOPPED}  │                       │
   │  App{spec+STOPPED} │<──────────────────────│                       │
   │<───────────────────│                       │                       │
```

---

## 10. Phased Implementation Plan

### Phase 1: Control Plane (CRUD + DB) + K8s Integration

1. Add `App` model to `runs/repository/models/app.go` (spec only, no status)
2. Add migration to `runs/migrations/migrations.go`
3. Create `AppRepo` interface and GORM implementation
4. Create `AppK8sClient` in `actions/k8s/app_client.go` (Deploy, Stop, GetStatus)
5. Implement `AppService` RPCs: Create, Get, Update, Stop, List
6. Wire in `runs/setup.go`, `runs/repository/repository.go`, `actions/setup.go`
7. **Unit tests** for repo + service (mock AppK8sClient)
8. **Integration tests** with envtest + Knative CRDs

### Phase 2: Streaming RPCs

1. Implement `Watch` RPC (backed by K8s watch on KService CRDs)
2. Implement `Lease` RPC stub (for future multi-cluster)
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
| Repository | Unit | GORM + SQLite in-memory |
| AppService | Unit | Mock repo + mock AppK8sClient (mockery) |
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

7. **List performance** — For `List` with many apps, reading status from individual KService CRDs could be slow. Should we batch via label-selector list, or add a cache?

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
| Repository pattern ref | `flyte2/runs/repository/` |
| Migration pattern ref | `flyte2/runs/migrations/migrations.go` |
| App framework | `flyte2/app/context.go` |
| Manager (unified binary) | `flyte2/manager/` |
