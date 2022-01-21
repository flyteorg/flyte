package mocks

import (
	"context"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateExecutionFunc func(
	ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error)
type RelaunchExecutionFunc func(
	ctx context.Context, request admin.ExecutionRelaunchRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error)
type RecoverExecutionFunc func(ctx context.Context, request admin.ExecutionRecoverRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error)
type CreateExecutionEventFunc func(ctx context.Context, request admin.WorkflowExecutionEventRequest) (
	*admin.WorkflowExecutionEventResponse, error)
type GetExecutionFunc func(ctx context.Context, request admin.WorkflowExecutionGetRequest) (*admin.Execution, error)
type UpdateExecutionFunc func(ctx context.Context, request admin.ExecutionUpdateRequest, requestedAt time.Time) (
	*admin.ExecutionUpdateResponse, error)
type GetExecutionDataFunc func(ctx context.Context, request admin.WorkflowExecutionGetDataRequest) (
	*admin.WorkflowExecutionGetDataResponse, error)
type ListExecutionFunc func(ctx context.Context, request admin.ResourceListRequest) (*admin.ExecutionList, error)
type TerminateExecutionFunc func(
	ctx context.Context, request admin.ExecutionTerminateRequest) (*admin.ExecutionTerminateResponse, error)

type MockExecutionManager struct {
	createExecutionFunc      CreateExecutionFunc
	relaunchExecutionFunc    RelaunchExecutionFunc
	RecoverExecutionFunc     RecoverExecutionFunc
	createExecutionEventFunc CreateExecutionEventFunc
	getExecutionFunc         GetExecutionFunc
	updateExecutionFunc      UpdateExecutionFunc
	getExecutionDataFunc     GetExecutionDataFunc
	listExecutionFunc        ListExecutionFunc
	terminateExecutionFunc   TerminateExecutionFunc
}

func (m *MockExecutionManager) SetCreateCallback(createFunction CreateExecutionFunc) {
	m.createExecutionFunc = createFunction
}

func (m *MockExecutionManager) CreateExecution(
	ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {
	if m.createExecutionFunc != nil {
		return m.createExecutionFunc(ctx, request, requestedAt)
	}
	return nil, nil
}

func (m *MockExecutionManager) SetRelaunchCallback(relaunchFunction RelaunchExecutionFunc) {
	m.relaunchExecutionFunc = relaunchFunction
}

func (m *MockExecutionManager) RelaunchExecution(
	ctx context.Context, request admin.ExecutionRelaunchRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {
	if m.relaunchExecutionFunc != nil {
		return m.relaunchExecutionFunc(ctx, request, requestedAt)
	}
	return nil, nil
}

func (m *MockExecutionManager) SetCreateEventCallback(createEventFunc CreateExecutionEventFunc) {
	m.createExecutionEventFunc = createEventFunc
}

func (m *MockExecutionManager) RecoverExecution(ctx context.Context, request admin.ExecutionRecoverRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {
	if m.RecoverExecutionFunc != nil {
		return m.RecoverExecutionFunc(ctx, request, requestedAt)
	}
	return &admin.ExecutionCreateResponse{}, nil
}

func (m *MockExecutionManager) CreateWorkflowEvent(
	ctx context.Context,
	request admin.WorkflowExecutionEventRequest) (*admin.WorkflowExecutionEventResponse, error) {
	if m.createExecutionEventFunc != nil {
		return m.createExecutionEventFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockExecutionManager) SetUpdateExecutionCallback(updateExecutionFunc UpdateExecutionFunc) {
	m.updateExecutionFunc = updateExecutionFunc
}

func (m *MockExecutionManager) UpdateExecution(ctx context.Context, request admin.ExecutionUpdateRequest,
	requestedAt time.Time) (*admin.ExecutionUpdateResponse, error) {
	if m.updateExecutionFunc != nil {
		return m.updateExecutionFunc(ctx, request, requestedAt)
	}
	return nil, nil
}

func (m *MockExecutionManager) SetGetCallback(getExecutionFunc GetExecutionFunc) {
	m.getExecutionFunc = getExecutionFunc
}

func (m *MockExecutionManager) GetExecution(
	ctx context.Context, request admin.WorkflowExecutionGetRequest) (*admin.Execution, error) {
	if m.getExecutionFunc != nil {
		return m.getExecutionFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockExecutionManager) SetGetDataCallback(getExecutionDataFunc GetExecutionDataFunc) {
	m.getExecutionDataFunc = getExecutionDataFunc
}

func (m *MockExecutionManager) GetExecutionData(ctx context.Context, request admin.WorkflowExecutionGetDataRequest) (
	*admin.WorkflowExecutionGetDataResponse, error) {
	if m.getExecutionDataFunc != nil {
		return m.getExecutionDataFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockExecutionManager) SetListCallback(listExecutionFunc ListExecutionFunc) {
	m.listExecutionFunc = listExecutionFunc
}

func (m *MockExecutionManager) ListExecutions(
	ctx context.Context, request admin.ResourceListRequest) (*admin.ExecutionList, error) {
	if m.listExecutionFunc != nil {
		return m.listExecutionFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockExecutionManager) SetTerminateExecutionCallback(terminateExecutionFunc TerminateExecutionFunc) {
	m.terminateExecutionFunc = terminateExecutionFunc
}

func (m *MockExecutionManager) TerminateExecution(
	ctx context.Context, request admin.ExecutionTerminateRequest) (*admin.ExecutionTerminateResponse, error) {
	if m.terminateExecutionFunc != nil {
		return m.terminateExecutionFunc(ctx, request)
	}
	return nil, nil
}
