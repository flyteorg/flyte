// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mock "github.com/stretchr/testify/mock"

	types "k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// Meta is an autogenerated mock type for the Meta type
type Meta struct {
	mock.Mock
}

type Meta_Expecter struct {
	mock *mock.Mock
}

func (_m *Meta) EXPECT() *Meta_Expecter {
	return &Meta_Expecter{mock: &_m.Mock}
}

// GetAnnotations provides a mock function with no fields
func (_m *Meta) GetAnnotations() map[string]string {
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

// Meta_GetAnnotations_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAnnotations'
type Meta_GetAnnotations_Call struct {
	*mock.Call
}

// GetAnnotations is a helper method to define mock.On call
func (_e *Meta_Expecter) GetAnnotations() *Meta_GetAnnotations_Call {
	return &Meta_GetAnnotations_Call{Call: _e.mock.On("GetAnnotations")}
}

func (_c *Meta_GetAnnotations_Call) Run(run func()) *Meta_GetAnnotations_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetAnnotations_Call) Return(_a0 map[string]string) *Meta_GetAnnotations_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetAnnotations_Call) RunAndReturn(run func() map[string]string) *Meta_GetAnnotations_Call {
	_c.Call.Return(run)
	return _c
}

// GetConsoleURL provides a mock function with no fields
func (_m *Meta) GetConsoleURL() string {
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

// Meta_GetConsoleURL_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetConsoleURL'
type Meta_GetConsoleURL_Call struct {
	*mock.Call
}

// GetConsoleURL is a helper method to define mock.On call
func (_e *Meta_Expecter) GetConsoleURL() *Meta_GetConsoleURL_Call {
	return &Meta_GetConsoleURL_Call{Call: _e.mock.On("GetConsoleURL")}
}

func (_c *Meta_GetConsoleURL_Call) Run(run func()) *Meta_GetConsoleURL_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetConsoleURL_Call) Return(_a0 string) *Meta_GetConsoleURL_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetConsoleURL_Call) RunAndReturn(run func() string) *Meta_GetConsoleURL_Call {
	_c.Call.Return(run)
	return _c
}

// GetCreationTimestamp provides a mock function with no fields
func (_m *Meta) GetCreationTimestamp() v1.Time {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetCreationTimestamp")
	}

	var r0 v1.Time
	if rf, ok := ret.Get(0).(func() v1.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1.Time)
	}

	return r0
}

// Meta_GetCreationTimestamp_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCreationTimestamp'
type Meta_GetCreationTimestamp_Call struct {
	*mock.Call
}

// GetCreationTimestamp is a helper method to define mock.On call
func (_e *Meta_Expecter) GetCreationTimestamp() *Meta_GetCreationTimestamp_Call {
	return &Meta_GetCreationTimestamp_Call{Call: _e.mock.On("GetCreationTimestamp")}
}

func (_c *Meta_GetCreationTimestamp_Call) Run(run func()) *Meta_GetCreationTimestamp_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetCreationTimestamp_Call) Return(_a0 v1.Time) *Meta_GetCreationTimestamp_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetCreationTimestamp_Call) RunAndReturn(run func() v1.Time) *Meta_GetCreationTimestamp_Call {
	_c.Call.Return(run)
	return _c
}

// GetDefinitionVersion provides a mock function with no fields
func (_m *Meta) GetDefinitionVersion() v1alpha1.WorkflowDefinitionVersion {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetDefinitionVersion")
	}

	var r0 v1alpha1.WorkflowDefinitionVersion
	if rf, ok := ret.Get(0).(func() v1alpha1.WorkflowDefinitionVersion); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1alpha1.WorkflowDefinitionVersion)
	}

	return r0
}

// Meta_GetDefinitionVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDefinitionVersion'
type Meta_GetDefinitionVersion_Call struct {
	*mock.Call
}

// GetDefinitionVersion is a helper method to define mock.On call
func (_e *Meta_Expecter) GetDefinitionVersion() *Meta_GetDefinitionVersion_Call {
	return &Meta_GetDefinitionVersion_Call{Call: _e.mock.On("GetDefinitionVersion")}
}

func (_c *Meta_GetDefinitionVersion_Call) Run(run func()) *Meta_GetDefinitionVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetDefinitionVersion_Call) Return(_a0 v1alpha1.WorkflowDefinitionVersion) *Meta_GetDefinitionVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetDefinitionVersion_Call) RunAndReturn(run func() v1alpha1.WorkflowDefinitionVersion) *Meta_GetDefinitionVersion_Call {
	_c.Call.Return(run)
	return _c
}

// GetEventVersion provides a mock function with no fields
func (_m *Meta) GetEventVersion() v1alpha1.EventVersion {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetEventVersion")
	}

	var r0 v1alpha1.EventVersion
	if rf, ok := ret.Get(0).(func() v1alpha1.EventVersion); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1alpha1.EventVersion)
	}

	return r0
}

// Meta_GetEventVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEventVersion'
type Meta_GetEventVersion_Call struct {
	*mock.Call
}

// GetEventVersion is a helper method to define mock.On call
func (_e *Meta_Expecter) GetEventVersion() *Meta_GetEventVersion_Call {
	return &Meta_GetEventVersion_Call{Call: _e.mock.On("GetEventVersion")}
}

func (_c *Meta_GetEventVersion_Call) Run(run func()) *Meta_GetEventVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetEventVersion_Call) Return(_a0 v1alpha1.EventVersion) *Meta_GetEventVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetEventVersion_Call) RunAndReturn(run func() v1alpha1.EventVersion) *Meta_GetEventVersion_Call {
	_c.Call.Return(run)
	return _c
}

// GetExecutionID provides a mock function with no fields
func (_m *Meta) GetExecutionID() v1alpha1.WorkflowExecutionIdentifier {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetExecutionID")
	}

	var r0 v1alpha1.WorkflowExecutionIdentifier
	if rf, ok := ret.Get(0).(func() v1alpha1.WorkflowExecutionIdentifier); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1alpha1.WorkflowExecutionIdentifier)
	}

	return r0
}

// Meta_GetExecutionID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetExecutionID'
type Meta_GetExecutionID_Call struct {
	*mock.Call
}

// GetExecutionID is a helper method to define mock.On call
func (_e *Meta_Expecter) GetExecutionID() *Meta_GetExecutionID_Call {
	return &Meta_GetExecutionID_Call{Call: _e.mock.On("GetExecutionID")}
}

func (_c *Meta_GetExecutionID_Call) Run(run func()) *Meta_GetExecutionID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetExecutionID_Call) Return(_a0 v1alpha1.WorkflowExecutionIdentifier) *Meta_GetExecutionID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetExecutionID_Call) RunAndReturn(run func() v1alpha1.WorkflowExecutionIdentifier) *Meta_GetExecutionID_Call {
	_c.Call.Return(run)
	return _c
}

// GetK8sWorkflowID provides a mock function with no fields
func (_m *Meta) GetK8sWorkflowID() types.NamespacedName {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetK8sWorkflowID")
	}

	var r0 types.NamespacedName
	if rf, ok := ret.Get(0).(func() types.NamespacedName); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(types.NamespacedName)
	}

	return r0
}

// Meta_GetK8sWorkflowID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetK8sWorkflowID'
type Meta_GetK8sWorkflowID_Call struct {
	*mock.Call
}

// GetK8sWorkflowID is a helper method to define mock.On call
func (_e *Meta_Expecter) GetK8sWorkflowID() *Meta_GetK8sWorkflowID_Call {
	return &Meta_GetK8sWorkflowID_Call{Call: _e.mock.On("GetK8sWorkflowID")}
}

func (_c *Meta_GetK8sWorkflowID_Call) Run(run func()) *Meta_GetK8sWorkflowID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetK8sWorkflowID_Call) Return(_a0 types.NamespacedName) *Meta_GetK8sWorkflowID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetK8sWorkflowID_Call) RunAndReturn(run func() types.NamespacedName) *Meta_GetK8sWorkflowID_Call {
	_c.Call.Return(run)
	return _c
}

// GetLabels provides a mock function with no fields
func (_m *Meta) GetLabels() map[string]string {
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

// Meta_GetLabels_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLabels'
type Meta_GetLabels_Call struct {
	*mock.Call
}

// GetLabels is a helper method to define mock.On call
func (_e *Meta_Expecter) GetLabels() *Meta_GetLabels_Call {
	return &Meta_GetLabels_Call{Call: _e.mock.On("GetLabels")}
}

func (_c *Meta_GetLabels_Call) Run(run func()) *Meta_GetLabels_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetLabels_Call) Return(_a0 map[string]string) *Meta_GetLabels_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetLabels_Call) RunAndReturn(run func() map[string]string) *Meta_GetLabels_Call {
	_c.Call.Return(run)
	return _c
}

// GetName provides a mock function with no fields
func (_m *Meta) GetName() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetName")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Meta_GetName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetName'
type Meta_GetName_Call struct {
	*mock.Call
}

// GetName is a helper method to define mock.On call
func (_e *Meta_Expecter) GetName() *Meta_GetName_Call {
	return &Meta_GetName_Call{Call: _e.mock.On("GetName")}
}

func (_c *Meta_GetName_Call) Run(run func()) *Meta_GetName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetName_Call) Return(_a0 string) *Meta_GetName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetName_Call) RunAndReturn(run func() string) *Meta_GetName_Call {
	_c.Call.Return(run)
	return _c
}

// GetNamespace provides a mock function with no fields
func (_m *Meta) GetNamespace() string {
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

// Meta_GetNamespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNamespace'
type Meta_GetNamespace_Call struct {
	*mock.Call
}

// GetNamespace is a helper method to define mock.On call
func (_e *Meta_Expecter) GetNamespace() *Meta_GetNamespace_Call {
	return &Meta_GetNamespace_Call{Call: _e.mock.On("GetNamespace")}
}

func (_c *Meta_GetNamespace_Call) Run(run func()) *Meta_GetNamespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetNamespace_Call) Return(_a0 string) *Meta_GetNamespace_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetNamespace_Call) RunAndReturn(run func() string) *Meta_GetNamespace_Call {
	_c.Call.Return(run)
	return _c
}

// GetOwnerReference provides a mock function with no fields
func (_m *Meta) GetOwnerReference() v1.OwnerReference {
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

// Meta_GetOwnerReference_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOwnerReference'
type Meta_GetOwnerReference_Call struct {
	*mock.Call
}

// GetOwnerReference is a helper method to define mock.On call
func (_e *Meta_Expecter) GetOwnerReference() *Meta_GetOwnerReference_Call {
	return &Meta_GetOwnerReference_Call{Call: _e.mock.On("GetOwnerReference")}
}

func (_c *Meta_GetOwnerReference_Call) Run(run func()) *Meta_GetOwnerReference_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetOwnerReference_Call) Return(_a0 v1.OwnerReference) *Meta_GetOwnerReference_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetOwnerReference_Call) RunAndReturn(run func() v1.OwnerReference) *Meta_GetOwnerReference_Call {
	_c.Call.Return(run)
	return _c
}

// GetRawOutputDataConfig provides a mock function with no fields
func (_m *Meta) GetRawOutputDataConfig() v1alpha1.RawOutputDataConfig {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetRawOutputDataConfig")
	}

	var r0 v1alpha1.RawOutputDataConfig
	if rf, ok := ret.Get(0).(func() v1alpha1.RawOutputDataConfig); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1alpha1.RawOutputDataConfig)
	}

	return r0
}

// Meta_GetRawOutputDataConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRawOutputDataConfig'
type Meta_GetRawOutputDataConfig_Call struct {
	*mock.Call
}

// GetRawOutputDataConfig is a helper method to define mock.On call
func (_e *Meta_Expecter) GetRawOutputDataConfig() *Meta_GetRawOutputDataConfig_Call {
	return &Meta_GetRawOutputDataConfig_Call{Call: _e.mock.On("GetRawOutputDataConfig")}
}

func (_c *Meta_GetRawOutputDataConfig_Call) Run(run func()) *Meta_GetRawOutputDataConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetRawOutputDataConfig_Call) Return(_a0 v1alpha1.RawOutputDataConfig) *Meta_GetRawOutputDataConfig_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetRawOutputDataConfig_Call) RunAndReturn(run func() v1alpha1.RawOutputDataConfig) *Meta_GetRawOutputDataConfig_Call {
	_c.Call.Return(run)
	return _c
}

// GetSecurityContext provides a mock function with no fields
func (_m *Meta) GetSecurityContext() core.SecurityContext {
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

// Meta_GetSecurityContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSecurityContext'
type Meta_GetSecurityContext_Call struct {
	*mock.Call
}

// GetSecurityContext is a helper method to define mock.On call
func (_e *Meta_Expecter) GetSecurityContext() *Meta_GetSecurityContext_Call {
	return &Meta_GetSecurityContext_Call{Call: _e.mock.On("GetSecurityContext")}
}

func (_c *Meta_GetSecurityContext_Call) Run(run func()) *Meta_GetSecurityContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetSecurityContext_Call) Return(_a0 core.SecurityContext) *Meta_GetSecurityContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetSecurityContext_Call) RunAndReturn(run func() core.SecurityContext) *Meta_GetSecurityContext_Call {
	_c.Call.Return(run)
	return _c
}

// GetServiceAccountName provides a mock function with no fields
func (_m *Meta) GetServiceAccountName() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetServiceAccountName")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Meta_GetServiceAccountName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetServiceAccountName'
type Meta_GetServiceAccountName_Call struct {
	*mock.Call
}

// GetServiceAccountName is a helper method to define mock.On call
func (_e *Meta_Expecter) GetServiceAccountName() *Meta_GetServiceAccountName_Call {
	return &Meta_GetServiceAccountName_Call{Call: _e.mock.On("GetServiceAccountName")}
}

func (_c *Meta_GetServiceAccountName_Call) Run(run func()) *Meta_GetServiceAccountName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_GetServiceAccountName_Call) Return(_a0 string) *Meta_GetServiceAccountName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_GetServiceAccountName_Call) RunAndReturn(run func() string) *Meta_GetServiceAccountName_Call {
	_c.Call.Return(run)
	return _c
}

// IsInterruptible provides a mock function with no fields
func (_m *Meta) IsInterruptible() bool {
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

// Meta_IsInterruptible_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsInterruptible'
type Meta_IsInterruptible_Call struct {
	*mock.Call
}

// IsInterruptible is a helper method to define mock.On call
func (_e *Meta_Expecter) IsInterruptible() *Meta_IsInterruptible_Call {
	return &Meta_IsInterruptible_Call{Call: _e.mock.On("IsInterruptible")}
}

func (_c *Meta_IsInterruptible_Call) Run(run func()) *Meta_IsInterruptible_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Meta_IsInterruptible_Call) Return(_a0 bool) *Meta_IsInterruptible_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Meta_IsInterruptible_Call) RunAndReturn(run func() bool) *Meta_IsInterruptible_Call {
	_c.Call.Return(run)
	return _c
}

// NewMeta creates a new instance of Meta. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMeta(t interface {
	mock.TestingT
	Cleanup(func())
}) *Meta {
	mock := &Meta{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
