package gormimpl

import (
	"context"
	"database/sql/driver"
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

const workflowID = uint(1)

var launchPlanSpec = []byte{1, 2}
var launchPlanClosure = []byte{3, 4}
var inactive = int32(admin.LaunchPlanState_INACTIVE)
var active = int32(admin.LaunchPlanState_ACTIVE)

func TestCreateLaunchPlan(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	err := launchPlanRepo.Create(context.Background(), models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: version,
		},
		Spec:       launchPlanSpec,
		WorkflowID: workflowID,
		Closure:    launchPlanClosure,
		State:      &inactive,
	})
	assert.NoError(t, err)
}

func getMockLaunchPlanResponseFromDb(expected models.LaunchPlan) map[string]interface{} {
	launchPlan := make(map[string]interface{})
	launchPlan["project"] = expected.Project
	launchPlan["domain"] = expected.Domain
	launchPlan["name"] = expected.Name
	launchPlan["version"] = expected.Version
	launchPlan["spec"] = expected.Spec
	launchPlan["workflow_id"] = expected.WorkflowID
	launchPlan["closure"] = expected.Closure
	launchPlan["state"] = expected.State
	return launchPlan
}

func TestGetLaunchPlan(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	launchPlans := make([]map[string]interface{}, 0)
	launchPlan := getMockLaunchPlanResponseFromDb(models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: version,
		},
		Spec:       launchPlanSpec,
		WorkflowID: workflowID,
		Closure:    launchPlanClosure,
		State:      &inactive,
	})
	launchPlans = append(launchPlans, launchPlan)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "launch_plans" WHERE "launch_plans"."project" = $1 AND "launch_plans"."domain" = $2 AND "launch_plans"."name" = $3 AND "launch_plans"."version" = $4 LIMIT 1`).WithReply(launchPlans)
	output, err := launchPlanRepo.Get(context.Background(), interfaces.Identifier{
		Project: project,
		Domain:  domain,
		Name:    name,
		Version: version,
	})
	assert.NoError(t, err)
	assert.Equal(t, project, output.Project)
	assert.Equal(t, domain, output.Domain)
	assert.Equal(t, name, output.Name)
	assert.Equal(t, version, output.Version)
	assert.Equal(t, launchPlanSpec, output.Spec)
}

func TestSetInactiveLaunchPlan(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	mockDb := GlobalMock.NewMock()
	updated := false
	mockDb.WithQuery(
		`UPDATE "launch_plans" SET "id"=$1,"updated_at"=$2,"project"=$3,"domain"=$4,"name"=$5,"version"=$6,"closure"=$7,"state"=$8 WHERE "project" = $9 AND "domain" = $10 AND "name" = $11 AND "version" = $12`).WithCallback(
		func(s string, values []driver.NamedValue) {
			updated = true
		},
	)

	err := launchPlanRepo.Update(context.Background(), models.LaunchPlan{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: version,
		},
		Closure: []byte{5, 6},
		State:   &inactive,
	})
	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestSetActiveLaunchPlan(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	mockQuery := GlobalMock.NewMock()
	updated := false
	mockQuery.WithQuery(
		`UPDATE "launch_plans" SET "id"=$1,"project"=$2,"domain"=$3,"name"=$4,"version"=$5,"closure"=$6,"state"=$7 WHERE "project" = $8 AND "domain" = $9 AND "name" = $10 AND "version" = $11`).WithCallback(
		func(s string, values []driver.NamedValue) {
			updated = true
		},
	)

	err := launchPlanRepo.SetActive(context.Background(), models.LaunchPlan{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "new version",
		},
		Closure: []byte{5, 6},
		State:   &active,
	}, &models.LaunchPlan{
		BaseModel: models.BaseModel{
			ID: 2,
		},
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "old version",
		},
		Closure: []byte{5, 6},
		State:   &inactive,
	})
	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestSetActiveLaunchPlan_NoCurrentlyActiveLaunchPlan(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	mockQuery := GlobalMock.NewMock()
	updated := false
	mockQuery.WithQuery(
		`UPDATE "launch_plans" SET "id"=$1,"project"=$2,"domain"=$3,"name"=$4,"version"=$5,"closure"=$6,"state"=$7 WHERE "project" = $8 AND "domain" = $9 AND "name" = $10 AND "version" = $11`).WithCallback(
		func(s string, values []driver.NamedValue) {
			updated = true
		},
	)
	err := launchPlanRepo.SetActive(context.Background(), models.LaunchPlan{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "new version",
		},
		Closure: []byte{5, 6},
		State:   &active,
	}, nil)
	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestListLaunchPlans(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	launchPlans := make([]map[string]interface{}, 0)
	versions := []string{"ABC", "XYZ"}
	for _, version := range versions {
		launchPlan := getMockLaunchPlanResponseFromDb(models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: project,
				Domain:  domain,
				Name:    name,
				Version: version,
			},
			Spec:       launchPlanSpec,
			WorkflowID: workflowID,
			Closure:    launchPlanClosure,
			State:      &inactive,
		})
		launchPlans = append(launchPlans, launchPlan)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithReply(launchPlans)

	collection, err := launchPlanRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.LaunchPlan, "project", project),
			getEqualityFilter(common.LaunchPlan, "domain", domain),
			getEqualityFilter(common.LaunchPlan, "name", name),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.LaunchPlans)
	assert.Len(t, collection.LaunchPlans, 2)
	for _, launchPlan := range collection.LaunchPlans {
		assert.Equal(t, project, launchPlan.Project)
		assert.Equal(t, domain, launchPlan.Domain)
		assert.Equal(t, name, launchPlan.Name)
		assert.Contains(t, versions, launchPlan.Version)
		assert.Equal(t, launchPlanSpec, launchPlan.Spec)
		assert.Equal(t, workflowID, launchPlan.WorkflowID)
	}
}

func TestListLaunchPlans_Pagination(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	launchPlans := make([]map[string]interface{}, 0)
	versions := []string{"ABC", "DEF"}
	for _, version := range versions {
		launchPlan := getMockLaunchPlanResponseFromDb(models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: project,
				Domain:  domain,
				Name:    name,
				Version: version,
			},
			Spec:       launchPlanSpec,
			WorkflowID: workflowID,
			Closure:    launchPlanClosure,
			State:      &inactive,
		})
		launchPlans = append(launchPlans, launchPlan)
	}

	GlobalMock := mocket.Catcher.Reset()

	GlobalMock.NewMock().WithQuery(
		`SELECT "launch_plans"."id","launch_plans"."created_at","launch_plans"."updated_at","launch_plans"."deleted_at","launch_plans"."project","launch_plans"."domain","launch_plans"."name","launch_plans"."version","launch_plans"."spec","launch_plans"."workflow_id","launch_plans"."closure","launch_plans"."state","launch_plans"."digest","launch_plans"."schedule_type" FROM "launch_plans" inner join workflows on launch_plans.workflow_id = workflows.id WHERE launch_plans.project = $1 AND launch_plans.domain = $2 AND launch_plans.name = $3 LIMIT 2 OFFSET 1`).WithReply(launchPlans)

	collection, err := launchPlanRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.LaunchPlan, "project", project),
			getEqualityFilter(common.LaunchPlan, "domain", domain),
			getEqualityFilter(common.LaunchPlan, "name", name),
		},
		Limit:  2,
		Offset: 1,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.LaunchPlans)
	assert.Len(t, collection.LaunchPlans, 2)
	for idx, launchPlan := range collection.LaunchPlans {
		assert.Equal(t, project, launchPlan.Project)
		assert.Equal(t, domain, launchPlan.Domain)
		assert.Equal(t, name, launchPlan.Name)
		assert.Equal(t, versions[idx], launchPlan.Version)
		assert.Equal(t, launchPlanSpec, launchPlan.Spec)
		assert.Equal(t, workflowID, launchPlan.WorkflowID)
		assert.Equal(t, launchPlanClosure, launchPlan.Closure)
	}
}

func TestListLaunchPlans_Filters(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	launchPlans := make([]map[string]interface{}, 0)
	launchPlan := getMockLaunchPlanResponseFromDb(models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "ABC",
		},
		Spec:       launchPlanSpec,
		WorkflowID: workflowID,
		Closure:    launchPlanClosure,
		State:      &inactive,
	})
	launchPlans = append(launchPlans, launchPlan)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	// Only match on queries that append the name filter
	GlobalMock.NewMock().WithQuery(`SELECT "launch_plans"."id","launch_plans"."created_at","launch_plans"."updated_at","launch_plans"."deleted_at","launch_plans"."project","launch_plans"."domain","launch_plans"."name","launch_plans"."version","launch_plans"."spec","launch_plans"."workflow_id","launch_plans"."closure","launch_plans"."state","launch_plans"."digest","launch_plans"."schedule_type" FROM "launch_plans" inner join workflows on launch_plans.workflow_id = workflows.id WHERE launch_plans.project = $1 AND launch_plans.domain = $2 AND launch_plans.name = $3 AND launch_plans.version = $4 LIMIT 20`).WithReply(launchPlans[0:1])

	collection, err := launchPlanRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.LaunchPlan, "project", project),
			getEqualityFilter(common.LaunchPlan, "domain", domain),
			getEqualityFilter(common.LaunchPlan, "name", name),
			getEqualityFilter(common.LaunchPlan, "version", "ABC"),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.LaunchPlans)
	assert.Len(t, collection.LaunchPlans, 1)
	assert.Equal(t, project, collection.LaunchPlans[0].Project)
	assert.Equal(t, domain, collection.LaunchPlans[0].Domain)
	assert.Equal(t, name, collection.LaunchPlans[0].Name)
	assert.Equal(t, "ABC", collection.LaunchPlans[0].Version)
	assert.Equal(t, launchPlanSpec, collection.LaunchPlans[0].Spec)
	assert.Equal(t, workflowID, collection.LaunchPlans[0].WorkflowID)
	assert.Equal(t, launchPlanClosure, collection.LaunchPlans[0].Closure)
}

func TestListLaunchPlans_Order(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	launchPlans := make([]map[string]interface{}, 0)

	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that include ordering by project
	mockQuery := GlobalMock.NewMock()
	mockQuery.WithQuery(`project desc`)
	mockQuery.WithReply(launchPlans)

	sortParameter, _ := common.NewSortParameter(&admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "project",
	}, models.LaunchPlanColumns)
	_, err := launchPlanRepo.List(context.Background(), interfaces.ListResourceInput{
		SortParameter: sortParameter,
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.LaunchPlan, "project", project),
			getEqualityFilter(common.LaunchPlan, "domain", domain),
			getEqualityFilter(common.LaunchPlan, "name", name),
			getEqualityFilter(common.LaunchPlan, "version", version),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.True(t, mockQuery.Triggered)
}

func TestListLaunchPlans_MissingParameters(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	_, err := launchPlanRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.LaunchPlan, "project", project),
			getEqualityFilter(common.LaunchPlan, "domain", domain),
			getEqualityFilter(common.LaunchPlan, "name", name),
			getEqualityFilter(common.LaunchPlan, "version", version),
		},
	})
	assert.EqualError(t, err, "missing and/or invalid parameters: limit")

	_, err = launchPlanRepo.List(context.Background(), interfaces.ListResourceInput{
		Limit: 20,
	})
	assert.EqualError(t, err, "missing and/or invalid parameters: filters")
}

func TestListLaunchPlansForWorkflow(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	launchPlans := make([]map[string]interface{}, 0)
	launchPlan := getMockLaunchPlanResponseFromDb(models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: version,
		},
		Spec:       launchPlanSpec,
		WorkflowID: workflowID,
		Closure:    launchPlanClosure,
		State:      &inactive,
	})
	launchPlans = append(launchPlans, launchPlan)

	GlobalMock := mocket.Catcher.Reset()
	// HACK: gorm orders the filters on join clauses non-deterministically. Ordering of filters doesn't affect
	// correctness, but because the mocket library only pattern matches on substrings, both variations of the (valid)
	// SQL that gorm produces are checked below.
	query := `SELECT "launch_plans"."id","launch_plans"."created_at","launch_plans"."updated_at","launch_plans"."deleted_at","launch_plans"."project","launch_plans"."domain","launch_plans"."name","launch_plans"."version","launch_plans"."spec","launch_plans"."workflow_id","launch_plans"."closure","launch_plans"."state","launch_plans"."digest","launch_plans"."schedule_type" FROM "launch_plans" inner join workflows on launch_plans.workflow_id = workflows.id WHERE launch_plans.project = $1 AND launch_plans.domain = $2 AND launch_plans.name = $3 AND workflows.deleted_at = $4 LIMIT 20`
	alternateQuery := `SELECT "launch_plans"."id","launch_plans"."created_at","launch_plans"."updated_at","launch_plans"."deleted_at","launch_plans"."project","launch_plans"."domain","launch_plans"."name","launch_plans"."version","launch_plans"."spec","launch_plans"."workflow_id","launch_plans"."closure","launch_plans"."state","launch_plans"."digest","launch_plans"."schedule_type" FROM "launch_plans" inner join workflows on launch_plans.workflow_id = workflows.id WHERE launch_plans.project = $1 AND launch_plans.domain = $2 AND launch_plans.name = $3 AND workflows.deleted_at = $4 LIMIT 20`
	GlobalMock.NewMock().WithQuery(query).WithReply(launchPlans)
	GlobalMock.NewMock().WithQuery(alternateQuery).WithReply(launchPlans)

	collection, err := launchPlanRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.LaunchPlan, "project", project),
			getEqualityFilter(common.LaunchPlan, "domain", domain),
			getEqualityFilter(common.LaunchPlan, "name", name),
			getEqualityFilter(common.Workflow, "deleted_at", "foo"),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.LaunchPlans)
	assert.Len(t, collection.LaunchPlans, 1)
	for _, launchPlan := range collection.LaunchPlans {
		assert.Equal(t, project, launchPlan.Project)
		assert.Equal(t, domain, launchPlan.Domain)
		assert.Equal(t, name, launchPlan.Name)
		assert.Contains(t, version, launchPlan.Version)
		assert.Equal(t, launchPlanSpec, launchPlan.Spec)
		assert.Equal(t, workflowID, launchPlan.WorkflowID)
		assert.Equal(t, launchPlanClosure, launchPlan.Closure)
	}
}

func TestListLaunchPlanIds(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	launchPlans := make([]map[string]interface{}, 0)
	versions := []string{"ABC", "XYZ"}
	for _, version := range versions {
		launchPlan := getMockLaunchPlanResponseFromDb(models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: project,
				Domain:  domain,
				Name:    name,
				Version: version,
			},
			Spec:       launchPlanSpec,
			WorkflowID: workflowID,
			Closure:    launchPlanClosure,
			State:      &inactive,
		})
		launchPlans = append(launchPlans, launchPlan)
		launchPlan = getMockLaunchPlanResponseFromDb(models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: project,
				Domain:  domain,
				Name:    "app.workflows.MyWorkflow",
				Version: version,
			},
			Spec:       launchPlanSpec,
			WorkflowID: uint(2),
			Closure:    launchPlanClosure,
			State:      &inactive,
		})
		launchPlans = append(launchPlans, launchPlan)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithReply(launchPlans)

	collection, err := launchPlanRepo.ListLaunchPlanIdentifiers(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.LaunchPlan, "project", project),
			getEqualityFilter(common.LaunchPlan, "domain", domain),
			getEqualityFilter(common.LaunchPlan, "name", name),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.LaunchPlans)
	assert.Len(t, collection.LaunchPlans, 4)
	for _, launchPlan := range collection.LaunchPlans {
		assert.Equal(t, project, launchPlan.Project)
		assert.Equal(t, domain, launchPlan.Domain)
		assert.True(t, launchPlan.Name == name || launchPlan.Name == "app.workflows.MyWorkflow")
		assert.Contains(t, versions, launchPlan.Version)
		assert.Equal(t, launchPlanSpec, launchPlan.Spec)
		assert.True(t, launchPlan.WorkflowID == workflowID || launchPlan.WorkflowID == uint(2))
	}
}
