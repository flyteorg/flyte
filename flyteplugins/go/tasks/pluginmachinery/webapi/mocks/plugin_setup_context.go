// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	promutils "github.com/flyteorg/flyte/flytestdlib/promutils"
	mock "github.com/stretchr/testify/mock"
)

// PluginSetupContext is an autogenerated mock type for the PluginSetupContext type
type PluginSetupContext struct {
	mock.Mock
}

type PluginSetupContext_MetricsScope struct {
	*mock.Call
}

func (_m PluginSetupContext_MetricsScope) Return(_a0 promutils.Scope) *PluginSetupContext_MetricsScope {
	return &PluginSetupContext_MetricsScope{Call: _m.Call.Return(_a0)}
}

func (_m *PluginSetupContext) OnMetricsScope() *PluginSetupContext_MetricsScope {
	c_call := _m.On("MetricsScope")
	return &PluginSetupContext_MetricsScope{Call: c_call}
}

func (_m *PluginSetupContext) OnMetricsScopeMatch(matchers ...interface{}) *PluginSetupContext_MetricsScope {
	c_call := _m.On("MetricsScope", matchers...)
	return &PluginSetupContext_MetricsScope{Call: c_call}
}

// MetricsScope provides a mock function with given fields:
func (_m *PluginSetupContext) MetricsScope() promutils.Scope {
	ret := _m.Called()

	var r0 promutils.Scope
	if rf, ok := ret.Get(0).(func() promutils.Scope); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(promutils.Scope)
		}
	}

	return r0
}
