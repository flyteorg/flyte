// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	context "context"

	storage "github.com/flyteorg/flyte/flytestdlib/storage"
	mock "github.com/stretchr/testify/mock"
)

// ReferenceConstructor is an autogenerated mock type for the ReferenceConstructor type
type ReferenceConstructor struct {
	mock.Mock
}

type ReferenceConstructor_Expecter struct {
	mock *mock.Mock
}

func (_m *ReferenceConstructor) EXPECT() *ReferenceConstructor_Expecter {
	return &ReferenceConstructor_Expecter{mock: &_m.Mock}
}

// ConstructReference provides a mock function with given fields: ctx, reference, nestedKeys
func (_m *ReferenceConstructor) ConstructReference(ctx context.Context, reference storage.DataReference, nestedKeys ...string) (storage.DataReference, error) {
	_va := make([]interface{}, len(nestedKeys))
	for _i := range nestedKeys {
		_va[_i] = nestedKeys[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, reference)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ConstructReference")
	}

	var r0 storage.DataReference
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, storage.DataReference, ...string) (storage.DataReference, error)); ok {
		return rf(ctx, reference, nestedKeys...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, storage.DataReference, ...string) storage.DataReference); ok {
		r0 = rf(ctx, reference, nestedKeys...)
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	if rf, ok := ret.Get(1).(func(context.Context, storage.DataReference, ...string) error); ok {
		r1 = rf(ctx, reference, nestedKeys...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReferenceConstructor_ConstructReference_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ConstructReference'
type ReferenceConstructor_ConstructReference_Call struct {
	*mock.Call
}

// ConstructReference is a helper method to define mock.On call
//   - ctx context.Context
//   - reference storage.DataReference
//   - nestedKeys ...string
func (_e *ReferenceConstructor_Expecter) ConstructReference(ctx interface{}, reference interface{}, nestedKeys ...interface{}) *ReferenceConstructor_ConstructReference_Call {
	return &ReferenceConstructor_ConstructReference_Call{Call: _e.mock.On("ConstructReference",
		append([]interface{}{ctx, reference}, nestedKeys...)...)}
}

func (_c *ReferenceConstructor_ConstructReference_Call) Run(run func(ctx context.Context, reference storage.DataReference, nestedKeys ...string)) *ReferenceConstructor_ConstructReference_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), args[1].(storage.DataReference), variadicArgs...)
	})
	return _c
}

func (_c *ReferenceConstructor_ConstructReference_Call) Return(_a0 storage.DataReference, _a1 error) *ReferenceConstructor_ConstructReference_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ReferenceConstructor_ConstructReference_Call) RunAndReturn(run func(context.Context, storage.DataReference, ...string) (storage.DataReference, error)) *ReferenceConstructor_ConstructReference_Call {
	_c.Call.Return(run)
	return _c
}

// NewReferenceConstructor creates a new instance of ReferenceConstructor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewReferenceConstructor(t interface {
	mock.TestingT
	Cleanup(func())
}) *ReferenceConstructor {
	mock := &ReferenceConstructor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
