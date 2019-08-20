package mocks

import (
	"context"

	"github.com/lyft/flyteplugins/go/tasks/v1/resourcemanager"
	"github.com/stretchr/testify/mock"
)

type AllocateResourceCall struct {
	*mock.Call
}

func (_a *AllocateResourceCall) Return(status resourcemanager.AllocationStatus, err error) *AllocateResourceCall {
	return &AllocateResourceCall{Call: _a.Call.Return(status, err)}
}

func (_m *ResourceManager) OnAllocateResource(ctx context.Context, namespace string, allocationToken string) *AllocateResourceCall {
	call := _m.On("AllocateResource", ctx, namespace, allocationToken)
	return &AllocateResourceCall{Call: call}
}

func (_m *ResourceManager) OnAllocateResourceWithMatchers(matchers []interface{}) *AllocateResourceCall {
	call := _m.On("AllocateResource", matchers...)
	return &AllocateResourceCall{Call: call}
}

type ReleaseResourceCall struct {
	*mock.Call
}

func (_a *ReleaseResourceCall) Return(err error) *ReleaseResourceCall {
	return &ReleaseResourceCall{Call: _a.Call.Return(err)}
}

func (_m *ResourceManager) OnReleaseResource(ctx context.Context, namespace string, allocationToken string) *ReleaseResourceCall {
	call := _m.On("ReleaseResource", ctx, namespace, allocationToken)
	return &ReleaseResourceCall{Call: call}
}

func (_m *ResourceManager) OnReleaseResourceWithMatchers(matchers []interface{}) *ReleaseResourceCall {
	call := _m.On("ReleaseResource", matchers...)
	return &ReleaseResourceCall{Call: call}
}
