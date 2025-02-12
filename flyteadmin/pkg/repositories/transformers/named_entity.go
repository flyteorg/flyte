package transformers

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func CreateNamedEntityModel(request *admin.NamedEntityUpdateRequest) models.NamedEntity {
	stateInt := int32(request.GetMetadata().GetState())
	return models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: request.GetResourceType(),
			Project:      request.GetId().GetProject(),
			Domain:       request.GetId().GetDomain(),
			Name:         request.GetId().GetName(),
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: request.GetMetadata().GetDescription(),
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
	var entityState int32
	if model.State != nil {
		entityState = *model.State
	}
	return admin.NamedEntity{
		ResourceType: model.ResourceType,
		Id: &admin.NamedEntityIdentifier{
			Project: model.Project,
			Domain:  model.Domain,
			Name:    model.Name,
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: model.Description,
			State:       admin.NamedEntityState(entityState),
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
