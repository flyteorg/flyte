package transformers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestCreateNamedEntityModel(t *testing.T) {
	state := int32(admin.NamedEntityState_NAMED_ENTITY_ACTIVE)
	model := CreateNamedEntityModel(&admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: "description",
			State:       admin.NamedEntityState_NAMED_ENTITY_ACTIVE,
		},
	})

	assert.Equal(t, models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: "description",
			State:       &state,
		},
	}, model)
}

func TestFromNamedEntityModel(t *testing.T) {
	entityState := int32(1)
	model := models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: "description",
			State:       &entityState,
		},
	}

	namedEntity := FromNamedEntityModel(model)
	assert.True(t, proto.Equal(&admin.NamedEntity{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: "description",
			State:       admin.NamedEntityState_NAMED_ENTITY_ARCHIVED,
		},
	}, &namedEntity))
}

func TestFromNamedEntityMetadataFields(t *testing.T) {
	model := models.NamedEntityMetadataFields{
		Description: "description",
	}

	metadata := FromNamedEntityMetadataFields(model)
	assert.True(t, proto.Equal(&admin.NamedEntityMetadata{
		Description: "description",
	}, &metadata))
}
