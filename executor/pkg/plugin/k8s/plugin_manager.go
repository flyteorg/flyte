package k8s

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/gpufault"
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
	name := taskCtx.GetTaskExecutionID().GetGeneratedName()
	if pm.plugin.GetProperties().GeneratedNameMaxLength != nil {
		if truncatedName, err := taskCtx.GetTaskExecutionID().GetGeneratedNameWith(0, *pm.plugin.GetProperties().GeneratedNameMaxLength); err == nil {
			name = truncatedName
		}
	}
	o.SetName(name)

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
		// Admission/validation rejections (e.g. a generated object or derived child name
		// exceeding k8s length limits) are deterministic: retrying re-submits the identical
		// object and fails the same way, leaving the execution stuck RUNNING. Fast-fail
		// instead of looping via UnknownTransition.
		if k8serrors.IsInvalid(err) {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("InvalidResource", err.Error(), nil)), nil
		}
		// Same for HTTP 400, which is what a validating admission webhook returns when it
		// rejects the spec outright. Distinct code from InvalidResource because a webhook
		// rejection and a field the apiserver itself found invalid are different diagnoses.
		if k8serrors.IsBadRequest(err) {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadRequest", err.Error(), nil)), nil
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
			// Failing to read plugin state is deterministic, the stored bytes don't
			// change between reconciles, so fail permanently instead of retrying.
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemFailureWithCleanup(errors.CorruptedPluginState,
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
		phaseInfo = pm.classifyGpuFailure(ctx, tCtx, resource, phaseInfo)
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

	recentEvents := pm.eventWatcher.List(objectKeyFor(resource), lastEventUpdate, lastEventRecordedAt)
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

func objectKeyFor(resource client.Object) watchedObjectKey {
	return watchedObjectKey{
		Namespace: resource.GetNamespace(),
		Name:      resource.GetName(),
		Kind:      resource.GetObjectKind().GroupVersionKind().Kind,
	}
}

// gpuFaultRelevanceWindow bounds how long before a failure a fault can still explain
// it. Which pod a fault belongs to is settled by the UID when the pod's UID is known
// (see classifyGpuFailure for the one case it is not); the window only separates the
// fault that explains this failure from one the node saw much earlier. It is measured
// from the failure's own time, not from when classification runs, so a slow reconcile
// cannot age a fault out. Thirty minutes spans the slow paths between a fault and the
// failure it causes: a container left wedged after a bus fault until the kubelet gives
// up on it, and a node going NotReady with its pods evicted only after the
// node-monitor grace period and eviction timeout.
//
// A fault that was still firing inside the window counts even if it started before it,
// because what the window bounds is how stale a fault's last sign of life may be, not how
// old the fault is. See faultOverlapsFailure.
const gpuFaultRelevanceWindow = 30 * time.Minute

// gpuFaultAfterFailureSlack is how far past the failure a fault may first be recorded and
// still count. The kernel line and the container's termination are stamped by different
// processes on the same node and the daemon reads the kernel log with a small lag, so a
// fault can first be recorded moments after the failure it caused; a fault that only
// started later than that cannot have caused it.
//
// It bounds when a fault started, not when it stopped. Hardware that keeps faulting after
// the container died goes on being observed for as long as it goes on faulting, and that
// says nothing about whether it caused the failure. See faultOverlapsFailure.
const gpuFaultAfterFailureSlack = 2 * time.Minute

// observedFault is a fault recorded against one pod, kept with the times that order it and
// the pod it was observed on. The pod name is what tells the user which worker of a
// distributed job the fault happened on.
type observedFault struct {
	fault *core.GpuFault
	// createdAt is when the fault was first recorded, which is when it happened. It is
	// what orders faults gathered from several pods into the single sequence
	// ClassifyFailure expects, so that the first fault of a severity is the earliest.
	createdAt time.Time
	podName   string
}

// classifyGpuFailure folds the GPU faults recorded against a failed attempt's pods into
// the failure the plugin reported, so that a fault the node saw becomes the code and the
// message the user reads. Anything that is not a failure is left alone.
//
// The pods it looks at are the tracked resource itself when the plugin tracks a Pod, and
// otherwise the child pods the plugin names through k8s.ChildPodDiscovery. A plugin that
// tracks a CRD and does not implement that interface contributes no pods and so no faults,
// which is exactly what it did before the interface existed.
func (pm *PluginManager) classifyGpuFailure(
	ctx context.Context,
	tCtx pluginsCore.TaskExecutionContext,
	resource client.Object,
	phaseInfo pluginsCore.PhaseInfo,
) pluginsCore.PhaseInfo {
	if pm.eventWatcher == nil || resource == nil || !phaseInfo.Phase().IsFailure() {
		return phaseInfo
	}

	// Every event cached for a pod, not only the ones since the last watermark: the Xid
	// that killed the task is usually recorded rounds before the pod's status catches up
	// with it, and by then the watermark has moved past it. What bounds the search is the
	// identity and the recency of each event.
	//
	// The attempt's own failure time is what a child pod's anchor is bounded by, and what
	// stands in for a pod that offers nothing better. Every CRD plugin stamps it with the
	// time of the reconcile that noticed the failure rather than with anything the kubelet
	// recorded, which is why each pod is anchored on itself below.
	attemptFailedAt := phaseInfoOccurredAt(phaseInfo)

	// A plugin that tracks the Pod is looking at the pod the failure is already reported
	// against, so naming it again would tell the user nothing. Only a CRD's child pods
	// need naming, which is what namesTheFaultingPod tracks.
	var observed []observedFault
	namesTheFaultingPod := false
	if pod, isPod := resource.(*v1.Pod); isPod {
		observed = pm.faultsOnPod(ctx, objectKeyFor(resource), resource.GetUID(), podFailureTime(pod, attemptFailedAt))
	} else {
		observed = pm.faultsOnChildPods(ctx, tCtx, resource, attemptFailedAt)
		namesTheFaultingPod = true
	}
	if len(observed) == 0 {
		return phaseInfo
	}

	// ClassifyFailure reads the list as a sequence in time, taking the first fault of the
	// severity it settles on. Faults gathered from several pods arrive grouped by pod, so
	// they have to be put back in order for first to mean earliest. The order is on when
	// each fault was first recorded, not on when it was last seen: a fault that is still
	// repeating has a later last observation than a one-shot fault that followed it, and
	// ordering on that would let a downstream symptom outrank the root cause.
	sort.SliceStable(observed, func(i, j int) bool {
		return observed[i].createdAt.Before(observed[j].createdAt)
	})

	faults := make([]*core.GpuFault, 0, len(observed))
	for _, o := range observed {
		faults = append(faults, o.fault)
	}

	classified := gpufault.ClassifyFailure(phaseInfo, faults)
	if namesTheFaultingPod {
		classified = attachFaultingPod(classified, observed)
	}
	return classified
}

// gpuFaultCodes are the codes ClassifyFailure puts on a failure it settled with a fault.
// Their presence is how the caller tells a failure the fault explained from one the fault
// only rode along with.
var gpuFaultCodes = sets.NewString(
	gpufault.CodeGpuXidError,
	gpufault.CodeGpuFallenOffBus,
	gpufault.CodeGpuEccUncorrectable,
	gpufault.CodeGpuRowRemapPending,
	gpufault.CodeGpuNvlinkError,
	gpufault.CodeGpuGspError,
)

// attachFaultingPod names the one pod whose fault the classification settled on. The fault
// sentence carries the Xid, the GPU and the node but not the pod, which on a job with many
// workers is not enough to act on, and the pod is not part of the fault contract itself
// because a fault reported as an event on a pod already says which pod it is about to
// anyone reading the event.
//
// Only the fault that decided the verdict is named. Naming every pod that saw any fault
// would put a misleading reason on a failure the faults did not explain, for example a
// plain OOMKilled that a warning happened to coincide with, and would emit one cluster
// event per pod when a whole node's worth of GPUs faults at once.
func attachFaultingPod(phaseInfo pluginsCore.PhaseInfo, observed []observedFault) pluginsCore.PhaseInfo {
	info := phaseInfo.Info()
	if info == nil || !gpuFaultCodes.Has(phaseInfo.Err().GetCode()) {
		return phaseInfo
	}

	// The fault the classification kept is carried on the error, so matching it back by
	// identity finds the pod it came from without repeating the precedence rules.
	settled := phaseInfo.Err().GetGpuFault()
	for _, o := range observed {
		if o.fault != settled || o.podName == "" {
			continue
		}
		occurredAt := o.createdAt
		info.AdditionalReasons = append(info.AdditionalReasons, pluginsCore.ReasonInfo{
			Reason:     fmt.Sprintf("GPU fault recorded on pod %s", o.podName),
			OccurredAt: &occurredAt,
		})
		return phaseInfo
	}

	return phaseInfo
}

// faultsOnChildPods gathers the faults recorded against the pods an operator expanded from
// the resource this plugin tracks. A plugin that cannot name its child pods, either because
// it does not implement the interface or because it declined for this resource, contributes
// nothing: there is no safe wider search, since an operator's child pod names carry a random
// suffix and the events are cached under those names.
func (pm *PluginManager) faultsOnChildPods(
	ctx context.Context,
	tCtx pluginsCore.TaskExecutionContext,
	resource client.Object,
	failureAt time.Time,
) []observedFault {
	discovery, ok := pm.plugin.(k8s.ChildPodDiscovery)
	if !ok || tCtx == nil {
		return nil
	}

	tracked := fmt.Sprintf("%s/%s", resource.GetNamespace(), resource.GetName())

	selector, err := discovery.ChildPods(ctx, tCtx.TaskExecutionMetadata(), resource)
	if err != nil {
		logger.Warnf(ctx, "plugin [%s] failed to name the child pods of %s: %v", pm.GetID(), tracked, err)
		return nil
	}
	if selector == nil {
		logger.Debugf(ctx, "plugin [%s] did not name the child pods of %s, so its GPU faults are not classified",
			pm.GetID(), tracked)
		return nil
	}

	podList := &v1.PodList{}
	listOptions := []client.ListOption{
		client.InNamespace(resource.GetNamespace()),
		client.MatchingLabelsSelector{Selector: selector},
	}
	if err := pm.kubeClient.GetClient().List(ctx, podList, listOptions...); err != nil {
		logger.Warnf(ctx, "failed to list the child pods of %s for GPU fault classification: %v", tracked, err)
		return nil
	}
	if len(podList.Items) == 0 {
		// The pods are gone by the time the failure is classified, most often because the
		// operator tore them down with the job. Nothing can be recovered: the events are
		// cached under pod names that are not derivable without the pods.
		logger.Debugf(ctx, "no child pods left for %s matching %s, so any GPU fault on them is not classified",
			tracked, selector.String())
		return nil
	}

	var observed []observedFault
	for i := range podList.Items {
		pod := &podList.Items[i]
		// A worker that finished its work cannot be the reason the job failed. It may
		// still have logged a fault mid-run, and crediting that to another pod's failure
		// would turn a plain user error into a hardware one. A pod that is still running
		// is kept: a worker wedged on a GPU that fell off the bus is exactly the case
		// this path exists for.
		if podSucceeded(pod) {
			continue
		}
		podKey := watchedObjectKey{Namespace: pod.Namespace, Name: pod.Name, Kind: "Pod"}
		observed = append(observed, pm.faultsOnPod(ctx, podKey, pod.UID, childPodFailureTime(pod, failureAt))...)
	}
	return observed
}

// podSucceeded reports whether this pod finished its work. The phase is the kubelet's own
// verdict; the container check catches the pod whose containers have all exited cleanly
// but whose phase has not caught up yet.
func podSucceeded(pod *v1.Pod) bool {
	if pod.Status.Phase == v1.PodSucceeded {
		return true
	}
	if pod.Status.Phase != v1.PodRunning || len(pod.Status.ContainerStatuses) == 0 {
		return false
	}
	for _, status := range pod.Status.ContainerStatuses {
		terminated := status.State.Terminated
		if terminated == nil || terminated.ExitCode != 0 {
			return false
		}
	}
	return true
}

// childPodFailureTime is podFailureTime bounded by the attempt's own failure.
//
// A child pod is anchored on itself, so a worker that died an hour before the operator
// admitted the job had failed is judged against its own death, which is the whole point of
// anchoring per pod. What it must not do is anchor later than the failure: an operator
// tearing its pods down long after the job failed would otherwise let faults recorded in
// the meantime explain it, and the slack after the failure is deliberately small. The risk
// the earlier direction leaves, a fault from a pod's earlier life explaining an unrelated
// failure, is what excluding succeeded pods above bounds.
func childPodFailureTime(pod *v1.Pod, failureAt time.Time) time.Time {
	anchor := podFailureTime(pod, failureAt)
	if !failureAt.IsZero() && anchor.After(failureAt) {
		return failureAt
	}
	return anchor
}

// faultsOnPod reads the faults recorded against one pod.
//
// It takes every event cached for the pod, not only the ones since the last watermark: the
// Xid that killed the task is usually recorded rounds before the pod's status catches up
// with it, and by then the watermark has moved past it. What bounds the search is the
// identity and the recency of each event, checked below.
func (pm *PluginManager) faultsOnPod(
	ctx context.Context,
	podKey watchedObjectKey,
	podUID k8stypes.UID,
	failureAt time.Time,
) []observedFault {
	events := pm.eventWatcher.List(podKey, time.Time{}, time.Time{})
	if len(events) == 0 {
		return nil
	}

	observed := make([]observedFault, 0, len(events))
	for _, event := range events {
		// Only events the GPU fault emitter wrote, recognized by their reason, are
		// parsed; the message prefix alone is free text anyone can put in an event.
		if event.Reason != gpufault.EventReasonXid && event.Reason != gpufault.EventReasonSXid {
			continue
		}
		// Events are cached under the pod's namespace and name, which a recreated pod
		// reuses, so the fault has to have been recorded against this very pod.
		// Identity is the event's regarding UID against the pod's. An event without one
		// is rejected: the API server does not fill that field, so its absence is a
		// client that did not say which object it meant. The pod's own UID is unknown
		// when the pod was deleted before this round reached it; the name match the
		// cache is keyed on is then all there is, and it is used knowingly: a same-name
		// replacement pod's faults could be credited here, a deliberate trade against
		// losing every fault on the path where the hardware most clearly failed. A child
		// pod always arrives with its UID, because it was read out of the cache as an
		// object rather than named from the task, so it never takes that trade.
		if event.RegardingUID == "" {
			continue
		}
		if podUID != "" && event.RegardingUID != podUID {
			continue
		}
		if !faultOverlapsFailure(event, failureAt) {
			logger.Debugf(ctx,
				"ignoring GPU fault event %q on %s: active %s to %s, which does not reach the failure at %s",
				event.Reason, podKey.Name, event.CreatedAt, event.LastObservedAt, failureAt)
			continue
		}
		if fault := gpufault.FromEventMessage(event.Message); fault != nil {
			// The last observation decided relevance above; what orders the fault against
			// the others is when it was first recorded, which is when it happened.
			observed = append(observed, observedFault{fault: fault, createdAt: event.CreatedAt, podName: podKey.Name})
		}
	}

	return observed
}

// faultOverlapsFailure reports whether a fault event was active close enough to the
// failure to explain it.
//
// A fault that keeps repeating is aggregated into a single event whose last observation
// moves with every repeat, so an event describes an interval and not a moment: it was
// first recorded at CreatedAt and was still firing at LastObservedAt. The failure has an
// interval of its own, the window before it in which a fault could have caused it and the
// small slack after it in which a fault it caused could still be recorded. The event
// counts when those two intervals overlap.
//
// Testing the last observation alone, as this used to, drops the fault that matters most:
// hardware that keeps faulting after the container died has a last observation well past
// the failure, so the longer it goes on the more certainly it was discarded. Testing the
// creation alone drops the opposite case, a fault that started before the window opened
// and was still firing when the task died. Overlap keeps both and still rejects a fault
// that only started after the failure, or one that had stopped firing before the window
// opened.
func faultOverlapsFailure(event *eventInfo, failureAt time.Time) bool {
	activeFrom, activeUntil := event.CreatedAt, event.LastObservedAt
	if activeFrom.IsZero() {
		activeFrom = activeUntil
	}
	if activeUntil.IsZero() {
		activeUntil = activeFrom
	}
	if activeFrom.IsZero() || activeUntil.Before(activeFrom) {
		// No usable time at all, or a last observation older than the creation, which no
		// honest recorder produces. Nothing can be concluded, so it does not explain.
		return false
	}

	relevantFrom := failureAt.Add(-gpuFaultRelevanceWindow)
	relevantUntil := failureAt.Add(gpuFaultAfterFailureSlack)

	return !activeFrom.After(relevantUntil) && !activeUntil.Before(relevantFrom)
}

// phaseInfoOccurredAt is the time the plugin put on the failure, or the zero time when it
// put none there.
func phaseInfoOccurredAt(phaseInfo pluginsCore.PhaseInfo) time.Time {
	if info := phaseInfo.Info(); info != nil && info.OccurredAt != nil {
		return *info.OccurredAt
	}
	return time.Time{}
}

// podFailureTime is the time a pod's own trouble is anchored on, which is what the fault
// relevance interval is centred on.
//
// A container's termination is stamped by the kubelet on the same node and clock as the
// fault events, so it is the closest thing to the moment a fault would have to explain. A
// pod on its way out without a terminated container is anchored on its deletion, which is
// what an eviction leaves behind.
//
// Only then does the plugin's own reported time stand in, and it is the last resort on
// purpose. It comes from GetLastTransitionOccurredAt, which for a pod that failed while
// its containers were still running is the time the container started, not the time
// anything went wrong. Anchoring a long-running task on its own start would put every real
// fault outside the window and quietly classify nothing.
//
// Init containers are not eligible. They finish before the workload starts, and a native
// sidecar declared among them is reaped after everything else, so either would anchor on a
// moment that has nothing to do with when the work died.
func podFailureTime(pod *v1.Pod, occurredAt time.Time) time.Time {
	latest := time.Time{}
	for _, status := range pod.Status.ContainerStatuses {
		terminated := status.State.Terminated
		if terminated == nil || terminated.FinishedAt.IsZero() {
			continue
		}
		if terminated.FinishedAt.After(latest) {
			latest = terminated.FinishedAt.Time
		}
	}

	switch {
	case !latest.IsZero():
		return latest
	case pod.DeletionTimestamp != nil && !pod.DeletionTimestamp.IsZero():
		return pod.DeletionTimestamp.Time
	case !occurredAt.IsZero():
		return occurredAt
	default:
		return time.Now()
	}
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
