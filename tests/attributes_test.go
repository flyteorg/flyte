// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/gormimpl"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"

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

func TestUpdateProjectAttributes(t *testing.T) {
	ctx := context.Background()
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

	// If we ever expose get/list ProjectAttributes APIs update the test below to call those instead.
	db := databaseConfig.OpenDbConnection(databaseConfig.NewPostgresConfigProvider(geDbConfig(), adminScope))
	defer db.Close()

	errorsTransformer := errors.NewPostgresErrorTransformer(adminScope.NewSubScope("project_attrs_errors"))
	projectRepo := gormimpl.NewProjectAttributesRepo(db, errorsTransformer, adminScope.NewSubScope("project_attrs"))

	attributes, err := projectRepo.Get(ctx, "admintests", admin.MatchableResource_TASK_RESOURCE.String())
	assert.Nil(t, err)

	projectAttributes, err := transformers.FromProjectAttributesModel(attributes)
	assert.True(t, proto.Equal(matchingAttributes, projectAttributes.MatchingAttributes))
}

func TestUpdateProjectDomainAttributes(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	req := admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "admintests",
			Domain:             "development",
			MatchingAttributes: matchingAttributes,
		},
	}

	_, err := client.UpdateProjectDomainAttributes(ctx, &req)
	assert.Nil(t, err)

	// If we ever expose get/list ProjectDomainAttributes APIs update the test below to call those instead.
	db := databaseConfig.OpenDbConnection(databaseConfig.NewPostgresConfigProvider(getLocalDbConfig(), adminScope))
	defer db.Close()

	errorsTransformer := errors.NewPostgresErrorTransformer(adminScope.NewSubScope("project_domain_attrs_errors"))
	projectDomainRepo := gormimpl.NewProjectDomainAttributesRepo(db, errorsTransformer, adminScope.NewSubScope("project_domain_attrs"))

	attributes, err := projectDomainRepo.Get(ctx, "admintests", "development",
		admin.MatchableResource_TASK_RESOURCE.String())
	assert.Nil(t, err)

	projectDomainAttributes, err := transformers.FromProjectDomainAttributesModel(attributes)
	assert.True(t, proto.Equal(matchingAttributes, projectDomainAttributes.MatchingAttributes))
}

func TestUpdateWorkflowAttributes(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

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

	// If we ever expose get/list WorkflowAttributes APIs update the test below to call those instead.
	db := databaseConfig.OpenDbConnection(databaseConfig.NewPostgresConfigProvider(getLocalDbConfig(), adminScope))
	defer db.Close()

	errorsTransformer := errors.NewPostgresErrorTransformer(adminScope.NewSubScope("workflow_attrs_errors"))
	workflowRepo := gormimpl.NewWorkflowAttributesRepo(db, errorsTransformer, adminScope.NewSubScope("workflow_attrs"))

	attributes, err := workflowRepo.Get(ctx, "admintests", "development", "workflow",
		admin.MatchableResource_TASK_RESOURCE.String())
	assert.Nil(t, err)

	workflowAttributes, err := transformers.FromWorkflowAttributesModel(attributes)
	assert.True(t, proto.Equal(matchingAttributes, workflowAttributes.MatchingAttributes))
}
