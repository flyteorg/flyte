package plugin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	flytesecret "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestResolveServiceAccount(t *testing.T) {
	// A service account set on the task's security context always wins over the default.
	sc := &core.SecurityContext{RunAs: &core.Identity{K8SServiceAccount: "custom-sa"}}
	require.Equal(t, "custom-sa", resolveServiceAccount(sc, "config-default-sa"))

	// With no run account, it falls back to the provided default. Getters are nil-safe.
	require.Equal(t, "config-default-sa", resolveServiceAccount(nil, "config-default-sa"))
	require.Equal(t, "config-default-sa", resolveServiceAccount(&core.SecurityContext{}, "config-default-sa"))

	// An empty default resolves to empty, so Kubernetes uses the pod namespace's `default`.
	require.Equal(t, "", resolveServiceAccount(nil, ""))
}

func TestNewTaskExecutionMetadata_UsesProjectedRunContext(t *testing.T) {
	interruptible := true
	taskAction := &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			Project:       "project",
			Domain:        "development",
			RunName:       "run-name",
			ActionName:    "action-name",
			RunOutputBase: "s3://bucket/run",
			EnvVars:       map[string]string{"TRACE_ID": "root-abc"},
			Interruptible: &interruptible,
		},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)
	require.Equal(t, "root-abc", meta.GetEnvironmentVariables()["TRACE_ID"])
	require.True(t, meta.IsInterruptible())
}

func TestNewTaskExecutionMetadata_UserEnvVarsCannotClobberInternal(t *testing.T) {
	taskAction := &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			Project:       "project",
			Domain:        "development",
			RunName:       "run-name",
			ActionName:    "action-name",
			RunOutputBase: "s3://bucket/run",
			EnvVars: map[string]string{
				"ACTION_NAME": "malicious-override",
				"RUN_NAME":    "malicious-override",
				"_U_RUN_BASE": "malicious-override",
				"USER_VAR":    "allowed",
			},
		},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)

	env := meta.GetEnvironmentVariables()
	require.Equal(t, "action-name", env["ACTION_NAME"])
	require.Equal(t, "run-name", env["RUN_NAME"])
	require.Equal(t, "s3://bucket/run", env["_U_RUN_BASE"])
	require.Equal(t, "allowed", env["USER_VAR"])
}

func TestNewTaskExecutionMetadata_UsesTaskTemplateID(t *testing.T) {
	taskID := &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "development",
		Name:         "task-name",
		Version:      "version",
	}
	taskTemplate, err := proto.Marshal(&core.TaskTemplate{Id: taskID})
	require.NoError(t, err)

	taskAction := &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			Project:       "project",
			Domain:        "development",
			RunName:       "run-name",
			ActionName:    "action-name",
			RunOutputBase: "s3://bucket/run",
			TaskTemplate:  taskTemplate,
		},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)

	got := meta.GetTaskExecutionID().GetID().GetTaskId()
	require.NotNil(t, got)
	require.Equal(t, taskID.GetResourceType(), got.GetResourceType())
	require.Equal(t, taskID.GetProject(), got.GetProject())
	require.Equal(t, taskID.GetDomain(), got.GetDomain())
	require.Equal(t, taskID.GetName(), got.GetName())
	require.Equal(t, taskID.GetVersion(), got.GetVersion())
}

func TestTaskExecutionID_GetGeneratedNameWith(t *testing.T) {
	t.Run("within max length", func(t *testing.T) {
		execID := &taskExecutionID{
			generatedName: "short-name",
		}
		name, err := execID.GetGeneratedNameWith(0, 50)
		require.NoError(t, err)
		require.Equal(t, "short-name", name)
	})

	t.Run("exceeds max length uses FixedLengthUniqueID", func(t *testing.T) {
		execID := &taskExecutionID{
			generatedName: "a-very-long-generated-name-that-exceeds-the-max-length-limit-12345",
		}
		maxLength := 20
		name, err := execID.GetGeneratedNameWith(0, maxLength)
		require.NoError(t, err)
		require.LessOrEqual(t, len(name), maxLength)

		expected, err := encoding.FixedLengthUniqueID(execID.generatedName, maxLength)
		require.NoError(t, err)
		require.Equal(t, expected, name)
	})

	t.Run("max length too small returns error", func(t *testing.T) {
		execID := &taskExecutionID{
			generatedName: "long-name",
		}
		_, err := execID.GetGeneratedNameWith(0, 2)
		require.Error(t, err)
	})
}

func TestNewTaskExecutionMetadata_ManagedLabel(t *testing.T) {
	// The manager cache selects on this label, so a Pod without it is invisible
	// to the executor. Plugins merge these labels into the Pod templates they
	// build, which is how operator-created Pods inherit it.
	newMeta := func(labels map[string]string) pluginsCore.TaskExecutionMetadata {
		meta, err := NewTaskExecutionMetadata(&flyteorgv1.TaskAction{
			ObjectMeta: metav1.ObjectMeta{Labels: labels},
			Spec: flyteorgv1.TaskActionSpec{
				Project:       "project",
				Domain:        "development",
				RunName:       "run-name",
				ActionName:    "action-name",
				RunOutputBase: "s3://bucket/run",
			},
		})
		require.NoError(t, err)
		return meta
	}

	t.Run("always set", func(t *testing.T) {
		require.Equal(t, flytek8s.ManagedLabelValue,
			newMeta(nil).GetLabels()[flytek8s.ManagedLabelKey])
	})

	t.Run("matches the cache selector", func(t *testing.T) {
		selector := labels.SelectorFromSet(labels.Set{
			flytek8s.ManagedLabelKey: flytek8s.ManagedLabelValue,
		})
		require.True(t, selector.Matches(labels.Set(newMeta(nil).GetLabels())))
	})

	t.Run("user labels cannot override it", func(t *testing.T) {
		meta := newMeta(map[string]string{flytek8s.ManagedLabelKey: "false"})
		require.Equal(t, flytek8s.ManagedLabelValue,
			meta.GetLabels()[flytek8s.ManagedLabelKey])
	})
}

func TestNewTaskExecutionMetadata_StampsIdentifyingLabels(t *testing.T) {
	taskTemplate, err := proto.Marshal(&core.TaskTemplate{
		Id: &core.Identifier{Name: "my_module.my_task"},
	})
	require.NoError(t, err)

	taskAction := &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			Project:       "project",
			Domain:        "development",
			RunName:       "run-name",
			ActionName:    "action-name",
			RunOutputBase: "s3://bucket/run",
			TaskTemplate:  taskTemplate,
		},
		Status: flyteorgv1.TaskActionStatus{Attempts: 3},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)

	labels := meta.GetLabels()
	require.Equal(t, "run-name", labels[RunLabel])
	require.Equal(t, "action-name", labels[ActionLabel])
	// Attempts are 1-based, and the third attempt is the second retry.
	require.Equal(t, "3", labels[AttemptLabel])
	require.Equal(t, "my_module.my_task", labels[TaskNameLabel])

	// The labels injected for secret scoping are still there.
	require.Equal(t, "project", labels[flytesecret.ProjectLabel])
	require.Equal(t, "development", labels[flytesecret.DomainLabel])
}

func TestNewTaskExecutionMetadata_LabelsAreSanitizedAndOptional(t *testing.T) {
	longName := strings.Repeat("a", 70)
	taskTemplate, err := proto.Marshal(&core.TaskTemplate{
		Id: &core.Identifier{Name: longName},
	})
	require.NoError(t, err)

	taskAction := &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			Project:      "project",
			Domain:       "development",
			RunName:      "run/name:with bad chars",
			TaskTemplate: taskTemplate,
		},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)

	labels := meta.GetLabels()
	require.Equal(t, "run-name-with-bad-chars", labels[RunLabel])
	require.Len(t, labels[TaskNameLabel], 63)
	// A first attempt that has not been recorded yet still reads as attempt 1.
	require.Equal(t, "1", labels[AttemptLabel])
	// An empty action name is left unstamped rather than stamped blank.
	require.NotContains(t, labels, ActionLabel)

	for key, value := range labels {
		require.Empty(t, validation.IsValidLabelValue(value), "label %s has invalid value %q", key, value)
	}
}

func TestNewTaskExecutionMetadata_CRDLabelsDoNotClobberIdentifyingLabels(t *testing.T) {
	taskAction := &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{RunLabel: "spoofed", "team": "data"},
		},
		Spec: flyteorgv1.TaskActionSpec{
			Project:    "project",
			Domain:     "development",
			RunName:    "run-name",
			ActionName: "action-name",
		},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)

	labels := meta.GetLabels()
	require.Equal(t, "run-name", labels[RunLabel])
	require.Equal(t, "data", labels["team"])
}
