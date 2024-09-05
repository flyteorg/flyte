package transformers

import (
	"context"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// CreateDescriptionEntityModel Transforms a TaskCreateRequest to a Description entity model
func CreateDescriptionEntityModel(
	descriptionEntity *admin.DescriptionEntity,
	id *core.Identifier) (*models.DescriptionEntity, error) {
	ctx := context.Background()
	if descriptionEntity == nil {
		return nil, nil
	}

	var longDescriptionBytes []byte
	var sourceCode models.SourceCode
	var err error

	if descriptionEntity.LongDescription != nil {
		longDescriptionBytes, err = proto.Marshal(descriptionEntity.LongDescription)
		if err != nil {
			logger.Errorf(ctx, "Failed to marshal LongDescription with error: %v", err)
			return nil, err
		}
	}

	if descriptionEntity.LongDescription != nil {
		longDescriptionBytes, err = proto.Marshal(descriptionEntity.LongDescription)
		if err != nil {
			logger.Errorf(ctx, "Failed to marshal LongDescription with error: %v", err)
			return nil, err
		}
	}
	if descriptionEntity.SourceCode != nil {
		sourceCode = models.SourceCode{Link: descriptionEntity.SourceCode.Link}
	}

	return &models.DescriptionEntity{
		DescriptionEntityKey: models.DescriptionEntityKey{
			ResourceType: id.ResourceType,
			Project:      id.Project,
			Domain:       id.Domain,
			Name:         id.Name,
			Version:      id.Version,
		},
		ShortDescription: descriptionEntity.ShortDescription,
		LongDescription:  longDescriptionBytes,
		SourceCode:       sourceCode,
	}, nil
}

func FromDescriptionEntityModel(descriptionEntityModel models.DescriptionEntity) (*admin.DescriptionEntity, error) {

	longDescription := admin.Description{}
	err := proto.Unmarshal(descriptionEntityModel.LongDescription, &longDescription)
	if err != nil {
		return nil, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal longDescription")
	}

	return &admin.DescriptionEntity{
		Id: &core.Identifier{
			ResourceType: descriptionEntityModel.ResourceType,
			Project:      descriptionEntityModel.Project,
			Domain:       descriptionEntityModel.Domain,
			Name:         descriptionEntityModel.Name,
			Version:      descriptionEntityModel.Version,
		},
		ShortDescription: descriptionEntityModel.ShortDescription,
		LongDescription:  &longDescription,
		SourceCode:       &admin.SourceCode{Link: descriptionEntityModel.Link},
	}, nil
}

func FromDescriptionEntityModels(descriptionEntityModels []models.DescriptionEntity) ([]*admin.DescriptionEntity, error) {
	descriptionEntities := make([]*admin.DescriptionEntity, len(descriptionEntityModels))
	for idx, descriptionEntityModel := range descriptionEntityModels {
		descriptionEntity, err := FromDescriptionEntityModel(descriptionEntityModel)
		if err != nil {
			return nil, err
		}
		descriptionEntities[idx] = descriptionEntity
	}
	return descriptionEntities, nil
}
