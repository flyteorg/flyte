# ClusterPlugin Interface + Ray Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `ClusterPlugin` interface (with `BuildClusterResource` + `BuildJobResource`) that creates a shared, long-lived cluster keyed by the hash of the job-spec proto, runs jobs on top of it, and migrate the Ray plugin to use it so identical Ray configs share one cluster.

**Architecture:** A new `ClusterPlugin` interface lives next to the existing `Plugin` in `pluginmachinery/k8s`. A new `ClusterPluginManager` (sibling to `PluginManager`) drives a 3-phase state machine — *ensure cluster → wait for ready → submit & watch job* — and implements `pluginsCore.Plugin`. The cluster object is named by `GetClusterName` (a proto hash), created without owner-ref/finalizer so it outlives any single task, and reaped by KubeRay's idle autoscaler. The job carries the normal owner-ref and binds to the cluster via KubeRay's `ClusterSelector`. Registration is extended so one `k8s.PluginEntry` can carry either a `Plugin` or a `ClusterPlugin`.

**Tech Stack:** Go, Kubernetes controller-runtime, KubeRay v1.5.2 (`rayv1`), Flyte plugin machinery, `flytestdlib/pbhash`, testify + controller-runtime `fake` client.

---

## File Structure

- `flyteplugins/go/tasks/pluginmachinery/k8s/plugin.go` — **modify**: add `ClusterPlugin` interface + new `PluginEntry` fields (`ClusterPlugin`, `ClusterResourceToWatch`).
- `flyteplugins/go/tasks/pluginmachinery/registry.go` — **modify**: validate the new `PluginEntry` shape in `RegisterK8sPlugin`.
- `flyteplugins/go/tasks/pluginmachinery/registry_test.go` — **create/modify**: validation tests.
- `executor/pkg/plugin/k8s/plugin_manager.go` — **modify**: extract reusable helpers (`addObjectMetadataWithName`, `finishSuccessfulTask`) so both managers share them.
- `executor/pkg/plugin/k8s/cluster_plugin_manager.go` — **create**: the `ClusterPluginManager` state machine.
- `executor/pkg/plugin/k8s/cluster_plugin_manager_test.go` — **create**: state-machine tests with a fake client.
- `executor/pkg/plugin/registry.go` — **modify**: construct `ClusterPluginManager` when `entry.ClusterPlugin != nil`.
- `flyteplugins/go/tasks/plugins/k8s/ray/ray.go` — **modify**: migrate `rayJobResourceHandler` from `Plugin` to `ClusterPlugin`.
- `flyteplugins/go/tasks/plugins/k8s/ray/ray_test.go` — **modify**: split tests for cluster vs job, add `GetClusterName` determinism tests.

---

## Task 1: Define the `ClusterPlugin` interface

**Files:**
- Modify: `flyteplugins/go/tasks/pluginmachinery/k8s/plugin.go` (after the `Plugin` interface, ~line 103)

- [ ] **Step 1: Add the `ClusterPlugin` interface**

Add this block immediately after the existing `Plugin` interface definition (after the closing brace at ~line 103):

```go
// ClusterPlugin is a simplified interface to author plugins that run a job on top of a shared,
// long-lived cluster resource. The cluster is created once (keyed by GetClusterName) and reused by
// every task whose spec hashes to the same name; the job is created only after the cluster is ready.
//
// Unlike Plugin, the cluster resource is intentionally NOT owned by the task execution: it has no
// owner reference and no finalizer, so completing or aborting one task never deletes a cluster that
// other tasks may still be using. Cleanup of idle clusters is delegated to the underlying operator
// (e.g. KubeRay's idle autoscaler). The job resource, by contrast, is owned normally.
type ClusterPlugin interface {
	// GetClusterName returns a deterministic name for the cluster backing this task. Implementations
	// typically hash the task's plugin-spec proto so that identical specs collapse onto the same
	// cluster. The framework sanitizes the result into a DNS1123 subdomain before use.
	GetClusterName(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (string, error)

	// BuildClusterResource builds the full cluster object (e.g. a RayCluster) that will be posted to
	// k8s. Name and namespace are assigned by the framework from GetClusterName.
	BuildClusterResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error)

	// BuildClusterIdentityResource builds a query object (type/object meta only) used to GET the
	// cluster for readiness checks.
	BuildClusterIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error)

	// IsClusterReady reports whether the cluster can accept jobs.
	IsClusterReady(ctx context.Context, pluginContext PluginContext, cluster client.Object) (bool, error)

	// BuildJobResource builds the full job object (e.g. a RayJob) bound to the ready cluster
	// identified by clusterName. The job is created with the normal owner reference / finalizer.
	BuildJobResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, clusterName string) (client.Object, error)

	// BuildJobIdentityResource builds a query object (type/object meta only) used to GET the job.
	BuildJobIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error)

	// GetJobPhase analyses the job resource and reports the status as a TaskPhase. This is analogous
	// to Plugin.GetTaskPhase and is expected to be relatively fast.
	GetJobPhase(ctx context.Context, pluginContext PluginContext, job client.Object) (pluginsCore.PhaseInfo, error)

	// GetProperties returns properties desired by the plugin (mirrors Plugin.GetProperties).
	GetProperties() PluginProperties
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/kevin/git/flyte && go build ./flyteplugins/go/tasks/pluginmachinery/k8s/...`
Expected: builds with no errors (no consumers yet).

- [ ] **Step 3: Commit**

```bash
git add flyteplugins/go/tasks/pluginmachinery/k8s/plugin.go
git commit -s -m "feat(k8s): add ClusterPlugin interface for shared-cluster plugins"
```

---

## Task 2: Extend `PluginEntry` with cluster fields

**Files:**
- Modify: `flyteplugins/go/tasks/pluginmachinery/k8s/plugin.go:14-30` (the `PluginEntry` struct)

- [ ] **Step 1: Add the new fields**

In the `PluginEntry` struct, add these fields after the existing `Plugin Plugin` field:

```go
	// ClusterPlugin, when non-nil, indicates this entry is driven by a ClusterPluginManager instead
	// of a PluginManager. Exactly one of Plugin or ClusterPlugin must be set. When ClusterPlugin is
	// set, ResourceToWatch must be the job resource type and ClusterResourceToWatch must be the
	// cluster resource type.
	ClusterPlugin ClusterPlugin
	// ClusterResourceToWatch is the cluster CRD type (e.g. &rayv1.RayCluster{}). Required when
	// ClusterPlugin is set, ignored otherwise.
	ClusterResourceToWatch client.Object
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/kevin/git/flyte && go build ./flyteplugins/go/tasks/pluginmachinery/k8s/...`
Expected: builds with no errors.

- [ ] **Step 3: Commit**

```bash
git add flyteplugins/go/tasks/pluginmachinery/k8s/plugin.go
git commit -s -m "feat(k8s): add ClusterPlugin/ClusterResourceToWatch to PluginEntry"
```

---

## Task 3: Validate the new `PluginEntry` shape in the registry

**Files:**
- Modify: `flyteplugins/go/tasks/pluginmachinery/registry.go:52-72` (`RegisterK8sPlugin`)
- Test: `flyteplugins/go/tasks/pluginmachinery/registry_test.go`

- [ ] **Step 1: Write the failing test**

Create or append to `flyteplugins/go/tasks/pluginmachinery/registry_test.go`:

```go
package pluginmachinery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
)

func TestRegisterK8sPlugin_ClusterPluginValidation(t *testing.T) {
	t.Run("cluster plugin requires ClusterResourceToWatch", func(t *testing.T) {
		reg := &taskPluginRegistry{}
		assert.Panics(t, func() {
			reg.RegisterK8sPlugin(k8s.PluginEntry{
				ID:                  "x",
				RegisteredTaskTypes: []pluginsCore.TaskType{"x"},
				ResourceToWatch:     &rayv1.RayJob{},
				ClusterPlugin:       stubClusterPlugin{},
				// ClusterResourceToWatch missing -> panic
			})
		})
	})

	t.Run("cannot set both Plugin and ClusterPlugin", func(t *testing.T) {
		reg := &taskPluginRegistry{}
		assert.Panics(t, func() {
			reg.RegisterK8sPlugin(k8s.PluginEntry{
				ID:                     "x",
				RegisteredTaskTypes:    []pluginsCore.TaskType{"x"},
				ResourceToWatch:        &rayv1.RayJob{},
				ClusterResourceToWatch: &rayv1.RayCluster{},
				Plugin:                 stubPlugin{},
				ClusterPlugin:          stubClusterPlugin{},
			})
		})
	})

	t.Run("valid cluster plugin registers", func(t *testing.T) {
		reg := &taskPluginRegistry{}
		assert.NotPanics(t, func() {
			reg.RegisterK8sPlugin(k8s.PluginEntry{
				ID:                     "x",
				RegisteredTaskTypes:    []pluginsCore.TaskType{"x"},
				ResourceToWatch:        &rayv1.RayJob{},
				ClusterResourceToWatch: &rayv1.RayCluster{},
				ClusterPlugin:          stubClusterPlugin{},
			})
		})
		assert.Len(t, reg.GetK8sPlugins(), 1)
	})
}
```

Add minimal stubs at the bottom of the same test file (they only need to satisfy the interfaces; methods can return zero values). If `stubPlugin` is tedious, instead reuse an existing fake from the package; otherwise define:

```go
type stubPlugin struct{ k8s.Plugin }
type stubClusterPlugin struct{ k8s.ClusterPlugin }
```

(Embedding the interface gives a nil-method stub that satisfies the type for these construction-only tests.)

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/kevin/git/flyte && go test ./flyteplugins/go/tasks/pluginmachinery/ -run TestRegisterK8sPlugin_ClusterPluginValidation -v`
Expected: FAIL — the "requires ClusterResourceToWatch" and "both set" cases do NOT panic yet.

- [ ] **Step 3: Implement validation**

In `RegisterK8sPlugin`, replace the existing nil/ResourceToWatch checks with logic that handles both shapes. Replace the block at lines 61-67 with:

```go
	if info.Plugin == nil && info.ClusterPlugin == nil {
		logger.Panicf(context.TODO(), "K8s plugin requires either Plugin or ClusterPlugin to be set")
	}

	if info.Plugin != nil && info.ClusterPlugin != nil {
		logger.Panicf(context.TODO(), "K8s plugin cannot set both Plugin and ClusterPlugin")
	}

	if info.ResourceToWatch == nil {
		logger.Panicf(context.TODO(), "The framework requires a K8s resource to watch, for valid plugin registration")
	}

	if info.ClusterPlugin != nil && info.ClusterResourceToWatch == nil {
		logger.Panicf(context.TODO(), "ClusterPlugin requires ClusterResourceToWatch to be set")
	}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/kevin/git/flyte && go test ./flyteplugins/go/tasks/pluginmachinery/ -run TestRegisterK8sPlugin_ClusterPluginValidation -v`
Expected: PASS (all three subtests).

- [ ] **Step 5: Commit**

```bash
git add flyteplugins/go/tasks/pluginmachinery/registry.go flyteplugins/go/tasks/pluginmachinery/registry_test.go
git commit -s -m "feat(k8s): validate ClusterPlugin registration shape"
```

---

## Task 4: Extract reusable helpers from `PluginManager`

The `ClusterPluginManager` needs (a) metadata stamping with a custom name and toggleable owner-ref/finalizer, and (b) the success/output-reading tail. Extract these from `PluginManager` so both managers share one implementation (DRY).

**Files:**
- Modify: `executor/pkg/plugin/k8s/plugin_manager.go`

- [ ] **Step 1: Add `addObjectMetadataWithName` and refactor `addObjectMetadata`**

Replace the existing `addObjectMetadata` (lines 79-97) with a version that delegates to a new helper:

```go
// addObjectMetadata stamps namespace/labels/annotations, the generated name, and (unless disabled)
// owner references and finalizers, on an object that IS owned by this task execution.
func (pm *PluginManager) addObjectMetadata(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig) {
	pm.addObjectMetadataWithName(taskCtx, o, cfg, taskCtx.GetTaskExecutionID().GetGeneratedName(), true)
}

// addObjectMetadataWithName is the general form. name overrides the object name; when injectOwnership
// is false, owner references and finalizers are never added (used for shared cluster resources that
// must outlive a single task execution).
func (pm *PluginManager) addObjectMetadataWithName(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig, name string, injectOwnership bool) {
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(pluginsUtils.UnionMaps(cfg.DefaultAnnotations, o.GetAnnotations(), pluginsUtils.CopyMap(taskCtx.GetAnnotations())))
	o.SetLabels(pluginsUtils.UnionMaps(cfg.DefaultLabels, o.GetLabels(), pluginsUtils.CopyMap(taskCtx.GetLabels())))
	o.SetName(name)

	if injectOwnership && !pm.plugin.GetProperties().DisableInjectOwnerReferences && !cfg.DisableInjectOwnerReferences {
		o.SetOwnerReferences([]metav1.OwnerReference{taskCtx.GetOwnerReference()})
	}

	if injectOwnership && cfg.InjectFinalizer && !pm.plugin.GetProperties().DisableInjectFinalizer {
		f := append(o.GetFinalizers(), "flyte/flytek8s")
		o.SetFinalizers(f)
	}

	if errs := validation.IsDNS1123Subdomain(o.GetName()); len(errs) > 0 {
		o.SetName(pluginsUtils.ConvertToDNS1123SubdomainCompatibleString(o.GetName()))
	}
}
```

Note: `addObjectMetadataWithName` references `pm.plugin.GetProperties()`. The `ClusterPluginManager` is a different type, so it will NOT call this method — it gets its own copy (Task 5). Keep this one on `*PluginManager` unchanged in behavior.

- [ ] **Step 2: Run existing manager build to confirm no behavior change**

Run: `cd /Users/kevin/git/flyte && go build ./executor/pkg/plugin/k8s/...`
Expected: builds clean. (`addObjectMetadata` still behaves identically: name = generated name, ownership injected.)

- [ ] **Step 3: Commit**

```bash
git add executor/pkg/plugin/k8s/plugin_manager.go
git commit -s -m "refactor(k8s): generalize addObjectMetadata for reuse by cluster manager"
```

---

## Task 5: Implement `ClusterPluginManager` — construction & phases

**Files:**
- Create: `executor/pkg/plugin/k8s/cluster_plugin_manager.go`

- [ ] **Step 1: Write the file skeleton (types, constructor, helpers)**

Create `executor/pkg/plugin/k8s/cluster_plugin_manager.go`:

```go
package k8s

import (
	"context"
	"fmt"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	pluginsUtils "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const clusterPluginStateVersion = 1

// ClusterPhase tracks the high-level phase of the ClusterPluginManager's state machine.
type ClusterPhase uint8

const (
	ClusterPhaseNotStarted ClusterPhase = iota
	ClusterPhaseClusterWait
	ClusterPhaseJobStarted
)

// ClusterPluginState is persisted between reconciliation rounds.
type ClusterPluginState struct {
	Phase          ClusterPhase
	ClusterName    string
	K8sPluginState k8s.PluginState
}

var _ pluginsCore.Plugin = &ClusterPluginManager{}

// ClusterPluginManager wraps a k8s.ClusterPlugin to implement pluginsCore.Plugin. It manages the
// lifecycle of creating a shared cluster, waiting for readiness, then submitting and watching a job.
type ClusterPluginManager struct {
	id         string
	plugin     k8s.ClusterPlugin
	kubeClient pluginsCore.KubeClient
}

// NewClusterPluginManager creates a ClusterPluginManager wrapping a k8s.ClusterPlugin.
func NewClusterPluginManager(id string, plugin k8s.ClusterPlugin, kubeClient pluginsCore.KubeClient) *ClusterPluginManager {
	return &ClusterPluginManager{id: id, plugin: plugin, kubeClient: kubeClient}
}

func (m *ClusterPluginManager) GetID() string { return m.id }

func (m *ClusterPluginManager) GetProperties() pluginsCore.PluginProperties {
	props := m.plugin.GetProperties()
	return pluginsCore.PluginProperties{GeneratedNameMaxLength: props.GeneratedNameMaxLength}
}

// sanitizeName ensures a name is a valid DNS1123 subdomain.
func sanitizeName(name string) string {
	if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
		return pluginsUtils.ConvertToDNS1123SubdomainCompatibleString(name)
	}
	return name
}

// addJobMetadata stamps the job object exactly like PluginManager does for an owned resource.
func (m *ClusterPluginManager) addJobMetadata(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig) {
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(pluginsUtils.UnionMaps(cfg.DefaultAnnotations, o.GetAnnotations(), pluginsUtils.CopyMap(taskCtx.GetAnnotations())))
	o.SetLabels(pluginsUtils.UnionMaps(cfg.DefaultLabels, o.GetLabels(), pluginsUtils.CopyMap(taskCtx.GetLabels())))
	o.SetName(taskCtx.GetTaskExecutionID().GetGeneratedName())

	if !m.plugin.GetProperties().DisableInjectOwnerReferences && !cfg.DisableInjectOwnerReferences {
		o.SetOwnerReferences([]metav1.OwnerReference{taskCtx.GetOwnerReference()})
	}
	if cfg.InjectFinalizer && !m.plugin.GetProperties().DisableInjectFinalizer {
		o.SetFinalizers(append(o.GetFinalizers(), "flyte/flytek8s"))
	}
	if errs := validation.IsDNS1123Subdomain(o.GetName()); len(errs) > 0 {
		o.SetName(pluginsUtils.ConvertToDNS1123SubdomainCompatibleString(o.GetName()))
	}
}

// addClusterMetadata stamps the cluster object with the deterministic shared name and NO owner ref /
// finalizer, so it survives beyond any single task execution.
func (m *ClusterPluginManager) addClusterMetadata(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig, clusterName string) {
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(pluginsUtils.UnionMaps(cfg.DefaultAnnotations, o.GetAnnotations(), pluginsUtils.CopyMap(taskCtx.GetAnnotations())))
	o.SetLabels(pluginsUtils.UnionMaps(cfg.DefaultLabels, o.GetLabels(), pluginsUtils.CopyMap(taskCtx.GetLabels())))
	o.SetName(sanitizeName(clusterName))
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/kevin/git/flyte && go build ./executor/pkg/plugin/k8s/...`
Expected: builds (unused imports `io`, `ioutils`, `core`, `errors`, `time`, `k8serrors`, `k8stypes`, `logger`, `fmt` will be used in the next step — if Go complains about unused imports now, temporarily proceed to Step 3 which uses them, then build).

- [ ] **Step 3: Add the `Handle` state machine + `Abort`/`Finalize`**

Append to the same file:

```go
// Handle drives the cluster->job state machine. Invoked every reconciliation round.
func (m *ClusterPluginManager) Handle(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	state := ClusterPluginState{}
	if v, err := tCtx.PluginStateReader().Get(&state); err != nil {
		if v != clusterPluginStateVersion {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure(errors.CorruptedPluginState,
				fmt.Sprintf("plugin state version mismatch expected [%d] got [%d]", clusterPluginStateVersion, v), nil)), nil
		}
		return pluginsCore.UnknownTransition, errors.Wrapf(errors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	cfg := config.GetK8sPluginConfig()
	var transition pluginsCore.Transition
	var err error
	newState := state

	switch state.Phase {
	case ClusterPhaseNotStarted:
		transition, newState, err = m.ensureCluster(ctx, tCtx, cfg)
	case ClusterPhaseClusterWait:
		transition, newState, err = m.waitForClusterThenLaunchJob(ctx, tCtx, cfg, state)
	default: // ClusterPhaseJobStarted
		transition, err = m.checkJobPhase(ctx, tCtx, &newState.K8sPluginState)
	}
	if err != nil {
		return transition, err
	}

	phaseInfo := transition.Info()
	newState.K8sPluginState = k8s.PluginState{
		Phase:        phaseInfo.Phase(),
		PhaseVersion: phaseInfo.Version(),
		Reason:       phaseInfo.Reason(),
	}
	if state != newState {
		if err := tCtx.PluginStateWriter().Put(clusterPluginStateVersion, &newState); err != nil {
			return pluginsCore.UnknownTransition, err
		}
	}
	return transition, nil
}

// ensureCluster creates the shared cluster if it does not already exist.
func (m *ClusterPluginManager) ensureCluster(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, cfg *config.K8sPluginConfig) (pluginsCore.Transition, ClusterPluginState, error) {
	state := ClusterPluginState{Phase: ClusterPhaseNotStarted}

	clusterName, err := m.plugin.GetClusterName(ctx, tCtx)
	if err != nil {
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition",
			fmt.Sprintf("failed to compute cluster name: %s", err.Error()), nil)), state, nil
	}
	clusterName = sanitizeName(clusterName)
	state.ClusterName = clusterName

	cluster, err := m.plugin.BuildClusterResource(ctx, tCtx)
	if err != nil {
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition",
			fmt.Sprintf("failed to build cluster resource: %s", err.Error()), nil)), state, nil
	}
	m.addClusterMetadata(tCtx.TaskExecutionMetadata(), cluster, cfg, clusterName)

	logger.Infof(ctx, "Creating shared cluster: Type:[%v], Object:[%v/%v]",
		cluster.GetObjectKind().GroupVersionKind(), cluster.GetNamespace(), cluster.GetName())
	if err := m.kubeClient.GetClient().Create(ctx, cluster); err != nil && !k8serrors.IsAlreadyExists(err) {
		if k8serrors.IsForbidden(err) {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure("RuntimeFailure", err.Error(), nil)), state, nil
		}
		return pluginsCore.UnknownTransition, state, errors.Wrapf("ClusterCreateFailed", err, "failed to create cluster resource")
	}

	state.Phase = ClusterPhaseClusterWait
	return pluginsCore.DoTransition(pluginsCore.PhaseInfoInitializing(
		time.Now(), pluginsCore.DefaultPhaseVersion, "cluster is creating", nil)), state, nil
}

// waitForClusterThenLaunchJob checks cluster readiness; once ready it submits the job.
func (m *ClusterPluginManager) waitForClusterThenLaunchJob(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, cfg *config.K8sPluginConfig, state ClusterPluginState) (pluginsCore.Transition, ClusterPluginState, error) {
	clusterObj, err := m.plugin.BuildClusterIdentityResource(ctx, tCtx.TaskExecutionMetadata())
	if err != nil {
		return pluginsCore.UnknownTransition, state, err
	}
	m.addClusterMetadata(tCtx.TaskExecutionMetadata(), clusterObj, cfg, state.ClusterName)

	nsName := k8stypes.NamespacedName{Namespace: clusterObj.GetNamespace(), Name: clusterObj.GetName()}
	if err := m.kubeClient.GetClient().Get(ctx, nsName, clusterObj); err != nil {
		if k8serrors.IsNotFound(err) {
			// Cluster was reaped or not yet visible; recreate on next round.
			state.Phase = ClusterPhaseNotStarted
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoInitializing(
				time.Now(), pluginsCore.DefaultPhaseVersion, "cluster not found, recreating", nil)), state, nil
		}
		return pluginsCore.UnknownTransition, state, err
	}

	pCtx := newPluginContext(tCtx, &state.K8sPluginState, m.kubeClient.GetClient())
	ready, err := m.plugin.IsClusterReady(ctx, pCtx, clusterObj)
	if err != nil {
		return pluginsCore.UnknownTransition, state, err
	}
	if !ready {
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoInitializing(
			time.Now(), pluginsCore.DefaultPhaseVersion, "waiting for cluster to be ready", nil)), state, nil
	}

	job, err := m.plugin.BuildJobResource(ctx, tCtx, state.ClusterName)
	if err != nil {
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition",
			fmt.Sprintf("failed to build job resource: %s", err.Error()), nil)), state, nil
	}
	m.addJobMetadata(tCtx.TaskExecutionMetadata(), job, cfg)

	logger.Infof(ctx, "Submitting job to shared cluster %s: Object:[%v/%v]", state.ClusterName, job.GetNamespace(), job.GetName())
	if err := m.kubeClient.GetClient().Create(ctx, job); err != nil && !k8serrors.IsAlreadyExists(err) {
		if k8serrors.IsForbidden(err) {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure("RuntimeFailure", err.Error(), nil)), state, nil
		}
		return pluginsCore.UnknownTransition, state, errors.Wrapf("JobCreateFailed", err, "failed to create job resource")
	}

	state.Phase = ClusterPhaseJobStarted
	return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(time.Now(), pluginsCore.DefaultPhaseVersion, "job submitted to cluster")), state, nil
}

// checkJobPhase fetches the job and reports its phase, reading outputs on success.
func (m *ClusterPluginManager) checkJobPhase(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, k8sPluginState *k8s.PluginState) (pluginsCore.Transition, error) {
	job, err := m.plugin.BuildJobIdentityResource(ctx, tCtx.TaskExecutionMetadata())
	if err != nil {
		return pluginsCore.UnknownTransition, err
	}
	m.addJobMetadata(tCtx.TaskExecutionMetadata(), job, config.GetK8sPluginConfig())

	nsName := k8stypes.NamespacedName{Namespace: job.GetNamespace(), Name: job.GetName()}
	if err := m.kubeClient.GetClient().Get(ctx, nsName, job); err != nil {
		if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure("JobDeletedExternally",
				fmt.Sprintf("job not found, name [%s]: %s", nsName.String(), err.Error()), nil)), nil
		}
		return pluginsCore.UnknownTransition, err
	}

	pCtx := newPluginContext(tCtx, k8sPluginState, m.kubeClient.GetClient())
	p, err := m.plugin.GetJobPhase(ctx, pCtx, job)
	if err != nil {
		return pluginsCore.UnknownTransition, err
	}

	if p.Phase() == k8sPluginState.Phase && p.Version() < k8sPluginState.PhaseVersion {
		p = p.WithVersion(k8sPluginState.PhaseVersion)
	}

	if p.Phase() == pluginsCore.PhaseSuccess {
		var opReader io.OutputReader
		if pCtx.ow == nil {
			opReader = ioutils.NewRemoteFileOutputReader(ctx, tCtx.DataStore(), tCtx.OutputWriter(), 0)
		} else {
			opReader = pCtx.ow.GetReader()
		}
		isErr, err := opReader.IsError(ctx)
		if err != nil {
			return pluginsCore.UnknownTransition, err
		}
		if isErr {
			taskErr, err := opReader.ReadError(ctx)
			if err != nil {
				return pluginsCore.UnknownTransition, err
			}
			if taskErr.ExecutionError == nil {
				taskErr.ExecutionError = &core.ExecutionError{Kind: core.ExecutionError_UNKNOWN, Code: "Unknown", Message: "Unknown"}
			}
			phase := pluginsCore.PhasePermanentFailure
			if taskErr.IsRecoverable {
				phase = pluginsCore.PhaseRetryableFailure
			}
			return pluginsCore.DoTransitionType(pluginsCore.TransitionTypeEphemeral,
				pluginsCore.PhaseInfoFailed(phase, taskErr.ExecutionError, p.Info())), nil
		}
		if err := tCtx.OutputWriter().Put(ctx, opReader); err != nil {
			return pluginsCore.UnknownTransition, err
		}
		return pluginsCore.DoTransition(p), nil
	}

	return pluginsCore.DoTransition(p), nil
}

// Abort deletes only the job resource. The shared cluster is intentionally left alone.
func (m *ClusterPluginManager) Abort(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) error {
	job, err := m.plugin.BuildJobIdentityResource(ctx, tCtx.TaskExecutionMetadata())
	if err != nil {
		logger.Errorf(ctx, "%v", err)
		return nil
	}
	m.addJobMetadata(tCtx.TaskExecutionMetadata(), job, config.GetK8sPluginConfig())
	if err := m.kubeClient.GetClient().Delete(ctx, job); err != nil && !k8serrors.IsNotFound(err) && !k8serrors.IsGone(err) {
		logger.Warningf(ctx, "Failed to abort job %v/%v: %v", job.GetNamespace(), job.GetName(), err)
		return err
	}
	return nil
}

// Finalize clears the job finalizer (and deletes it if configured). The cluster is never touched.
func (m *ClusterPluginManager) Finalize(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) error {
	job, err := m.plugin.BuildJobIdentityResource(ctx, tCtx.TaskExecutionMetadata())
	if err != nil {
		logger.Errorf(ctx, "%v", err)
		return nil
	}
	m.addJobMetadata(tCtx.TaskExecutionMetadata(), job, config.GetK8sPluginConfig())
	nsName := k8stypes.NamespacedName{Namespace: job.GetNamespace(), Name: job.GetName()}

	if err := m.kubeClient.GetClient().Get(ctx, nsName, job); err != nil {
		if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
			return nil
		}
		return err
	}
	if len(job.GetFinalizers()) > 0 {
		job.SetFinalizers([]string{})
		if err := m.kubeClient.GetClient().Update(ctx, job); err != nil {
			if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
				return nil
			}
			return err
		}
	}

	cfg := config.GetK8sPluginConfig()
	if cfg.DeleteResourceOnFinalize && !m.plugin.GetProperties().DisableDeleteResourceOnFinalize {
		if err := m.kubeClient.GetClient().Delete(ctx, job); err != nil && !k8serrors.IsNotFound(err) && !k8serrors.IsGone(err) {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 4: Verify it compiles**

Run: `cd /Users/kevin/git/flyte && go build ./executor/pkg/plugin/k8s/...`
Expected: builds clean. If an error like `errors.Wrapf` signature mismatch appears, check `flyteplugins/go/tasks/errors` — `Wrapf(code ErrorCode, err error, msgfmt string, args...)`; the string literals `"ClusterCreateFailed"`/`"JobCreateFailed"` must be cast to `stdErrors.ErrorCode(...)` exactly as `plugin_manager.go` does. Adjust the import + calls to mirror `plugin_manager.go:118`.

- [ ] **Step 5: Commit**

```bash
git add executor/pkg/plugin/k8s/cluster_plugin_manager.go
git commit -s -m "feat(k8s): add ClusterPluginManager state machine"
```

---

## Task 6: Test the `ClusterPluginManager` state machine

**Files:**
- Create: `executor/pkg/plugin/k8s/cluster_plugin_manager_test.go`

This task verifies the three phases with a controller-runtime fake client and a hand-written fake `ClusterPlugin`. Use the `mocks.KubeClient` from `flyteplugins/go/tasks/pluginmachinery/core/mocks` wrapping a `fake.NewClientBuilder()` client, and `mocks.TaskExecutionContext` / `TaskExecutionMetadata` / `TaskExecutionID` for the task context (mirror the patterns in `flyteplugins/go/tasks/plugins/k8s/ray/ray_test.go::dummyRayTaskContext` and the `PluginStateReader`/`PluginStateWriter` mocks).

- [ ] **Step 1: Write the fake ClusterPlugin + test scaffolding**

Create `executor/pkg/plugin/k8s/cluster_plugin_manager_test.go`:

```go
package k8s

import (
	"context"
	"testing"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	coremocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
)

// fakeClusterPlugin is a minimal ClusterPlugin used to drive the manager in tests.
type fakeClusterPlugin struct {
	clusterReady bool
	jobPhase     pluginsCore.Phase
}

func (f *fakeClusterPlugin) GetClusterName(_ context.Context, _ pluginsCore.TaskExecutionContext) (string, error) {
	return "shared-cluster", nil
}
func (f *fakeClusterPlugin) BuildClusterResource(_ context.Context, _ pluginsCore.TaskExecutionContext) (client.Object, error) {
	return &rayv1.RayCluster{TypeMeta: metav1.TypeMeta{Kind: "RayCluster", APIVersion: rayv1.SchemeGroupVersion.String()}}, nil
}
func (f *fakeClusterPlugin) BuildClusterIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &rayv1.RayCluster{TypeMeta: metav1.TypeMeta{Kind: "RayCluster", APIVersion: rayv1.SchemeGroupVersion.String()}}, nil
}
func (f *fakeClusterPlugin) IsClusterReady(_ context.Context, _ k8s.PluginContext, _ client.Object) (bool, error) {
	return f.clusterReady, nil
}
func (f *fakeClusterPlugin) BuildJobResource(_ context.Context, _ pluginsCore.TaskExecutionContext, clusterName string) (client.Object, error) {
	return &rayv1.RayJob{
		TypeMeta: metav1.TypeMeta{Kind: "RayJob", APIVersion: rayv1.SchemeGroupVersion.String()},
		Spec:     rayv1.RayJobSpec{ClusterSelector: map[string]string{"ray.io/cluster": clusterName}},
	}, nil
}
func (f *fakeClusterPlugin) BuildJobIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &rayv1.RayJob{TypeMeta: metav1.TypeMeta{Kind: "RayJob", APIVersion: rayv1.SchemeGroupVersion.String()}}, nil
}
func (f *fakeClusterPlugin) GetJobPhase(_ context.Context, _ k8s.PluginContext, _ client.Object) (pluginsCore.PhaseInfo, error) {
	switch f.jobPhase {
	case pluginsCore.PhaseSuccess:
		return pluginsCore.PhaseInfoSuccess(nil), nil
	case pluginsCore.PhaseRunning:
		return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, nil), nil
	default:
		return pluginsCore.PhaseInfoQueued(metav1.Now().Time, pluginsCore.DefaultPhaseVersion, "queued"), nil
	}
}
func (f *fakeClusterPlugin) GetProperties() k8s.PluginProperties { return k8s.PluginProperties{} }

// fakeKubeClient wraps a controller-runtime fake client to satisfy pluginsCore.KubeClient.
func newFakeKubeClient(t *testing.T, objs ...runtime.Object) pluginsCore.KubeClient {
	_ = rayv1.AddToScheme(scheme.Scheme)
	c := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objs...).Build()
	kc := &coremocks.KubeClient{}
	kc.EXPECT().GetClient().Return(c).Maybe()
	kc.EXPECT().GetCache().Return(nil).Maybe()
	return kc
}
```

Note on plugin-state mocks: implement an in-memory `PluginStateReader`/`PluginStateWriter` pair (the Ray tests under `ray_test.go` around line 890 show the `mocks.PluginStateReader` usage). The reader must decode into `*ClusterPluginState`; the writer captures the latest state so the next `Handle` round reads it back. Build a small helper `taskContextForState(t, state ClusterPluginState, objs ...runtime.Object)` returning a `*coremocks.TaskExecutionContext` whose `PluginStateReader().Get` copies `state` into the target and whose `PluginStateWriter().Put` records into a captured pointer. Mirror `dummyRayTaskContext` for `TaskExecutionMetadata`/`GetNamespace`/`GetTaskExecutionID().GetGeneratedName()` (return e.g. `"job-name"` and namespace `"ns"`).

- [ ] **Step 2: Write the phase tests**

Append:

```go
func TestClusterPluginManager_NotStarted_CreatesClusterNoOwnerRef(t *testing.T) {
	ctx := context.Background()
	kc := newFakeKubeClient(t)
	m := NewClusterPluginManager("test", &fakeClusterPlugin{}, kc)

	tCtx, captured := taskContextForState(t, ClusterPluginState{Phase: ClusterPhaseNotStarted})
	tr, err := m.Handle(ctx, tCtx)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseInitializing, tr.Info().Phase())
	assert.Equal(t, ClusterPhaseClusterWait, captured.state.Phase)
	assert.Equal(t, "shared-cluster", captured.state.ClusterName)

	// Cluster exists with no owner reference / finalizer.
	got := &rayv1.RayCluster{}
	err = kc.GetClient().Get(ctx, client.ObjectKey{Namespace: "ns", Name: "shared-cluster"}, got)
	assert.NoError(t, err)
	assert.Empty(t, got.OwnerReferences)
	assert.Empty(t, got.Finalizers)
}

func TestClusterPluginManager_ClusterWait_NotReady_StaysInitializing(t *testing.T) {
	ctx := context.Background()
	cluster := &rayv1.RayCluster{ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "shared-cluster"}}
	kc := newFakeKubeClient(t, cluster)
	m := NewClusterPluginManager("test", &fakeClusterPlugin{clusterReady: false}, kc)

	tCtx, _ := taskContextForState(t, ClusterPluginState{Phase: ClusterPhaseClusterWait, ClusterName: "shared-cluster"})
	tr, err := m.Handle(ctx, tCtx)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseInitializing, tr.Info().Phase())

	// No job was created.
	jobs := &rayv1.RayJobList{}
	_ = kc.GetClient().List(ctx, jobs)
	assert.Len(t, jobs.Items, 0)
}

func TestClusterPluginManager_ClusterWait_Ready_CreatesJobWithSelector(t *testing.T) {
	ctx := context.Background()
	cluster := &rayv1.RayCluster{ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "shared-cluster"}}
	kc := newFakeKubeClient(t, cluster)
	m := NewClusterPluginManager("test", &fakeClusterPlugin{clusterReady: true}, kc)

	tCtx, captured := taskContextForState(t, ClusterPluginState{Phase: ClusterPhaseClusterWait, ClusterName: "shared-cluster"})
	tr, err := m.Handle(ctx, tCtx)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, tr.Info().Phase())
	assert.Equal(t, ClusterPhaseJobStarted, captured.state.Phase)

	got := &rayv1.RayJob{}
	err = kc.GetClient().Get(ctx, client.ObjectKey{Namespace: "ns", Name: "job-name"}, got)
	assert.NoError(t, err)
	assert.Equal(t, "shared-cluster", got.Spec.ClusterSelector["ray.io/cluster"])
	assert.NotEmpty(t, got.OwnerReferences) // job IS owned
}

func TestClusterPluginManager_JobStarted_MapsRunning(t *testing.T) {
	ctx := context.Background()
	job := &rayv1.RayJob{ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "job-name"}}
	kc := newFakeKubeClient(t, job)
	m := NewClusterPluginManager("test", &fakeClusterPlugin{jobPhase: pluginsCore.PhaseRunning}, kc)

	tCtx, _ := taskContextForState(t, ClusterPluginState{Phase: ClusterPhaseJobStarted, ClusterName: "shared-cluster"})
	tr, err := m.Handle(ctx, tCtx)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, tr.Info().Phase())
}

func TestClusterPluginManager_Abort_DeletesJobNotCluster(t *testing.T) {
	ctx := context.Background()
	cluster := &rayv1.RayCluster{ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "shared-cluster"}}
	job := &rayv1.RayJob{ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "job-name"}}
	kc := newFakeKubeClient(t, cluster, job)
	m := NewClusterPluginManager("test", &fakeClusterPlugin{}, kc)

	tCtx, _ := taskContextForState(t, ClusterPluginState{Phase: ClusterPhaseJobStarted, ClusterName: "shared-cluster"})
	assert.NoError(t, m.Abort(ctx, tCtx))

	// Job gone, cluster remains.
	assert.True(t, apierrorsIsNotFound(kc.GetClient().Get(ctx, client.ObjectKey{Namespace: "ns", Name: "job-name"}, &rayv1.RayJob{})))
	assert.NoError(t, kc.GetClient().Get(ctx, client.ObjectKey{Namespace: "ns", Name: "shared-cluster"}, &rayv1.RayCluster{}))
}
```

Add a tiny helper near the imports for the abort assertion:

```go
func apierrorsIsNotFound(err error) bool { return err != nil && k8serrors.IsNotFound(err) }
```

and import `k8serrors "k8s.io/apimachinery/pkg/api/errors"` plus the unused `corev1`/`mock` imports if the state-mock helper needs them (drop any that go unused).

- [ ] **Step 3: Run tests to verify they fail (then pass)**

Run: `cd /Users/kevin/git/flyte && go test ./executor/pkg/plugin/k8s/ -run TestClusterPluginManager -v`
Expected: initially may fail while wiring the `taskContextForState` state-mock helper. Iterate on the helper until all five tests PASS. The most common fixups: ensure `PluginStateReader().Get` type-switches on `*ClusterPluginState`, and `PluginStateWriter().Put` stores into `captured.state`.

- [ ] **Step 4: Commit**

```bash
git add executor/pkg/plugin/k8s/cluster_plugin_manager_test.go
git commit -s -m "test(k8s): cover ClusterPluginManager cluster/job lifecycle"
```

---

## Task 7: Wire `ClusterPluginManager` into the executor registry

**Files:**
- Modify: `executor/pkg/plugin/registry.go:54-82` (k8s plugin loading loop)

- [ ] **Step 1: Construct the right manager per entry**

Replace the body of the `for _, entry := range r.pluginRegistry.GetK8sPlugins()` loop (the part that builds `pm`) so it picks the manager based on `entry.ClusterPlugin`:

```go
	for _, entry := range r.pluginRegistry.GetK8sPlugins() {
		var plugin pluginsCore.Plugin
		if entry.ClusterPlugin != nil {
			plugin = executorK8s.NewClusterPluginManager(
				entry.ID,
				entry.ClusterPlugin,
				r.setupCtx.KubeClient(),
			)
		} else {
			pm := executorK8s.NewPluginManager(
				entry.ID,
				entry.Plugin,
				r.setupCtx.KubeClient(),
			)
			if err := pm.InitializeObjectEventWatcher(ctx); err != nil {
				return fmt.Errorf("failed to initialize k8s object event watcher for plugin %s: %w", entry.ID, err)
			}
			plugin = pm
		}

		for _, taskType := range entry.RegisteredTaskTypes {
			if existing, ok := r.plugins[taskType]; ok {
				logger.Warnf(ctx, "Task type %q already registered by plugin %q, overwriting with %q",
					taskType, existing.GetID(), entry.ID)
			}
			r.plugins[taskType] = plugin
		}

		if entry.IsDefault {
			if r.defaultPlugin != nil {
				logger.Warnf(ctx, "Multiple default plugins found, overwriting %q with %q",
					r.defaultPlugin.GetID(), entry.ID)
			}
			r.defaultPlugin = plugin
		}

		logger.Infof(ctx, "Registered k8s plugin [%s] for task types %v", entry.ID, entry.RegisteredTaskTypes)
	}
```

(Event watching for the cluster-plugin job/cluster types is out of scope for v1 — the `ClusterPluginManager` does not attach object events. This keeps the change small; a follow-up can add it. Note this limitation in the commit message.)

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/kevin/git/flyte && go build ./executor/...`
Expected: builds clean.

- [ ] **Step 3: Commit**

```bash
git add executor/pkg/plugin/registry.go
git commit -s -m "feat(executor): route ClusterPlugin entries to ClusterPluginManager"
```

---

## Task 8: Migrate the Ray plugin to `ClusterPlugin`

This is the largest task. Split `constructRayJob` into a cluster builder and a job builder, add `GetClusterName`, and rewrite the interface methods.

**Files:**
- Modify: `flyteplugins/go/tasks/plugins/k8s/ray/ray.go`

- [ ] **Step 1: Add constants and `GetClusterName`**

Add to the `const` block (line 37-48):

```go
	KindRayCluster        = "RayCluster"
	rayClusterLabelKey    = "ray.io/cluster"
	clusterNameHashLength = 16
```

Add imports: `"strings"` already present; add `"github.com/flyteorg/flyte/v2/flytestdlib/pbhash"`.

Add the method:

```go
// GetClusterName returns a deterministic cluster name derived from the hash of the RayJob plugin
// spec proto, so two tasks with identical Ray configuration share one cluster.
func (rayJobResourceHandler) GetClusterName(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (string, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return "", flyteerr.Errorf(flyteerr.BadTaskSpecification, "unable to fetch task specification [%v]", err.Error())
	}
	rayJob := plugins.RayJob{}
	if err := utils.UnmarshalStruct(taskTemplate.GetCustom(), &rayJob); err != nil { //nolint: staticcheck
		return "", flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
	}
	hash, err := pbhash.ComputeHashString(ctx, &rayJob)
	if err != nil {
		return "", flyteerr.Errorf(flyteerr.BadTaskSpecification, "failed to hash RayJob spec: %v", err.Error())
	}
	// ComputeHashString returns standard base64; sanitize + truncate into a DNS1123-safe suffix.
	safe := utils.ConvertToDNS1123SubdomainCompatibleString(strings.ToLower(hash))
	if len(safe) > clusterNameHashLength {
		safe = safe[:clusterNameHashLength]
	}
	return "ray-" + safe, nil
}
```

(Confirm `utils.ConvertToDNS1123SubdomainCompatibleString` exists in `pluginmachinery/utils`; it is the same helper used by the plugin manager via `pluginsUtils`. If the import alias differs, match the existing `utils` import in ray.go.)

- [ ] **Step 2: Refactor `constructRayJob` into cluster + job builders**

Rename/replace `constructRayJob` so the cluster-spec construction returns a standalone `*rayv1.RayCluster`. Add a `buildRayClusterSpec` helper holding the current spec-building logic (lines 145-238, everything up to the `jobSpec`), then:

```go
// BuildClusterResource builds a standalone RayCluster keyed by the spec hash. It carries the
// ray.io/cluster label so jobs can bind via ClusterSelector, and is created without owner ref /
// finalizer by the framework so it can be shared.
func (h rayJobResourceHandler) BuildClusterResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	rayJob, podSpec, objectMeta, primaryContainerIdx, primaryContainer, err := h.readRaySpec(ctx, taskCtx)
	if err != nil {
		return nil, err
	}
	headNodeRayStartParams := buildHeadStartParams(rayJob)
	podSpec.ServiceAccountName = GetConfig().ServiceAccount

	clusterSpec, err := buildRayClusterSpec(taskCtx, rayJob, objectMeta, *podSpec, headNodeRayStartParams, primaryContainerIdx, *primaryContainer)
	if err != nil {
		return nil, err
	}

	clusterName, err := h.GetClusterName(ctx, taskCtx)
	if err != nil {
		return nil, err
	}

	return &rayv1.RayCluster{
		TypeMeta: metav1.TypeMeta{Kind: KindRayCluster, APIVersion: rayv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{rayClusterLabelKey: clusterName},
		},
		Spec: *clusterSpec,
	}, nil
}
```

Extract two small helpers used by both the cluster and job paths to avoid duplication (DRY):

```go
// readRaySpec unmarshals the RayJob proto and builds the base pod spec + primary container.
func (rayJobResourceHandler) readRaySpec(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (*plugins.RayJob, *v1.PodSpec, *metav1.ObjectMeta, int, *v1.Container, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, nil, nil, 0, nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "unable to fetch task specification [%v]", err.Error())
	} else if taskTemplate == nil {
		return nil, nil, nil, 0, nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "nil task specification")
	}
	rayJob := plugins.RayJob{}
	if err := utils.UnmarshalStruct(taskTemplate.GetCustom(), &rayJob); err != nil { //nolint: staticcheck
		return nil, nil, nil, 0, nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
	}
	podSpec, objectMeta, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, nil, nil, 0, nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create pod spec: [%v]", err.Error())
	}
	var primaryContainer *v1.Container
	var primaryContainerIdx int
	for idx, c := range podSpec.Containers {
		if c.Name == primaryContainerName {
			c := c
			primaryContainer = &c
			primaryContainerIdx = idx
			break
		}
	}
	if primaryContainer == nil {
		return nil, nil, nil, 0, nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to get primary container from the pod")
	}
	return &rayJob, podSpec, objectMeta, primaryContainerIdx, primaryContainer, nil
}

// buildHeadStartParams reproduces the head-node ray-start parameter defaulting from the old BuildResource.
func buildHeadStartParams(rayJob *plugins.RayJob) map[string]string {
	cfg := GetConfig()
	headNodeRayStartParams := make(map[string]string)
	if rayJob.RayCluster.HeadGroupSpec != nil && rayJob.RayCluster.HeadGroupSpec.RayStartParams != nil {
		headNodeRayStartParams = rayJob.RayCluster.HeadGroupSpec.RayStartParams
	} else if headNode := cfg.Defaults.HeadNode; len(headNode.StartParameters) > 0 {
		headNodeRayStartParams = headNode.StartParameters
	}
	if _, exist := headNodeRayStartParams[IncludeDashboard]; !exist {
		headNodeRayStartParams[IncludeDashboard] = strconv.FormatBool(cfg.IncludeDashboard)
	}
	if _, exist := headNodeRayStartParams[NodeIPAddress]; !exist {
		headNodeRayStartParams[NodeIPAddress] = cfg.Defaults.HeadNode.IPAddress
	}
	if _, exist := headNodeRayStartParams[DashboardHost]; !exist {
		headNodeRayStartParams[DashboardHost] = cfg.DashboardHost
	}
	if _, exists := headNodeRayStartParams[DisableUsageStatsStartParameter]; !exists && !cfg.EnableUsageStats {
		headNodeRayStartParams[DisableUsageStatsStartParameter] = DisableUsageStatsStartParameterVal
	}
	return headNodeRayStartParams
}
```

`buildRayClusterSpec` = the current `constructRayJob` body from line 146 through the `rayClusterSpec` assembly (lines 146-229), returning `(*rayv1.RayClusterSpec, error)` instead of building a RayJob. Keep `EnableInTreeAutoscaling` and set the autoscaler idle timeout so idle clusters self-reap:

```go
	enableAutoscaling := true
	rayClusterSpec.EnableInTreeAutoscaling = &enableAutoscaling
```

(Leave existing per-group logic intact; only the return type changes.)

- [ ] **Step 3: Implement `BuildJobResource` with `ClusterSelector`**

```go
// BuildJobResource builds a RayJob that binds to an existing shared cluster via ClusterSelector.
func (h rayJobResourceHandler) BuildJobResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, clusterName string) (client.Object, error) {
	rayJob, podSpec, objectMeta, primaryContainerIdx, primaryContainer, err := h.readRaySpec(ctx, taskCtx)
	if err != nil {
		return nil, err
	}
	_ = podSpec
	_ = objectMeta
	_ = primaryContainerIdx

	// Entrypoint comes from the primary container args (same as before).
	runtimeEnvYaml := rayJob.RuntimeEnvYaml
	if rayJob.RuntimeEnv != "" && rayJob.RuntimeEnvYaml == "" { //nolint: staticcheck
		runtimeEnvYaml, err = convertBase64RuntimeEnvToYaml(rayJob.RuntimeEnv) //nolint: staticcheck
		if err != nil {
			return nil, err
		}
	}

	jobSpec := rayv1.RayJobSpec{
		RayClusterSpec:  nil, // bind to existing cluster, do not create a new one
		ClusterSelector: map[string]string{rayClusterLabelKey: clusterName},
		Entrypoint:      strings.Join(primaryContainer.Args, " "),
		RuntimeEnvYAML:  runtimeEnvYaml,
		// ShutdownAfterJobFinishes MUST stay false: the cluster is shared and must outlive the job.
		ShutdownAfterJobFinishes: false,
	}

	return &rayv1.RayJob{
		TypeMeta:   metav1.TypeMeta{Kind: KindRayJob, APIVersion: rayv1.SchemeGroupVersion.String()},
		Spec:       jobSpec,
		ObjectMeta: *objectMeta,
	}, nil
}
```

(When using `ClusterSelector`, KubeRay requires `RayClusterSpec` to be nil and does NOT use `SubmitterPodTemplate` from a fresh cluster; the submitter still runs. Keep the submitter default behavior — if `BuildResourceRaySubmitterPodAffinity` test depended on a submitter template, set `SubmitterPodTemplate` via `buildSubmitterPodTemplate` using the *cluster* head spec rebuilt here, or drop that assertion. Decide during Step 6 test fixups.)

- [ ] **Step 4: Implement identity + readiness + phase methods**

```go
func (rayJobResourceHandler) BuildClusterIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &rayv1.RayCluster{
		TypeMeta: metav1.TypeMeta{Kind: KindRayCluster, APIVersion: rayv1.SchemeGroupVersion.String()},
	}, nil
}

func (rayJobResourceHandler) BuildJobIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &rayv1.RayJob{
		TypeMeta: metav1.TypeMeta{Kind: KindRayJob, APIVersion: rayv1.SchemeGroupVersion.String()},
	}, nil
}

// IsClusterReady reports readiness based on the RayCluster status state.
func (rayJobResourceHandler) IsClusterReady(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object) (bool, error) {
	cluster, ok := resource.(*rayv1.RayCluster)
	if !ok {
		return false, fmt.Errorf("unexpected resource type: expected *RayCluster, got %T", resource)
	}
	return cluster.Status.State == rayv1.Ready, nil
}
```

Rename the existing `GetTaskPhase` method to `GetJobPhase` (signature is identical; it already takes `k8s.PluginContext` and `client.Object` and type-asserts to `*rayv1.RayJob`). Delete the old `BuildResource`, `BuildIdentityResource`, `IsTerminal`, and `GetCompletionTime` methods (the `ClusterPlugin` interface does not require `GarbageCollectable`). If any other code references them, update those references.

- [ ] **Step 5: Update `init()` registration**

Replace the `RegisterK8sPlugin` call (lines 810-831) so it registers a cluster plugin:

```go
func init() {
	if err := rayv1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                     rayTaskType,
			RegisteredTaskTypes:    []pluginsCore.TaskType{rayTaskType},
			ResourceToWatch:        &rayv1.RayJob{},
			ClusterResourceToWatch: &rayv1.RayCluster{},
			ClusterPlugin:          rayJobResourceHandler{},
			IsDefault:              false,
			CustomKubeClient: func(ctx context.Context) (pluginsCore.KubeClient, error) {
				remoteConfig := GetConfig().RemoteClusterConfig
				if !remoteConfig.Enabled {
					return nil, nil
				}
				kubeConfig, err := k8s.KubeClientConfig(remoteConfig.Endpoint, remoteConfig.Auth)
				if err != nil {
					return nil, err
				}
				return k8s.NewDefaultKubeClient(kubeConfig)
			},
		})
}
```

- [ ] **Step 6: Build and fix compile errors**

Run: `cd /Users/kevin/git/flyte && go build ./flyteplugins/go/tasks/plugins/k8s/ray/...`
Expected: builds. Remove now-unused helpers (`constructRayJob` if fully replaced) and unused imports. Keep `buildSubmitterPodTemplate`, `buildHeadPodTemplate`, `buildWorkerPodTemplate` (still used by `buildRayClusterSpec`).

- [ ] **Step 7: Commit**

```bash
git add flyteplugins/go/tasks/plugins/k8s/ray/ray.go
git commit -s -m "feat(ray): migrate Ray plugin to ClusterPlugin with shared cluster"
```

---

## Task 9: Update Ray unit tests for the split

**Files:**
- Modify: `flyteplugins/go/tasks/plugins/k8s/ray/ray_test.go`

- [ ] **Step 1: Add `GetClusterName` determinism tests**

Append:

```go
func TestGetClusterNameDeterministic(t *testing.T) {
	ctx := context.Background()
	h := rayJobResourceHandler{}

	taskCtx := dummyRayTaskContext(dummyRayTaskTemplate("id", dummyRayCustomObj()), resourceRequirements, nil, "", serviceAccount)
	n1, err := h.GetClusterName(ctx, taskCtx)
	assert.NoError(t, err)
	n2, err := h.GetClusterName(ctx, taskCtx)
	assert.NoError(t, err)
	assert.Equal(t, n1, n2, "same spec must yield same cluster name")
	assert.True(t, strings.HasPrefix(n1, "ray-"))
	assert.LessOrEqual(t, len(n1), 47)
}

func TestGetClusterNameDiffersForDifferentSpecs(t *testing.T) {
	ctx := context.Background()
	h := rayJobResourceHandler{}

	a := dummyRayCustomObj()
	b := dummyRayCustomObj()
	b.RuntimeEnvYaml = "pip: [numpy]" // perturb the spec
	taskA := dummyRayTaskContext(dummyRayTaskTemplate("id", a), resourceRequirements, nil, "", serviceAccount)
	taskB := dummyRayTaskContext(dummyRayTaskTemplate("id", b), resourceRequirements, nil, "", serviceAccount)
	nA, _ := h.GetClusterName(ctx, taskA)
	nB, _ := h.GetClusterName(ctx, taskB)
	assert.NotEqual(t, nA, nB)
}
```

(Use whatever the existing helpers are named — `dummyRayTaskTemplate`/`dummyRayCustomObj`/`resourceRequirements`/`serviceAccount` appear in the current test file; adjust call signatures to match.)

- [ ] **Step 2: Convert `TestBuildResourceRay*` to cluster + job assertions**

For each existing `TestBuildResourceRay*` test that called `rayJobResourceHandler{}.BuildResource(...)` and asserted on the returned `*rayv1.RayJob`, split the assertions:

- Call `BuildClusterResource(...)`; assert the returned `*rayv1.RayCluster` has the head/worker group specs, service type, autoscaling, and the `ray.io/cluster` label.
- Call `BuildJobResource(ctx, taskCtx, "ray-test")`; assert the returned `*rayv1.RayJob` has `Spec.RayClusterSpec == nil`, `Spec.ClusterSelector["ray.io/cluster"] == "ray-test"`, correct `Entrypoint`, and `ShutdownAfterJobFinishes == false`.

Example rewrite of the core assertion in `TestBuildResourceRay`:

```go
	h := rayJobResourceHandler{}
	clusterObj, err := h.BuildClusterResource(context.Background(), taskCtx)
	require.NoError(t, err)
	rayCluster, ok := clusterObj.(*rayv1.RayCluster)
	require.True(t, ok)
	assert.Equal(t, 1, len(rayCluster.Spec.WorkerGroupSpecs))
	assert.Equal(t, "ray-head", rayCluster.Spec.HeadGroupSpec.Template.Spec.Containers[0].Name)
	assert.Contains(t, rayCluster.Labels, "ray.io/cluster")

	jobObj, err := h.BuildJobResource(context.Background(), taskCtx, "ray-test")
	require.NoError(t, err)
	rayJob, ok := jobObj.(*rayv1.RayJob)
	require.True(t, ok)
	assert.Nil(t, rayJob.Spec.RayClusterSpec)
	assert.Equal(t, "ray-test", rayJob.Spec.ClusterSelector["ray.io/cluster"])
	assert.False(t, rayJob.Spec.ShutdownAfterJobFinishes)
```

- [ ] **Step 3: Rename `GetTaskPhase` test calls to `GetJobPhase`**

Every test that calls `rayJobResourceHandler{}.GetTaskPhase(...)` must now call `GetJobPhase(...)` (same args/returns). Update the call sites; the assertions stay identical.

- [ ] **Step 4: Add an `IsClusterReady` test**

```go
func TestIsClusterReady(t *testing.T) {
	h := rayJobResourceHandler{}
	notReady := &rayv1.RayCluster{}
	ok, err := h.IsClusterReady(context.Background(), nil, notReady)
	assert.NoError(t, err)
	assert.False(t, ok)

	ready := &rayv1.RayCluster{Status: rayv1.RayClusterStatus{State: rayv1.Ready}}
	ok, err = h.IsClusterReady(context.Background(), nil, ready)
	assert.NoError(t, err)
	assert.True(t, ok)
}
```

- [ ] **Step 5: Remove tests for deleted methods**

Delete or convert any tests referencing the removed `BuildResource`, `BuildIdentityResource`, `IsTerminal`, `GetCompletionTime`. (Tests for terminal-state/completion-time no longer apply; the GC interface was dropped for cluster plugins.)

- [ ] **Step 6: Run the full Ray test suite**

Run: `cd /Users/kevin/git/flyte && go test ./flyteplugins/go/tasks/plugins/k8s/ray/... -v`
Expected: PASS. Fix assertion drift (names, label keys) until green.

- [ ] **Step 7: Commit**

```bash
git add flyteplugins/go/tasks/plugins/k8s/ray/ray_test.go
git commit -s -m "test(ray): cover ClusterPlugin split (cluster/job/readiness/name)"
```

---

## Task 10: Full build, vet, and module tidy

**Files:** none (verification only)

- [ ] **Step 1: Build everything**

Run: `cd /Users/kevin/git/flyte && go build ./...`
Expected: no errors.

- [ ] **Step 2: Vet the changed packages**

Run: `cd /Users/kevin/git/flyte && go vet ./flyteplugins/go/tasks/plugins/k8s/ray/... ./executor/pkg/plugin/... ./flyteplugins/go/tasks/pluginmachinery/...`
Expected: clean.

- [ ] **Step 3: Run the affected test packages**

Run: `cd /Users/kevin/git/flyte && go test ./flyteplugins/go/tasks/plugins/k8s/ray/... ./executor/pkg/plugin/... ./flyteplugins/go/tasks/pluginmachinery/...`
Expected: PASS.

- [ ] **Step 4: go mod tidy (if imports changed)**

Run: `cd /Users/kevin/git/flyte && go mod tidy`
Expected: no unexpected churn; commit `go.mod`/`go.sum` only if they changed.

- [ ] **Step 5: Commit any tidy changes**

```bash
git add -A
git commit -s -m "chore: go mod tidy after ClusterPlugin changes" || echo "nothing to commit"
```

---

## Notes / Known limitations (for the PR description)

- **Event watching**: `ClusterPluginManager` does not attach recent k8s object events to phase info (the existing `PluginManager` does). Deferred to a follow-up; noted in Task 7.
- **Cluster GC**: clusters are never deleted by Flyte. They rely on KubeRay's autoscaler `IdleTimeoutSeconds`. Operators must enable autoscaling (the Ray plugin forces `EnableInTreeAutoscaling=true`). Document the recommended idle timeout in Ray plugin config.
- **GarbageCollectable**: cluster plugins intentionally drop the `IsTerminal`/`GetCompletionTime` interface. The external GC continues to operate on standard `Plugin`s only.
- **Concurrency**: two tasks hashing to the same cluster may both attempt `Create`; `IsAlreadyExists` is swallowed so the loser proceeds to wait. Confirmed safe.
