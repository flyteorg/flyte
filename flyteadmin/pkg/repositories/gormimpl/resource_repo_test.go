package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

const resourceTestWorkflowName = "workflow"
const resourceTypeStr = "resource-type"

func TestCreateWorkflowAttributes(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	GlobalMock.Logging = true
	query.WithQuery(
		`INSERT INTO "resources" ("created_at","updated_at","deleted_at","project","domain","workflow","launch_plan","resource_type","priority","attributes") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING "id"`)

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

func getMockResourceResponseFromDb(expected models.Resource) map[string]interface{} {
	metadata := make(map[string]interface{})
	metadata["resource_type"] = expected.ResourceType
	metadata["project"] = expected.Project
	metadata["domain"] = expected.Domain
	metadata["priority"] = 2
	return metadata
}

func TestUpdateWorkflowAttributes_WithExisting(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	results := make([]map[string]interface{}, 0)
	metadata := getMockResourceResponseFromDb(models.Resource{
		ResourceType: resourceType.String(),
		Project:      project,
		Domain:       domain,
		Priority:     2,
	})
	results = append(results, metadata)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	mockSelectQuery := GlobalMock.NewMock()
	mockSelectQuery.WithQuery(
		`SELECT * FROM "resources" WHERE "resources"."project" = $1 AND "resources"."domain" = $2 AND "resources"."resource_type" = $3 AND "resources"."priority" = $4 ORDER BY "resources"."id" LIMIT 1`).WithReply(results)

	mockSaveQuery := GlobalMock.NewMock()
	mockSaveQuery.WithQuery(
		`INSERT INTO "resources" ("created_at","updated_at","deleted_at","project","domain","workflow","launch_plan","resource_type","priority","attributes") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`)

	err := resourceRepo.CreateOrUpdate(context.Background(), models.Resource{
		ResourceType: resourceType.String(),
		Project:      project,
		Domain:       domain,
		Priority:     2,
	})
	assert.NoError(t, err)
	assert.True(t, mockSelectQuery.Triggered)
	assert.True(t, mockSaveQuery.Triggered)
}

func TestGetWorkflowAttributes(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	response := make(map[string]interface{})
	response["project"] = "project"
	response["domain"] = "domain"
	response["workflow"] = resourceTestWorkflowName
	response["resource_type"] = resourceTypeStr
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "resources" WHERE resource_type = $1 AND domain IN ($2,$3) AND project IN ($4,$5) AND workflow IN ($6,$7) AND launch_plan IN ($8) ORDER BY priority desc,"resources"."id" LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := resourceRepo.Get(context.Background(), interfaces.ResourceID{Project: "project", Domain: "domain", Workflow: "workflow", ResourceType: "resource"})
	assert.Nil(t, err)
	assert.Equal(t, "project", output.Project)
	assert.Equal(t, "domain", output.Domain)
	assert.Equal(t, "workflow", output.Workflow)
	assert.Equal(t, resourceTypeStr, output.ResourceType)
	assert.Equal(t, []byte("attrs"), output.Attributes)
}

func TestProjectDomainAttributes(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	response := make(map[string]interface{})
	response[project] = project
	response[domain] = domain
	response["resource_type"] = resourceTypeStr
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "resources" WHERE resource_type = $1 AND domain IN ($2,$3) AND project IN ($4,$5) AND workflow IN ($6) AND launch_plan IN ($7) ORDER BY priority desc,"resources"."id" LIMIT 1`).WithReply(
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

func TestProjectLevelAttributes(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	response := make(map[string]interface{})
	response[project] = project
	response[domain] = ""
	response["resource_type"] = "resource-type"
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "resources" WHERE resource_type = $1 AND domain = '' AND project = $2 AND workflow = '' AND launch_plan = '' ORDER BY priority desc,"resources"."id" LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := resourceRepo.GetProjectLevel(context.Background(), interfaces.ResourceID{Project: "project", Domain: "", ResourceType: "resource"})
	assert.Nil(t, err)
	assert.Equal(t, project, output.Project)
	assert.Equal(t, "", output.Domain)
	assert.Equal(t, "", output.Workflow)
	assert.Equal(t, "resource-type", output.ResourceType)
	assert.Equal(t, []byte("attrs"), output.Attributes)

	// Must have a project defined
	_, err = resourceRepo.GetProjectLevel(context.Background(), interfaces.ResourceID{Project: "", Domain: "", ResourceType: "resource"})
	assert.Error(t, err)
}

func TestGetRawWorkflowAttributes(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	response := make(map[string]interface{})
	response[project] = project
	response[domain] = domain
	response["workflow"] = resourceTestWorkflowName
	response["resource_type"] = "resource"
	response["launch_plan"] = "launch_plan"
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "resources" WHERE "resources"."project" = $1 AND "resources"."domain" = $2 AND "resources"."workflow" = $3 AND "resources"."launch_plan" = $4 AND "resources"."resource_type" = $5 ORDER BY "resources"."id" LIMIT 1`).WithReply(
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
	GlobalMock.Logging = true
	query := GlobalMock.NewMock()
	fakeResponse := query.WithQuery(
		`DELETE FROM "resources" WHERE "resources"."project" = $1 AND "resources"."domain" = $2 AND "resources"."workflow" = $3 AND "resources"."launch_plan" = $4 AND "resources"."resource_type" = $5`)

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

	fakeResponse := query.WithQuery(`SELECT * FROM "resources" WHERE "resources"."resource_type" = $1 ORDER BY priority desc`).WithReply(
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

func TestGetError(t *testing.T) {
	resourceRepo := NewResourceRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "resources" WHERE resource_type = $1 AND domain IN ($2,$3) AND project IN ($4,$5) AND workflow IN ($6,$7) AND launch_plan IN ($8) ORDER BY priority desc,"resources"."id" LIMIT 1`).WithError(gorm.ErrRecordNotFound)

	output, err := resourceRepo.Get(context.Background(), interfaces.ResourceID{Project: "project", Domain: "domain", Workflow: "workflow", ResourceType: "resource"})
	assert.Error(t, err)
	assert.Equal(t, "", output.Project)
	assert.Equal(t, "", output.Domain)
	assert.Equal(t, "", output.Workflow)
}
