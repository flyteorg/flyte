// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	interfaces "github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"

	mock "github.com/stretchr/testify/mock"
)

// ResourceInterface is an autogenerated mock type for the ResourceInterface type
type ResourceInterface struct {
	mock.Mock
}

type ResourceInterface_DeleteOrgAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_DeleteOrgAttributes) Return(_a0 *admin.OrgAttributesDeleteResponse, _a1 error) *ResourceInterface_DeleteOrgAttributes {
	return &ResourceInterface_DeleteOrgAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnDeleteOrgAttributes(ctx context.Context, request admin.OrgAttributesDeleteRequest) *ResourceInterface_DeleteOrgAttributes {
	c_call := _m.On("DeleteOrgAttributes", ctx, request)
	return &ResourceInterface_DeleteOrgAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnDeleteOrgAttributesMatch(matchers ...interface{}) *ResourceInterface_DeleteOrgAttributes {
	c_call := _m.On("DeleteOrgAttributes", matchers...)
	return &ResourceInterface_DeleteOrgAttributes{Call: c_call}
}

// DeleteOrgAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) DeleteOrgAttributes(ctx context.Context, request admin.OrgAttributesDeleteRequest) (*admin.OrgAttributesDeleteResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.OrgAttributesDeleteResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.OrgAttributesDeleteRequest) *admin.OrgAttributesDeleteResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.OrgAttributesDeleteResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.OrgAttributesDeleteRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_DeleteProjectAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_DeleteProjectAttributes) Return(_a0 *admin.ProjectAttributesDeleteResponse, _a1 error) *ResourceInterface_DeleteProjectAttributes {
	return &ResourceInterface_DeleteProjectAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnDeleteProjectAttributes(ctx context.Context, request admin.ProjectAttributesDeleteRequest) *ResourceInterface_DeleteProjectAttributes {
	c_call := _m.On("DeleteProjectAttributes", ctx, request)
	return &ResourceInterface_DeleteProjectAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnDeleteProjectAttributesMatch(matchers ...interface{}) *ResourceInterface_DeleteProjectAttributes {
	c_call := _m.On("DeleteProjectAttributes", matchers...)
	return &ResourceInterface_DeleteProjectAttributes{Call: c_call}
}

// DeleteProjectAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) DeleteProjectAttributes(ctx context.Context, request admin.ProjectAttributesDeleteRequest) (*admin.ProjectAttributesDeleteResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.ProjectAttributesDeleteResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.ProjectAttributesDeleteRequest) *admin.ProjectAttributesDeleteResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.ProjectAttributesDeleteResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.ProjectAttributesDeleteRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_DeleteProjectDomainAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_DeleteProjectDomainAttributes) Return(_a0 *admin.ProjectDomainAttributesDeleteResponse, _a1 error) *ResourceInterface_DeleteProjectDomainAttributes {
	return &ResourceInterface_DeleteProjectDomainAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnDeleteProjectDomainAttributes(ctx context.Context, request admin.ProjectDomainAttributesDeleteRequest) *ResourceInterface_DeleteProjectDomainAttributes {
	c_call := _m.On("DeleteProjectDomainAttributes", ctx, request)
	return &ResourceInterface_DeleteProjectDomainAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnDeleteProjectDomainAttributesMatch(matchers ...interface{}) *ResourceInterface_DeleteProjectDomainAttributes {
	c_call := _m.On("DeleteProjectDomainAttributes", matchers...)
	return &ResourceInterface_DeleteProjectDomainAttributes{Call: c_call}
}

// DeleteProjectDomainAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) DeleteProjectDomainAttributes(ctx context.Context, request admin.ProjectDomainAttributesDeleteRequest) (*admin.ProjectDomainAttributesDeleteResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.ProjectDomainAttributesDeleteResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.ProjectDomainAttributesDeleteRequest) *admin.ProjectDomainAttributesDeleteResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.ProjectDomainAttributesDeleteResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.ProjectDomainAttributesDeleteRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_DeleteWorkflowAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_DeleteWorkflowAttributes) Return(_a0 *admin.WorkflowAttributesDeleteResponse, _a1 error) *ResourceInterface_DeleteWorkflowAttributes {
	return &ResourceInterface_DeleteWorkflowAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnDeleteWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesDeleteRequest) *ResourceInterface_DeleteWorkflowAttributes {
	c_call := _m.On("DeleteWorkflowAttributes", ctx, request)
	return &ResourceInterface_DeleteWorkflowAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnDeleteWorkflowAttributesMatch(matchers ...interface{}) *ResourceInterface_DeleteWorkflowAttributes {
	c_call := _m.On("DeleteWorkflowAttributes", matchers...)
	return &ResourceInterface_DeleteWorkflowAttributes{Call: c_call}
}

// DeleteWorkflowAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) DeleteWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesDeleteRequest) (*admin.WorkflowAttributesDeleteResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.WorkflowAttributesDeleteResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.WorkflowAttributesDeleteRequest) *admin.WorkflowAttributesDeleteResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.WorkflowAttributesDeleteResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.WorkflowAttributesDeleteRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_GetOrgAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_GetOrgAttributes) Return(_a0 *admin.OrgAttributesGetResponse, _a1 error) *ResourceInterface_GetOrgAttributes {
	return &ResourceInterface_GetOrgAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnGetOrgAttributes(ctx context.Context, request admin.OrgAttributesGetRequest) *ResourceInterface_GetOrgAttributes {
	c_call := _m.On("GetOrgAttributes", ctx, request)
	return &ResourceInterface_GetOrgAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnGetOrgAttributesMatch(matchers ...interface{}) *ResourceInterface_GetOrgAttributes {
	c_call := _m.On("GetOrgAttributes", matchers...)
	return &ResourceInterface_GetOrgAttributes{Call: c_call}
}

// GetOrgAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) GetOrgAttributes(ctx context.Context, request admin.OrgAttributesGetRequest) (*admin.OrgAttributesGetResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.OrgAttributesGetResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.OrgAttributesGetRequest) *admin.OrgAttributesGetResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.OrgAttributesGetResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.OrgAttributesGetRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_GetProjectAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_GetProjectAttributes) Return(_a0 *admin.ProjectAttributesGetResponse, _a1 error) *ResourceInterface_GetProjectAttributes {
	return &ResourceInterface_GetProjectAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnGetProjectAttributes(ctx context.Context, request admin.ProjectAttributesGetRequest) *ResourceInterface_GetProjectAttributes {
	c_call := _m.On("GetProjectAttributes", ctx, request)
	return &ResourceInterface_GetProjectAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnGetProjectAttributesMatch(matchers ...interface{}) *ResourceInterface_GetProjectAttributes {
	c_call := _m.On("GetProjectAttributes", matchers...)
	return &ResourceInterface_GetProjectAttributes{Call: c_call}
}

// GetProjectAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) GetProjectAttributes(ctx context.Context, request admin.ProjectAttributesGetRequest) (*admin.ProjectAttributesGetResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.ProjectAttributesGetResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.ProjectAttributesGetRequest) *admin.ProjectAttributesGetResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.ProjectAttributesGetResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.ProjectAttributesGetRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_GetProjectDomainAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_GetProjectDomainAttributes) Return(_a0 *admin.ProjectDomainAttributesGetResponse, _a1 error) *ResourceInterface_GetProjectDomainAttributes {
	return &ResourceInterface_GetProjectDomainAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnGetProjectDomainAttributes(ctx context.Context, request admin.ProjectDomainAttributesGetRequest) *ResourceInterface_GetProjectDomainAttributes {
	c_call := _m.On("GetProjectDomainAttributes", ctx, request)
	return &ResourceInterface_GetProjectDomainAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnGetProjectDomainAttributesMatch(matchers ...interface{}) *ResourceInterface_GetProjectDomainAttributes {
	c_call := _m.On("GetProjectDomainAttributes", matchers...)
	return &ResourceInterface_GetProjectDomainAttributes{Call: c_call}
}

// GetProjectDomainAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) GetProjectDomainAttributes(ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (*admin.ProjectDomainAttributesGetResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.ProjectDomainAttributesGetResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.ProjectDomainAttributesGetRequest) *admin.ProjectDomainAttributesGetResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.ProjectDomainAttributesGetResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.ProjectDomainAttributesGetRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_GetResource struct {
	*mock.Call
}

func (_m ResourceInterface_GetResource) Return(_a0 *interfaces.ResourceResponse, _a1 error) *ResourceInterface_GetResource {
	return &ResourceInterface_GetResource{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnGetResource(ctx context.Context, request interfaces.ResourceRequest) *ResourceInterface_GetResource {
	c_call := _m.On("GetResource", ctx, request)
	return &ResourceInterface_GetResource{Call: c_call}
}

func (_m *ResourceInterface) OnGetResourceMatch(matchers ...interface{}) *ResourceInterface_GetResource {
	c_call := _m.On("GetResource", matchers...)
	return &ResourceInterface_GetResource{Call: c_call}
}

// GetResource provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) GetResource(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *interfaces.ResourceResponse
	if rf, ok := ret.Get(0).(func(context.Context, interfaces.ResourceRequest) *interfaces.ResourceResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*interfaces.ResourceResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, interfaces.ResourceRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_GetWorkflowAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_GetWorkflowAttributes) Return(_a0 *admin.WorkflowAttributesGetResponse, _a1 error) *ResourceInterface_GetWorkflowAttributes {
	return &ResourceInterface_GetWorkflowAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnGetWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesGetRequest) *ResourceInterface_GetWorkflowAttributes {
	c_call := _m.On("GetWorkflowAttributes", ctx, request)
	return &ResourceInterface_GetWorkflowAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnGetWorkflowAttributesMatch(matchers ...interface{}) *ResourceInterface_GetWorkflowAttributes {
	c_call := _m.On("GetWorkflowAttributes", matchers...)
	return &ResourceInterface_GetWorkflowAttributes{Call: c_call}
}

// GetWorkflowAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) GetWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesGetRequest) (*admin.WorkflowAttributesGetResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.WorkflowAttributesGetResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.WorkflowAttributesGetRequest) *admin.WorkflowAttributesGetResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.WorkflowAttributesGetResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.WorkflowAttributesGetRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_ListAll struct {
	*mock.Call
}

func (_m ResourceInterface_ListAll) Return(_a0 *admin.ListMatchableAttributesResponse, _a1 error) *ResourceInterface_ListAll {
	return &ResourceInterface_ListAll{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnListAll(ctx context.Context, request admin.ListMatchableAttributesRequest) *ResourceInterface_ListAll {
	c_call := _m.On("ListAll", ctx, request)
	return &ResourceInterface_ListAll{Call: c_call}
}

func (_m *ResourceInterface) OnListAllMatch(matchers ...interface{}) *ResourceInterface_ListAll {
	c_call := _m.On("ListAll", matchers...)
	return &ResourceInterface_ListAll{Call: c_call}
}

// ListAll provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) ListAll(ctx context.Context, request admin.ListMatchableAttributesRequest) (*admin.ListMatchableAttributesResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.ListMatchableAttributesResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.ListMatchableAttributesRequest) *admin.ListMatchableAttributesResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.ListMatchableAttributesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.ListMatchableAttributesRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_UpdateOrgAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_UpdateOrgAttributes) Return(_a0 *admin.OrgAttributesUpdateResponse, _a1 error) *ResourceInterface_UpdateOrgAttributes {
	return &ResourceInterface_UpdateOrgAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnUpdateOrgAttributes(ctx context.Context, request admin.OrgAttributesUpdateRequest) *ResourceInterface_UpdateOrgAttributes {
	c_call := _m.On("UpdateOrgAttributes", ctx, request)
	return &ResourceInterface_UpdateOrgAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnUpdateOrgAttributesMatch(matchers ...interface{}) *ResourceInterface_UpdateOrgAttributes {
	c_call := _m.On("UpdateOrgAttributes", matchers...)
	return &ResourceInterface_UpdateOrgAttributes{Call: c_call}
}

// UpdateOrgAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) UpdateOrgAttributes(ctx context.Context, request admin.OrgAttributesUpdateRequest) (*admin.OrgAttributesUpdateResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.OrgAttributesUpdateResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.OrgAttributesUpdateRequest) *admin.OrgAttributesUpdateResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.OrgAttributesUpdateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.OrgAttributesUpdateRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_UpdateProjectAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_UpdateProjectAttributes) Return(_a0 *admin.ProjectAttributesUpdateResponse, _a1 error) *ResourceInterface_UpdateProjectAttributes {
	return &ResourceInterface_UpdateProjectAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnUpdateProjectAttributes(ctx context.Context, request admin.ProjectAttributesUpdateRequest) *ResourceInterface_UpdateProjectAttributes {
	c_call := _m.On("UpdateProjectAttributes", ctx, request)
	return &ResourceInterface_UpdateProjectAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnUpdateProjectAttributesMatch(matchers ...interface{}) *ResourceInterface_UpdateProjectAttributes {
	c_call := _m.On("UpdateProjectAttributes", matchers...)
	return &ResourceInterface_UpdateProjectAttributes{Call: c_call}
}

// UpdateProjectAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) UpdateProjectAttributes(ctx context.Context, request admin.ProjectAttributesUpdateRequest) (*admin.ProjectAttributesUpdateResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.ProjectAttributesUpdateResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.ProjectAttributesUpdateRequest) *admin.ProjectAttributesUpdateResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.ProjectAttributesUpdateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.ProjectAttributesUpdateRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_UpdateProjectDomainAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_UpdateProjectDomainAttributes) Return(_a0 *admin.ProjectDomainAttributesUpdateResponse, _a1 error) *ResourceInterface_UpdateProjectDomainAttributes {
	return &ResourceInterface_UpdateProjectDomainAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnUpdateProjectDomainAttributes(ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) *ResourceInterface_UpdateProjectDomainAttributes {
	c_call := _m.On("UpdateProjectDomainAttributes", ctx, request)
	return &ResourceInterface_UpdateProjectDomainAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnUpdateProjectDomainAttributesMatch(matchers ...interface{}) *ResourceInterface_UpdateProjectDomainAttributes {
	c_call := _m.On("UpdateProjectDomainAttributes", matchers...)
	return &ResourceInterface_UpdateProjectDomainAttributes{Call: c_call}
}

// UpdateProjectDomainAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) UpdateProjectDomainAttributes(ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (*admin.ProjectDomainAttributesUpdateResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.ProjectDomainAttributesUpdateResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.ProjectDomainAttributesUpdateRequest) *admin.ProjectDomainAttributesUpdateResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.ProjectDomainAttributesUpdateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.ProjectDomainAttributesUpdateRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ResourceInterface_UpdateWorkflowAttributes struct {
	*mock.Call
}

func (_m ResourceInterface_UpdateWorkflowAttributes) Return(_a0 *admin.WorkflowAttributesUpdateResponse, _a1 error) *ResourceInterface_UpdateWorkflowAttributes {
	return &ResourceInterface_UpdateWorkflowAttributes{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ResourceInterface) OnUpdateWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesUpdateRequest) *ResourceInterface_UpdateWorkflowAttributes {
	c_call := _m.On("UpdateWorkflowAttributes", ctx, request)
	return &ResourceInterface_UpdateWorkflowAttributes{Call: c_call}
}

func (_m *ResourceInterface) OnUpdateWorkflowAttributesMatch(matchers ...interface{}) *ResourceInterface_UpdateWorkflowAttributes {
	c_call := _m.On("UpdateWorkflowAttributes", matchers...)
	return &ResourceInterface_UpdateWorkflowAttributes{Call: c_call}
}

// UpdateWorkflowAttributes provides a mock function with given fields: ctx, request
func (_m *ResourceInterface) UpdateWorkflowAttributes(ctx context.Context, request admin.WorkflowAttributesUpdateRequest) (*admin.WorkflowAttributesUpdateResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.WorkflowAttributesUpdateResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.WorkflowAttributesUpdateRequest) *admin.WorkflowAttributesUpdateResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.WorkflowAttributesUpdateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.WorkflowAttributesUpdateRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
