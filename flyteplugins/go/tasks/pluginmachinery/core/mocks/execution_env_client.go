// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	core "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	mock "github.com/stretchr/testify/mock"

	structpb "google.golang.org/protobuf/types/known/structpb"
)

// ExecutionEnvClient is an autogenerated mock type for the ExecutionEnvClient type
type ExecutionEnvClient struct {
	mock.Mock
}

type ExecutionEnvClient_Create struct {
	*mock.Call
}

func (_m ExecutionEnvClient_Create) Return(_a0 *structpb.Struct, _a1 error) *ExecutionEnvClient_Create {
	return &ExecutionEnvClient_Create{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ExecutionEnvClient) OnCreate(ctx context.Context, executionEnvID core.ExecutionEnvID, executionEnvSpec *structpb.Struct) *ExecutionEnvClient_Create {
	c_call := _m.On("Create", ctx, executionEnvID, executionEnvSpec)
	return &ExecutionEnvClient_Create{Call: c_call}
}

func (_m *ExecutionEnvClient) OnCreateMatch(matchers ...interface{}) *ExecutionEnvClient_Create {
	c_call := _m.On("Create", matchers...)
	return &ExecutionEnvClient_Create{Call: c_call}
}

// Create provides a mock function with given fields: ctx, executionEnvID, executionEnvSpec
func (_m *ExecutionEnvClient) Create(ctx context.Context, executionEnvID core.ExecutionEnvID, executionEnvSpec *structpb.Struct) (*structpb.Struct, error) {
	ret := _m.Called(ctx, executionEnvID, executionEnvSpec)

	var r0 *structpb.Struct
	if rf, ok := ret.Get(0).(func(context.Context, core.ExecutionEnvID, *structpb.Struct) *structpb.Struct); ok {
		r0 = rf(ctx, executionEnvID, executionEnvSpec)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*structpb.Struct)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, core.ExecutionEnvID, *structpb.Struct) error); ok {
		r1 = rf(ctx, executionEnvID, executionEnvSpec)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ExecutionEnvClient_Get struct {
	*mock.Call
}

func (_m ExecutionEnvClient_Get) Return(_a0 *structpb.Struct) *ExecutionEnvClient_Get {
	return &ExecutionEnvClient_Get{Call: _m.Call.Return(_a0)}
}

func (_m *ExecutionEnvClient) OnGet(ctx context.Context, executionEnvID core.ExecutionEnvID) *ExecutionEnvClient_Get {
	c_call := _m.On("Get", ctx, executionEnvID)
	return &ExecutionEnvClient_Get{Call: c_call}
}

func (_m *ExecutionEnvClient) OnGetMatch(matchers ...interface{}) *ExecutionEnvClient_Get {
	c_call := _m.On("Get", matchers...)
	return &ExecutionEnvClient_Get{Call: c_call}
}

// Get provides a mock function with given fields: ctx, executionEnvID
func (_m *ExecutionEnvClient) Get(ctx context.Context, executionEnvID core.ExecutionEnvID) *structpb.Struct {
	ret := _m.Called(ctx, executionEnvID)

	var r0 *structpb.Struct
	if rf, ok := ret.Get(0).(func(context.Context, core.ExecutionEnvID) *structpb.Struct); ok {
		r0 = rf(ctx, executionEnvID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*structpb.Struct)
		}
	}

	return r0
}

type ExecutionEnvClient_Status struct {
	*mock.Call
}

func (_m ExecutionEnvClient_Status) Return(_a0 interface{}, _a1 error) *ExecutionEnvClient_Status {
	return &ExecutionEnvClient_Status{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ExecutionEnvClient) OnStatus(ctx context.Context, executionEnvID core.ExecutionEnvID) *ExecutionEnvClient_Status {
	c_call := _m.On("Status", ctx, executionEnvID)
	return &ExecutionEnvClient_Status{Call: c_call}
}

func (_m *ExecutionEnvClient) OnStatusMatch(matchers ...interface{}) *ExecutionEnvClient_Status {
	c_call := _m.On("Status", matchers...)
	return &ExecutionEnvClient_Status{Call: c_call}
}

// Status provides a mock function with given fields: ctx, executionEnvID
func (_m *ExecutionEnvClient) Status(ctx context.Context, executionEnvID core.ExecutionEnvID) (interface{}, error) {
	ret := _m.Called(ctx, executionEnvID)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(context.Context, core.ExecutionEnvID) interface{}); ok {
		r0 = rf(ctx, executionEnvID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, core.ExecutionEnvID) error); ok {
		r1 = rf(ctx, executionEnvID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
