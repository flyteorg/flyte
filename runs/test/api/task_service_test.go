package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
)

const (
	testOrg     = "test-org"
	testProject = "test-project"
	testDomain  = "test-domain"
)

func TestDeployTask(t *testing.T) {
	t.Cleanup(func() {
		cleanupTestDB(t)
	})

	ctx := context.Background()
	httpClient := newClient()
	opts := []connect.ClientOption{}

	taskClient := taskconnect.NewTaskServiceClient(httpClient, endpoint, opts...)

	// Deploy a task
	taskID := &task.TaskIdentifier{
		Org:     testOrg,
		Project: testProject,
		Domain:  testDomain,
		Name:    "test-task",
		Version: uniqueString(),
	}

	deployResp, err := taskClient.DeployTask(ctx, connect.NewRequest(&task.DeployTaskRequest{
		TaskId: taskID,
		Spec: &task.TaskSpec{
			TaskTemplate: &core.TaskTemplate{
				Type: "container",
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Image: "alpine:latest",
						Args:  []string{"echo", "hello"},
					},
				},
			},
		},
	}))

	require.NoError(t, err)
	require.NotNil(t, deployResp)
	t.Logf("Task deployed successfully: %v", taskID)

	// Get task details
	getTaskDetailsResp, err := taskClient.GetTaskDetails(ctx, connect.NewRequest(&task.GetTaskDetailsRequest{
		TaskId: taskID,
	}))

	require.NoError(t, err)
	require.NotNil(t, getTaskDetailsResp)
	details := getTaskDetailsResp.Msg.GetDetails()
	assert.Equal(t, taskID.GetOrg(), details.GetTaskId().GetOrg())
	assert.Equal(t, taskID.GetProject(), details.GetTaskId().GetProject())
	assert.Equal(t, taskID.GetDomain(), details.GetTaskId().GetDomain())
	assert.Equal(t, taskID.GetName(), details.GetTaskId().GetName())
	assert.Equal(t, taskID.GetVersion(), details.GetTaskId().GetVersion())
	t.Logf("Task details retrieved successfully: %v", details)
}

func TestListTasks(t *testing.T) {
	t.Cleanup(func() {
		cleanupTestDB(t)
	})

	ctx := context.Background()
	httpClient := newClient()
	opts := []connect.ClientOption{}

	taskClient := taskconnect.NewTaskServiceClient(httpClient, endpoint, opts...)

	// Deploy a task first
	taskID := &task.TaskIdentifier{
		Org:     testOrg,
		Project: testProject,
		Domain:  testDomain,
		Name:    "list-test-task",
		Version: uniqueString(),
	}

	deployResp, err := taskClient.DeployTask(ctx, connect.NewRequest(&task.DeployTaskRequest{
		TaskId: taskID,
		Spec: &task.TaskSpec{
			TaskTemplate: &core.TaskTemplate{
				Type: "container",
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Image: "alpine:latest",
						Args:  []string{"echo", "hello"},
					},
				},
			},
		},
	}))
	require.NotNil(t, deployResp)
	require.NoError(t, err)

	// List tasks
	listResp, err := taskClient.ListTasks(ctx, connect.NewRequest(&task.ListTasksRequest{
		ScopeBy: &task.ListTasksRequest_Org{
			Org: "test-org",
		},
		Request: &common.ListRequest{
			Limit: 10,
		},
	}))
	require.NoError(t, err)
	require.NotNil(t, listResp)

	tasks := listResp.Msg.GetTasks()
	assert.Equal(t, 1, len(tasks))
	t.Logf("Listed %d tasks", len(tasks))
}
