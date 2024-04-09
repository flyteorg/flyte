package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

type UpdateConfigurationInput struct {
	VersionToUpdate  string
	NewConfiguration *models.ConfigurationDocument
}

type ConfigurationDocumentRepoInterface interface {
	GetActive(ctx context.Context) (models.ConfigurationDocument, error)
	Update(ctx context.Context, input *UpdateConfigurationInput) error
	CreateOrUpdateResource(ctx context.Context, input models.Resource, updateConfigurationInput *UpdateConfigurationInput) error
}
