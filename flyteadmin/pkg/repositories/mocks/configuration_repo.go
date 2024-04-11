package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

type GetActiveFunc func(ctx context.Context) (models.ConfigurationDocument, error)
type UpdateFunc func(ctx context.Context, input *interfaces.UpdateConfigurationInput) error
type CreateFunc func(ctx context.Context, input *models.ConfigurationDocument) error

type MockConfigurationRepo struct {
	getActiveFunction GetActiveFunc
	updateFunction    UpdateFunc
	createFunction    CreateFunc
}

func (m *MockConfigurationRepo) GetActive(ctx context.Context) (models.ConfigurationDocument, error) {
	if m.getActiveFunction != nil {
		return m.getActiveFunction(ctx)
	}
	return models.ConfigurationDocument{}, nil
}

func (m *MockConfigurationRepo) SetGetActiveCallback(getActiveFunction GetActiveFunc) {
	m.getActiveFunction = getActiveFunction
}

func (m *MockConfigurationRepo) Update(ctx context.Context, input *interfaces.UpdateConfigurationInput) error {
	if m.updateFunction != nil {
		return m.updateFunction(ctx, input)
	}
	return nil
}

func (m *MockConfigurationRepo) SetUpdateCallback(updateFunction UpdateFunc) {
	m.updateFunction = updateFunction
}

func (m *MockConfigurationRepo) Create(ctx context.Context, input *models.ConfigurationDocument) error {
	if m.createFunction != nil {
		return m.createFunction(ctx, input)
	}
	return nil
}

func (m *MockConfigurationRepo) SetCreateCallback(createFunction CreateFunc) {
	m.createFunction = createFunction
}

func NewMockConfigurationRepo() interfaces.ConfigurationDocumentRepoInterface {
	return &MockConfigurationRepo{}
}
