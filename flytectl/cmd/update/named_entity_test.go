package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/mock"
)

func testNamedEntityUpdate(
	resourceType core.ResourceType,
	setup func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity),
	asserter func(s *testutils.TestStruct, err error),
) {
	testNamedEntityUpdateWithMockSetup(
		resourceType,
		/* mockSetup */ func(s *testutils.TestStruct, namedEntity *admin.NamedEntity) {
			s.MockAdminClient.
				OnGetNamedEntityMatch(
					s.Ctx,
					mock.Anything).
				Return(namedEntity, nil)
			s.MockAdminClient.
				OnUpdateNamedEntityMatch(s.Ctx, mock.Anything).
				Return(&admin.NamedEntityUpdateResponse{}, nil)
		},
		setup,
		asserter,
	)
}

func testNamedEntityUpdateWithMockSetup(
	resourceType core.ResourceType,
	mockSetup func(s *testutils.TestStruct, namedEntity *admin.NamedEntity),
	setup func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	defer s.TearDown()

	config := &NamedEntityConfig{}
	target := newTestNamedEntity(resourceType)

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	if setup != nil {
		setup(&s, config, target)
		defer s.TearDown()
	}

	updateMetadataFactory := getUpdateMetadataFactory(resourceType)

	args := []string{target.Id.Name}
	err := updateMetadataFactory(config)(s.Ctx, args, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}
}

func newTestNamedEntity(resourceType core.ResourceType) *admin.NamedEntity {
	return &admin.NamedEntity{
		Id: &admin.NamedEntityIdentifier{
			Name:    testutils.RandomName(12),
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		},
		ResourceType: resourceType,
		Metadata: &admin.NamedEntityMetadata{
			State:       admin.NamedEntityState_NAMED_ENTITY_ACTIVE,
			Description: testutils.RandomName(50),
		},
	}
}

func getUpdateMetadataFactory(resourceType core.ResourceType) func(namedEntityConfig *NamedEntityConfig) func(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	switch resourceType {
	case core.ResourceType_LAUNCH_PLAN:
		return getUpdateLPMetaFunc
	case core.ResourceType_TASK:
		return getUpdateTaskFunc
	case core.ResourceType_WORKFLOW:
		return getUpdateWorkflowFunc
	}

	panic(fmt.Sprintf("no known mapping exists between resource type %s and "+
		"corresponding update metadata factory function", resourceType))
}
