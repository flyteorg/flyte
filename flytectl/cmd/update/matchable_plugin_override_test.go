package update

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	pluginoverride "github.com/flyteorg/flytectl/cmd/config/subcommand/plugin_override"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flytectl/pkg/ext"
)

const (
	validProjectPluginOverrideFilePath       = "testdata/valid_project_plugin_override.yaml"
	validProjectDomainPluginOverrideFilePath = "testdata/valid_project_domain_plugin_override.yaml"
	validWorkflowPluginOverrideFilePath      = "testdata/valid_workflow_plugin_override.yaml"
)

func TestPluginOverrideUpdateRequiresAttributeFile(t *testing.T) {
	testWorkflowPluginOverrideUpdate(
		/* setup */ nil,
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "attrFile is mandatory")
			s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
}

func TestPluginOverrideUpdateFailsWhenAttributeFileDoesNotExist(t *testing.T) {
	testWorkflowPluginOverrideUpdate(
		/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes) {
			config.AttrFile = testDataNonExistentFile
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "unable to read from testdata/non-existent-file yaml file")
			s.UpdaterExt.AssertNotCalled(t, "FetchWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
}

func TestPluginOverrideUpdateFailsWhenAttributeFileIsMalformed(t *testing.T) {
	testWorkflowPluginOverrideUpdate(
		/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes) {
			config.AttrFile = testDataInvalidAttrFile
			config.Force = true
		},
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\"")
			s.UpdaterExt.AssertNotCalled(t, "FetchWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
}

func TestPluginOverrideUpdateHappyPath(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowPluginOverrideFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainPluginOverrideFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectPluginOverrideFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project`)
			})
	})
}

func TestPluginOverrideUpdateFailsWithoutForceFlag(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowPluginOverrideFilePath
				config.Force = false
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.ErrorContains(t, err, "update aborted by user")
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainPluginOverrideFilePath
				config.Force = false
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.ErrorContains(t, err, "update aborted by user")
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectPluginOverrideFilePath
				config.Force = false
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.ErrorContains(t, err, "update aborted by user")
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func TestPluginOverrideUpdateDoesNothingWithDryRunFlag(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowPluginOverrideFilePath
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainPluginOverrideFilePath
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectPluginOverrideFilePath
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func TestPluginOverrideUpdateIgnoresForceFlagWithDryRun(t *testing.T) {
	t.Run("workflow without --force", func(t *testing.T) {
		testWorkflowPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowPluginOverrideFilePath
				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("workflow with --force", func(t *testing.T) {
		testWorkflowPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowPluginOverrideFilePath
				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain without --force", func(t *testing.T) {
		testProjectDomainPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainPluginOverrideFilePath
				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain with --force", func(t *testing.T) {
		testProjectDomainPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainPluginOverrideFilePath
				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project without --force", func(t *testing.T) {
		testProjectPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectPluginOverrideFilePath
				config.Force = false
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project with --force", func(t *testing.T) {
		testProjectPluginOverrideUpdate(
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectPluginOverrideFilePath
				config.Force = true
				config.DryRun = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertNotCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func TestPluginOverrideUpdateSucceedsWhenAttributesDoNotExist(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowPluginOverrideUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
				s.FetcherExt.
					OnFetchWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					OnUpdateWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, mock.Anything).
					Return(nil)
			},
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowPluginOverrideFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainPluginOverrideUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
				s.FetcherExt.
					OnFetchProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					OnUpdateProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, mock.Anything).
					Return(nil)
			},
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainPluginOverrideFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project and domain development`)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectPluginOverrideUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
				s.FetcherExt.
					OnFetchProjectAttributesMatch(s.Ctx, target.Project, admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					OnUpdateProjectAttributesMatch(s.Ctx, target.Project, mock.Anything).
					Return(nil)
			},
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectPluginOverrideFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Nil(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
				s.TearDownAndVerifyContains(t, `Updated attributes from flytesnacks project`)
			})
	})
}

func TestPluginOverrideUpdateFailsWhenAdminClientFails(t *testing.T) {
	t.Run("workflow", func(t *testing.T) {
		testWorkflowPluginOverrideUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
				s.FetcherExt.
					OnFetchWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(&admin.WorkflowAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					OnUpdateWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, mock.Anything).
					Return(fmt.Errorf("network error"))
			},
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes) {
				config.AttrFile = validWorkflowPluginOverrideFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Error(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("domain", func(t *testing.T) {
		testProjectDomainPluginOverrideUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
				s.FetcherExt.
					OnFetchProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(&admin.ProjectDomainAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					OnUpdateProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, mock.Anything).
					Return(fmt.Errorf("network error"))
			},
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes) {
				config.AttrFile = validProjectDomainPluginOverrideFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Error(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectDomainAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
	})

	t.Run("project", func(t *testing.T) {
		testProjectPluginOverrideUpdateWithMockSetup(
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
				s.FetcherExt.
					OnFetchProjectAttributesMatch(s.Ctx, target.Project, admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(&admin.ProjectAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					OnUpdateProjectAttributesMatch(s.Ctx, target.Project, mock.Anything).
					Return(fmt.Errorf("network error"))
			},
			/* setup */ func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes) {
				config.AttrFile = validProjectPluginOverrideFilePath
				config.Force = true
			},
			/* assert */ func(s *testutils.TestStruct, err error) {
				assert.Error(t, err)
				s.UpdaterExt.AssertCalled(t, "UpdateProjectAttributes", mock.Anything, mock.Anything, mock.Anything)
			})
	})
}

func testWorkflowPluginOverrideUpdate(
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testWorkflowPluginOverrideUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
			s.FetcherExt.
				OnFetchWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, admin.MatchableResource_PLUGIN_OVERRIDE).
				Return(&admin.WorkflowAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				OnUpdateWorkflowAttributesMatch(s.Ctx, target.Project, target.Domain, target.Workflow, mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testWorkflowPluginOverrideUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, target *admin.WorkflowAttributes),
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	pluginoverride.DefaultUpdateConfig = &pluginoverride.AttrUpdateConfig{}
	target := newTestWorkflowPluginOverride()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	if setup != nil {
		setup(&s, pluginoverride.DefaultUpdateConfig, target)
	}

	err := updatePluginOverridesFunc(s.Ctx, nil, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	pluginoverride.DefaultUpdateConfig = &pluginoverride.AttrUpdateConfig{}
}

func newTestWorkflowPluginOverride() *admin.WorkflowAttributes {
	return &admin.WorkflowAttributes{
		// project, domain, and workflow names need to be same as in the tests spec files in testdata folder
		Project:  "flytesnacks",
		Domain:   "development",
		Workflow: "core.control_flow.merge_sort.merge_sort",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_PluginOverrides{
				PluginOverrides: &admin.PluginOverrides{
					Overrides: []*admin.PluginOverride{
						{
							TaskType: testutils.RandomName(15),
							PluginId: []string{
								testutils.RandomName(12),
								testutils.RandomName(12),
								testutils.RandomName(12),
							},
							MissingPluginBehavior: admin.PluginOverride_FAIL,
						},
					},
				},
			},
		},
	}
}

func testProjectPluginOverrideUpdate(
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testProjectPluginOverrideUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
			s.FetcherExt.
				OnFetchProjectAttributesMatch(s.Ctx, target.Project, admin.MatchableResource_PLUGIN_OVERRIDE).
				Return(&admin.ProjectAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				OnUpdateProjectAttributesMatch(s.Ctx, target.Project, mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testProjectPluginOverrideUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, target *admin.ProjectAttributes),
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	pluginoverride.DefaultUpdateConfig = &pluginoverride.AttrUpdateConfig{}
	target := newTestProjectPluginOverride()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	if setup != nil {
		setup(&s, pluginoverride.DefaultUpdateConfig, target)
	}

	err := updatePluginOverridesFunc(s.Ctx, nil, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	pluginoverride.DefaultUpdateConfig = &pluginoverride.AttrUpdateConfig{}
}

func newTestProjectPluginOverride() *admin.ProjectAttributes {
	return &admin.ProjectAttributes{
		// project name needs to be same as in the tests spec files in testdata folder
		Project: "flytesnacks",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_PluginOverrides{
				PluginOverrides: &admin.PluginOverrides{
					Overrides: []*admin.PluginOverride{
						{
							TaskType: testutils.RandomName(15),
							PluginId: []string{
								testutils.RandomName(12),
								testutils.RandomName(12),
								testutils.RandomName(12),
							},
							MissingPluginBehavior: admin.PluginOverride_FAIL,
						},
					},
				},
			},
		},
	}
}

func testProjectDomainPluginOverrideUpdate(
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testProjectDomainPluginOverrideUpdateWithMockSetup(
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
			s.FetcherExt.
				OnFetchProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, admin.MatchableResource_PLUGIN_OVERRIDE).
				Return(&admin.ProjectDomainAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				OnUpdateProjectDomainAttributesMatch(s.Ctx, target.Project, target.Domain, mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testProjectDomainPluginOverrideUpdateWithMockSetup(
	mockSetup func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes),
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup()
	pluginoverride.DefaultUpdateConfig = &pluginoverride.AttrUpdateConfig{}
	target := newTestProjectDomainPluginOverride()

	if mockSetup != nil {
		mockSetup(&s, target)
	}

	if setup != nil {
		setup(&s, pluginoverride.DefaultUpdateConfig, target)
	}

	err := updatePluginOverridesFunc(s.Ctx, nil, s.CmdCtx)

	if asserter != nil {
		asserter(&s, err)
	}

	// cleanup
	pluginoverride.DefaultUpdateConfig = &pluginoverride.AttrUpdateConfig{}
}

func newTestProjectDomainPluginOverride() *admin.ProjectDomainAttributes {
	return &admin.ProjectDomainAttributes{
		// project and domain names need to be same as in the tests spec files in testdata folder
		Project: "flytesnacks",
		Domain:  "development",
		MatchingAttributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_PluginOverrides{
				PluginOverrides: &admin.PluginOverrides{
					Overrides: []*admin.PluginOverride{
						{
							TaskType: testutils.RandomName(15),
							PluginId: []string{
								testutils.RandomName(12),
								testutils.RandomName(12),
								testutils.RandomName(12),
							},
							MissingPluginBehavior: admin.PluginOverride_FAIL,
						},
					},
				},
			},
		},
	}
}
