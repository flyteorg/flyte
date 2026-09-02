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
	executorConfig "github.com/flyteorg/flyte/v2/executor/pkg/config"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	flytesecret "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	stdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
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

func TestNewTaskExecutionMetadata_Interruptible(t *testing.T) {
	marshalTemplate := func(t *testing.T, tmpl *core.TaskTemplate) []byte {
		data, err := proto.Marshal(tmpl)
		require.NoError(t, err)
		return data
	}
	templateWith := func(t *testing.T, interruptible bool) []byte {
		return marshalTemplate(t, &core.TaskTemplate{
			Metadata: &core.TaskMetadata{
				InterruptibleValue: &core.TaskMetadata_Interruptible{Interruptible: interruptible},
			},
		})
	}
	newMeta := func(t *testing.T, runLevel *bool, taskTemplate []byte) pluginsCore.TaskExecutionMetadata {
		meta, err := NewTaskExecutionMetadata(&flyteorgv1.TaskAction{
			Spec: flyteorgv1.TaskActionSpec{
				Project:       "project",
				Domain:        "development",
				RunName:       "run-name",
				ActionName:    "action-name",
				RunOutputBase: "s3://bucket/run",
				Interruptible: runLevel,
				TaskTemplate:  taskTemplate,
			},
		})
		require.NoError(t, err)
		return meta
	}

	t.Run("defaults to false", func(t *testing.T) {
		require.False(t, newMeta(t, nil, nil).IsInterruptible())
	})

	t.Run("template metadata applies when run level is unset", func(t *testing.T) {
		require.True(t, newMeta(t, nil, templateWith(t, true)).IsInterruptible())
		require.False(t, newMeta(t, nil, templateWith(t, false)).IsInterruptible())
	})

	t.Run("template without the flag leaves the default", func(t *testing.T) {
		require.False(t, newMeta(t, nil, marshalTemplate(t, &core.TaskTemplate{Metadata: &core.TaskMetadata{}})).IsInterruptible())
	})

	t.Run("run level wins over template metadata", func(t *testing.T) {
		runTrue, runFalse := true, false
		require.True(t, newMeta(t, &runTrue, templateWith(t, false)).IsInterruptible())
		require.False(t, newMeta(t, &runFalse, templateWith(t, true)).IsInterruptible())
	})
}

func TestAboveInterruptibleFailureThreshold(t *testing.T) {
	tests := []struct {
		name        string
		burnt       uint32
		maxAttempts uint32
		threshold   int32
		want        bool
	}{
		// A complement of one covers the final attempt of the budget and nothing earlier.
		{name: "complement of one, first of three", burnt: 0, maxAttempts: 3, threshold: -1, want: false},
		{name: "complement of one, second of three", burnt: 1, maxAttempts: 3, threshold: -1, want: false},
		{name: "complement of one, last of three", burnt: 2, maxAttempts: 3, threshold: -1, want: true},
		{name: "complement of one, beyond the budget", burnt: 3, maxAttempts: 3, threshold: -1, want: true},

		// A wider complement covers correspondingly more of the tail.
		{name: "complement of two, first of three", burnt: 0, maxAttempts: 3, threshold: -2, want: false},
		{name: "complement of two, second of three", burnt: 1, maxAttempts: 3, threshold: -2, want: true},

		// Zero is rejected at startup; the arithmetic degrades to a complement of zero,
		// covering no attempt in the budget.
		{name: "zero threshold spares the last attempt", burnt: 2, maxAttempts: 3, threshold: 0, want: false},
		{name: "zero threshold trips past the budget", burnt: 3, maxAttempts: 3, threshold: 0, want: true},

		// A complement at or beyond the budget covers every attempt rather than wrapping
		// the subtraction and silently covering none.
		{name: "complement equal to the budget", burnt: 0, maxAttempts: 1, threshold: -1, want: true},
		{name: "complement wider than the budget", burnt: 0, maxAttempts: 2, threshold: -5, want: true},

		// A positive threshold is an absolute count of burnt attempts.
		{name: "positive threshold not yet reached", burnt: 1, maxAttempts: 3, threshold: 2, want: false},
		{name: "positive threshold reached", burnt: 2, maxAttempts: 3, threshold: 2, want: true},
		{name: "positive threshold above the budget never trips", burnt: 2, maxAttempts: 3, threshold: 100, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, aboveInterruptibleFailureThreshold(tt.burnt, tt.maxAttempts, tt.threshold))
		})
	}
}

func TestInterruptibleThresholdReached(t *testing.T) {
	// A task with two retries and the default system-failure budget.
	const maxAttempts, maxSystemAttempts = 3, 4

	tests := []struct {
		name           string
		retryAttempt   uint32
		systemFailures uint32
		threshold      int32
		want           bool
	}{
		{name: "first attempt, nothing burnt", retryAttempt: 0, systemFailures: 0, threshold: -1, want: false},
		{name: "middle retry", retryAttempt: 1, systemFailures: 0, threshold: -1, want: false},
		{name: "final retry", retryAttempt: 2, systemFailures: 0, threshold: -1, want: true},

		// A reclaimed node charged to the system budget trips the fallback on its own,
		// without the task having burnt any of its own retries.
		{name: "final system attempt", retryAttempt: 0, systemFailures: 3, threshold: -1, want: true},
		{name: "system budget not yet exhausted", retryAttempt: 0, systemFailures: 2, threshold: -1, want: false},

		// Neither budget being exhausted leaves the task on interruptible capacity.
		{name: "both budgets partly burnt", retryAttempt: 1, systemFailures: 2, threshold: -1, want: false},

		// A threshold above both budgets disables the fallback entirely.
		{name: "threshold above both budgets", retryAttempt: 2, systemFailures: 3, threshold: 100, want: false},

		// Zero is rejected at startup; the arithmetic degrades to covering no attempt.
		{name: "zero threshold spares the final retry", retryAttempt: 2, systemFailures: 3, threshold: 0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, interruptibleThresholdReached(
				tt.retryAttempt, tt.systemFailures, maxAttempts, maxSystemAttempts, tt.threshold))
		})
	}
}

func TestMaxSystemAttempts(t *testing.T) {
	// A configured maximum is a retry count, so the attempt budget is one larger.
	require.Equal(t, uint32(8), maxSystemAttempts(7))

	// An unset or nonsensical maximum falls back to the reconciler's default of 3.
	require.Equal(t, uint32(4), maxSystemAttempts(0))
	require.Equal(t, uint32(4), maxSystemAttempts(-1))
}

func TestNewTaskExecutionMetadata_InterruptibleFailureThreshold(t *testing.T) {
	// withThreshold swaps in an executor config for the duration of the subtest.
	withThreshold := func(t *testing.T, threshold int32) {
		section := stdconfig.GetSection("executor")
		previous := section.GetConfig().(*executorConfig.Config)
		swapped := *previous
		swapped.InterruptibleFailureThreshold = threshold
		require.NoError(t, section.SetConfig(&swapped))
		t.Cleanup(func() { require.NoError(t, section.SetConfig(previous)) })
	}
	templateWithRetries := func(t *testing.T, retries uint32) []byte {
		data, err := proto.Marshal(&core.TaskTemplate{
			Metadata: &core.TaskMetadata{
				InterruptibleValue: &core.TaskMetadata_Interruptible{Interruptible: true},
				Retries:            &core.RetryStrategy{Retries: retries},
			},
		})
		require.NoError(t, err)
		return data
	}
	newMeta := func(t *testing.T, taskTemplate []byte, attempts, systemFailures uint32) pluginsCore.TaskExecutionMetadata {
		meta, err := NewTaskExecutionMetadata(&flyteorgv1.TaskAction{
			Spec: flyteorgv1.TaskActionSpec{
				Project:       "project",
				Domain:        "development",
				RunName:       "run-name",
				ActionName:    "action-name",
				RunOutputBase: "s3://bucket/run",
				TaskTemplate:  taskTemplate,
			},
			Status: flyteorgv1.TaskActionStatus{
				Attempts:       attempts,
				SystemFailures: systemFailures,
			},
		})
		require.NoError(t, err)
		return meta
	}

	t.Run("last of the task's own attempts falls back", func(t *testing.T) {
		withThreshold(t, -1)
		template := templateWithRetries(t, 2)
		require.True(t, newMeta(t, template, 1, 0).IsInterruptible())
		require.True(t, newMeta(t, template, 2, 0).IsInterruptible())
		require.False(t, newMeta(t, template, 3, 0).IsInterruptible())
	})

	t.Run("last system attempt falls back", func(t *testing.T) {
		withThreshold(t, -1)
		// The default system-failure budget allows four attempts, so the fourth is the last.
		template := templateWithRetries(t, 2)
		require.True(t, newMeta(t, template, 1, 2).IsInterruptible())
		require.False(t, newMeta(t, template, 1, 3).IsInterruptible())
	})

	t.Run("a task without retries keeps its single interruptible attempt", func(t *testing.T) {
		withThreshold(t, -1)
		require.True(t, newMeta(t, templateWithRetries(t, 0), 1, 0).IsInterruptible())
	})

	t.Run("a threshold above both budgets disables the fallback", func(t *testing.T) {
		withThreshold(t, 100)
		template := templateWithRetries(t, 2)
		require.True(t, newMeta(t, template, 3, 0).IsInterruptible())
		require.True(t, newMeta(t, template, 1, 3).IsInterruptible())
	})

	t.Run("a non-interruptible task is unaffected", func(t *testing.T) {
		withThreshold(t, -1)
		data, err := proto.Marshal(&core.TaskTemplate{
			Metadata: &core.TaskMetadata{Retries: &core.RetryStrategy{Retries: 2}},
		})
		require.NoError(t, err)
		require.False(t, newMeta(t, data, 1, 0).IsInterruptible())
		require.False(t, newMeta(t, data, 3, 0).IsInterruptible())
	})

	t.Run("the configured threshold is reported to plugins", func(t *testing.T) {
		withThreshold(t, -2)
		require.Equal(t, int32(-2), newMeta(t, nil, 1, 0).GetInterruptibleFailureThreshold())
	})
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
