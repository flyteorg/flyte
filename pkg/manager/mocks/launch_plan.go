package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateLaunchPlanFunc func(ctx context.Context, request admin.LaunchPlanCreateRequest) (
	*admin.LaunchPlanCreateResponse, error)
type UpdateLaunchPlanFunc func(ctx context.Context, request admin.LaunchPlanUpdateRequest) (
	*admin.LaunchPlanUpdateResponse, error)
type GetLaunchPlanFunc func(ctx context.Context, request admin.ObjectGetRequest) (
	*admin.LaunchPlan, error)
type GetActiveLaunchPlanFunc func(ctx context.Context, request admin.ActiveLaunchPlanRequest) (
	*admin.LaunchPlan, error)
type ListLaunchPlansFunc func(ctx context.Context, request admin.ResourceListRequest) (
	*admin.LaunchPlanList, error)
type ListLaunchPlanIdsFunc func(ctx context.Context, request admin.NamedEntityIdentifierListRequest) (
	*admin.NamedEntityIdentifierList, error)
type ListActiveLaunchPlansFunc func(ctx context.Context, request admin.ActiveLaunchPlanListRequest) (
	*admin.LaunchPlanList, error)

type MockLaunchPlanManager struct {
	createLaunchPlanFunc      CreateLaunchPlanFunc
	updateLaunchPlanFunc      UpdateLaunchPlanFunc
	getLaunchPlanFunc         GetLaunchPlanFunc
	getActiveLaunchPlanFunc   GetActiveLaunchPlanFunc
	listLaunchPlansFunc       ListLaunchPlansFunc
	listLaunchPlanIdsFunc     ListLaunchPlanIdsFunc
	listActiveLaunchPlansFunc ListActiveLaunchPlansFunc
}

func (r *MockLaunchPlanManager) SetCreateCallback(createFunction CreateLaunchPlanFunc) {
	r.createLaunchPlanFunc = createFunction
}

func (r *MockLaunchPlanManager) CreateLaunchPlan(
	ctx context.Context,
	request admin.LaunchPlanCreateRequest) (*admin.LaunchPlanCreateResponse, error) {
	if r.createLaunchPlanFunc != nil {
		return r.createLaunchPlanFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockLaunchPlanManager) SetUpdateLaunchPlan(updateFunction UpdateLaunchPlanFunc) {
	r.updateLaunchPlanFunc = updateFunction
}

func (r *MockLaunchPlanManager) UpdateLaunchPlan(ctx context.Context, request admin.LaunchPlanUpdateRequest) (
	*admin.LaunchPlanUpdateResponse, error) {
	if r.updateLaunchPlanFunc != nil {
		return r.updateLaunchPlanFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockLaunchPlanManager) GetLaunchPlan(ctx context.Context, request admin.ObjectGetRequest) (
	*admin.LaunchPlan, error) {
	if r.getLaunchPlanFunc != nil {
		return r.getLaunchPlanFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockLaunchPlanManager) SetGetActiveLaunchPlanCallback(plansFunc GetActiveLaunchPlanFunc) {
	r.getActiveLaunchPlanFunc = plansFunc
}

func (r *MockLaunchPlanManager) GetActiveLaunchPlan(ctx context.Context, request admin.ActiveLaunchPlanRequest) (
	*admin.LaunchPlan, error) {
	if r.getActiveLaunchPlanFunc != nil {
		return r.getActiveLaunchPlanFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockLaunchPlanManager) SetListLaunchPlansCallback(listLaunchPlansFunc ListLaunchPlansFunc) {
	r.listLaunchPlansFunc = listLaunchPlansFunc
}

func (r *MockLaunchPlanManager) ListLaunchPlans(ctx context.Context, request admin.ResourceListRequest) (
	*admin.LaunchPlanList, error) {
	if r.listLaunchPlansFunc != nil {
		return r.listLaunchPlansFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockLaunchPlanManager) SetListActiveLaunchPlansCallback(plansFunc ListActiveLaunchPlansFunc) {
	r.listActiveLaunchPlansFunc = plansFunc
}

func (r *MockLaunchPlanManager) ListActiveLaunchPlans(ctx context.Context, request admin.ActiveLaunchPlanListRequest) (
	*admin.LaunchPlanList, error) {
	if r.listActiveLaunchPlansFunc != nil {
		return r.listActiveLaunchPlansFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockLaunchPlanManager) ListLaunchPlanIds(ctx context.Context, request admin.NamedEntityIdentifierListRequest) (
	*admin.NamedEntityIdentifierList, error) {
	if r.listLaunchPlanIdsFunc != nil {
		return r.ListLaunchPlanIds(ctx, request)
	}
	return nil, nil
}

func NewMockLaunchPlanManager() interfaces.LaunchPlanInterface {
	return &MockLaunchPlanManager{}
}
