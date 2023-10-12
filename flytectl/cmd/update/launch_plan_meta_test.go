package update

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLaunchPlanMetadataCanBeActivated(t *testing.T) {
	testNamedEntityUpdate(core.ResourceType_LAUNCH_PLAN,
		/* setup */ func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity) {
			namedEntity.Metadata.State = admin.NamedEntityState_NAMED_ENTITY_ARCHIVED
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertCalled(
				t, "UpdateNamedEntity", s.Ctx,
				mock.MatchedBy(
					func(r *admin.NamedEntityUpdateRequest) bool {
						return r.GetMetadata().GetState() == admin.NamedEntityState_NAMED_ENTITY_ACTIVE
					}))
		})
}

func TestLaunchPlanMetadataCanBeArchived(t *testing.T) {
	testNamedEntityUpdate(core.ResourceType_LAUNCH_PLAN,
		/* setup */ func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity) {
			namedEntity.Metadata.State = admin.NamedEntityState_NAMED_ENTITY_ACTIVE
			config.Archive = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertCalled(
				t, "UpdateNamedEntity", s.Ctx,
				mock.MatchedBy(
					func(r *admin.NamedEntityUpdateRequest) bool {
						return r.GetMetadata().GetState() == admin.NamedEntityState_NAMED_ENTITY_ARCHIVED
					}))
		})
}

func TestLaunchPlanMetadataCannotBeActivatedAndArchivedAtTheSameTime(t *testing.T) {
	testNamedEntityUpdate(core.ResourceType_LAUNCH_PLAN,
		/* setup */ func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity) {
			config.Activate = true
			config.Archive = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "Specify either activate or archive")
			s.MockAdminClient.AssertNotCalled(t, "UpdateNamedEntity", mock.Anything, mock.Anything)
		})
}

func TestLaunchPlanMetadataUpdateDoesNothingWhenThereAreNoChanges(t *testing.T) {
	testNamedEntityUpdate(core.ResourceType_LAUNCH_PLAN,
		/* setup */ func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity) {
			namedEntity.Metadata.State = admin.NamedEntityState_NAMED_ENTITY_ACTIVE
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateNamedEntity", mock.Anything, mock.Anything)
		})
}

func TestLaunchPlanMetadataUpdateWithoutForceFlagFails(t *testing.T) {
	testNamedEntityUpdate(core.ResourceType_LAUNCH_PLAN,
		/* setup */ func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity) {
			namedEntity.Metadata.State = admin.NamedEntityState_NAMED_ENTITY_ARCHIVED
			config.Activate = true
			config.Force = false
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "update aborted by user")
			s.MockAdminClient.AssertNotCalled(t, "UpdateNamedEntity", mock.Anything, mock.Anything)
		})
}

func TestLaunchPlanMetadataUpdateDoesNothingWithDryRunFlag(t *testing.T) {
	testNamedEntityUpdate(core.ResourceType_LAUNCH_PLAN,
		/* setup */ func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity) {
			namedEntity.Metadata.State = admin.NamedEntityState_NAMED_ENTITY_ARCHIVED
			config.Activate = true
			config.DryRun = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateNamedEntity", mock.Anything, mock.Anything)
		})
}

func TestForceFlagIsIgnoredWithDryRunDuringLaunchPlanMetadataUpdate(t *testing.T) {
	t.Run("without --force", func(t *testing.T) {
		testNamedEntityUpdate(core.ResourceType_LAUNCH_PLAN,
			/* setup */ func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity) {
				namedEntity.Metadata.State = admin.NamedEntityState_NAMED_ENTITY_ARCHIVED
				config.Activate = true

				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.MockAdminClient.AssertNotCalled(t, "UpdateNamedEntity", mock.Anything, mock.Anything)
			})
	})

	t.Run("with --force", func(t *testing.T) {
		testNamedEntityUpdate(core.ResourceType_LAUNCH_PLAN,
			/* setup */ func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity) {
				namedEntity.Metadata.State = admin.NamedEntityState_NAMED_ENTITY_ARCHIVED
				config.Activate = true

				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.MockAdminClient.AssertNotCalled(t, "UpdateNamedEntity", mock.Anything, mock.Anything)
			})
	})
}

func TestLaunchPlanMetadataUpdateFailsWhenLaunchPlanDoesNotExist(t *testing.T) {
	testNamedEntityUpdateWithMockSetup(
		core.ResourceType_LAUNCH_PLAN,
		/* mockSetup */ func(s *testutils.TestStruct, namedEntity *admin.NamedEntity) {
			s.MockAdminClient.
				OnGetNamedEntityMatch(
					s.Ctx,
					mock.MatchedBy(func(r *admin.NamedEntityGetRequest) bool {
						return r.ResourceType == namedEntity.ResourceType &&
							cmp.Equal(r.Id, namedEntity.Id)
					})).
				Return(nil, ext.NewNotFoundError("named entity not found"))
			s.MockAdminClient.
				OnUpdateNamedEntityMatch(s.Ctx, mock.Anything).
				Return(&admin.NamedEntityUpdateResponse{}, nil)
		},
		/* setup */ nil,
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Error(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateNamedEntity", mock.Anything, mock.Anything)
		},
	)
}

func TestLaunchPlanMetadataUpdateFailsWhenAdminClientFails(t *testing.T) {
	testNamedEntityUpdateWithMockSetup(
		core.ResourceType_LAUNCH_PLAN,
		/* mockSetup */ func(s *testutils.TestStruct, namedEntity *admin.NamedEntity) {
			s.MockAdminClient.
				OnGetNamedEntityMatch(
					s.Ctx,
					mock.MatchedBy(func(r *admin.NamedEntityGetRequest) bool {
						return r.ResourceType == namedEntity.ResourceType &&
							cmp.Equal(r.Id, namedEntity.Id)
					})).
				Return(namedEntity, nil)
			s.MockAdminClient.
				OnUpdateNamedEntityMatch(s.Ctx, mock.Anything).
				Return(nil, fmt.Errorf("network error"))
		},
		/* setup */ func(s *testutils.TestStruct, config *NamedEntityConfig, namedEntity *admin.NamedEntity) {
			namedEntity.Metadata.State = admin.NamedEntityState_NAMED_ENTITY_ARCHIVED
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Error(t, err)
			s.MockAdminClient.AssertCalled(t, "UpdateNamedEntity", mock.Anything, mock.Anything)
		},
	)
}

func TestLaunchPlanMetadataUpdateRequiresLaunchPlanName(t *testing.T) {
	s := testutils.Setup()
	config := &NamedEntityConfig{}

	err := getUpdateLPMetaFunc(config)(s.Ctx, nil, s.CmdCtx)

	assert.ErrorContains(t, err, "launch plan name wasn't passed")
}
