package update

import (
	"fmt"
	"testing"

	pluginoverride "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/plugin_override"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/flyteorg/flyte/flytectl/pkg/ext"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validProjectPluginOverrideFilePath       = "testdata/valid_project_plugin_override.yaml"
	validProjectDomainPluginOverrideFilePath = "testdata/valid_project_domain_plugin_override.yaml"
	validWorkflowPluginOverrideFilePath      = "testdata/valid_workflow_plugin_override.yaml"
)

func TestPluginOverrideUpdateRequiresAttributeFile(t *testing.T) {
	testWorkflowPluginOverrideUpdate(
		t,
		/* setup */ nil,
		/* assert */ func(s *testutils.TestStruct, err error) {
			assert.ErrorContains(t, err, "attrFile is mandatory")
			s.UpdaterExt.AssertNotCalled(t, "UpdateWorkflowAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
}

func TestPluginOverrideUpdateFailsWhenAttributeFileDoesNotExist(t *testing.T) {
	testWorkflowPluginOverrideUpdate(
		t,
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
		t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
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
			t,
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
				s.FetcherExt.
					EXPECT().FetchWorkflowAttributes(s.Ctx, target.GetProject(), target.GetDomain(), target.GetWorkflow(), admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					EXPECT().UpdateWorkflowAttributes(s.Ctx, target.GetProject(), target.GetDomain(), target.GetWorkflow(), mock.Anything).
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
			t,
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
				s.FetcherExt.
					EXPECT().FetchProjectDomainAttributes(s.Ctx, target.GetProject(), target.GetDomain(), admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					EXPECT().UpdateProjectDomainAttributes(s.Ctx, target.GetProject(), target.GetDomain(), mock.Anything).
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
			t,
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
				s.FetcherExt.
					EXPECT().FetchProjectAttributes(s.Ctx, target.GetProject(), admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(nil, ext.NewNotFoundError("attribute"))
				s.UpdaterExt.
					EXPECT().UpdateProjectAttributes(s.Ctx, target.GetProject(), mock.Anything).
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
			t,
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
				s.FetcherExt.
					EXPECT().FetchWorkflowAttributes(s.Ctx, target.GetProject(), target.GetDomain(), target.GetWorkflow(), admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(&admin.WorkflowAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					EXPECT().UpdateWorkflowAttributes(s.Ctx, target.GetProject(), target.GetDomain(), target.GetWorkflow(), mock.Anything).
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
			t,
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
				s.FetcherExt.
					EXPECT().FetchProjectDomainAttributes(s.Ctx, target.GetProject(), target.GetDomain(), admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(&admin.ProjectDomainAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					EXPECT().UpdateProjectDomainAttributes(s.Ctx, target.GetProject(), target.GetDomain(), mock.Anything).
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
			t,
			/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
				s.FetcherExt.
					EXPECT().FetchProjectAttributes(s.Ctx, target.GetProject(), admin.MatchableResource_PLUGIN_OVERRIDE).
					Return(&admin.ProjectAttributesGetResponse{Attributes: target}, nil)
				s.UpdaterExt.
					EXPECT().UpdateProjectAttributes(s.Ctx, target.GetProject(), mock.Anything).
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
	t *testing.T,
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testWorkflowPluginOverrideUpdateWithMockSetup(
		t,
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.WorkflowAttributes) {
			s.FetcherExt.
				EXPECT().FetchWorkflowAttributes(s.Ctx, target.GetProject(), target.GetDomain(), target.GetWorkflow(), admin.MatchableResource_PLUGIN_OVERRIDE).
				Return(&admin.WorkflowAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				EXPECT().UpdateWorkflowAttributes(s.Ctx, target.GetProject(), target.GetDomain(), target.GetWorkflow(), mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testWorkflowPluginOverrideUpdateWithMockSetup(
	t *testing.T,
	mockSetup func(s *testutils.TestStruct, target *admin.WorkflowAttributes),
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.WorkflowAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup(t)

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
	t *testing.T,
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testProjectPluginOverrideUpdateWithMockSetup(
		t,
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectAttributes) {
			s.FetcherExt.
				EXPECT().FetchProjectAttributes(s.Ctx, target.GetProject(), admin.MatchableResource_PLUGIN_OVERRIDE).
				Return(&admin.ProjectAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				EXPECT().UpdateProjectAttributes(s.Ctx, target.GetProject(), mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testProjectPluginOverrideUpdateWithMockSetup(
	t *testing.T,
	mockSetup func(s *testutils.TestStruct, target *admin.ProjectAttributes),
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup(t)

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
	t *testing.T,
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	testProjectDomainPluginOverrideUpdateWithMockSetup(
		t,
		/* mockSetup */ func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes) {
			s.FetcherExt.
				EXPECT().FetchProjectDomainAttributes(s.Ctx, target.GetProject(), target.GetDomain(), admin.MatchableResource_PLUGIN_OVERRIDE).
				Return(&admin.ProjectDomainAttributesGetResponse{Attributes: target}, nil)
			s.UpdaterExt.
				EXPECT().UpdateProjectDomainAttributes(s.Ctx, target.GetProject(), target.GetDomain(), mock.Anything).
				Return(nil)
		},
		setup,
		asserter,
	)
}

func testProjectDomainPluginOverrideUpdateWithMockSetup(
	t *testing.T,
	mockSetup func(s *testutils.TestStruct, target *admin.ProjectDomainAttributes),
	setup func(s *testutils.TestStruct, config *pluginoverride.AttrUpdateConfig, target *admin.ProjectDomainAttributes),
	asserter func(s *testutils.TestStruct, err error),
) {
	s := testutils.Setup(t)

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
