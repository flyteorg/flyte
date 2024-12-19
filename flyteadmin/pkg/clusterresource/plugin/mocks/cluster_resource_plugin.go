// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	plugin "github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/plugin"
	mock "github.com/stretchr/testify/mock"
)

// ClusterResourcePlugin is an autogenerated mock type for the ClusterResourcePlugin type
type ClusterResourcePlugin struct {
	mock.Mock
}

type ClusterResourcePlugin_BatchUpdateClusterResourceState struct {
	*mock.Call
}

func (_m ClusterResourcePlugin_BatchUpdateClusterResourceState) Return(_a0 plugin.BatchUpdateClusterResourceStateOutput, _a1 []plugin.BatchUpdateClusterResourceStateError, _a2 error) *ClusterResourcePlugin_BatchUpdateClusterResourceState {
	return &ClusterResourcePlugin_BatchUpdateClusterResourceState{Call: _m.Call.Return(_a0, _a1, _a2)}
}

func (_m *ClusterResourcePlugin) OnBatchUpdateClusterResourceState(ctx context.Context, input *plugin.BatchUpdateClusterResourceStateInput) *ClusterResourcePlugin_BatchUpdateClusterResourceState {
	c_call := _m.On("BatchUpdateClusterResourceState", ctx, input)
	return &ClusterResourcePlugin_BatchUpdateClusterResourceState{Call: c_call}
}

func (_m *ClusterResourcePlugin) OnBatchUpdateClusterResourceStateMatch(matchers ...interface{}) *ClusterResourcePlugin_BatchUpdateClusterResourceState {
	c_call := _m.On("BatchUpdateClusterResourceState", matchers...)
	return &ClusterResourcePlugin_BatchUpdateClusterResourceState{Call: c_call}
}

// BatchUpdateClusterResourceState provides a mock function with given fields: ctx, input
func (_m *ClusterResourcePlugin) BatchUpdateClusterResourceState(ctx context.Context, input *plugin.BatchUpdateClusterResourceStateInput) (plugin.BatchUpdateClusterResourceStateOutput, []plugin.BatchUpdateClusterResourceStateError, error) {
	ret := _m.Called(ctx, input)

	var r0 plugin.BatchUpdateClusterResourceStateOutput
	if rf, ok := ret.Get(0).(func(context.Context, *plugin.BatchUpdateClusterResourceStateInput) plugin.BatchUpdateClusterResourceStateOutput); ok {
		r0 = rf(ctx, input)
	} else {
		r0 = ret.Get(0).(plugin.BatchUpdateClusterResourceStateOutput)
	}

	var r1 []plugin.BatchUpdateClusterResourceStateError
	if rf, ok := ret.Get(1).(func(context.Context, *plugin.BatchUpdateClusterResourceStateInput) []plugin.BatchUpdateClusterResourceStateError); ok {
		r1 = rf(ctx, input)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]plugin.BatchUpdateClusterResourceStateError)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, *plugin.BatchUpdateClusterResourceStateInput) error); ok {
		r2 = rf(ctx, input)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

type ClusterResourcePlugin_GetDeprovisionProjectFilter struct {
	*mock.Call
}

func (_m ClusterResourcePlugin_GetDeprovisionProjectFilter) Return(_a0 string, _a1 error) *ClusterResourcePlugin_GetDeprovisionProjectFilter {
	return &ClusterResourcePlugin_GetDeprovisionProjectFilter{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ClusterResourcePlugin) OnGetDeprovisionProjectFilter(ctx context.Context) *ClusterResourcePlugin_GetDeprovisionProjectFilter {
	c_call := _m.On("GetDeprovisionProjectFilter", ctx)
	return &ClusterResourcePlugin_GetDeprovisionProjectFilter{Call: c_call}
}

func (_m *ClusterResourcePlugin) OnGetDeprovisionProjectFilterMatch(matchers ...interface{}) *ClusterResourcePlugin_GetDeprovisionProjectFilter {
	c_call := _m.On("GetDeprovisionProjectFilter", matchers...)
	return &ClusterResourcePlugin_GetDeprovisionProjectFilter{Call: c_call}
}

// GetDeprovisionProjectFilter provides a mock function with given fields: ctx
func (_m *ClusterResourcePlugin) GetDeprovisionProjectFilter(ctx context.Context) (string, error) {
	ret := _m.Called(ctx)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context) string); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ClusterResourcePlugin_GetProvisionProjectFilter struct {
	*mock.Call
}

func (_m ClusterResourcePlugin_GetProvisionProjectFilter) Return(_a0 string, _a1 error) *ClusterResourcePlugin_GetProvisionProjectFilter {
	return &ClusterResourcePlugin_GetProvisionProjectFilter{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ClusterResourcePlugin) OnGetProvisionProjectFilter(ctx context.Context) *ClusterResourcePlugin_GetProvisionProjectFilter {
	c_call := _m.On("GetProvisionProjectFilter", ctx)
	return &ClusterResourcePlugin_GetProvisionProjectFilter{Call: c_call}
}

func (_m *ClusterResourcePlugin) OnGetProvisionProjectFilterMatch(matchers ...interface{}) *ClusterResourcePlugin_GetProvisionProjectFilter {
	c_call := _m.On("GetProvisionProjectFilter", matchers...)
	return &ClusterResourcePlugin_GetProvisionProjectFilter{Call: c_call}
}

// GetProvisionProjectFilter provides a mock function with given fields: ctx
func (_m *ClusterResourcePlugin) GetProvisionProjectFilter(ctx context.Context) (string, error) {
	ret := _m.Called(ctx)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context) string); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
