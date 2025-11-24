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
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	"github.com/flyteorg/flyte/flytestdlib/bitarray"
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
			tasklog.TemplateLogPlugin{
				Name:        "foo",
				DisplayName: "bar",
				TemplateURIs: []tasklog.TemplateURI{
					"/console/projects/{{.executionProject}}/domains/{{.executionDomain}}/executions/{{.executionName}}/nodeId/{{.nodeID}}/taskId/{{.taskID}}/attempt/{{.taskRetryAttempt}}/mappedIndex/{{.subtaskExecutionIndex}}/mappedAttempt/{{.subtaskRetryAttempt}}/view/logs?duration=all",
				},
			},
		},
	}

	mapLogPlugin, err := logs.InitializeLogPlugins(logConfig)
	assert.Nil(t, err)

	// create NodeExecutionContext
	nCtx := &mocks.NodeExecutionContext{}
	nCtx.On("CurrentAttempt").Return(uint32(0))

	executionContext := &execmocks.ExecutionContext{}
	executionContext.On("GetEventVersion").Return(v1alpha1.EventVersion1)
	executionContext.On("GetParentInfo").Return(nil)
	executionContext.On("GetTask", taskRef).Return(
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
	nCtx.On("ExecutionContext").Return(executionContext)

	nCtx.On("Node").Return(&arrayNodeSpec)

	nodeExecutionMetadata := &mocks.NodeExecutionMetadata{}
	nodeExecutionMetadata.On("GetNamespace").Return("node_namespace")
	nodeExecutionMetadata.On("GetNodeExecutionID").Return(&idlcore.NodeExecutionIdentifier{
		NodeId: "node_id",
		ExecutionId: &idlcore.WorkflowExecutionIdentifier{
			Project: "node_project",
			Domain:  "node_domain",
			Name:    "node_name",
		},
	})
	nodeExecutionMetadata.On("GetOwnerID").Return(types.NamespacedName{
		Namespace: "wf_namespace",
		Name:      "wf_name",
	})
	nCtx.On("NodeExecutionMetadata").Return(nodeExecutionMetadata)

	nCtx.On("NodeID").Return("foo")

	// call `getPluginLogs`
	logs, err := getPluginLogs(mapLogPlugin, nCtx, 1, 0)
	assert.Nil(t, err)

	assert.Equal(t, len(logConfig.Templates), len(logs))
	assert.Equal(t, "bar", logs[0].GetName())
	assert.Equal(t, "/console/projects/node_project/domains/node_domain/executions/node_name/nodeId/foo/taskId/task_name/attempt/0/mappedIndex/1/mappedAttempt/0/view/logs?duration=all", logs[0].GetUri())
}

func TestGetExternalResourceID(t *testing.T) {

	tests := []struct {
		nodeID                     string
		currentNodeAttempt         uint32
		index                      int
		retryAttempt               uint32
		expectedExternalResourceID string
	}{
		{
			nodeID:                     "n2",
			currentNodeAttempt:         0,
			index:                      0,
			retryAttempt:               0,
			expectedExternalResourceID: "exec_name-n2-0-n0-0",
		},
		{
			nodeID:                     "n0",
			currentNodeAttempt:         1,
			index:                      2,
			retryAttempt:               3,
			expectedExternalResourceID: "exec_name-n0-1-n2-3",
		},
	}

	for _, test := range tests {
		execContext := &execmocks.ExecutionContext{}
		execContext.On("GetEventVersion").Return(v1alpha1.EventVersion0)

		nodeExecMetadata := &mocks.NodeExecutionMetadata{}
		nodeExecMetadata.On("GetOwnerID").Return(types.NamespacedName{Name: "exec_name"})

		nCtx := &mocks.NodeExecutionContext{}
		nCtx.On("NodeID").Return(test.nodeID)
		nCtx.On("ExecutionContext").Return(execContext)
		nCtx.On("NodeExecutionMetadata").Return(nodeExecMetadata)
		nCtx.On("CurrentAttempt").Return(test.currentNodeAttempt)

		externalResourceID, err := generateExternalResourceID(nCtx, test.index, test.retryAttempt)
		assert.Nil(t, err)
		assert.Equal(t, test.expectedExternalResourceID, externalResourceID)
	}
}

func TestUpdateExternalResourceSubnodePhases(t *testing.T) {
	tests := []struct {
		name                             string
		subNodePhases                    []v1alpha1.NodePhase
		subNodeRetryAttempts             []uint32
		existingExternalResources        []*event.ExternalResourceInfo
		expectedUpdatedExternalResources []*event.ExternalResourceInfo
	}{
		{
			name: "all new resources - empty existing",
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseSucceeded,
				v1alpha1.NodePhaseFailed,
				v1alpha1.NodePhaseRunning,
			},
			subNodeRetryAttempts:      []uint32{0, 1, 2},
			existingExternalResources: []*event.ExternalResourceInfo{},
			expectedUpdatedExternalResources: []*event.ExternalResourceInfo{
				{
					ExternalId:   "test-name-test-node-0-n0-0",
					Index:        0,
					RetryAttempt: 0,
					Phase:        idlcore.TaskExecution_SUCCEEDED,
				},
				{
					ExternalId:   "test-name-test-node-0-n1-1",
					Index:        1,
					RetryAttempt: 1,
					Phase:        idlcore.TaskExecution_FAILED,
				},
				{
					ExternalId:   "test-name-test-node-0-n2-2",
					Index:        2,
					RetryAttempt: 2,
					Phase:        idlcore.TaskExecution_RUNNING,
				},
			},
		},
		{
			name: "some existing resources - preserve order",
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseSucceeded,
				v1alpha1.NodePhaseFailed,
				v1alpha1.NodePhaseRunning,
				v1alpha1.NodePhaseQueued,
			},
			subNodeRetryAttempts: []uint32{0, 1, 2, 3},
			existingExternalResources: []*event.ExternalResourceInfo{
				{
					ExternalId:   "existing-0",
					Index:        0,
					RetryAttempt: 0,
					Phase:        idlcore.TaskExecution_SUCCEEDED,
				},
				{
					ExternalId:   "existing-2",
					Index:        2,
					RetryAttempt: 2,
					Phase:        idlcore.TaskExecution_RUNNING,
				},
			},
			expectedUpdatedExternalResources: []*event.ExternalResourceInfo{
				{
					ExternalId:   "existing-0",
					Index:        0,
					RetryAttempt: 0,
					Phase:        idlcore.TaskExecution_SUCCEEDED,
				},
				{
					ExternalId:   "test-name-test-node-0-n1-1",
					Index:        1,
					RetryAttempt: 1,
					Phase:        idlcore.TaskExecution_FAILED,
				},
				{
					ExternalId:   "existing-2",
					Index:        2,
					RetryAttempt: 2,
					Phase:        idlcore.TaskExecution_RUNNING,
				},
				{
					ExternalId:   "test-name-test-node-0-n3-3",
					Index:        3,
					RetryAttempt: 3,
					Phase:        idlcore.TaskExecution_QUEUED,
				},
			},
		},
		{
			name: "all existing resources - no new ones",
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseSucceeded,
				v1alpha1.NodePhaseFailed,
			},
			subNodeRetryAttempts: []uint32{0, 1},
			existingExternalResources: []*event.ExternalResourceInfo{
				{
					ExternalId:   "existing-0",
					Index:        0,
					RetryAttempt: 0,
					Phase:        idlcore.TaskExecution_SUCCEEDED,
				},
				{
					ExternalId:   "existing-1",
					Index:        1,
					RetryAttempt: 1,
					Phase:        idlcore.TaskExecution_FAILED,
				},
			},
			expectedUpdatedExternalResources: []*event.ExternalResourceInfo{
				{
					ExternalId:   "existing-0",
					Index:        0,
					RetryAttempt: 0,
					Phase:        idlcore.TaskExecution_SUCCEEDED,
				},
				{
					ExternalId:   "existing-1",
					Index:        1,
					RetryAttempt: 1,
					Phase:        idlcore.TaskExecution_FAILED,
				},
			},
		},
		{
			name: "preserve existing resource data",
			subNodePhases: []v1alpha1.NodePhase{
				v1alpha1.NodePhaseSucceeded,
				v1alpha1.NodePhaseFailed,
				v1alpha1.NodePhaseRunning,
			},
			subNodeRetryAttempts: []uint32{5, 10, 15}, // Different from existing
			existingExternalResources: []*event.ExternalResourceInfo{
				{
					ExternalId:   "preserve-this-id",
					Index:        1,
					RetryAttempt: 99,                              // Should preserve this, not use 10
					Phase:        idlcore.TaskExecution_SUCCEEDED, // Should preserve this
					Logs: []*idlcore.TaskLog{
						{Name: "test-log"},
					},
					CacheStatus: idlcore.CatalogCacheStatus_CACHE_HIT,
				},
			},
			expectedUpdatedExternalResources: []*event.ExternalResourceInfo{
				{
					ExternalId:   "test-name-test-node-0-n0-5",
					Index:        0,
					RetryAttempt: 5,
					Phase:        idlcore.TaskExecution_SUCCEEDED,
				},
				{
					ExternalId:   "preserve-this-id",
					Index:        1,
					RetryAttempt: 99, // Preserved from existing
					Phase:        idlcore.TaskExecution_SUCCEEDED,
					Logs: []*idlcore.TaskLog{
						{Name: "test-log"},
					},
					CacheStatus: idlcore.CatalogCacheStatus_CACHE_HIT,
				},
				{
					ExternalId:   "test-name-test-node-0-n2-15",
					Index:        2,
					RetryAttempt: 15,
					Phase:        idlcore.TaskExecution_RUNNING,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nCtx := &mocks.NodeExecutionContext{}
			nodeStateReader := &mocks.NodeStateReader{}
			executionContext := &execmocks.ExecutionContext{}
			nodeExecutionMetadata := &mocks.NodeExecutionMetadata{}

			arrayNodeState := &handler.ArrayNodeState{}
			if len(test.subNodePhases) > 0 {
				subNodePhasesArray, err := bitarray.NewCompactArray(uint(len(test.subNodePhases)), bitarray.Item(v1alpha1.NodePhaseRecovered))
				assert.NoError(t, err)
				subNodeRetryAttemptsArray, err := bitarray.NewCompactArray(uint(len(test.subNodeRetryAttempts)), bitarray.Item(100))
				assert.NoError(t, err)

				for i, phase := range test.subNodePhases {
					subNodePhasesArray.SetItem(i, bitarray.Item(phase))
				}
				for i, retryAttempt := range test.subNodeRetryAttempts {
					subNodeRetryAttemptsArray.SetItem(i, bitarray.Item(retryAttempt))
				}

				arrayNodeState.SubNodePhases = subNodePhasesArray
				arrayNodeState.SubNodeRetryAttempts = subNodeRetryAttemptsArray
			} else {
				subNodePhasesArray, err := bitarray.NewCompactArray(0, bitarray.Item(v1alpha1.NodePhaseRecovered))
				assert.NoError(t, err)
				subNodeRetryAttemptsArray, err := bitarray.NewCompactArray(0, bitarray.Item(100))
				assert.NoError(t, err)
				arrayNodeState.SubNodePhases = subNodePhasesArray
				arrayNodeState.SubNodeRetryAttempts = subNodeRetryAttemptsArray
			}

			executionContext.On("GetEventVersion").Return(v1alpha1.EventVersion0)

			nodeExecutionMetadata.On("GetOwnerID").Return(types.NamespacedName{
				Namespace: "test-namespace",
				Name:      "test-name",
			})

			nCtx.On("NodeStateReader").Return(nodeStateReader)
			nCtx.On("ExecutionContext").Return(executionContext)
			nCtx.On("NodeExecutionMetadata").Return(nodeExecutionMetadata)
			nCtx.On("NodeID").Return("test-node")
			nCtx.On("CurrentAttempt").Return(uint32(0))
			nodeStateReader.On("GetArrayNodeState").Return(*arrayNodeState)

			result, err := updateExternalResourceSubnodePhases(nCtx, test.existingExternalResources)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, test.expectedUpdatedExternalResources, result)
		})
	}
}
