package plugin

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
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
