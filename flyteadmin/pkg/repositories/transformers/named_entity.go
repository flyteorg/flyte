package transformers

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func CreateNamedEntityModel(request *admin.NamedEntityUpdateRequest) models.NamedEntity {
	stateInt := int32(request.Metadata.State)

	var hasTrigger *bool
	if request.Metadata.HasTrigger != nil {
		hasTrigger = &request.Metadata.HasTrigger.Value
	}

	return models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: request.ResourceType,
			Project:      request.Id.Project,
			Domain:       request.Id.Domain,
			Name:         request.Id.Name,
			Org:          request.Id.Org,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: request.Metadata.Description,
			State:       &stateInt,
			HasTrigger:  hasTrigger,
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

	var hasTrigger *wrapperspb.BoolValue
	if model.HasTrigger != nil {
		hasTrigger = wrapperspb.Bool(*model.HasTrigger)
	}

	return admin.NamedEntity{
		ResourceType: model.ResourceType,
		Id: &admin.NamedEntityIdentifier{
			Project: model.Project,
			Domain:  model.Domain,
			Name:    model.Name,
			Org:     model.Org,
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: model.Description,
			State:       admin.NamedEntityState(entityState),
			HasTrigger:  hasTrigger,
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
