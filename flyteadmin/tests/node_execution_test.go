//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
)

const nodeID = "nodey"
const inputURI = "input uri"

var nodeExecutionId = &core.NodeExecutionIdentifier{
	NodeId: nodeID,
	ExecutionId: &core.WorkflowExecutionIdentifier{
		Project: project,
		Domain:  domain,
		Name:    name,
	},
}

func TestCreateNodeExecution(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	occurredAt := time.Now()
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)
	_, err := client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id:    nodeExecutionId,
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: occurredAtProto,
		},
	})
	assert.Nil(t, err)

	response, err := client.GetNodeExecution(ctx, &admin.NodeExecutionGetRequest{
		Id: nodeExecutionId,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(nodeExecutionId, response.Id))
	assert.Equal(t, core.NodeExecution_RUNNING, response.Closure.Phase)
	assert.Equal(t, inputURI, response.InputUri)
	assert.True(t, proto.Equal(occurredAtProto, response.Closure.StartedAt))
}

func TestCreateNodeExecutionWithParent(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	occurredAt := time.Now()
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)
	_, err := client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id:    nodeExecutionId,
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: occurredAtProto,
		},
	})
	assert.Nil(t, err)

	response, err := client.GetNodeExecution(ctx, &admin.NodeExecutionGetRequest{
		Id: nodeExecutionId,
	})
	assert.Nil(t, err)
	assert.False(t, response.Metadata.IsParentNode)
	assert.True(t, proto.Equal(nodeExecutionId, response.Id))

	_, err = client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id: &core.NodeExecutionIdentifier{
				NodeId:      "child",
				ExecutionId: nodeExecutionId.ExecutionId,
			},
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: occurredAtProto,
			SpecNodeId: "spec",
			RetryGroup: "1",
			ParentNodeMetadata: &event.ParentNodeExecutionMetadata{
				NodeId: nodeExecutionId.NodeId,
			},
		},
	})
	response, err = client.GetNodeExecution(ctx, &admin.NodeExecutionGetRequest{
		Id: &core.NodeExecutionIdentifier{
			NodeId:      "child",
			ExecutionId: nodeExecutionId.ExecutionId,
		},
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
		NodeId:      "child",
		ExecutionId: nodeExecutionId.ExecutionId,
	}, response.Id))
	assert.Nil(t, err)
	assert.False(t, response.Metadata.IsParentNode)

	response, err = client.GetNodeExecution(ctx, &admin.NodeExecutionGetRequest{
		Id: nodeExecutionId,
	})
	assert.Nil(t, err)
	assert.True(t, response.Metadata.IsParentNode)
	assert.True(t, proto.Equal(nodeExecutionId, response.Id))
}

func TestCreateAndUpdateNodeExecution(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	beganRunningAt := time.Now()
	beganRunningAtProto, _ := ptypes.TimestampProto(beganRunningAt)
	_, err := client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id:    nodeExecutionId,
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: beganRunningAtProto,
		},
	})
	assert.Nil(t, err)

	// Create another node execution to assert that we only (and successfully) update the desired node execution.
	otherBeganRunningAt := beganRunningAt.Add(10 * time.Second)
	otherBeganRunningAtProto, _ := ptypes.TimestampProto(otherBeganRunningAt)
	otherNodeExecutionID := &core.NodeExecutionIdentifier{
		NodeId: "other node",
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
	}
	_, err = client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "other request id",
		Event: &event.NodeExecutionEvent{
			Id:    otherNodeExecutionID,
			Phase: core.NodeExecution_QUEUED,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: otherBeganRunningAtProto,
		},
	})
	assert.Nil(t, err)

	succeededAt := beganRunningAt.Add(time.Minute)
	succeededAtProto, _ := ptypes.TimestampProto(succeededAt)
	outputURI := "output uri"
	_, err = client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id:    nodeExecutionId,
			Phase: core.NodeExecution_SUCCEEDED,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: succeededAtProto,
			OutputResult: &event.NodeExecutionEvent_OutputUri{
				OutputUri: outputURI,
			},
		},
	})
	assert.Nil(t, err)

	response, err := client.GetNodeExecution(ctx, &admin.NodeExecutionGetRequest{
		Id: nodeExecutionId,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(nodeExecutionId, response.Id))
	assert.Equal(t, core.NodeExecution_SUCCEEDED, response.Closure.Phase)
	assert.Equal(t, inputURI, response.InputUri)
	assert.Equal(t, outputURI, response.Closure.GetOutputUri())
	assert.True(t, proto.Equal(beganRunningAtProto, response.Closure.StartedAt))
	assert.True(t, proto.Equal(succeededAtProto, response.Closure.UpdatedAt))

	// Assert the other node execution remains unchanged.
	response, err = client.GetNodeExecution(ctx, &admin.NodeExecutionGetRequest{
		Id: otherNodeExecutionID,
	})
	assert.Nil(t, err)
	assert.Equal(t, core.NodeExecution_QUEUED, response.Closure.Phase)
}

func TestCreateAndListNodeExecutions(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	occurredAt := time.Now()
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)
	_, err := client.CreateWorkflowEvent(ctx, &admin.WorkflowExecutionEventRequest{
		RequestId: "request id",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: project,
				Domain:  domain,
				Name:    name,
			},
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: occurredAtProto,
		},
	})
	assert.Nil(t, err)
	_, err = client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id:    nodeExecutionId,
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: occurredAtProto,
		},
	})
	assert.Nil(t, err)

	response, err := client.ListNodeExecutions(ctx, &admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		Limit: 10,
	})

	assert.Nil(t, err)
	assert.Len(t, response.NodeExecutions, 1)
	nodeExecutionResponse := response.NodeExecutions[0]
	assert.True(t, proto.Equal(nodeExecutionId, nodeExecutionResponse.Id))
	assert.Equal(t, core.NodeExecution_RUNNING, nodeExecutionResponse.Closure.Phase)
	assert.Equal(t, inputURI, nodeExecutionResponse.InputUri)
	assert.True(t, proto.Equal(occurredAtProto, nodeExecutionResponse.Closure.StartedAt))
}

func TestListNodeExecutionWithParent(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	occurredAt := time.Now()
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)
	_, err := client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id:    nodeExecutionId,
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: occurredAtProto,
		},
	})
	assert.Nil(t, err)

	_, err = client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id: &core.NodeExecutionIdentifier{
				NodeId:      "child",
				ExecutionId: nodeExecutionId.ExecutionId,
			},
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: occurredAtProto,
			SpecNodeId: "spec",
			RetryGroup: "1",
			ParentNodeMetadata: &event.ParentNodeExecutionMetadata{
				NodeId: nodeExecutionId.NodeId,
			},
		},
	})
	assert.Nil(t, err)

	_, err = client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id: &core.NodeExecutionIdentifier{
				NodeId:      "child2",
				ExecutionId: nodeExecutionId.ExecutionId,
			},
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: occurredAtProto,
			SpecNodeId: "spec",
			RetryGroup: "1",
			ParentNodeMetadata: &event.ParentNodeExecutionMetadata{
				NodeId: nodeExecutionId.NodeId,
			},
		},
	})
	assert.Nil(t, err)

	response, err := client.ListNodeExecutions(ctx, &admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		Limit: 10,
	})

	assert.Nil(t, err)
	assert.Len(t, response.NodeExecutions, 1)
	nodeExecutionResponse := response.NodeExecutions[0]
	assert.True(t, proto.Equal(nodeExecutionId, nodeExecutionResponse.Id))
	assert.Equal(t, core.NodeExecution_RUNNING, nodeExecutionResponse.Closure.Phase)
	assert.Equal(t, inputURI, nodeExecutionResponse.InputUri)
	assert.True(t, proto.Equal(occurredAtProto, nodeExecutionResponse.Closure.StartedAt))

	response, err = client.ListNodeExecutions(ctx, &admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		UniqueParentId: nodeExecutionId.NodeId,
		Limit:          10,
	})

	assert.Nil(t, err)
	assert.Len(t, response.NodeExecutions, 2)
	nodeExecutionResponse = response.NodeExecutions[0]
	assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
		NodeId:      "child2",
		ExecutionId: nodeExecutionId.ExecutionId,
	}, nodeExecutionResponse.Id))
	assert.Equal(t, core.NodeExecution_RUNNING, nodeExecutionResponse.Closure.Phase)
	assert.Equal(t, inputURI, nodeExecutionResponse.InputUri)
	assert.True(t, proto.Equal(occurredAtProto, nodeExecutionResponse.Closure.StartedAt))

	nodeExecutionResponse = response.NodeExecutions[1]
	assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
		NodeId:      "child",
		ExecutionId: nodeExecutionId.ExecutionId,
	}, nodeExecutionResponse.Id))
	assert.Equal(t, core.NodeExecution_RUNNING, nodeExecutionResponse.Closure.Phase)
	assert.Equal(t, inputURI, nodeExecutionResponse.InputUri)
	assert.True(t, proto.Equal(occurredAtProto, nodeExecutionResponse.Closure.StartedAt))
}

func TestCreateChildNodeExecutionForTaskExecution(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	occurredAt := time.Now()
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)

	createTaskAndNodeExecution(ctx, t, client, conn, occurredAtProto)

	_, err := client.CreateTaskEvent(ctx, &admin.TaskExecutionEventRequest{
		RequestId: "request id",
		Event: &event.TaskExecutionEvent{
			TaskId:                taskIdentifier,
			ParentNodeExecutionId: nodeExecutionId,
			Phase:                 core.TaskExecution_RUNNING,
			RetryAttempt:          1,
			OccurredAt:            occurredAtProto,
		},
	})
	assert.Nil(t, err)

	childOccurredAt := occurredAt.Add(time.Minute)
	childOccurredAtProto, _ := ptypes.TimestampProto(childOccurredAt)
	childNodeExecutionID := &core.NodeExecutionIdentifier{
		NodeId: "child_node",
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
	}
	_, err = client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id:    childNodeExecutionID,
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: childOccurredAtProto,
			ParentTaskMetadata: &event.ParentTaskExecutionMetadata{
				Id: taskExecutionIdentifier,
			},
		},
	})
	assert.Nil(t, err)

	response, err := client.GetNodeExecution(ctx, &admin.NodeExecutionGetRequest{
		Id: nodeExecutionId,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(nodeExecutionId, response.Id))

	listResponse, err := client.ListNodeExecutions(ctx, &admin.NodeExecutionListRequest{
		WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		Limit: 10,
	})

	assert.Nil(t, err)
	assert.Len(t, listResponse.NodeExecutions, 1)
	assert.True(t, proto.Equal(nodeExecutionId, listResponse.NodeExecutions[0].Id),
		"list requests should only return top-level nodes for a workflow")

	listForTaskResponse, err := client.ListNodeExecutionsForTask(ctx, &admin.NodeExecutionForTaskListRequest{
		TaskExecutionId: &core.TaskExecutionIdentifier{
			TaskId:          taskExecutionIdentifier.TaskId,
			NodeExecutionId: nodeExecutionId,
			RetryAttempt:    taskExecutionIdentifier.RetryAttempt,
		},
		Limit: 10,
	})

	assert.Nil(t, err)
	assert.Len(t, listForTaskResponse.NodeExecutions, 1)
	assert.True(t, proto.Equal(childNodeExecutionID, listForTaskResponse.NodeExecutions[0].Id),
		"list for task requests should only return nodes launched by a specific task")

	// While we're testing, validate that the parent task execution is correctly marked as a parent.
	taskExecutionResp, err := client.GetTaskExecution(ctx, &admin.TaskExecutionGetRequest{
		Id: taskExecutionIdentifier,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(taskExecutionIdentifier, taskExecutionResp.Id))
	assert.True(t, taskExecutionResp.IsParent)
}
