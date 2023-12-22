package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
)

const shortDescription = "hello"

func TestGetDescriptionEntity(t *testing.T) {
	descriptionEntityRepo := NewDescriptionEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	descriptionEntities := make([]map[string]interface{}, 0)
	descriptionEntity := getMockDescriptionEntityResponseFromDb(version, []byte{1, 2})
	descriptionEntities = append(descriptionEntities, descriptionEntity)

	output, err := descriptionEntityRepo.Get(context.Background(), &core.Identifier{
		ResourceType: resourceType,
		Project:      project,
		Domain:       domain,
		Name:         name,
		Version:      version,
	})
	assert.Empty(t, output)
	assert.EqualError(t, err, "Test transformer failed to find transformation to apply")

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "description_entities" WHERE project = $1 AND domain = $2 AND name = $3 AND version = $4 LIMIT 1`).
		WithReply(descriptionEntities)
	output, err = descriptionEntityRepo.Get(context.Background(), &core.Identifier{
		ResourceType: resourceType,
		Project:      project,
		Domain:       domain,
		Name:         name,
		Version:      version,
	})
	assert.Empty(t, err)
	assert.Equal(t, project, output.Project)
	assert.Equal(t, domain, output.Domain)
	assert.Equal(t, name, output.Name)
	assert.Equal(t, version, output.Version)
	assert.Equal(t, shortDescription, output.ShortDescription)
}

func TestListDescriptionEntities(t *testing.T) {
	descriptionEntityRepo := NewDescriptionEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	descriptionEntities := make([]map[string]interface{}, 0)
	versions := []string{"ABC", "XYZ"}
	for _, version := range versions {
		descriptionEntity := getMockDescriptionEntityResponseFromDb(version, []byte{1, 2})
		descriptionEntities = append(descriptionEntities, descriptionEntity)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithReply(descriptionEntities)

	identifierScope := &admin.NamedEntityIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}
	collection, err := descriptionEntityRepo.List(context.Background(), common.Workflow, interfaces.ListResourceInput{
		IdentifierScope: identifierScope,
	})
	assert.Equal(t, 0, len(collection.Entities))
	assert.Error(t, err)

	collection, err = descriptionEntityRepo.List(context.Background(), common.Workflow, interfaces.ListResourceInput{
		IdentifierScope: identifierScope,
		Limit:           20,
	})
	assert.Empty(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Entities)
	assert.Len(t, collection.Entities, 2)
	for _, descriptionEntity := range collection.Entities {
		assert.Equal(t, project, descriptionEntity.Project)
		assert.Equal(t, domain, descriptionEntity.Domain)
		assert.Equal(t, name, descriptionEntity.Name)
		assert.Contains(t, versions, descriptionEntity.Version)
		assert.Equal(t, shortDescription, descriptionEntity.ShortDescription)
	}
}

func getMockDescriptionEntityResponseFromDb(version string, digest []byte) map[string]interface{} {
	descriptionEntity := make(map[string]interface{})
	descriptionEntity["resource_type"] = resourceType
	descriptionEntity["project"] = project
	descriptionEntity["domain"] = domain
	descriptionEntity["name"] = name
	descriptionEntity["version"] = version
	descriptionEntity["Digest"] = digest
	descriptionEntity["ShortDescription"] = shortDescription
	return descriptionEntity
}
