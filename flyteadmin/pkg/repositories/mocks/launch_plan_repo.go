// Mock implementation of a workflow repo to be used for tests.
package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type CreateLaunchPlanFunc func(input models.LaunchPlan) error
type UpdateLaunchPlanFunc func(input models.LaunchPlan) error
type SetActiveLaunchPlanFunc func(toEnable models.LaunchPlan, toDisable *models.LaunchPlan) error
type GetLaunchPlanFunc func(input interfaces.Identifier) (models.LaunchPlan, error)
type ListLaunchPlanFunc func(input interfaces.ListResourceInput) (interfaces.LaunchPlanCollectionOutput, error)
type ListLaunchPlanIdentifiersFunc func(input interfaces.ListResourceInput) (
	interfaces.LaunchPlanCollectionOutput, error)

type MockLaunchPlanRepo struct {
	createFunction    CreateLaunchPlanFunc
	updateFunction    UpdateLaunchPlanFunc
	setActiveFunction SetActiveLaunchPlanFunc
	getFunction       GetLaunchPlanFunc
	listFunction      ListLaunchPlanFunc
	listIdsFunction   ListLaunchPlanIdentifiersFunc
}

func (r *MockLaunchPlanRepo) Create(ctx context.Context, input models.LaunchPlan) error {
	if r.createFunction != nil {
		return r.createFunction(input)
	}
	return nil
}

func (r *MockLaunchPlanRepo) SetCreateCallback(createFunction CreateLaunchPlanFunc) {
	r.createFunction = createFunction
}

func (r *MockLaunchPlanRepo) Update(ctx context.Context, launchPlan models.LaunchPlan) error {
	if r.updateFunction != nil {
		return r.updateFunction(launchPlan)
	}
	return nil
}

func (r *MockLaunchPlanRepo) SetUpdateCallback(updateFunction UpdateLaunchPlanFunc) {
	r.updateFunction = updateFunction
}

func (r *MockLaunchPlanRepo) SetActive(
	ctx context.Context, toEnable models.LaunchPlan, toDisable *models.LaunchPlan) error {
	if r.setActiveFunction != nil {
		return r.setActiveFunction(toEnable, toDisable)
	}
	return nil
}

func (r *MockLaunchPlanRepo) SetSetActiveCallback(setActiveFunction SetActiveLaunchPlanFunc) {
	r.setActiveFunction = setActiveFunction
}

func (r *MockLaunchPlanRepo) Get(
	ctx context.Context, input interfaces.Identifier) (models.LaunchPlan, error) {
	if r.getFunction != nil {
		return r.getFunction(input)
	}
	return models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: input.Project,
			Domain:  input.Domain,
			Name:    input.Name,
			Version: input.Version,
		},
	}, nil
}

func (r *MockLaunchPlanRepo) SetGetCallback(getFunction GetLaunchPlanFunc) {
	r.getFunction = getFunction
}

func (r *MockLaunchPlanRepo) List(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.LaunchPlanCollectionOutput, error) {
	if r.listFunction != nil {
		return r.listFunction(input)
	}
	return interfaces.LaunchPlanCollectionOutput{}, nil
}

func (r *MockLaunchPlanRepo) SetListCallback(listFunction ListLaunchPlanFunc) {
	r.listFunction = listFunction
}

func (r *MockLaunchPlanRepo) ListLaunchPlanIdentifiers(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.LaunchPlanCollectionOutput, error) {
	if r.listIdsFunction != nil {
		return r.listIdsFunction(input)
	}
	return interfaces.LaunchPlanCollectionOutput{}, nil
}

func (r *MockLaunchPlanRepo) SetListLaunchPlanIdentifiersCallback(fn ListLaunchPlanIdentifiersFunc) {
	r.listIdsFunction = fn
}

func NewMockLaunchPlanRepo() interfaces.LaunchPlanRepoInterface {
	return &MockLaunchPlanRepo{}
}
