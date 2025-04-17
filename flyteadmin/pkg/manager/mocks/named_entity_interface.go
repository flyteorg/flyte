// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	mock "github.com/stretchr/testify/mock"
)

// NamedEntityInterface is an autogenerated mock type for the NamedEntityInterface type
type NamedEntityInterface struct {
	mock.Mock
}

type NamedEntityInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *NamedEntityInterface) EXPECT() *NamedEntityInterface_Expecter {
	return &NamedEntityInterface_Expecter{mock: &_m.Mock}
}

// GetNamedEntity provides a mock function with given fields: ctx, request
func (_m *NamedEntityInterface) GetNamedEntity(ctx context.Context, request *admin.NamedEntityGetRequest) (*admin.NamedEntity, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for GetNamedEntity")
	}

	var r0 *admin.NamedEntity
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.NamedEntityGetRequest) (*admin.NamedEntity, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.NamedEntityGetRequest) *admin.NamedEntity); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.NamedEntity)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.NamedEntityGetRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NamedEntityInterface_GetNamedEntity_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNamedEntity'
type NamedEntityInterface_GetNamedEntity_Call struct {
	*mock.Call
}

// GetNamedEntity is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.NamedEntityGetRequest
func (_e *NamedEntityInterface_Expecter) GetNamedEntity(ctx interface{}, request interface{}) *NamedEntityInterface_GetNamedEntity_Call {
	return &NamedEntityInterface_GetNamedEntity_Call{Call: _e.mock.On("GetNamedEntity", ctx, request)}
}

func (_c *NamedEntityInterface_GetNamedEntity_Call) Run(run func(ctx context.Context, request *admin.NamedEntityGetRequest)) *NamedEntityInterface_GetNamedEntity_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.NamedEntityGetRequest))
	})
	return _c
}

func (_c *NamedEntityInterface_GetNamedEntity_Call) Return(_a0 *admin.NamedEntity, _a1 error) *NamedEntityInterface_GetNamedEntity_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NamedEntityInterface_GetNamedEntity_Call) RunAndReturn(run func(context.Context, *admin.NamedEntityGetRequest) (*admin.NamedEntity, error)) *NamedEntityInterface_GetNamedEntity_Call {
	_c.Call.Return(run)
	return _c
}

// ListNamedEntities provides a mock function with given fields: ctx, request
func (_m *NamedEntityInterface) ListNamedEntities(ctx context.Context, request *admin.NamedEntityListRequest) (*admin.NamedEntityList, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for ListNamedEntities")
	}

	var r0 *admin.NamedEntityList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.NamedEntityListRequest) (*admin.NamedEntityList, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.NamedEntityListRequest) *admin.NamedEntityList); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.NamedEntityList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.NamedEntityListRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NamedEntityInterface_ListNamedEntities_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListNamedEntities'
type NamedEntityInterface_ListNamedEntities_Call struct {
	*mock.Call
}

// ListNamedEntities is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.NamedEntityListRequest
func (_e *NamedEntityInterface_Expecter) ListNamedEntities(ctx interface{}, request interface{}) *NamedEntityInterface_ListNamedEntities_Call {
	return &NamedEntityInterface_ListNamedEntities_Call{Call: _e.mock.On("ListNamedEntities", ctx, request)}
}

func (_c *NamedEntityInterface_ListNamedEntities_Call) Run(run func(ctx context.Context, request *admin.NamedEntityListRequest)) *NamedEntityInterface_ListNamedEntities_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.NamedEntityListRequest))
	})
	return _c
}

func (_c *NamedEntityInterface_ListNamedEntities_Call) Return(_a0 *admin.NamedEntityList, _a1 error) *NamedEntityInterface_ListNamedEntities_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NamedEntityInterface_ListNamedEntities_Call) RunAndReturn(run func(context.Context, *admin.NamedEntityListRequest) (*admin.NamedEntityList, error)) *NamedEntityInterface_ListNamedEntities_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateNamedEntity provides a mock function with given fields: ctx, request
func (_m *NamedEntityInterface) UpdateNamedEntity(ctx context.Context, request *admin.NamedEntityUpdateRequest) (*admin.NamedEntityUpdateResponse, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for UpdateNamedEntity")
	}

	var r0 *admin.NamedEntityUpdateResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.NamedEntityUpdateRequest) (*admin.NamedEntityUpdateResponse, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.NamedEntityUpdateRequest) *admin.NamedEntityUpdateResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.NamedEntityUpdateResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.NamedEntityUpdateRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NamedEntityInterface_UpdateNamedEntity_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateNamedEntity'
type NamedEntityInterface_UpdateNamedEntity_Call struct {
	*mock.Call
}

// UpdateNamedEntity is a helper method to define mock.On call
//   - ctx context.Context
//   - request *admin.NamedEntityUpdateRequest
func (_e *NamedEntityInterface_Expecter) UpdateNamedEntity(ctx interface{}, request interface{}) *NamedEntityInterface_UpdateNamedEntity_Call {
	return &NamedEntityInterface_UpdateNamedEntity_Call{Call: _e.mock.On("UpdateNamedEntity", ctx, request)}
}

func (_c *NamedEntityInterface_UpdateNamedEntity_Call) Run(run func(ctx context.Context, request *admin.NamedEntityUpdateRequest)) *NamedEntityInterface_UpdateNamedEntity_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*admin.NamedEntityUpdateRequest))
	})
	return _c
}

func (_c *NamedEntityInterface_UpdateNamedEntity_Call) Return(_a0 *admin.NamedEntityUpdateResponse, _a1 error) *NamedEntityInterface_UpdateNamedEntity_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *NamedEntityInterface_UpdateNamedEntity_Call) RunAndReturn(run func(context.Context, *admin.NamedEntityUpdateRequest) (*admin.NamedEntityUpdateResponse, error)) *NamedEntityInterface_UpdateNamedEntity_Call {
	_c.Call.Return(run)
	return _c
}

// NewNamedEntityInterface creates a new instance of NamedEntityInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNamedEntityInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *NamedEntityInterface {
	mock := &NamedEntityInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
