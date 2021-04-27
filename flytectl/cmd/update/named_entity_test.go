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

func NamedEntitySetup() {
	ctx = testutils.Ctx
	cmdCtx = testutils.CmdCtx
	mockClient = testutils.MockClient
}

func TestNamedEntity(t *testing.T) {
	testutils.Setup()
	NamedEntitySetup()
	mockClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(&admin.NamedEntityUpdateResponse{}, nil)
	namedEntityConfig = &NamedEntityConfig{Archive: false, Activate: true, Description: "named entity description"}
	assert.Nil(t, namedEntityConfig.UpdateNamedEntity(ctx, "namedEntity", "project", "domain",
		core.ResourceType_WORKFLOW, cmdCtx))
	namedEntityConfig = &NamedEntityConfig{Archive: true, Activate: false, Description: "named entity description"}
	assert.Nil(t, namedEntityConfig.UpdateNamedEntity(ctx, "namedEntity", "project", "domain",
		core.ResourceType_WORKFLOW, cmdCtx))
}

func TestNamedEntityValidationFailure(t *testing.T) {
	testutils.Setup()
	NamedEntitySetup()
	namedEntityConfig = &NamedEntityConfig{Archive: true, Activate: true, Description: "named entity description"}
	assert.NotNil(t, namedEntityConfig.UpdateNamedEntity(ctx, "namedEntity", "project", "domain",
		core.ResourceType_WORKFLOW, cmdCtx))
}

func TestNamedEntityFailure(t *testing.T) {
	testutils.Setup()
	NamedEntitySetup()
	namedEntityConfig = &NamedEntityConfig{Archive: true, Activate: true, Description: "named entity description"}
	mockClient.OnUpdateNamedEntityMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to update"))
	assert.NotNil(t, namedEntityConfig.UpdateNamedEntity(ctx, "namedEntity", "project", "domain",
		core.ResourceType_WORKFLOW, cmdCtx))
}
