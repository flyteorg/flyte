package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

type MockConfigurationRepo struct {
}

func (r *MockConfigurationRepo) GetActive(ctx context.Context) (models.Configuration, error) {
	return models.Configuration{}, nil
}

func (r *MockConfigurationRepo) EraseActiveAndCreate(ctx context.Context, versionToUpdate string, newConfiguration models.Configuration) error {
	return nil
}

func NewMockConfigurationRepo() interfaces.ConfigurationRepoInterface {
	return &MockConfigurationRepo{}
}
