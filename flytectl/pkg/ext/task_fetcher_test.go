package ext

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyte/flytectl/pkg/filters"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	adminFetcherExt  AdminFetcherExtClient
	adminClient      *mocks.AdminServiceClient
	ctx              context.Context
	taskListResponse *admin.TaskList
	taskFilter       = filters.Filters{}
	task1            *admin.Task
)

func getTaskFetcherSetup() {
	ctx = context.Background()
	adminClient = new(mocks.AdminServiceClient)
	adminFetcherExt = AdminFetcherExtClient{AdminClient: adminClient}

	sortedListLiteralType := core.Variable{
		Type: &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
	}
	variableMap := map[string]*core.Variable{
		"sorted_list1": &sortedListLiteralType,
		"sorted_list2": &sortedListLiteralType,
	}

	task1 := &admin.Task{
		Id: &core.Identifier{
			Name:    "task1",
			Version: "v1",
		},
		Closure: &admin.TaskClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
			CompiledTask: &core.CompiledTask{
				Template: &core.TaskTemplate{
					Interface: &core.TypedInterface{
						Inputs: &core.VariableMap{
							Variables: variableMap,
						},
					},
				},
			},
		},
	}

	task2 := &admin.Task{
		Id: &core.Identifier{
			Name:    "task1",
			Version: "v2",
		},
		Closure: &admin.TaskClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			CompiledTask: &core.CompiledTask{
				Template: &core.TaskTemplate{
					Interface: &core.TypedInterface{
						Inputs: &core.VariableMap{
							Variables: variableMap,
						},
					},
				},
			},
		},
	}

	tasks := []*admin.Task{task2, task1}

	taskListResponse = &admin.TaskList{
		Tasks: tasks,
	}
}

func TestFetchAllVerOfTask(t *testing.T) {
	getTaskFetcherSetup()
	adminClient.OnListTasksMatch(mock.Anything, mock.Anything).Return(taskListResponse, nil)
	_, err := adminFetcherExt.FetchAllVerOfTask(ctx, "taskName", "project", "domain", taskFilter)
	assert.Nil(t, err)
}

func TestFetchTaskVersion(t *testing.T) {
	getTaskFetcherSetup()
	adminClient.OnGetTaskMatch(mock.Anything, mock.Anything).Return(task1, nil)
	_, err := adminFetcherExt.FetchTaskVersion(ctx, "task1", "v1", "project", "domain")
	assert.Nil(t, err)
}

func TestFetchAllVerOfTaskError(t *testing.T) {
	getTaskFetcherSetup()
	adminClient.OnListTasksMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchAllVerOfTask(ctx, "taskName", "project", "domain", taskFilter)
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestFetchAllVerOfTaskFilterError(t *testing.T) {
	getTaskFetcherSetup()
	taskFilter = filters.Filters{
		FieldSelector: "hello=",
	}
	adminClient.OnListTasksMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchAllVerOfTask(ctx, "taskName", "project", "domain", taskFilter)
	assert.NotNil(t, err)
}

func TestFetchAllVerOfTaskEmptyResponse(t *testing.T) {
	taskListResponse := &admin.TaskList{}
	getTaskFetcherSetup()
	taskFilter = filters.Filters{
		FieldSelector: "",
	}
	adminClient.OnListTasksMatch(mock.Anything, mock.Anything).Return(taskListResponse, nil)
	_, err := adminFetcherExt.FetchAllVerOfTask(ctx, "taskName", "project", "domain", taskFilter)
	assert.Equal(t, fmt.Errorf("no tasks retrieved for taskName"), err)
}

func TestFetchTaskLatestVersion(t *testing.T) {
	getTaskFetcherSetup()
	adminClient.OnListTasksMatch(mock.Anything, mock.Anything).Return(taskListResponse, nil)
	_, err := adminFetcherExt.FetchTaskLatestVersion(ctx, "taskName", "project", "domain", taskFilter)
	assert.Nil(t, err)
}

func TestFetchTaskLatestVersionError(t *testing.T) {
	taskListResponse := &admin.TaskList{}
	getTaskFetcherSetup()
	adminClient.OnListTasksMatch(mock.Anything, mock.Anything).Return(taskListResponse, nil)
	_, err := adminFetcherExt.FetchTaskLatestVersion(ctx, "taskName", "project", "domain", taskFilter)
	assert.Equal(t, fmt.Errorf("no tasks retrieved for taskName"), err)
}
