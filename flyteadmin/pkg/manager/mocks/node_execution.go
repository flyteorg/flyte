package mocks

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateNodeEventFunc func(ctx context.Context, request admin.NodeExecutionEventRequest) (
	*admin.NodeExecutionEventResponse, error)
type GetNodeExecutionFunc func(ctx context.Context, request admin.NodeExecutionGetRequest) (*admin.NodeExecution, error)
type ListNodeExecutionsFunc func(
	ctx context.Context, request admin.NodeExecutionListRequest) (*admin.NodeExecutionList, error)
type ListNodeExecutionsForTaskFunc func(ctx context.Context, request admin.NodeExecutionForTaskListRequest) (
	*admin.NodeExecutionList, error)
type GetNodeExecutionDataFunc func(
	ctx context.Context, request admin.NodeExecutionGetDataRequest) (*admin.NodeExecutionGetDataResponse, error)

type MockNodeExecutionManager struct {
	createNodeEventFunc           CreateNodeEventFunc
	getNodeExecutionFunc          GetNodeExecutionFunc
	listNodeExecutionsFunc        ListNodeExecutionsFunc
	listNodeExecutionsForTaskFunc ListNodeExecutionsForTaskFunc
	getNodeExecutionDataFunc      GetNodeExecutionDataFunc
}

func (m *MockNodeExecutionManager) SetCreateNodeEventCallback(createNodeEventFunc CreateNodeEventFunc) {
	m.createNodeEventFunc = createNodeEventFunc
}

func (m *MockNodeExecutionManager) CreateNodeEvent(
	ctx context.Context,
	request admin.NodeExecutionEventRequest) (*admin.NodeExecutionEventResponse, error) {
	if m.createNodeEventFunc != nil {
		return m.createNodeEventFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockNodeExecutionManager) SetGetNodeExecutionFunc(getNodeExecutionFunc GetNodeExecutionFunc) {
	m.getNodeExecutionFunc = getNodeExecutionFunc
}

func (m *MockNodeExecutionManager) GetNodeExecution(
	ctx context.Context, request admin.NodeExecutionGetRequest) (*admin.NodeExecution, error) {
	if m.getNodeExecutionFunc != nil {
		return m.getNodeExecutionFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockNodeExecutionManager) SetListNodeExecutionsFunc(listNodeExecutionsFunc ListNodeExecutionsFunc) {
	m.listNodeExecutionsFunc = listNodeExecutionsFunc
}

func (m *MockNodeExecutionManager) ListNodeExecutions(
	ctx context.Context, request admin.NodeExecutionListRequest) (*admin.NodeExecutionList, error) {
	if m.listNodeExecutionsFunc != nil {
		return m.listNodeExecutionsFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockNodeExecutionManager) SetListNodeExecutionsForTaskFunc(listNodeExecutionsForTaskFunc ListNodeExecutionsForTaskFunc) {
	m.listNodeExecutionsForTaskFunc = listNodeExecutionsForTaskFunc
}

func (m *MockNodeExecutionManager) ListNodeExecutionsForTask(
	ctx context.Context, request admin.NodeExecutionForTaskListRequest) (*admin.NodeExecutionList, error) {
	if m.listNodeExecutionsForTaskFunc != nil {
		return m.listNodeExecutionsForTaskFunc(ctx, request)
	}
	return nil, nil
}

func (m *MockNodeExecutionManager) SetGetNodeExecutionDataFunc(getNodeExecutionDataFunc GetNodeExecutionDataFunc) {
	m.getNodeExecutionDataFunc = getNodeExecutionDataFunc
}

func (m *MockNodeExecutionManager) GetNodeExecutionData(
	ctx context.Context, request admin.NodeExecutionGetDataRequest) (*admin.NodeExecutionGetDataResponse, error) {
	if m.getNodeExecutionDataFunc != nil {
		return m.getNodeExecutionDataFunc(ctx, request)
	}
	return nil, nil
}
