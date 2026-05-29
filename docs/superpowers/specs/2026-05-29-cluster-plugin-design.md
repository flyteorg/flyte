# ClusterPlugin Interface + Ray Migration — Design

Date: 2026-05-29
Status: Approved (pending spec review)

## Problem

The current k8s plugin model (`flyteplugins/.../pluginmachinery/k8s/plugin.go`,
`Plugin` interface) assumes **one task → one k8s resource**. `BuildResource`
returns a single object, the `PluginManager`
(`executor/pkg/plugin/k8s/plugin_manager.go`) creates it, watches it, and tears
it down when the task execution ends.

For Ray today (`flyteplugins/.../plugins/k8s/ray/ray.go`) this means every Ray
task produces a `RayJob` with an **embedded** `RayClusterSpec`. KubeRay spins up
a brand-new Ray cluster per task, runs the job, and tears the cluster down. Two
Ray tasks with identical cluster configuration cannot share a cluster, so we pay
full cluster cold-start cost on every task.

We want a plugin shape where:

1. A plugin declares **two** resources: a long-lived *cluster* and a *job* that
   runs on top of it.
2. The framework creates the cluster only if it does not already exist, waits
   for it to be ready, then submits the job against it.
3. The cluster is named by the **hash of the job spec proto**, so two tasks with
   the same Ray config deterministically target the *same* cluster.

## Goals

- Add a `ClusterPlugin` interface alongside the existing `Plugin` interface.
- Add a `ClusterPluginManager` that drives the create-or-reuse-cluster → wait →
  submit-job → watch-job state machine, implementing `pluginsCore.Plugin`.
- Migrate the Ray plugin to `ClusterPlugin`: a standalone `RayCluster` named
  `<hash(RayJob proto)>` plus a `RayJob` bound to it via `ClusterSelector`.
- Preserve the existing `Plugin` path and all other plugins unchanged.

## Non-Goals

- Reference-counting / explicit cluster deletion. Shared clusters are reaped by
  KubeRay idle timeout (autoscaler `IdleTimeoutSeconds`); the plugin never
  deletes a cluster. (Decision: "Idle TTL, no owner ref".)
- Changing the `flyteidl2` Ray proto. The hash is computed over the existing
  `plugins.RayJob` proto unmarshalled from the task template.
- Multi-cluster scheduling/placement policy beyond "same spec → same cluster".

## Key facts established during exploration

- `PluginManager.Handle` is a 2-phase state machine persisted via
  `PluginStateReader/Writer`: `PluginPhaseNotStarted` → `launchResource`
  (`BuildResource` + `Create`) → `PluginPhaseStarted` → `checkResourcePhase`
  (`BuildIdentityResource` + `Get` + `GetTaskPhase`). Object name is always
  `taskCtx.GetTaskExecutionID().GetGeneratedName()`, sanitized to DNS1123.
- Plugins register via `pluginmachinery.PluginRegistry().RegisterK8sPlugin(k8s.PluginEntry{...})`.
  At startup `executor/pkg/plugin/registry.go::Registry.Initialize` wraps every
  `k8s.PluginEntry` in `executorK8s.NewPluginManager(id, plugin, kubeClient)` and
  maps task types → manager.
- KubeRay **v1.5.2** (go.mod) supports `RayJobSpec.ClusterSelector`
  (`map[string]string`, key `ray.io/cluster` =
  `utils.RayClusterLabelKey`/`RayJobClusterSelectorKey`). When set, the RayJob
  controller skips cluster creation and looks the cluster up by namespaced name,
  requeuing until it exists (`rayjob_controller.go` ~L460-480). `RayClusterSpec`
  must be left nil in that case.
- `flytestdlib/pbhash.ComputeHashString(ctx, msg)` returns a deterministic
  sha256 (base64 RawURL) of a proto — basis for the cluster name.

## Design

### 1. `ClusterPlugin` interface (new)

In `flyteplugins/go/tasks/pluginmachinery/k8s/plugin.go`:

```go
// ClusterPlugin authors plugins that run a job on top of a shared, long-lived
// cluster resource. The cluster is created once (keyed by GetClusterName) and
// reused by every task whose spec hashes to the same name; the job is created
// only after the cluster is ready.
type ClusterPlugin interface {
    // GetClusterName returns a deterministic name for the cluster backing this
    // task. Implementations typically hash the task's plugin-spec proto so that
    // identical specs collapse onto the same cluster. The framework sanitizes
    // the result to a DNS1123 subdomain.
    GetClusterName(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (string, error)

    // BuildClusterResource builds the cluster object (e.g. a RayCluster). It is
    // created without an owner reference or finalizer so that it outlives any
    // single task execution. Name/namespace are set by the framework.
    BuildClusterResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error)

    // BuildClusterIdentityResource returns a minimal object (Type/Object meta)
    // used to GET the cluster for readiness checks.
    BuildClusterIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error)

    // IsClusterReady reports whether the cluster can accept jobs.
    IsClusterReady(ctx context.Context, pluginContext PluginContext, cluster client.Object) (bool, error)

    // BuildJobResource builds the job object bound to the (ready) cluster
    // identified by clusterName — e.g. a RayJob with ClusterSelector set. Unlike
    // the cluster, the job gets the normal owner reference / finalizer.
    BuildJobResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, clusterName string) (client.Object, error)

    // BuildJobIdentityResource returns a minimal object used to GET the job.
    BuildJobIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error)

    // GetJobPhase analyses the job resource and reports task phase (analogous to
    // Plugin.GetTaskPhase).
    GetJobPhase(ctx context.Context, pluginContext PluginContext, job client.Object) (pluginsCore.PhaseInfo, error)

    // GetProperties mirrors Plugin.GetProperties.
    GetProperties() PluginProperties
}
```

Notes:
- `GarbageCollectable` is **not** required: clusters are not GC'd by Flyte and
  jobs are owned via owner-reference, so the external GC continues to watch the
  job resource type.
- The cluster object name comes from `GetClusterName` (sanitized), **not** from
  `GetGeneratedName()` — that is the whole point of sharing.

### 2. `PluginEntry` extension

`k8s.PluginEntry` gains optional fields so a single registration carries both
watched types and the cluster plugin:

```go
type PluginEntry struct {
    // ... existing fields ...

    // ClusterPlugin, when set, indicates this entry is driven by a
    // ClusterPluginManager instead of a PluginManager. Plugin must be nil.
    ClusterPlugin ClusterPlugin
    // ClusterResourceToWatch is the cluster CRD type (e.g. &rayv1.RayCluster{}).
    ClusterResourceToWatch client.Object
    // ResourceToWatch remains the *job* type (e.g. &rayv1.RayJob{}).
}
```

`RegisterK8sPlugin` validation: exactly one of `Plugin` / `ClusterPlugin` must be
non-nil; if `ClusterPlugin` is set, both `ResourceToWatch` (job) and
`ClusterResourceToWatch` must be set.

### 3. `ClusterPluginManager` (new) — state machine

New file `executor/pkg/plugin/k8s/cluster_plugin_manager.go`, implementing
`pluginsCore.Plugin` (GetID/GetProperties/Handle/Abort/Finalize). Persisted
state extends the phase enum:

```
ClusterPhaseNotStarted   -> ensure cluster
ClusterPhaseClusterWait  -> cluster created, waiting for readiness
ClusterPhaseJobStarted   -> job submitted, watching job
```

`Handle` logic per round:

1. **NotStarted**
   - `clusterName = sanitize(GetClusterName(...))`
   - GET cluster by `{namespace, clusterName}`.
     - If not found: `BuildClusterResource`, set name=clusterName,
       namespace, labels (incl. `ray.io/cluster=clusterName`), **no owner ref,
       no finalizer** (force `DisableInjectOwnerReferences`/`DisableInjectFinalizer`
       for the cluster object regardless of config), `Create`
       (ignore `IsAlreadyExists` — handles the race where two tasks create the
       same cluster concurrently).
   - Persist `clusterName` + phase `ClusterWait`, return `PhaseInitializing`
     ("cluster is creating").
2. **ClusterWait**
   - GET cluster; if `IsClusterReady` false → stay `PhaseInitializing`.
   - If ready → `BuildJobResource(ctx, tCtx, clusterName)`, apply normal
     `addObjectMetadata` (job name = `GetGeneratedName()`, **with** owner
     ref/finalizer), `Create` (ignore `IsAlreadyExists`). Persist phase
     `JobStarted`, return `PhaseQueued`.
3. **JobStarted**
   - GET job via `BuildJobIdentityResource`; `GetJobPhase`; reuse the existing
     output-reading / error / phase-version logic from
     `PluginManager.checkResourcePhase` (factor the shared tail into a helper so
     both managers use it).

`Abort` / `Finalize`: act on the **job** resource only (delete/clear finalizers
exactly like `PluginManager`). The cluster is intentionally left alone.

`addObjectMetadata` is refactored to take an option controlling
name + owner/finalizer injection so the cluster path can reuse it
(`addClusterObjectMetadata` sets name=clusterName, no owner ref, no finalizer).

### 4. Registry wiring

`executor/pkg/plugin/registry.go::Initialize`: for each `k8s.PluginEntry`, if
`entry.ClusterPlugin != nil` construct `executorK8s.NewClusterPluginManager(
entry.ID, entry.ClusterPlugin, kubeClient)`; else the existing
`NewPluginManager`. Both implement `pluginsCore.Plugin`, so the rest of the map
wiring (task-type → plugin, default, event watcher) is unchanged. Event watcher
init covers both the job and cluster types.

### 5. Ray migration

`flyteplugins/.../plugins/k8s/ray/ray.go`:

- `rayJobResourceHandler` implements `ClusterPlugin` instead of `Plugin`.
- `GetClusterName`: unmarshal `plugins.RayJob` from the task template, then
  `name = "ray-" + sanitize(pbhash.ComputeHashString(ctx, &rayJob)[:16])`.
  (Truncate + lowercase via `ConvertToDNS1123SubdomainCompatibleString`; 16 hex/
  base32 chars keeps collision probability negligible while staying within the
  47-char limit and leaving room for KubeRay's pod-name suffixes.) The hash is
  computed over the *proto* so it is independent of task execution id.
- `BuildClusterResource`: the current `constructRayJob` cluster-building logic,
  but returns a standalone `*rayv1.RayCluster` (TypeMeta `RayCluster`) carrying
  `rayClusterSpec`. Set label `ray.io/cluster=<clusterName>` (so the selector
  matches) and enable autoscaling + `AutoscalerOptions.IdleTimeoutSeconds` from
  config for idle teardown.
- `BuildClusterIdentityResource`: minimal `&rayv1.RayCluster{TypeMeta:...}`.
- `IsClusterReady`: cluster ready when
  `RayCluster.Status.State == rayv1.Ready` (head pod up). Keep the existing
  head-pod readiness helper as a fallback signal.
- `BuildJobResource`: a `*rayv1.RayJob` with `Spec.RayClusterSpec = nil`,
  `Spec.ClusterSelector = {ray.io/cluster: clusterName}`, `Entrypoint`,
  `RuntimeEnvYAML`, `SubmitterPodTemplate`, and `ShutdownAfterJobFinishes=false`
  (cluster must survive the job for reuse).
- `BuildJobIdentityResource`: current `BuildIdentityResource` (minimal RayJob).
- `GetJobPhase`: the current `GetTaskPhase` body (status mapping, logs,
  dashboard), unchanged.
- `init()`: register with `ClusterPlugin: rayJobResourceHandler{}`,
  `ResourceToWatch: &rayv1.RayJob{}`,
  `ClusterResourceToWatch: &rayv1.RayCluster{}`.
- Add `rayv1.RayCluster` to the scheme (already via `rayv1.AddToScheme`).

## Error handling

- Cluster `Create` racing another task: ignore `IsAlreadyExists` (same as job
  path today).
- Cluster vanished while in `JobStarted` (idle-reaped mid-job): KubeRay marks the
  RayJob failed; `GetJobPhase` surfaces it as a (retryable) failure — on retry
  the manager re-enters `NotStarted` and recreates the cluster.
- `GetClusterName` error → `PhaseInfoFailure("BadTaskDefinition", ...)`.
- Cluster never becomes ready: bounded by the task's own timeout; manager keeps
  returning `PhaseInitializing`.

## Testing

New `executor/pkg/plugin/k8s/cluster_plugin_manager_test.go` (currently no test
exists for plugin_manager.go — add a fake-client harness):

- NotStarted creates cluster with name=hash, no owner ref, no finalizer.
- Second task with same spec → cluster already exists → no duplicate create,
  proceeds to wait.
- ClusterWait stays Initializing until `IsClusterReady`, then creates job with
  owner ref + `ClusterSelector`.
- JobStarted maps job phases (success/failure/running) and reads outputs.
- Abort/Finalize delete only the job, never the cluster.
- `IsAlreadyExists` on concurrent cluster create is swallowed.

Ray unit tests (`ray_test.go`):
- `GetClusterName` deterministic + identical for two identical specs, different
  for differing specs.
- `BuildClusterResource` returns RayCluster with correct label + autoscaling/idle
  timeout; `BuildJobResource` returns RayJob with nil RayClusterSpec and correct
  ClusterSelector.
- Adapt existing `constructRayJob` assertions to the split.

## Files touched

- `flyteplugins/go/tasks/pluginmachinery/k8s/plugin.go` — `ClusterPlugin`,
  `PluginEntry` fields.
- `flyteplugins/go/tasks/pluginmachinery/registry.go` — `RegisterK8sPlugin`
  validation.
- `executor/pkg/plugin/k8s/cluster_plugin_manager.go` — new manager.
- `executor/pkg/plugin/k8s/plugin_manager.go` — refactor `addObjectMetadata` +
  extract shared success/output tail.
- `executor/pkg/plugin/registry.go` — construct the right manager.
- `flyteplugins/go/tasks/plugins/k8s/ray/ray.go` — migrate to `ClusterPlugin`.
- Tests as above.
