package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

var taskIdentifier = &core.Identifier{
	ResourceType: core.ResourceType_TASK,
	Name:         "Name",
	Domain:       "Domain",
	Project:      "Project",
	Version:      "Version",
}

func TestTaskHappyCase(t *testing.T) {
	ctx := context.Background()

	mockTaskManager := mocks.MockTaskManager{}
	mockTaskManager.SetCreateCallback(
		func(ctx context.Context,
			request *admin.TaskCreateRequest) (*admin.TaskCreateResponse, error) {
			return &admin.TaskCreateResponse{}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		taskManager: &mockTaskManager,
	})

	resp, err := mockServer.CreateTask(ctx, &admin.TaskCreateRequest{
		Id: taskIdentifier,
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
}

func TestTaskError(t *testing.T) {
	ctx := context.Background()

	mockTaskManager := mocks.MockTaskManager{}
	mockTaskManager.SetCreateCallback(
		func(ctx context.Context,
			request *admin.TaskCreateRequest) (*admin.TaskCreateResponse, error) {
			return nil, errors.GetMissingEntityError(core.ResourceType_TASK.String(), request.Id)
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		taskManager: &mockTaskManager,
	})

	resp, err := mockServer.CreateTask(ctx, &admin.TaskCreateRequest{
		Id: &core.Identifier{
			Project: "project",
			Domain:  "staging",
			Name:    "name",
			Version: "version",
		},
	})
	assert.Nil(t, resp)

	utils.AssertEqualWithSanitizedRegex(t, "missing entity of type TASK with "+
		"identifier project:\"project\" domain:\"staging\" name:\"name\" version:\"version\"", err.Error())
}

func TestListUniqueTaskIds(t *testing.T) {
	ctx := context.Background()

	mockTaskManager := mocks.MockTaskManager{}
	mockTaskManager.SetListUniqueIdsFunc(func(ctx context.Context, request *admin.NamedEntityIdentifierListRequest) (
		*admin.NamedEntityIdentifierList, error) {

		assert.Equal(t, "staging", request.Domain)
		return nil, nil
	})
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		taskManager: &mockTaskManager,
	})

	resp, err := mockServer.ListTaskIds(ctx, &admin.NamedEntityIdentifierListRequest{
		Project: "project",
		Domain:  "staging",
	})

	assert.NoError(t, err)
	assert.Nil(t, resp)
}
