// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	core "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	mock "github.com/stretchr/testify/mock"
)

// ResourceRegistrar is an autogenerated mock type for the ResourceRegistrar type
type ResourceNegotiator struct {
	mock.Mock
}

type ResourceNegotiator_RegisterResourceQuota struct {
	*mock.Call
}

func (_m ResourceNegotiator_RegisterResourceQuota) Return(_a0 error) *ResourceNegotiator_RegisterResourceQuota {
	return &ResourceNegotiator_RegisterResourceQuota{Call: _m.Call.Return(_a0)}
}

func (_m *ResourceNegotiator) OnRegisterResourceQuota(ctx context.Context, namespace core.ResourceNamespace, quota int) *ResourceNegotiator_RegisterResourceQuota {
	c := _m.On("RegisterResourceQuota", ctx, namespace, quota)
	return &ResourceNegotiator_RegisterResourceQuota{Call: c}
}

func (_m *ResourceNegotiator) OnRegisterResourceQuotaMatch(matchers ...interface{}) *ResourceNegotiator_RegisterResourceQuota {
	c := _m.On("RegisterResourceQuota", matchers...)
	return &ResourceNegotiator_RegisterResourceQuota{Call: c}
}

// RegisterResourceQuota provides a mock function with given fields: ctx, namespace, quota
func (_m *ResourceNegotiator) RegisterResourceQuota(ctx context.Context, namespace core.ResourceNamespace, quota int) error {
	ret := _m.Called(ctx, namespace, quota)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, core.ResourceNamespace, int) error); ok {
		r0 = rf(ctx, namespace, quota)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
