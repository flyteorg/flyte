// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/api/core/v1"
)

// ShardStrategy is an autogenerated mock type for the ShardStrategy type
type ShardStrategy struct {
	mock.Mock
}

type ShardStrategy_Expecter struct {
	mock *mock.Mock
}

func (_m *ShardStrategy) EXPECT() *ShardStrategy_Expecter {
	return &ShardStrategy_Expecter{mock: &_m.Mock}
}

// GetPodCount provides a mock function with no fields
func (_m *ShardStrategy) GetPodCount() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetPodCount")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// ShardStrategy_GetPodCount_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPodCount'
type ShardStrategy_GetPodCount_Call struct {
	*mock.Call
}

// GetPodCount is a helper method to define mock.On call
func (_e *ShardStrategy_Expecter) GetPodCount() *ShardStrategy_GetPodCount_Call {
	return &ShardStrategy_GetPodCount_Call{Call: _e.mock.On("GetPodCount")}
}

func (_c *ShardStrategy_GetPodCount_Call) Run(run func()) *ShardStrategy_GetPodCount_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ShardStrategy_GetPodCount_Call) Return(_a0 int) *ShardStrategy_GetPodCount_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ShardStrategy_GetPodCount_Call) RunAndReturn(run func() int) *ShardStrategy_GetPodCount_Call {
	_c.Call.Return(run)
	return _c
}

// HashCode provides a mock function with no fields
func (_m *ShardStrategy) HashCode() (uint32, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for HashCode")
	}

	var r0 uint32
	var r1 error
	if rf, ok := ret.Get(0).(func() (uint32, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() uint32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint32)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ShardStrategy_HashCode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HashCode'
type ShardStrategy_HashCode_Call struct {
	*mock.Call
}

// HashCode is a helper method to define mock.On call
func (_e *ShardStrategy_Expecter) HashCode() *ShardStrategy_HashCode_Call {
	return &ShardStrategy_HashCode_Call{Call: _e.mock.On("HashCode")}
}

func (_c *ShardStrategy_HashCode_Call) Run(run func()) *ShardStrategy_HashCode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ShardStrategy_HashCode_Call) Return(_a0 uint32, _a1 error) *ShardStrategy_HashCode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ShardStrategy_HashCode_Call) RunAndReturn(run func() (uint32, error)) *ShardStrategy_HashCode_Call {
	_c.Call.Return(run)
	return _c
}

// UpdatePodSpec provides a mock function with given fields: pod, containerName, podIndex
func (_m *ShardStrategy) UpdatePodSpec(pod *v1.PodSpec, containerName string, podIndex int) error {
	ret := _m.Called(pod, containerName, podIndex)

	if len(ret) == 0 {
		panic("no return value specified for UpdatePodSpec")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1.PodSpec, string, int) error); ok {
		r0 = rf(pod, containerName, podIndex)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ShardStrategy_UpdatePodSpec_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdatePodSpec'
type ShardStrategy_UpdatePodSpec_Call struct {
	*mock.Call
}

// UpdatePodSpec is a helper method to define mock.On call
//   - pod *v1.PodSpec
//   - containerName string
//   - podIndex int
func (_e *ShardStrategy_Expecter) UpdatePodSpec(pod interface{}, containerName interface{}, podIndex interface{}) *ShardStrategy_UpdatePodSpec_Call {
	return &ShardStrategy_UpdatePodSpec_Call{Call: _e.mock.On("UpdatePodSpec", pod, containerName, podIndex)}
}

func (_c *ShardStrategy_UpdatePodSpec_Call) Run(run func(pod *v1.PodSpec, containerName string, podIndex int)) *ShardStrategy_UpdatePodSpec_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*v1.PodSpec), args[1].(string), args[2].(int))
	})
	return _c
}

func (_c *ShardStrategy_UpdatePodSpec_Call) Return(_a0 error) *ShardStrategy_UpdatePodSpec_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ShardStrategy_UpdatePodSpec_Call) RunAndReturn(run func(*v1.PodSpec, string, int) error) *ShardStrategy_UpdatePodSpec_Call {
	_c.Call.Return(run)
	return _c
}

// NewShardStrategy creates a new instance of ShardStrategy. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewShardStrategy(t interface {
	mock.TestingT
	Cleanup(func())
}) *ShardStrategy {
	mock := &ShardStrategy{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
