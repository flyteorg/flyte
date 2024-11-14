package array

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	execmocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
)

type bufferedEventRecorder struct {
	taskExecutionEvents []*event.TaskExecutionEvent
	nodeExecutionEvents []*event.NodeExecutionEvent
}

func (b *bufferedEventRecorder) RecordTaskEvent(ctx context.Context, taskExecutionEvent *event.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	b.taskExecutionEvents = append(b.taskExecutionEvents, taskExecutionEvent)
	return nil
}

func (b *bufferedEventRecorder) RecordNodeEvent(ctx context.Context, nodeExecutionEvent *event.NodeExecutionEvent, eventConfig *config.EventConfig) error {
	b.nodeExecutionEvents = append(b.nodeExecutionEvents, nodeExecutionEvent)
	return nil
}

func newBufferedEventRecorder() *bufferedEventRecorder {
	return &bufferedEventRecorder{}
}

func TestGetPluginLogs(t *testing.T) {
	// initialize log plugin
	logConfig := &logs.LogConfig{
		Templates: []tasklog.TemplateLogPlugin{
			{
				Name:        "foo",
				DisplayName: "bar",
				TemplateURIs: []tasklog.TemplateURI{
					"/console/projects/{{.executionProject}}/domains/{{.executionDomain}}/executions/{{.executionName}}/nodeId/{{.nodeID}}/taskId/{{.taskID}}/attempt/{{.taskRetryAttempt}}/mappedIndex/{{.subtaskExecutionIndex}}/mappedAttempt/{{.subtaskRetryAttempt}}/view/logs?duration=all",
				},
			},
			{
				Name:        "log_link",
				DisplayName: "Log Link",
				TemplateURIs: []tasklog.TemplateURI{
					"/{{.podName}}/{{.containerName}}",
				},
			},
		},
	}

	mapLogPlugin, err := logs.InitializeLogPlugins(logConfig)
	assert.Nil(t, err)

	// create NodeExecutionContext
	nCtx := &mocks.NodeExecutionContext{}
	nCtx.OnCurrentAttempt().Return(uint32(0))

	executionContext := &execmocks.ExecutionContext{}
	executionContext.OnGetEventVersion().Return(1)
	executionContext.OnGetParentInfo().Return(nil)
	executionContext.OnGetTask(taskRef).Return(
		&v1alpha1.TaskSpec{
			TaskTemplate: &idlcore.TaskTemplate{
				Id: &idlcore.Identifier{
					ResourceType: idlcore.ResourceType_TASK,
					Project:      "task_project",
					Domain:       "task_domain",
					Name:         "task_name",
					Version:      "task_version",
				},
			},
		},
		nil,
	)
	nCtx.OnExecutionContext().Return(executionContext)

	nCtx.OnNode().Return(&arrayNodeSpecTaskMinStore)

	nodeExecutionMetadata := &mocks.NodeExecutionMetadata{}
	nodeExecutionMetadata.OnGetNamespace().Return("node_namespace")
	nodeExecutionMetadata.OnGetNodeExecutionID().Return(&idlcore.NodeExecutionIdentifier{
		NodeId: "node_id",
		ExecutionId: &idlcore.WorkflowExecutionIdentifier{
			Project: "node_project",
			Domain:  "node_domain",
			Name:    "node_name",
		},
	})
	nodeExecutionMetadata.OnGetOwnerID().Return(types.NamespacedName{
		Namespace: "wf_namespace",
		Name:      "wf_name",
	})
	nCtx.OnNodeExecutionMetadata().Return(nodeExecutionMetadata)

	nCtx.OnNodeID().Return("foo")

	// call `getPluginLogs`
	logs, err := getPluginLogs(mapLogPlugin, nCtx, 1, 0, "test_pod", "test_cont")
	assert.Nil(t, err)

	assert.Equal(t, len(logConfig.Templates), len(logs))
	assert.Equal(t, "bar", logs[0].Name)
	assert.Equal(t, "/console/projects/node_project/domains/node_domain/executions/node_name/nodeId/foo/taskId/task_name/attempt/0/mappedIndex/1/mappedAttempt/0/view/logs?duration=all", logs[0].Uri)
	assert.Equal(t, "Log Link", logs[1].Name)
	assert.Equal(t, "/test_pod/test_cont", logs[1].Uri)
}
