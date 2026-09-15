package k8s

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"k8s.io/apimachinery/pkg/labels"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/gpufault"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8sMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestAddObjectMetadata_GeneratedNameMaxLength(t *testing.T) {
	longName := "a-very-long-task-execution-generated-name-that-exceeds-the-limit-12345"

	t.Run("nil GeneratedNameMaxLength preserves generated name", func(t *testing.T) {
		mockPlugin := k8sMocks.NewPlugin(t)
		mockPlugin.EXPECT().GetProperties().Return(k8s.PluginProperties{
			GeneratedNameMaxLength: nil,
		})

		pm := NewPluginManager("test-plugin", mockPlugin, nil)

		taskMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
		taskMetadata.EXPECT().GetNamespace().Return("test-ns")
		taskMetadata.EXPECT().GetAnnotations().Return(nil)
		taskMetadata.EXPECT().GetLabels().Return(nil)
		taskMetadata.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})

		tID := &pluginsCoreMock.TaskExecutionID{}
		tID.EXPECT().GetGeneratedName().Return(longName)
		taskMetadata.EXPECT().GetTaskExecutionID().Return(tID)

		pod := &v1.Pod{}
		pm.addObjectMetadata(taskMetadata, pod, &config.K8sPluginConfig{})

		assert.Equal(t, longName, pod.GetName())
		assert.Equal(t, "test-ns", pod.GetNamespace())
	})

	t.Run("GeneratedNameMaxLength set truncates long generated name", func(t *testing.T) {
		maxLength := 20
		mockPlugin := k8sMocks.NewPlugin(t)
		mockPlugin.EXPECT().GetProperties().Return(k8s.PluginProperties{
			GeneratedNameMaxLength: ptr.To(maxLength),
		})

		pm := NewPluginManager("test-plugin", mockPlugin, nil)

		taskMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
		taskMetadata.EXPECT().GetNamespace().Return("test-ns")
		taskMetadata.EXPECT().GetAnnotations().Return(nil)
		taskMetadata.EXPECT().GetLabels().Return(nil)
		taskMetadata.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})

		tID := &pluginsCoreMock.TaskExecutionID{}
		tID.EXPECT().GetGeneratedName().Return(longName)
		tID.EXPECT().GetGeneratedNameWith(0, maxLength).Return(encoding.FixedLengthUniqueID(longName, maxLength))
		taskMetadata.EXPECT().GetTaskExecutionID().Return(tID)

		pod := &v1.Pod{}
		pm.addObjectMetadata(taskMetadata, pod, &config.K8sPluginConfig{})

		expectedName, err := encoding.FixedLengthUniqueID(longName, maxLength)
		require.NoError(t, err)

		assert.Equal(t, expectedName, pod.GetName())
		assert.LessOrEqual(t, len(pod.GetName()), maxLength)
		assert.Equal(t, "test-ns", pod.GetNamespace())
	})

	t.Run("GeneratedNameMaxLength set with short name preserves name", func(t *testing.T) {
		maxLength := 50
		shortName := "short-name"
		mockPlugin := k8sMocks.NewPlugin(t)
		mockPlugin.EXPECT().GetProperties().Return(k8s.PluginProperties{
			GeneratedNameMaxLength: ptr.To(maxLength),
		})

		pm := NewPluginManager("test-plugin", mockPlugin, nil)

		taskMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
		taskMetadata.EXPECT().GetNamespace().Return("test-ns")
		taskMetadata.EXPECT().GetAnnotations().Return(nil)
		taskMetadata.EXPECT().GetLabels().Return(nil)
		taskMetadata.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})

		tID := &pluginsCoreMock.TaskExecutionID{}
		tID.EXPECT().GetGeneratedName().Return(shortName)
		tID.EXPECT().GetGeneratedNameWith(0, maxLength).Return(shortName, nil)
		taskMetadata.EXPECT().GetTaskExecutionID().Return(tID)

		pod := &v1.Pod{}
		pm.addObjectMetadata(taskMetadata, pod, &config.K8sPluginConfig{})

		assert.Equal(t, shortName, pod.GetName())
	})
}

// metadataMock builds a TaskExecutionMetadata sufficient for addObjectMetadata, returning the
// given generated name.
func metadataMock(generatedName string) *pluginsCoreMock.TaskExecutionMetadata {
	tID := &pluginsCoreMock.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return(generatedName).Maybe()

	meta := &pluginsCoreMock.TaskExecutionMetadata{}
	meta.EXPECT().GetTaskExecutionID().Return(tID).Maybe()
	meta.EXPECT().GetNamespace().Return("ns")
	meta.EXPECT().GetAnnotations().Return(nil)
	meta.EXPECT().GetLabels().Return(nil)
	meta.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})
	return meta
}

// TestLaunchResource_InvalidCreateFastFails verifies that a deterministic admission/validation
// rejection (e.g. a derived name exceeding k8s length limits) fast-fails instead of looping via
// UnknownTransition and leaving the execution stuck RUNNING.
func TestLaunchResource_InvalidCreateFastFails(t *testing.T) {
	invalidErr := k8serrors.NewInvalid(
		schema.GroupKind{Kind: "JobSet"}, "too-long-name",
		field.ErrorList{field.Invalid(field.NewPath("metadata", "name"), "x", "must be no more than 63 characters")},
	)
	fakeClient := fake.NewClientBuilder().
		WithScheme(k8sscheme.Scheme).
		WithInterceptorFuncs(interceptor.Funcs{
			Create: func(context.Context, client.WithWatch, client.Object, ...client.CreateOption) error {
				return invalidErr
			},
		}).
		Build()

	kubeClient := &pluginsCoreMock.KubeClient{}
	kubeClient.EXPECT().GetClient().Return(fakeClient)

	plugin := &k8sMocks.Plugin{}
	plugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})
	plugin.EXPECT().BuildResource(mock.Anything, mock.Anything).Return(&v1.Pod{}, nil)

	tCtx := &pluginsCoreMock.TaskExecutionContext{}
	tCtx.EXPECT().TaskExecutionMetadata().Return(metadataMock("too-long-name"))

	pm := NewPluginManager("test", plugin, kubeClient)

	transition, err := pm.launchResource(context.Background(), tCtx)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhasePermanentFailure, transition.Info().Phase())
	assert.Equal(t, "InvalidResource", transition.Info().Err().GetCode())
}

// A validating admission webhook rejecting the pod spec surfaces as HTTP 400, which is neither
// Forbidden nor Invalid: without its own branch it reaches the catch-all and the executor keeps
// retrying a request that can never succeed.
func TestLaunchResourceErrors(t *testing.T) {
	tests := []struct {
		name      string
		createErr error
		wantPhase pluginsCore.Phase
		wantCode  string
		wantErr   bool
	}{
		{
			name:      "webhook 400 fails",
			createErr: k8serrors.NewBadRequest(`admission webhook "validate.example.com" denied the request: spec.containers[0].image is required`),
			wantPhase: pluginsCore.PhasePermanentFailure,
			wantCode:  "BadRequest",
		},
		{
			// A rejection can also arrive with no reason set and a bare 400; IsBadRequest
			// falls back to the status code for that shape.
			name: "bare 400 fails",
			createErr: &k8serrors.StatusError{ErrStatus: metav1.Status{
				Status:  metav1.StatusFailure,
				Code:    http.StatusBadRequest,
				Message: "the server rejected our request",
			}},
			wantPhase: pluginsCore.PhasePermanentFailure,
			wantCode:  "BadRequest",
		},
		{
			name: "invalid fails",
			createErr: k8serrors.NewInvalid(
				schema.GroupKind{Kind: "Pod"}, "name",
				field.ErrorList{field.Invalid(field.NewPath("metadata", "name"), "x", "must be no more than 63 characters")},
			),
			wantPhase: pluginsCore.PhasePermanentFailure,
			wantCode:  "InvalidResource",
		},
		{
			name:      "forbidden retries",
			createErr: k8serrors.NewForbidden(schema.GroupResource{Resource: "pods"}, "name", fmt.Errorf("exceeded quota")),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  "RuntimeFailure",
		},
		{
			name:      "unknown retries",
			createErr: k8serrors.NewInternalError(fmt.Errorf("etcd unavailable")),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(k8sscheme.Scheme).
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(context.Context, client.WithWatch, client.Object, ...client.CreateOption) error {
						return tt.createErr
					},
				}).
				Build()

			kubeClient := &pluginsCoreMock.KubeClient{}
			kubeClient.EXPECT().GetClient().Return(fakeClient)

			plugin := &k8sMocks.Plugin{}
			plugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})
			plugin.EXPECT().BuildResource(mock.Anything, mock.Anything).Return(&v1.Pod{}, nil)

			tCtx := &pluginsCoreMock.TaskExecutionContext{}
			tCtx.EXPECT().TaskExecutionMetadata().Return(metadataMock("name"))

			pm := NewPluginManager("test", plugin, kubeClient)

			transition, err := pm.launchResource(context.Background(), tCtx)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, pluginsCore.UnknownTransition, transition)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantPhase, transition.Info().Phase())
			assert.Equal(t, tt.wantCode, transition.Info().Err().GetCode())
			assert.Equal(t, core.ExecutionError_USER, transition.Info().Err().GetKind())
		})
	}
}

func TestHandle_CorruptedPluginStateFailsPermanently(t *testing.T) {
	stateReader := &pluginsCoreMock.PluginStateReader{}
	// (0, err) matches PluginStateManager's contract: the version is zeroed on decode failure.
	stateReader.EXPECT().Get(mock.Anything).Return(uint8(0), fmt.Errorf("gob: decode failed"))

	tCtx := &pluginsCoreMock.TaskExecutionContext{}
	tCtx.EXPECT().PluginStateReader().Return(stateReader)

	plugin := &k8sMocks.Plugin{}
	plugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})

	pm := NewPluginManager("test", plugin, &pluginsCoreMock.KubeClient{})

	transition, err := pm.Handle(context.Background(), tCtx)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhasePermanentFailure, transition.Info().Phase())
	assert.Equal(t, string(errors.CorruptedPluginState), transition.Info().Err().GetCode())
	assert.Equal(t, core.ExecutionError_SYSTEM, transition.Info().Err().GetKind())
}

// fakeEventWatcher stands in for the informer-backed watcher and records the window it
// was asked for, so a test can assert that classification looks at the whole attempt.
type fakeEventWatcher struct {
	events            map[watchedObjectKey][]*eventInfo
	lastCreatedAfter  time.Time
	lastRecordedAfter time.Time
}

func (w *fakeEventWatcher) List(objectKey watchedObjectKey, createdAfter time.Time, recordedAfter time.Time) []*eventInfo {
	w.lastCreatedAfter = createdAfter
	w.lastRecordedAfter = recordedAfter
	return w.events[objectKey]
}

// testPodUID is the UID of the pod the fault events are recorded against, so that a test
// can hand classification an event that belongs to some other incarnation of the pod.
const testPodUID k8stypes.UID = "pod-uid"

func gpuFaultEvent(code int, severity gpufault.Severity, createdAt time.Time) *eventInfo {
	return gpuFaultEventFor(code, severity, createdAt, createdAt, testPodUID)
}

func gpuFaultEventFor(
	code int,
	severity gpufault.Severity,
	createdAt time.Time,
	lastObservedAt time.Time,
	regardingUID k8stypes.UID,
) *eventInfo {
	message := gpufault.FormatEventMessage(
		gpufault.Fault{
			Kind:     gpufault.KindXid,
			Code:     code,
			Name:     gpufault.NameFor(gpufault.KindXid, code),
			Severity: severity,
			PCI:      "0000:3b:00.0",
		},
		gpufault.Attribution{NodeName: "ip-10-0-0-1", GPUUUID: "GPU-1234", GPUIndex: 0},
	)
	return &eventInfo{
		Message:        message,
		Reason:         "GPUXidError",
		CreatedAt:      createdAt,
		RecordedAt:     createdAt,
		LastObservedAt: lastObservedAt,
		RegardingUID:   regardingUID,
	}
}

func failedPod() *v1.Pod {
	pod := &v1.Pod{
		TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "pod", UID: testPodUID},
	}
	return pod
}

func TestClassifyGpuFailure(t *testing.T) {
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}
	// Recency is measured against the clock now, so the fixtures have to sit relative to
	// it: base is inside the relevance window, stale is well outside it.
	base := time.Now().Add(-time.Minute)
	stale := time.Now().Add(-2 * gpufault.RelevanceWindow)

	tests := []struct {
		name        string
		events      []*eventInfo
		phaseInfo   pluginsCore.PhaseInfo
		wantPhase   pluginsCore.Phase
		wantCode    string
		wantKind    core.ExecutionError_ErrorKind
		wantFault   bool
		wantMessage string
	}{
		{
			name:      "a critical xid makes the failure a system retryable one",
			events:    []*eventInfo{gpuFaultEvent(79, gpufault.SeverityCritical, base)},
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", nil),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  gpufault.CodeGpuFallenOffBus,
			wantKind:  core.ExecutionError_SYSTEM,
			wantFault: true,
		},
		{
			name:      "a user xid names the failure but keeps the verdict",
			events:    []*eventInfo{gpuFaultEvent(31, gpufault.SeverityUser, base)},
			phaseInfo: pluginsCore.PhaseInfoFailure("UnknownError", "exit code 1", nil),
			wantPhase: pluginsCore.PhasePermanentFailure,
			wantCode:  gpufault.CodeGpuXidError,
			wantKind:  core.ExecutionError_USER,
			wantFault: true,
		},
		{
			name:      "a warning only rides along",
			events:    []*eventInfo{gpuFaultEvent(92, gpufault.SeverityWarn, base)},
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("OOMKilled", "oom", nil),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  "OOMKilled",
			wantKind:  core.ExecutionError_USER,
			wantFault: true,
		},
		{
			name: "events that are not gpu faults are ignored",
			events: []*eventInfo{
				{Message: "Back-off restarting failed container", CreatedAt: base, RecordedAt: base},
			},
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("OOMKilled", "oom", nil),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  "OOMKilled",
			wantKind:  core.ExecutionError_USER,
			wantFault: false,
		},
		{
			name: "a gpu-health message under an ordinary reason is not trusted",
			events: []*eventInfo{
				{Message: "Back-off restarting failed container", CreatedAt: base, RecordedAt: base},
			},
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("OOMKilled", "oom", nil),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  "OOMKilled",
			wantKind:  core.ExecutionError_USER,
			wantFault: false,
		},
		{
			name:      "no events at all",
			events:    nil,
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("OOMKilled", "oom", nil),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  "OOMKilled",
			wantKind:  core.ExecutionError_USER,
			wantFault: false,
		},
		{
			name: "the first critical fault of the attempt wins",
			events: []*eventInfo{
				gpuFaultEvent(31, gpufault.SeverityUser, base),
				gpuFaultEvent(74, gpufault.SeverityCritical, base.Add(time.Second)),
				gpuFaultEvent(79, gpufault.SeverityCritical, base.Add(2*time.Second)),
			},
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", nil),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  gpufault.CodeGpuNvlinkError,
			wantKind:  core.ExecutionError_SYSTEM,
			wantFault: true,
		},
		{
			name: "a fault recorded against an earlier pod of the same name is ignored",
			events: []*eventInfo{
				gpuFaultEventFor(79, gpufault.SeverityCritical, base, base, "some-other-pod-uid"),
			},
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("OOMKilled", "oom", nil),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  "OOMKilled",
			wantKind:  core.ExecutionError_USER,
			wantFault: false,
		},
		{
			name: "an old fault still being observed now is classified",
			events: []*eventInfo{
				gpuFaultEventFor(79, gpufault.SeverityCritical, stale, base, testPodUID),
			},
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", nil),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  gpufault.CodeGpuFallenOffBus,
			wantKind:  core.ExecutionError_SYSTEM,
			wantFault: true,
		},
		{
			name: "a fault last observed long ago is ignored",
			events: []*eventInfo{
				gpuFaultEventFor(79, gpufault.SeverityCritical, stale, stale, testPodUID),
			},
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("OOMKilled", "oom", nil),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  "OOMKilled",
			wantKind:  core.ExecutionError_USER,
			wantFault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher := &fakeEventWatcher{events: map[watchedObjectKey][]*eventInfo{key: tt.events}}
			pm := NewPluginManager("test-plugin", nil, nil)
			pm.eventWatcher = watcher

			got := pm.classifyGpuFailure(failedPod(), tt.phaseInfo)

			assert.Equal(t, tt.wantPhase, got.Phase())
			require.NotNil(t, got.Err())
			assert.Equal(t, tt.wantCode, got.Err().GetCode())
			assert.Equal(t, tt.wantKind, got.Err().GetKind())
			if tt.wantFault {
				assert.NotNil(t, got.Err().GetGpuFault())
			} else {
				assert.Nil(t, got.Err().GetGpuFault())
			}
			// The search itself is unbounded: which events count is decided per event,
			// on the pod they name and on when they were last observed.
			assert.True(t, watcher.lastCreatedAfter.IsZero())
			assert.True(t, watcher.lastRecordedAfter.IsZero())
		})
	}
}

func TestClassifyGpuFailureRelevanceIsAnchoredOnTheFailure(t *testing.T) {
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}
	failedAt := time.Now().Add(-2 * time.Hour)
	phase := pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", &pluginsCore.TaskInfo{OccurredAt: &failedAt})

	t.Run("a fault just before an old failure still classifies a late reconcile", func(t *testing.T) {
		observed := failedAt.Add(-5 * time.Minute)
		watcher := &fakeEventWatcher{events: map[watchedObjectKey][]*eventInfo{
			key: {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, testPodUID)},
		}}
		pm := NewPluginManager("test-plugin", nil, nil)
		pm.eventWatcher = watcher
		got := pm.classifyGpuFailure(failedPod(), phase)
		assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
	})

	t.Run("a fault observed well after the failure does not explain it", func(t *testing.T) {
		observed := failedAt.Add(10 * time.Minute)
		watcher := &fakeEventWatcher{events: map[watchedObjectKey][]*eventInfo{
			key: {gpuFaultEventFor(79, gpufault.SeverityCritical, observed, observed, testPodUID)},
		}}
		pm := NewPluginManager("test-plugin", nil, nil)
		pm.eventWatcher = watcher
		got := pm.classifyGpuFailure(failedPod(), phase)
		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
	})
}

// TestClassifyGpuFailureRelevanceIsAnInterval covers the shape of a fault event that a
// single timestamp cannot express. A fault that keeps repeating is aggregated into one
// event whose last observation moves with every repeat, so the event says it was first
// recorded at one time and was still firing at another. What decides relevance is whether
// that stretch of time reaches the failure.
func TestClassifyGpuFailureRelevanceIsAnInterval(t *testing.T) {
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}
	failedAt := time.Now().Add(-30 * time.Minute)
	phase := pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", &pluginsCore.TaskInfo{OccurredAt: &failedAt})

	classify := func(t *testing.T, createdAt, lastObservedAt time.Time) pluginsCore.PhaseInfo {
		t.Helper()
		watcher := &fakeEventWatcher{events: map[watchedObjectKey][]*eventInfo{
			key: {gpuFaultEventFor(79, gpufault.SeverityCritical, createdAt, lastObservedAt, testPodUID)},
		}}
		pm := NewPluginManager("test-plugin", nil, nil)
		pm.eventWatcher = watcher
		return pm.classifyGpuFailure(failedPod(), phase)
	}

	t.Run("a fault that started before the failure and kept firing after it explains it", func(t *testing.T) {
		// The GPU faulted a minute before the container died and went on faulting for
		// eleven minutes afterwards, which is what dying hardware does. Judging this by
		// its last observation alone would discard it, and the longer the hardware kept
		// faulting the more certainly it would be discarded.
		got := classify(t, failedAt.Add(-time.Minute), failedAt.Add(11*time.Minute))
		assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
		assert.NotNil(t, got.Err().GetGpuFault())
	})

	t.Run("a fault older than the window but still firing at the failure explains it", func(t *testing.T) {
		// First recorded forty minutes before the failure, so outside the window, but
		// still going when the task died. Judging this by its creation alone would
		// discard it.
		got := classify(t, failedAt.Add(-40*time.Minute), failedAt)
		assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
		assert.NotNil(t, got.Err().GetGpuFault())
	})

	t.Run("a fault that only started after the failure does not explain it", func(t *testing.T) {
		started := failedAt.Add(gpufault.AfterFailureSlack + time.Minute)
		got := classify(t, started, started.Add(5*time.Minute))
		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
	})

	t.Run("a fault that stopped firing before the window opened does not explain it", func(t *testing.T) {
		got := classify(t, failedAt.Add(-90*time.Minute), failedAt.Add(-40*time.Minute))
		assert.Equal(t, "UnknownError", got.Err().GetCode())
		assert.Nil(t, got.Err().GetGpuFault())
	})
}

// TestClassifyGpuFailureAnchorsOnThePodNotItsStartTime covers the pod that failed without
// any container terminating. GetLastTransitionOccurredAt then reports the time the running
// container started, so the failure the plugin hands over is stamped hours before anything
// went wrong, and anchoring on it would put every real fault outside the window.
func TestClassifyGpuFailureAnchorsOnThePodNotItsStartTime(t *testing.T) {
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}

	startedAt := time.Now().Add(-6 * time.Hour)
	evictedAt := time.Now().Add(-2 * time.Minute)
	faultedAt := evictedAt.Add(-time.Minute)

	// A long-running task, evicted while its container was still running. The plugin's
	// reported time is the container's start.
	pod := failedPod()
	deletion := metav1.NewTime(evictedAt)
	pod.DeletionTimestamp = &deletion
	pod.Status.ContainerStatuses = []v1.ContainerStatus{{
		Name:  "primary",
		State: v1.ContainerState{Running: &v1.ContainerStateRunning{StartedAt: metav1.NewTime(startedAt)}},
	}}

	phase := pluginsCore.PhaseInfoRetryableFailure("Interrupted", "pod evicted",
		&pluginsCore.TaskInfo{OccurredAt: &startedAt})

	watcher := &fakeEventWatcher{events: map[watchedObjectKey][]*eventInfo{
		key: {gpuFaultEventFor(79, gpufault.SeverityCritical, faultedAt, faultedAt, testPodUID)},
	}}
	pm := NewPluginManager("test-plugin", nil, nil)
	pm.eventWatcher = watcher

	got := pm.classifyGpuFailure(pod, phase)

	assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
	assert.Equal(t, core.ExecutionError_SYSTEM, got.Err().GetKind())
	require.NotNil(t, got.Err().GetGpuFault())
}

func TestClassifyGpuFailureIdentity(t *testing.T) {
	base := time.Now().Add(-time.Minute)
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}
	phase := pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", nil)

	t.Run("an event without a regarding UID is not trusted", func(t *testing.T) {
		watcher := &fakeEventWatcher{events: map[watchedObjectKey][]*eventInfo{
			key: {gpuFaultEventFor(79, gpufault.SeverityCritical, base, base, "")},
		}}
		pm := NewPluginManager("test-plugin", nil, nil)
		pm.eventWatcher = watcher
		got := pm.classifyGpuFailure(failedPod(), phase)
		assert.Nil(t, got.Err().GetGpuFault())
		assert.Equal(t, "UnknownError", got.Err().GetCode())
	})

	t.Run("a pod whose UID is unknown is matched by name", func(t *testing.T) {
		// The pod was deleted before this round saw it, so the identity object the
		// manager builds carries no UID. The fault recorded against the pod name
		// must still classify the failure.
		watcher := &fakeEventWatcher{events: map[watchedObjectKey][]*eventInfo{
			key: {gpuFaultEvent(79, gpufault.SeverityCritical, base)},
		}}
		pm := NewPluginManager("test-plugin", nil, nil)
		pm.eventWatcher = watcher
		pod := failedPod()
		pod.UID = ""
		got := pm.classifyGpuFailure(pod, phase)
		assert.Equal(t, gpufault.CodeGpuFallenOffBus, got.Err().GetCode())
		assert.NotNil(t, got.Err().GetGpuFault())
	})
}

func TestClassifyGpuFailureSkips(t *testing.T) {
	key := watchedObjectKey{Namespace: "ns", Name: "pod", Kind: "Pod"}
	base := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	events := map[watchedObjectKey][]*eventInfo{key: {gpuFaultEvent(79, gpufault.SeverityCritical, base)}}

	tests := []struct {
		name      string
		resource  client.Object
		phaseInfo pluginsCore.PhaseInfo
		noWatcher bool
	}{
		{
			name:      "the task did not fail",
			resource:  failedPod(),
			phaseInfo: pluginsCore.PhaseInfoRunning(1, nil),
		},
		{
			name:      "the task succeeded",
			resource:  failedPod(),
			phaseInfo: pluginsCore.PhaseInfoSuccess(nil),
		},
		{
			name: "the resource is not a pod",
			resource: &v1.Service{
				TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "pod"},
			},
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", nil),
		},
		{
			name:      "there is no event watcher",
			resource:  failedPod(),
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", nil),
			noWatcher: true,
		},
		{
			name:      "there is no resource",
			resource:  nil,
			phaseInfo: pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPluginManager("test-plugin", nil, nil)
			if !tt.noWatcher {
				pm.eventWatcher = &fakeEventWatcher{events: events}
			}

			got := pm.classifyGpuFailure(tt.resource, tt.phaseInfo)

			assert.Equal(t, tt.phaseInfo.Phase(), got.Phase())
			if tt.phaseInfo.Err() != nil {
				assert.Equal(t, tt.phaseInfo.Err().GetCode(), got.Err().GetCode())
				assert.Nil(t, got.Err().GetGpuFault())
			}
		})
	}
}

// TestAddObjectMetadata_ManagedLabel verifies that the label the manager's Pod cache selects
// on survives addObjectMetadata. NewTaskExecutionMetadata is the single injection point; this
// asserts nothing here drops it, since a Pod without the label is invisible to the executor
// that created it and checkResourcePhase would report it as deleted externally.
func TestAddObjectMetadata_ManagedLabel(t *testing.T) {
	newManager := func(t *testing.T) *PluginManager {
		mockPlugin := k8sMocks.NewPlugin(t)
		mockPlugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})
		return NewPluginManager("test-plugin", mockPlugin, nil)
	}

	// Mirrors what NewTaskExecutionMetadata injects into every task's labels.
	managedLabels := map[string]string{
		flytek8s.ManagedLabelKey: flytek8s.ManagedLabelValue,
	}

	newMeta := func(labels map[string]string) *pluginsCoreMock.TaskExecutionMetadata {
		tID := &pluginsCoreMock.TaskExecutionID{}
		tID.EXPECT().GetGeneratedName().Return("name").Maybe()
		meta := &pluginsCoreMock.TaskExecutionMetadata{}
		meta.EXPECT().GetTaskExecutionID().Return(tID).Maybe()
		meta.EXPECT().GetNamespace().Return("ns")
		meta.EXPECT().GetAnnotations().Return(nil)
		meta.EXPECT().GetLabels().Return(labels)
		meta.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})
		return meta
	}

	t.Run("propagates to the object", func(t *testing.T) {
		pod := &v1.Pod{}
		newManager(t).addObjectMetadata(newMeta(managedLabels), pod, &config.K8sPluginConfig{})

		assert.Equal(t, flytek8s.ManagedLabelValue, pod.GetLabels()[flytek8s.ManagedLabelKey])
	})

	t.Run("matches the selector the manager cache is configured with", func(t *testing.T) {
		pod := &v1.Pod{}
		newManager(t).addObjectMetadata(newMeta(managedLabels), pod, &config.K8sPluginConfig{})

		selector := labels.SelectorFromSet(labels.Set{
			flytek8s.ManagedLabelKey: flytek8s.ManagedLabelValue,
		})
		assert.True(t, selector.Matches(labels.Set(pod.GetLabels())))
	})

	t.Run("not dropped by platform defaults or labels already on the object", func(t *testing.T) {
		pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{flytek8s.ManagedLabelKey: "object-value"},
		}}
		cfg := &config.K8sPluginConfig{
			DefaultLabels: map[string]string{flytek8s.ManagedLabelKey: "default-value"},
		}
		newManager(t).addObjectMetadata(newMeta(managedLabels), pod, cfg)

		assert.Equal(t, flytek8s.ManagedLabelValue, pod.GetLabels()[flytek8s.ManagedLabelKey])
	})
}
