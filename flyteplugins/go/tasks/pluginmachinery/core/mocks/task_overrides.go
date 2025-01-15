// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	flyteidlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/api/core/v1"
)

// TaskOverrides is an autogenerated mock type for the TaskOverrides type
type TaskOverrides struct {
	mock.Mock
}

type TaskOverrides_GetAnnotations struct {
	*mock.Call
}

func (_m TaskOverrides_GetAnnotations) Return(_a0 map[string]string) *TaskOverrides_GetAnnotations {
	return &TaskOverrides_GetAnnotations{Call: _m.Call.Return(_a0)}
}

func (_m *TaskOverrides) OnGetAnnotations() *TaskOverrides_GetAnnotations {
	c_call := _m.On("GetAnnotations")
	return &TaskOverrides_GetAnnotations{Call: c_call}
}

func (_m *TaskOverrides) OnGetAnnotationsMatch(matchers ...interface{}) *TaskOverrides_GetAnnotations {
	c_call := _m.On("GetAnnotations", matchers...)
	return &TaskOverrides_GetAnnotations{Call: c_call}
}

// GetAnnotations provides a mock function with given fields:
func (_m *TaskOverrides) GetAnnotations() map[string]string {
	ret := _m.Called()

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

type TaskOverrides_GetConfig struct {
	*mock.Call
}

func (_m TaskOverrides_GetConfig) Return(_a0 *v1.ConfigMap) *TaskOverrides_GetConfig {
	return &TaskOverrides_GetConfig{Call: _m.Call.Return(_a0)}
}

func (_m *TaskOverrides) OnGetConfig() *TaskOverrides_GetConfig {
	c_call := _m.On("GetConfig")
	return &TaskOverrides_GetConfig{Call: c_call}
}

func (_m *TaskOverrides) OnGetConfigMatch(matchers ...interface{}) *TaskOverrides_GetConfig {
	c_call := _m.On("GetConfig", matchers...)
	return &TaskOverrides_GetConfig{Call: c_call}
}

// GetConfig provides a mock function with given fields:
func (_m *TaskOverrides) GetConfig() *v1.ConfigMap {
	ret := _m.Called()

	var r0 *v1.ConfigMap
	if rf, ok := ret.Get(0).(func() *v1.ConfigMap); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	return r0
}

type TaskOverrides_GetContainerImage struct {
	*mock.Call
}

func (_m TaskOverrides_GetContainerImage) Return(_a0 string) *TaskOverrides_GetContainerImage {
	return &TaskOverrides_GetContainerImage{Call: _m.Call.Return(_a0)}
}

func (_m *TaskOverrides) OnGetContainerImage() *TaskOverrides_GetContainerImage {
	c_call := _m.On("GetContainerImage")
	return &TaskOverrides_GetContainerImage{Call: c_call}
}

func (_m *TaskOverrides) OnGetContainerImageMatch(matchers ...interface{}) *TaskOverrides_GetContainerImage {
	c_call := _m.On("GetContainerImage", matchers...)
	return &TaskOverrides_GetContainerImage{Call: c_call}
}

// GetContainerImage provides a mock function with given fields:
func (_m *TaskOverrides) GetContainerImage() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

type TaskOverrides_GetExtendedResources struct {
	*mock.Call
}

func (_m TaskOverrides_GetExtendedResources) Return(_a0 *flyteidlcore.ExtendedResources) *TaskOverrides_GetExtendedResources {
	return &TaskOverrides_GetExtendedResources{Call: _m.Call.Return(_a0)}
}

func (_m *TaskOverrides) OnGetExtendedResources() *TaskOverrides_GetExtendedResources {
	c_call := _m.On("GetExtendedResources")
	return &TaskOverrides_GetExtendedResources{Call: c_call}
}

func (_m *TaskOverrides) OnGetExtendedResourcesMatch(matchers ...interface{}) *TaskOverrides_GetExtendedResources {
	c_call := _m.On("GetExtendedResources", matchers...)
	return &TaskOverrides_GetExtendedResources{Call: c_call}
}

// GetExtendedResources provides a mock function with given fields:
func (_m *TaskOverrides) GetExtendedResources() *flyteidlcore.ExtendedResources {
	ret := _m.Called()

	var r0 *flyteidlcore.ExtendedResources
	if rf, ok := ret.Get(0).(func() *flyteidlcore.ExtendedResources); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flyteidlcore.ExtendedResources)
		}
	}

	return r0
}

type TaskOverrides_GetLabels struct {
	*mock.Call
}

func (_m TaskOverrides_GetLabels) Return(_a0 map[string]string) *TaskOverrides_GetLabels {
	return &TaskOverrides_GetLabels{Call: _m.Call.Return(_a0)}
}

func (_m *TaskOverrides) OnGetLabels() *TaskOverrides_GetLabels {
	c_call := _m.On("GetLabels")
	return &TaskOverrides_GetLabels{Call: c_call}
}

func (_m *TaskOverrides) OnGetLabelsMatch(matchers ...interface{}) *TaskOverrides_GetLabels {
	c_call := _m.On("GetLabels", matchers...)
	return &TaskOverrides_GetLabels{Call: c_call}
}

// GetLabels provides a mock function with given fields:
func (_m *TaskOverrides) GetLabels() map[string]string {
	ret := _m.Called()

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

type TaskOverrides_GetResources struct {
	*mock.Call
}

func (_m TaskOverrides_GetResources) Return(_a0 *v1.ResourceRequirements) *TaskOverrides_GetResources {
	return &TaskOverrides_GetResources{Call: _m.Call.Return(_a0)}
}

func (_m *TaskOverrides) OnGetResources() *TaskOverrides_GetResources {
	c_call := _m.On("GetResources")
	return &TaskOverrides_GetResources{Call: c_call}
}

func (_m *TaskOverrides) OnGetResourcesMatch(matchers ...interface{}) *TaskOverrides_GetResources {
	c_call := _m.On("GetResources", matchers...)
	return &TaskOverrides_GetResources{Call: c_call}
}

// GetResources provides a mock function with given fields:
func (_m *TaskOverrides) GetResources() *v1.ResourceRequirements {
	ret := _m.Called()

	var r0 *v1.ResourceRequirements
	if rf, ok := ret.Get(0).(func() *v1.ResourceRequirements); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ResourceRequirements)
		}
	}

	return r0
}
