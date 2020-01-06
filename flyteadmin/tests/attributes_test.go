// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/stretchr/testify/assert"

	databaseConfig "github.com/lyft/flyteadmin/pkg/repositories/config"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
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

func TestProjectAttributes(t *testing.T) {
	ctx := context.Background()

	db := databaseConfig.OpenDbConnection(databaseConfig.NewPostgresConfigProvider(getLocalDbConfig(), adminScope))
	truncateTableForTesting(db, "project_attributes")
	db.Close()

	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	req := admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{
			Project:            "admintests",
			MatchingAttributes: matchingAttributes,
		},
	}

	_, err := client.UpdateProjectAttributes(ctx, &req)
	assert.Nil(t, err)

	response, err := client.GetProjectAttributes(ctx, &admin.ProjectAttributesGetRequest{
		Project:      "admintests",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectAttributesGetResponse{
		Attributes: &admin.ProjectAttributes{
			Project:            "admintests",
			MatchingAttributes: matchingAttributes,
		},
	}, response))

	_, err = client.DeleteProjectAttributes(ctx, &admin.ProjectAttributesDeleteRequest{
		Project:      "admintests",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.Nil(t, err)

	_, err = client.GetProjectAttributes(ctx, &admin.ProjectAttributesGetRequest{
		Project:      "admintests",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.EqualError(t, err, "rpc error: code = NotFound desc = entry not found")
}

func TestUpdateProjectDomainAttributes(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	db := databaseConfig.OpenDbConnection(databaseConfig.NewPostgresConfigProvider(getLocalDbConfig(), adminScope))
	truncateTableForTesting(db, "project_domain_attributes")
	db.Close()

	req := admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: matchingAttributes,
		},
	}

	_, err := client.UpdateProjectDomainAttributes(ctx, &req)
	assert.Nil(t, err)

	response, err := client.GetProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: matchingAttributes,
		},
	}, response))

	_, err = client.DeleteProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesDeleteRequest{
		Project:      "admintests",
		Domain:       "development",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.Nil(t, err)

	_, err = client.GetProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.EqualError(t, err, "rpc error: code = NotFound desc = entry not found")
}

func TestUpdateWorkflowAttributes(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	db := databaseConfig.OpenDbConnection(databaseConfig.NewPostgresConfigProvider(getLocalDbConfig(), adminScope))
	truncateTableForTesting(db, "workflow_attributes")
	db.Close()

	req := admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Project:            "admintests",
			Domain:             "development",
			Workflow:           "workflow",
			MatchingAttributes: matchingAttributes,
		},
	}

	_, err := client.UpdateWorkflowAttributes(ctx, &req)
	assert.Nil(t, err)

	response, err := client.GetWorkflowAttributes(ctx, &admin.WorkflowAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
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
	assert.Nil(t, err)

	_, err = client.GetWorkflowAttributes(ctx, &admin.WorkflowAttributesGetRequest{
		Project:      "admintests",
		Domain:       "development",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})
	assert.EqualError(t, err, "rpc error: code = NotFound desc = entry not found")
}
