package gormimpl

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

const resourceTestWorkflowName = "workflow"

func TestCreateWorkflowAttributes(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	query.WithQuery(
		`INSERT INTO "resources" ("created_at","updated_at","deleted_at","project","domain",` +
			`"workflow","launch_plan","resource_type","priority","attributes") VALUES (?,?,?,?,?,?,?,?,?,?)`)

	err := resourceRepo.CreateOrUpdate(context.Background(), models.Resource{
		Project:      "project",
		Domain:       "domain",
		Workflow:     resourceTestWorkflowName,
		ResourceType: "resource",
		Priority:     models.ResourcePriorityLaunchPlanLevel,
		Attributes:   []byte("attrs"),
	})
	assert.NoError(t, err)
	assert.True(t, query.Triggered)
}

func TestGetWorkflowAttributes(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	response := make(map[string]interface{})
	response["project"] = "project"
	response["domain"] = "domain"
	response["workflow"] = resourceTestWorkflowName
	response["resource_type"] = "resource-type"
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "resources"  WHERE "resources"."deleted_at" IS NULL AND` +
		` ((resource_type = resource AND domain = domain AND project IN (,project)` +
		` AND workflow IN (,workflow) AND launch_plan IN ())) ORDER BY` +
		` priority desc,"resources"."id" ASC LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := resourceRepo.Get(context.Background(), interfaces.ResourceID{Project: "project", Domain: "domain", Workflow: "workflow", ResourceType: "resource"})
	assert.Nil(t, err)
	assert.Equal(t, "project", output.Project)
	assert.Equal(t, "domain", output.Domain)
	assert.Equal(t, "workflow", output.Workflow)
	assert.Equal(t, "resource-type", output.ResourceType)
	assert.Equal(t, []byte("attrs"), output.Attributes)
}

func TestProjectDomainAttributes(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	response := make(map[string]interface{})
	response[project] = project
	response[domain] = domain
	response["resource_type"] = "resource-type"
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "resources"  WHERE "resources"."deleted_at" IS NULL AND` +
		` ((resource_type = resource AND domain = domain AND project IN (,project)` +
		` AND workflow IN () AND launch_plan IN ())) ORDER BY` +
		` priority desc,"resources"."id" ASC LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := resourceRepo.Get(context.Background(), interfaces.ResourceID{Project: "project", Domain: "domain", ResourceType: "resource"})
	assert.Nil(t, err)
	assert.Equal(t, project, output.Project)
	assert.Equal(t, domain, output.Domain)
	assert.Equal(t, "", output.Workflow)
	assert.Equal(t, "resource-type", output.ResourceType)
	assert.Equal(t, []byte("attrs"), output.Attributes)
}

func TestGetRawWorkflowAttributes(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	response := make(map[string]interface{})
	response[project] = project
	response[domain] = domain
	response["workflow"] = resourceTestWorkflowName
	response["resource_type"] = "resource"
	response["launch_plan"] = "launch_plan"
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "resources"  WHERE "resources"."deleted_at" IS NULL AND (("resources"."project" = project) AND ` +
		`("resources"."domain" = domain) AND ("resources"."workflow" = workflow) AND ` +
		`("resources"."launch_plan" = launch_plan) AND ("resources"."resource_type" = resource)) ORDER BY "resources"."id" ASC LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := resourceRepo.GetRaw(context.Background(), interfaces.ResourceID{Project: "project", Domain: "domain", Workflow: "workflow", LaunchPlan: "launch_plan", ResourceType: "resource"})
	assert.Nil(t, err)
	assert.Equal(t, project, output.Project)
	assert.Equal(t, domain, output.Domain)
	assert.Equal(t, "workflow", output.Workflow)
	assert.Equal(t, "launch_plan", output.LaunchPlan)
	assert.Equal(t, "resource", output.ResourceType)
	assert.Equal(t, []byte("attrs"), output.Attributes)
}

func TestDeleteWorkflowAttributes(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	fakeResponse := query.WithQuery(
		`DELETE FROM "resources"  WHERE ("resources"."project" = ?) AND ` +
			`("resources"."domain" = ?) AND ("resources"."workflow" = ?) AND ` +
			`("resources"."launch_plan" = ?) AND ("resources"."resource_type" = ?)`)

	err := resourceRepo.Delete(context.Background(), interfaces.ResourceID{Project: "project", Domain: "domain", Workflow: "workflow", LaunchPlan: "launch_plan", ResourceType: "resource"})
	assert.Nil(t, err)
	assert.True(t, fakeResponse.Triggered)
}

func TestListAll(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	query := GlobalMock.NewMock()

	response := make(map[string]interface{})
	response[project] = project
	response[domain] = domain
	response["workflow"] = resourceTestWorkflowName
	response["resource_type"] = "resource"
	response["launch_plan"] = "launch_plan"
	response["attributes"] = []byte("attrs")

	fakeResponse := query.WithQuery(`SELECT * FROM "resources"  WHERE "resources"."deleted_at" IS NULL AND ` +
		`(("resources"."resource_type" = resource)) ORDER BY priority desc`).WithReply(
		[]map[string]interface{}{response})
	output, err := resourceRepo.ListAll(context.Background(), "resource")
	assert.Nil(t, err)
	assert.Len(t, output, 1)
	assert.Equal(t, project, output[0].Project)
	assert.Equal(t, domain, output[0].Domain)
	assert.Equal(t, "workflow", output[0].Workflow)
	assert.Equal(t, "launch_plan", output[0].LaunchPlan)
	assert.Equal(t, "resource", output[0].ResourceType)
	assert.Equal(t, []byte("attrs"), output[0].Attributes)
	assert.True(t, fakeResponse.Triggered)
}
