//go:build integration
// +build integration

package tests

import (
	"context"
	"sync"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

var taskResourceAttributes = &admin.TaskResourceAttributes{
	Defaults: &admin.TaskResourceSpec{
		Cpu:              "1",
		Gpu:              "2",
		Memory:           "1000Mi",
		EphemeralStorage: "x",
	},
	Limits: &admin.TaskResourceSpec{
		Cpu:              "3",
		Gpu:              "4",
		Memory:           "2000Mi",
		EphemeralStorage: "y",
	},
}

var executionQueueAttributes = &admin.ExecutionQueueAttributes{
	Tags: []string{
		"4", "5", "6",
	},
}

var workflowExecutionConfig = &admin.WorkflowExecutionConfig{
	MaxParallelism: 5,
}

func TestConfiguration(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	rollbackToDefaultConfiguration(db)
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	configurationId := &admin.ConfigurationID{
		Project: "admintests",
		Domain:  "development",
	}

	currentConfiguration, err := client.GetConfiguration(ctx, &admin.ConfigurationGetRequest{
		Id: configurationId,
	})
	assert.Nil(t, err)
	assert.NotNil(t, currentConfiguration)
	assert.True(t, proto.Equal(configurationId, currentConfiguration.Id))

	updatedConfiguration, err := client.UpdateConfiguration(ctx, &admin.ConfigurationUpdateRequest{
		Id:              configurationId,
		VersionToUpdate: currentConfiguration.Version,
		Configuration: &admin.Configuration{
			TaskResourceAttributes:   taskResourceAttributes,
			ExecutionQueueAttributes: executionQueueAttributes,
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, updatedConfiguration)
	assert.True(t, proto.Equal(configurationId, updatedConfiguration.Id))

	expectedConfiguration := currentConfiguration.Configuration
	expectedConfiguration.TaskResourceAttributes = &admin.TaskResourceAttributesWithSource{
		Source: admin.AttributesSource_PROJECT_DOMAIN,
		Value:  taskResourceAttributes,
		Metadata: &admin.AttributeMetadata{
			IsMutable: &admin.AttributeIsMutable{
				Value: true,
			},
		},
	}
	expectedConfiguration.ExecutionQueueAttributes = &admin.ExecutionQueueAttributesWithSource{
		Source: admin.AttributesSource_PROJECT_DOMAIN,
		Value:  executionQueueAttributes,
		Metadata: &admin.AttributeMetadata{
			IsMutable: &admin.AttributeIsMutable{
				Value: true,
			},
		},
	}
	assert.True(t, proto.Equal(expectedConfiguration, updatedConfiguration.Configuration))
}

func TestDefaultConfiguration(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	rollbackToDefaultConfiguration(db)
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	configurationId := &admin.ConfigurationID{
		Project: "admintests",
		Domain:  "development",
	}
	defaultConfiguration, err := client.GetConfiguration(ctx, &admin.ConfigurationGetRequest{
		Id:                             configurationId,
		OnlyGetLowerLevelConfiguration: true,
	})
	assert.Nil(t, err)
	assert.NotNil(t, defaultConfiguration)
	assert.True(t, proto.Equal(configurationId, defaultConfiguration.Id))

	currentConfiguration, err := client.GetConfiguration(ctx, &admin.ConfigurationGetRequest{
		Id: configurationId,
	})
	assert.Nil(t, err)
	assert.NotNil(t, currentConfiguration)
	assert.True(t, proto.Equal(configurationId, currentConfiguration.Id))

	assert.Equal(t, defaultConfiguration.Configuration, currentConfiguration.Configuration)

	updatedConfiguration, err := client.UpdateConfiguration(ctx, &admin.ConfigurationUpdateRequest{
		Id:              configurationId,
		VersionToUpdate: currentConfiguration.Version,
		Configuration: &admin.Configuration{
			TaskResourceAttributes: taskResourceAttributes,
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, updatedConfiguration)
	assert.True(t, proto.Equal(configurationId, updatedConfiguration.Id))
	defaultConfiguration.Configuration.TaskResourceAttributes = &admin.TaskResourceAttributesWithSource{
		Source: admin.AttributesSource_PROJECT_DOMAIN,
		Value:  taskResourceAttributes,
		Metadata: &admin.AttributeMetadata{
			IsMutable: &admin.AttributeIsMutable{
				Value: true,
			},
		},
	}
	assert.True(t, proto.Equal(defaultConfiguration.Configuration, updatedConfiguration.Configuration))
}

func TestConcurrentConfiguration(t *testing.T) {
	for round := 0; round < 10; round++ {
		ctx := context.Background()
		db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
		assert.Nil(t, err)
		rollbackToDefaultConfiguration(db)
		sqlDB, err := db.DB()
		assert.Nil(t, err)
		err = sqlDB.Close()
		assert.Nil(t, err)

		client, conn := GetTestAdminServiceClient()
		defer conn.Close()

		projects := []string{"project", "admintests", "flytekit"}
		domains := []string{"staging", "development", "production"}
		configurations := []*admin.Configuration{
			{
				TaskResourceAttributes: taskResourceAttributes,
			},
			{
				ExecutionQueueAttributes: executionQueueAttributes,
			},
		}

		var wg sync.WaitGroup
		numRoutines := 9 // number of goroutines to launch

		for i := 0; i < numRoutines; i++ {
			wg.Add(1)
			i := i
			go func() {
				defer wg.Done()

				configurationId := &admin.ConfigurationID{
					Project: projects[i/3],
					Domain:  domains[i%3],
				}
				defaultConfiguration, err := client.GetConfiguration(ctx, &admin.ConfigurationGetRequest{
					Id: configurationId,
				})
				assert.Nil(t, err)
				assert.NotNil(t, defaultConfiguration)
				assert.True(t, proto.Equal(configurationId, defaultConfiguration.Id))

				counter := 0
				for counter < 10 {
					currentConfiguration, err := client.GetConfiguration(ctx, &admin.ConfigurationGetRequest{
						Id: configurationId,
					})
					assert.Nil(t, err)
					assert.NotNil(t, currentConfiguration)
					assert.True(t, proto.Equal(configurationId, currentConfiguration.Id))

					updatedConfiguration, err := client.UpdateConfiguration(ctx, &admin.ConfigurationUpdateRequest{
						Id:              configurationId,
						VersionToUpdate: currentConfiguration.Version,
						Configuration:   configurations[counter%2],
					})
					if err != nil {
						assert.Equal(t, "rpc error: code = InvalidArgument desc = The document you are trying to update is outdated. Please try again.", err.Error())
						continue
					}
					assert.NotNil(t, updatedConfiguration)
					assert.True(t, proto.Equal(configurationId, updatedConfiguration.Id))

					expectedConfiguration := *(*defaultConfiguration).Configuration
					if counter%2 == 0 {
						expectedConfiguration.TaskResourceAttributes = &admin.TaskResourceAttributesWithSource{
							Source: admin.AttributesSource_PROJECT_DOMAIN,
							Value:  taskResourceAttributes,
							Metadata: &admin.AttributeMetadata{
								IsMutable: &admin.AttributeIsMutable{
									Value: true,
								},
							},
						}
					} else {
						expectedConfiguration.ExecutionQueueAttributes = &admin.ExecutionQueueAttributesWithSource{
							Source: admin.AttributesSource_PROJECT_DOMAIN,
							Value:  executionQueueAttributes,
							Metadata: &admin.AttributeMetadata{
								IsMutable: &admin.AttributeIsMutable{
									Value: true,
								},
							},
						}
					}
					assert.True(t, proto.Equal(&expectedConfiguration, updatedConfiguration.Configuration))
					counter++
				}
			}()
		}
		wg.Wait()
	}
}

func TestVersion(t *testing.T) {
	ctx := context.Background()
	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	rollbackToDefaultConfiguration(db)
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	projects := []string{"project", "admintests", "flytekit"}
	domains := []string{"staging", "development", "production"}

	defaultConfiguration, err := client.GetConfiguration(ctx, &admin.ConfigurationGetRequest{
		Id: &admin.ConfigurationID{
			Project: projects[0],
			Domain:  domains[0],
		},
	})
	defaultVersion := defaultConfiguration.Version
	currentVersion := defaultVersion

	for i := 0; i < 9; i++ {
		configurationId := &admin.ConfigurationID{
			Project: projects[i/3],
			Domain:  domains[i%3],
		}

		updatedConfiguration, err := client.UpdateConfiguration(ctx, &admin.ConfigurationUpdateRequest{
			Id:              configurationId,
			VersionToUpdate: currentVersion,
			Configuration: &admin.Configuration{
				TaskResourceAttributes:   taskResourceAttributes,
				ExecutionQueueAttributes: executionQueueAttributes,
			},
		})
		assert.Nil(t, err)
		assert.NotNil(t, updatedConfiguration)
		assert.True(t, proto.Equal(configurationId, updatedConfiguration.Id))
		currentVersion = updatedConfiguration.Version
	}

	for i := 0; i < 9; i++ {
		configurationId := &admin.ConfigurationID{
			Project: projects[i/3],
			Domain:  domains[i%3],
		}

		updatedConfiguration, err := client.UpdateConfiguration(ctx, &admin.ConfigurationUpdateRequest{
			Id:              configurationId,
			VersionToUpdate: currentVersion,
			Configuration:   &admin.Configuration{},
		})
		assert.Nil(t, err)
		assert.NotNil(t, updatedConfiguration)
		assert.True(t, proto.Equal(configurationId, updatedConfiguration.Id))
		currentVersion = updatedConfiguration.Version
	}
	assert.Equal(t, defaultVersion, currentVersion)
}

func TestUpdateConfiguration(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	rollbackToDefaultConfiguration(db)
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	currentConfiguration, err := client.GetConfiguration(ctx, &admin.ConfigurationGetRequest{
		Id: &admin.ConfigurationID{
			Project: "admintests",
			Domain:  "development",
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, currentConfiguration)
	assert.True(t, proto.Equal(&admin.ConfigurationID{
		Project: "admintests",
		Domain:  "development",
	}, currentConfiguration.Id))

	updatedConfiguration, err := client.UpdateConfiguration(ctx, &admin.ConfigurationUpdateRequest{
		Id:              &admin.ConfigurationID{},
		VersionToUpdate: currentConfiguration.Version,
		Configuration: &admin.Configuration{
			WorkflowExecutionConfig: workflowExecutionConfig,
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, updatedConfiguration)
	assert.True(t, proto.Equal(&admin.ConfigurationID{}, updatedConfiguration.Id))

	updatedConfiguration, err = client.UpdateConfiguration(ctx, &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: "admintests",
		},
		VersionToUpdate: updatedConfiguration.Version,
		Configuration: &admin.Configuration{
			ExecutionQueueAttributes: executionQueueAttributes,
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, updatedConfiguration)
	assert.True(t, proto.Equal(&admin.ConfigurationID{
		Project: "admintests",
	}, updatedConfiguration.Id))

	updatedConfiguration, err = client.UpdateConfiguration(ctx, &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: "admintests",
			Domain:  "development",
		},
		VersionToUpdate: updatedConfiguration.Version,
		Configuration: &admin.Configuration{
			TaskResourceAttributes: taskResourceAttributes,
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, updatedConfiguration)
	assert.True(t, proto.Equal(&admin.ConfigurationID{
		Project: "admintests",
		Domain:  "development",
	}, updatedConfiguration.Id))

	expectedConfiguration := currentConfiguration.Configuration
	expectedConfiguration.TaskResourceAttributes = &admin.TaskResourceAttributesWithSource{
		Source: admin.AttributesSource_PROJECT_DOMAIN,
		Value:  taskResourceAttributes,
		Metadata: &admin.AttributeMetadata{
			IsMutable: &admin.AttributeIsMutable{
				Value: true,
			},
		},
	}
	expectedConfiguration.ExecutionQueueAttributes = &admin.ExecutionQueueAttributesWithSource{
		Source: admin.AttributesSource_PROJECT,
		Value:  executionQueueAttributes,
		Metadata: &admin.AttributeMetadata{
			IsMutable: &admin.AttributeIsMutable{
				Value: true,
			},
		},
	}
	expectedConfiguration.WorkflowExecutionConfig = &admin.WorkflowExecutionConfigWithSource{
		Source: admin.AttributesSource_ORG,
		Value:  workflowExecutionConfig,
		Metadata: &admin.AttributeMetadata{
			IsMutable: &admin.AttributeIsMutable{
				Value: true,
			},
		},
	}
	assert.True(t, proto.Equal(expectedConfiguration, updatedConfiguration.Configuration))
}
