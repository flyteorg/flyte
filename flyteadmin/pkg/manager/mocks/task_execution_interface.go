// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	mock "github.com/stretchr/testify/mock"
)

// TaskExecutionInterface is an autogenerated mock type for the TaskExecutionInterface type
type TaskExecutionInterface struct {
	mock.Mock
}

type TaskExecutionInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *TaskExecutionInterface) EXPECT() *TaskExecutionInterface_Expecter {
	return &TaskExecutionInterface_Expecter{mock: &_m.Mock}
}

// CreateTaskExecutionEvent provides a mock function with given fields: ctx, request
func (_m *TaskExecutionInterface) CreateTaskExecutionEvent(ctx context.Context, request *admin.TaskExecutionEventRequest) (*admin.TaskExecutionEventResponse, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for CreateTaskExecutionEvent")
	}

	var r0 *admin.TaskExecutionEventResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.TaskExecutionEventRequest) (*admin.TaskExecutionEventResponse, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.TaskExecutionEventRequest) *admin.TaskExecutionEventResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.TaskExecutionEventResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.TaskExecutionEventRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TaskExecutionInterface_CreateTaskExecutionEvent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateTaskExecutionEvent'
type TaskExecutionInterface_CreateTaskExecutionEvent_Call struct {
	*mock.Call
}

// CreateTaskExecutionEvent is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.TaskExecutionEventRequest
func (_e *TaskExecutionInterface_Expecter) CreateTaskExecutionEvent(ctx interface{}, request interface{}) *TaskExecutionInterface_CreateTaskExecutionEvent_Call {
	return &TaskExecutionInterface_CreateTaskExecutionEvent_Call{Call: _e.mock.On("CreateTaskExecutionEvent", ctx, request)}
}

func (_c *TaskExecutionInterface_CreateTaskExecutionEvent_Call) Run(run func(ctx context.Context, request *admin.TaskExecutionEventRequest)) *TaskExecutionInterface_CreateTaskExecutionEvent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.TaskExecutionEventRequest))
	})
	return _c
}

func (_c *TaskExecutionInterface_CreateTaskExecutionEvent_Call) Return(_a0 *admin.TaskExecutionEventResponse, _a1 error) *TaskExecutionInterface_CreateTaskExecutionEvent_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TaskExecutionInterface_CreateTaskExecutionEvent_Call) RunAndReturn(run func(context.Context, *admin.TaskExecutionEventRequest) (*admin.TaskExecutionEventResponse, error)) *TaskExecutionInterface_CreateTaskExecutionEvent_Call {
	_c.Call.Return(run)
	return _c
}

// GetTaskExecution provides a mock function with given fields: ctx, request
func (_m *TaskExecutionInterface) GetTaskExecution(ctx context.Context, request *admin.TaskExecutionGetRequest) (*admin.TaskExecution, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for GetTaskExecution")
	}

	var r0 *admin.TaskExecution
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.TaskExecutionGetRequest) (*admin.TaskExecution, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.TaskExecutionGetRequest) *admin.TaskExecution); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.TaskExecution)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.TaskExecutionGetRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TaskExecutionInterface_GetTaskExecution_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTaskExecution'
type TaskExecutionInterface_GetTaskExecution_Call struct {
	*mock.Call
}

// GetTaskExecution is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.TaskExecutionGetRequest
func (_e *TaskExecutionInterface_Expecter) GetTaskExecution(ctx interface{}, request interface{}) *TaskExecutionInterface_GetTaskExecution_Call {
	return &TaskExecutionInterface_GetTaskExecution_Call{Call: _e.mock.On("GetTaskExecution", ctx, request)}
}

func (_c *TaskExecutionInterface_GetTaskExecution_Call) Run(run func(ctx context.Context, request *admin.TaskExecutionGetRequest)) *TaskExecutionInterface_GetTaskExecution_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.TaskExecutionGetRequest))
	})
	return _c
}

func (_c *TaskExecutionInterface_GetTaskExecution_Call) Return(_a0 *admin.TaskExecution, _a1 error) *TaskExecutionInterface_GetTaskExecution_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TaskExecutionInterface_GetTaskExecution_Call) RunAndReturn(run func(context.Context, *admin.TaskExecutionGetRequest) (*admin.TaskExecution, error)) *TaskExecutionInterface_GetTaskExecution_Call {
	_c.Call.Return(run)
	return _c
}

// GetTaskExecutionData provides a mock function with given fields: ctx, request
func (_m *TaskExecutionInterface) GetTaskExecutionData(ctx context.Context, request *admin.TaskExecutionGetDataRequest) (*admin.TaskExecutionGetDataResponse, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for GetTaskExecutionData")
	}

	var r0 *admin.TaskExecutionGetDataResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.TaskExecutionGetDataRequest) (*admin.TaskExecutionGetDataResponse, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.TaskExecutionGetDataRequest) *admin.TaskExecutionGetDataResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.TaskExecutionGetDataResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.TaskExecutionGetDataRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TaskExecutionInterface_GetTaskExecutionData_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTaskExecutionData'
type TaskExecutionInterface_GetTaskExecutionData_Call struct {
	*mock.Call
}

// GetTaskExecutionData is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.TaskExecutionGetDataRequest
func (_e *TaskExecutionInterface_Expecter) GetTaskExecutionData(ctx interface{}, request interface{}) *TaskExecutionInterface_GetTaskExecutionData_Call {
	return &TaskExecutionInterface_GetTaskExecutionData_Call{Call: _e.mock.On("GetTaskExecutionData", ctx, request)}
}

func (_c *TaskExecutionInterface_GetTaskExecutionData_Call) Run(run func(ctx context.Context, request *admin.TaskExecutionGetDataRequest)) *TaskExecutionInterface_GetTaskExecutionData_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.TaskExecutionGetDataRequest))
	})
	return _c
}

func (_c *TaskExecutionInterface_GetTaskExecutionData_Call) Return(_a0 *admin.TaskExecutionGetDataResponse, _a1 error) *TaskExecutionInterface_GetTaskExecutionData_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TaskExecutionInterface_GetTaskExecutionData_Call) RunAndReturn(run func(context.Context, *admin.TaskExecutionGetDataRequest) (*admin.TaskExecutionGetDataResponse, error)) *TaskExecutionInterface_GetTaskExecutionData_Call {
	_c.Call.Return(run)
	return _c
}

// ListTaskExecutions provides a mock function with given fields: ctx, request
func (_m *TaskExecutionInterface) ListTaskExecutions(ctx context.Context, request *admin.TaskExecutionListRequest) (*admin.TaskExecutionList, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for ListTaskExecutions")
	}

	var r0 *admin.TaskExecutionList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.TaskExecutionListRequest) (*admin.TaskExecutionList, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.TaskExecutionListRequest) *admin.TaskExecutionList); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.TaskExecutionList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.TaskExecutionListRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TaskExecutionInterface_ListTaskExecutions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListTaskExecutions'
type TaskExecutionInterface_ListTaskExecutions_Call struct {
	*mock.Call
}

// ListTaskExecutions is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.TaskExecutionListRequest
func (_e *TaskExecutionInterface_Expecter) ListTaskExecutions(ctx interface{}, request interface{}) *TaskExecutionInterface_ListTaskExecutions_Call {
	return &TaskExecutionInterface_ListTaskExecutions_Call{Call: _e.mock.On("ListTaskExecutions", ctx, request)}
}

func (_c *TaskExecutionInterface_ListTaskExecutions_Call) Run(run func(ctx context.Context, request *admin.TaskExecutionListRequest)) *TaskExecutionInterface_ListTaskExecutions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.TaskExecutionListRequest))
	})
	return _c
}

func (_c *TaskExecutionInterface_ListTaskExecutions_Call) Return(_a0 *admin.TaskExecutionList, _a1 error) *TaskExecutionInterface_ListTaskExecutions_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TaskExecutionInterface_ListTaskExecutions_Call) RunAndReturn(run func(context.Context, *admin.TaskExecutionListRequest) (*admin.TaskExecutionList, error)) *TaskExecutionInterface_ListTaskExecutions_Call {
	_c.Call.Return(run)
	return _c
}

// NewTaskExecutionInterface creates a new instance of TaskExecutionInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTaskExecutionInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *TaskExecutionInterface {
	mock := &TaskExecutionInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
