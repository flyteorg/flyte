package k8s

import (
	"context"
	"fmt"
	"sync"
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

const pluginStateVersion = 1

// PluginPhase tracks the high-level phase of the PluginManager's state machine.
type PluginPhase uint8

const (
	PluginPhaseNotStarted PluginPhase = iota
	PluginPhaseStarted
)

// PluginState is the state persisted by the PluginManager between reconciliation rounds.
type PluginState struct {
	Phase               PluginPhase
	K8sPluginState      k8s.PluginState
	LastEventUpdate     time.Time
	LastEventRecordedAt time.Time
}

var _ pluginsCore.Plugin = &PluginManager{}

// PluginManager wraps a k8s.Plugin to implement pluginsCore.Plugin. It manages the lifecycle
// of creating, monitoring, aborting, and finalizing Kubernetes resources for task execution.
type PluginManager struct {
	id         string
	plugin     k8s.Plugin
	kubeClient pluginsCore.KubeClient

	eventWatcher     objectEventWatcher
	eventWatcherOnce sync.Once
	eventWatcherErr  error
}

// NewPluginManager creates a PluginManager that wraps a k8s.Plugin.
func NewPluginManager(id string, plugin k8s.Plugin, kubeClient pluginsCore.KubeClient) *PluginManager {
	return &PluginManager{
		id:         id,
		plugin:     plugin,
		kubeClient: kubeClient,
	}
}

func (pm *PluginManager) GetID() string {
	return pm.id
}

func (pm *PluginManager) GetProperties() pluginsCore.PluginProperties {
	props := pm.plugin.GetProperties()
	return pluginsCore.PluginProperties{
		GeneratedNameMaxLength: props.GeneratedNameMaxLength,
	}
}

func (pm *PluginManager) addObjectMetadata(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig) {
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(pluginsUtils.UnionMaps(cfg.DefaultAnnotations, o.GetAnnotations(), pluginsUtils.CopyMap(taskCtx.GetAnnotations())))
	o.SetLabels(pluginsUtils.UnionMaps(cfg.DefaultLabels, o.GetLabels(), pluginsUtils.CopyMap(taskCtx.GetLabels())))
	o.SetName(taskCtx.GetTaskExecutionID().GetGeneratedName())

	if !pm.plugin.GetProperties().DisableInjectOwnerReferences && !cfg.DisableInjectOwnerReferences {
		o.SetOwnerReferences([]metav1.OwnerReference{taskCtx.GetOwnerReference()})
	}

	if cfg.InjectFinalizer && !pm.plugin.GetProperties().DisableInjectFinalizer {
		f := append(o.GetFinalizers(), "flyte/flytek8s")
		o.SetFinalizers(f)
	}

	if errs := validation.IsDNS1123Subdomain(o.GetName()); len(errs) > 0 {
		o.SetName(pluginsUtils.ConvertToDNS1123SubdomainCompatibleString(o.GetName()))
	}
}

func (pm *PluginManager) launchResource(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	o, err := pm.plugin.BuildResource(ctx, tCtx)
	if err != nil {
		return pluginsCore.UnknownTransition, err
	}

	pm.addObjectMetadata(tCtx.TaskExecutionMetadata(), o, config.GetK8sPluginConfig())
	logger.Infof(ctx, "Creating Object: Type:[%v], Object:[%v/%v]", o.GetObjectKind().GroupVersionKind(), o.GetNamespace(), o.GetName())

	err = pm.kubeClient.GetClient().Create(ctx, o)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		if k8serrors.IsForbidden(err) {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure("RuntimeFailure", err.Error(), nil)), nil
		}
		if k8serrors.IsRequestEntityTooLargeError(err) {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("EntityTooLarge", err.Error(), nil)), nil
		}
		reason := k8serrors.ReasonForError(err)
		logger.Errorf(ctx, "Failed to launch job, system error. err: %v", err)
		return pluginsCore.UnknownTransition, errors.Wrapf(stdErrors.ErrorCode(reason), err, "failed to create resource")
	}

	return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(time.Now(), pluginsCore.DefaultPhaseVersion, "task submitted to K8s")), nil
}

func (pm *PluginManager) getResource(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	o, err := pm.plugin.BuildIdentityResource(ctx, tCtx.TaskExecutionMetadata())
	if err != nil {
		logger.Errorf(ctx, "Failed to build the Resource with name: %v. Error: %v",
			tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return nil, err
	}
	pm.addObjectMetadata(tCtx.TaskExecutionMetadata(), o, config.GetK8sPluginConfig())
	return o, nil
}

func (pm *PluginManager) checkResourcePhase(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, o client.Object, k8sPluginState *k8s.PluginState) (pluginsCore.Transition, error) {
	nsName := k8stypes.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}

	if err := pm.kubeClient.GetClient().Get(ctx, nsName, o); err != nil {
		if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) || k8serrors.IsResourceExpired(err) {
			logger.Warningf(ctx, "Failed to find the Resource with name: %v. Error: %v", nsName, err)
			failureReason := fmt.Sprintf("resource not found, name [%s]. reason: %s", nsName.String(), err.Error())
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure("ResourceDeletedExternally", failureReason, nil)), nil
		}
		logger.Warningf(ctx, "Failed to retrieve Resource Details with name: %v. Error: %v", nsName, err)
		return pluginsCore.UnknownTransition, err
	}

	pCtx := newPluginContext(tCtx, k8sPluginState, pm.kubeClient.GetClient())
	p, err := pm.plugin.GetTaskPhase(ctx, pCtx, o)
	if err != nil {
		logger.Warnf(ctx, "failed to check status of resource in plugin [%s], with error: %s", pm.GetID(), err.Error())
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

	if !p.Phase().IsTerminal() && o.GetDeletionTimestamp() != nil {
		failureReason := fmt.Sprintf("object [%s] terminated unexpectedly in the background", nsName.String())
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure("UnexpectedObjectDeletion", failureReason, nil)), nil
	}

	return pluginsCore.DoTransition(p), nil
}

// Handle implements pluginsCore.Plugin. It is invoked for every reconciliation round.
func (pm *PluginManager) Handle(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	pluginState := PluginState{}
	if v, err := tCtx.PluginStateReader().Get(&pluginState); err != nil {
		if v != pluginStateVersion {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure(errors.CorruptedPluginState,
				fmt.Sprintf("plugin state version mismatch expected [%d] got [%d]", pluginStateVersion, v), nil)), nil
		}
		return pluginsCore.UnknownTransition, errors.Wrapf(errors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	var err error
	var transition pluginsCore.Transition
	pluginPhase := pluginState.Phase
	var resource client.Object

	if pluginState.Phase == PluginPhaseNotStarted {
		transition, err = pm.launchResource(ctx, tCtx)
		if err == nil && transition.Info().Phase() == pluginsCore.PhaseQueued {
			pluginPhase = PluginPhaseStarted
		}
	} else {
		o, getErr := pm.getResource(ctx, tCtx)
		if getErr != nil {
			transition, err = pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition",
				fmt.Sprintf("Failed to build resource, caused by: %s", getErr.Error()), nil)), nil
		} else {
			resource = o
			transition, err = pm.checkResourcePhase(ctx, tCtx, o, &pluginState.K8sPluginState)
		}
	}

	if err != nil {
		return transition, err
	}

	phaseInfo := transition.Info()
	lastEventUpdate := pluginState.LastEventUpdate
	lastEventRecordedAt := pluginState.LastEventRecordedAt
	if resource != nil {
		phaseInfo, lastEventUpdate, lastEventRecordedAt = pm.attachRecentObjectEvents(
			resource,
			phaseInfo,
			pluginState.K8sPluginState,
			lastEventUpdate,
			lastEventRecordedAt,
		)
		transition.SetInfo(phaseInfo)
	}

	newPluginState := PluginState{
		Phase: pluginPhase,
		K8sPluginState: k8s.PluginState{
			Phase:        phaseInfo.Phase(),
			PhaseVersion: phaseInfo.Version(),
			Reason:       phaseInfo.Reason(),
		},
		LastEventUpdate:     lastEventUpdate,
		LastEventRecordedAt: lastEventRecordedAt,
	}
	if pluginState != newPluginState {
		if err := tCtx.PluginStateWriter().Put(pluginStateVersion, &newPluginState); err != nil {
			return pluginsCore.UnknownTransition, err
		}
	}

	return transition, nil
}

func (pm *PluginManager) initEventWatcher(ctx context.Context) {
	if pm.eventWatcher != nil {
		return
	}

	pm.eventWatcherOnce.Do(func() {
		pm.eventWatcher, pm.eventWatcherErr = newControllerRuntimeEventWatcher(ctx, pm.kubeClient.GetCache())
		if pm.eventWatcherErr != nil {
			logger.Warnf(ctx, "Failed to initialize k8s object event watcher for plugin [%s]: %v", pm.GetID(), pm.eventWatcherErr)
		}
	})
}

// InitializeObjectEventWatcher starts watching Kubernetes object events for this plugin.
// It is intended to be called during plugin initialization (before task handling starts).
func (pm *PluginManager) InitializeObjectEventWatcher(ctx context.Context) error {
	pm.initEventWatcher(ctx)
	if pm.eventWatcherErr != nil {
		return fmt.Errorf("failed to initialize k8s object event watcher for plugin %s: %w", pm.GetID(), pm.eventWatcherErr)
	}
	return nil
}

func (pm *PluginManager) attachRecentObjectEvents(
	resource client.Object,
	phaseInfo pluginsCore.PhaseInfo,
	lastObservedState k8s.PluginState,
	lastEventUpdate time.Time,
	lastEventRecordedAt time.Time,
) (pluginsCore.PhaseInfo, time.Time, time.Time) {
	if pm.eventWatcher == nil || resource == nil {
		return phaseInfo, lastEventUpdate, lastEventRecordedAt
	}

	info := phaseInfo.Info()
	if info == nil {
		return phaseInfo, lastEventUpdate, lastEventRecordedAt
	}

	objectKey := watchedObjectKey{
		Namespace: resource.GetNamespace(),
		Name:      resource.GetName(),
		Kind:      resource.GetObjectKind().GroupVersionKind().Kind,
	}
	recentEvents := pm.eventWatcher.List(objectKey, lastEventUpdate, lastEventRecordedAt)
	if len(recentEvents) == 0 {
		return phaseInfo, lastEventUpdate, lastEventRecordedAt
	}

	for _, event := range recentEvents {
		info.AdditionalReasons = append(info.AdditionalReasons, pluginsCore.ReasonInfo{
			Reason:     event.Message,
			OccurredAt: &event.CreatedAt,
		})
		lastEventUpdate = event.CreatedAt
		lastEventRecordedAt = event.RecordedAt
	}

	if phaseInfo.Phase() == lastObservedState.Phase && phaseInfo.Version() <= lastObservedState.PhaseVersion {
		phaseInfo = phaseInfo.WithVersion(lastObservedState.PhaseVersion + 1)
	}

	return phaseInfo, lastEventUpdate, lastEventRecordedAt
}

// Abort implements pluginsCore.Plugin. Called when the task should be killed/aborted.
func (pm *PluginManager) Abort(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) error {
	logger.Infof(ctx, "KillTask invoked. We will attempt to delete object [%v].",
		tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())

	o, err := pm.getResource(ctx, tCtx)
	if err != nil {
		logger.Errorf(ctx, "%v", err)
		return nil
	}

	deleteResource := true
	abortOverride, hasAbortOverride := pm.plugin.(k8s.PluginAbortOverride)

	resourceToFinalize := o
	var behavior k8s.AbortBehavior

	if hasAbortOverride {
		behavior, err = abortOverride.OnAbort(ctx, tCtx, o)
		deleteResource = err == nil && behavior.DeleteResource
		if err == nil && behavior.Resource != nil {
			resourceToFinalize = behavior.Resource
		}
	}

	if err != nil {
		// fall through to error check below
	} else if deleteResource {
		err = pm.kubeClient.GetClient().Delete(ctx, resourceToFinalize)
	} else {
		if behavior.Patch != nil && behavior.Update == nil {
			err = pm.kubeClient.GetClient().Patch(ctx, resourceToFinalize, behavior.Patch.Patch, behavior.Patch.Options...)
		} else if behavior.Patch == nil && behavior.Update != nil {
			err = pm.kubeClient.GetClient().Update(ctx, resourceToFinalize, behavior.Update.Options...)
		} else {
			err = fmt.Errorf("AbortBehavior for resource %v must specify either a Patch or an Update operation if Delete is set to false", resourceToFinalize.GetName())
		}
		if behavior.DeleteOnErr && err != nil {
			logger.Warningf(ctx, "Failed to apply AbortBehavior for resource %v with error %v. Will attempt to delete.", resourceToFinalize.GetName(), err)
			err = pm.kubeClient.GetClient().Delete(ctx, resourceToFinalize)
		}
	}

	if err != nil && !k8serrors.IsNotFound(err) && !k8serrors.IsGone(err) {
		logger.Warningf(ctx, "Failed to abort Resource with name: %v/%v. Error: %v",
			resourceToFinalize.GetNamespace(), resourceToFinalize.GetName(), err)
		return err
	}

	return nil
}

// Finalize implements pluginsCore.Plugin. Called after Handle or Abort to clean up resources.
func (pm *PluginManager) Finalize(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) error {
	o, err := pm.getResource(ctx, tCtx)
	if err != nil {
		logger.Errorf(ctx, "%v", err)
		return nil
	}

	nsName := k8stypes.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}

	// Clear finalizers
	if err := pm.kubeClient.GetClient().Get(ctx, nsName, o); err != nil {
		if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
			return nil
		}
		return err
	}

	if len(o.GetFinalizers()) > 0 {
		o.SetFinalizers([]string{})
		if err := pm.kubeClient.GetClient().Update(ctx, o); err != nil {
			if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
				return nil
			}
			logger.Warningf(ctx, "Failed to clear finalizers for Resource: %v. Error: %v", nsName, err)
			return err
		}
	}

	cfg := config.GetK8sPluginConfig()
	if cfg.DeleteResourceOnFinalize && !pm.plugin.GetProperties().DisableDeleteResourceOnFinalize {
		if err := pm.kubeClient.GetClient().Delete(ctx, o); err != nil {
			if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
				return nil
			}
			logger.Warningf(ctx, "Failed to delete Resource: %v. Error: %v", nsName, err)
			return err
		}
	}

	return nil
}
