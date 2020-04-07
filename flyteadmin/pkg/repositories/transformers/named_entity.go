package transformers

import (
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

func CreateNamedEntityModel(request *admin.NamedEntityUpdateRequest) models.NamedEntity {
	stateInt := int32(request.Metadata.State)
	return models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: request.ResourceType,
			Project:      request.Id.Project,
			Domain:       request.Id.Domain,
			Name:         request.Id.Name,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: request.Metadata.Description,
			State:       &stateInt,
		},
	}
}

func FromNamedEntityMetadataFields(metadata models.NamedEntityMetadataFields) admin.NamedEntityMetadata {
	return admin.NamedEntityMetadata{
		Description: metadata.Description,
	}
}

func FromNamedEntityModel(model models.NamedEntity) admin.NamedEntity {
	return admin.NamedEntity{
		ResourceType: model.ResourceType,
		Id: &admin.NamedEntityIdentifier{
			Project: model.Project,
			Domain:  model.Domain,
			Name:    model.Name,
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: model.Description,
		},
	}
}

func FromNamedEntityModels(models []models.NamedEntity) []*admin.NamedEntity {
	namedEntities := make([]*admin.NamedEntity, len(models))
	for idx, model := range models {
		namedEntity := FromNamedEntityModel(model)
		namedEntities[idx] = &namedEntity
	}
	return namedEntities
}
