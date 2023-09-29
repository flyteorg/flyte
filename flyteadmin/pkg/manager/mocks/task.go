package mocks

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateTaskFunc func(ctx context.Context, request admin.TaskCreateRequest) (*admin.TaskCreateResponse, error)
type ListUniqueIdsFunc func(ctx context.Context, request admin.NamedEntityIdentifierListRequest) (*admin.NamedEntityIdentifierList, error)

type MockTaskManager struct {
	createTaskFunc    CreateTaskFunc
	listUniqueIdsFunc ListUniqueIdsFunc
}

func (r *MockTaskManager) SetCreateCallback(createFunction CreateTaskFunc) {
	r.createTaskFunc = createFunction
}

func (r *MockTaskManager) CreateTask(
	ctx context.Context,
	request admin.TaskCreateRequest) (*admin.TaskCreateResponse, error) {
	if r.createTaskFunc != nil {
		return r.createTaskFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockTaskManager) GetTask(ctx context.Context, request admin.ObjectGetRequest) (*admin.Task, error) {
	return nil, nil
}

func (r *MockTaskManager) ListTasks(ctx context.Context, request admin.ResourceListRequest) (*admin.TaskList, error) {
	return nil, nil
}

func (r *MockTaskManager) SetListUniqueIdsFunc(fn ListUniqueIdsFunc) {
	r.listUniqueIdsFunc = fn
}

func (r *MockTaskManager) ListUniqueTaskIdentifiers(ctx context.Context, request admin.NamedEntityIdentifierListRequest) (
	*admin.NamedEntityIdentifierList, error) {

	if r.listUniqueIdsFunc != nil {
		return r.listUniqueIdsFunc(ctx, request)
	}

	return nil, nil
}
