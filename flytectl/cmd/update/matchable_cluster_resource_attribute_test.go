package update

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/clusterresourceattribute"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flytectl/pkg/ext"
)

const (
	validWorkflowClusterResourceAttributesFilePath      = "testdata/valid_workflow_cluster_attribute.yaml"
	validProjectDomainClusterResourceAttributesFilePath = "testdata/valid_project_domain_cluster_attribute.yaml"
	validProjectClusterResourceAttributesFilePath       = "testdata/valid_project_cluster_attribute.yaml"
)

func TestClusterResourceAttributeUpdateRequiresAttributeFile(t *testing.T) {
	testWorkflowClusterResourceAttributeUpdate(
		/* setup */ nil,
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "attrFile is mandatory")
			s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
}

func TestClusterResourceAttributeUpdateFailsWhenAttributeFileDoesNotExist(t *testing.T) {
	testWorkflowClusterResourceAttributeUpdate(
		/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
			config.AttrFile = testDataNonExistentFile
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "unable to read from testdata/non-existent-file yaml file")
			s.UpdaterExt.AssertNotCalled(t, "FetchWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
}

func TestClusterResourceAttributeUpdateFailsWhenAttributeFileIsMalformed(t *testing.T) {
	testWorkflowClusterResourceAttributeUpdate(
		/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
			config.AttrFile = testDataInvalidAttrFile
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\"")
			s.UpdaterExt.AssertNotCalled(t, "FetchWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
}

func TestClusterResourceAttributeUpdateHappyPath(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowClusterResourceAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainClusterResourceAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectClusterResourceAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project`)
			})
	})
}

func TestClusterResourceAttributeUpdateFailsWithoutForceFlag(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowClusterResourceAttributesFilePath
				config.Force = false
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.ErrorContains(t, err, "update aborted by user")
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainClusterResourceAttributesFilePath
				config.Force = false
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.ErrorContains(t, err, "update aborted by user")
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectClusterResourceAttributesFilePath
				config.Force = false
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.ErrorContains(t, err, "update aborted by user")
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func TestClusterResourceAttributeUpdateDoesNothingWithDryRunFlag(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowClusterResourceAttributesFilePath
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainClusterResourceAttributesFilePath
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectClusterResourceAttributesFilePath
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func TestClusterResourceAttributeUpdateIgnoresForceFlagWithDryRun(t *testing.T) {
	t.Run("workflow without --force", func(t *testing.T) {
		testWorkflowClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowClusterResourceAttributesFilePath
				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("workflow with --force", func(t *testing.T) {
		testWorkflowClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowClusterResourceAttributesFilePath
				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain without --force", func(t *testing.T) {
		testProjectDomainClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainClusterResourceAttributesFilePath
				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain with --force", func(t *testing.T) {
		testProjectDomainClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainClusterResourceAttributesFilePath
				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project without --force", func(t *testing.T) {
		testProjectClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectClusterResourceAttributesFilePath
				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project with --force", func(t *testing.T) {
		testProjectClusterResourceAttributeUpdate(
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectClusterResourceAttributesFilePath
				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func TestClusterResourceAttributeUpdateSucceedsWhenAttributesDoNotExist(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowClusterResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
				s.FetcherExt.
					OnFetchWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, admin.MatchableResource_CLUSTER_RESOURCE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					OnUpdateWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, mock.Anything).
					Return(nil)
			},
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowClusterResourceAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainClusterResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
				s.FetcherExt.
					OnFetchProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, admin.MatchableResource_CLUSTER_RESOURCE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					OnUpdateProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, mock.Anything).
					Return(nil)
			},
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainClusterResourceAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectClusterResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
				s.FetcherExt.
					OnFetchProjectAttributesMatch(s.Ctx, target.Project, admin.MatchableResource_CLUSTER_RESOURCE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					OnUpdateProjectAttributesMatch(s.Ctx, target.Project, mock.Anything).
					Return(nil)
			},
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectClusterResourceAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project`)
			})
	})
}

func TestClusterResourceAttributeUpdateFailsWhenAdminClientFails(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowClusterResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
				s.FetcherExt.
					OnFetchWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, admin.MatchableResource_CLUSTER_RESOURCE).
					Return(&admin.WorkflowAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					OnUpdateWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, mock.Anything).
					Return(fmt.Errorf("network error"))
			},
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowClusterResourceAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Error(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainClusterResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
				s.FetcherExt.
					OnFetchProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, admin.MatchableResource_CLUSTER_RESOURCE).
					Return(&admin.ProjectDomainAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					OnUpdateProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, mock.Anything).
					Return(fmt.Errorf("network error"))
			},
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainClusterResourceAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Error(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectClusterResourceAttributeUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
				s.FetcherExt.
					OnFetchProjectAttributesMatch(s.Ctx, target.Project, admin.MatchableResource_CLUSTER_RESOURCE).
					Return(&admin.ProjectAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					OnUpdateProjectAttributesMatch(s.Ctx, target.Project, mock.Anything).
					Return(fmt.Errorf("network error"))
			},
			/* setup */ func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectClusterResourceAttributesFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Error(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func testWorkflowClusterResourceAttributeUpdate(
	setup func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testWorkflowClusterResourceAttributeUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
			s.FetcherExt.
				OnFetchWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, admin.MatchableResource_CLUSTER_RESOURCE).
				Return(&admin.WorkflowAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				OnUpdateWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testWorkflowClusterResourceAttributeUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, target *admin.WorkflowAttributes),
	setup func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.WorkflowAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	clusterresourceattribute.DefaultUpdateConfig = &clusterresourceattribute.AttrUpdateConfig{}
	target := newTestWorkflowClusterResourceAttribute()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	if setup != nil {
		setup(&s, clusterresourceattribute.DefaultUpdateConfig, target)
	}

	err := updateClusterResourceAttributesFunc(s.Ctx, nil, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	clusterresourceattribute.DefaultUpdateConfig = &clusterresourceattribute.AttrUpdateConfig{}
}

func newTestWorkflowClusterResourceAttribute() *admin.WorkflowAttributes {
	return &admin.WorkflowAttributes{
		// project, domain, and workflow names need to be same as in the tests spec files in testdata folder
		Project:  "flytesnacks",
		Domain:   "development",
		Workflow: "core.control_flow.merge_sort.merge_sort",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ClusterResourceAttributes{
				ClusterResourceAttributes: &admin.ClusterResourceAttributes{
					Attributes: map[string]string{
						testutils.RandomName(5): testutils.RandomName(10),
						testutils.RandomName(5): testutils.RandomName(10),
						testutils.RandomName(5): testutils.RandomName(10),
					},
				},
			},
		},
	}
}

func testProjectClusterResourceAttributeUpdate(
	setup func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testProjectClusterResourceAttributeUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
			s.FetcherExt.
				OnFetchProjectAttributesMatch(s.Ctx, target.Project, admin.MatchableResource_CLUSTER_RESOURCE).
				Return(&admin.ProjectAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				OnUpdateProjectAttributesMatch(s.Ctx, target.Project, mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testProjectClusterResourceAttributeUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, target *admin.ProjectAttributes),
	setup func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	clusterresourceattribute.DefaultUpdateConfig = &clusterresourceattribute.AttrUpdateConfig{}
	target := newTestProjectClusterResourceAttribute()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	if setup != nil {
		setup(&s, clusterresourceattribute.DefaultUpdateConfig, target)
	}

	err := updateClusterResourceAttributesFunc(s.Ctx, nil, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	clusterresourceattribute.DefaultUpdateConfig = &clusterresourceattribute.AttrUpdateConfig{}
}

func newTestProjectClusterResourceAttribute() *admin.ProjectAttributes {
	return &admin.ProjectAttributes{
		// project name needs to be same as in the tests spec files in testdata folder
		Project: "flytesnacks",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ClusterResourceAttributes{
				ClusterResourceAttributes: &admin.ClusterResourceAttributes{
					Attributes: map[string]string{
						testutils.RandomName(5): testutils.RandomName(10),
						testutils.RandomName(5): testutils.RandomName(10),
						testutils.RandomName(5): testutils.RandomName(10),
					},
				},
			},
		},
	}
}

func testProjectDomainClusterResourceAttributeUpdate(
	setup func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testProjectDomainClusterResourceAttributeUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
			s.FetcherExt.
				OnFetchProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, admin.MatchableResource_CLUSTER_RESOURCE).
				Return(&admin.ProjectDomainAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				OnUpdateProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testProjectDomainClusterResourceAttributeUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes),
	setup func(s *testutils.TestStruct, config *clusterresourceattribute.AttrUpdateConfig, target *admin.ProjectDomainAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	clusterresourceattribute.DefaultUpdateConfig = &clusterresourceattribute.AttrUpdateConfig{}
	target := newTestProjectDomainClusterResourceAttribute()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	if setup != nil {
		setup(&s, clusterresourceattribute.DefaultUpdateConfig, target)
	}

	err := updateClusterResourceAttributesFunc(s.Ctx, nil, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	clusterresourceattribute.DefaultUpdateConfig = &clusterresourceattribute.AttrUpdateConfig{}
}

func newTestProjectDomainClusterResourceAttribute() *admin.ProjectDomainAttributes {
	return &admin.ProjectDomainAttributes{
		// project and domain names need to be same as in the tests spec files in testdata folder
		Project: "flytesnacks",
		Domain:  "development",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ClusterResourceAttributes{
				ClusterResourceAttributes: &admin.ClusterResourceAttributes{
					Attributes: map[string]string{
						testutils.RandomName(5): testutils.RandomName(10),
						testutils.RandomName(5): testutils.RandomName(10),
						testutils.RandomName(5): testutils.RandomName(10),
					},
				},
			},
		},
	}
}
