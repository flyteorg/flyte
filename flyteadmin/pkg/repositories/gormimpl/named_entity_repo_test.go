package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
)

func getMockNamedEntityResponseFromDb(expected models.NamedEntity) map[string]interface{} {
	metadata := make(map[string]interface{})
	metadata["resource_type"] = expected.ResourceType
	metadata["project"] = expected.Project
	metadata["domain"] = expected.Domain
	metadata["name"] = expected.Name
	metadata["description"] = expected.Description
	metadata["state"] = expected.State
	metadata["org"] = expected.Org
	metadata["has_trigger"] = expected.HasTrigger
	return metadata
}

func TestGetNamedEntity(t *testing.T) {
	metadataRepo := NewNamedEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	results := make([]map[string]interface{}, 0)
	metadata := getMockNamedEntityResponseFromDb(models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      project,
			Domain:       domain,
			Name:         name,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: description,
		},
	})
	results = append(results, metadata)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(
		`SELECT workflows.project,workflows.domain,workflows.name,workflows.org,'2' AS resource_type,named_entity_metadata.description,named_entity_metadata.state,named_entity_metadata.has_trigger FROM "workflows" LEFT JOIN named_entity_metadata ON named_entity_metadata.resource_type = 2 AND named_entity_metadata.project = workflows.project AND named_entity_metadata.domain = workflows.domain AND named_entity_metadata.name = workflows.name AND named_entity_metadata.org = workflows.org WHERE workflows.project = $1 AND workflows.domain = $2 AND workflows.name = $3 AND workflows.org = $4 LIMIT 1`,
	).WithReply(results)
	output, err := metadataRepo.Get(context.Background(), interfaces.GetNamedEntityInput{
		ResourceType: resourceType,
		Project:      project,
		Domain:       domain,
		Name:         name,
		Org:          testOrg,
	})
	assert.NoError(t, err)
	assert.Equal(t, project, output.Project)
	assert.Equal(t, domain, output.Domain)
	assert.Equal(t, name, output.Name)
	assert.Equal(t, resourceType, output.ResourceType)
	assert.Equal(t, description, output.Description)
	assert.Nil(t, output.HasTrigger)
}

func TestUpdateNamedEntity_WithExisting(t *testing.T) {
	metadataRepo := NewNamedEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	const updatedDescription = "updated description"

	results := make([]map[string]interface{}, 0)
	activeState := int32(admin.NamedEntityState_NAMED_ENTITY_ACTIVE)
	metadata := getMockNamedEntityResponseFromDb(models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      project,
			Domain:       domain,
			Name:         name,
			Org:          testOrg,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: description,
			State:       &activeState,
		},
	})
	results = append(results, metadata)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(
		`SELECT "named_entity_metadata"."created_at","named_entity_metadata"."updated_at","named_entity_metadata"."deleted_at","named_entity_metadata"."resource_type","named_entity_metadata"."project","named_entity_metadata"."domain","named_entity_metadata"."name","named_entity_metadata"."org","named_entity_metadata"."description","named_entity_metadata"."state","named_entity_metadata"."has_trigger" FROM "named_entity_metadata" WHERE "named_entity_metadata"."resource_type" = $1 AND "named_entity_metadata"."project" = $2 AND "named_entity_metadata"."domain" = $3 AND "named_entity_metadata"."name" = $4 AND "named_entity_metadata"."org" = $5 ORDER BY "named_entity_metadata"."id" LIMIT 1`,
	).WithReply(results)

	mockQuery := GlobalMock.NewMock()
	mockQuery.WithQuery(
		`UPDATE "named_entity_metadata" SET "description"=$1,"state"=$2,"updated_at"=$3 WHERE "named_entity_metadata"."resource_type" = $4 AND "named_entity_metadata"."project" = $5 AND "named_entity_metadata"."domain" = $6 AND "named_entity_metadata"."name" = $7 AND "named_entity_metadata"."org" = $8 AND "resource_type" = $9 AND "project" = $10 AND "domain" = $11 AND "name" = $12 AND "org" = $13`,
	)

	err := metadataRepo.Update(context.Background(), models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      project,
			Domain:       domain,
			Name:         name,
			Org:          testOrg,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: updatedDescription,
			State:       &activeState,
		},
	})
	assert.NoError(t, err)
	assert.True(t, mockQuery.Triggered)
}

func TestUpdateNamedEntity_CreateNew(t *testing.T) {
	metadataRepo := NewNamedEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	const updatedDescription = "updated description"

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	mockQuery := GlobalMock.NewMock()
	mockQuery.WithQuery(
		`INSERT INTO "named_entity_metadata" ("created_at","updated_at","deleted_at","resource_type","project","domain","name","org","description","state","has_trigger") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
	)

	err := metadataRepo.Update(context.Background(), models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      project,
			Domain:       domain,
			Name:         name,
			Org:          testOrg,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: updatedDescription,
		},
	})
	assert.NoError(t, err)
	assert.True(t, mockQuery.Triggered)
}

func TestListNamedEntity(t *testing.T) {
	metadataRepo := NewNamedEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	results := make([]map[string]interface{}, 0)
	metadata := getMockNamedEntityResponseFromDb(models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      project,
			Domain:       domain,
			Name:         name,
			Org:          testOrg,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: description,
		},
	})
	results = append(results, metadata)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	mockQuery := GlobalMock.NewMock()

	mockQuery.WithQuery(
		`SELECT entities.project,entities.domain,entities.name,entities.org,'2' AS resource_type,named_entity_metadata.description,named_entity_metadata.state,named_entity_metadata.has_trigger FROM "named_entity_metadata" RIGHT JOIN (SELECT project,domain,name,org FROM "workflows" WHERE "domain" = $1 AND "org" = $2 AND "project" = $3 GROUP BY org, project, domain, name ORDER BY name desc LIMIT 20) AS entities ON named_entity_metadata.resource_type = 2 AND named_entity_metadata.project = entities.project AND named_entity_metadata.domain = entities.domain AND named_entity_metadata.name = entities.name AND named_entity_metadata.org = entities.org GROUP BY entities.project, entities.domain, entities.name, entities.org, named_entity_metadata.description, named_entity_metadata.state, named_entity_metadata.has_trigger ORDER BY name desc`,
	).WithReply(results)

	sortParameter, _ := common.NewSortParameter(&admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "name",
	}, models.NamedEntityColumns)
	output, err := metadataRepo.List(context.Background(), interfaces.ListNamedEntityInput{
		ResourceType: resourceType,
		Project:      "admintests",
		Domain:       "development",
		ListResourceInput: interfaces.ListResourceInput{
			Limit:         20,
			SortParameter: sortParameter,
		},
		Org: testOrg,
	})
	assert.NoError(t, err)
	assert.Len(t, output.Entities, 1)
	assert.Nil(t, output.Entities[0].HasTrigger)
}
