package resources

import (
	"context"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	// pkg/runtime/interfaces/application_configuration.go
	"testing"

	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	commonTestUtils "github.com/flyteorg/flyteadmin/pkg/common/testutils"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

const project = "project"
const domain = "domain"
const workflow = "workflow"
const python = "python"
const hive = "hive"

func TestUpdateWorkflowAttributes(t *testing.T) {
	request := admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Project:            project,
			Domain:             domain,
			Workflow:           workflow,
			MatchingAttributes: testutils.ExecutionQueueAttributes,
		},
	}
	db := mocks.NewMockRepository()
	expectedSerializedAttrs, _ := proto.Marshal(testutils.ExecutionQueueAttributes)
	var createOrUpdateCalled bool
	db.ResourceRepo().(*mocks.MockResourceRepo).CreateOrUpdateFunction = func(
		ctx context.Context, input models.Resource) error {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, workflow, input.Workflow)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), input.ResourceType)
		assert.EqualValues(t, expectedSerializedAttrs, input.Attributes)
		createOrUpdateCalled = true
		return nil
	}
	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	_, err := manager.UpdateWorkflowAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, createOrUpdateCalled)
}

func TestUpdateWorkflowAttributes_CreateOrMerge(t *testing.T) {
	request := admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Project:            project,
			Domain:             domain,
			Workflow:           workflow,
			MatchingAttributes: commonTestUtils.GetPluginOverridesAttributes(map[string][]string{"python": {"plugin a"}}),
		},
	}

	t.Run("create only", func(t *testing.T) {
		db := mocks.NewMockRepository()
		db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID repoInterfaces.ResourceID) (
			models.Resource, error) {
			return models.Resource{}, errors.NewFlyteAdminError(codes.NotFound, "foo")
		}
		var createOrUpdateCalled bool
		db.ResourceRepo().(*mocks.MockResourceRepo).CreateOrUpdateFunction = func(ctx context.Context, input models.Resource) error {
			assert.Equal(t, project, input.Project)
			assert.Equal(t, domain, input.Domain)
			assert.Equal(t, workflow, input.Workflow)

			var attributesToBeSaved admin.MatchingAttributes
			err := proto.Unmarshal(input.Attributes, &attributesToBeSaved)
			if err != nil {
				t.Fatal(err)
			}
			assert.Len(t, attributesToBeSaved.GetPluginOverrides().Overrides, 1)
			assert.True(t, proto.Equal(attributesToBeSaved.GetPluginOverrides().Overrides[0], &admin.PluginOverride{
				TaskType: "python",
				PluginId: []string{"plugin a"}}))

			createOrUpdateCalled = true
			return nil
		}
		manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
		_, err := manager.UpdateWorkflowAttributes(context.Background(), request)
		assert.NoError(t, err)
		assert.True(t, createOrUpdateCalled)
	})
	t.Run("merge update", func(t *testing.T) {
		db := mocks.NewMockRepository()
		db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID repoInterfaces.ResourceID) (
			models.Resource, error) {
			existingAttributes := commonTestUtils.GetPluginOverridesAttributes(map[string][]string{
				"hive":   {"plugin b"},
				"python": {"plugin c"},
			})
			bytes, err := proto.Marshal(existingAttributes)
			if err != nil {
				t.Fatal(err)
			}
			return models.Resource{
				Project:    project,
				Domain:     domain,
				Workflow:   workflow,
				Attributes: bytes,
			}, nil
		}
		var createOrUpdateCalled bool
		db.ResourceRepo().(*mocks.MockResourceRepo).CreateOrUpdateFunction = func(ctx context.Context, input models.Resource) error {
			assert.Equal(t, project, input.Project)
			assert.Equal(t, domain, input.Domain)
			assert.Equal(t, workflow, input.Workflow)

			var attributesToBeSaved admin.MatchingAttributes
			err := proto.Unmarshal(input.Attributes, &attributesToBeSaved)
			if err != nil {
				t.Fatal(err)
			}

			assert.Len(t, attributesToBeSaved.GetPluginOverrides().Overrides, 2)
			for _, override := range attributesToBeSaved.GetPluginOverrides().Overrides {
				if override.TaskType == python {
					assert.EqualValues(t, []string{"plugin a"}, override.PluginId)
				} else if override.TaskType == hive {
					assert.EqualValues(t, []string{"plugin b"}, override.PluginId)
				} else {
					t.Errorf("Unexpected task type [%s] plugin override committed to db", override.TaskType)
				}
			}
			createOrUpdateCalled = true
			return nil
		}
		manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
		_, err := manager.UpdateWorkflowAttributes(context.Background(), request)
		assert.NoError(t, err)
		assert.True(t, createOrUpdateCalled)
	})
}

func TestGetWorkflowAttributes(t *testing.T) {
	request := admin.WorkflowAttributesGetRequest{
		Project:      project,
		Domain:       domain,
		Workflow:     workflow,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(
		ctx context.Context, ID repoInterfaces.ResourceID) (models.Resource, error) {
		assert.Equal(t, project, ID.Project)
		assert.Equal(t, domain, ID.Domain)
		assert.Equal(t, workflow, ID.Workflow)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), ID.ResourceType)
		expectedSerializedAttrs, _ := proto.Marshal(testutils.ExecutionQueueAttributes)
		return models.Resource{
			Project:      project,
			Domain:       domain,
			Workflow:     workflow,
			ResourceType: "resource",
			Attributes:   expectedSerializedAttrs,
		}, nil
	}
	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	response, err := manager.GetWorkflowAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.WorkflowAttributesGetResponse{
		Attributes: &admin.WorkflowAttributes{
			Project:            project,
			Domain:             domain,
			Workflow:           workflow,
			MatchingAttributes: testutils.ExecutionQueueAttributes,
		},
	}, response))
}

func TestDeleteWorkflowAttributes(t *testing.T) {
	request := admin.WorkflowAttributesDeleteRequest{
		Project:      project,
		Domain:       domain,
		Workflow:     workflow,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	db.ResourceRepo().(*mocks.MockResourceRepo).DeleteFunction = func(
		ctx context.Context, ID repoInterfaces.ResourceID) error {
		assert.Equal(t, project, ID.Project)
		assert.Equal(t, domain, ID.Domain)
		assert.Equal(t, workflow, ID.Workflow)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), ID.ResourceType)
		return nil
	}
	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	_, err := manager.DeleteWorkflowAttributes(context.Background(), request)
	assert.Nil(t, err)
}

func TestUpdateProjectDomainAttributes(t *testing.T) {
	request := admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            project,
			Domain:             domain,
			MatchingAttributes: testutils.ExecutionQueueAttributes,
		},
	}
	db := mocks.NewMockRepository()
	expectedSerializedAttrs, _ := proto.Marshal(testutils.ExecutionQueueAttributes)
	var createOrUpdateCalled bool
	db.ResourceRepo().(*mocks.MockResourceRepo).CreateOrUpdateFunction = func(
		ctx context.Context, input models.Resource) error {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, domain, input.Domain)
		assert.Equal(t, "", input.Workflow)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), input.ResourceType)
		assert.EqualValues(t, expectedSerializedAttrs, input.Attributes)
		createOrUpdateCalled = true
		return nil
	}
	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	_, err := manager.UpdateProjectDomainAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, createOrUpdateCalled)
}

func TestUpdateProjectDomainAttributes_CreateOrMerge(t *testing.T) {
	request := admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            project,
			Domain:             domain,
			MatchingAttributes: commonTestUtils.GetPluginOverridesAttributes(map[string][]string{"python": {"plugin a"}}),
		},
	}

	t.Run("create only", func(t *testing.T) {
		db := mocks.NewMockRepository()
		db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID repoInterfaces.ResourceID) (
			models.Resource, error) {
			return models.Resource{}, errors.NewFlyteAdminError(codes.NotFound, "foo")
		}
		var createOrUpdateCalled bool
		db.ResourceRepo().(*mocks.MockResourceRepo).CreateOrUpdateFunction = func(ctx context.Context, input models.Resource) error {
			assert.Equal(t, project, input.Project)
			assert.Equal(t, domain, input.Domain)

			var attributesToBeSaved admin.MatchingAttributes
			err := proto.Unmarshal(input.Attributes, &attributesToBeSaved)
			if err != nil {
				t.Fatal(err)
			}
			assert.Len(t, attributesToBeSaved.GetPluginOverrides().Overrides, 1)
			assert.True(t, proto.Equal(attributesToBeSaved.GetPluginOverrides().Overrides[0], &admin.PluginOverride{
				TaskType: python,
				PluginId: []string{"plugin a"}}))

			createOrUpdateCalled = true
			return nil
		}
		manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
		_, err := manager.UpdateProjectDomainAttributes(context.Background(), request)
		assert.NoError(t, err)
		assert.True(t, createOrUpdateCalled)
	})
	t.Run("merge update", func(t *testing.T) {
		db := mocks.NewMockRepository()
		db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID repoInterfaces.ResourceID) (
			models.Resource, error) {
			existingAttributes := commonTestUtils.GetPluginOverridesAttributes(map[string][]string{
				"hive":   {"plugin b"},
				"python": {"plugin c"},
			})
			bytes, err := proto.Marshal(existingAttributes)
			if err != nil {
				t.Fatal(err)
			}
			return models.Resource{
				Project:    project,
				Domain:     domain,
				Attributes: bytes,
			}, nil
		}
		var createOrUpdateCalled bool
		db.ResourceRepo().(*mocks.MockResourceRepo).CreateOrUpdateFunction = func(ctx context.Context, input models.Resource) error {
			assert.Equal(t, project, input.Project)
			assert.Equal(t, domain, input.Domain)

			var attributesToBeSaved admin.MatchingAttributes
			err := proto.Unmarshal(input.Attributes, &attributesToBeSaved)
			if err != nil {
				t.Fatal(err)
			}

			assert.Len(t, attributesToBeSaved.GetPluginOverrides().Overrides, 2)
			for _, override := range attributesToBeSaved.GetPluginOverrides().Overrides {
				if override.TaskType == python {
					assert.EqualValues(t, []string{"plugin a"}, override.PluginId)
				} else if override.TaskType == hive {
					assert.EqualValues(t, []string{"plugin b"}, override.PluginId)
				} else {
					t.Errorf("Unexpected task type [%s] plugin override committed to db", override.TaskType)
				}
			}
			createOrUpdateCalled = true
			return nil
		}
		manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
		_, err := manager.UpdateProjectDomainAttributes(context.Background(), request)
		assert.NoError(t, err)
		assert.True(t, createOrUpdateCalled)
	})
}

func TestGetProjectDomainAttributes(t *testing.T) {
	request := admin.ProjectDomainAttributesGetRequest{
		Project:      project,
		Domain:       domain,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(
		ctx context.Context, ID repoInterfaces.ResourceID) (models.Resource, error) {
		assert.Equal(t, project, ID.Project)
		assert.Equal(t, domain, ID.Domain)
		assert.Equal(t, "", ID.Workflow)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), ID.ResourceType)
		expectedSerializedAttrs, _ := proto.Marshal(testutils.ExecutionQueueAttributes)
		return models.Resource{
			Project:      project,
			Domain:       domain,
			ResourceType: "resource",
			Attributes:   expectedSerializedAttrs,
		}, nil
	}
	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	response, err := manager.GetProjectDomainAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            project,
			Domain:             domain,
			MatchingAttributes: testutils.ExecutionQueueAttributes,
		},
	}, response))
}

func TestDeleteProjectDomainAttributes(t *testing.T) {
	request := admin.ProjectDomainAttributesDeleteRequest{
		Project:      project,
		Domain:       domain,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	db.ResourceRepo().(*mocks.MockResourceRepo).DeleteFunction = func(
		ctx context.Context, ID repoInterfaces.ResourceID) error {
		assert.Equal(t, project, ID.Project)
		assert.Equal(t, domain, ID.Domain)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), ID.ResourceType)
		return nil
	}
	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	_, err := manager.DeleteProjectDomainAttributes(context.Background(), request)
	assert.Nil(t, err)
}

func TestUpdateProjectAttributes(t *testing.T) {
	request := admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{
			Project:            project,
			MatchingAttributes: testutils.WorkflowExecutionConfigSample,
		},
	}
	db := mocks.NewMockRepository()
	expectedSerializedAttrs, _ := proto.Marshal(testutils.WorkflowExecutionConfigSample)
	var createOrUpdateCalled bool
	db.ResourceRepo().(*mocks.MockResourceRepo).CreateOrUpdateFunction = func(
		ctx context.Context, input models.Resource) error {
		assert.Equal(t, project, input.Project)
		assert.Equal(t, "", input.Domain)
		assert.Equal(t, "", input.Workflow)
		assert.Equal(t, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG.String(), input.ResourceType)
		assert.EqualValues(t, expectedSerializedAttrs, input.Attributes)
		createOrUpdateCalled = true
		return nil
	}
	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	_, err := manager.UpdateProjectAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, createOrUpdateCalled)

	// Test empty attributes
	request = admin.ProjectAttributesUpdateRequest{Attributes: nil}
	_, err = manager.UpdateProjectAttributes(context.Background(), request)
	assert.Error(t, err)

	// Test error handling
	db.ResourceRepo().(*mocks.MockResourceRepo).CreateOrUpdateFunction = func(
		ctx context.Context, input models.Resource) error {
		return errors.NewFlyteAdminErrorf(123, "123")
	}
	request = admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{
			Project:            project,
			MatchingAttributes: testutils.WorkflowExecutionConfigSample,
		},
	}
	_, err = manager.UpdateProjectAttributes(context.Background(), request)
	assert.Error(t, err, "123")
}

func TestUpdateProjectAttributes_CreateOrMerge(t *testing.T) {
	request := admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{
			Project:            project,
			MatchingAttributes: commonTestUtils.GetPluginOverridesAttributes(map[string][]string{"python": {"plugin a"}}),
		},
	}

	t.Run("create only", func(t *testing.T) {
		db := mocks.NewMockRepository()
		db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID repoInterfaces.ResourceID) (
			models.Resource, error) {
			return models.Resource{}, errors.NewFlyteAdminError(codes.NotFound, "foo")
		}
		var createOrUpdateCalled bool
		db.ResourceRepo().(*mocks.MockResourceRepo).CreateOrUpdateFunction = func(ctx context.Context, input models.Resource) error {
			assert.Equal(t, project, input.Project)
			assert.Equal(t, "", input.Domain)

			var attributesToBeSaved admin.MatchingAttributes
			err := proto.Unmarshal(input.Attributes, &attributesToBeSaved)
			if err != nil {
				t.Fatal(err)
			}
			assert.Len(t, attributesToBeSaved.GetPluginOverrides().Overrides, 1)
			assert.True(t, proto.Equal(attributesToBeSaved.GetPluginOverrides().Overrides[0], &admin.PluginOverride{
				TaskType: python,
				PluginId: []string{"plugin a"}}))

			createOrUpdateCalled = true
			return nil
		}
		manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
		_, err := manager.UpdateProjectAttributes(context.Background(), request)
		assert.NoError(t, err)
		assert.True(t, createOrUpdateCalled)
	})
	t.Run("merge update", func(t *testing.T) {
		db := mocks.NewMockRepository()
		db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID repoInterfaces.ResourceID) (
			models.Resource, error) {
			existingAttributes := commonTestUtils.GetPluginOverridesAttributes(map[string][]string{
				"hive":   {"plugin b"},
				"python": {"plugin c"},
			})
			bytes, err := proto.Marshal(existingAttributes)
			if err != nil {
				t.Fatal(err)
			}
			return models.Resource{
				Project:    project,
				Attributes: bytes,
			}, nil
		}
		var createOrUpdateCalled bool
		db.ResourceRepo().(*mocks.MockResourceRepo).CreateOrUpdateFunction = func(ctx context.Context, input models.Resource) error {
			assert.Equal(t, project, input.Project)
			assert.Equal(t, "", input.Domain)

			var attributesToBeSaved admin.MatchingAttributes
			err := proto.Unmarshal(input.Attributes, &attributesToBeSaved)
			if err != nil {
				t.Fatal(err)
			}

			assert.Len(t, attributesToBeSaved.GetPluginOverrides().Overrides, 2)
			for _, override := range attributesToBeSaved.GetPluginOverrides().Overrides {
				if override.TaskType == python {
					assert.EqualValues(t, []string{"plugin a"}, override.PluginId)
				} else if override.TaskType == hive {
					assert.EqualValues(t, []string{"plugin b"}, override.PluginId)
				} else {
					t.Errorf("Unexpected task type [%s] plugin override committed to db", override.TaskType)
				}
			}
			createOrUpdateCalled = true
			return nil
		}
		manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
		_, err := manager.UpdateProjectAttributes(context.Background(), request)
		assert.NoError(t, err)
		assert.True(t, createOrUpdateCalled)
	})
}

func TestGetProjectAttributes(t *testing.T) {
	request := admin.ProjectAttributesGetRequest{
		Project:      project,
		ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
	}
	db := mocks.NewMockRepository()

	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(
		ctx context.Context, ID repoInterfaces.ResourceID) (models.Resource, error) {

		assert.Equal(t, project, ID.Project)
		assert.Equal(t, "", ID.Domain)
		assert.Equal(t, "", ID.Workflow)
		assert.Equal(t, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG.String(), ID.ResourceType)
		expectedSerializedAttrs, _ := proto.Marshal(testutils.WorkflowExecutionConfigSample)
		return models.Resource{
			Project:      project,
			Domain:       "",
			ResourceType: "resource",
			Attributes:   expectedSerializedAttrs,
		}, nil
	}
	response, err := manager.GetProjectAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectAttributesGetResponse{
		Attributes: &admin.ProjectAttributes{
			Project:            project,
			MatchingAttributes: testutils.WorkflowExecutionConfigSample,
		},
	}, response))

	// unrecognized errors are thrown
	db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(
		ctx context.Context, ID repoInterfaces.ResourceID) (models.Resource, error) {

		return models.Resource{}, errors.NewFlyteAdminErrorf(5323, "random code")
	}
	_, err = manager.GetProjectAttributes(context.Background(), request)
	assert.Error(t, err)
}

func TestGetProjectAttributes_ConfigLookup(t *testing.T) {
	request := admin.ProjectAttributesGetRequest{
		Project:      project,
		ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
	}
	db := mocks.NewMockRepository()
	db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(
		ctx context.Context, ID repoInterfaces.ResourceID) (models.Resource, error) {
		// return not found to trigger loading from config
		return models.Resource{}, errors.NewFlyteAdminError(codes.NotFound, "not found message")
	}
	config := runtimeMocks.MockApplicationProvider{}
	manager := NewResourceManager(db, &config)

	t.Run("config 1", func(t *testing.T) {
		appConfig := runtimeInterfaces.ApplicationConfig{
			MaxParallelism:       3,
			K8SServiceAccount:    "testserviceaccount",
			Labels:               map[string]string{"lab1": "name"},
			OutputLocationPrefix: "s3://test-bucket",
		}
		config.SetTopLevelConfig(appConfig)

		response, err := manager.GetProjectAttributes(context.Background(), request)
		assert.Nil(t, err)
		assert.True(t, proto.Equal(&admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Project: project,
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
							MaxParallelism: 3,
							SecurityContext: &core.SecurityContext{
								RunAs: &core.Identity{K8SServiceAccount: "testserviceaccount"},
							},
							RawOutputDataConfig: &admin.RawOutputDataConfig{
								OutputLocationPrefix: "s3://test-bucket",
							},
							Labels: &admin.Labels{
								Values: map[string]string{"lab1": "name"},
							},
						},
					},
				},
			},
		}, response))
	})

	t.Run("config 2", func(t *testing.T) {
		appConfig := runtimeInterfaces.ApplicationConfig{
			MaxParallelism:   3,
			AssumableIamRole: "myrole",
		}
		config.SetTopLevelConfig(appConfig)

		response, err := manager.GetProjectAttributes(context.Background(), request)
		assert.Nil(t, err)
		assert.True(t, proto.Equal(&admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Project: project,
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
							MaxParallelism: 3,
							SecurityContext: &core.SecurityContext{
								RunAs: &core.Identity{IamRole: "myrole"},
							},
						},
					},
				},
			},
		}, response))
	})

	t.Run("config 3", func(t *testing.T) {
		appConfig := runtimeInterfaces.ApplicationConfig{
			MaxParallelism: 3,
			Annotations:    map[string]string{"ann1": "val1"},
		}
		config.SetTopLevelConfig(appConfig)

		response, err := manager.GetProjectAttributes(context.Background(), request)
		assert.Nil(t, err)
		assert.True(t, proto.Equal(&admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Project: project,
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
							MaxParallelism: 3,
							Annotations: &admin.Annotations{
								Values: map[string]string{"ann1": "val1"},
							},
						},
					},
				},
			},
		}, response))
	})

	t.Run("config not merged if not wec", func(t *testing.T) {
		appConfig := runtimeInterfaces.ApplicationConfig{
			MaxParallelism:       3,
			K8SServiceAccount:    "testserviceaccount",
			Labels:               map[string]string{"lab1": "name"},
			OutputLocationPrefix: "s3://test-bucket",
		}
		config.SetTopLevelConfig(appConfig)
		request := admin.ProjectAttributesGetRequest{
			Project:      project,
			ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
		}

		_, err := manager.GetProjectAttributes(context.Background(), request)
		assert.Error(t, err)
		ec, ok := err.(errors.FlyteAdminError)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, ec.Code())
	})
}

func TestDeleteProjectAttributes(t *testing.T) {
	request := admin.ProjectAttributesDeleteRequest{
		Project:      project,
		ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
	}
	db := mocks.NewMockRepository()
	db.ResourceRepo().(*mocks.MockResourceRepo).DeleteFunction = func(
		ctx context.Context, ID repoInterfaces.ResourceID) error {
		assert.Equal(t, project, ID.Project)
		assert.Equal(t, "", ID.Domain)
		assert.Equal(t, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG.String(), ID.ResourceType)
		return nil
	}
	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	_, err := manager.DeleteProjectAttributes(context.Background(), request)
	assert.Nil(t, err)
}

func TestGetResource(t *testing.T) {
	request := interfaces.ResourceRequest{
		Project:      project,
		Domain:       domain,
		Workflow:     workflow,
		LaunchPlan:   "launch_plan",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(
		ctx context.Context, ID repoInterfaces.ResourceID) (models.Resource, error) {
		assert.Equal(t, project, ID.Project)
		assert.Equal(t, domain, ID.Domain)
		assert.Equal(t, workflow, ID.Workflow)
		assert.Equal(t, "launch_plan", ID.LaunchPlan)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), ID.ResourceType)
		expectedSerializedAttrs, _ := proto.Marshal(testutils.ExecutionQueueAttributes)
		return models.Resource{
			Project:      ID.Project,
			Domain:       ID.Domain,
			Workflow:     ID.Workflow,
			LaunchPlan:   ID.LaunchPlan,
			ResourceType: ID.ResourceType,
			Attributes:   expectedSerializedAttrs,
		}, nil
	}
	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	response, err := manager.GetResource(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, request.Project, response.Project)
	assert.Equal(t, request.Domain, response.Domain)
	assert.Equal(t, request.Workflow, response.Workflow)
	assert.Equal(t, request.LaunchPlan, response.LaunchPlan)
	assert.Equal(t, request.ResourceType.String(), response.ResourceType)
	assert.True(t, proto.Equal(response.Attributes, testutils.ExecutionQueueAttributes))
}

func TestListAllResources(t *testing.T) {
	db := mocks.NewMockRepository()
	projectAttributes := admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ClusterResourceAttributes{
			ClusterResourceAttributes: &admin.ClusterResourceAttributes{
				Attributes: map[string]string{
					"foo": "foofoo",
				},
			},
		},
	}
	marshaledProjectAttrs, _ := proto.Marshal(&projectAttributes)
	workflowAttributes := admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ClusterResourceAttributes{
			ClusterResourceAttributes: &admin.ClusterResourceAttributes{
				Attributes: map[string]string{
					"bar": "barbar",
				},
			},
		},
	}
	marshaledWorkflowAttrs, _ := proto.Marshal(&workflowAttributes)
	db.ResourceRepo().(*mocks.MockResourceRepo).ListAllFunction = func(ctx context.Context, resourceType string) (
		[]models.Resource, error) {
		assert.Equal(t, admin.MatchableResource_CLUSTER_RESOURCE.String(), resourceType)
		return []models.Resource{
			{
				Project:      "projectA",
				ResourceType: admin.MatchableResource_CLUSTER_RESOURCE.String(),
				Attributes:   marshaledProjectAttrs,
			},
			{
				Project:      "projectB",
				Domain:       "development",
				Workflow:     "workflow",
				ResourceType: admin.MatchableResource_CLUSTER_RESOURCE.String(),
				Attributes:   marshaledWorkflowAttrs,
			},
		}, nil
	}
	manager := NewResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains())
	response, err := manager.ListAll(context.Background(), admin.ListMatchableAttributesRequest{
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
	})
	assert.Nil(t, err)
	assert.NotNil(t, response.Configurations)
	assert.Len(t, response.Configurations, 2)
	assert.True(t, proto.Equal(&admin.MatchableAttributesConfiguration{
		Project:    "projectA",
		Attributes: &projectAttributes,
	}, response.Configurations[0]))
	assert.True(t, proto.Equal(&admin.MatchableAttributesConfiguration{
		Project:    "projectB",
		Domain:     "development",
		Workflow:   "workflow",
		Attributes: &workflowAttributes,
	}, response.Configurations[1]))
}
