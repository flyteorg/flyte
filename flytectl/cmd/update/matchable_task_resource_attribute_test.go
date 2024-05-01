package update

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/taskresourceattribute"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flytectl/pkg/ext"
)

const (
	validProjectTaskAttributesFilePath       = "testdata/valid_project_task_attribute.yaml"
	validProjectDomainTaskAttributesFilePath = "testdata/valid_project_domain_task_attribute.yaml"
	validWorkflowTaskAttributesFilePath      = "testdata/valid_workflow_task_attribute.yaml"
)

func TestTaskResourceAttributeUpdateRequiresAttributeFile(t *testing.T) {
	testWorkflowTaskResourceAttributeUpdate(
		/* setup */ nil,
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "attrFile is mandatory")
			s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
}

func TestTaskResourceAttributeUpdateFailsWhenAttributeFileDoesNotExist(t *testing.T) {
	testWorkflowTaskResourceAttributeUpdate(
		/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
			config.AttrFile = testDataNonExistentFile
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "unable to read from testdata/non-existent-file yaml file")
			s.UpdaterExt.AssertNotCalled(t, "FetchWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
}

func TestTaskResourceAttributeUpdateFailsWhenAttributeFileIsMalformed(t *testing.T) {
	testWorkflowTaskResourceAttributeUpdate(
		/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
			config.AttrFile = testDataInvalidAttrFile
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\"")
			s.UpdaterExt.AssertNotCalled(t, "FetchWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
}

func TestTaskResourceAttributeUpdateHappyPath(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowTaskAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainTaskAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectTaskAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project`)
			})
	})
}

func TestTaskResourceAttributeUpdateFailsWithoutForceFlag(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowTaskAttributesFilePath
				config.Force = false
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.ErrorContains(t, err, "update aborted by user")
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainTaskAttributesFilePath
				config.Force = false
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.ErrorContains(t, err, "update aborted by user")
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectTaskAttributesFilePath
				config.Force = false
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.ErrorContains(t, err, "update aborted by user")
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func TestTaskResourceAttributeUpdateDoesNothingWithDryRunFlag(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowTaskAttributesFilePath
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainTaskAttributesFilePath
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectTaskAttributesFilePath
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func TestTaskResourceAttributeUpdateIgnoresForceFlagWithDryRun(t *testing.T) {
	t.Run("workflow without --force", func(t *testing.T) {
		testWorkflowTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowTaskAttributesFilePath
				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("workflow with --force", func(t *testing.T) {
		testWorkflowTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowTaskAttributesFilePath
				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain without --force", func(t *testing.T) {
		testProjectDomainTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainTaskAttributesFilePath
				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain with --force", func(t *testing.T) {
		testProjectDomainTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainTaskAttributesFilePath
				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project without --force", func(t *testing.T) {
		testProjectTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectTaskAttributesFilePath
				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project with --force", func(t *testing.T) {
		testProjectTaskResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectTaskAttributesFilePath
				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func TestTaskResourceAttributeUpdateSucceedsWhenAttributesDoNotExist(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowTaskResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
				s.FetcherExt.
					OnFetchWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, admin.MatchableResource_TASK_RESOURCE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					OnUpdateWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, mock.Anything).
					Return(nil)
			},
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowTaskAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainTaskResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
				s.FetcherExt.
					OnFetchProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, admin.MatchableResource_TASK_RESOURCE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					OnUpdateProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, mock.Anything).
					Return(nil)
			},
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainTaskAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectTaskResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
				s.FetcherExt.
					OnFetchProjectAttributesMatch(s.Ctx, target.Project, admin.MatchableResource_TASK_RESOURCE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					OnUpdateProjectAttributesMatch(s.Ctx, target.Project, mock.Anything).
					Return(nil)
			},
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectTaskAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project`)
			})
	})
}

func TestTaskResourceAttributeUpdateFailsWhenAdminClientFails(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowTaskResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
				s.FetcherExt.
					OnFetchWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, admin.MatchableResource_TASK_RESOURCE).
					Return(&admin.WorkflowAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					OnUpdateWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, mock.Anything).
					Return(fmt.Errorf("network error"))
			},
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowTaskAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Error(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainTaskResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
				s.FetcherExt.
					OnFetchProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, admin.MatchableResource_TASK_RESOURCE).
					Return(&admin.ProjectDomainAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					OnUpdateProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, mock.Anything).
					Return(fmt.Errorf("network error"))
			},
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainTaskAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Error(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectTaskResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
				s.FetcherExt.
					OnFetchProjectAttributesMatch(s.Ctx, target.Project, admin.MatchableResource_TASK_RESOURCE).
					Return(&admin.ProjectAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					OnUpdateProjectAttributesMatch(s.Ctx, target.Project, mock.Anything).
					Return(fmt.Errorf("network error"))
			},
			/* setup */ func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectTaskAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Error(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func testWorkflowTaskResourceAttributeUpdate(
	setup func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testWorkflowTaskResourceAttributeUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
			s.FetcherExt.
				OnFetchWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, admin.MatchableResource_TASK_RESOURCE).
				Return(&admin.WorkflowAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				OnUpdateWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testWorkflowTaskResourceAttributeUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, target *admin.WorkflowAttributes),
	setup func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	taskresourceattribute.DefaultUpdateConfig = &taskresourceattribute.AttrUpdateConfig{}
	target := newTestWorkflowTaskResourceAttribute()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	if setup != nil {
		setup(&s, taskresourceattribute.DefaultUpdateConfig, target)
	}

	err := updateTaskResourceAttributesFunc(s.Ctx, nil, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	taskresourceattribute.DefaultUpdateConfig = &taskresourceattribute.AttrUpdateConfig{}
}

func newTestWorkflowTaskResourceAttribute() *admin.WorkflowAttributes {
	return &admin.WorkflowAttributes{
		// project, domain, and workflow names need to be same as in the tests spec files in testdata folder
		Project:  "flytesnacks",
		Domain:   "development",
		Workflow: "core.control_flow.merge_sort.merge_sort",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_TaskResourceAttributes{
				TaskResourceAttributes: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu:    testutils.RandomName(2),
						Memory: testutils.RandomName(5),
					},
				},
			},
		},
	}
}

func testProjectTaskResourceAttributeUpdate(
	setup func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testProjectTaskResourceAttributeUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
			s.FetcherExt.
				OnFetchProjectAttributesMatch(s.Ctx, target.Project, admin.MatchableResource_TASK_RESOURCE).
				Return(&admin.ProjectAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				OnUpdateProjectAttributesMatch(s.Ctx, target.Project, mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testProjectTaskResourceAttributeUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, target *admin.ProjectAttributes),
	setup func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	taskresourceattribute.DefaultUpdateConfig = &taskresourceattribute.AttrUpdateConfig{}
	target := newTestProjectTaskResourceAttribute()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	if setup != nil {
		setup(&s, taskresourceattribute.DefaultUpdateConfig, target)
	}

	err := updateTaskResourceAttributesFunc(s.Ctx, nil, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	taskresourceattribute.DefaultUpdateConfig = &taskresourceattribute.AttrUpdateConfig{}
}

func newTestProjectTaskResourceAttribute() *admin.ProjectAttributes {
	return &admin.ProjectAttributes{
		// project name needs to be same as in the tests spec files in testdata folder
		Project: "flytesnacks",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_TaskResourceAttributes{
				TaskResourceAttributes: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu:    testutils.RandomName(2),
						Memory: testutils.RandomName(5),
					},
				},
			},
		},
	}
}

func testProjectDomainTaskResourceAttributeUpdate(
	setup func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testProjectDomainTaskResourceAttributeUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
			s.FetcherExt.
				OnFetchProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, admin.MatchableResource_TASK_RESOURCE).
				Return(&admin.ProjectDomainAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				OnUpdateProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testProjectDomainTaskResourceAttributeUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes),
	setup func(s *testutils.TestStruct, config *taskresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	taskresourceattribute.DefaultUpdateConfig = &taskresourceattribute.AttrUpdateConfig{}
	target := newTestProjectDomainTaskResourceAttribute()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	if setup != nil {
		setup(&s, taskresourceattribute.DefaultUpdateConfig, target)
	}

	err := updateTaskResourceAttributesFunc(s.Ctx, nil, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	taskresourceattribute.DefaultUpdateConfig = &taskresourceattribute.AttrUpdateConfig{}
}

func newTestProjectDomainTaskResourceAttribute() *admin.ProjectDomainAttributes {
	return &admin.ProjectDomainAttributes{
		// project and domain names need to be same as in the tests spec files in testdata folder
		Project: "flytesnacks",
		Domain:  "development",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_TaskResourceAttributes{
				TaskResourceAttributes: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu:    testutils.RandomName(2),
						Memory: testutils.RandomName(5),
					},
				},
			},
		},
	}
}
