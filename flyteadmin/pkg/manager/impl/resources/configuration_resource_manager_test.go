package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	managerMocks "github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const version = "v1"

func TestUpdateWorkflowAttributes_Configuration(t *testing.T) {
	request := &admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Org:                org,
			Project:            project,
			Domain:             domain,
			Workflow:           workflow,
			MatchingAttributes: testutils.UpdateTaskResourceAttributes,
		},
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	expectedConfigurationUpdateRequest := &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org:      org,
			Project:  project,
			Domain:   domain,
			Workflow: workflow,
		},
		VersionToUpdate: version,
		Configuration: &admin.Configuration{
			ExecutionQueueAttributes: testutils.WorkflowExecutionQueueAttributes.GetExecutionQueueAttributes(),
			TaskResourceAttributes:   testutils.UpdateTaskResourceAttributes.GetTaskResourceAttributes(),
		},
	}
	mockConfigManager.On("UpdateConfiguration", mock.Anything, expectedConfigurationUpdateRequest).Return(&admin.ConfigurationUpdateResponse{}, nil)
	_, err = manager.UpdateWorkflowAttributes(context.Background(), request)
	assert.NoError(t, err)
	mockConfigManager.AssertExpectations(t)
}

func TestGetWorkflowAttributes_Configuration(t *testing.T) {
	request := &admin.WorkflowAttributesGetRequest{
		Org:          org,
		Project:      project,
		Domain:       domain,
		Workflow:     workflow,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	response, err := manager.GetWorkflowAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.WorkflowAttributesGetResponse{
		Attributes: &admin.WorkflowAttributes{
			Org:                org,
			Project:            project,
			Domain:             domain,
			Workflow:           workflow,
			MatchingAttributes: testutils.WorkflowExecutionQueueAttributes,
		},
	}, response))
	mockConfigManager.AssertExpectations(t)
}

func TestDeleteWorkflowAttributes_Configuration(t *testing.T) {
	request := &admin.WorkflowAttributesDeleteRequest{
		Org:          org,
		Project:      project,
		Domain:       domain,
		Workflow:     workflow,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	expectedConfigurationUpdateRequest := &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org:      org,
			Project:  project,
			Domain:   domain,
			Workflow: workflow,
		},
		VersionToUpdate: version,
		Configuration: &admin.Configuration{
			TaskResourceAttributes: testutils.WorkflowTaskResourceAttributes.GetTaskResourceAttributes(),
		},
	}
	mockConfigManager.On("UpdateConfiguration", mock.Anything, expectedConfigurationUpdateRequest).Return(&admin.ConfigurationUpdateResponse{}, nil)
	_, err = manager.DeleteWorkflowAttributes(context.Background(), request)
	assert.Nil(t, err)
	mockConfigManager.AssertExpectations(t)
}

func TestUpdateProjectDomainAttributes_Configuration(t *testing.T) {
	request := &admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Org:                org,
			Project:            project,
			Domain:             domain,
			MatchingAttributes: testutils.UpdateTaskResourceAttributes,
		},
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	expectedConfigurationUpdateRequest := &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org:     org,
			Project: project,
			Domain:  domain,
		},
		VersionToUpdate: version,
		Configuration: &admin.Configuration{
			ExecutionQueueAttributes: testutils.ProjectDomainExecutionQueueAttributes.GetExecutionQueueAttributes(),
			TaskResourceAttributes:   testutils.UpdateTaskResourceAttributes.GetTaskResourceAttributes(),
		},
	}
	mockConfigManager.On("UpdateConfiguration", mock.Anything, expectedConfigurationUpdateRequest).Return(&admin.ConfigurationUpdateResponse{}, nil)
	_, err = manager.UpdateProjectDomainAttributes(context.Background(), request)
	assert.NoError(t, err)
	mockConfigManager.AssertExpectations(t)
}

func TestGetProjectDomainAttributes_Configuration(t *testing.T) {
	request := &admin.ProjectDomainAttributesGetRequest{
		Org:          org,
		Project:      project,
		Domain:       domain,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	response, err := manager.GetProjectDomainAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Org:                org,
			Project:            project,
			Domain:             domain,
			MatchingAttributes: testutils.ProjectDomainExecutionQueueAttributes,
		},
	}, response))
	mockConfigManager.AssertExpectations(t)
}

func TestDeleteProjectDomainAttributes_Configuration(t *testing.T) {
	request := &admin.ProjectDomainAttributesDeleteRequest{
		Org:          org,
		Project:      project,
		Domain:       domain,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	expectedConfigurationUpdateRequest := &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org:     org,
			Project: project,
			Domain:  domain,
		},
		VersionToUpdate: version,
		Configuration: &admin.Configuration{
			TaskResourceAttributes: testutils.ProjectDomainTaskResourceAttributes.GetTaskResourceAttributes(),
		},
	}
	mockConfigManager.On("UpdateConfiguration", mock.Anything, expectedConfigurationUpdateRequest).Return(&admin.ConfigurationUpdateResponse{}, nil)
	_, err = manager.DeleteProjectDomainAttributes(context.Background(), request)
	assert.Nil(t, err)
	mockConfigManager.AssertExpectations(t)
}

func TestUpdateProjectAttributes_Configuration(t *testing.T) {
	request := &admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{
			Org:                org,
			Project:            project,
			MatchingAttributes: testutils.UpdateTaskResourceAttributes,
		},
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	expectedConfigurationUpdateRequest := &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org:     org,
			Project: project,
		},
		VersionToUpdate: version,
		Configuration: &admin.Configuration{
			ExecutionQueueAttributes: testutils.ProjectExecutionQueueAttributes.GetExecutionQueueAttributes(),
			TaskResourceAttributes:   testutils.UpdateTaskResourceAttributes.GetTaskResourceAttributes(),
		},
	}
	mockConfigManager.On("UpdateConfiguration", mock.Anything, expectedConfigurationUpdateRequest).Return(&admin.ConfigurationUpdateResponse{}, nil)
	_, err = manager.UpdateProjectAttributes(context.Background(), request)
	assert.NoError(t, err)
	mockConfigManager.AssertExpectations(t)
}

func TestGetProjectAttributes_Configuration(t *testing.T) {
	request := &admin.ProjectAttributesGetRequest{
		Org:          org,
		Project:      project,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	response, err := manager.GetProjectAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectAttributesGetResponse{
		Attributes: &admin.ProjectAttributes{
			Org:                org,
			Project:            project,
			MatchingAttributes: testutils.ProjectExecutionQueueAttributes,
		},
	}, response))
	mockConfigManager.AssertExpectations(t)
}

func TestGetProjectAttributes_ConfigLookup_Configuration(t *testing.T) {
	request := &admin.ProjectAttributesGetRequest{
		Org:          org,
		Project:      project,
		ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
	}

	db := mocks.NewMockRepository()

	t.Run("config 1", func(t *testing.T) {
		mockConfigManager := new(managerMocks.ConfigurationInterface)
		manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)
		mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
		assert.Nil(t, err)
		key, err := testutils.EncodeDocumentKey(&util.GlobalConfigurationKey)
		assert.Nil(t, err)
		mockConfigurations[key] = &admin.Configuration{
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
		}
		mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
			Configurations: mockConfigurations,
			Version:        version,
		}, nil)

		response, err := manager.GetProjectAttributes(context.Background(), request)
		fmt.Println(err)
		assert.Nil(t, err)
		assert.True(t, proto.Equal(&admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Org:     org,
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
		mockConfigManager.AssertExpectations(t)
	})

	t.Run("config 2", func(t *testing.T) {
		mockConfigManager := new(managerMocks.ConfigurationInterface)
		manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)
		mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
		assert.Nil(t, err)
		key, err := testutils.EncodeDocumentKey(&util.GlobalConfigurationKey)
		assert.Nil(t, err)
		mockConfigurations[key] = &admin.Configuration{
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
				MaxParallelism: 3,
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{IamRole: "myrole"},
				},
			},
		}
		mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
			Configurations: mockConfigurations,
			Version:        version,
		}, nil)

		response, err := manager.GetProjectAttributes(context.Background(), request)
		assert.Nil(t, err)
		fmt.Println(response)
		assert.True(t, proto.Equal(&admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Org:     org,
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
		mockConfigManager.AssertExpectations(t)
	})

	t.Run("config 3", func(t *testing.T) {
		mockConfigManager := new(managerMocks.ConfigurationInterface)
		manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)
		mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
		assert.Nil(t, err)
		key, err := testutils.EncodeDocumentKey(&util.GlobalConfigurationKey)
		assert.Nil(t, err)
		mockConfigurations[key] = &admin.Configuration{
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
				MaxParallelism: 3,
				Annotations: &admin.Annotations{
					Values: map[string]string{"ann1": "val1"},
				},
			},
		}
		mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
			Configurations: mockConfigurations,
			Version:        version,
		}, nil)

		response, err := manager.GetProjectAttributes(context.Background(), request)
		assert.Nil(t, err)
		fmt.Println(response)
		assert.True(t, proto.Equal(&admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Org:     org,
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
		mockConfigManager.AssertExpectations(t)
	})

	t.Run("config not merged if not wec", func(t *testing.T) {
		mockConfigManager := new(managerMocks.ConfigurationInterface)
		manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)
		mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
		assert.Nil(t, err)
		key, err := testutils.EncodeDocumentKey(&util.GlobalConfigurationKey)
		assert.Nil(t, err)
		mockConfigurations[key] = &admin.Configuration{
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
		}
		mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
			Configurations: mockConfigurations,
			Version:        version,
		}, nil)

		request := &admin.ProjectAttributesGetRequest{
			Org:          org,
			Project:      project,
			ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
		}

		_, err = manager.GetProjectAttributes(context.Background(), request)
		assert.Error(t, err)
		ec, ok := err.(errors.FlyteAdminError)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, ec.Code())
		mockConfigManager.AssertExpectations(t)
	})
}

func TestDeleteProjectAttributes_Configuration(t *testing.T) {
	request := &admin.ProjectAttributesDeleteRequest{
		Org:          org,
		Project:      project,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	expectedConfigurationUpdateRequest := &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org:     org,
			Project: project,
		},
		VersionToUpdate: version,
		Configuration: &admin.Configuration{
			TaskResourceAttributes: testutils.ProjectTaskResourceAttributes.GetTaskResourceAttributes(),
		},
	}
	mockConfigManager.On("UpdateConfiguration", mock.Anything, expectedConfigurationUpdateRequest).Return(&admin.ConfigurationUpdateResponse{}, nil)
	_, err = manager.DeleteProjectAttributes(context.Background(), request)
	assert.Nil(t, err)
	mockConfigManager.AssertExpectations(t)
}

func TestUpdateOrgAttributes_Configuration(t *testing.T) {
	request := &admin.OrgAttributesUpdateRequest{
		Attributes: &admin.OrgAttributes{
			Org:                org,
			MatchingAttributes: testutils.UpdateTaskResourceAttributes,
		},
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	expectedConfigurationUpdateRequest := &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org: org,
		},
		VersionToUpdate: version,
		Configuration: &admin.Configuration{
			ExecutionQueueAttributes: testutils.OrgExecutionQueueAttributes.GetExecutionQueueAttributes(),
			TaskResourceAttributes:   testutils.UpdateTaskResourceAttributes.GetTaskResourceAttributes(),
		},
	}
	mockConfigManager.On("UpdateConfiguration", mock.Anything, expectedConfigurationUpdateRequest).Return(&admin.ConfigurationUpdateResponse{}, nil)
	_, err = manager.UpdateOrgAttributes(context.Background(), request)
	assert.NoError(t, err)
	mockConfigManager.AssertExpectations(t)
}

func TestGetOrgAttributes_Configuration(t *testing.T) {
	request := &admin.OrgAttributesGetRequest{
		Org:          org,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	response, err := manager.GetOrgAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.OrgAttributesGetResponse{
		Attributes: &admin.OrgAttributes{
			Org:                org,
			MatchingAttributes: testutils.OrgExecutionQueueAttributes,
		},
	}, response))
	mockConfigManager.AssertExpectations(t)
}

func TestDeleteOrgAttributes_Configuration(t *testing.T) {
	request := &admin.OrgAttributesDeleteRequest{
		Org:          org,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	expectedConfigurationUpdateRequest := &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org: org,
		},
		VersionToUpdate: version,
		Configuration: &admin.Configuration{
			TaskResourceAttributes: testutils.OrgTaskResourceAttributes.GetTaskResourceAttributes(),
		},
	}
	mockConfigManager.On("UpdateConfiguration", mock.Anything, expectedConfigurationUpdateRequest).Return(&admin.ConfigurationUpdateResponse{}, nil)
	_, err = manager.DeleteOrgAttributes(context.Background(), request)
	assert.Nil(t, err)
	mockConfigManager.AssertExpectations(t)
}

func TestGetResource_Configuration(t *testing.T) {
	request := interfaces.ResourceRequest{
		Org:          org,
		Project:      project,
		Domain:       domain,
		Workflow:     workflow,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	response, err := manager.GetResource(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, request.Org, response.Org)
	assert.Equal(t, request.Project, response.Project)
	assert.Equal(t, request.Domain, response.Domain)
	assert.Equal(t, request.Workflow, response.Workflow)
	assert.Equal(t, request.ResourceType.String(), response.ResourceType)
	assert.True(t, proto.Equal(response.Attributes, testutils.WorkflowExecutionQueueAttributes))
	mockConfigManager.AssertExpectations(t)
}

func TestListAll_Configuration(t *testing.T) {
	db := mocks.NewMockRepository()
	mockConfigManager := new(managerMocks.ConfigurationInterface)
	manager := NewConfigurationResourceManager(db, testutils.GetApplicationConfigWithDefaultDomains(), mockConfigManager)

	mockConfigurations, err := testutils.MockConfigurations(org, project, domain, workflow)
	assert.Nil(t, err)

	mockConfigManager.On("GetReadOnlyActiveDocument", mock.Anything).Return(&admin.ConfigurationDocument{
		Configurations: mockConfigurations,
		Version:        version,
	}, nil)

	response, err := manager.ListAll(context.Background(), &admin.ListMatchableAttributesRequest{
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	})
	assert.Nil(t, err)
	assert.NotNil(t, response.Configurations)
	assert.Len(t, response.Configurations, 4)
	assert.True(t, proto.Equal(&admin.MatchableAttributesConfiguration{
		Org:        org,
		Project:    project,
		Domain:     domain,
		Workflow:   workflow,
		Attributes: testutils.WorkflowExecutionQueueAttributes,
	}, response.Configurations[0]))
	assert.True(t, proto.Equal(&admin.MatchableAttributesConfiguration{
		Org:        org,
		Project:    project,
		Domain:     domain,
		Attributes: testutils.ProjectDomainExecutionQueueAttributes,
	}, response.Configurations[1]))
	assert.True(t, proto.Equal(&admin.MatchableAttributesConfiguration{
		Org:        org,
		Project:    project,
		Attributes: testutils.ProjectExecutionQueueAttributes,
	}, response.Configurations[2]))
	assert.True(t, proto.Equal(&admin.MatchableAttributesConfiguration{
		Org:        org,
		Attributes: testutils.OrgExecutionQueueAttributes,
	}, response.Configurations[3]))
	mockConfigManager.AssertExpectations(t)
}
