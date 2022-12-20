package tests

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestCreateLaunchPlanHappyCase(t *testing.T) {
	ctx := context.Background()

	mockLaunchPlanManager := mocks.MockLaunchPlanManager{}
	mockLaunchPlanManager.SetCreateCallback(
		func(ctx context.Context,
			request admin.LaunchPlanCreateRequest) (*admin.LaunchPlanCreateResponse, error) {
			return &admin.LaunchPlanCreateResponse{}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		launchPlanManager: &mockLaunchPlanManager,
	})

	resp, err := mockServer.CreateLaunchPlan(ctx, &admin.LaunchPlanCreateRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Name:         "Name",
			Domain:       "Domain",
			Project:      "Project",
			Version:      "Version",
		},
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
}

func TestCreateLaunchPlanError(t *testing.T) {
	ctx := context.Background()

	mockLaunchPlanManager := mocks.MockLaunchPlanManager{}
	mockLaunchPlanManager.SetCreateCallback(
		func(ctx context.Context,
			request admin.LaunchPlanCreateRequest) (*admin.LaunchPlanCreateResponse, error) {
			return nil, errors.GetMissingEntityError(core.ResourceType_LAUNCH_PLAN.String(), request.Id)
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		launchPlanManager: &mockLaunchPlanManager,
	})

	resp, err := mockServer.CreateLaunchPlan(ctx, &admin.LaunchPlanCreateRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Name:         "Name",
			Domain:       "Domain",
			Project:      "Project",
			Version:      "Version",
		},
	})
	assert.Nil(t, resp)
	assert.EqualError(t, err, "missing entity of type LAUNCH_PLAN with "+
		"identifier resource_type:LAUNCH_PLAN project:\"Project\" domain:\"Domain\" name:\"Name\" version:\"Version\" ")
}

func TestGetActiveLaunchPlan(t *testing.T) {
	ctx := context.Background()

	mockLaunchPlanManager := mocks.MockLaunchPlanManager{}
	mockLaunchPlanManager.SetGetActiveLaunchPlanCallback(
		func(ctx context.Context,
			request admin.ActiveLaunchPlanRequest) (*admin.LaunchPlan, error) {
			return &admin.LaunchPlan{}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		launchPlanManager: &mockLaunchPlanManager,
	})

	resp, err := mockServer.GetActiveLaunchPlan(ctx, &admin.ActiveLaunchPlanRequest{
		Id: &admin.NamedEntityIdentifier{
			Name:    "Name",
			Domain:  "Domain",
			Project: "Project",
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestGetActiveLaunchPlan_Error(t *testing.T) {
	ctx := context.Background()

	mockLaunchPlanManager := mocks.MockLaunchPlanManager{}
	mockLaunchPlanManager.SetGetActiveLaunchPlanCallback(
		func(ctx context.Context,
			request admin.ActiveLaunchPlanRequest) (*admin.LaunchPlan, error) {
			return nil, errors.GetInvalidInputError("invalid input")
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		launchPlanManager: &mockLaunchPlanManager,
	})

	resp, err := mockServer.GetActiveLaunchPlan(ctx, &admin.ActiveLaunchPlanRequest{
		Id: &admin.NamedEntityIdentifier{
			Name:    "Name",
			Domain:  "Domain",
			Project: "",
		},
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListActiveLaunchPlans(t *testing.T) {
	ctx := context.Background()

	mockLaunchPlanManager := mocks.MockLaunchPlanManager{}
	mockLaunchPlanManager.SetListActiveLaunchPlansCallback(
		func(ctx context.Context,
			request admin.ActiveLaunchPlanListRequest) (*admin.LaunchPlanList, error) {
			return &admin.LaunchPlanList{}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		launchPlanManager: &mockLaunchPlanManager,
	})

	resp, err := mockServer.ListActiveLaunchPlans(ctx, &admin.ActiveLaunchPlanListRequest{
		Project: "project",
		Domain:  "domain",
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestListActiveLaunchPlans_Error(t *testing.T) {
	ctx := context.Background()

	mockLaunchPlanManager := mocks.MockLaunchPlanManager{}
	mockLaunchPlanManager.SetListActiveLaunchPlansCallback(
		func(ctx context.Context,
			request admin.ActiveLaunchPlanListRequest) (*admin.LaunchPlanList, error) {
			return nil, errors.GetInvalidInputError("oops")
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		launchPlanManager: &mockLaunchPlanManager,
	})

	resp, err := mockServer.ListActiveLaunchPlans(ctx, &admin.ActiveLaunchPlanListRequest{
		Project: "project",
		Domain:  "domain",
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}
