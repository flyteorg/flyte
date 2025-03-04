// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	interfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	mock "github.com/stretchr/testify/mock"
)

// ClusterPoolAssignmentConfiguration is an autogenerated mock type for the ClusterPoolAssignmentConfiguration type
type ClusterPoolAssignmentConfiguration struct {
	mock.Mock
}

type ClusterPoolAssignmentConfiguration_Expecter struct {
	mock *mock.Mock
}

func (_m *ClusterPoolAssignmentConfiguration) EXPECT() *ClusterPoolAssignmentConfiguration_Expecter {
	return &ClusterPoolAssignmentConfiguration_Expecter{mock: &_m.Mock}
}

// GetClusterPoolAssignments provides a mock function with no fields
func (_m *ClusterPoolAssignmentConfiguration) GetClusterPoolAssignments() map[string]interfaces.ClusterPoolAssignment {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetClusterPoolAssignments")
	}

	var r0 map[string]interfaces.ClusterPoolAssignment
	if rf, ok := ret.Get(0).(func() map[string]interfaces.ClusterPoolAssignment); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interfaces.ClusterPoolAssignment)
		}
	}

	return r0
}

// ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetClusterPoolAssignments'
type ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments_Call struct {
	*mock.Call
}

// GetClusterPoolAssignments is a helper method to define mock.On call
func (_e *ClusterPoolAssignmentConfiguration_Expecter) GetClusterPoolAssignments() *ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments_Call {
	return &ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments_Call{Call: _e.mock.On("GetClusterPoolAssignments")}
}

func (_c *ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments_Call) Run(run func()) *ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments_Call) Return(_a0 map[string]interfaces.ClusterPoolAssignment) *ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments_Call) RunAndReturn(run func() map[string]interfaces.ClusterPoolAssignment) *ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments_Call {
	_c.Call.Return(run)
	return _c
}

// NewClusterPoolAssignmentConfiguration creates a new instance of ClusterPoolAssignmentConfiguration. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClusterPoolAssignmentConfiguration(t interface {
	mock.TestingT
	Cleanup(func())
}) *ClusterPoolAssignmentConfiguration {
	mock := &ClusterPoolAssignmentConfiguration{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
