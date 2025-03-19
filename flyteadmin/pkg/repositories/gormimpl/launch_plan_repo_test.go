package gormimpl

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
)

const workflowID = uint(1)

var launchPlanSpec = []byte{1, 2}
var launchPlanClosure = []byte{3, 4}
var inactive = int32(admin.LaunchPlanState_INACTIVE)
var active = int32(admin.LaunchPlanState_ACTIVE)

type MockNamedEntityUpdater struct {
	Updated             bool
	InputtedNamedEntity models.NamedEntity
	ReturnError         bool
}

func (m *MockNamedEntityUpdater) Update(ctx context.Context, tx *gorm.DB, namedEntity models.NamedEntity, errTransformer adminErrors.ErrorTransformer) error {
	m.Updated = true
	m.InputtedNamedEntity = namedEntity
	if m.ReturnError {
		return errors.New("injected error")
	}
	return nil
}

func TestCreateLaunchPlan(t *testing.T) {

	trueVal := true
	falseVal := false
	tests := []struct {
		name                    string
		launchPlanCondition     models.LaunchConditionType
		updateNamedEntityError  bool
		expectNamedEntityUpdate bool
		hasTrigger              *bool
	}{
		{
			name:                    "create with launch condition type ARTIFACT",
			launchPlanCondition:     models.LaunchConditionTypeARTIFACT,
			expectNamedEntityUpdate: true,
			hasTrigger:              &trueVal,
		},
		{
			name:                    "create with launch condition type SCHEDULE",
			launchPlanCondition:     models.LaunchConditionTypeSCHED,
			expectNamedEntityUpdate: true,
			hasTrigger:              &trueVal,
		},
		{
			name:                    "create without launch condition",
			expectNamedEntityUpdate: true,
			hasTrigger:              &falseVal,
		},
		{
			name:                   "error with updating named entity",
			launchPlanCondition:    models.LaunchConditionTypeSCHED,
			updateNamedEntityError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockUpdater := &MockNamedEntityUpdater{ReturnError: test.updateNamedEntityError}
			launchPlanRepo := &LaunchPlanRepo{
				db:                 GetDbForTest(t),
				errorTransformer:   adminErrors.NewTestErrorTransformer(),
				metrics:            newMetrics(mockScope.NewTestScope()),
				launchPlanMetrics:  launchPlanMetrics{},
				namedEntityUpdater: mockUpdater,
			}

			err := launchPlanRepo.Create(context.Background(), models.LaunchPlan{
				LaunchPlanKey: models.LaunchPlanKey{
					Project: project,
					Domain:  domain,
					Name:    name,
					Version: version,
					Org:     Org,
				},
				Spec:                launchPlanSpec,
				WorkflowID:          workflowID,
				Closure:             launchPlanClosure,
				State:               &inactive,
				LaunchConditionType: &test.launchPlanCondition, // #nosec G601
			})
			if test.updateNamedEntityError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if test.expectNamedEntityUpdate {
				assert.True(t, mockUpdater.Updated)
				assert.Equal(t, test.hasTrigger, mockUpdater.InputtedNamedEntity.HasTrigger)
				assert.Equal(t, core.ResourceType_LAUNCH_PLAN, mockUpdater.InputtedNamedEntity.ResourceType)
				assert.Equal(t, project, mockUpdater.InputtedNamedEntity.Project)
				assert.Equal(t, domain, mockUpdater.InputtedNamedEntity.Domain)
				assert.Equal(t, name, mockUpdater.InputtedNamedEntity.Name)
				assert.Equal(t, Org, mockUpdater.InputtedNamedEntity.Org)
			} else {
				assert.False(t, mockUpdater.Updated)
			}
		})
	}
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
	launchPlan["org"] = expected.Org
	return launchPlan
}

func TestGetLaunchPlan(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())

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
		`SELECT * FROM "launch_plans" WHERE "launch_plans"."project" = $1 AND "launch_plans"."domain" = $2 AND "launch_plans"."name" = $3 AND "launch_plans"."version" = $4 AND "org" = $5 LIMIT 1`).WithReply(launchPlans)
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
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	mockDb := GlobalMock.NewMock()
	updated := false
	mockDb.WithQuery(
		`UPDATE "launch_plans" SET "id"=$1,"updated_at"=$2,"project"=$3,"domain"=$4,"name"=$5,"version"=$6,"closure"=$7,"state"=$8 WHERE "org" = $9 AND "project" = $10 AND "domain" = $11 AND "name" = $12 AND "version" = $13`).WithCallback(
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
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	mockQuery := GlobalMock.NewMock()
	updated := false
	mockQuery.WithQuery(
		`UPDATE "launch_plans" SET "id"=$1,"project"=$2,"domain"=$3,"name"=$4,"version"=$5,"closure"=$6,"state"=$7 WHERE "org" = $8 AND "project" = $9 AND "domain" = $10 AND "name" = $11 AND "version" = $12`).WithCallback(
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
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	mockQuery := GlobalMock.NewMock()
	updated := false
	mockQuery.WithQuery(
		`UPDATE "launch_plans" SET "id"=$1,"project"=$2,"domain"=$3,"name"=$4,"version"=$5,"closure"=$6,"state"=$7 WHERE "org" = $8 AND "project" = $9 AND "domain" = $10 AND "name" = $11 AND "version" = $12`).WithCallback(
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
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())

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
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())

	launchPlans := make([]map[string]interface{}, 0)
	versions := []string{"ABC", "DEF"}
	for _, version := range versions {
		launchPlan := getMockLaunchPlanResponseFromDb(models.LaunchPlan{
			LaunchPlanKey: models.LaunchPlanKey{
				Project: project,
				Domain:  domain,
				Name:    name,
				Version: version,
				Org:     testOrg,
			},
			Spec:       launchPlanSpec,
			WorkflowID: workflowID,
			Closure:    launchPlanClosure,
			State:      &inactive,
		})
		launchPlans = append(launchPlans, launchPlan)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(
		`SELECT "launch_plans"."id","launch_plans"."created_at","launch_plans"."updated_at","launch_plans"."deleted_at","launch_plans"."project","launch_plans"."domain","launch_plans"."name","launch_plans"."version","launch_plans"."org","launch_plans"."spec","launch_plans"."workflow_id","launch_plans"."closure","launch_plans"."state","launch_plans"."digest","launch_plans"."schedule_type","launch_plans"."launch_condition_type" FROM "launch_plans" inner join workflows on launch_plans.workflow_id = workflows.id WHERE launch_plans.project = $1 AND launch_plans.domain = $2 AND launch_plans.name = $3 AND launch_plans.org = $4 LIMIT 2 OFFSET 1`).WithReply(launchPlans)

	collection, err := launchPlanRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.LaunchPlan, "project", project),
			getEqualityFilter(common.LaunchPlan, "domain", domain),
			getEqualityFilter(common.LaunchPlan, "name", name),
			getEqualityFilter(common.LaunchPlan, "org", testOrg),
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
		assert.Equal(t, testOrg, launchPlan.Org)
		assert.Equal(t, versions[idx], launchPlan.Version)
		assert.Equal(t, launchPlanSpec, launchPlan.Spec)
		assert.Equal(t, workflowID, launchPlan.WorkflowID)
		assert.Equal(t, launchPlanClosure, launchPlan.Closure)
	}
}

func TestListLaunchPlans_Filters(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())

	launchPlans := make([]map[string]interface{}, 0)
	launchPlan := getMockLaunchPlanResponseFromDb(models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: "ABC",
			Org:     testOrg,
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
	GlobalMock.NewMock().WithQuery(`SELECT "launch_plans"."id","launch_plans"."created_at","launch_plans"."updated_at","launch_plans"."deleted_at","launch_plans"."project","launch_plans"."domain","launch_plans"."name","launch_plans"."version","launch_plans"."org","launch_plans"."spec","launch_plans"."workflow_id","launch_plans"."closure","launch_plans"."state","launch_plans"."digest","launch_plans"."schedule_type","launch_plans"."launch_condition_type" FROM "launch_plans" inner join workflows on launch_plans.workflow_id = workflows.id WHERE launch_plans.project = $1 AND launch_plans.domain = $2 AND launch_plans.name = $3 AND launch_plans.version = $4 AND launch_plans.org = $5 LIMIT 20`).WithReply(launchPlans[0:1])

	collection, err := launchPlanRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.LaunchPlan, "project", project),
			getEqualityFilter(common.LaunchPlan, "domain", domain),
			getEqualityFilter(common.LaunchPlan, "name", name),
			getEqualityFilter(common.LaunchPlan, "version", "ABC"),
			getEqualityFilter(common.LaunchPlan, "org", testOrg),
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
	assert.Equal(t, testOrg, collection.LaunchPlans[0].Org)
	assert.Equal(t, "ABC", collection.LaunchPlans[0].Version)
	assert.Equal(t, launchPlanSpec, collection.LaunchPlans[0].Spec)
	assert.Equal(t, workflowID, collection.LaunchPlans[0].WorkflowID)
	assert.Equal(t, launchPlanClosure, collection.LaunchPlans[0].Closure)
}

func TestListLaunchPlans_Order(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())
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
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())
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
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())

	launchPlans := make([]map[string]interface{}, 0)
	launchPlan := getMockLaunchPlanResponseFromDb(models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: version,
			Org:     testOrg,
		},
		Spec:       launchPlanSpec,
		WorkflowID: workflowID,
		Closure:    launchPlanClosure,
		State:      &inactive,
	})
	launchPlans = append(launchPlans, launchPlan)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	// HACK: gorm orders the filters on join clauses non-deterministically. Ordering of filters doesn't affect
	// correctness, but because the mocket library only pattern matches on substrings, both variations of the (valid)
	// SQL that gorm produces are checked below.
	query := `SELECT "launch_plans"."id","launch_plans"."created_at","launch_plans"."updated_at","launch_plans"."deleted_at","launch_plans"."project","launch_plans"."domain","launch_plans"."name","launch_plans"."version","launch_plans"."org","launch_plans"."spec","launch_plans"."workflow_id","launch_plans"."closure","launch_plans"."state","launch_plans"."digest","launch_plans"."schedule_type","launch_plans"."launch_condition_type" FROM "launch_plans" inner join workflows on launch_plans.workflow_id = workflows.id WHERE launch_plans.project = $1 AND launch_plans.domain = $2 AND launch_plans.name = $3 AND launch_plans.org = $4 AND workflows.deleted_at = $5 LIMIT 20`
	alternateQuery := `SELECT "launch_plans"."id","launch_plans"."created_at","launch_plans"."updated_at","launch_plans"."deleted_at","launch_plans"."project","launch_plans"."domain","launch_plans"."name","launch_plans"."version","launch_plans"."org","launch_plans"."spec","launch_plans"."workflow_id","launch_plans"."closure","launch_plans"."state","launch_plans"."digest","launch_plans"."schedule_type","launch_plans"."launch_condition_type" FROM "launch_plans" inner join workflows on launch_plans.workflow_id = workflows.id WHERE launch_plans.project = $1 AND launch_plans.domain = $2 AND launch_plans.name = $3 AND launch_plans.org = $4 AND workflows.deleted_at = $5 LIMIT 20`
	GlobalMock.NewMock().WithQuery(query).WithReply(launchPlans)
	GlobalMock.NewMock().WithQuery(alternateQuery).WithReply(launchPlans)

	collection, err := launchPlanRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.LaunchPlan, "project", project),
			getEqualityFilter(common.LaunchPlan, "domain", domain),
			getEqualityFilter(common.LaunchPlan, "name", name),
			getEqualityFilter(common.LaunchPlan, "org", testOrg),
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
		assert.Equal(t, testOrg, launchPlan.Org)
		assert.Contains(t, version, launchPlan.Version)
		assert.Equal(t, launchPlanSpec, launchPlan.Spec)
		assert.Equal(t, workflowID, launchPlan.WorkflowID)
		assert.Equal(t, launchPlanClosure, launchPlan.Closure)
	}
}

func TestListLaunchPlanIds(t *testing.T) {
	launchPlanRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())

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

func TestCountLaunchPlans(t *testing.T) {
	executionRepo := NewLaunchPlanRepo(GetDbForTest(t), adminErrors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithQuery(
		`SELECT count(*) FROM "launch_plans"`).WithReply([]map[string]interface{}{{"rows": 2}})

	count, err := executionRepo.Count(context.Background(), interfaces.CountResourceInput{})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
