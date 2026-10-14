package controller

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

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
		assert.Equal(t, taskAction.Name+"-", ev.GenerateName)
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

	t.Run("info annotation decodes to the original event", func(t *testing.T) {
		event := testActionEvent(common.ActionPhase_ACTION_PHASE_FAILED, 3)
		event.ErrorInfo = &workflow.ErrorInfo{Kind: workflow.ErrorInfo_KIND_SYSTEM, Code: "Evicted"}

		ev, err := buildActionEventK8s(taskAction, event, "taskaction-controller-host1")
		assert.NoError(t, err)

		decoded := &workflow.ActionEvent{}
		assert.NoError(t, protojson.Unmarshal([]byte(ev.Annotations[annInfo]), decoded))
		assert.True(t, proto.Equal(event, decoded))
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

func TestEventLevelOf(t *testing.T) {
	tests := []struct {
		name      string
		phase     common.ActionPhase
		version   uint32
		prevPhase common.ActionPhase
		expected  EventLevel
	}{
		{"succeeded", common.ActionPhase_ACTION_PHASE_SUCCEEDED, 0, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelTerminal},
		{"failed", common.ActionPhase_ACTION_PHASE_FAILED, 0, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelTerminal},
		{"aborted", common.ActionPhase_ACTION_PHASE_ABORTED, 0, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelTerminal},
		{"timed out", common.ActionPhase_ACTION_PHASE_TIMED_OUT, 0, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelTerminal},
		{"system retry", common.ActionPhase_ACTION_PHASE_QUEUED, systemRetryEventVersionBase, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelTerminal},
		{"just below system retry range", common.ActionPhase_ACTION_PHASE_QUEUED, systemRetryEventVersionBase - 1, common.ActionPhase_ACTION_PHASE_QUEUED, EventLevelDebug},
		{"first event", common.ActionPhase_ACTION_PHASE_QUEUED, 0, common.ActionPhase_ACTION_PHASE_UNSPECIFIED, EventLevelInfo},
		{"phase change", common.ActionPhase_ACTION_PHASE_RUNNING, 0, common.ActionPhase_ACTION_PHASE_QUEUED, EventLevelInfo},
		{"same phase, higher version", common.ActionPhase_ACTION_PHASE_RUNNING, 2, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelDebug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eventLevelOf(testActionEvent(tt.phase, tt.version), tt.prevPhase)
			assert.Equal(t, tt.expected, got)
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
		level         EventLevel
		phase         common.ActionPhase
		prevPhase     common.ActionPhase
		createdReason string
		recorded      string
	}{
		{
			name:          "terminal event is created through the client",
			level:         EventLevelTerminal,
			phase:         common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			prevPhase:     common.ActionPhase_ACTION_PHASE_RUNNING,
			createdReason: string(ReasonActionSucceeded),
		},
		{
			name:      "phase change at info goes through the recorder",
			level:     EventLevelInfo,
			phase:     common.ActionPhase_ACTION_PHASE_RUNNING,
			prevPhase: common.ActionPhase_ACTION_PHASE_QUEUED,
			recorded:  "Normal ActionRunning action a0 attempt 1 running",
		},
		{
			name:      "phase change at terminal is filtered",
			level:     EventLevelTerminal,
			phase:     common.ActionPhase_ACTION_PHASE_RUNNING,
			prevPhase: common.ActionPhase_ACTION_PHASE_QUEUED,
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
				K8sEventLevel:     tt.level,
				reportingInstance: "taskaction-controller-host1",
			}

			reconciler.emitK8sEvent(ctx, testTaskAction(), testActionEvent(tt.phase, 0), tt.prevPhase)

			created := &eventsv1.EventList{}
			require.NoError(t, k8sClient.List(ctx, created))
			if tt.createdReason == "" {
				assert.Empty(t, created.Items)
			} else {
				require.Len(t, created.Items, 1)
				assert.Equal(t, tt.createdReason, created.Items[0].Reason)
				assert.Equal(t, "taskaction-controller-host1", created.Items[0].ReportingInstance)
			}

			select {
			case got := <-recorder.Events:
				assert.Equal(t, tt.recorded, got)
			default:
				assert.Empty(t, tt.recorded)
			}
		})
	}
}
