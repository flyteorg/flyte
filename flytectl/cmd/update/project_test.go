package update

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flytectl/pkg/ext"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProjectCanBeActivated(t *testing.T) {
	testProjectUpdate(
		/* setup */ func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project) {
			project.State = admin.Project_ARCHIVED
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertCalled(
				t, "UpdateProject", s.Ctx,
				mock.MatchedBy(
					func(r *admin.Project) bool {
						return r.State == admin.Project_ACTIVE
					}))
		})
}

func TestProjectCanBeArchived(t *testing.T) {
	testProjectUpdate(
		/* setup */ func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project) {
			project.State = admin.Project_ACTIVE
			config.Archive = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertCalled(
				t, "UpdateProject", s.Ctx,
				mock.MatchedBy(
					func(r *admin.Project) bool {
						return r.State == admin.Project_ARCHIVED
					}))
		})
}

func TestProjectCannotBeActivatedAndArchivedAtTheSameTime(t *testing.T) {
	testProjectUpdate(
		/* setup */ func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project) {
			config.Activate = true
			config.Archive = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "Specify either activate or archive")
			s.MockAdminClient.AssertNotCalled(t, "UpdateProject", mock.Anything, mock.Anything)
		})
}

func TestProjectUpdateDoesNothingWhenThereAreNoChanges(t *testing.T) {
	testProjectUpdate(
		/* setup */ func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project) {
			project.State = admin.Project_ACTIVE
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateProject", mock.Anything, mock.Anything)
		})
}

func TestProjectUpdateWithoutForceFlagFails(t *testing.T) {
	testProjectUpdate(
		/* setup */ func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project) {
			project.State = admin.Project_ARCHIVED
			config.Activate = true
			config.Force = false
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "update aborted by user")
			s.MockAdminClient.AssertNotCalled(t, "UpdateProject", mock.Anything, mock.Anything)
		})
}

func TestProjectUpdateDoesNothingWithDryRunFlag(t *testing.T) {
	testProjectUpdate(
		/* setup */ func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project) {
			project.State = admin.Project_ARCHIVED
			config.Activate = true
			config.DryRun = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Nil(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateProject", mock.Anything, mock.Anything)
		})
}

func TestForceFlagIsIgnoredWithDryRunDuringProjectUpdate(t *testing.T) {
	t.Run("without --force", func(t *testing.T) {
		testProjectUpdate(
			/* setup */ func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project) {
				project.State = admin.Project_ARCHIVED
				config.Activate = true

				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.MockAdminClient.AssertNotCalled(t, "UpdateProject", mock.Anything, mock.Anything)
			})
	})

	t.Run("with --force", func(t *testing.T) {
		testProjectUpdate(
			/* setup */ func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project) {
				project.State = admin.Project_ARCHIVED
				config.Activate = true

				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.MockAdminClient.AssertNotCalled(t, "UpdateProject", mock.Anything, mock.Anything)
			})
	})
}

func TestProjectUpdateFailsWhenProjectDoesNotExist(t *testing.T) {
	testProjectUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, project *admin.Project) {
			s.FetcherExt.
				OnGetProjectByID(s.Ctx, project.Id).
				Return(nil, ext.NewNotFoundError("project not found"))
			s.MockAdminClient.
				OnUpdateProjectMatch(s.Ctx, mock.Anything).
				Return(&admin.ProjectUpdateResponse{}, nil)
		},
		/* setup */ nil,
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Error(t, err)
			s.MockAdminClient.AssertNotCalled(t, "UpdateProject", mock.Anything, mock.Anything)
		},
	)
}

func TestProjectUpdateFailsWhenAdminClientFails(t *testing.T) {
	testProjectUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, project *admin.Project) {
			s.FetcherExt.
				OnGetProjectByID(s.Ctx, project.Id).
				Return(project, nil)
			s.MockAdminClient.
				OnUpdateProjectMatch(s.Ctx, mock.Anything).
				Return(nil, fmt.Errorf("network error"))
		},
		/* setup */ func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project) {
			project.State = admin.Project_ARCHIVED
			config.Activate = true
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.Error(t, err)
			s.MockAdminClient.AssertCalled(t, "UpdateProject", mock.Anything, mock.Anything)
		},
	)
}

func TestProjectUpdateRequiresProjectId(t *testing.T) {
	testProjectUpdate(
		/* setup */ func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project) {
			config.ID = ""
		},
		func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "project id wasn't passed")
		})
}

func testProjectUpdate(
	setup func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project),
	asserter func(s *testutils.TestStruct, err error),
) {
	testProjectUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, project *admin.Project) {
			s.FetcherExt.
				OnGetProjectByID(s.Ctx, project.Id).
				Return(project, nil)
			s.MockAdminClient.
				OnUpdateProjectMatch(s.Ctx, mock.Anything).
				Return(&admin.ProjectUpdateResponse{}, nil)
		},
		setup,
		asserter,
	)
}

func testProjectUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, project *admin.Project),
	setup func(s *testutils.TestStruct, config *project.ConfigProject, project *admin.Project),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	target := newTestProject()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	project.DefaultProjectConfig = &project.ConfigProject{
		ID: target.Id,
	}
	config.GetConfig().Project = ""
	config.GetConfig().Domain = ""
	if setup != nil {
		setup(&s, project.DefaultProjectConfig, target)
	}

	err := updateProjectsFunc(s.Ctx, nil, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	project.DefaultProjectConfig = &project.ConfigProject{}
	config.GetConfig().Project = ""
	config.GetConfig().Domain = ""
}

func newTestProject() *admin.Project {
	return &admin.Project{
		Id:    testutils.RandomName(12),
		Name:  testutils.RandomName(12),
		State: admin.Project_ACTIVE,
		Domains: []*admin.Domain{
			{
				Id:   testutils.RandomName(12),
				Name: testutils.RandomName(12),
			},
		},
		Description: testutils.RandomName(12),
		Labels: &admin.Labels{
			Values: map[string]string{
				testutils.RandomName(5): testutils.RandomName(12),
				testutils.RandomName(5): testutils.RandomName(12),
				testutils.RandomName(5): testutils.RandomName(12),
			},
		},
	}
}
