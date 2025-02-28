// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	mock "github.com/stretchr/testify/mock"
)

// WorkflowInterface is an autogenerated mock type for the WorkflowInterface type
type WorkflowInterface struct {
	mock.Mock
}

type WorkflowInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *WorkflowInterface) EXPECT() *WorkflowInterface_Expecter {
	return &WorkflowInterface_Expecter{mock: &_m.Mock}
}

// CreateWorkflow provides a mock function with given fields: ctx, request
func (_m *WorkflowInterface) CreateWorkflow(ctx context.Context, request *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for CreateWorkflow")
	}

	var r0 *admin.WorkflowCreateResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.WorkflowCreateRequest) *admin.WorkflowCreateResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.WorkflowCreateResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.WorkflowCreateRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WorkflowInterface_CreateWorkflow_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateWorkflow'
type WorkflowInterface_CreateWorkflow_Call struct {
	*mock.Call
}

// CreateWorkflow is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.WorkflowCreateRequest
func (_e *WorkflowInterface_Expecter) CreateWorkflow(ctx interface{}, request interface{}) *WorkflowInterface_CreateWorkflow_Call {
	return &WorkflowInterface_CreateWorkflow_Call{Call: _e.mock.On("CreateWorkflow", ctx, request)}
}

func (_c *WorkflowInterface_CreateWorkflow_Call) Run(run func(ctx context.Context, request *admin.WorkflowCreateRequest)) *WorkflowInterface_CreateWorkflow_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.WorkflowCreateRequest))
	})
	return _c
}

func (_c *WorkflowInterface_CreateWorkflow_Call) Return(_a0 *admin.WorkflowCreateResponse, _a1 error) *WorkflowInterface_CreateWorkflow_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *WorkflowInterface_CreateWorkflow_Call) RunAndReturn(run func(context.Context, *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error)) *WorkflowInterface_CreateWorkflow_Call {
	_c.Call.Return(run)
	return _c
}

// GetWorkflow provides a mock function with given fields: ctx, request
func (_m *WorkflowInterface) GetWorkflow(ctx context.Context, request *admin.ObjectGetRequest) (*admin.Workflow, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for GetWorkflow")
	}

	var r0 *admin.Workflow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.ObjectGetRequest) (*admin.Workflow, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.ObjectGetRequest) *admin.Workflow); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.Workflow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.ObjectGetRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WorkflowInterface_GetWorkflow_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetWorkflow'
type WorkflowInterface_GetWorkflow_Call struct {
	*mock.Call
}

// GetWorkflow is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.ObjectGetRequest
func (_e *WorkflowInterface_Expecter) GetWorkflow(ctx interface{}, request interface{}) *WorkflowInterface_GetWorkflow_Call {
	return &WorkflowInterface_GetWorkflow_Call{Call: _e.mock.On("GetWorkflow", ctx, request)}
}

func (_c *WorkflowInterface_GetWorkflow_Call) Run(run func(ctx context.Context, request *admin.ObjectGetRequest)) *WorkflowInterface_GetWorkflow_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.ObjectGetRequest))
	})
	return _c
}

func (_c *WorkflowInterface_GetWorkflow_Call) Return(_a0 *admin.Workflow, _a1 error) *WorkflowInterface_GetWorkflow_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *WorkflowInterface_GetWorkflow_Call) RunAndReturn(run func(context.Context, *admin.ObjectGetRequest) (*admin.Workflow, error)) *WorkflowInterface_GetWorkflow_Call {
	_c.Call.Return(run)
	return _c
}

// ListWorkflowIdentifiers provides a mock function with given fields: ctx, request
func (_m *WorkflowInterface) ListWorkflowIdentifiers(ctx context.Context, request *admin.NamedEntityIdentifierListRequest) (*admin.NamedEntityIdentifierList, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for ListWorkflowIdentifiers")
	}

	var r0 *admin.NamedEntityIdentifierList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.NamedEntityIdentifierListRequest) (*admin.NamedEntityIdentifierList, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.NamedEntityIdentifierListRequest) *admin.NamedEntityIdentifierList); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.NamedEntityIdentifierList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.NamedEntityIdentifierListRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WorkflowInterface_ListWorkflowIdentifiers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListWorkflowIdentifiers'
type WorkflowInterface_ListWorkflowIdentifiers_Call struct {
	*mock.Call
}

// ListWorkflowIdentifiers is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.NamedEntityIdentifierListRequest
func (_e *WorkflowInterface_Expecter) ListWorkflowIdentifiers(ctx interface{}, request interface{}) *WorkflowInterface_ListWorkflowIdentifiers_Call {
	return &WorkflowInterface_ListWorkflowIdentifiers_Call{Call: _e.mock.On("ListWorkflowIdentifiers", ctx, request)}
}

func (_c *WorkflowInterface_ListWorkflowIdentifiers_Call) Run(run func(ctx context.Context, request *admin.NamedEntityIdentifierListRequest)) *WorkflowInterface_ListWorkflowIdentifiers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.NamedEntityIdentifierListRequest))
	})
	return _c
}

func (_c *WorkflowInterface_ListWorkflowIdentifiers_Call) Return(_a0 *admin.NamedEntityIdentifierList, _a1 error) *WorkflowInterface_ListWorkflowIdentifiers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *WorkflowInterface_ListWorkflowIdentifiers_Call) RunAndReturn(run func(context.Context, *admin.NamedEntityIdentifierListRequest) (*admin.NamedEntityIdentifierList, error)) *WorkflowInterface_ListWorkflowIdentifiers_Call {
	_c.Call.Return(run)
	return _c
}

// ListWorkflows provides a mock function with given fields: ctx, request
func (_m *WorkflowInterface) ListWorkflows(ctx context.Context, request *admin.ResourceListRequest) (*admin.WorkflowList, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for ListWorkflows")
	}

	var r0 *admin.WorkflowList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.ResourceListRequest) (*admin.WorkflowList, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.ResourceListRequest) *admin.WorkflowList); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.WorkflowList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.ResourceListRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WorkflowInterface_ListWorkflows_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListWorkflows'
type WorkflowInterface_ListWorkflows_Call struct {
	*mock.Call
}

// ListWorkflows is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.ResourceListRequest
func (_e *WorkflowInterface_Expecter) ListWorkflows(ctx interface{}, request interface{}) *WorkflowInterface_ListWorkflows_Call {
	return &WorkflowInterface_ListWorkflows_Call{Call: _e.mock.On("ListWorkflows", ctx, request)}
}

func (_c *WorkflowInterface_ListWorkflows_Call) Run(run func(ctx context.Context, request *admin.ResourceListRequest)) *WorkflowInterface_ListWorkflows_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.ResourceListRequest))
	})
	return _c
}

func (_c *WorkflowInterface_ListWorkflows_Call) Return(_a0 *admin.WorkflowList, _a1 error) *WorkflowInterface_ListWorkflows_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *WorkflowInterface_ListWorkflows_Call) RunAndReturn(run func(context.Context, *admin.ResourceListRequest) (*admin.WorkflowList, error)) *WorkflowInterface_ListWorkflows_Call {
	_c.Call.Return(run)
	return _c
}

// NewWorkflowInterface creates a new instance of WorkflowInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWorkflowInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *WorkflowInterface {
	mock := &WorkflowInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
