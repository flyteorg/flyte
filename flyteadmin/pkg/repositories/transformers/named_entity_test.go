package transformers

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestCreateNamedEntityModel(t *testing.T) {

	model := CreateNamedEntityModel(&admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: "description",
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
		},
	}, model)
}

func TestFromNamedEntityModel(t *testing.T) {
	model := models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: "description",
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
