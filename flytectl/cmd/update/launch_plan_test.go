package update

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/launchplan"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLaunchPlanCanBeActivated(t *testing.T) {
	testLaunchPlanUpdate(
		/* setup */ func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan) {
			launchplan.Closure.State = admin.LaunchPlanState_INACTIVE
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertCalled(
				t, "UpdateLaunchPlan", s.Ctx,
				mock.MatchedBy(
					func(r *admin.LaunchPlanUpdateRequest) bool {
						return r.State == admin.LaunchPlanState_ACTIVE
					}))
		})
}

func TestLaunchPlanCanBeArchived(t *testing.T) {
	testLaunchPlanUpdate(
		/* setup */ func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan) {
			launchplan.Closure.State = admin.LaunchPlanState_ACTIVE
			config.Archive = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertCalled(
				t, "UpdateLaunchPlan", s.Ctx,
				mock.MatchedBy(
					func(r *admin.LaunchPlanUpdateRequest) bool {
						return r.State == admin.LaunchPlanState_INACTIVE
					}))
		})
}

func TestLaunchPlanCannotBeActivatedAndArchivedAtTheSameTime(t *testing.T) {
	testLaunchPlanUpdate(
		/* setup */ func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan) {
			config.Activate = true
			config.Archive = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "Specify either activate or archive")
			s.MockAdminClient.AssertNotCalled(t, "UpdateLaunchPlan", mock.Anything, mock.Anything)
		})
}

func TestLaunchPlanUpdateDoesNothingWhenThereAreNoChanges(t *testing.T) {
	testLaunchPlanUpdate(
		/* setup */ func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan) {
			launchplan.Closure.State = admin.LaunchPlanState_ACTIVE
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateLaunchPlan", mock.Anything, mock.Anything)
		})
}

func TestLaunchPlanUpdateWithoutForceFlagFails(t *testing.T) {
	testLaunchPlanUpdate(
		/* setup */ func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan) {
			launchplan.Closure.State = admin.LaunchPlanState_INACTIVE
			config.Activate = true
			config.Force = false
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "update aborted by user")
			s.MockAdminClient.AssertNotCalled(t, "UpdateLaunchPlan", mock.Anything, mock.Anything)
		})
}

func TestLaunchPlanUpdateDoesNothingWithDryRunFlag(t *testing.T) {
	testLaunchPlanUpdate(
		/* setup */ func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan) {
			launchplan.Closure.State = admin.LaunchPlanState_INACTIVE
			config.Activate = true
			config.DryRun = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateLaunchPlan", mock.Anything, mock.Anything)
		})
}

func TestForceFlagIsIgnoredWithDryRunDuringLaunchPlanUpdate(t *testing.T) {
	t.Run("without --force", func(t *testing.T) {
		testLaunchPlanUpdate(
			/* setup */ func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan) {
				launchplan.Closure.State = admin.LaunchPlanState_INACTIVE
				config.Activate = true

				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.MockAdminClient.AssertNotCalled(t, "UpdateLaunchPlan", mock.Anything, mock.Anything)
			})
	})

	t.Run("with --force", func(t *testing.T) {
		testLaunchPlanUpdate(
			/* setup */ func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan) {
				launchplan.Closure.State = admin.LaunchPlanState_INACTIVE
				config.Activate = true

				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.MockAdminClient.AssertNotCalled(t, "UpdateLaunchPlan", mock.Anything, mock.Anything)
			})
	})
}

func TestLaunchPlanUpdateFailsWhenLaunchPlanDoesNotExist(t *testing.T) {
	testLaunchPlanUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, launchplan *admin.LaunchPlan) {
			s.MockAdminClient.
				OnGetLaunchPlanMatch(
					s.Ctx,
					mock.MatchedBy(func(r *admin.ObjectGetRequest) bool {
						return cmp.Equal(r.Id, launchplan.Id)
					})).
				Return(nil, ext.NewNotFoundError("launch plan not found"))
			s.MockAdminClient.
				OnUpdateLaunchPlanMatch(s.Ctx, mock.Anything).
				Return(&admin.LaunchPlanUpdateResponse{}, nil)
		},
		/* setup */ nil,
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Error(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateLaunchPlan", mock.Anything, mock.Anything)
		},
	)
}

func TestLaunchPlanUpdateFailsWhenAdminClientFails(t *testing.T) {
	testLaunchPlanUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, launchplan *admin.LaunchPlan) {
			s.MockAdminClient.
				OnGetLaunchPlanMatch(
					s.Ctx,
					mock.MatchedBy(func(r *admin.ObjectGetRequest) bool {
						return cmp.Equal(r.Id, launchplan.Id)
					})).
				Return(launchplan, nil)
			s.MockAdminClient.
				OnUpdateLaunchPlanMatch(s.Ctx, mock.Anything).
				Return(nil, fmt.Errorf("network error"))
		},
		/* setup */ func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan) {
			launchplan.Closure.State = admin.LaunchPlanState_INACTIVE
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Error(t, err)
			s.MockAdminClient.AssertCalled(t, "UpdateLaunchPlan", mock.Anything, mock.Anything)
		},
	)
}

func TestLaunchPlanUpdateRequiresLaunchPlanName(t *testing.T) {
	s := testutils.Setup()
	launchplan.UConfig = &launchplan.UpdateConfig{}

	launchplan.UConfig.Version = testutils.RandomName(2)
	err := updateLPFunc(s.Ctx, nil, s.CmdCtx)

	assert.ErrorContains(t, err, "launch plan name wasn't passed")

	// cleanup
	launchplan.UConfig = &launchplan.UpdateConfig{}
}

func TestLaunchPlanUpdateRequiresLaunchPlanVersion(t *testing.T) {
	s := testutils.Setup()
	launchplan.UConfig = &launchplan.UpdateConfig{}

	name := testutils.RandomName(12)
	err := updateLPFunc(s.Ctx, []string{name}, s.CmdCtx)

	assert.ErrorContains(t, err, "launch plan version wasn't passed")

	// cleanup
	launchplan.UConfig = &launchplan.UpdateConfig{}
}

func testLaunchPlanUpdate(
	setup func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan),
	asserter func(s *testutils.TestStruct, err error),
) {
	testLaunchPlanUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, launchplan *admin.LaunchPlan) {
			s.MockAdminClient.
				OnGetLaunchPlanMatch(
					s.Ctx,
					mock.MatchedBy(func(r *admin.ObjectGetRequest) bool {
						return cmp.Equal(r.Id, launchplan.Id)
					})).
				Return(launchplan, nil)
			s.MockAdminClient.
				OnUpdateLaunchPlanMatch(s.Ctx, mock.Anything).
				Return(&admin.LaunchPlanUpdateResponse{}, nil)
		},
		setup,
		asserter,
	)
}

func testLaunchPlanUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, launchplan *admin.LaunchPlan),
	setup func(s *testutils.TestStruct, config *launchplan.UpdateConfig, launchplan *admin.LaunchPlan),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	target := newTestLaunchPlan()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	launchplan.UConfig = &launchplan.UpdateConfig{}
	if setup != nil {
		setup(&s, launchplan.UConfig, target)
	}

	args := []string{target.Id.Name}
	launchplan.UConfig.Version = target.Id.Version
	err := updateLPFunc(s.Ctx, args, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	launchplan.UConfig = &launchplan.UpdateConfig{}
}

func newTestLaunchPlan() *admin.LaunchPlan {
	return &admin.LaunchPlan{
		Id: &core.Identifier{
			Name:         testutils.RandomName(12),
			Project:      config.GetConfig().Project,
			Domain:       config.GetConfig().Domain,
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Version:      testutils.RandomName(2),
		},
		Closure: &admin.LaunchPlanClosure{
			State: admin.LaunchPlanState_ACTIVE,
		},
	}
}
