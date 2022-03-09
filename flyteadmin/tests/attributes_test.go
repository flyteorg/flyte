//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/repositories"

	"github.com/golang/protobuf/proto"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

var matchingAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_TaskResourceAttributes{
		TaskResourceAttributes: &admin.TaskResourceAttributes{
			Defaults: &admin.TaskResourceSpec{
				Cpu: "1",
			},
		},
	},
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
	truncateTableForTesting(db, "resources")
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
	truncateTableForTesting(db, "resources")
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	req := admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: matchingAttributes,
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
			MatchingAttributes: matchingAttributes,
		},
	}, response))

	var updatedMatchingAttributes = &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_TaskResourceAttributes{
			TaskResourceAttributes: &admin.TaskResourceAttributes{
				Defaults: &admin.TaskResourceSpec{
					Cpu: "1",
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
			MatchingAttributes: matchingAttributes,
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
	assert.EqualError(t, err, "rpc error: code = NotFound desc = Resource [{Project:admintests Domain:development Workflow: LaunchPlan: ResourceType:TASK_RESOURCE}] not found")
}

func TestUpdateWorkflowAttributes(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	truncateTableForTesting(db, "resources")
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	req := admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Project:            "admintests",
			Domain:             "development",
			Workflow:           "workflow",
			MatchingAttributes: matchingAttributes,
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
			MatchingAttributes: matchingAttributes,
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
	assert.EqualError(t, err, "rpc error: code = NotFound desc = Resource [{Project:admintests Domain:development Workflow:workflow LaunchPlan: ResourceType:TASK_RESOURCE}] not found")
}

func TestListAllMatchableAttributes(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	db, err := repositories.GetDB(ctx, getDbConfig(), getLoggerConfig())
	assert.Nil(t, err)
	truncateTableForTesting(db, "resources")
	sqlDB, err := db.DB()
	assert.Nil(t, err)
	err = sqlDB.Close()
	assert.Nil(t, err)

	req := admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: matchingAttributes,
		},
	}

	_, err = client.UpdateProjectDomainAttributes(ctx, &req)
	assert.Nil(t, err)

	response, err := client.ListMatchableAttributes(ctx, &admin.ListMatchableAttributesRequest{
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.Nil(t, err)
	assert.Len(t, response.Configurations, 1)
	assert.True(t, proto.Equal(&admin.MatchableAttributesConfiguration{
		Project:    "admintests",
		Domain:     "development",
		Attributes: matchingAttributes,
	}, response.Configurations[0]))

}
