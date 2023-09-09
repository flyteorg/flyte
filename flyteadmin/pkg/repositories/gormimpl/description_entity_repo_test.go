package gormimpl

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

const shortDescription = "hello"

func TestGetDescriptionEntity(t *testing.T) {
	descriptionEntityRepo := NewDescriptionEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	descriptionEntities := make([]map[string]interface{}, 0)
	descriptionEntity := getMockDescriptionEntityResponseFromDb(version, []byte{1, 2})
	descriptionEntities = append(descriptionEntities, descriptionEntity)

	output, err := descriptionEntityRepo.Get(context.Background(), interfaces.GetDescriptionEntityInput{
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
	output, err = descriptionEntityRepo.Get(context.Background(), interfaces.GetDescriptionEntityInput{
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

	collection, err := descriptionEntityRepo.List(context.Background(), interfaces.ListResourceInput{})
	assert.Equal(t, 0, len(collection.Entities))
	assert.Error(t, err)

	collection, err = descriptionEntityRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Workflow, "project", project),
			getEqualityFilter(common.Workflow, "domain", domain),
			getEqualityFilter(common.Workflow, "name", name),
		},
		Limit: 20,
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

func TestGetDescriptionEntityFilters(t *testing.T) {
	filters, err := getDescriptionEntityFilters(resourceType, project, domain, name, version)
	entity := common.ResourceTypeToEntity[resourceType]
	assert.NoError(t, err)

	projectFilter, err := common.NewSingleValueFilter(entity, common.Equal, Project, project)
	assert.NoError(t, err)
	assert.Equal(t, filters[0], projectFilter)

	domainFilter, err := common.NewSingleValueFilter(entity, common.Equal, Domain, domain)
	assert.NoError(t, err)
	assert.Equal(t, filters[1], domainFilter)

	nameFilter, err := common.NewSingleValueFilter(entity, common.Equal, Name, name)
	assert.NoError(t, err)
	assert.Equal(t, filters[2], nameFilter)

	versionFilter, err := common.NewSingleValueFilter(entity, common.Equal, Version, version)
	assert.NoError(t, err)
	assert.Equal(t, filters[3], versionFilter)
}
