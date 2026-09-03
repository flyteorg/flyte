package k8s

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/gpufault"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8sMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// childPodPlugin is a plugin that tracks a CRD and can name the pods an operator expanded
// from it. Only ChildPods is ever called on it, so the rest of k8s.Plugin is left nil.
type childPodPlugin struct {
	k8s.Plugin
	selector labels.Selector
	err      error
}

func (p childPodPlugin) ChildPods(_ context.Context, taskCtx pluginsCore.TaskExecutionMetadata, _ client.Object) (labels.Selector, error) {
	if p.err != nil {
		return nil, p.err
	}
	if p.selector != nil {
		return p.selector, nil
	}
	return flytek8s.AttemptPodSelector(taskCtx), nil
}

var _ k8s.ChildPodDiscovery = childPodPlugin{}

const childPodNamespace = "ns"

// attemptLabels are the labels the framework stamps on the attempt and every plugin merges
// into the pod templates it builds, so the operator's pods carry them.
func attemptLabels() map[string]string {
	return map[string]string{
		flytek8s.ManagedLabelKey: flytek8s.ManagedLabelValue,
		flytek8s.RunLabel:        "run-abc",
		flytek8s.ActionLabel:     "a0",
		flytek8s.AttemptLabel:    "1",
	}
}

func childPodTaskContext(t *testing.T, podLabels map[string]string) pluginsCore.TaskExecutionContext {
	t.Helper()
	meta := &pluginsCoreMock.TaskExecutionMetadata{}
	meta.EXPECT().GetLabels().Return(podLabels)
	tCtx := &pluginsCoreMock.TaskExecutionContext{}
	tCtx.EXPECT().TaskExecutionMetadata().Return(meta)
	return tCtx
}

// trackedJobSet stands in for any CRD a plugin tracks instead of a Pod.
func trackedJobSet() *jobsetv1alpha2.JobSet {
	return &jobsetv1alpha2.JobSet{
		TypeMeta:   metav1.TypeMeta{Kind: "JobSet", APIVersion: jobsetv1alpha2.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Namespace: childPodNamespace, Name: "job", UID: "jobset-uid"},
	}
}

type workerPodOpts struct {
	name       string
	uid        k8stypes.UID
	podLabels  map[string]string
	finishedAt time.Time
	exitCode   int32
	phase      v1.PodPhase
}

func workerPod(opts workerPodOpts) *v1.Pod {
	podLabels := opts.podLabels
	if podLabels == nil {
		podLabels = attemptLabels()
	}
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: childPodNamespace,
			Name:      opts.name,
			UID:       opts.uid,
			Labels:    podLabels,
		},
		Status: v1.PodStatus{Phase: opts.phase},
	}
	if !opts.finishedAt.IsZero() {
		exitCode := opts.exitCode
		if exitCode == 0 && opts.phase != v1.PodSucceeded {
			exitCode = 1
		}
		pod.Status.ContainerStatuses = []v1.ContainerStatus{{
			Name: "primary",
			State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{
				ExitCode:   exitCode,
				FinishedAt: metav1.NewTime(opts.finishedAt),
			}},
		}}
	}
	return pod
}

func childPodManager(t *testing.T, plugin k8s.Plugin, events map[watchedObjectKey][]*eventInfo, pods ...*v1.Pod) *PluginManager {
	t.Helper()
	builder := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme)
	for _, pod := range pods {
		builder = builder.WithObjects(pod)
	}
	kubeClient := &pluginsCoreMock.KubeClient{}
	kubeClient.EXPECT().GetClient().Return(builder.Build()).Maybe()

	pm := NewPluginManager("test-plugin", plugin, kubeClient)
	pm.eventWatcher = &fakeEventWatcher{events: events}
	return pm
}

func podEventKey(name string) watchedObjectKey {
	return watchedObjectKey{Namespace: childPodNamespace, Name: name, Kind: "Pod"}
}

func crdFailure() pluginsCore.PhaseInfo {
	// Every CRD plugin stamps the failure with the time of the reconcile that noticed it,
	// which is why a child pod's own termination is the better anchor.
	now := time.Now()
	return pluginsCore.PhaseInfoRetryableFailure("UnknownError", "JobSet failed", &pluginsCore.TaskInfo{OccurredAt: &now})
}

func TestClassifyGpuFailureOnChildPods(t *testing.T) {
	t.Run("the earliest critical fault across workers explains the failure", func(t *testing.T) {
		// Two workers fault. The later one is the one whose exit the operator reported,
		// but the fault that explains the job is the first one the hardware produced.
		early := time.Now().Add(-3 * time.Minute)
		late := time.Now().Add(-1 * time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-1"): {gpuFaultEventFor(48, gpufault.SeverityCritical, late, late, "worker-1-uid")},
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, early, early, "worker-0-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: early}),
			workerPod(workerPodOpts{name: "job-worker-1", uid: "worker-1-uid", finishedAt: late}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		require.NotNil(t, got.Err())
		assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
		assert.Equal(t, core.ExecutionError_SYSTEM, got.Err().GetKind())
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, got.Phase())
		require.NotNil(t, got.Err().GetGpuFault())
		assert.EqualValues(t, 79, got.Err().GetGpuFault().GetCode())
	})

	t.Run("the pods a fault was seen on are named for the user", func(t *testing.T) {
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, "worker-0-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: observed}),
			workerPod(workerPodOpts{name: "job-worker-1", uid: "worker-1-uid", finishedAt: observed}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		require.NotNil(t, got.Info())
		reasons := make([]string, 0, len(got.Info().AdditionalReasons))
		for _, reason := range got.Info().AdditionalReasons {
			reasons = append(reasons, reason.Reason)
		}
		// Only the worker that actually faulted is named, so the user is not sent to
		// read the logs of every pod in the job.
		assert.Equal(t, []string{"GPU fault recorded on pod job-worker-0"}, reasons)
	})

	t.Run("a plugin that cannot name its child pods classifies nothing", func(t *testing.T) {
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, "worker-0-uid")},
		}
		// The run label did not survive sanitization, so the attempt cannot be identified.
		podLabels := attemptLabels()
		delete(podLabels, flytek8s.RunLabel)
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: observed}),
		)

		phase := crdFailure()
		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, podLabels), trackedJobSet(), phase)

		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
	})

	// A plugin that tracks a CRD and does not implement ChildPodDiscovery has to behave
	// exactly as it did before the interface existed. Everything here is real: a plugin
	// that genuinely lacks the method, a task context that is present, and fault events
	// well inside the relevance window keyed to a pod that genuinely exists and matches
	// the attempt, so the only reason nothing is classified is the missing interface.
	t.Run("a plugin that does not implement the interface is left alone", func(t *testing.T) {
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, "worker-0-uid")},
		}
		plugin := k8sMocks.NewPlugin(t)
		require.NotImplements(t, (*k8s.ChildPodDiscovery)(nil), plugin)

		pm := childPodManager(t, plugin, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: observed}),
		)

		// The same fixture reclassifies when the plugin can name its pods, so the
		// assertions below are about the interface and nothing else.
		withDiscovery := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: observed}),
		)
		reclassified := withDiscovery.classifyGpuFailure(
			context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())
		require.Equal(t, gpufault.CodeGpuFallenOffBus, reclassified.Err().GetCode())

		phase := crdFailure()
		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), phase)

		assert.Equal(t, phase.Phase(), got.Phase())
		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
		assert.Empty(t, got.Info().AdditionalReasons)
	})

	t.Run("a plugin that fails to name its child pods classifies nothing", func(t *testing.T) {
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, "worker-0-uid")},
		}
		pm := childPodManager(t, childPodPlugin{err: assert.AnError}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: observed}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
	})

	t.Run("a pod from another attempt of the same action is not consulted", func(t *testing.T) {
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, "worker-0-uid")},
		}
		// The pod left behind by the previous attempt still carries the run and action,
		// but not this attempt.
		previousAttempt := attemptLabels()
		previousAttempt[flytek8s.AttemptLabel] = "1"
		current := attemptLabels()
		current[flytek8s.AttemptLabel] = "2"

		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", podLabels: previousAttempt, finishedAt: observed}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, current), trackedJobSet(), crdFailure())

		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
	})

	t.Run("a pod that is already gone yields nothing rather than a wrong answer", func(t *testing.T) {
		// The operator tore the pods down with the job. The fault is still cached under
		// the pod's name, but nothing can name that pod any more, and an operator's child
		// pod name is not derivable because it carries a random suffix.
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, "worker-0-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
	})

	t.Run("a fault recorded against an earlier incarnation of the pod is rejected", func(t *testing.T) {
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, "an-older-pod-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: observed}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
	})

	t.Run("relevance is anchored on the pod that terminated, not on the reconcile", func(t *testing.T) {
		// The worker died and its fault was recorded two hours ago; the operator only
		// reported the job as failed on this reconcile. Anchoring on the reconcile would
		// age the fault out, anchoring on the pod's own termination keeps it.
		diedAt := time.Now().Add(-2 * time.Hour)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, diedAt, diedAt, "worker-0-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: diedAt}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
	})

	t.Run("a fault long before the pod terminated does not explain the failure", func(t *testing.T) {
		diedAt := time.Now().Add(-time.Minute)
		longBefore := diedAt.Add(-2 * gpuFaultRelevanceWindow)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, longBefore, longBefore, "worker-0-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: diedAt}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
	})

	t.Run("a critical fault on one worker outranks a user fault on another", func(t *testing.T) {
		// The user fault came first in time, but severity decides which fault names the
		// failure, and a critical fault is not the workload's doing.
		userAt := time.Now().Add(-3 * time.Minute)
		criticalAt := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(31, gpufault.SeverityUser, userAt, userAt, "worker-0-uid")},
			podEventKey("job-worker-1"): {gpuFaultEventFor(79, gpufault.SeverityCritical, criticalAt, criticalAt, "worker-1-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: userAt}),
			workerPod(workerPodOpts{name: "job-worker-1", uid: "worker-1-uid", finishedAt: criticalAt}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
		assert.Equal(t, core.ExecutionError_SYSTEM, got.Err().GetKind())
	})
}

func TestClassifyGpuFailureOnChildPodsSelection(t *testing.T) {
	t.Run("a worker that finished its work does not contribute its faults", func(t *testing.T) {
		// The classic shape of this: an MPI worker exits 0 after logging a critical Xid
		// mid-run, while the launcher fails on the user's own code. Crediting the worker's
		// fault would turn a user error into a system one and stop charging the retry.
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, "worker-0-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{
				name: "job-worker-0", uid: "worker-0-uid", finishedAt: observed,
				exitCode: 0, phase: v1.PodSucceeded,
			}),
			workerPod(workerPodOpts{name: "job-launcher", uid: "launcher-uid", finishedAt: observed, phase: v1.PodFailed}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
	})

	t.Run("a worker still running does contribute its faults", func(t *testing.T) {
		// A worker wedged on a GPU that fell off the bus never terminates, and it is
		// exactly the case this path exists for.
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, "worker-0-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", phase: v1.PodRunning}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
	})

	t.Run("the fault that happened first wins, not the one last seen", func(t *testing.T) {
		// The root cause is still repeating, so its last observation is newer than that of
		// the one-shot fault that followed it. Ordering on the last observation would let
		// the downstream symptom name the failure.
		rootCauseAt := time.Now().Add(-10 * time.Minute)
		stillRepeatingAt := time.Now().Add(-time.Minute)
		downstreamAt := time.Now().Add(-5 * time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {
				gpuFaultEventFor(79, gpufault.SeverityCritical, rootCauseAt, stillRepeatingAt, "worker-0-uid"),
			},
			podEventKey("job-worker-1"): {
				gpuFaultEventFor(48, gpufault.SeverityCritical, downstreamAt, downstreamAt, "worker-1-uid"),
			},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", phase: v1.PodRunning}),
			workerPod(workerPodOpts{name: "job-worker-1", uid: "worker-1-uid", finishedAt: downstreamAt}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
		require.Len(t, got.Info().AdditionalReasons, 1)
		assert.Equal(t, "GPU fault recorded on pod job-worker-0", got.Info().AdditionalReasons[0].Reason)
	})
}

func TestAttachFaultingPod(t *testing.T) {
	t.Run("names only the pod whose fault settled the verdict", func(t *testing.T) {
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(92, gpufault.SeverityWarn, observed, observed, "worker-0-uid")},
			podEventKey("job-worker-1"): {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, "worker-1-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: observed}),
			workerPod(workerPodOpts{name: "job-worker-1", uid: "worker-1-uid", finishedAt: observed}),
		)

		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), crdFailure())

		// The warning on worker-0 is in window and rode along, but the critical on
		// worker-1 is what named the failure, so worker-0 must not be pointed at.
		require.Len(t, got.Info().AdditionalReasons, 1)
		assert.Equal(t, "GPU fault recorded on pod job-worker-1", got.Info().AdditionalReasons[0].Reason)
	})

	t.Run("names nothing when the faults did not explain the failure", func(t *testing.T) {
		// A task that ran out of memory keeps that verdict, and a warning that happened to
		// coincide rides along as data only. Pointing at a pod here would send the user
		// looking for hardware trouble that did not cause anything.
		observed := time.Now().Add(-time.Minute)
		events := map[watchedObjectKey][]*eventInfo{
			podEventKey("job-worker-0"): {gpuFaultEventFor(92, gpufault.SeverityWarn, observed, observed, "worker-0-uid")},
		}
		pm := childPodManager(t, childPodPlugin{}, events,
			workerPod(workerPodOpts{name: "job-worker-0", uid: "worker-0-uid", finishedAt: observed}),
		)

		now := time.Now()
		phase := pluginsCore.PhaseInfoFailure("OOMKilled", "out of memory", &pluginsCore.TaskInfo{OccurredAt: &now})
		got := pm.classifyGpuFailure(context.Background(), childPodTaskContext(t, attemptLabels()), trackedJobSet(), phase)

		assert.Equal(t, "OOMKilled", got.Err().GetCode())
		assert.NotNil(t, got.Err().GetGpuFault())
		assert.Empty(t, got.Info().AdditionalReasons)
	})

	t.Run("a pod that names itself is not named again", func(t *testing.T) {
		// On the single pod path the failure is already reported against that pod, so a
		// reason pointing at it would be self-referential noise.
		base := time.Now().Add(-time.Minute)
		watcher := &fakeEventWatcher{events: map[watchedObjectKey][]*eventInfo{
			{Namespace: "ns", Name: "pod", Kind: "Pod"}: {gpuFaultEvent(79, gpufault.SeverityCritical, base)},
		}}
		pm := NewPluginManager("test-plugin", nil, nil)
		pm.eventWatcher = watcher

		phase := pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", &pluginsCore.TaskInfo{})
		got := pm.classifyGpuFailure(context.Background(), nil, failedPod(), phase)

		assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
		assert.Empty(t, got.Info().AdditionalReasons)
	})
}

// The anchor itself is podFailureTime, covered in plugin_manager_test.go. What is specific
// to a child pod is the bound against the attempt's own failure.
func TestChildPodFailureTime(t *testing.T) {
	failureAt := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)

	t.Run("anchors the pod on itself when it died before the job failed", func(t *testing.T) {
		diedAt := failureAt.Add(-40 * time.Minute)
		pod := &v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{
			{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.NewTime(diedAt)}}},
		}}}
		assert.Equal(t, diedAt, childPodFailureTime(pod, failureAt))
	})

	t.Run("never anchors later than the attempt's own failure", func(t *testing.T) {
		// A pod torn down twenty minutes after the job failed must not let faults recorded
		// in the meantime explain it; the slack past the failure is deliberately small.
		deletion := metav1.NewTime(failureAt.Add(20 * time.Minute))
		pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &deletion}}
		assert.Equal(t, failureAt, childPodFailureTime(pod, failureAt))

		late := metav1.NewTime(failureAt.Add(20 * time.Minute))
		pod = &v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{
			{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: late}}},
		}}}
		assert.Equal(t, failureAt, childPodFailureTime(pod, failureAt))
	})

	t.Run("does not bound against a failure time the plugin never reported", func(t *testing.T) {
		// With no reported time podFailureTime falls back to now, and bounding that
		// against the zero time would anchor every pod in 1970.
		assert.WithinDuration(t, time.Now(), childPodFailureTime(&v1.Pod{}, time.Time{}), time.Minute)
	})
}
