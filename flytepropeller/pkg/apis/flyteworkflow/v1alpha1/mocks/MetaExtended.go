// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mock "github.com/stretchr/testify/mock"

	types "k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// MetaExtended is an autogenerated mock type for the MetaExtended type
type MetaExtended struct {
	mock.Mock
}

type MetaExtended_Expecter struct {
	mock *mock.Mock
}

func (_m *MetaExtended) EXPECT() *MetaExtended_Expecter {
	return &MetaExtended_Expecter{mock: &_m.Mock}
}

// FindSubWorkflow provides a mock function with given fields: subID
func (_m *MetaExtended) FindSubWorkflow(subID string) v1alpha1.ExecutableSubWorkflow {
	ret := _m.Called(subID)

	if len(ret) == 0 {
		panic("no return value specified for FindSubWorkflow")
	}

	var r0 v1alpha1.ExecutableSubWorkflow
	if rf, ok := ret.Get(0).(func(string) v1alpha1.ExecutableSubWorkflow); ok {
		r0 = rf(subID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.ExecutableSubWorkflow)
		}
	}

	return r0
}

// MetaExtended_FindSubWorkflow_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindSubWorkflow'
type MetaExtended_FindSubWorkflow_Call struct {
	*mock.Call
}

// FindSubWorkflow is a helper method to define mock.On call
//   - subID string
func (_e *MetaExtended_Expecter) FindSubWorkflow(subID interface{}) *MetaExtended_FindSubWorkflow_Call {
	return &MetaExtended_FindSubWorkflow_Call{Call: _e.mock.On("FindSubWorkflow", subID)}
}

func (_c *MetaExtended_FindSubWorkflow_Call) Run(run func(subID string)) *MetaExtended_FindSubWorkflow_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MetaExtended_FindSubWorkflow_Call) Return(_a0 v1alpha1.ExecutableSubWorkflow) *MetaExtended_FindSubWorkflow_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_FindSubWorkflow_Call) RunAndReturn(run func(string) v1alpha1.ExecutableSubWorkflow) *MetaExtended_FindSubWorkflow_Call {
	_c.Call.Return(run)
	return _c
}

// GetAnnotations provides a mock function with given fields:
func (_m *MetaExtended) GetAnnotations() map[string]string {
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

// MetaExtended_GetAnnotations_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAnnotations'
type MetaExtended_GetAnnotations_Call struct {
	*mock.Call
}

// GetAnnotations is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetAnnotations() *MetaExtended_GetAnnotations_Call {
	return &MetaExtended_GetAnnotations_Call{Call: _e.mock.On("GetAnnotations")}
}

func (_c *MetaExtended_GetAnnotations_Call) Run(run func()) *MetaExtended_GetAnnotations_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetAnnotations_Call) Return(_a0 map[string]string) *MetaExtended_GetAnnotations_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetAnnotations_Call) RunAndReturn(run func() map[string]string) *MetaExtended_GetAnnotations_Call {
	_c.Call.Return(run)
	return _c
}

// GetConsoleURL provides a mock function with given fields:
func (_m *MetaExtended) GetConsoleURL() string {
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

// MetaExtended_GetConsoleURL_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetConsoleURL'
type MetaExtended_GetConsoleURL_Call struct {
	*mock.Call
}

// GetConsoleURL is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetConsoleURL() *MetaExtended_GetConsoleURL_Call {
	return &MetaExtended_GetConsoleURL_Call{Call: _e.mock.On("GetConsoleURL")}
}

func (_c *MetaExtended_GetConsoleURL_Call) Run(run func()) *MetaExtended_GetConsoleURL_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetConsoleURL_Call) Return(_a0 string) *MetaExtended_GetConsoleURL_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetConsoleURL_Call) RunAndReturn(run func() string) *MetaExtended_GetConsoleURL_Call {
	_c.Call.Return(run)
	return _c
}

// GetCreationTimestamp provides a mock function with given fields:
func (_m *MetaExtended) GetCreationTimestamp() v1.Time {
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

// MetaExtended_GetCreationTimestamp_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCreationTimestamp'
type MetaExtended_GetCreationTimestamp_Call struct {
	*mock.Call
}

// GetCreationTimestamp is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetCreationTimestamp() *MetaExtended_GetCreationTimestamp_Call {
	return &MetaExtended_GetCreationTimestamp_Call{Call: _e.mock.On("GetCreationTimestamp")}
}

func (_c *MetaExtended_GetCreationTimestamp_Call) Run(run func()) *MetaExtended_GetCreationTimestamp_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetCreationTimestamp_Call) Return(_a0 v1.Time) *MetaExtended_GetCreationTimestamp_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetCreationTimestamp_Call) RunAndReturn(run func() v1.Time) *MetaExtended_GetCreationTimestamp_Call {
	_c.Call.Return(run)
	return _c
}

// GetDefinitionVersion provides a mock function with given fields:
func (_m *MetaExtended) GetDefinitionVersion() v1alpha1.WorkflowDefinitionVersion {
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

// MetaExtended_GetDefinitionVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDefinitionVersion'
type MetaExtended_GetDefinitionVersion_Call struct {
	*mock.Call
}

// GetDefinitionVersion is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetDefinitionVersion() *MetaExtended_GetDefinitionVersion_Call {
	return &MetaExtended_GetDefinitionVersion_Call{Call: _e.mock.On("GetDefinitionVersion")}
}

func (_c *MetaExtended_GetDefinitionVersion_Call) Run(run func()) *MetaExtended_GetDefinitionVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetDefinitionVersion_Call) Return(_a0 v1alpha1.WorkflowDefinitionVersion) *MetaExtended_GetDefinitionVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetDefinitionVersion_Call) RunAndReturn(run func() v1alpha1.WorkflowDefinitionVersion) *MetaExtended_GetDefinitionVersion_Call {
	_c.Call.Return(run)
	return _c
}

// GetEventVersion provides a mock function with given fields:
func (_m *MetaExtended) GetEventVersion() v1alpha1.EventVersion {
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

// MetaExtended_GetEventVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEventVersion'
type MetaExtended_GetEventVersion_Call struct {
	*mock.Call
}

// GetEventVersion is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetEventVersion() *MetaExtended_GetEventVersion_Call {
	return &MetaExtended_GetEventVersion_Call{Call: _e.mock.On("GetEventVersion")}
}

func (_c *MetaExtended_GetEventVersion_Call) Run(run func()) *MetaExtended_GetEventVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetEventVersion_Call) Return(_a0 v1alpha1.EventVersion) *MetaExtended_GetEventVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetEventVersion_Call) RunAndReturn(run func() v1alpha1.EventVersion) *MetaExtended_GetEventVersion_Call {
	_c.Call.Return(run)
	return _c
}

// GetExecutionID provides a mock function with given fields:
func (_m *MetaExtended) GetExecutionID() v1alpha1.WorkflowExecutionIdentifier {
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

// MetaExtended_GetExecutionID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetExecutionID'
type MetaExtended_GetExecutionID_Call struct {
	*mock.Call
}

// GetExecutionID is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetExecutionID() *MetaExtended_GetExecutionID_Call {
	return &MetaExtended_GetExecutionID_Call{Call: _e.mock.On("GetExecutionID")}
}

func (_c *MetaExtended_GetExecutionID_Call) Run(run func()) *MetaExtended_GetExecutionID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetExecutionID_Call) Return(_a0 v1alpha1.WorkflowExecutionIdentifier) *MetaExtended_GetExecutionID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetExecutionID_Call) RunAndReturn(run func() v1alpha1.WorkflowExecutionIdentifier) *MetaExtended_GetExecutionID_Call {
	_c.Call.Return(run)
	return _c
}

// GetExecutionStatus provides a mock function with given fields:
func (_m *MetaExtended) GetExecutionStatus() v1alpha1.ExecutableWorkflowStatus {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetExecutionStatus")
	}

	var r0 v1alpha1.ExecutableWorkflowStatus
	if rf, ok := ret.Get(0).(func() v1alpha1.ExecutableWorkflowStatus); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.ExecutableWorkflowStatus)
		}
	}

	return r0
}

// MetaExtended_GetExecutionStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetExecutionStatus'
type MetaExtended_GetExecutionStatus_Call struct {
	*mock.Call
}

// GetExecutionStatus is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetExecutionStatus() *MetaExtended_GetExecutionStatus_Call {
	return &MetaExtended_GetExecutionStatus_Call{Call: _e.mock.On("GetExecutionStatus")}
}

func (_c *MetaExtended_GetExecutionStatus_Call) Run(run func()) *MetaExtended_GetExecutionStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetExecutionStatus_Call) Return(_a0 v1alpha1.ExecutableWorkflowStatus) *MetaExtended_GetExecutionStatus_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetExecutionStatus_Call) RunAndReturn(run func() v1alpha1.ExecutableWorkflowStatus) *MetaExtended_GetExecutionStatus_Call {
	_c.Call.Return(run)
	return _c
}

// GetK8sWorkflowID provides a mock function with given fields:
func (_m *MetaExtended) GetK8sWorkflowID() types.NamespacedName {
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

// MetaExtended_GetK8sWorkflowID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetK8sWorkflowID'
type MetaExtended_GetK8sWorkflowID_Call struct {
	*mock.Call
}

// GetK8sWorkflowID is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetK8sWorkflowID() *MetaExtended_GetK8sWorkflowID_Call {
	return &MetaExtended_GetK8sWorkflowID_Call{Call: _e.mock.On("GetK8sWorkflowID")}
}

func (_c *MetaExtended_GetK8sWorkflowID_Call) Run(run func()) *MetaExtended_GetK8sWorkflowID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetK8sWorkflowID_Call) Return(_a0 types.NamespacedName) *MetaExtended_GetK8sWorkflowID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetK8sWorkflowID_Call) RunAndReturn(run func() types.NamespacedName) *MetaExtended_GetK8sWorkflowID_Call {
	_c.Call.Return(run)
	return _c
}

// GetLabels provides a mock function with given fields:
func (_m *MetaExtended) GetLabels() map[string]string {
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

// MetaExtended_GetLabels_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLabels'
type MetaExtended_GetLabels_Call struct {
	*mock.Call
}

// GetLabels is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetLabels() *MetaExtended_GetLabels_Call {
	return &MetaExtended_GetLabels_Call{Call: _e.mock.On("GetLabels")}
}

func (_c *MetaExtended_GetLabels_Call) Run(run func()) *MetaExtended_GetLabels_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetLabels_Call) Return(_a0 map[string]string) *MetaExtended_GetLabels_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetLabels_Call) RunAndReturn(run func() map[string]string) *MetaExtended_GetLabels_Call {
	_c.Call.Return(run)
	return _c
}

// GetName provides a mock function with given fields:
func (_m *MetaExtended) GetName() string {
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

// MetaExtended_GetName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetName'
type MetaExtended_GetName_Call struct {
	*mock.Call
}

// GetName is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetName() *MetaExtended_GetName_Call {
	return &MetaExtended_GetName_Call{Call: _e.mock.On("GetName")}
}

func (_c *MetaExtended_GetName_Call) Run(run func()) *MetaExtended_GetName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetName_Call) Return(_a0 string) *MetaExtended_GetName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetName_Call) RunAndReturn(run func() string) *MetaExtended_GetName_Call {
	_c.Call.Return(run)
	return _c
}

// GetNamespace provides a mock function with given fields:
func (_m *MetaExtended) GetNamespace() string {
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

// MetaExtended_GetNamespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNamespace'
type MetaExtended_GetNamespace_Call struct {
	*mock.Call
}

// GetNamespace is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetNamespace() *MetaExtended_GetNamespace_Call {
	return &MetaExtended_GetNamespace_Call{Call: _e.mock.On("GetNamespace")}
}

func (_c *MetaExtended_GetNamespace_Call) Run(run func()) *MetaExtended_GetNamespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetNamespace_Call) Return(_a0 string) *MetaExtended_GetNamespace_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetNamespace_Call) RunAndReturn(run func() string) *MetaExtended_GetNamespace_Call {
	_c.Call.Return(run)
	return _c
}

// GetOwnerReference provides a mock function with given fields:
func (_m *MetaExtended) GetOwnerReference() v1.OwnerReference {
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

// MetaExtended_GetOwnerReference_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOwnerReference'
type MetaExtended_GetOwnerReference_Call struct {
	*mock.Call
}

// GetOwnerReference is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetOwnerReference() *MetaExtended_GetOwnerReference_Call {
	return &MetaExtended_GetOwnerReference_Call{Call: _e.mock.On("GetOwnerReference")}
}

func (_c *MetaExtended_GetOwnerReference_Call) Run(run func()) *MetaExtended_GetOwnerReference_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetOwnerReference_Call) Return(_a0 v1.OwnerReference) *MetaExtended_GetOwnerReference_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetOwnerReference_Call) RunAndReturn(run func() v1.OwnerReference) *MetaExtended_GetOwnerReference_Call {
	_c.Call.Return(run)
	return _c
}

// GetRawOutputDataConfig provides a mock function with given fields:
func (_m *MetaExtended) GetRawOutputDataConfig() v1alpha1.RawOutputDataConfig {
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

// MetaExtended_GetRawOutputDataConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRawOutputDataConfig'
type MetaExtended_GetRawOutputDataConfig_Call struct {
	*mock.Call
}

// GetRawOutputDataConfig is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetRawOutputDataConfig() *MetaExtended_GetRawOutputDataConfig_Call {
	return &MetaExtended_GetRawOutputDataConfig_Call{Call: _e.mock.On("GetRawOutputDataConfig")}
}

func (_c *MetaExtended_GetRawOutputDataConfig_Call) Run(run func()) *MetaExtended_GetRawOutputDataConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetRawOutputDataConfig_Call) Return(_a0 v1alpha1.RawOutputDataConfig) *MetaExtended_GetRawOutputDataConfig_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetRawOutputDataConfig_Call) RunAndReturn(run func() v1alpha1.RawOutputDataConfig) *MetaExtended_GetRawOutputDataConfig_Call {
	_c.Call.Return(run)
	return _c
}

// GetSecurityContext provides a mock function with given fields:
func (_m *MetaExtended) GetSecurityContext() core.SecurityContext {
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

// MetaExtended_GetSecurityContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSecurityContext'
type MetaExtended_GetSecurityContext_Call struct {
	*mock.Call
}

// GetSecurityContext is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetSecurityContext() *MetaExtended_GetSecurityContext_Call {
	return &MetaExtended_GetSecurityContext_Call{Call: _e.mock.On("GetSecurityContext")}
}

func (_c *MetaExtended_GetSecurityContext_Call) Run(run func()) *MetaExtended_GetSecurityContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetSecurityContext_Call) Return(_a0 core.SecurityContext) *MetaExtended_GetSecurityContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetSecurityContext_Call) RunAndReturn(run func() core.SecurityContext) *MetaExtended_GetSecurityContext_Call {
	_c.Call.Return(run)
	return _c
}

// GetServiceAccountName provides a mock function with given fields:
func (_m *MetaExtended) GetServiceAccountName() string {
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

// MetaExtended_GetServiceAccountName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetServiceAccountName'
type MetaExtended_GetServiceAccountName_Call struct {
	*mock.Call
}

// GetServiceAccountName is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) GetServiceAccountName() *MetaExtended_GetServiceAccountName_Call {
	return &MetaExtended_GetServiceAccountName_Call{Call: _e.mock.On("GetServiceAccountName")}
}

func (_c *MetaExtended_GetServiceAccountName_Call) Run(run func()) *MetaExtended_GetServiceAccountName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_GetServiceAccountName_Call) Return(_a0 string) *MetaExtended_GetServiceAccountName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_GetServiceAccountName_Call) RunAndReturn(run func() string) *MetaExtended_GetServiceAccountName_Call {
	_c.Call.Return(run)
	return _c
}

// GetTask provides a mock function with given fields: id
func (_m *MetaExtended) GetTask(id string) (v1alpha1.ExecutableTask, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for GetTask")
	}

	var r0 v1alpha1.ExecutableTask
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (v1alpha1.ExecutableTask, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) v1alpha1.ExecutableTask); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.ExecutableTask)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MetaExtended_GetTask_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTask'
type MetaExtended_GetTask_Call struct {
	*mock.Call
}

// GetTask is a helper method to define mock.On call
//   - id string
func (_e *MetaExtended_Expecter) GetTask(id interface{}) *MetaExtended_GetTask_Call {
	return &MetaExtended_GetTask_Call{Call: _e.mock.On("GetTask", id)}
}

func (_c *MetaExtended_GetTask_Call) Run(run func(id string)) *MetaExtended_GetTask_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MetaExtended_GetTask_Call) Return(_a0 v1alpha1.ExecutableTask, _a1 error) *MetaExtended_GetTask_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MetaExtended_GetTask_Call) RunAndReturn(run func(string) (v1alpha1.ExecutableTask, error)) *MetaExtended_GetTask_Call {
	_c.Call.Return(run)
	return _c
}

// IsInterruptible provides a mock function with given fields:
func (_m *MetaExtended) IsInterruptible() bool {
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

// MetaExtended_IsInterruptible_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsInterruptible'
type MetaExtended_IsInterruptible_Call struct {
	*mock.Call
}

// IsInterruptible is a helper method to define mock.On call
func (_e *MetaExtended_Expecter) IsInterruptible() *MetaExtended_IsInterruptible_Call {
	return &MetaExtended_IsInterruptible_Call{Call: _e.mock.On("IsInterruptible")}
}

func (_c *MetaExtended_IsInterruptible_Call) Run(run func()) *MetaExtended_IsInterruptible_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MetaExtended_IsInterruptible_Call) Return(_a0 bool) *MetaExtended_IsInterruptible_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MetaExtended_IsInterruptible_Call) RunAndReturn(run func() bool) *MetaExtended_IsInterruptible_Call {
	_c.Call.Return(run)
	return _c
}

// NewMetaExtended creates a new instance of MetaExtended. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMetaExtended(t interface {
	mock.TestingT
	Cleanup(func())
}) *MetaExtended {
	mock := &MetaExtended{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
