package configurations

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/configurations/plugin"
	utilMocks "github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/configurations/plugin/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/flytestdlib/storage/mocks"
)

const s3Path = "s3://bucket/key"

func TestGetReadOnlyActiveDocument(t *testing.T) {
	ctx := context.Background()
	db := mocks.NewMockRepository()
	mockConfig := &runtimeMocks.Configuration{}
	mockPBStore := &storageMocks.ComposedProtobufStore{}
	mockStorage := &storage.DataStore{
		ComposedProtobufStore: mockPBStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, plugins.NewRegistry(), ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: s3Path,
		Active:           true,
	}, nil)
	mockPBStore.On("ReadProtobuf", mock.Anything, mock.MatchedBy(func(reference storage.DataReference) bool {
		return reference.String() == s3Path
	}), mock.AnythingOfType("*admin.ConfigurationDocument")).Return(nil).Run(func(args mock.Arguments) {
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Version = "v1"
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Configurations = map[string]*admin.Configuration{
			"key": {
				TaskResourceAttributes: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu: "1",
						Gpu: "2",
					},
					Limits: &admin.TaskResourceSpec{
						Cpu: "3",
						Gpu: "4",
					},
				},
			},
		}
	})

	activeDocument, err := configurationManager.GetReadOnlyActiveDocument(ctx)
	assert.Nil(t, err)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).AssertExpectations(t)
	mockPBStore.AssertExpectations(t)
	assert.True(t, proto.Equal(&activeDocument, &admin.ConfigurationDocument{
		Version: "v1",
		Configurations: map[string]*admin.Configuration{
			"key": {
				TaskResourceAttributes: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu: "1",
						Gpu: "2",
					},
					Limits: &admin.TaskResourceSpec{
						Cpu: "3",
						Gpu: "4",
					},
				},
			},
		},
	}))
}

func TestGetEditableActiveDocument(t *testing.T) {
	ctx := context.Background()
	db := mocks.NewMockRepository()
	mockConfig := &runtimeMocks.Configuration{}
	mockPBStore := &storageMocks.ComposedProtobufStore{}
	mockStorage := &storage.DataStore{
		ComposedProtobufStore: mockPBStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}
	// mock config
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, plugins.NewRegistry(), ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: s3Path,
		Active:           true,
	}, nil)
	mockPBStore.On("ReadProtobuf", mock.Anything, mock.MatchedBy(func(reference storage.DataReference) bool {
		return reference.String() == s3Path
	}), mock.AnythingOfType("*admin.ConfigurationDocument")).Return(nil).Run(func(args mock.Arguments) {
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Version = "v1"
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Configurations = map[string]*admin.Configuration{
			"key": {
				TaskResourceAttributes: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu: "1",
						Gpu: "2",
					},
					Limits: &admin.TaskResourceSpec{
						Cpu: "3",
						Gpu: "4",
					},
				},
			},
		}
	})

	activeDocument, err := configurationManager.GetEditableActiveDocument(ctx)
	assert.Nil(t, err)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).AssertExpectations(t)
	mockPBStore.AssertExpectations(t)
	assert.True(t, proto.Equal(&activeDocument, &admin.ConfigurationDocument{
		Version: "v1",
		Configurations: map[string]*admin.Configuration{
			"key": {
				TaskResourceAttributes: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu: "1",
						Gpu: "2",
					},
					Limits: &admin.TaskResourceSpec{
						Cpu: "3",
						Gpu: "4",
					},
				},
			},
		},
	}))
}

func TestGetActiveDocument_DBError(t *testing.T) {
	ctx := context.Background()
	db := mocks.NewMockRepository()
	mockConfig := &runtimeMocks.Configuration{}
	mockStorage := &storage.DataStore{
		ComposedProtobufStore: &storageMocks.ComposedProtobufStore{},
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}
	// mock config
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, plugins.NewRegistry(), ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{}, assert.AnError)

	activeDocument, err := configurationManager.GetReadOnlyActiveDocument(ctx)
	assert.NotNil(t, err)
	assert.True(t, proto.Equal(&activeDocument, &admin.ConfigurationDocument{}))
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).AssertExpectations(t)
}

func TestGetActiveDocument_StoreError(t *testing.T) {
	ctx := context.Background()
	db := mocks.NewMockRepository()
	mockConfig := &runtimeMocks.Configuration{}
	mockPBStore := &storageMocks.ComposedProtobufStore{}
	mockStorage := &storage.DataStore{
		ComposedProtobufStore: mockPBStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}
	// mock config
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, plugins.NewRegistry(), ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: s3Path,
		Active:           true,
	}, nil)
	mockPBStore.On("ReadProtobuf", mock.Anything, mock.Anything, mock.AnythingOfType("*admin.ConfigurationDocument")).Return(assert.AnError)

	activeDocument, err := configurationManager.GetReadOnlyActiveDocument(ctx)
	assert.NotNil(t, err)
	assert.True(t, proto.Equal(&activeDocument, &admin.ConfigurationDocument{}))
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).AssertExpectations(t)
	mockPBStore.AssertExpectations(t)
}

func TestGetConfiguration(t *testing.T) {
	ctx := context.Background()
	db := mocks.NewMockRepository()
	mockConfig := &runtimeMocks.Configuration{}
	mockPBStore := &storageMocks.ComposedProtobufStore{}
	mockStorage := &storage.DataStore{
		ComposedProtobufStore: mockPBStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}
	mockProjectConfigurationPlugin := utilMocks.MockProjectConfigurationPlugin{}
	pluginRegistry := plugins.NewRegistry()
	pluginRegistry.RegisterDefault(plugins.PluginIDProjectConfiguration, &mockProjectConfigurationPlugin)
	// Mock config
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	applicationConfig.On("GetDomainsConfig").Return(&runtimeInterfaces.DomainsConfig{
		runtimeInterfaces.Domain{
			ID:   "domain",
			Name: "domain",
		},
	})
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, pluginRegistry, ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	// Mock repo
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: s3Path,
		Active:           true,
	}, nil)
	// Mock store
	mockPBStore.On("ReadProtobuf", mock.Anything, mock.MatchedBy(func(reference storage.DataReference) bool {
		return reference.String() == s3Path
	}), mock.AnythingOfType("*admin.ConfigurationDocument")).Return(nil).Run(func(args mock.Arguments) {
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Version = "v1"
		configurations := make(map[string]*admin.Configuration)
		projectDomainKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{
			Org:     "org",
			Project: "project",
			Domain:  "domain",
		})
		assert.Nil(t, err)
		configurations[projectDomainKey] = &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "1",
					Gpu: "2",
				},
			},
		}
		projectKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{
			Org:     "org",
			Project: "project",
		})
		assert.Nil(t, err)
		configurations[projectKey] = &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "5",
					Gpu: "6",
				},
			},
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
				MaxParallelism: 1,
			},
		}
		orgKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{
			Org: "org",
		})
		assert.Nil(t, err)
		configurations[orgKey] = &admin.Configuration{
			ClusterAssignment: &admin.ClusterAssignment{
				ClusterPoolName: "cluster",
			},
		}
		globalKey, err := util.EncodeConfigurationDocumentKey(ctx, &util.GlobalConfigurationKey)
		assert.Nil(t, err)
		configurations[globalKey] = &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "7",
					Gpu: "8",
				},
			},
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
				Tags: []string{
					"foo", "bar", "baz",
				},
			},
		}
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Configurations = configurations
	})
	// Mock plugin
	mockProjectConfigurationPlugin.SetGetAttributeIsMutableCallback(func(ctx context.Context, input *plugin.GetAttributeIsMutable) (map[admin.MatchableResource]*admin.AttributeIsMutable, error) {
		return map[admin.MatchableResource]*admin.AttributeIsMutable{
			admin.MatchableResource_TASK_RESOURCE: {
				Value: true,
			},
			admin.MatchableResource_CLUSTER_RESOURCE: {
				Value: true,
			},
			admin.MatchableResource_EXECUTION_QUEUE: {
				Value: true,
			},
			admin.MatchableResource_EXECUTION_CLUSTER_LABEL: {
				Value: true,
			},
			admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_PLUGIN_OVERRIDE: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_CLUSTER_ASSIGNMENT: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_EXTERNAL_RESOURCE: {
				Value:  false,
				Reason: "Cannot be modified",
			},
		}, nil
	})

	response, err := configurationManager.GetConfiguration(ctx, admin.ConfigurationGetRequest{
		Id: &admin.ConfigurationID{
			Org:     "org",
			Project: "project",
			Domain:  "domain",
		},
	})
	assert.Nil(t, err)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).AssertExpectations(t)
	applicationConfig.AssertExpectations(t)
	mockPBStore.AssertExpectations(t)
	assert.True(t, proto.Equal(response, &admin.ConfigurationGetResponse{
		Id:      &admin.ConfigurationID{Org: "org", Project: "project", Domain: "domain"},
		Version: "v1",
		Configuration: &admin.ConfigurationWithSource{
			TaskResourceAttributes: &admin.TaskResourceAttributesWithSource{
				Source: admin.AttributesSource_PROJECT_DOMAIN,
				Value: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu: "1",
						Gpu: "2",
					},
				},
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			ClusterResourceAttributes: &admin.ClusterResourceAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value: &admin.ExecutionQueueAttributes{
					Tags: []string{
						"foo", "bar", "baz",
					},
				},
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			ExecutionClusterLabel: &admin.ExecutionClusterLabelWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			QualityOfService: &admin.QualityOfServiceWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			PluginOverrides: &admin.PluginOverridesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfigWithSource{
				Source: admin.AttributesSource_PROJECT,
				Value: &admin.WorkflowExecutionConfig{
					MaxParallelism: 1,
				},
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			ClusterAssignment: &admin.ClusterAssignmentWithSource{
				Source: admin.AttributesSource_ORG,
				Value: &admin.ClusterAssignment{
					ClusterPoolName: "cluster",
				},
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			ExternalResourceAttributes: &admin.ExternalResourceAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
		},
	}))
}

func TestUpdateProjectDomainConfiguration(t *testing.T) {
	ctx := context.Background()
	db := mocks.NewMockRepository()
	mockConfig := &runtimeMocks.Configuration{}
	mockPBStore := &storageMocks.ComposedProtobufStore{}
	mockRefConstructor := &storageMocks.ReferenceConstructor{}
	mockStorage := &storage.DataStore{
		ComposedProtobufStore: mockPBStore,
		ReferenceConstructor:  mockRefConstructor,
	}
	mockProjectConfigurationPlugin := utilMocks.MockProjectConfigurationPlugin{}
	pluginRegistry := plugins.NewRegistry()
	pluginRegistry.RegisterDefault(plugins.PluginIDProjectConfiguration, &mockProjectConfigurationPlugin)
	// Mock config
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	applicationConfig.On("GetDomainsConfig").Return(&runtimeInterfaces.DomainsConfig{
		runtimeInterfaces.Domain{
			ID:   "domain",
			Name: "domain",
		},
	})
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, pluginRegistry, ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	// Mock repo
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: s3Path,
		Active:           true,
	}, nil)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("Update", mock.Anything, mock.MatchedBy(func(input *interfaces.UpdateConfigurationInput) bool {
		if input.VersionToUpdate != "v1" {
			return false
		}
		return *input.NewConfigurationMetadata == models.ConfigurationDocumentMetadata{
			Version:          "/2q3/pKW4h7lf1uExxhntIXaoI3dj0wL+JwEIrTIHUg=",
			DocumentLocation: "s3://bucket/v2",
			Active:           true,
		}
	})).Return(nil)
	// Mock store
	mockPBStore.On("ReadProtobuf", mock.Anything, mock.MatchedBy(func(reference storage.DataReference) bool {
		return reference.String() == s3Path
	}), mock.AnythingOfType("*admin.ConfigurationDocument")).Return(nil).Run(func(args mock.Arguments) {
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Version = "v1"
		configurations := make(map[string]*admin.Configuration)
		projectDomainKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{
			Org:     "org",
			Project: "project",
			Domain:  "domain",
		})
		assert.Nil(t, err)
		configurations[projectDomainKey] = &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "1",
					Gpu: "2",
				},
			},
		}
		projectKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{
			Org:     "org",
			Project: "project",
		})
		assert.Nil(t, err)
		configurations[projectKey] = &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "5",
					Gpu: "6",
				},
			},
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
				MaxParallelism: 1,
			},
		}
		globalKey, err := util.EncodeConfigurationDocumentKey(ctx, &util.GlobalConfigurationKey)
		assert.Nil(t, err)
		configurations[globalKey] = &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "7",
					Gpu: "8",
				},
			},
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
				Tags: []string{
					"foo", "bar", "baz",
				},
			},
		}
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Configurations = configurations
	})
	mockPBStore.On("GetBaseContainerFQN", mock.Anything).Return(storage.DataReference("s3://bucket"))
	mockRefConstructor.On("ConstructReference", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(storage.DataReference("s3://bucket/v2"), nil)
	mockMetadata := &storageMocks.Metadata{}
	mockMetadata.On("Exists").Return(false)
	mockPBStore.On("Head", mock.Anything, mock.Anything).Return(mockMetadata, nil)
	mockPBStore.On("WriteProtobuf", mock.Anything, mock.MatchedBy(func(reference storage.DataReference) bool {
		return reference.String() == "s3://bucket/v2"
	}), mock.Anything, mock.AnythingOfType("*admin.ConfigurationDocument")).Return(nil)
	// Mock plugin
	mockProjectConfigurationPlugin.SetGetAttributeIsMutableCallback(func(ctx context.Context, input *plugin.GetAttributeIsMutable) (map[admin.MatchableResource]*admin.AttributeIsMutable, error) {
		return map[admin.MatchableResource]*admin.AttributeIsMutable{
			admin.MatchableResource_TASK_RESOURCE: {
				Value: true,
			},
			admin.MatchableResource_CLUSTER_RESOURCE: {
				Value: true,
			},
			admin.MatchableResource_EXECUTION_QUEUE: {
				Value: true,
			},
			admin.MatchableResource_EXECUTION_CLUSTER_LABEL: {
				Value: true,
			},
			admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_PLUGIN_OVERRIDE: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_CLUSTER_ASSIGNMENT: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_EXTERNAL_RESOURCE: {
				Value:  false,
				Reason: "Cannot be modified",
			},
		}, nil
	})

	response, err := configurationManager.UpdateConfiguration(ctx, admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org:     "org",
			Project: "project",
			Domain:  "domain",
		},
		VersionToUpdate: "v1",
		Configuration: &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "3",
					Gpu: "4",
				},
			},
		},
	})
	assert.Nil(t, err)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).AssertExpectations(t)
	applicationConfig.AssertExpectations(t)
	mockRefConstructor.AssertExpectations(t)
	mockPBStore.AssertExpectations(t)

	fmt.Printf("Response: %v\n", response)
	assert.True(t, proto.Equal(response, &admin.ConfigurationUpdateResponse{
		Id:      &admin.ConfigurationID{Org: "org", Project: "project", Domain: "domain"},
		Version: "/2q3/pKW4h7lf1uExxhntIXaoI3dj0wL+JwEIrTIHUg=",
		Configuration: &admin.ConfigurationWithSource{
			TaskResourceAttributes: &admin.TaskResourceAttributesWithSource{
				Source: admin.AttributesSource_PROJECT_DOMAIN,
				Value: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu: "3",
						Gpu: "4",
					},
				},
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			ClusterResourceAttributes: &admin.ClusterResourceAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value: &admin.ExecutionQueueAttributes{
					Tags: []string{
						"foo", "bar", "baz",
					},
				},
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			ExecutionClusterLabel: &admin.ExecutionClusterLabelWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			QualityOfService: &admin.QualityOfServiceWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			PluginOverrides: &admin.PluginOverridesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfigWithSource{
				Source: admin.AttributesSource_PROJECT,
				Value: &admin.WorkflowExecutionConfig{
					MaxParallelism: 1,
				},
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			ClusterAssignment: &admin.ClusterAssignmentWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			ExternalResourceAttributes: &admin.ExternalResourceAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
		},
	}))
}

func TestUpdateProjectDomainConfiguration_UpdatingMutableAttributesError(t *testing.T) {
	ctx := context.Background()
	db := mocks.NewMockRepository()
	mockConfig := &runtimeMocks.Configuration{}
	mockPBStore := &storageMocks.ComposedProtobufStore{}
	mockRefConstructor := &storageMocks.ReferenceConstructor{}
	mockStorage := &storage.DataStore{
		ComposedProtobufStore: mockPBStore,
		ReferenceConstructor:  mockRefConstructor,
	}
	mockProjectConfigurationPlugin := utilMocks.MockProjectConfigurationPlugin{}
	pluginRegistry := plugins.NewRegistry()
	pluginRegistry.RegisterDefault(plugins.PluginIDProjectConfiguration, &mockProjectConfigurationPlugin)
	// Mock config
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	applicationConfig.On("GetDomainsConfig").Return(&runtimeInterfaces.DomainsConfig{
		runtimeInterfaces.Domain{
			ID:   "domain",
			Name: "domain",
		},
	})
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, pluginRegistry, ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	// Mock repo
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: s3Path,
		Active:           true,
	}, nil)
	// Mock store
	mockPBStore.On("ReadProtobuf", mock.Anything, mock.MatchedBy(func(reference storage.DataReference) bool {
		return reference.String() == s3Path
	}), mock.AnythingOfType("*admin.ConfigurationDocument")).Return(nil).Run(func(args mock.Arguments) {
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Version = "v1"
		configurations := make(map[string]*admin.Configuration)
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Configurations = configurations
	})
	// Mock plugin
	mockProjectConfigurationPlugin.SetGetAttributeIsMutableCallback(func(ctx context.Context, input *plugin.GetAttributeIsMutable) (map[admin.MatchableResource]*admin.AttributeIsMutable, error) {
		return map[admin.MatchableResource]*admin.AttributeIsMutable{
			admin.MatchableResource_TASK_RESOURCE: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_CLUSTER_RESOURCE: {
				Value: true,
			},
			admin.MatchableResource_EXECUTION_QUEUE: {
				Value: true,
			},
			admin.MatchableResource_EXECUTION_CLUSTER_LABEL: {
				Value: true,
			},
			admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_PLUGIN_OVERRIDE: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_CLUSTER_ASSIGNMENT: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_EXTERNAL_RESOURCE: {
				Value:  false,
				Reason: "Cannot be modified",
			},
		}, nil
	})

	_, err = configurationManager.UpdateConfiguration(ctx, admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org:     "org",
			Project: "project",
			Domain:  "domain",
		},
		VersionToUpdate: "v1",
		Configuration: &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "3",
					Gpu: "4",
				},
			},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "attributes not editable: [TASK_RESOURCE]")
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).AssertExpectations(t)
	applicationConfig.AssertExpectations(t)
	mockPBStore.AssertExpectations(t)
}

func TestUpdateOrgConfiguration(t *testing.T) {
	ctx := context.Background()
	db := mocks.NewMockRepository()
	mockConfig := &runtimeMocks.Configuration{}
	mockPBStore := &storageMocks.ComposedProtobufStore{}
	mockRefConstructor := &storageMocks.ReferenceConstructor{}
	mockStorage := &storage.DataStore{
		ComposedProtobufStore: mockPBStore,
		ReferenceConstructor:  mockRefConstructor,
	}
	mockProjectConfigurationPlugin := utilMocks.MockProjectConfigurationPlugin{}
	pluginRegistry := plugins.NewRegistry()
	pluginRegistry.RegisterDefault(plugins.PluginIDProjectConfiguration, &mockProjectConfigurationPlugin)
	// Mock config
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, pluginRegistry, ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	// Mock repo
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: s3Path,
		Active:           true,
	}, nil)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("Update", mock.Anything, mock.MatchedBy(func(input *interfaces.UpdateConfigurationInput) bool {
		if input.VersionToUpdate != "v1" {
			return false
		}
		return *input.NewConfigurationMetadata == models.ConfigurationDocumentMetadata{
			Version:          "xsqt/y06Uiiae023ZmNbFkYnQHNBtucHThzcj3i2aeE=",
			DocumentLocation: "s3://bucket/v2",
			Active:           true,
		}
	})).Return(nil)
	// Mock store
	mockPBStore.On("ReadProtobuf", mock.Anything, mock.MatchedBy(func(reference storage.DataReference) bool {
		return reference.String() == s3Path
	}), mock.AnythingOfType("*admin.ConfigurationDocument")).Return(nil).Run(func(args mock.Arguments) {
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Version = "v1"
		configurations := make(map[string]*admin.Configuration)
		projectKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{
			Org:     "org",
			Project: "project",
		})
		assert.Nil(t, err)
		configurations[projectKey] = &admin.Configuration{
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
				MaxParallelism: 1,
			},
		}
		orgKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{
			Org: "org",
		})
		assert.Nil(t, err)
		configurations[orgKey] = &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "1",
					Gpu: "2",
				},
			},
		}
		globalKey, err := util.EncodeConfigurationDocumentKey(ctx, &util.GlobalConfigurationKey)
		assert.Nil(t, err)
		configurations[globalKey] = &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "3",
					Gpu: "4",
				},
			},
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
				Tags: []string{
					"foo", "bar", "baz",
				},
			},
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
				MaxParallelism: 2,
			},
		}
		args.Get(2).(proto.Message).(*admin.ConfigurationDocument).Configurations = configurations
	})
	mockPBStore.On("GetBaseContainerFQN", mock.Anything).Return(storage.DataReference("s3://bucket"))
	mockRefConstructor.On("ConstructReference", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(storage.DataReference("s3://bucket/v2"), nil)
	mockMetadata := &storageMocks.Metadata{}
	mockMetadata.On("Exists").Return(false)
	mockPBStore.On("Head", mock.Anything, mock.Anything).Return(mockMetadata, nil)
	mockPBStore.On("WriteProtobuf", mock.Anything, mock.MatchedBy(func(reference storage.DataReference) bool {
		return reference.String() == "s3://bucket/v2"
	}), mock.Anything, mock.AnythingOfType("*admin.ConfigurationDocument")).Return(nil)
	// Mock plugin
	mockProjectConfigurationPlugin.SetGetAttributeIsMutableCallback(func(ctx context.Context, input *plugin.GetAttributeIsMutable) (map[admin.MatchableResource]*admin.AttributeIsMutable, error) {
		return map[admin.MatchableResource]*admin.AttributeIsMutable{
			admin.MatchableResource_TASK_RESOURCE: {
				Value: true,
			},
			admin.MatchableResource_CLUSTER_RESOURCE: {
				Value: true,
			},
			admin.MatchableResource_EXECUTION_QUEUE: {
				Value: true,
			},
			admin.MatchableResource_EXECUTION_CLUSTER_LABEL: {
				Value: true,
			},
			admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_PLUGIN_OVERRIDE: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_CLUSTER_ASSIGNMENT: {
				Value:  false,
				Reason: "Cannot be modified",
			},
			admin.MatchableResource_EXTERNAL_RESOURCE: {
				Value:  false,
				Reason: "Cannot be modified",
			},
		}, nil
	})

	response, err := configurationManager.UpdateConfiguration(ctx, admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Org: "org",
		},
		VersionToUpdate: "v1",
		Configuration: &admin.Configuration{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "5",
					Gpu: "6",
				},
			},
		},
	})
	assert.Nil(t, err)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).AssertExpectations(t)
	mockRefConstructor.AssertExpectations(t)
	mockPBStore.AssertExpectations(t)

	assert.True(t, proto.Equal(response, &admin.ConfigurationUpdateResponse{
		Id:      &admin.ConfigurationID{Org: "org"},
		Version: "xsqt/y06Uiiae023ZmNbFkYnQHNBtucHThzcj3i2aeE=",
		Configuration: &admin.ConfigurationWithSource{
			TaskResourceAttributes: &admin.TaskResourceAttributesWithSource{
				Source: admin.AttributesSource_ORG,
				Value: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu: "5",
						Gpu: "6",
					},
				},
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			ClusterResourceAttributes: &admin.ClusterResourceAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value: &admin.ExecutionQueueAttributes{
					Tags: []string{
						"foo", "bar", "baz",
					},
				},
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			ExecutionClusterLabel: &admin.ExecutionClusterLabelWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value: true,
					},
				},
			},
			QualityOfService: &admin.QualityOfServiceWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			PluginOverrides: &admin.PluginOverridesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfigWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value: &admin.WorkflowExecutionConfig{
					MaxParallelism: 2,
				},
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			ClusterAssignment: &admin.ClusterAssignmentWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
			ExternalResourceAttributes: &admin.ExternalResourceAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value:  nil,
				Metadata: &admin.AttributeMetadata{
					IsMutable: &admin.AttributeIsMutable{
						Value:  false,
						Reason: "Cannot be modified",
					},
				},
			},
		},
	}))
}

func TestNewAttributes(t *testing.T) {
	existingAttributes := sets.NewString(
		"TASK_RESOURCE",
		"CLUSTER_RESOURCE",
		"EXECUTION_QUEUE",
		"EXECUTION_CLUSTER_LABEL",
		"QUALITY_OF_SERVICE_SPECIFICATION",
		"PLUGIN_OVERRIDE",
		"WORKFLOW_EXECUTION_CONFIG",
		"CLUSTER_ASSIGNMENT",
		"EXTERNAL_RESOURCE",
	)
	for _, resource := range admin.MatchableResource_name {
		if !existingAttributes.Has(resource) {
			t.Fatalf("You are adding a new attributes [%s] to matchable_resource.proto, be sure to also modify [flyteidl/protos/flyteidl/admin/configuration.proto, flyteadmin/pkg/manager/impl/util/configuration.go, flyteadmin/pkg/repositories/transformers/configuration.go]", resource)
		}
	}

	existingAttributesInConfiguration := sets.NewString(
		"TaskResourceAttributes",
		"ClusterResourceAttributes",
		"ExecutionQueueAttributes",
		"ExecutionClusterLabel",
		"QualityOfService",
		"PluginOverrides",
		"WorkflowExecutionConfig",
		"ClusterAssignment",
		"ExternalResourceAttributes",
	)
	configType := reflect.ValueOf(admin.Configuration{}).Type()
	for i := 0; i < configType.NumField(); i++ {
		fieldName := configType.Field(i).Name
		if fieldName == "state" || fieldName == "sizeCache" || fieldName == "unknownFields" {
			continue
		}
		if !existingAttributesInConfiguration.Has(fieldName) {
			t.Fatalf("You are adding a new attributes [%s] to configuration.proto, be sure to also modify [flyteidl/protos/flyteidl/admin/matchable_resource.proto, flyteadmin/pkg/manager/impl/util/configuration.go, flyteadmin/pkg/repositories/transformers/configuration.go]", fieldName)
		}
	}
}
