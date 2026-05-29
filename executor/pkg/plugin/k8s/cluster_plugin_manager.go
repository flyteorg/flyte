package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	pluginsUtils "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	stdErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const clusterPluginStateVersion = 1

// ClusterPhase tracks the high-level phase of the ClusterPluginManager's state machine.
type ClusterPhase uint8

const (
	// ClusterPhaseNotStarted indicates the cluster has not yet been created.
	ClusterPhaseNotStarted ClusterPhase = iota
	// ClusterPhaseClusterWait indicates the cluster is being created and we are waiting for it to be ready.
	ClusterPhaseClusterWait
	// ClusterPhaseJobStarted indicates the job has been submitted to a ready cluster and is being watched.
	ClusterPhaseJobStarted
)

// ClusterPluginState is the state persisted by the ClusterPluginManager between reconciliation rounds.
type ClusterPluginState struct {
	Phase          ClusterPhase
	ClusterName    string
	K8sPluginState k8s.PluginState
}

var _ pluginsCore.Plugin = &ClusterPluginManager{}

// ClusterPluginManager wraps a k8s.ClusterPlugin to implement pluginsCore.Plugin. It drives a
// multi-step lifecycle: create-or-reuse a shared cluster, wait for the cluster to be ready, submit
// a job bound to it, and watch the job. The cluster is intentionally NOT owned by the task execution
// so that completing or aborting one task never deletes a cluster other tasks may still be using.
type ClusterPluginManager struct {
	id         string
	plugin     k8s.ClusterPlugin
	kubeClient pluginsCore.KubeClient
}

// NewClusterPluginManager creates a ClusterPluginManager that wraps a k8s.ClusterPlugin.
func NewClusterPluginManager(id string, plugin k8s.ClusterPlugin, kubeClient pluginsCore.KubeClient) *ClusterPluginManager {
	return &ClusterPluginManager{
		id:         id,
		plugin:     plugin,
		kubeClient: kubeClient,
	}
}

func (m *ClusterPluginManager) GetID() string {
	return m.id
}

func (m *ClusterPluginManager) GetProperties() pluginsCore.PluginProperties {
	return pluginsCore.PluginProperties{
		GeneratedNameMaxLength: m.plugin.GetProperties().GeneratedNameMaxLength,
	}
}

// sanitizeName returns the name unchanged if it is already a valid DNS1123 subdomain, otherwise it
// converts it into a DNS1123-subdomain-compatible string.
func sanitizeName(name string) string {
	if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
		return pluginsUtils.ConvertToDNS1123SubdomainCompatibleString(name)
	}
	return name
}

// addJobMetadata stamps namespace/annotations/labels, the generated name, owner references and
// finalizers (unless disabled) on a job object that IS owned by this task execution. Mirrors
// PluginManager.addObjectMetadata.
func (m *ClusterPluginManager) addJobMetadata(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig) {
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(pluginsUtils.UnionMaps(cfg.DefaultAnnotations, o.GetAnnotations(), pluginsUtils.CopyMap(taskCtx.GetAnnotations())))
	o.SetLabels(pluginsUtils.UnionMaps(cfg.DefaultLabels, o.GetLabels(), pluginsUtils.CopyMap(taskCtx.GetLabels())))
	o.SetName(taskCtx.GetTaskExecutionID().GetGeneratedName())

	if !m.plugin.GetProperties().DisableInjectOwnerReferences && !cfg.DisableInjectOwnerReferences {
		o.SetOwnerReferences([]metav1.OwnerReference{taskCtx.GetOwnerReference()})
	}

	if cfg.InjectFinalizer && !m.plugin.GetProperties().DisableInjectFinalizer {
		f := append(o.GetFinalizers(), "flyte/flytek8s")
		o.SetFinalizers(f)
	}

	if errs := validation.IsDNS1123Subdomain(o.GetName()); len(errs) > 0 {
		o.SetName(pluginsUtils.ConvertToDNS1123SubdomainCompatibleString(o.GetName()))
	}
}

// addClusterMetadata stamps namespace/annotations/labels and the (sanitized) cluster name on a
// cluster object. Unlike addJobMetadata it adds NO owner reference and NO finalizer, because the
// cluster is shared and must outlive any single task execution.
func (m *ClusterPluginManager) addClusterMetadata(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig, clusterName string) {
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(pluginsUtils.UnionMaps(cfg.DefaultAnnotations, o.GetAnnotations(), pluginsUtils.CopyMap(taskCtx.GetAnnotations())))
	o.SetLabels(pluginsUtils.UnionMaps(cfg.DefaultLabels, o.GetLabels(), pluginsUtils.CopyMap(taskCtx.GetLabels())))
	o.SetName(sanitizeName(clusterName))
}

// ensureCluster creates (or reuses) the shared cluster resource and transitions the state machine
// into ClusterPhaseClusterWait.
func (m *ClusterPluginManager) ensureCluster(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, state ClusterPluginState) (pluginsCore.Transition, ClusterPluginState, error) {
	cfg := config.GetK8sPluginConfig()

	clusterName, err := m.plugin.GetClusterName(ctx, tCtx)
	if err != nil {
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition",
			fmt.Sprintf("Failed to determine cluster name, caused by: %s", err.Error()), nil)), state, nil
	}
	clusterName = sanitizeName(clusterName)
	state.ClusterName = clusterName

	cluster, err := m.plugin.BuildClusterResource(ctx, tCtx)
	if err != nil {
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition",
			fmt.Sprintf("Failed to build cluster resource, caused by: %s", err.Error()), nil)), state, nil
	}

	m.addClusterMetadata(tCtx.TaskExecutionMetadata(), cluster, cfg, clusterName)
	logger.Infof(ctx, "Creating Cluster: Type:[%v], Object:[%v/%v]", cluster.GetObjectKind().GroupVersionKind(), cluster.GetNamespace(), cluster.GetName())

	err = m.kubeClient.GetClient().Create(ctx, cluster)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		if k8serrors.IsForbidden(err) {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure("RuntimeFailure", err.Error(), nil)), state, nil
		}
		reason := k8serrors.ReasonForError(err)
		logger.Errorf(ctx, "Failed to create cluster resource, system error. err: %v", err)
		return pluginsCore.UnknownTransition, state, errors.Wrapf(stdErrors.ErrorCode(reason), err, "failed to create cluster resource")
	}

	state.Phase = ClusterPhaseClusterWait
	return pluginsCore.DoTransition(pluginsCore.PhaseInfoInitializing(time.Now(), pluginsCore.DefaultPhaseVersion, "cluster is creating", nil)), state, nil
}

// waitForClusterThenLaunchJob waits for the cluster to become ready and, once it is, submits the job
// and transitions into ClusterPhaseJobStarted.
func (m *ClusterPluginManager) waitForClusterThenLaunchJob(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, state ClusterPluginState) (pluginsCore.Transition, ClusterPluginState, error) {
	cfg := config.GetK8sPluginConfig()

	clusterObj, err := m.plugin.BuildClusterIdentityResource(ctx, tCtx.TaskExecutionMetadata())
	if err != nil {
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition",
			fmt.Sprintf("Failed to build cluster identity resource, caused by: %s", err.Error()), nil)), state, nil
	}
	m.addClusterMetadata(tCtx.TaskExecutionMetadata(), clusterObj, cfg, state.ClusterName)

	nsName := k8stypes.NamespacedName{Namespace: clusterObj.GetNamespace(), Name: clusterObj.GetName()}
	if err := m.kubeClient.GetClient().Get(ctx, nsName, clusterObj); err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Warningf(ctx, "Cluster %v not found, recreating. Error: %v", nsName, err)
			state.Phase = ClusterPhaseNotStarted
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoInitializing(time.Now(), pluginsCore.DefaultPhaseVersion, "cluster not found, recreating", nil)), state, nil
		}
		logger.Warningf(ctx, "Failed to retrieve cluster details with name: %v. Error: %v", nsName, err)
		return pluginsCore.UnknownTransition, state, err
	}

	pCtx := newPluginContext(tCtx, &state.K8sPluginState, m.kubeClient.GetClient())
	ready, err := m.plugin.IsClusterReady(ctx, pCtx, clusterObj)
	if err != nil {
		logger.Warnf(ctx, "failed to check cluster readiness in plugin [%s], with error: %s", m.GetID(), err.Error())
		return pluginsCore.UnknownTransition, state, err
	}
	if !ready {
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoInitializing(time.Now(), pluginsCore.DefaultPhaseVersion, "waiting for cluster to be ready", nil)), state, nil
	}

	job, err := m.plugin.BuildJobResource(ctx, tCtx, state.ClusterName)
	if err != nil {
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition",
			fmt.Sprintf("Failed to build job resource, caused by: %s", err.Error()), nil)), state, nil
	}

	m.addJobMetadata(tCtx.TaskExecutionMetadata(), job, cfg)
	logger.Infof(ctx, "Creating Job: Type:[%v], Object:[%v/%v]", job.GetObjectKind().GroupVersionKind(), job.GetNamespace(), job.GetName())

	err = m.kubeClient.GetClient().Create(ctx, job)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		if k8serrors.IsForbidden(err) {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure("RuntimeFailure", err.Error(), nil)), state, nil
		}
		reason := k8serrors.ReasonForError(err)
		logger.Errorf(ctx, "Failed to create job resource, system error. err: %v", err)
		return pluginsCore.UnknownTransition, state, errors.Wrapf(stdErrors.ErrorCode(reason), err, "failed to create job resource")
	}

	state.Phase = ClusterPhaseJobStarted
	return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(time.Now(), pluginsCore.DefaultPhaseVersion, "job submitted to cluster")), state, nil
}

// checkJobPhase fetches the job and reports its phase, reading outputs once the job succeeds.
// Mirrors PluginManager.checkResourcePhase.
func (m *ClusterPluginManager) checkJobPhase(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, k8sPluginState *k8s.PluginState) (pluginsCore.Transition, error) {
	job, err := m.plugin.BuildJobIdentityResource(ctx, tCtx.TaskExecutionMetadata())
	if err != nil {
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition",
			fmt.Sprintf("Failed to build job identity resource, caused by: %s", err.Error()), nil)), nil
	}
	m.addJobMetadata(tCtx.TaskExecutionMetadata(), job, config.GetK8sPluginConfig())

	nsName := k8stypes.NamespacedName{Namespace: job.GetNamespace(), Name: job.GetName()}
	if err := m.kubeClient.GetClient().Get(ctx, nsName, job); err != nil {
		if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
			logger.Warningf(ctx, "Failed to find the Job with name: %v. Error: %v", nsName, err)
			failureReason := fmt.Sprintf("job not found, name [%s]. reason: %s", nsName.String(), err.Error())
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure("JobDeletedExternally", failureReason, nil)), nil
		}
		logger.Warningf(ctx, "Failed to retrieve Job details with name: %v. Error: %v", nsName, err)
		return pluginsCore.UnknownTransition, err
	}

	pCtx := newPluginContext(tCtx, k8sPluginState, m.kubeClient.GetClient())
	p, err := m.plugin.GetJobPhase(ctx, pCtx, job)
	if err != nil {
		logger.Warnf(ctx, "failed to check status of job in plugin [%s], with error: %s", m.GetID(), err.Error())
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
		y, err := opReader.IsError(ctx)
		if err != nil {
			return pluginsCore.UnknownTransition, err
		}
		if y {
			taskErr, err := opReader.ReadError(ctx)
			if err != nil {
				return pluginsCore.UnknownTransition, err
			}

			if taskErr.ExecutionError == nil {
				taskErr.ExecutionError = &core.ExecutionError{Kind: core.ExecutionError_UNKNOWN, Code: "Unknown", Message: "Unknown"}
			}
			var phase pluginsCore.Phase
			if taskErr.IsRecoverable {
				phase = pluginsCore.PhaseRetryableFailure
			} else {
				phase = pluginsCore.PhasePermanentFailure
			}
			return pluginsCore.DoTransitionType(
				pluginsCore.TransitionTypeEphemeral,
				pluginsCore.PhaseInfoFailed(phase, taskErr.ExecutionError, p.Info()),
			), nil
		}

		if err := tCtx.OutputWriter().Put(ctx, opReader); err != nil {
			return pluginsCore.UnknownTransition, err
		}
		return pluginsCore.DoTransition(p), nil
	}

	// If the job is being deleted in the background while still non-terminal, surface it as a
	// system-retryable failure (mirrors PluginManager.checkResourcePhase). Note: this checks only
	// the job, never the shared cluster, which is owned by no single task execution.
	if !p.Phase().IsTerminal() && job.GetDeletionTimestamp() != nil {
		failureReason := fmt.Sprintf("object [%s] terminated unexpectedly in the background", nsName.String())
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure("UnexpectedObjectDeletion", failureReason, nil)), nil
	}

	return pluginsCore.DoTransition(p), nil
}

// Handle implements pluginsCore.Plugin. It is invoked for every reconciliation round.
func (m *ClusterPluginManager) Handle(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	state := ClusterPluginState{}
	if v, err := tCtx.PluginStateReader().Get(&state); err != nil {
		if v != clusterPluginStateVersion {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure(errors.CorruptedPluginState,
				fmt.Sprintf("plugin state version mismatch expected [%d] got [%d]", clusterPluginStateVersion, v), nil)), nil
		}
		return pluginsCore.UnknownTransition, errors.Wrapf(errors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	var err error
	var transition pluginsCore.Transition
	newState := state

	switch state.Phase {
	case ClusterPhaseNotStarted:
		transition, newState, err = m.ensureCluster(ctx, tCtx, newState)
	case ClusterPhaseClusterWait:
		transition, newState, err = m.waitForClusterThenLaunchJob(ctx, tCtx, newState)
	default:
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

// getJob builds the job identity resource and stamps the job metadata onto it.
func (m *ClusterPluginManager) getJob(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	o, err := m.plugin.BuildJobIdentityResource(ctx, tCtx.TaskExecutionMetadata())
	if err != nil {
		logger.Errorf(ctx, "Failed to build the Job with name: %v. Error: %v",
			tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return nil, err
	}
	m.addJobMetadata(tCtx.TaskExecutionMetadata(), o, config.GetK8sPluginConfig())
	return o, nil
}

// Abort implements pluginsCore.Plugin. It deletes only the job; the shared cluster is never touched.
func (m *ClusterPluginManager) Abort(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) error {
	logger.Infof(ctx, "KillTask invoked. We will attempt to delete job [%v].",
		tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())

	o, err := m.getJob(ctx, tCtx)
	if err != nil {
		logger.Errorf(ctx, "%v", err)
		return nil
	}

	if err := m.kubeClient.GetClient().Delete(ctx, o); err != nil && !k8serrors.IsNotFound(err) && !k8serrors.IsGone(err) {
		logger.Warningf(ctx, "Failed to abort Job with name: %v/%v. Error: %v", o.GetNamespace(), o.GetName(), err)
		return err
	}

	return nil
}

// Finalize implements pluginsCore.Plugin. It clears finalizers and optionally deletes the job; the
// shared cluster is never touched.
func (m *ClusterPluginManager) Finalize(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) error {
	o, err := m.getJob(ctx, tCtx)
	if err != nil {
		logger.Errorf(ctx, "%v", err)
		return nil
	}

	nsName := k8stypes.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}

	// Clear finalizers
	if err := m.kubeClient.GetClient().Get(ctx, nsName, o); err != nil {
		if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
			return nil
		}
		return err
	}

	if len(o.GetFinalizers()) > 0 {
		o.SetFinalizers([]string{})
		if err := m.kubeClient.GetClient().Update(ctx, o); err != nil {
			if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
				return nil
			}
			logger.Warningf(ctx, "Failed to clear finalizers for Job: %v. Error: %v", nsName, err)
			return err
		}
	}

	cfg := config.GetK8sPluginConfig()
	if cfg.DeleteResourceOnFinalize && !m.plugin.GetProperties().DisableDeleteResourceOnFinalize {
		if err := m.kubeClient.GetClient().Delete(ctx, o); err != nil {
			if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
				return nil
			}
			logger.Warningf(ctx, "Failed to delete Job: %v. Error: %v", nsName, err)
			return err
		}
	}

	return nil
}
