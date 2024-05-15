package configurations

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/mocks"
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
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, ShouldNotBootstrapOrUpdateDefault)
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
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, ShouldNotBootstrapOrUpdateDefault)
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
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, ShouldNotBootstrapOrUpdateDefault)
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
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, ShouldNotBootstrapOrUpdateDefault)
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
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	// Mock repo
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: s3Path,
		Active:           true,
	}, nil)
	// Mock config
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	applicationConfig.On("GetDomainsConfig").Return(&runtimeInterfaces.DomainsConfig{
		runtimeInterfaces.Domain{
			ID:   "domain",
			Name: "domain",
		},
	})
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
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
		globalKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{})
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
			},
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfigWithSource{
				Source: admin.AttributesSource_PROJECT,
				Value: &admin.WorkflowExecutionConfig{
					MaxParallelism: 1,
				},
			},
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value: &admin.ExecutionQueueAttributes{
					Tags: []string{
						"foo", "bar", "baz",
					},
				},
			},
		},
	}))
}

func TestGetDefaultConfiguration(t *testing.T) {
	ctx := context.Background()
	db := mocks.NewMockRepository()
	mockConfig := &runtimeMocks.Configuration{}
	mockPBStore := &storageMocks.ComposedProtobufStore{}
	mockStorage := &storage.DataStore{
		ComposedProtobufStore: mockPBStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	// Mock repo
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: s3Path,
		Active:           true,
	}, nil)
	// Mock config
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	applicationConfig.On("GetDomainsConfig").Return(&runtimeInterfaces.DomainsConfig{
		runtimeInterfaces.Domain{
			ID:   "domain",
			Name: "domain",
		},
	})
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
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
		domainKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{
			Domain: "domain",
		})
		assert.Nil(t, err)
		configurations[domainKey] = &admin.Configuration{
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
				MaxParallelism: 2,
			},
		}
		globalKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{})
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

	response, err := configurationManager.GetConfiguration(ctx, admin.ConfigurationGetRequest{
		Id: &admin.ConfigurationID{
			Domain: "domain",
		},
	})
	assert.Nil(t, err)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).AssertExpectations(t)
	applicationConfig.AssertExpectations(t)
	mockPBStore.AssertExpectations(t)
	assert.True(t, proto.Equal(response, &admin.ConfigurationGetResponse{
		Id:      &admin.ConfigurationID{Domain: "domain"},
		Version: "v1",
		Configuration: &admin.ConfigurationWithSource{
			TaskResourceAttributes: &admin.TaskResourceAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu: "7",
						Gpu: "8",
					},
				},
			},
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfigWithSource{
				Source: admin.AttributesSource_DOMAIN,
				Value: &admin.WorkflowExecutionConfig{
					MaxParallelism: 2,
				},
			},
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value: &admin.ExecutionQueueAttributes{
					Tags: []string{
						"foo", "bar", "baz",
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
	configurationManager, err := NewConfigurationManager(ctx, db, mockConfig, mockStorage, ShouldNotBootstrapOrUpdateDefault)
	assert.Nil(t, err)
	// Mock repo
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("GetActive", mock.Anything).Return(models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: "s3://bucket/v1",
		Active:           true,
	}, nil)
	db.ConfigurationRepo().(*mocks.ConfigurationRepoInterface).On("Update", mock.Anything, mock.MatchedBy(func(input *interfaces.UpdateConfigurationInput) bool {
		if input.VersionToUpdate != "v1" {
			return false
		}
		return *input.NewConfigurationMetadata == models.ConfigurationDocumentMetadata{
			Version:          "x/Jg9dpEFSgcOydAjamhQRezcQrJHq8FictmbduZX6A=",
			DocumentLocation: "s3://bucket/v2",
			Active:           true,
		}
	})).Return(nil)
	// Mock config
	applicationConfig := &runtimeMocks.ApplicationConfiguration{}
	applicationConfig.On("GetDomainsConfig").Return(&runtimeInterfaces.DomainsConfig{
		runtimeInterfaces.Domain{
			ID:   "domain",
			Name: "domain",
		},
	})
	mockConfig.On("ApplicationConfiguration").Return(applicationConfig)
	// Mock store
	mockPBStore.On("ReadProtobuf", mock.Anything, mock.MatchedBy(func(reference storage.DataReference) bool {
		return reference.String() == "s3://bucket/v1"
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
		globalKey, err := util.EncodeConfigurationDocumentKey(ctx, &admin.ConfigurationID{})
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

	response, err := configurationManager.UpdateProjectDomainConfiguration(ctx, admin.ConfigurationUpdateRequest{
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
	assert.True(t, proto.Equal(response, &admin.ConfigurationUpdateResponse{
		Id:      &admin.ConfigurationID{Org: "org", Project: "project", Domain: "domain"},
		Version: "x/Jg9dpEFSgcOydAjamhQRezcQrJHq8FictmbduZX6A=",
		Configuration: &admin.ConfigurationWithSource{
			TaskResourceAttributes: &admin.TaskResourceAttributesWithSource{
				Source: admin.AttributesSource_PROJECT_DOMAIN,
				Value: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu: "3",
						Gpu: "4",
					},
				},
			},
			WorkflowExecutionConfig: &admin.WorkflowExecutionConfigWithSource{
				Source: admin.AttributesSource_PROJECT,
				Value: &admin.WorkflowExecutionConfig{
					MaxParallelism: 1,
				},
			},
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributesWithSource{
				Source: admin.AttributesSource_GLOBAL,
				Value: &admin.ExecutionQueueAttributes{
					Tags: []string{
						"foo", "bar", "baz",
					},
				},
			},
		},
	}))
}
