package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestCreateWorkflowAttributes(t *testing.T) {
	workflowRepo := NewWorkflowAttributesRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	query := GlobalMock.NewMock()
	query.WithQuery(
		`INSERT INTO "workflow_attributes" ("id","created_at","updated_at","deleted_at","project","domain",` +
			`"workflow","resource","attributes") VALUES (?,?,?,?,?,?,?,?,?)`)

	err := workflowRepo.CreateOrUpdate(context.Background(), models.WorkflowAttributes{
		Project:    "project",
		Domain:     "domain",
		Workflow:   "workflow",
		Resource:   "resource",
		Attributes: []byte("attrs"),
	})
	assert.NoError(t, err)
	assert.True(t, query.Triggered)
}

func TestGetWorkflowAttributes(t *testing.T) {
	workflowRepo := NewWorkflowAttributesRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	response := make(map[string]interface{})
	response["project"] = "project"
	response["domain"] = "domain"
	response["workflow"] = "workflow"
	response["resource"] = "resource"
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "workflow_attributes"  WHERE "workflow_attributes"."deleted_at" IS NULL AND` +
		` (("workflow_attributes"."project" = project) AND ("workflow_attributes"."domain" = domain) AND ` +
		`("workflow_attributes"."workflow" = workflow) AND ("workflow_attributes"."resource" = resource)) ORDER BY ` +
		`"workflow_attributes"."id" ASC LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := workflowRepo.Get(context.Background(), "project", "domain", "workflow", "resource")
	assert.Nil(t, err)
	assert.Equal(t, "project", output.Project)
	assert.Equal(t, "domain", output.Domain)
	assert.Equal(t, "workflow", output.Workflow)
	assert.Equal(t, "resource", output.Resource)
	assert.Equal(t, []byte("attrs"), output.Attributes)
}

func TestDeleteWorkflowAttributes(t *testing.T) {
	workflowRepo := NewWorkflowAttributesRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	fakeResponse := query.WithQuery(
		`DELETE FROM "workflow_attributes"  WHERE ("workflow_attributes"."project" = ?) AND ` +
			`("workflow_attributes"."domain" = ?) AND ("workflow_attributes"."workflow" = ?) AND ` +
			`("workflow_attributes"."resource" = ?)`)

	err := workflowRepo.Delete(context.Background(), "project", "domain", "workflow", "resource")
	assert.Nil(t, err)
	assert.True(t, fakeResponse.Triggered)
}
