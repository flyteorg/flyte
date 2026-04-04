package plugin

import (
	"testing"

	"github.com/stretchr/testify/require"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

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
