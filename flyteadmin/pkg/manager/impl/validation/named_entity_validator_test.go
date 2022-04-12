package validation

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestValidateNamedEntityGetRequest(t *testing.T) {
	assert.Nil(t, ValidateNamedEntityGetRequest(admin.NamedEntityGetRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	}))

	assert.NotNil(t, ValidateNamedEntityGetRequest(admin.NamedEntityGetRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Domain: "domain",
			Name:   "name",
		},
	}))

	assert.NotNil(t, ValidateNamedEntityGetRequest(admin.NamedEntityGetRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Name:    "name",
		},
	}))

	assert.NotNil(t, ValidateNamedEntityGetRequest(admin.NamedEntityGetRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
		},
	}))

	assert.NotNil(t, ValidateNamedEntityGetRequest(admin.NamedEntityGetRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	}))
}

func TestValidateNamedEntityUpdateRequest(t *testing.T) {
	assert.Nil(t, ValidateNamedEntityUpdateRequest(admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: "description",
		},
	}))

	assert.NotNil(t, ValidateNamedEntityUpdateRequest(admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Domain: "domain",
			Name:   "name",
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: "description",
		},
	}))

	assert.NotNil(t, ValidateNamedEntityUpdateRequest(admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Name:    "name",
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: "description",
		},
	}))

	assert.NotNil(t, ValidateNamedEntityUpdateRequest(admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: "description",
		},
	}))

	assert.NotNil(t, ValidateNamedEntityUpdateRequest(admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	}))
	assert.Equal(t, codes.InvalidArgument, ValidateNamedEntityUpdateRequest(admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Metadata: &admin.NamedEntityMetadata{
			State: admin.NamedEntityState_NAMED_ENTITY_ARCHIVED,
		},
	}).(errors.FlyteAdminError).Code())
	assert.Nil(t, ValidateNamedEntityUpdateRequest(admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Metadata: &admin.NamedEntityMetadata{
			State: admin.NamedEntityState_NAMED_ENTITY_ARCHIVED,
		},
	}))
	assert.Nil(t, ValidateNamedEntityUpdateRequest(admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_TASK,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Metadata: &admin.NamedEntityMetadata{
			State: admin.NamedEntityState_NAMED_ENTITY_ARCHIVED,
		},
	}))
}

func TestValidateNamedEntityListRequest(t *testing.T) {
	assert.Nil(t, ValidateNamedEntityListRequest(admin.NamedEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
		Limit:        2,
	}))

	assert.NotNil(t, ValidateNamedEntityListRequest(admin.NamedEntityListRequest{
		Project: "project",
		Domain:  "domain",
		Limit:   2,
	}))

	assert.NotNil(t, ValidateNamedEntityListRequest(admin.NamedEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Domain:       "domain",
		Limit:        2,
	}))

	assert.NotNil(t, ValidateNamedEntityListRequest(admin.NamedEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Limit:        2,
	}))

	assert.NotNil(t, ValidateNamedEntityListRequest(admin.NamedEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
	}))
}
