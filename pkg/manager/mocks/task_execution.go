package mocks

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateTaskExecutionEventFunc func(ctx context.Context, request admin.TaskExecutionEventRequest) (
	*admin.TaskExecutionEventResponse, error)
type GetTaskExecutionFunc func(ctx context.Context, request admin.TaskExecutionGetRequest) (
	*admin.TaskExecution, error)
type ListTaskExecutionsFunc func(ctx context.Context, request admin.TaskExecutionListRequest) (
	*admin.TaskExecutionList, error)
type GetTaskExecutionDataFunc func(ctx context.Context, request admin.TaskExecutionGetDataRequest) (
	*admin.TaskExecutionGetDataResponse, error)

type MockTaskExecutionManager struct {
	createTaskExecutionEventFunc CreateTaskExecutionEventFunc
	getTaskExecutionFunc         GetTaskExecutionFunc
	listTaskExecutionsFunc       ListTaskExecutionsFunc
	getTaskExecutionDataFunc     GetTaskExecutionDataFunc
}

func (m *MockTaskExecutionManager) CreateTaskExecutionEvent(
	ctx context.Context, request admin.TaskExecutionEventRequest) (*admin.TaskExecutionEventResponse, error) {
	if m.createTaskExecutionEventFunc != nil {
		return m.createTaskExecutionEventFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockTaskExecutionManager) SetCreateTaskEventCallback(
	createFunc CreateTaskExecutionEventFunc) {
	m.createTaskExecutionEventFunc = createFunc
}

func (m *MockTaskExecutionManager) GetTaskExecution(
	ctx context.Context, request admin.TaskExecutionGetRequest) (*admin.TaskExecution, error) {
	if m.getTaskExecutionFunc != nil {
		return m.getTaskExecutionFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockTaskExecutionManager) SetGetTaskExecutionCallback(
	getTaskExecutionFunc GetTaskExecutionFunc) {
	m.getTaskExecutionFunc = getTaskExecutionFunc
}

func (m *MockTaskExecutionManager) ListTaskExecutions(
	ctx context.Context, request admin.TaskExecutionListRequest) (*admin.TaskExecutionList, error) {
	if m.listTaskExecutionsFunc != nil {
		return m.listTaskExecutionsFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockTaskExecutionManager) SetListTaskExecutionsCallback(
	listTaskExecutionsFunc ListTaskExecutionsFunc) {
	m.listTaskExecutionsFunc = listTaskExecutionsFunc
}

func (m *MockTaskExecutionManager) GetTaskExecutionData(
	ctx context.Context, request admin.TaskExecutionGetDataRequest) (*admin.TaskExecutionGetDataResponse, error) {
	if m.getTaskExecutionDataFunc != nil {
		return m.getTaskExecutionDataFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockTaskExecutionManager) SetGetTaskExecutionDataCallback(
	getTaskExecutionDataFunc GetTaskExecutionDataFunc) {
	m.getTaskExecutionDataFunc = getTaskExecutionDataFunc
}
