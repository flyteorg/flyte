package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

var typedInterface = []byte{1, 2, 3}

const remoteSpecIdentifier = "remote spec id"

func TestCreateWorkflow(t *testing.T) {
	workflowRepo := NewWorkflowRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	err := workflowRepo.Create(context.Background(), models.Workflow{
		WorkflowKey: models.WorkflowKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: version,
		},
		TypedInterface:          typedInterface,
		RemoteClosureIdentifier: remoteSpecIdentifier,
	}, &models.DescriptionEntity{ShortDescription: "hello"})
	assert.NoError(t, err)
}

func getMockWorkflowResponseFromDb(version string, typedInterface []byte) map[string]interface{} {
	workflow := make(map[string]interface{})
	workflow["project"] = project
	workflow["domain"] = domain
	workflow["name"] = name
	workflow["version"] = version
	workflow["typed_interface"] = typedInterface
	workflow["remote_closure_identifier"] = remoteSpecIdentifier
	return workflow
}

func TestGetWorkflow(t *testing.T) {
	workflowRepo := NewWorkflowRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	workflows := make([]map[string]interface{}, 0)
	workflow := getMockWorkflowResponseFromDb(version, typedInterface)
	workflows = append(workflows, workflow)

	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "workflows" WHERE "workflows"."project" = $1 AND "workflows"."domain" = $2 AND "workflows"."name" = $3 AND "workflows"."version" = $4 LIMIT 1`).WithReply(workflows)
	output, err := workflowRepo.Get(context.Background(), interfaces.Identifier{
		Project: project,
		Domain:  domain,
		Name:    name,
		Version: version,
	})
	assert.Empty(t, err)
	assert.Equal(t, project, output.Project)
	assert.Equal(t, domain, output.Domain)
	assert.Equal(t, name, output.Name)
	assert.Equal(t, version, output.Version)
	assert.Equal(t, typedInterface, output.TypedInterface)
	assert.Equal(t, remoteSpecIdentifier, output.RemoteClosureIdentifier)
}

func TestListWorkflows(t *testing.T) {
	workflowRepo := NewWorkflowRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	workflows := make([]map[string]interface{}, 0)
	versions := []string{"ABC", "XYZ"}
	for _, version := range versions {
		workflow := getMockWorkflowResponseFromDb(version, typedInterface)
		workflows = append(workflows, workflow)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithReply(workflows)

	collection, err := workflowRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Workflow, "project", project),
			getEqualityFilter(common.Workflow, "domain", domain),
			getEqualityFilter(common.Workflow, "name", name),
		},
		Limit: 20,
	})
	assert.Empty(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Workflows)
	assert.Len(t, collection.Workflows, 2)
	for _, workflow := range collection.Workflows {
		assert.Equal(t, project, workflow.Project)
		assert.Equal(t, domain, workflow.Domain)
		assert.Equal(t, name, workflow.Name)
		assert.Contains(t, versions, workflow.Version)
		assert.Equal(t, typedInterface, workflow.TypedInterface)
		assert.Equal(t, remoteSpecIdentifier, workflow.RemoteClosureIdentifier)
	}
}

func TestListWorkflows_Pagination(t *testing.T) {
	workflowRepo := NewWorkflowRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	workflows := make([]map[string]interface{}, 0)
	versions := []string{"ABC", "DEF"}
	for _, version := range versions {
		workflow := getMockWorkflowResponseFromDb(version, typedInterface)
		workflows = append(workflows, workflow)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithReply(workflows)

	collection, err := workflowRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Workflow, "project", project),
			getEqualityFilter(common.Workflow, "domain", domain),
			getEqualityFilter(common.Workflow, "name", name),
		},
		Limit: 2,
	})
	assert.Empty(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Workflows)
	assert.Len(t, collection.Workflows, 2)
	for idx, workflow := range collection.Workflows {
		assert.Equal(t, project, workflow.Project)
		assert.Equal(t, domain, workflow.Domain)
		assert.Equal(t, name, workflow.Name)
		assert.Equal(t, versions[idx], workflow.Version)
		assert.Equal(t, typedInterface, workflow.TypedInterface)
		assert.Equal(t, remoteSpecIdentifier, workflow.RemoteClosureIdentifier)
	}
}

func TestListWorkflows_Filters(t *testing.T) {
	workflowRepo := NewWorkflowRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	workflows := make([]map[string]interface{}, 0)
	workflow := getMockWorkflowResponseFromDb("ABC", typedInterface)
	workflows = append(workflows, workflow)

	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that append the name filter
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "workflows" WHERE project = $1 AND domain = $2 AND name = $3 AND version = $4 LIMIT 20`).WithReply(workflows[0:1])

	collection, err := workflowRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Workflow, "project", project),
			getEqualityFilter(common.Workflow, "domain", domain),
			getEqualityFilter(common.Workflow, "name", name),
			getEqualityFilter(common.Workflow, "version", "ABC"),
		},
		Limit: 20,
	})
	assert.Empty(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Workflows)
	assert.Len(t, collection.Workflows, 1)
	assert.Equal(t, project, collection.Workflows[0].Project)
	assert.Equal(t, domain, collection.Workflows[0].Domain)
	assert.Equal(t, name, collection.Workflows[0].Name)
	assert.Equal(t, "ABC", collection.Workflows[0].Version)
	assert.Equal(t, typedInterface, collection.Workflows[0].TypedInterface)
	assert.Equal(t, remoteSpecIdentifier, collection.Workflows[0].RemoteClosureIdentifier)
}

func TestListWorkflows_Order(t *testing.T) {
	workflowRepo := NewWorkflowRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	workflows := make([]map[string]interface{}, 0)

	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that include ordering by project
	mockQuery := GlobalMock.NewMock()
	mockQuery.WithQuery(`project desc`)
	mockQuery.WithReply(workflows)

	sortParameter, _ := common.NewSortParameter(&admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "project",
	}, models.WorkflowColumns)
	_, err := workflowRepo.List(context.Background(), interfaces.ListResourceInput{
		SortParameter: sortParameter,
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Workflow, "project", project),
			getEqualityFilter(common.Workflow, "domain", domain),
			getEqualityFilter(common.Workflow, "name", name),
			getEqualityFilter(common.Workflow, "version", "ABC"),
		},
		Limit: 20,
	})
	assert.Empty(t, err)
	assert.True(t, mockQuery.Triggered)
}

func TestListWorkflows_MissingParameters(t *testing.T) {
	workflowRepo := NewWorkflowRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	_, err := workflowRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Workflow, "project", project),
			getEqualityFilter(common.Workflow, "domain", domain),
			getEqualityFilter(common.Workflow, "name", name),
			getEqualityFilter(common.Workflow, "version", version),
		},
	})
	assert.Equal(t, err.Error(), "missing and/or invalid parameters: limit")

	_, err = workflowRepo.List(context.Background(), interfaces.ListResourceInput{
		Limit: 20,
	})
	assert.Equal(t, err.Error(), "missing and/or invalid parameters: filters")
}

func TestListWorkflowIds(t *testing.T) {
	workflowRepo := NewWorkflowRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	workflows := make([]map[string]interface{}, 0)
	versions := []string{"ABC", "DEF"}
	// Instead of different versions, we should be returning different identifiers since the point of this list ids
	// function is to group away the versions, but for the purpose of this unit test, different versions will suffice.
	for _, version := range versions {
		workflow := getMockWorkflowResponseFromDb(version, typedInterface)
		workflows = append(workflows, workflow)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithReply(workflows)

	collection, err := workflowRepo.ListIdentifiers(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Workflow, "project", project),
			getEqualityFilter(common.Workflow, "domain", domain),
		},
		Limit: 2,
	})
	assert.Empty(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Workflows)
	assert.Len(t, collection.Workflows, 2)
	for idx, workflow := range collection.Workflows {
		assert.Equal(t, project, workflow.Project)
		assert.Equal(t, domain, workflow.Domain)
		assert.Equal(t, name, workflow.Name)
		assert.Equal(t, versions[idx], workflow.Version)
		assert.Equal(t, typedInterface, workflow.TypedInterface)
		assert.Equal(t, remoteSpecIdentifier, workflow.RemoteClosureIdentifier)
	}
}

func TestListWorkflowIds_MissingParameters(t *testing.T) {
	workflowRepo := NewWorkflowRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	_, err := workflowRepo.ListIdentifiers(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Workflow, "project", project),
			getEqualityFilter(common.Workflow, "domain", domain),
		},
	})

	assert.Equal(t, err.Error(), "missing and/or invalid parameters: limit")
}
