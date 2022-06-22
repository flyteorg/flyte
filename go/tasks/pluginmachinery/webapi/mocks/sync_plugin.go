// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	core "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	mock "github.com/stretchr/testify/mock"

	webapi "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
)

// SyncPlugin is an autogenerated mock type for the SyncPlugin type
type SyncPlugin struct {
	mock.Mock
}

type SyncPlugin_Do struct {
	*mock.Call
}

func (_m SyncPlugin_Do) Return(phase core.PhaseInfo, err error) *SyncPlugin_Do {
	return &SyncPlugin_Do{Call: _m.Call.Return(phase, err)}
}

func (_m *SyncPlugin) OnDo(ctx context.Context, tCtx webapi.TaskExecutionContext) *SyncPlugin_Do {
	c_call := _m.On("Do", ctx, tCtx)
	return &SyncPlugin_Do{Call: c_call}
}

func (_m *SyncPlugin) OnDoMatch(matchers ...interface{}) *SyncPlugin_Do {
	c_call := _m.On("Do", matchers...)
	return &SyncPlugin_Do{Call: c_call}
}

// Do provides a mock function with given fields: ctx, tCtx
func (_m *SyncPlugin) Do(ctx context.Context, tCtx webapi.TaskExecutionContext) (core.PhaseInfo, error) {
	ret := _m.Called(ctx, tCtx)

	var r0 core.PhaseInfo
	if rf, ok := ret.Get(0).(func(context.Context, webapi.TaskExecutionContext) core.PhaseInfo); ok {
		r0 = rf(ctx, tCtx)
	} else {
		r0 = ret.Get(0).(core.PhaseInfo)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, webapi.TaskExecutionContext) error); ok {
		r1 = rf(ctx, tCtx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type SyncPlugin_GetConfig struct {
	*mock.Call
}

func (_m SyncPlugin_GetConfig) Return(_a0 webapi.PluginConfig) *SyncPlugin_GetConfig {
	return &SyncPlugin_GetConfig{Call: _m.Call.Return(_a0)}
}

func (_m *SyncPlugin) OnGetConfig() *SyncPlugin_GetConfig {
	c_call := _m.On("GetConfig")
	return &SyncPlugin_GetConfig{Call: c_call}
}

func (_m *SyncPlugin) OnGetConfigMatch(matchers ...interface{}) *SyncPlugin_GetConfig {
	c_call := _m.On("GetConfig", matchers...)
	return &SyncPlugin_GetConfig{Call: c_call}
}

// GetConfig provides a mock function with given fields:
func (_m *SyncPlugin) GetConfig() webapi.PluginConfig {
	ret := _m.Called()

	var r0 webapi.PluginConfig
	if rf, ok := ret.Get(0).(func() webapi.PluginConfig); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(webapi.PluginConfig)
	}

	return r0
}
