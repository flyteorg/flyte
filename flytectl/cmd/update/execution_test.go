package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/execution"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExecutionCanBeActivated(t *testing.T) {
	testExecutionUpdate(
		/* setup */ func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution) {
			execution.Closure.StateChangeDetails.State = admin.ExecutionState_EXECUTION_ARCHIVED
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertCalled(
				t, "UpdateExecution", s.Ctx,
				mock.MatchedBy(
					func(r *admin.ExecutionUpdateRequest) bool {
						return r.State == admin.ExecutionState_EXECUTION_ACTIVE
					}))
		})
}

func TestExecutionCanBeArchived(t *testing.T) {
	testExecutionUpdate(
		/* setup */ func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution) {
			execution.Closure.StateChangeDetails.State = admin.ExecutionState_EXECUTION_ACTIVE
			config.Archive = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertCalled(
				t, "UpdateExecution", s.Ctx,
				mock.MatchedBy(
					func(r *admin.ExecutionUpdateRequest) bool {
						return r.State == admin.ExecutionState_EXECUTION_ARCHIVED
					}))
		})
}

func TestExecutionCannotBeActivatedAndArchivedAtTheSameTime(t *testing.T) {
	testExecutionUpdate(
		/* setup */ func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution) {
			config.Activate = true
			config.Archive = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "Specify either activate or archive")
			s.MockAdminClient.AssertNotCalled(t, "UpdateExecution", mock.Anything, mock.Anything)
		})
}

func TestExecutionUpdateDoesNothingWhenThereAreNoChanges(t *testing.T) {
	testExecutionUpdate(
		/* setup */ func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution) {
			execution.Closure.StateChangeDetails.State = admin.ExecutionState_EXECUTION_ACTIVE
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateExecution", mock.Anything, mock.Anything)
		})
}

func TestExecutionUpdateWithoutForceFlagFails(t *testing.T) {
	testExecutionUpdate(
		/* setup */ func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution) {
			execution.Closure.StateChangeDetails.State = admin.ExecutionState_EXECUTION_ARCHIVED
			config.Activate = true
			config.Force = false
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "update aborted by user")
			s.MockAdminClient.AssertNotCalled(t, "UpdateExecution", mock.Anything, mock.Anything)
		})
}

func TestExecutionUpdateDoesNothingWithDryRunFlag(t *testing.T) {
	testExecutionUpdate(
		/* setup */ func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution) {
			execution.Closure.StateChangeDetails.State = admin.ExecutionState_EXECUTION_ARCHIVED
			config.Activate = true
			config.DryRun = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateExecution", mock.Anything, mock.Anything)
		})
}

func TestForceFlagIsIgnoredWithDryRunDuringExecutionUpdate(t *testing.T) {
	t.Run("without --force", func(t *testing.T) {
		testExecutionUpdate(
			/* setup */ func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution) {
				execution.Closure.StateChangeDetails.State = admin.ExecutionState_EXECUTION_ARCHIVED
				config.Activate = true

				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.MockAdminClient.AssertNotCalled(t, "UpdateExecution", mock.Anything, mock.Anything)
			})
	})

	t.Run("with --force", func(t *testing.T) {
		testExecutionUpdate(
			/* setup */ func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution) {
				execution.Closure.StateChangeDetails.State = admin.ExecutionState_EXECUTION_ARCHIVED
				config.Activate = true

				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.MockAdminClient.AssertNotCalled(t, "UpdateExecution", mock.Anything, mock.Anything)
			})
	})
}

func TestExecutionUpdateFailsWhenExecutionDoesNotExist(t *testing.T) {
	testExecutionUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, execution *admin.Execution) {
			s.FetcherExt.
				OnFetchExecution(s.Ctx, execution.Id.Name, execution.Id.Project, execution.Id.Domain).
				Return(nil, ext.NewNotFoundError("execution not found"))
			s.MockAdminClient.
				OnUpdateExecutionMatch(s.Ctx, mock.Anything).
				Return(&admin.ExecutionUpdateResponse{}, nil)
		},
		/* setup */ nil,
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Error(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateExecution", mock.Anything, mock.Anything)
		},
	)
}

func TestExecutionUpdateFailsWhenAdminClientFails(t *testing.T) {
	testExecutionUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, execution *admin.Execution) {
			s.FetcherExt.
				OnFetchExecution(s.Ctx, execution.Id.Name, execution.Id.Project, execution.Id.Domain).
				Return(execution, nil)
			s.MockAdminClient.
				OnUpdateExecutionMatch(s.Ctx, mock.Anything).
				Return(nil, fmt.Errorf("network error"))
		},
		/* setup */ func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution) {
			execution.Closure.StateChangeDetails.State = admin.ExecutionState_EXECUTION_ARCHIVED
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Error(t, err)
			s.MockAdminClient.AssertCalled(t, "UpdateExecution", mock.Anything, mock.Anything)
		},
	)
}

func TestExecutionUpdateRequiresExecutionName(t *testing.T) {
	s := testutils.Setup()
	err := updateExecutionFunc(s.Ctx, nil, s.CmdCtx)

	assert.ErrorContains(t, err, "execution name wasn't passed")
}

func testExecutionUpdate(
	setup func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution),
	asserter func(s *testutils.TestStruct, err error),
) {
	testExecutionUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, execution *admin.Execution) {
			s.FetcherExt.
				OnFetchExecution(s.Ctx, execution.Id.Name, execution.Id.Project, execution.Id.Domain).
				Return(execution, nil)
			s.MockAdminClient.
				OnUpdateExecutionMatch(s.Ctx, mock.Anything).
				Return(&admin.ExecutionUpdateResponse{}, nil)
		},
		setup,
		asserter,
	)
}

func testExecutionUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, execution *admin.Execution),
	setup func(s *testutils.TestStruct, config *execution.UpdateConfig, execution *admin.Execution),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	target := newTestExecution()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	execution.UConfig = &execution.UpdateConfig{}
	if setup != nil {
		setup(&s, execution.UConfig, target)
	}

	args := []string{target.Id.Name}
	err := updateExecutionFunc(s.Ctx, args, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	execution.UConfig = &execution.UpdateConfig{}
}

func newTestExecution() *admin.Execution {
	return &admin.Execution{
		Id: &core.WorkflowExecutionIdentifier{
			Name:    testutils.RandomName(12),
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		},
		Closure: &admin.ExecutionClosure{
			StateChangeDetails: &admin.ExecutionStateChangeDetails{
				State: admin.ExecutionState_EXECUTION_ACTIVE,
			},
		},
	}
}
