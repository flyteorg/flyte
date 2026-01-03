package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
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
	// Ensure a clean test database
	cleanupTestDB(t)

	// Cleanup after test
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

	// Get versions of the task
	getVersionsResp, err := taskClient.ListVersions(ctx, connect.NewRequest(&task.ListVersionsRequest{
		TaskName: &task.TaskName{
			Org:     testOrg,
			Project: testProject,
			Domain:  testDomain,
			Name:    "test-task",
		},
	}))

	require.NoError(t, err)
	require.NotNil(t, getVersionsResp)
	versions := getVersionsResp.Msg.GetVersions()
	assert.Equal(t, 1, len(versions))
	assert.Equal(t, taskID.GetVersion(), versions[0].Version)
	t.Logf("Task versions retrieved successfully: %v", versions)
}
