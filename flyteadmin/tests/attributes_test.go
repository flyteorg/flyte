//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	runtimeIfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

var matchingTaskResourceAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_TaskResourceAttributes{
		TaskResourceAttributes: &admin.TaskResourceAttributes{
			Defaults: &admin.TaskResourceSpec{
				Cpu: "1",
			},
			Limits: &admin.TaskResourceSpec{
				Cpu: "2",
			},
		},
	},
}

var matchingExecutionQueueAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
		ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
			Tags: []string{
				"4", "5", "6",
			},
		},
	},
}

func rollbackToDefaultConfiguration(db *gorm.DB) {
	ctx := context.Background()
	var defaultConfiguration models.ConfigurationDocumentMetadata
	err := db.Order("created_at asc").First(&defaultConfiguration).Error
	if err != nil {
		logger.Fatalf(ctx, "Error fetching default configuration %v", err)
	}
	err = db.Where("id != ?", defaultConfiguration.ID).Delete(&models.ConfigurationDocumentMetadata{}).Error
	if err != nil {
		logger.Fatalf(ctx, "Error deleting configurations %v", err)
	}
	err = db.Model(&models.ConfigurationDocumentMetadata{}).Where(&models.ConfigurationDocumentMetadata{
		Version: defaultConfiguration.Version,
	}).Update("active", true).Error
	if err != nil {
		logger.Fatalf(ctx, "Error updating default configuration %v", err)
	}
}

func TestUpdateClusterResourceAttributes(t *testing.T) {
	clusterMatchingAttributes := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ClusterResourceAttributes{
			ClusterResourceAttributes: &admin.ClusterResourceAttributes{
				Attributes: map[string]string{
					"key": "value",
				},
			},
		},
	}

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	configuration := runtime.NewConfigurationProvider()
	if configuration.ApplicationConfiguration().GetTopLevelConfig().ResourceAttributesMode == runtimeIfaces.ResourceAttributesModeConfiguration {
		rollbackToDefaultConfiguration(db)
	} else {
		truncateTableForTesting(db, "resources")
	}

	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	req := admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: clusterMatchingAttributes,
		},
	}

	_, err = client.UpdateProjectDomainAttributes(ctx, &req)
	fmt.Println(err)
	assert.Nil(t, err)

	listResp, err := client.ListMatchableAttributes(ctx, &admin.ListMatchableAttributesRequest{
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(listResp.GetConfigurations()))

	response, err := client.GetProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: clusterMatchingAttributes,
		},
	}, response))

	var updatedClusterMatchingAttributes = &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ClusterResourceAttributes{
			ClusterResourceAttributes: &admin.ClusterResourceAttributes{
				Attributes: map[string]string{
					"key": "value2",
				},
			},
		},
	}
	req = admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: updatedClusterMatchingAttributes,
		},
	}

	_, err = client.UpdateProjectDomainAttributes(ctx, &req)
	fmt.Println(err)
	assert.Nil(t, err)

	listResp, err = client.ListMatchableAttributes(ctx, &admin.ListMatchableAttributesRequest{
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(listResp.GetConfigurations()))

	response, err = client.GetProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: updatedClusterMatchingAttributes,
		},
	}, response))

}

func TestUpdateProjectDomainAttributes(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	configuration := runtime.NewConfigurationProvider()
	if configuration.ApplicationConfiguration().GetTopLevelConfig().ResourceAttributesMode == runtimeIfaces.ResourceAttributesModeConfiguration {
		rollbackToDefaultConfiguration(db)
	} else {
		truncateTableForTesting(db, "resources")
	}
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	req := admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: matchingTaskResourceAttributes,
		},
	}

	_, err = client.UpdateProjectDomainAttributes(ctx, &req)
	fmt.Println(err)
	assert.Nil(t, err)

	response, err := client.GetProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: matchingTaskResourceAttributes,
		},
	}, response))

	var updatedMatchingAttributes = &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_TaskResourceAttributes{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "1",
				},
				Limits: &admin.TaskResourceSpec{
					Cpu: "2",
				},
			},
		},
	}
	req = admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: updatedMatchingAttributes,
		},
	}

	_, err = client.UpdateProjectDomainAttributes(ctx, &req)
	fmt.Println(err)
	assert.Nil(t, err)

	response, err = client.GetProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: updatedMatchingAttributes,
		},
	}, response))

	workflowResponse, err := client.GetWorkflowAttributes(ctx, &admin.WorkflowAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, err)
	// Testing that if overrides are not set at workflow level, the one from Project-Domain is returned
	assert.True(t, proto.Equal(&admin.WorkflowAttributesGetResponse{
		Attributes: &admin.WorkflowAttributes{
			Project:            "admintests",
			Domain:             "development",
			Workflow:           "",
			MatchingAttributes: matchingTaskResourceAttributes,
		},
	}, workflowResponse))

	_, err = client.DeleteProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesDeleteRequest{
		Project:      "admintests",
		Domain:       "development",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, err)

	response, err = client.GetProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, response)
	assert.EqualError(t, err, "rpc error: code = NotFound desc = Resource [{Project:admintests Domain:development Workflow: LaunchPlan: ResourceType:TASK_RESOURCE Org:}] not found")
}

func TestUpdateWorkflowAttributes(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	configuration := runtime.NewConfigurationProvider()
	if configuration.ApplicationConfiguration().GetTopLevelConfig().ResourceAttributesMode == runtimeIfaces.ResourceAttributesModeConfiguration {
		rollbackToDefaultConfiguration(db)
	} else {
		truncateTableForTesting(db, "resources")
	}
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	req := admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Project:            "admintests",
			Domain:             "development",
			Workflow:           "workflow",
			MatchingAttributes: matchingTaskResourceAttributes,
		},
	}

	_, err = client.UpdateWorkflowAttributes(ctx, &req)
	fmt.Println(err)
	assert.Nil(t, err)

	response, err := client.GetWorkflowAttributes(ctx, &admin.WorkflowAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.WorkflowAttributesGetResponse{
		Attributes: &admin.WorkflowAttributes{
			Project:            "admintests",
			Domain:             "development",
			Workflow:           "workflow",
			MatchingAttributes: matchingTaskResourceAttributes,
		},
	}, response))

	_, err = client.DeleteWorkflowAttributes(ctx, &admin.WorkflowAttributesDeleteRequest{
		Project:      "admintests",
		Domain:       "development",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	fmt.Println(err)
	assert.Nil(t, err)

	_, err = client.GetWorkflowAttributes(ctx, &admin.WorkflowAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.EqualError(t, err, "rpc error: code = NotFound desc = Resource [{Project:admintests Domain:development Workflow:workflow LaunchPlan: ResourceType:TASK_RESOURCE Org:}] not found")
}

func TestListAllMatchableAttributes(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	configuration := runtime.NewConfigurationProvider()
	if configuration.ApplicationConfiguration().GetTopLevelConfig().ResourceAttributesMode == runtimeIfaces.ResourceAttributesModeConfiguration {
		rollbackToDefaultConfiguration(db)
	} else {
		truncateTableForTesting(db, "resources")
	}
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	_, err = client.UpdateProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: matchingTaskResourceAttributes,
		},
	})
	assert.Nil(t, err)

	_, err = client.UpdateWorkflowAttributes(ctx, &admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Project:            "admintests",
			Domain:             "development",
			Workflow:           "workflow",
			MatchingAttributes: matchingTaskResourceAttributes,
		},
	})
	assert.Nil(t, err)

	_, err = client.UpdateProjectAttributes(ctx, &admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{
			Project:            "admintests",
			MatchingAttributes: matchingTaskResourceAttributes,
		},
	})
	assert.Nil(t, err)

	_, err = client.UpdateProjectAttributes(ctx, &admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{
			Project:            "admintests",
			MatchingAttributes: matchingExecutionQueueAttributes,
		},
	})
	assert.Nil(t, err)

	response, err := client.ListMatchableAttributes(ctx, &admin.ListMatchableAttributesRequest{
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.Nil(t, err)
	assert.Len(t, response.Configurations, 3)
	assert.True(t, proto.Equal(&admin.MatchableAttributesConfiguration{
		Project:    "admintests",
		Domain:     "development",
		Workflow:   "workflow",
		Attributes: matchingTaskResourceAttributes,
	}, response.Configurations[0]))
	assert.True(t, proto.Equal(&admin.MatchableAttributesConfiguration{
		Project:    "admintests",
		Domain:     "development",
		Attributes: matchingTaskResourceAttributes,
	}, response.Configurations[1]))
	assert.True(t, proto.Equal(&admin.MatchableAttributesConfiguration{
		Project:    "admintests",
		Attributes: matchingTaskResourceAttributes,
	}, response.Configurations[2]))

}
