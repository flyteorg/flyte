package controller

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

func TestTruncateUTF8(t *testing.T) {
	longMessage := strings.Repeat("task failed ", 100)

	tests := []struct {
		name     string
		s        string
		max      int
		expected string
	}{
		{name: "empty input", s: "", max: 5, expected: ""},
		{name: "shorter than max", s: "task failed", max: 20, expected: "task failed"},
		{name: "exactly max", s: "task failed", max: 11, expected: "task failed"},
		{name: "word over max", s: "failed", max: 4, expected: "fail"},
		{name: "sentence over max", s: "the task failed after three retries", max: 15, expected: "the task failed"},
		{name: "zero max", s: "task failed", max: 0, expected: ""},
		{name: "message over note limit", s: longMessage, max: 1024, expected: longMessage[:1024]},
		// "task " is 5 bytes and each euro sign is 3, so byte 9 is inside the second euro sign.
		{name: "cut inside a multi-byte character", s: "task \u20ac\u20ac!", max: 9, expected: "task \u20ac"},
		{name: "invalid byte inside the limit", s: "ab\xffcd", max: 10, expected: "abcd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateUTF8(tt.s, tt.max)
			assert.Equal(t, tt.expected, got)
			assert.LessOrEqual(t, len(got), tt.max)
			assert.True(t, utf8.ValidString(got))
		})
	}
}

// testTaskAction returns the TaskAction that the k8s event of an action event refers to.
func testTaskAction() *flyteorgv1.TaskAction {
	return &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "run1-a0",
			Namespace: "flytesnacks-development",
			UID:       types.UID("uid-1"),
		},
	}
}

func TestBuildActionEventK8s(t *testing.T) {
	taskAction := testTaskAction()

	t.Run("fields the apiserver checks", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_SUCCEEDED, 0)

		ev, err := buildActionEventK8s(taskAction, event, "taskaction-controller-host1")
		assert.NoError(t, err)

		assert.Equal(t, taskAction.Namespace, ev.Namespace)
		assert.Equal(t, ev.Namespace, ev.Regarding.Namespace)
		assert.Equal(t, "uid-1.1.succeeded.0", ev.Name)
		assert.Empty(t, ev.GenerateName, "a generated name would let a retried reconcile create a duplicate")
		assert.Equal(t, flyteorgv1.GroupVersion.String(), ev.Regarding.APIVersion)
		assert.Equal(t, "TaskAction", ev.Regarding.Kind)
		assert.Equal(t, taskAction.Name, ev.Regarding.Name)
		assert.Equal(t, taskAction.UID, ev.Regarding.UID)

		assert.NotEmpty(t, ev.Type)
		assert.NotEmpty(t, ev.Reason)
		assert.NotEmpty(t, ev.Action)
		assert.NotEmpty(t, ev.ReportingController)
		assert.Equal(t, "taskaction-controller-host1", ev.ReportingInstance)
	})

	t.Run("no error annotations without error info", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_SUCCEEDED, 0)

		ev, err := buildActionEventK8s(taskAction, event, "taskaction-controller-host1")
		assert.NoError(t, err)

		_, ok := ev.Annotations[annErrorKind]
		assert.False(t, ok)
		_, ok = ev.Annotations[annErrorCode]
		assert.False(t, ok)
	})

	t.Run("error annotations with error info", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_FAILED, 0)
		event.ErrorInfo = &workflow.ErrorInfo{
			Kind:    workflow.ErrorInfo_KIND_USER,
			Code:    "OOMKilled",
			Message: "container exceeded its memory limit",
		}

		ev, err := buildActionEventK8s(taskAction, event, "taskaction-controller-host1")
		assert.NoError(t, err)

		assert.Equal(t, "KIND_USER", ev.Annotations[annErrorKind])
		assert.Equal(t, "OOMKilled", ev.Annotations[annErrorCode])
	})

	t.Run("control plane context lands on the event", func(t *testing.T) {
		child := testTaskAction()
		parent := "a0"
		child.Spec.ParentActionName = &parent
		child.Spec.TaskType = "container"
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_SUCCEEDED, 0)

		ev, err := buildActionEventK8s(child, event, "taskaction-controller-host1")
		assert.NoError(t, err)

		assert.Equal(t, "a0", ev.Annotations[annParentActionName])
		assert.Equal(t, "container", ev.Annotations[annTaskType])
	})

	t.Run("no payload annotation, the note carries the error", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_FAILED, 3)
		event.ErrorInfo = &workflow.ErrorInfo{Kind: workflow.ErrorInfo_KIND_SYSTEM, Code: "Evicted", Message: "node pressure"}

		ev, err := buildActionEventK8s(taskAction, event, "taskaction-controller-host1")
		assert.NoError(t, err)

		assert.NotContains(t, ev.Annotations, annPrefix+"info")
		assert.Contains(t, ev.Note, "node pressure")
	})
}

func TestControlPlaneContext(t *testing.T) {
	t.Run("root action without labels sets nothing", func(t *testing.T) {
		taskAction := testTaskAction()
		taskAction.Labels = map[string]string{"flyte.org/run": "run1", "flyte.org/is-root": "true"}

		assert.Empty(t, controlPlaneContext(taskAction))
	})

	t.Run("child action sets parent, task type and group", func(t *testing.T) {
		taskAction := testTaskAction()
		parent := "a0"
		taskAction.Spec.ParentActionName = &parent
		taskAction.Spec.TaskType = "spark"
		taskAction.Spec.Group = "map-1"

		got := controlPlaneContext(taskAction)

		assert.Equal(t, map[string]string{
			annParentActionName: "a0",
			annTaskType:         "spark",
			annGroup:            "map-1",
		}, got)
	})

	t.Run("empty parent name counts as root", func(t *testing.T) {
		taskAction := testTaskAction()
		empty := ""
		taskAction.Spec.ParentActionName = &empty

		_, ok := controlPlaneContext(taskAction)[annParentActionName]
		assert.False(t, ok)
	})

	t.Run("user labels become one JSON annotation without flyte.org keys", func(t *testing.T) {
		taskAction := testTaskAction()
		taskAction.Labels = map[string]string{
			"flyte.org/run":     "run1",
			"flyte.org/is-root": "true",
			"team":              "ml",
			"cost-center":       "123",
		}

		got := controlPlaneContext(taskAction)

		decoded := map[string]string{}
		require.NoError(t, json.Unmarshal([]byte(got[annLabels]), &decoded))
		assert.Equal(t, map[string]string{"team": "ml", "cost-center": "123"}, decoded)
	})
}

// testActionEvent returns an event for action a0 of run run1, attempt 1.
func testActionEvent(phase common.ActionPhase, version uint32) *workflow.ActionEvent {
	return &workflow.ActionEvent{
		Id: &common.ActionIdentifier{
			Run:  &common.RunIdentifier{Project: "flytesnacks", Domain: "development", Name: "run1"},
			Name: "a0",
		},
		Attempt: 1,
		Phase:   phase,
		Version: version,
		Cluster: "cluster1",
	}
}

func TestShouldEmitK8sEvent(t *testing.T) {
	const (
		queued      = common.ActionPhase_ACTION_PHASE_QUEUED
		running     = common.ActionPhase_ACTION_PHASE_RUNNING
		succeeded   = common.ActionPhase_ACTION_PHASE_SUCCEEDED
		failed      = common.ActionPhase_ACTION_PHASE_FAILED
		aborted     = common.ActionPhase_ACTION_PHASE_ABORTED
		timedOut    = common.ActionPhase_ACTION_PHASE_TIMED_OUT
		unspecified = common.ActionPhase_ACTION_PHASE_UNSPECIFIED
	)
	tests := []struct {
		name      string
		mode      K8sEventMode
		phase     common.ActionPhase
		version   uint32
		prevPhase common.ActionPhase
		want      bool
	}{
		// off emits nothing, not even terminal phases.
		{"off: succeeded", K8sEventOff, succeeded, 0, running, false},
		{"off: phase change", K8sEventOff, running, 0, queued, false},

		// terminal emits terminal phases and system retries only.
		{"terminal: succeeded", K8sEventTerminal, succeeded, 0, running, true},
		{"terminal: failed", K8sEventTerminal, failed, 0, running, true},
		{"terminal: aborted", K8sEventTerminal, aborted, 0, running, true},
		{"terminal: timed out", K8sEventTerminal, timedOut, 0, running, true},
		{"terminal: system retry", K8sEventTerminal, queued, systemRetryEventVersionBase, running, true},
		{"terminal: just below system retry range", K8sEventTerminal, queued, systemRetryEventVersionBase - 1, queued, false},
		{"terminal: first event", K8sEventTerminal, queued, 0, unspecified, false},
		{"terminal: phase change", K8sEventTerminal, running, 0, queued, false},

		// all emits every phase change, but never a version bump within the same phase.
		{"all: succeeded", K8sEventAll, succeeded, 0, running, true},
		{"all: system retry", K8sEventAll, queued, systemRetryEventVersionBase, running, true},
		{"all: first event", K8sEventAll, queued, 0, unspecified, true},
		{"all: phase change", K8sEventAll, running, 0, queued, true},
		{"all: same phase, higher version", K8sEventAll, running, 2, running, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldEmitK8sEvent(tt.mode, testActionEvent(tt.phase, tt.version), tt.prevPhase)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEventType(t *testing.T) {
	tests := []struct {
		name     string
		phase    common.ActionPhase
		version  uint32
		expected string
	}{
		{"running", common.ActionPhase_ACTION_PHASE_RUNNING, 0, corev1.EventTypeNormal},
		{"succeeded", common.ActionPhase_ACTION_PHASE_SUCCEEDED, 0, corev1.EventTypeNormal},
		{"aborted is requested, not a fault", common.ActionPhase_ACTION_PHASE_ABORTED, 0, corev1.EventTypeNormal},
		{"failed", common.ActionPhase_ACTION_PHASE_FAILED, 0, corev1.EventTypeWarning},
		{"timed out", common.ActionPhase_ACTION_PHASE_TIMED_OUT, 0, corev1.EventTypeWarning},
		{"system retry", common.ActionPhase_ACTION_PHASE_QUEUED, systemRetryEventVersionBase, corev1.EventTypeWarning},
		{"queued", common.ActionPhase_ACTION_PHASE_QUEUED, 0, corev1.EventTypeNormal}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, eventType(testActionEvent(tt.phase, tt.version)))
		})
	}
}

func TestEventReason(t *testing.T) {
	t.Run("system retry is checked before the phase", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_QUEUED, systemRetryEventVersionBase)
		assert.Equal(t, string(ReasonActionSystemRetry), eventReason(event))
	})

	t.Run("failed", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_FAILED, 0)
		assert.Equal(t, string(ReasonActionFailed), eventReason(event))
	})

	t.Run("every phase has a valid reason", func(t *testing.T) {
		for value, name := range common.ActionPhase_name {
			event := testActionEvent(common.ActionPhase(value), 0)
			reason := eventReason(event)
			assert.NotEmpty(t, reason, name)
			assert.LessOrEqual(t, len(reason), 128, name)
		}
	})
}

func TestHumanSummary(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_TIMED_OUT, 0)
		assert.Equal(t, "action a0 attempt 1 timed out", humanSummary(event))
	})

	t.Run("system retry", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_QUEUED, systemRetryEventVersionBase)
		assert.Equal(t, "action a0 attempt 1 restarted after a system failure", humanSummary(event))
	})

	t.Run("with error", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_FAILED, 0)
		event.ErrorInfo = &workflow.ErrorInfo{
			Kind:    workflow.ErrorInfo_KIND_USER,
			Code:    "OOMKilled",
			Message: "container exceeded its memory limit",
		}
		assert.Equal(t,
			`action a0 attempt 1 failed: USER error "OOMKilled": container exceeded its memory limit`,
			humanSummary(event))
	})

	t.Run("long message fits in the note", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_FAILED, 0)
		event.ErrorInfo = &workflow.ErrorInfo{Message: strings.Repeat("stack frame ", 200)}
		assert.LessOrEqual(t, len(humanSummary(event)), noteLimit)
	})
}

func TestEmitK8sEvent(t *testing.T) {
	tests := []struct {
		name          string
		mode          K8sEventMode
		phase         common.ActionPhase
		prevPhase     common.ActionPhase
		createdReason string
	}{
		{
			name:          "terminal event is created through the client",
			mode:          K8sEventTerminal,
			phase:         common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			prevPhase:     common.ActionPhase_ACTION_PHASE_RUNNING,
			createdReason: string(ReasonActionSucceeded),
		},
		{
			name:          "phase change at all is created through the client",
			mode:          K8sEventAll,
			phase:         common.ActionPhase_ACTION_PHASE_RUNNING,
			prevPhase:     common.ActionPhase_ACTION_PHASE_QUEUED,
			createdReason: string(ReasonActionRunning),
		},
		{
			name:      "version bump at all is filtered",
			mode:      K8sEventAll,
			phase:     common.ActionPhase_ACTION_PHASE_RUNNING,
			prevPhase: common.ActionPhase_ACTION_PHASE_RUNNING,
		},
		{
			name:      "phase change at terminal is filtered",
			mode:      K8sEventTerminal,
			phase:     common.ActionPhase_ACTION_PHASE_RUNNING,
			prevPhase: common.ActionPhase_ACTION_PHASE_QUEUED,
		},
		{
			name:      "terminal event at off is filtered",
			mode:      K8sEventOff,
			phase:     common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			prevPhase: common.ActionPhase_ACTION_PHASE_RUNNING,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			k8sClient := fake.NewClientBuilder().Build()
			recorder := events.NewFakeRecorder(1)
			reconciler := &TaskActionReconciler{
				Client:            k8sClient,
				Recorder:          recorder,
				K8sEventMode:      tt.mode,
				reportingInstance: "taskaction-controller-host1",
			}

			require.NoError(t, reconciler.emitK8sEvent(ctx, testTaskAction(), testActionEvent(tt.phase, 0), tt.prevPhase))

			created := &eventsv1.EventList{}
			require.NoError(t, k8sClient.List(ctx, created))
			if tt.createdReason == "" {
				assert.Empty(t, created.Items)
			} else {
				require.Len(t, created.Items, 1)
				assert.Equal(t, tt.createdReason, created.Items[0].Reason)
				assert.Equal(t, "taskaction-controller-host1", created.Items[0].ReportingInstance)
			}

			// Nothing goes through the recorder any more: every emitted event is created directly.
			select {
			case got := <-recorder.Events:
				t.Fatalf("unexpected recorder event %q", got)
			default:
			}
		})
	}
}

func TestK8sEventName(t *testing.T) {
	const (
		queued    = common.ActionPhase_ACTION_PHASE_QUEUED
		waiting   = common.ActionPhase_ACTION_PHASE_WAITING_FOR_RESOURCES
		running   = common.ActionPhase_ACTION_PHASE_RUNNING
		succeeded = common.ActionPhase_ACTION_PHASE_SUCCEEDED
		timedOut  = common.ActionPhase_ACTION_PHASE_TIMED_OUT
	)
	taskAction := testTaskAction()

	tests := []struct {
		name     string
		phase    common.ActionPhase
		attempt  uint32
		version  uint32
		expected string
	}{
		{name: "terminal phase", phase: succeeded, attempt: 1, expected: "uid-1.1.succeeded.0"},
		{name: "later attempt and version", phase: running, attempt: 2, version: 3, expected: "uid-1.2.running.3"},
		{name: "underscore becomes a dash", phase: waiting, attempt: 1, expected: "uid-1.1.waiting-for-resources.0"},
		{name: "two underscores", phase: timedOut, attempt: 1, expected: "uid-1.1.timed-out.0"},
		{
			name: "system retry version", phase: queued, attempt: 1, version: systemRetryEventVersionBase,
			expected: "uid-1.1.queued.1073741824",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := testActionEvent(tt.phase, tt.version)
			event.Attempt = tt.attempt
			assert.Equal(t, tt.expected, k8sEventName(taskAction, event))
		})
	}

	t.Run("every phase gives a valid and distinct name", func(t *testing.T) {
		// A real UID and the largest attempt and version give the longest name.
		longest := testTaskAction()
		longest.UID = types.UID("4d0e7a19-2c4b-4f0e-9a51-7f3c2b1d6e80")
		seen := map[string]common.ActionPhase{}
		for value := range common.ActionPhase_name {
			phase := common.ActionPhase(value)
			event := testActionEvent(phase, math.MaxUint32)
			event.Attempt = math.MaxUint32

			name := k8sEventName(longest, event)
			assert.Empty(t, validation.IsDNS1123Subdomain(name), "phase %s gives invalid name %q", phase, name)
			prev, dup := seen[name]
			assert.False(t, dup, "phases %s and %s give the same name %q", prev, phase, name)
			seen[name] = phase
		}
	})

	t.Run("same action event gives the same name", func(t *testing.T) {
		a := testActionEvent(common.ActionPhase_ACTION_PHASE_FAILED, 2)
		b := testActionEvent(common.ActionPhase_ACTION_PHASE_FAILED, 2)
		b.ErrorInfo = &workflow.ErrorInfo{Code: "OOMKilled"} // not part of the key
		assert.Equal(t, k8sEventName(taskAction, a), k8sEventName(taskAction, b))
	})

	t.Run("recreated TaskAction gets new names", func(t *testing.T) {
		recreated := testTaskAction()
		recreated.UID = types.UID("uid-2")
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_SUCCEEDED, 0)
		assert.NotEqual(t, k8sEventName(taskAction, event), k8sEventName(recreated, event))
	})
}

func TestEmitK8sEventTwiceCreatesOne(t *testing.T) {
	ctx := context.Background()
	k8sClient := fake.NewClientBuilder().Build()
	reconciler := &TaskActionReconciler{
		Client:            k8sClient,
		K8sEventMode:      K8sEventAll,
		reportingInstance: "taskaction-controller-host1",
	}
	taskAction := testTaskAction()
	queued := common.ActionPhase_ACTION_PHASE_QUEUED
	running := common.ActionPhase_ACTION_PHASE_RUNNING
	succeeded := common.ActionPhase_ACTION_PHASE_SUCCEEDED

	// A retried reconcile emits the same action event again; the second create
	// gets AlreadyExists and is skipped.
	require.NoError(t, reconciler.emitK8sEvent(ctx, taskAction, testActionEvent(running, 0), queued))
	require.NoError(t, reconciler.emitK8sEvent(ctx, taskAction, testActionEvent(running, 0), queued))

	created := &eventsv1.EventList{}
	require.NoError(t, k8sClient.List(ctx, created))
	require.Len(t, created.Items, 1)
	assert.Equal(t, "uid-1.1.running.0", created.Items[0].Name)

	// A different action event of the same action still gets its own k8s Event.
	require.NoError(t, reconciler.emitK8sEvent(ctx, taskAction, testActionEvent(succeeded, 0), running))
	require.NoError(t, k8sClient.List(ctx, created))
	assert.Len(t, created.Items, 2)
}

func TestRecordEventEmitsK8sEvent(t *testing.T) {
	ctx := context.Background()
	taskAction := testTaskAction()
	running := common.ActionPhase_ACTION_PHASE_RUNNING
	queued := common.ActionPhase_ACTION_PHASE_QUEUED

	t.Run("emits once the events service stored the event", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().Build()
		reconciler := &TaskActionReconciler{
			Client:       k8sClient,
			eventsClient: &fakeEventsClient{},
			K8sEventMode: K8sEventAll,
		}

		require.NoError(t, reconciler.recordEvent(ctx, taskAction, testActionEvent(running, 0), queued))

		created := &eventsv1.EventList{}
		require.NoError(t, k8sClient.List(ctx, created))
		require.Len(t, created.Items, 1, "recordEvent creates the k8s Event itself, before any status write")
		assert.Equal(t, "uid-1.1.running.0", created.Items[0].Name)
	})

	t.Run("no k8s Event when the events service fails", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().Build()
		reconciler := &TaskActionReconciler{
			Client:       k8sClient,
			eventsClient: &failingEventsClient{failures: 1},
			K8sEventMode: K8sEventAll,
		}

		require.Error(t, reconciler.recordEvent(ctx, taskAction, testActionEvent(running, 0), queued))

		created := &eventsv1.EventList{}
		require.NoError(t, k8sClient.List(ctx, created))
		assert.Empty(t, created.Items, "the DB write comes first; a k8s Event must not exist without its row")
	})

	t.Run("returns a create failure that a retry can fix", func(t *testing.T) {
		k8sClient := failEventCreate(1, apierrors.NewServiceUnavailable("apiserver unavailable"))
		reconciler := &TaskActionReconciler{
			Client:       k8sClient,
			eventsClient: &fakeEventsClient{},
			K8sEventMode: K8sEventAll,
		}

		err := reconciler.recordEvent(ctx, taskAction, testActionEvent(running, 0), queued)
		require.Error(t, err, "the reconcile must requeue before the status write")
		assert.True(t, apierrors.IsServiceUnavailable(err), "the API error is wrapped, not replaced")

		// The requeued reconcile records the event again and creates the k8s Event.
		require.NoError(t, reconciler.recordEvent(ctx, taskAction, testActionEvent(running, 0), queued))
		created := &eventsv1.EventList{}
		require.NoError(t, k8sClient.List(ctx, created))
		require.Len(t, created.Items, 1)
		assert.Equal(t, "uid-1.1.running.0", created.Items[0].Name)
	})

	t.Run("drops a create failure that a retry cannot fix", func(t *testing.T) {
		eventsGR := schema.GroupResource{Group: "events.k8s.io", Resource: "events"}
		eventGK := schema.GroupKind{Group: "events.k8s.io", Kind: "Event"}
		eventName := "uid-1.1.running.0"
		permanent := map[string]error{
			// Forbidden is also what a create in a terminating namespace returns.
			"forbidden":   apierrors.NewForbidden(eventsGR, eventName, errors.New("no RBAC")),
			"invalid":     apierrors.NewInvalid(eventGK, eventName, nil),
			"bad request": apierrors.NewBadRequest("bad Event"),
		}
		for name, createErr := range permanent {
			t.Run(name, func(t *testing.T) {
				k8sClient := failEventCreate(1, createErr)
				reconciler := &TaskActionReconciler{
					Client:       k8sClient,
					eventsClient: &fakeEventsClient{},
					K8sEventMode: K8sEventAll,
				}

				require.NoError(t, reconciler.recordEvent(ctx, taskAction, testActionEvent(running, 0), queued),
					"a retry cannot fix it, so the action must not block on it")

				created := &eventsv1.EventList{}
				require.NoError(t, k8sClient.List(ctx, created))
				assert.Empty(t, created.Items, "the k8s Event is dropped, not created")
			})
		}
	})
}

// failEventCreate returns a fake client whose first n creates of an events.k8s.io
// Event fail with createErr.
func failEventCreate(n int, createErr error) client.WithWatch {
	failures := n
	return interceptor.NewClient(fake.NewClientBuilder().Build(), interceptor.Funcs{
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			if _, ok := obj.(*eventsv1.Event); ok && failures > 0 {
				failures--
				return createErr
			}
			return c.Create(ctx, obj, opts...)
		},
	})
}
