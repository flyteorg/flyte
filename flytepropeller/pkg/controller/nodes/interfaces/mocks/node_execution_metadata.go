// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

	mock "github.com/stretchr/testify/mock"

	types "k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeExecutionMetadata is an autogenerated mock type for the NodeExecutionMetadata type
type NodeExecutionMetadata struct {
	mock.Mock
}

type NodeExecutionMetadata_Expecter struct {
	mock *mock.Mock
}

func (_m *NodeExecutionMetadata) EXPECT() *NodeExecutionMetadata_Expecter {
	return &NodeExecutionMetadata_Expecter{mock: &_m.Mock}
}

// GetAnnotations provides a mock function with given fields:
func (_m *NodeExecutionMetadata) GetAnnotations() map[string]string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetAnnotations")
	}

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func() map[string]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	return r0
}

// NodeExecutionMetadata_GetAnnotations_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAnnotations'
type NodeExecutionMetadata_GetAnnotations_Call struct {
	*mock.Call
}

// GetAnnotations is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) GetAnnotations() *NodeExecutionMetadata_GetAnnotations_Call {
	return &NodeExecutionMetadata_GetAnnotations_Call{Call: _e.mock.On("GetAnnotations")}
}

func (_c *NodeExecutionMetadata_GetAnnotations_Call) Run(run func()) *NodeExecutionMetadata_GetAnnotations_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_GetAnnotations_Call) Return(_a0 map[string]string) *NodeExecutionMetadata_GetAnnotations_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_GetAnnotations_Call) RunAndReturn(run func() map[string]string) *NodeExecutionMetadata_GetAnnotations_Call {
	_c.Call.Return(run)
	return _c
}

// GetConsoleURL provides a mock function with given fields:
func (_m *NodeExecutionMetadata) GetConsoleURL() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetConsoleURL")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NodeExecutionMetadata_GetConsoleURL_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetConsoleURL'
type NodeExecutionMetadata_GetConsoleURL_Call struct {
	*mock.Call
}

// GetConsoleURL is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) GetConsoleURL() *NodeExecutionMetadata_GetConsoleURL_Call {
	return &NodeExecutionMetadata_GetConsoleURL_Call{Call: _e.mock.On("GetConsoleURL")}
}

func (_c *NodeExecutionMetadata_GetConsoleURL_Call) Run(run func()) *NodeExecutionMetadata_GetConsoleURL_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_GetConsoleURL_Call) Return(_a0 string) *NodeExecutionMetadata_GetConsoleURL_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_GetConsoleURL_Call) RunAndReturn(run func() string) *NodeExecutionMetadata_GetConsoleURL_Call {
	_c.Call.Return(run)
	return _c
}

// GetInterruptibleFailureThreshold provides a mock function with given fields:
func (_m *NodeExecutionMetadata) GetInterruptibleFailureThreshold() int32 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetInterruptibleFailureThreshold")
	}

	var r0 int32
	if rf, ok := ret.Get(0).(func() int32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int32)
	}

	return r0
}

// NodeExecutionMetadata_GetInterruptibleFailureThreshold_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetInterruptibleFailureThreshold'
type NodeExecutionMetadata_GetInterruptibleFailureThreshold_Call struct {
	*mock.Call
}

// GetInterruptibleFailureThreshold is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) GetInterruptibleFailureThreshold() *NodeExecutionMetadata_GetInterruptibleFailureThreshold_Call {
	return &NodeExecutionMetadata_GetInterruptibleFailureThreshold_Call{Call: _e.mock.On("GetInterruptibleFailureThreshold")}
}

func (_c *NodeExecutionMetadata_GetInterruptibleFailureThreshold_Call) Run(run func()) *NodeExecutionMetadata_GetInterruptibleFailureThreshold_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_GetInterruptibleFailureThreshold_Call) Return(_a0 int32) *NodeExecutionMetadata_GetInterruptibleFailureThreshold_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_GetInterruptibleFailureThreshold_Call) RunAndReturn(run func() int32) *NodeExecutionMetadata_GetInterruptibleFailureThreshold_Call {
	_c.Call.Return(run)
	return _c
}

// GetK8sServiceAccount provides a mock function with given fields:
func (_m *NodeExecutionMetadata) GetK8sServiceAccount() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetK8sServiceAccount")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NodeExecutionMetadata_GetK8sServiceAccount_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetK8sServiceAccount'
type NodeExecutionMetadata_GetK8sServiceAccount_Call struct {
	*mock.Call
}

// GetK8sServiceAccount is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) GetK8sServiceAccount() *NodeExecutionMetadata_GetK8sServiceAccount_Call {
	return &NodeExecutionMetadata_GetK8sServiceAccount_Call{Call: _e.mock.On("GetK8sServiceAccount")}
}

func (_c *NodeExecutionMetadata_GetK8sServiceAccount_Call) Run(run func()) *NodeExecutionMetadata_GetK8sServiceAccount_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_GetK8sServiceAccount_Call) Return(_a0 string) *NodeExecutionMetadata_GetK8sServiceAccount_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_GetK8sServiceAccount_Call) RunAndReturn(run func() string) *NodeExecutionMetadata_GetK8sServiceAccount_Call {
	_c.Call.Return(run)
	return _c
}

// GetLabels provides a mock function with given fields:
func (_m *NodeExecutionMetadata) GetLabels() map[string]string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetLabels")
	}

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func() map[string]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	return r0
}

// NodeExecutionMetadata_GetLabels_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLabels'
type NodeExecutionMetadata_GetLabels_Call struct {
	*mock.Call
}

// GetLabels is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) GetLabels() *NodeExecutionMetadata_GetLabels_Call {
	return &NodeExecutionMetadata_GetLabels_Call{Call: _e.mock.On("GetLabels")}
}

func (_c *NodeExecutionMetadata_GetLabels_Call) Run(run func()) *NodeExecutionMetadata_GetLabels_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_GetLabels_Call) Return(_a0 map[string]string) *NodeExecutionMetadata_GetLabels_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_GetLabels_Call) RunAndReturn(run func() map[string]string) *NodeExecutionMetadata_GetLabels_Call {
	_c.Call.Return(run)
	return _c
}

// GetNamespace provides a mock function with given fields:
func (_m *NodeExecutionMetadata) GetNamespace() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetNamespace")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NodeExecutionMetadata_GetNamespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNamespace'
type NodeExecutionMetadata_GetNamespace_Call struct {
	*mock.Call
}

// GetNamespace is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) GetNamespace() *NodeExecutionMetadata_GetNamespace_Call {
	return &NodeExecutionMetadata_GetNamespace_Call{Call: _e.mock.On("GetNamespace")}
}

func (_c *NodeExecutionMetadata_GetNamespace_Call) Run(run func()) *NodeExecutionMetadata_GetNamespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_GetNamespace_Call) Return(_a0 string) *NodeExecutionMetadata_GetNamespace_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_GetNamespace_Call) RunAndReturn(run func() string) *NodeExecutionMetadata_GetNamespace_Call {
	_c.Call.Return(run)
	return _c
}

// GetNodeExecutionID provides a mock function with given fields:
func (_m *NodeExecutionMetadata) GetNodeExecutionID() *core.NodeExecutionIdentifier {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetNodeExecutionID")
	}

	var r0 *core.NodeExecutionIdentifier
	if rf, ok := ret.Get(0).(func() *core.NodeExecutionIdentifier); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.NodeExecutionIdentifier)
		}
	}

	return r0
}

// NodeExecutionMetadata_GetNodeExecutionID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNodeExecutionID'
type NodeExecutionMetadata_GetNodeExecutionID_Call struct {
	*mock.Call
}

// GetNodeExecutionID is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) GetNodeExecutionID() *NodeExecutionMetadata_GetNodeExecutionID_Call {
	return &NodeExecutionMetadata_GetNodeExecutionID_Call{Call: _e.mock.On("GetNodeExecutionID")}
}

func (_c *NodeExecutionMetadata_GetNodeExecutionID_Call) Run(run func()) *NodeExecutionMetadata_GetNodeExecutionID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_GetNodeExecutionID_Call) Return(_a0 *core.NodeExecutionIdentifier) *NodeExecutionMetadata_GetNodeExecutionID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_GetNodeExecutionID_Call) RunAndReturn(run func() *core.NodeExecutionIdentifier) *NodeExecutionMetadata_GetNodeExecutionID_Call {
	_c.Call.Return(run)
	return _c
}

// GetOwnerID provides a mock function with given fields:
func (_m *NodeExecutionMetadata) GetOwnerID() types.NamespacedName {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetOwnerID")
	}

	var r0 types.NamespacedName
	if rf, ok := ret.Get(0).(func() types.NamespacedName); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(types.NamespacedName)
	}

	return r0
}

// NodeExecutionMetadata_GetOwnerID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOwnerID'
type NodeExecutionMetadata_GetOwnerID_Call struct {
	*mock.Call
}

// GetOwnerID is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) GetOwnerID() *NodeExecutionMetadata_GetOwnerID_Call {
	return &NodeExecutionMetadata_GetOwnerID_Call{Call: _e.mock.On("GetOwnerID")}
}

func (_c *NodeExecutionMetadata_GetOwnerID_Call) Run(run func()) *NodeExecutionMetadata_GetOwnerID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_GetOwnerID_Call) Return(_a0 types.NamespacedName) *NodeExecutionMetadata_GetOwnerID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_GetOwnerID_Call) RunAndReturn(run func() types.NamespacedName) *NodeExecutionMetadata_GetOwnerID_Call {
	_c.Call.Return(run)
	return _c
}

// GetOwnerReference provides a mock function with given fields:
func (_m *NodeExecutionMetadata) GetOwnerReference() v1.OwnerReference {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetOwnerReference")
	}

	var r0 v1.OwnerReference
	if rf, ok := ret.Get(0).(func() v1.OwnerReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1.OwnerReference)
	}

	return r0
}

// NodeExecutionMetadata_GetOwnerReference_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOwnerReference'
type NodeExecutionMetadata_GetOwnerReference_Call struct {
	*mock.Call
}

// GetOwnerReference is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) GetOwnerReference() *NodeExecutionMetadata_GetOwnerReference_Call {
	return &NodeExecutionMetadata_GetOwnerReference_Call{Call: _e.mock.On("GetOwnerReference")}
}

func (_c *NodeExecutionMetadata_GetOwnerReference_Call) Run(run func()) *NodeExecutionMetadata_GetOwnerReference_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_GetOwnerReference_Call) Return(_a0 v1.OwnerReference) *NodeExecutionMetadata_GetOwnerReference_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_GetOwnerReference_Call) RunAndReturn(run func() v1.OwnerReference) *NodeExecutionMetadata_GetOwnerReference_Call {
	_c.Call.Return(run)
	return _c
}

// GetSecurityContext provides a mock function with given fields:
func (_m *NodeExecutionMetadata) GetSecurityContext() core.SecurityContext {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSecurityContext")
	}

	var r0 core.SecurityContext
	if rf, ok := ret.Get(0).(func() core.SecurityContext); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(core.SecurityContext)
	}

	return r0
}

// NodeExecutionMetadata_GetSecurityContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSecurityContext'
type NodeExecutionMetadata_GetSecurityContext_Call struct {
	*mock.Call
}

// GetSecurityContext is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) GetSecurityContext() *NodeExecutionMetadata_GetSecurityContext_Call {
	return &NodeExecutionMetadata_GetSecurityContext_Call{Call: _e.mock.On("GetSecurityContext")}
}

func (_c *NodeExecutionMetadata_GetSecurityContext_Call) Run(run func()) *NodeExecutionMetadata_GetSecurityContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_GetSecurityContext_Call) Return(_a0 core.SecurityContext) *NodeExecutionMetadata_GetSecurityContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_GetSecurityContext_Call) RunAndReturn(run func() core.SecurityContext) *NodeExecutionMetadata_GetSecurityContext_Call {
	_c.Call.Return(run)
	return _c
}

// IsInterruptible provides a mock function with given fields:
func (_m *NodeExecutionMetadata) IsInterruptible() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsInterruptible")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// NodeExecutionMetadata_IsInterruptible_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsInterruptible'
type NodeExecutionMetadata_IsInterruptible_Call struct {
	*mock.Call
}

// IsInterruptible is a helper method to define mock.On call
func (_e *NodeExecutionMetadata_Expecter) IsInterruptible() *NodeExecutionMetadata_IsInterruptible_Call {
	return &NodeExecutionMetadata_IsInterruptible_Call{Call: _e.mock.On("IsInterruptible")}
}

func (_c *NodeExecutionMetadata_IsInterruptible_Call) Run(run func()) *NodeExecutionMetadata_IsInterruptible_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionMetadata_IsInterruptible_Call) Return(_a0 bool) *NodeExecutionMetadata_IsInterruptible_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionMetadata_IsInterruptible_Call) RunAndReturn(run func() bool) *NodeExecutionMetadata_IsInterruptible_Call {
	_c.Call.Return(run)
	return _c
}

// NewNodeExecutionMetadata creates a new instance of NodeExecutionMetadata. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNodeExecutionMetadata(t interface {
	mock.TestingT
	Cleanup(func())
}) *NodeExecutionMetadata {
	mock := &NodeExecutionMetadata{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
