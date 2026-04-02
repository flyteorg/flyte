# OSS Flyte v2 — App Service Design Doc

**Status:** Draft
**Author:** Kevin
**Date:** 2026-04-02

---

## 1. Overview

This document specifies the design for implementing the **App Service** in OSS Flyte v2. The App Service enables users to deploy and manage long-running applications (serving endpoints, LLM inference, APIs, dashboards) alongside Flyte workflow executions.

### Goals

- Deploy long-running containerized apps (FastAPI, Flask, vLLM, Streamlit, etc.) via the Flyte control plane
- Autoscale apps based on request rate or concurrency
- Provide ingress (public URLs) for deployed apps
- Manage app lifecycle: create, update, scale, stop, delete
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
│  │ AppService ───────┼──┤ ├── Watch  (TaskAction)  │  │ CRs        │ │
│  │ (CRUD + DB)       │  │ ├── Abort  (TaskAction)  │  │            │ │
│  │                   │  │ │                        │  │            │ │
│  │ Create ──Deploy──>│  │ AppK8sClient (NEW)       │  │            │ │
│  │ Update ──Deploy──>│  │ ├── Deploy (KService)    │  │            │ │
│  │ Delete ──Undeploy>│  │ ├── Undeploy (KService)  │  │            │ │
│  │ UpdateStatus <────┼──┤ ├── Watch  (KService)    │  │            │ │
│  │                   │  │ └── Replicas (Pods)      │  │            │ │
│  └──────────┬────────┘  └────────────┬─────────────┘  └──────┬─────┘ │
│             │                        │                       │       │
│             ▼                        ▼                       ▼       │
│      ┌────────────┐         ┌──────────────────────────────────┐     │
│      │ PostgreSQL/ │         │ Kubernetes                       │     │
│      │ SQLite      │         │ ┌──────────────┐ ┌────────────┐ │     │
│      │ (apps table)│         │ │ TaskAction CRs│ │ KService   │ │     │
│      └────────────┘         │ │ (tasks)       │ │ CRs (apps) │ │     │
│                              │ └──────────────┘ └────────────┘ │     │
│                              └──────────────────────────────────┘     │
└──────────────────────────────────────────────────────────────────────┘
```

**Key patterns to follow:**
- **SetupContext**: Services register handlers on `sc.Mux`, background workers via `sc.AddWorker`
- **buf connect**: gRPC handlers generated from proto definitions
- **GORM + gormigrate**: Database models with migration framework
- **Repository pattern**: `interfaces/` → `impl/` → `models/` → `transformers/`
- **Background workers**: Long-running goroutines for reconciliation

**Key architectural insight — ActionsService pattern:**
The ActionsService already establishes the pattern for bridging gRPC ↔ K8s CRs:
1. `Enqueue()` creates a TaskAction CR in Kubernetes
2. A K8s watcher monitors CR status changes
3. Status updates flow back to RunService via `InternalRunService`

**The App Service should follow this same pattern** — the ActionsService (or a new AppActionsService) directly creates/watches Knative Service CRs. The executor is not involved; it only handles TaskAction reconciliation.

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
| `Identifier` | org, project, domain, name |
| `Replica` | Individual pod/instance info |
| `AutoscalingConfig` | min/max replicas, scaling metric (RPS or concurrency) |
| `IngressConfig` | private flag, subdomain, cname |

**DeploymentStatus state machine (single-cluster, no lease):**
```
PENDING → DEPLOYING → ACTIVE
                    → FAILED
                    → SCALING_UP → ACTIVE
                    → SCALING_DOWN → ACTIVE
ACTIVE → STOPPED (user sets desired_state=STOPPED)
STOPPED → PENDING (user sets desired_state=ACTIVE)
Any → FAILED (on error)
```

> For single-cluster OSS, we skip UNASSIGNED and ASSIGNED entirely — no lease
> negotiation is needed. AppService creates the app and immediately calls
> AppK8sClient.Deploy(), which creates the KService. The app starts in PENDING.

### 3.2 Database Model

New file: `runs/repository/models/app.go`

```go
package models

import (
    "database/sql"
    "time"
)

type App struct {
    ID uint `gorm:"primaryKey"`

    // App Identifier (unique composite key)
    Org     string `gorm:"not null;uniqueIndex:idx_apps_identifier,priority:1"`
    Project string `gorm:"not null;uniqueIndex:idx_apps_identifier,priority:2"`
    Domain  string `gorm:"not null;uniqueIndex:idx_apps_identifier,priority:3"`
    Name    string `gorm:"not null;uniqueIndex:idx_apps_identifier,priority:4"`

    // Serialized protobuf (flyteidl2.app.Spec)
    Spec []byte `gorm:"type:bytea;not null"`
    // Serialized protobuf (flyteidl2.app.Status)
    Status []byte `gorm:"type:bytea;not null"`

    // Denormalized for queries and optimistic locking
    DesiredState int32  `gorm:"not null;default:0;index:idx_apps_desired_state"`
    Phase        int32  `gorm:"not null;default:0;index:idx_apps_phase"`
    Revision     uint64 `gorm:"not null;default:1"`
    SpecDigest   []byte `gorm:"type:bytea"`

    // Denormalized for ingress lookups
    PublicHost string `gorm:"uniqueIndex:idx_apps_public_host"`

    // Autoscaling state
    CurrentReplicas uint32 `gorm:"not null;default:0"`

    // Metadata
    CreatedBy string
    CreatedAt time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_apps_created"`
    UpdatedAt time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP"`
    DeletedAt sql.NullTime `gorm:"index:idx_apps_deleted"`
}

func (App) TableName() string { return "apps" }
```

**Design decisions:**
- Spec and Status stored as serialized protobuf bytes (same pattern as `Action.ActionSpec` / `Action.ActionDetails`)
- `DesiredState` and `Phase` denormalized as integers for efficient queries and indexing
- `SpecDigest` enables change detection without deserializing the full spec
- `Revision` for optimistic concurrency control
- `PublicHost` unique index for ingress-based lookups
- Soft delete via `DeletedAt`

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
    Get(ctx context.Context, org, project, domain, name string) (*models.App, error)
    GetByHost(ctx context.Context, host string) (*models.App, error)
    Update(ctx context.Context, oldRevision uint64, app *models.App) (*models.App, error)
    UpdateStatus(ctx context.Context, oldRevision uint64, app *models.App) (*models.App, error)
    Delete(ctx context.Context, org, project, domain, name string) error
    List(ctx context.Context, filters ListFilters, limit int, token string) ([]*models.App, string, error)
    ListByPhase(ctx context.Context, phases []int32) ([]*models.App, error)
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
- **UpdateStatus**: `UPDATE ... WHERE revision = ?` (status-only, no spec_digest check)
- **Get**: Lookup by composite key or by `PublicHost`
- **List**: Filter by org, project, phase; paginate by `updated_at` cursor
- **Delete**: Soft delete (set `DeletedAt`)

---

## 5. Service Layer

### 5.1 Architecture: AppService ↔ AppK8sClient

```
┌─────────────────────────────────────────────────────────────────┐
│  Manager (unified binary)                                        │
│                                                                  │
│  ┌──────────────────────┐        ┌────────────────────────────┐ │
│  │ RunService process    │        │ ActionsService process     │ │
│  │                      │        │                            │ │
│  │  AppService          │        │  ActionsService (tasks)    │ │
│  │  ├── AppRepo (DB)    │        │  ├── ActionsClient (K8s)   │ │
│  │  ├── AppNotifier     │        │  │                         │ │
│  │  └── → Deploy() ─────┼───────>│  AppK8sClient (NEW)       │ │
│  │       → Undeploy()   │  HTTP  │  ├── Create/Update KService│ │
│  │                      │  call  │  ├── Watch KService status │ │
│  │  ← UpdateStatus() ◄──┼───────┤  ├── → UpdateStatus() ─────┤ │
│  │                      │        │  └── Pod queries (replicas)│ │
│  └──────────┬───────────┘        └─────────────┬──────────────┘ │
│             │                                   │                │
│             ▼                                   ▼                │
│      ┌────────────┐                    ┌──────────────┐         │
│      │ PostgreSQL/ │                    │ Kubernetes    │         │
│      │ SQLite      │                    │ (KService CRs│         │
│      │ (apps table)│                    │  + Pods)     │         │
│      └────────────┘                    └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

This mirrors the existing pattern:
- **RunService → ActionsService.Enqueue() → creates TaskAction CR**
- **AppService → AppK8sClient.Deploy() → creates KService CR**

### 5.2 AppService Implementation

Replace the stub in `runs/service/app_service.go`:

```go
type AppService struct {
    appconnect.UnimplementedAppServiceHandler
    repo      interfaces.AppRepo
    deployer  AppDeployerInterface  // calls AppK8sClient via connect client
    cfg       *AppConfig
    notifier  *AppNotifier          // in-memory pub/sub for Watch
}
```

### 5.3 RPC Implementations

#### Create

```
User → Create(CreateRequest{app})
  1. Validate spec (container or pod payload required)
  2. Generate public host from subdomain pattern: "{name}-{project}-{domain}.{base_domain}"
  3. Set initial status: Phase=PENDING, Revision=1
  4. Compute spec_digest = SHA256(spec)
  5. Serialize spec + status to protobuf bytes
  6. repo.Create(app)
  7. deployer.Deploy(app)  → AppK8sClient creates KService in K8s
  8. Notify watchers (AppNotifier.Broadcast CreateEvent)
  9. Return created app
```

#### Get

```
User → Get(GetRequest{app_id | ingress})
  1. If app_id: repo.Get(org, project, domain, name)
  2. If ingress: repo.GetByHost(host)
  3. Return app
```

#### Update

```
User → Update(UpdateRequest{app, reason})
  1. Compute new spec_digest
  2. If spec changed (digest differs):
     a. Increment revision
     b. Append Condition: DEPLOYING with reason
  3. If only desired_state changed:
     a. Increment revision
     b. If STOPPED: append STOPPED condition
     c. If ACTIVE: append PENDING condition (triggers redeployment)
  4. repo.Update(oldRevision, app) — optimistic lock
  5. Notify watchers
  6. Return updated app
```

#### UpdateStatus

```
AppK8sClient → UpdateStatus(UpdateStatusRequest{app})
  1. Append condition to status.conditions (trim to max N)
  2. Update phase, current_replicas
  3. repo.UpdateStatus(oldRevision, app)
  4. Notify watchers
  5. Return updated app
```

#### Delete

```
User → Delete(DeleteRequest{app_id})
  1. Verify desired_state == STOPPED and phase is terminal
  2. repo.Delete(org, project, domain, name)
  3. Notify watchers (DeleteEvent)
```

#### List

```
User → List(ListRequest{filter_by, limit, token})
  1. Extract filter: org, project, or all
  2. repo.List(filters, limit, token)
  3. Return apps + next_page_token
```

#### Watch (server-streaming)

```
User → Watch(WatchRequest{target})
  1. Subscribe to AppNotifier for target scope (org/project/app_id)
  2. Send initial state (current apps matching target)
  3. Loop:
     a. Wait for notification on channel
     b. Send WatchResponse (CreateEvent | UpdateEvent | DeleteEvent)
  4. Unsubscribe on context cancel
```

#### Lease (server-streaming) — optional, for future multi-cluster

In single-cluster mode, the Lease RPC is not needed because AppK8sClient (co-located
in the ActionsService process) directly creates KServices on Create/Update calls.
The Lease RPC exists in the proto for future multi-cluster support where a remote
cluster operator would pull apps via streaming lease.

For now, the stub implementation is sufficient:

```go
func (s *AppService) Lease(...) error {
    return connect.NewError(connect.CodeUnimplemented,
        errors.New("Lease is not needed in single-cluster mode"))
}
```

### 5.4 AppNotifier (in-memory pub/sub)

New file: `runs/service/app_notifier.go`

```go
type AppNotifier struct {
    mu          sync.RWMutex
    subscribers map[string][]chan *flyteapp.WatchResponse  // key = scope string
}

func (n *AppNotifier) Subscribe(scope string) (<-chan *flyteapp.WatchResponse, func())
func (n *AppNotifier) Broadcast(scope string, event *flyteapp.WatchResponse)
```

Scope keys: `org:{org}`, `project:{org}/{project}`, `app:{org}/{project}/{domain}/{name}`

> For single-process (Manager) mode, in-memory channels are sufficient. For distributed
> deployment, this can later be backed by PostgreSQL LISTEN/NOTIFY.

---

## 6. K8s Integration — ActionsService Pattern (No Executor Changes)

### 6.1 Design Decision: ActionsService, Not Executor

The existing ActionsService already establishes the pattern for K8s CR lifecycle:

```
ActionsService today (for task execution):
  Enqueue() → creates TaskAction CR → K8s watcher monitors status → reports back to RunService

App deployment follows the SAME pattern:
  Deploy()  → creates KService CR  → K8s watcher monitors status → reports back to AppService
```

The **executor** reconciles TaskAction CRs for task execution — a completely different concern.
Apps should be handled by a new **AppK8sClient** inside the ActionsService (or a new sibling service),
which directly creates/watches Knative Service CRs, just like `ActionsClient` creates/watches TaskAction CRs.

### 6.2 Approach: Knative Service

Apps are deployed as **Knative Services** (KServices), which provide:
- Autoscaling (scale-to-zero, scale-from-zero)
- Revision management (blue-green deploys)
- Traffic routing
- Request timeout enforcement

**Dependency:** Knative Serving must be installed in the cluster.

### 6.3 AppK8sClient — KService Lifecycle Manager

New file: `actions/k8s/app_client.go`

Mirrors the existing `ActionsClient` pattern:

```go
type AppK8sClient struct {
    k8sClient   client.WithWatch    // same as ActionsClient
    k8sCache    cache.Cache
    namespace   string
    appClient   appconnect.AppServiceClient  // calls back to AppService
    subscribers map[string][]chan AppUpdate   // in-memory pub/sub
}

// Interface
type AppK8sClientInterface interface {
    // Deploy creates or updates a KService for the given app
    Deploy(ctx context.Context, app *flyteapp.App) error

    // Undeploy deletes the KService for the given app
    Undeploy(ctx context.Context, appID *flyteapp.Identifier) error

    // GetReplicas lists pods for an app
    GetReplicas(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Replica, error)

    // DeleteReplica deletes a specific pod
    DeleteReplica(ctx context.Context, replicaID *flyteapp.ReplicaIdentifier) error

    // StartWatching begins watching KService objects (background goroutine)
    StartWatching(ctx context.Context) error

    // StopWatching stops the watcher
    StopWatching()
}
```

**StartWatching()** — K8s watch on KService objects with label `flyte.org/app-managed=true`:

```
On KService event:
  1. Extract app identifier from annotation "flyte.org/app-id"
  2. Map KService status → App DeploymentStatus:
     - LatestCreated != LatestReady → DEPLOYING
     - Ready=True → ACTIVE
     - Ready=False with error → PENDING (with error message)
     - Being deleted → STOPPED
  3. Read replica count from KService status
  4. Call AppService.UpdateStatus() on the control plane
  5. Notify local subscribers (for Watch/Lease RPCs)
```

### 6.4 Deploy(): App Proto → KService

```
AppService.Create()
    │ persists to DB
    │ calls AppK8sClient.Deploy(app)
    ▼
AppK8sClient.Deploy(app):
    │
    ├── 1. Build PodSpec from App.Spec
    │      Container/K8sPod → core.TaskTemplate
    │      → flytek8s.ToK8sPodSpec() → corev1.PodSpec
    │      Post-process:
    │        ├── Add container ports from App.Spec
    │        ├── Add tolerations (flyte.org/node-role=worker)
    │        ├── Add identity labels (org/project/domain/name)
    │        └── Add env vars (FLYTE_APP_NAME, etc.)
    │
    ├── 2. Build KService manifest
    │      Name: "{org}-{project}-{domain}-{name}"
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
    └── Done (KService watcher handles status updates asynchronously)
```

### 6.5 Replica Server (Phase 4 — optional)

New file: `actions/k8s/app_replica.go`

**What it is:** A pod-level observability and management layer that lets users
inspect individual pods backing an app, without requiring direct `kubectl` access.

**RPCs:**
- `ListReplica` — show all pods for an app with per-pod status (running, pending, crash-looping)
- `GetReplica` — get status of a specific pod
- `DeleteReplica` — force-kill a specific pod (manual escape hatch)
- `WatchReplicas` — stream pod status changes (for UI dashboards)

**Why it exists:** Knative and Kubernetes already handle unhealthy pods automatically
via liveness/readiness probes and the autoscaler. The Replica Server is for cases
where the user needs visibility or manual control — e.g., a pod serving stale model
weights that still passes health checks, or debugging why a specific replica is slow.

**Priority:** Low. The core app lifecycle (Create → Deploy → ACTIVE) works without it.
Users can fall back to `kubectl` for pod-level debugging. Defer to Phase 4.

```go
func (c *AppK8sClient) GetReplicas(ctx, appID) ([]*flyteapp.Replica, error) {
    // List pods with label selector: org + project + domain + name
    // Convert pod status → Replica proto (reuse flytek8s.DemystifyPending)
}

func (c *AppK8sClient) DeleteReplica(ctx, replicaID) error {
    // Force-kill a specific pod; Knative auto-replaces it
}
```

### 6.6 App Deletion Flow

```
User calls Update(desired_state=STOPPED)
    │
    ▼
AppService.Update()
    │ Updates DB
    │ Calls AppK8sClient.Undeploy(appID)
    ▼
AppK8sClient.Undeploy()
    │ Deletes KService from cluster
    ▼
KService watcher sees deletion
    │ Calls AppService.UpdateStatus(STOPPED, replicas=0)
    ▼
User calls Delete()
    │ Verifies phase=STOPPED
    │ Soft-deletes from DB
    ▼
Done
```

### 6.7 Registration in ActionsService Setup

Modified: `actions/setup.go`

```go
func Setup(ctx context.Context, sc *app.SetupContext) error {
    // ... existing TaskAction setup ...

    // NEW: App K8s client (KService lifecycle)
    if cfg.Apps.Enabled {
        appServiceURL := cfg.Apps.AppServiceURL
        if sc.BaseURL != "" {
            appServiceURL = sc.BaseURL
        }
        appServiceClient := flyteappconnect.NewAppServiceClient(http.DefaultClient, appServiceURL)

        appK8sClient := actionsk8s.NewAppK8sClient(
            sc.K8sClient,
            sc.K8sCache,
            cfg.Apps.Namespace,  // default: "flyte-apps"
            appServiceClient,
        )
        if err := appK8sClient.StartWatching(ctx); err != nil {
            return fmt.Errorf("actions: failed to start KService watcher: %w", err)
        }
        sc.AddWorker("app-kservice-watcher", func(ctx context.Context) error {
            <-ctx.Done()
            appK8sClient.StopWatching()
            return nil
        })

        // Make appK8sClient available to AppService via context or dependency injection
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
│   │   └── app.go              # App DB model
│   └── transformers/
│       └── app.go              # Proto ↔ Model conversion
├── service/
│   ├── app_service.go          # Replace stub with full implementation
│   └── app_notifier.go         # In-memory pub/sub

actions/
├── k8s/
│   ├── app_client.go           # AppK8sClient — KService CRUD + watcher
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
├── setup.go                    # Wire real AppService with repo + notifier + appK8sClient

actions/
├── setup.go                    # Register AppK8sClient + KService watcher
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
appNotifier := service.NewAppNotifier()
appSvc := service.NewAppService(appRepo, appNotifier, appK8sClient, &service.AppConfig{
    PublicURLPattern: cfg.Apps.PublicURLPattern,
    MaxConditions:    cfg.Apps.MaxConditions,
})
```

> Note: `appK8sClient` is created in `actions/setup.go` and made available
> to the AppService either via `SetupContext` or by having AppService call
> the AppK8sClient through a connect client (same pattern as RunService → ActionsService).

### 7.4 ActionsService Setup (actions/setup.go)

```go
// Add to existing Setup():
if cfg.Apps.Enabled {
    appK8sClient := actionsk8s.NewAppK8sClient(sc.K8sClient, sc.K8sCache, ...)
    appK8sClient.StartWatching(ctx)
    sc.AddWorker("app-kservice-watcher", ...)
}
```

---

## 8. Configuration

### 8.1 Runs Service Config

Add to `runs/config/config.go`:

```go
type AppsConfig struct {
    // PublicURLPattern is a Go template for generating public URLs.
    // Variables: {{.Name}}, {{.Project}}, {{.Domain}}, {{.Org}}
    // Example: "https://{{.Name}}-{{.Project}}.apps.flyte.example.com"
    PublicURLPattern string `json:"publicUrlPattern" pflag:",URL pattern for app ingress"`

    // MaxConditions is the max number of conditions to keep in status history.
    MaxConditions int `json:"maxConditions" pflag:",Max status conditions to retain" default:"50"`
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

    // AppServiceURL is the URL of the AppService (for status callbacks).
    AppServiceURL string `json:"appServiceUrl" pflag:",AppService URL for status updates"`

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
SDK/CLI                    AppService             AppK8sClient (ActionsService)
   │                          │                       │
   │  Create(app_spec)        │                       │
   │─────────────────────────>│                       │
   │                          │ Validate spec         │
   │                          │ Generate public host  │
   │                          │ INSERT to DB          │
   │                          │ Deploy(app)           │
   │                          │──────────────────────>│
   │                          │                       │ Build PodSpec
   │                          │                       │ Create KService
   │  CreateResponse(app)     │                       │
   │<─────────────────────────│                       │
   │                          │                       │
   │                          │   (Knative deploys)   │
   │                          │                       │
   │                          │                       │ KService watcher
   │                          │                       │ sees Ready=True
   │                          │  UpdateStatus(ACTIVE) │
   │                          │<──────────────────────│
   │                          │                       │
   │  Watch() stream          │                       │
   │─────────────────────────>│                       │
   │  UpdateEvent(ACTIVE)     │                       │
   │<─────────────────────────│                       │
   │                          │                       │
   │  App is live at:         │                       │
   │  https://myapp-proj.apps.flyte.example.com       │
```

### 9.2 Update (Spec Change)

```
SDK/CLI                    AppService             AppK8sClient
   │                          │                       │
   │  Update(new_spec)        │                       │
   │─────────────────────────>│                       │
   │                          │ Compute new spec_digest│
   │                          │ Increment revision     │
   │                          │ UPDATE DB (optimistic  │
   │                          │ lock)                  │
   │                          │ Deploy(updated_app)    │
   │                          │──────────────────────>│
   │                          │                       │ Compare spec-sha
   │                          │                       │ Update KService
   │                          │                       │ (Knative creates
   │                          │                       │  new Revision)
   │                          │  UpdateStatus         │
   │                          │  (DEPLOYING)          │
   │                          │<──────────────────────│
   │                          │                       │
   │                          │  UpdateStatus(ACTIVE) │
   │                          │<──────────────────────│
```

### 9.3 Scale Event

```
                           AppService             AppK8sClient
                              │                       │
  (traffic increases)         │                       │
                              │                       │ Knative autoscaler
                              │                       │ adds replicas
                              │                       │
                              │                       │ KService watcher
                              │                       │ detects replica change
                              │  UpdateStatus         │
                              │  (SCALING_UP, n→m)    │
                              │<──────────────────────│
                              │                       │
                              │                       │ actual == desired
                              │  UpdateStatus         │
                              │  (ACTIVE, replicas=m) │
                              │<──────────────────────│
```

---

## 10. Phased Implementation Plan

### Phase 1: Control Plane (CRUD + DB)

1. Add `App` model to `runs/repository/models/app.go`
2. Add migration to `runs/migrations/migrations.go`
3. Create `AppRepo` interface and GORM implementation
4. Create `app_notifier.go` (in-memory pub/sub)
5. Implement `AppService` RPCs: Create, Get, Update, Delete, List
6. Wire in `runs/setup.go` and `runs/repository/repository.go`
7. **Unit tests** for repo + service

### Phase 2: K8s Integration (AppK8sClient in ActionsService)

1. Create `AppK8sClient` in `actions/k8s/app_client.go` (KService CRUD)
2. Create KService watcher (watches labeled KServices, reports status back)
3. Create `app_replica.go` in `actions/k8s/` (pod queries)
4. Register in `actions/setup.go`
5. Wire AppService → AppK8sClient (Deploy/Undeploy calls)
6. **Integration tests** with envtest + Knative CRDs

### Phase 3: Streaming RPCs + Status Loop

1. Implement `Watch` RPC (server-streaming with AppNotifier)
2. Implement `Lease` RPC (server-streaming, for future multi-cluster)
3. Implement `UpdateStatus` RPC (called by AppK8sClient watcher)
4. **End-to-end tests** with Knative in kind/minikube

### Phase 4: Polish

1. Log streaming (`AppLogsService.TailLogs` → K8s pod logs)
2. Ingress configuration (Knative domain mapping or Istio VirtualService)
3. Helm chart updates (`charts/flyte-binary/`, `charts/flyte-sandbox/`)
4. SDK integration (Python `union.app()` decorator)
5. Documentation

---

## 11. Testing Strategy

| Layer | Test Type | Tool |
|-------|-----------|------|
| Repository | Unit | GORM + SQLite in-memory |
| AppService | Unit | Mock repo + mock AppK8sClient (mockery) |
| Notifier | Unit | Goroutine + channel assertions |
| AppK8sClient | Integration | envtest (controller-runtime) + Knative CRDs |
| E2E | Integration | kind + Knative + flyte-sandbox |

---

## 12. Open Questions

1. **Knative dependency** — Should we support a non-Knative fallback (raw Deployment + HPA + Ingress) for clusters without Knative? This adds complexity but broadens adoption.

2. **Namespace strategy** — One namespace per app vs. shared namespace (`flyte-apps`)? Another option is `org-project-domain` namespaces for isolation.

3. **Ingress controller** — Knative typically uses Kourier or Istio. Should we prescribe one, or be agnostic?

4. **Auth for app endpoints** — How should app endpoints be authenticated? Knative doesn't provide auth out of the box. Options: Istio auth policy, app-level API keys, or proxy sidecar.

5. **Scale-to-zero** — Knative supports scale-to-zero. Should we expose `min_replicas: 0` as a valid option, or require at least 1?

6. **Resource quotas** — Should apps share the same resource pool as workflow executions, or have a separate quota?

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
