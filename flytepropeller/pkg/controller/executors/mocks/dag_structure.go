// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// DAGStructure is an autogenerated mock type for the DAGStructure type
type DAGStructure struct {
	mock.Mock
}

type DAGStructure_Expecter struct {
	mock *mock.Mock
}

func (_m *DAGStructure) EXPECT() *DAGStructure_Expecter {
	return &DAGStructure_Expecter{mock: &_m.Mock}
}

// FromNode provides a mock function with given fields: id
func (_m *DAGStructure) FromNode(id string) ([]string, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for FromNode")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DAGStructure_FromNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FromNode'
type DAGStructure_FromNode_Call struct {
	*mock.Call
}

// FromNode is a helper method to define mock.On call
//   - id string
func (_e *DAGStructure_Expecter) FromNode(id interface{}) *DAGStructure_FromNode_Call {
	return &DAGStructure_FromNode_Call{Call: _e.mock.On("FromNode", id)}
}

func (_c *DAGStructure_FromNode_Call) Run(run func(id string)) *DAGStructure_FromNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *DAGStructure_FromNode_Call) Return(_a0 []string, _a1 error) *DAGStructure_FromNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DAGStructure_FromNode_Call) RunAndReturn(run func(string) ([]string, error)) *DAGStructure_FromNode_Call {
	_c.Call.Return(run)
	return _c
}

// ToNode provides a mock function with given fields: id
func (_m *DAGStructure) ToNode(id string) ([]string, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for ToNode")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DAGStructure_ToNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ToNode'
type DAGStructure_ToNode_Call struct {
	*mock.Call
}

// ToNode is a helper method to define mock.On call
//   - id string
func (_e *DAGStructure_Expecter) ToNode(id interface{}) *DAGStructure_ToNode_Call {
	return &DAGStructure_ToNode_Call{Call: _e.mock.On("ToNode", id)}
}

func (_c *DAGStructure_ToNode_Call) Run(run func(id string)) *DAGStructure_ToNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *DAGStructure_ToNode_Call) Return(_a0 []string, _a1 error) *DAGStructure_ToNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DAGStructure_ToNode_Call) RunAndReturn(run func(string) ([]string, error)) *DAGStructure_ToNode_Call {
	_c.Call.Return(run)
	return _c
}

// NewDAGStructure creates a new instance of DAGStructure. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDAGStructure(t interface {
	mock.TestingT
	Cleanup(func())
}) *DAGStructure {
	mock := &DAGStructure{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
