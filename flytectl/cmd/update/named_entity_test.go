package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

func TestNamedEntity(t *testing.T) {
	s := testutils.Setup()
	s.MockAdminClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(&admin.NamedEntityUpdateResponse{}, nil)
	namedEntityConfig = &NamedEntityConfig{Archive: false, Activate: true, Description: "named entity description"}
	assert.Nil(t, namedEntityConfig.UpdateNamedEntity(s.Ctx, "namedEntity", "project", "domain",
		core.ResourceType_WORKFLOW, s.CmdCtx))
	namedEntityConfig = &NamedEntityConfig{Archive: true, Activate: false, Description: "named entity description"}
	assert.Nil(t, namedEntityConfig.UpdateNamedEntity(s.Ctx, "namedEntity", "project", "domain",
		core.ResourceType_WORKFLOW, s.CmdCtx))
}

func TestNamedEntityValidationFailure(t *testing.T) {
	s := testutils.Setup()
	namedEntityConfig := &NamedEntityConfig{Archive: true, Activate: true, Description: "named entity description"}
	assert.NotNil(t, namedEntityConfig.UpdateNamedEntity(s.Ctx, "namedEntity", "project", "domain",
		core.ResourceType_WORKFLOW, s.CmdCtx))
}

func TestNamedEntityFailure(t *testing.T) {
	s := testutils.Setup()
	namedEntityConfig := &NamedEntityConfig{Archive: true, Activate: true, Description: "named entity description"}
	s.MockAdminClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to update"))
	assert.NotNil(t, namedEntityConfig.UpdateNamedEntity(s.Ctx, "namedEntity", "project", "domain",
		core.ResourceType_WORKFLOW, s.CmdCtx))
}
