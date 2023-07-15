package k8s

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"
	podPlugin "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/pod"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/stretchr/testify/assert"
)

func TestSubTaskExecutionContext(t *testing.T) {
	ctx := context.Background()

	tCtx := getMockTaskExecutionContext(ctx, 0)
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	assert.Nil(t, err)

	executionIndex := 0
	originalIndex := 5
	retryAttempt := uint64(1)
	systemFailures := uint64(0)

	stCtx, err := NewSubTaskExecutionContext(ctx, tCtx, taskTemplate, executionIndex, originalIndex, retryAttempt, systemFailures)
	assert.Nil(t, err)

	assert.Equal(t, fmt.Sprintf("notfound-%d-%d", executionIndex, retryAttempt), stCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())

	subtaskTemplate, err := stCtx.TaskReader().Read(ctx)
	assert.Nil(t, err)
	assert.Equal(t, int32(2), subtaskTemplate.TaskTypeVersion)
	assert.Equal(t, podPlugin.ContainerTaskType, subtaskTemplate.Type)
	assert.Equal(t, storage.DataReference("/prefix/"), stCtx.OutputWriter().GetOutputPrefixPath())
	assert.Equal(t, storage.DataReference("/raw_prefix/5/1"), stCtx.OutputWriter().GetRawOutputPrefix())
	assert.Equal(t,
		&tasklog.TemplateVarsByScheme{
			TaskExecution: tasklog.TemplateVars{
				{Regex: logTemplateRegexes.ParentName, Value: "notfound"},
				{Regex: logTemplateRegexes.ExecutionIndex, Value: "0"},
				{Regex: logTemplateRegexes.RetryAttempt, Value: "1"},
				{Regex: logTemplateRegexes.ParentRetryAttempt, Value: "0"},
			},
		},
		stCtx.TaskExecutionMetadata().GetTaskExecutionID().(SubTaskExecutionID).TemplateVarsByScheme(),
	)
}
