package transformers

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestCreateNamedEntityModel(t *testing.T) {
	state := int32(admin.NamedEntityState_NAMED_ENTITY_ACTIVE)
	trueVal := true
	falseVal := false
	tests := []struct {
		name               string
		inputHasTrigger    *wrapperspb.BoolValue
		expectedHasTrigger *bool
	}{
		{
			name:               "has_trigger nil",
			inputHasTrigger:    nil,
			expectedHasTrigger: nil,
		},
		{
			name:               "has_trigger true",
			inputHasTrigger:    wrapperspb.Bool(true),
			expectedHasTrigger: &trueVal,
		},
		{
			name:               "has_trigger false",
			inputHasTrigger:    wrapperspb.Bool(false),
			expectedHasTrigger: &falseVal,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := CreateNamedEntityModel(&admin.NamedEntityUpdateRequest{
				ResourceType: core.ResourceType_WORKFLOW,
				Id: &admin.NamedEntityIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
					Org:     testOrg,
				},
				Metadata: &admin.NamedEntityMetadata{
					Description: "description",
					State:       admin.NamedEntityState_NAMED_ENTITY_ACTIVE,
					HasTrigger:  test.inputHasTrigger,
				},
			})

			assert.Equal(t, models.NamedEntity{
				NamedEntityKey: models.NamedEntityKey{
					ResourceType: core.ResourceType_WORKFLOW,
					Project:      "project",
					Domain:       "domain",
					Name:         "name",
					Org:          testOrg,
				},
				NamedEntityMetadataFields: models.NamedEntityMetadataFields{
					Description: "description",
					State:       &state,
					HasTrigger:  test.expectedHasTrigger,
				},
			}, model)
		})
	}
}

func TestFromNamedEntityModel(t *testing.T) {
	entityState := int32(1)
	trueVal := true
	model := models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
			Org:          testOrg,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: "description",
			State:       &entityState,
			HasTrigger:  &trueVal,
		},
	}

	namedEntity := FromNamedEntityModel(model)
	assert.True(t, proto.Equal(&admin.NamedEntity{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Org:     testOrg,
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: "description",
			State:       admin.NamedEntityState_NAMED_ENTITY_ARCHIVED,
			HasTrigger:  wrapperspb.Bool(true),
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
