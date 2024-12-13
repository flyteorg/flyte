// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	mock "github.com/stretchr/testify/mock"

	service "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

// AdminDeleterExtInterface is an autogenerated mock type for the AdminDeleterExtInterface type
type AdminDeleterExtInterface struct {
	mock.Mock
}

type AdminDeleterExtInterface_AdminServiceClient struct {
	*mock.Call
}

func (_m AdminDeleterExtInterface_AdminServiceClient) Return(_a0 service.AdminServiceClient) *AdminDeleterExtInterface_AdminServiceClient {
	return &AdminDeleterExtInterface_AdminServiceClient{Call: _m.Call.Return(_a0)}
}

func (_m *AdminDeleterExtInterface) OnAdminServiceClient() *AdminDeleterExtInterface_AdminServiceClient {
	c_call := _m.On("AdminServiceClient")
	return &AdminDeleterExtInterface_AdminServiceClient{Call: c_call}
}

func (_m *AdminDeleterExtInterface) OnAdminServiceClientMatch(matchers ...interface{}) *AdminDeleterExtInterface_AdminServiceClient {
	c_call := _m.On("AdminServiceClient", matchers...)
	return &AdminDeleterExtInterface_AdminServiceClient{Call: c_call}
}

// AdminServiceClient provides a mock function with given fields:
func (_m *AdminDeleterExtInterface) AdminServiceClient() service.AdminServiceClient {
	ret := _m.Called()

	var r0 service.AdminServiceClient
	if rf, ok := ret.Get(0).(func() service.AdminServiceClient); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(service.AdminServiceClient)
		}
	}

	return r0
}

type AdminDeleterExtInterface_DeleteProjectAttributes struct {
	*mock.Call
}

func (_m AdminDeleterExtInterface_DeleteProjectAttributes) Return(_a0 error) *AdminDeleterExtInterface_DeleteProjectAttributes {
	return &AdminDeleterExtInterface_DeleteProjectAttributes{Call: _m.Call.Return(_a0)}
}

func (_m *AdminDeleterExtInterface) OnDeleteProjectAttributes(ctx context.Context, project string, rsType admin.MatchableResource) *AdminDeleterExtInterface_DeleteProjectAttributes {
	c_call := _m.On("DeleteProjectAttributes", ctx, project, rsType)
	return &AdminDeleterExtInterface_DeleteProjectAttributes{Call: c_call}
}

func (_m *AdminDeleterExtInterface) OnDeleteProjectAttributesMatch(matchers ...interface{}) *AdminDeleterExtInterface_DeleteProjectAttributes {
	c_call := _m.On("DeleteProjectAttributes", matchers...)
	return &AdminDeleterExtInterface_DeleteProjectAttributes{Call: c_call}
}

// DeleteProjectAttributes provides a mock function with given fields: ctx, project, rsType
func (_m *AdminDeleterExtInterface) DeleteProjectAttributes(ctx context.Context, project string, rsType admin.MatchableResource) error {
	ret := _m.Called(ctx, project, rsType)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, admin.MatchableResource) error); ok {
		r0 = rf(ctx, project, rsType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type AdminDeleterExtInterface_DeleteProjectDomainAttributes struct {
	*mock.Call
}

func (_m AdminDeleterExtInterface_DeleteProjectDomainAttributes) Return(_a0 error) *AdminDeleterExtInterface_DeleteProjectDomainAttributes {
	return &AdminDeleterExtInterface_DeleteProjectDomainAttributes{Call: _m.Call.Return(_a0)}
}

func (_m *AdminDeleterExtInterface) OnDeleteProjectDomainAttributes(ctx context.Context, project string, domain string, rsType admin.MatchableResource) *AdminDeleterExtInterface_DeleteProjectDomainAttributes {
	c_call := _m.On("DeleteProjectDomainAttributes", ctx, project, domain, rsType)
	return &AdminDeleterExtInterface_DeleteProjectDomainAttributes{Call: c_call}
}

func (_m *AdminDeleterExtInterface) OnDeleteProjectDomainAttributesMatch(matchers ...interface{}) *AdminDeleterExtInterface_DeleteProjectDomainAttributes {
	c_call := _m.On("DeleteProjectDomainAttributes", matchers...)
	return &AdminDeleterExtInterface_DeleteProjectDomainAttributes{Call: c_call}
}

// DeleteProjectDomainAttributes provides a mock function with given fields: ctx, project, domain, rsType
func (_m *AdminDeleterExtInterface) DeleteProjectDomainAttributes(ctx context.Context, project string, domain string, rsType admin.MatchableResource) error {
	ret := _m.Called(ctx, project, domain, rsType)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, admin.MatchableResource) error); ok {
		r0 = rf(ctx, project, domain, rsType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type AdminDeleterExtInterface_DeleteWorkflowAttributes struct {
	*mock.Call
}

func (_m AdminDeleterExtInterface_DeleteWorkflowAttributes) Return(_a0 error) *AdminDeleterExtInterface_DeleteWorkflowAttributes {
	return &AdminDeleterExtInterface_DeleteWorkflowAttributes{Call: _m.Call.Return(_a0)}
}

func (_m *AdminDeleterExtInterface) OnDeleteWorkflowAttributes(ctx context.Context, project string, domain string, name string, rsType admin.MatchableResource) *AdminDeleterExtInterface_DeleteWorkflowAttributes {
	c_call := _m.On("DeleteWorkflowAttributes", ctx, project, domain, name, rsType)
	return &AdminDeleterExtInterface_DeleteWorkflowAttributes{Call: c_call}
}

func (_m *AdminDeleterExtInterface) OnDeleteWorkflowAttributesMatch(matchers ...interface{}) *AdminDeleterExtInterface_DeleteWorkflowAttributes {
	c_call := _m.On("DeleteWorkflowAttributes", matchers...)
	return &AdminDeleterExtInterface_DeleteWorkflowAttributes{Call: c_call}
}

// DeleteWorkflowAttributes provides a mock function with given fields: ctx, project, domain, name, rsType
func (_m *AdminDeleterExtInterface) DeleteWorkflowAttributes(ctx context.Context, project string, domain string, name string, rsType admin.MatchableResource) error {
	ret := _m.Called(ctx, project, domain, name, rsType)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, admin.MatchableResource) error); ok {
		r0 = rf(ctx, project, domain, name, rsType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
