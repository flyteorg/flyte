// +build integration

package tests

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils/labeled"

	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"

	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	ptypesStruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/lyft/flyteadmin/pkg/manager/impl/testutils"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
)

const taskExecInputURI = "s3://flyte/metadata/admin/input/uri"

var taskIdentifier = &core.Identifier{
	ResourceType: core.ResourceType_TASK,
	Project:      project,
	Domain:       "development",
	Name:         "task name",
	Version:      "task version",
}

var taskExecutionIdentifier = &core.TaskExecutionIdentifier{
	TaskId:          taskIdentifier,
	NodeExecutionId: nodeExecutionId,
	RetryAttempt:    1,
}

func createTaskAndNodeExecution(
	ctx context.Context, t *testing.T, client service.AdminServiceClient, conn *grpc.ClientConn,
	occurredAtProto *timestamp.Timestamp) {
	_, err := client.CreateTask(ctx, &admin.TaskCreateRequest{
		Id:   taskIdentifier,
		Spec: testutils.GetValidTaskRequest().Spec,
	})
	assert.Nil(t, err)

	_, err = client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id:         nodeExecutionId,
			Phase:      core.NodeExecution_RUNNING,
			InputUri:   inputURI,
			OccurredAt: occurredAtProto,
		},
	})
	assert.Nil(t, err)
}

func TestCreateTaskExecution(t *testing.T) {
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
			InputUri:              taskExecInputURI,
		},
	})
	assert.Nil(t, err)
	response, err := client.GetTaskExecution(ctx, &admin.TaskExecutionGetRequest{
		Id: taskExecutionIdentifier,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(taskExecutionIdentifier, response.Id))
	assert.Equal(t, core.TaskExecution_RUNNING, response.Closure.Phase)
	assert.Equal(t, taskExecInputURI, response.InputUri)
}

func TestCreateAndUpdateTaskExecution(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	beganAt := time.Now()
	beganAtProto, _ := ptypes.TimestampProto(beganAt)

	createTaskAndNodeExecution(ctx, t, client, conn, beganAtProto)

	// Create first attempt
	_, err := client.CreateTaskEvent(ctx, &admin.TaskExecutionEventRequest{
		RequestId: "request id",
		Event: &event.TaskExecutionEvent{
			TaskId:                taskIdentifier,
			ParentNodeExecutionId: nodeExecutionId,
			Phase:                 core.TaskExecution_FAILED,
			OccurredAt:            beganAtProto,
			InputUri:              taskExecInputURI,
			RetryAttempt:          0,
		},
	})
	assert.Nil(t, err)
	// And make sure we get it back okay.
	_, err = client.GetTaskExecution(ctx, &admin.TaskExecutionGetRequest{
		Id: &core.TaskExecutionIdentifier{
			TaskId:          taskIdentifier,
			NodeExecutionId: nodeExecutionId,
		},
	})
	assert.Nil(t, err)

	// Create second attempt
	_, err = client.CreateTaskEvent(ctx, &admin.TaskExecutionEventRequest{
		RequestId: "request id",
		Event: &event.TaskExecutionEvent{
			TaskId:                taskIdentifier,
			ParentNodeExecutionId: nodeExecutionId,
			Phase:                 core.TaskExecution_RUNNING,
			RetryAttempt:          1,
			OccurredAt:            beganAtProto,
			InputUri:              taskExecInputURI,
		},
	})
	assert.Nil(t, err)
	// And make sure we get it back okay.
	_, err = client.GetTaskExecution(ctx, &admin.TaskExecutionGetRequest{
		Id: taskExecutionIdentifier,
	})
	assert.Nil(t, err)
	endedAt := beganAt.Add(time.Minute)
	endedAtProto, _ := ptypes.TimestampProto(endedAt)
	_, err = client.CreateTaskEvent(ctx, &admin.TaskExecutionEventRequest{
		RequestId: "request id",
		Event: &event.TaskExecutionEvent{
			TaskId:                taskIdentifier,
			ParentNodeExecutionId: nodeExecutionId,
			Phase:                 core.TaskExecution_SUCCEEDED,
			RetryAttempt:          1,
			OccurredAt:            endedAtProto,
			InputUri:              taskExecInputURI,
		},
	})
	assert.Nil(t, err)
	response, err := client.GetTaskExecution(ctx, &admin.TaskExecutionGetRequest{
		Id: taskExecutionIdentifier,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(taskExecutionIdentifier, response.Id))
	assert.Equal(t, core.TaskExecution_SUCCEEDED, response.Closure.Phase)
	// Float comparisons from timestamp conversions are annoying. Approximately equal is good enough!
	assert.Contains(t, []int64{59, 60, 61}, response.Closure.Duration.Seconds)
	assert.True(t, proto.Equal(beganAtProto, response.Closure.CreatedAt))
	assert.True(t, proto.Equal(endedAtProto, response.Closure.UpdatedAt))
	assert.Equal(t, taskExecInputURI, response.InputUri)
}

func TestCreateAndUpdateTaskExecutionPhaseVersion(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	beganAt := time.Now()
	beganAtProto, _ := ptypes.TimestampProto(beganAt)

	createTaskAndNodeExecution(ctx, t, client, conn, beganAtProto)

	// Create first event
	_, err := client.CreateTaskEvent(ctx, &admin.TaskExecutionEventRequest{
		RequestId: "request id",
		Event: &event.TaskExecutionEvent{
			TaskId:                taskIdentifier,
			ParentNodeExecutionId: nodeExecutionId,
			Phase:                 core.TaskExecution_RUNNING,
			OccurredAt:            beganAtProto,
			InputUri:              taskExecInputURI,
			RetryAttempt:          0,
		},
	})
	assert.Nil(t, err)
	// And make sure we get it back okay.
	_, err = client.GetTaskExecution(ctx, &admin.TaskExecutionGetRequest{
		Id: &core.TaskExecutionIdentifier{
			TaskId:          taskIdentifier,
			NodeExecutionId: nodeExecutionId,
		},
	})
	assert.Nil(t, err)

	customInfo := ptypesStruct.Struct{
		Fields: map[string]*ptypesStruct.Value{
			"phase": {
				Kind: &ptypesStruct.Value_StringValue{
					StringValue: "value",
				},
			},
		},
	}

	// Create second event
	_, err = client.CreateTaskEvent(ctx, &admin.TaskExecutionEventRequest{
		RequestId: "request id",
		Event: &event.TaskExecutionEvent{
			TaskId:                taskIdentifier,
			ParentNodeExecutionId: nodeExecutionId,
			Phase:                 core.TaskExecution_RUNNING,
			PhaseVersion:          1,
			OccurredAt:            beganAtProto,
			InputUri:              taskExecInputURI,
			RetryAttempt:          0,
			CustomInfo:            &customInfo,
		},
	})
	assert.Nil(t, err)

	response, err := client.GetTaskExecution(ctx, &admin.TaskExecutionGetRequest{
		Id: &core.TaskExecutionIdentifier{
			TaskId:          taskIdentifier,
			NodeExecutionId: nodeExecutionId,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, core.TaskExecution_RUNNING, response.Closure.Phase)
	assert.True(t, proto.Equal(&customInfo, response.Closure.CustomInfo))
}

func TestCreateAndListTaskExecution(t *testing.T) {
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
			InputUri:              taskExecInputURI,
		},
	})
	assert.Nil(t, err)
	response, err := client.ListTaskExecutions(ctx, &admin.TaskExecutionListRequest{
		NodeExecutionId: nodeExecutionId,
		Limit:           10,
		Filters: "eq(task.project, project)+eq(task.domain, development)+eq(task.name, task name)+" +
			"eq(task.version, task version)+eq(task_execution.retry_attempt, 1)",
	})
	assert.Len(t, response.TaskExecutions, 1)
	assert.Nil(t, err)
}

func TestGetTaskExecutionData(t *testing.T) {
	labeled.SetMetricKeys(contextutils.TaskIDKey)
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	// Write output data to local memory store
	u, err := url.Parse("http://minio:9000")
	if err != nil {
		t.Fatalf("Failed to parse local memory store url %v", err)
	}
	store, err := storage.NewDataStore(&storage.Config{
		Type:          storage.TypeMinio,
		InitContainer: "flyte",
		Connection: storage.ConnectionConfig{
			AccessKey:  "minio",
			AuthType:   "accesskey",
			SecretKey:  "miniostorage",
			DisableSSL: true,
			Endpoint:   config.URL{URL: *u},
			Region:     "my-region",
		},
	}, mockScope.NewTestScope().NewSubScope("task_exec"))
	if err != nil {
		t.Fatalf("Failed to initialize storage config: %v", err)
	}

	if err != nil {
		t.Fatalf(err.Error())
	}

	ctx := context.Background()
	inputRef, err := store.ConstructReference(ctx, store.GetBaseContainerFQN(ctx), "metadata", "admin", "input", "uri")
	if err != nil {
		t.Fatalf("Failed to construct data reference [%s]. Error: %v", taskExecInputURI, err)
	}
	dataToStore := "task execution input data"
	err = store.WriteRaw(ctx, inputRef, int64(len(dataToStore)), storage.Options{}, strings.NewReader(dataToStore))
	if err != nil {
		t.Fatalf("Failed to write data. Error: %v", err)
	}

	outputRef, err := store.ConstructReference(ctx, store.GetBaseContainerFQN(ctx), "metadata", "admin", "output", "uri")
	if err != nil {
		t.Fatalf("Failed to construct data reference. Error: %v", err)
	}

	dataToStore = "task execution output data"
	err = store.WriteRaw(ctx, outputRef, int64(len(dataToStore)), storage.Options{}, strings.NewReader(dataToStore))
	if err != nil {
		t.Fatalf("Failed to write data. Error: %v", err)
	}

	beganAt := time.Now()
	beganAtProto, _ := ptypes.TimestampProto(beganAt)

	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	createTaskAndNodeExecution(ctx, t, client, conn, beganAtProto)

	// Create first attempt
	_, err = client.CreateTaskEvent(ctx, &admin.TaskExecutionEventRequest{
		RequestId: "request id",
		Event: &event.TaskExecutionEvent{
			TaskId:                taskIdentifier,
			ParentNodeExecutionId: nodeExecutionId,
			Phase:                 core.TaskExecution_SUCCEEDED,
			OccurredAt:            beganAtProto,
			InputUri:              taskExecInputURI,
			OutputResult: &event.TaskExecutionEvent_OutputUri{
				OutputUri: "s3://flyte/metadata/admin/output/uri",
			},
			RetryAttempt: 0,
		},
	})
	assert.Nil(t, err)

	// And make sure we get back the data okay.
	resp, err := client.GetTaskExecutionData(ctx, &admin.TaskExecutionGetDataRequest{
		Id: &core.TaskExecutionIdentifier{
			TaskId:          taskIdentifier,
			NodeExecutionId: nodeExecutionId,
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	assert.NotEmpty(t, resp.Inputs.Url)
	assert.Equal(t, int64(25), resp.Inputs.Bytes)
	assert.NotEmpty(t, resp.Outputs.Url)
	assert.Equal(t, int64(26), resp.Outputs.Bytes)
}
