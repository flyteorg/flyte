package impl

import (
	"context"
	"testing"

	"github.com/lyft/flyteadmin/pkg/common"

	"github.com/lyft/flyteadmin/pkg/manager/impl/testutils"
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/lyft/flyteadmin/pkg/repositories/mocks"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/lyft/flyteadmin/pkg/runtime/mocks"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

var namedEntityIdentifier = admin.NamedEntityIdentifier{
	Project: project,
	Domain:  domain,
	Name:    name,
}

var badIdentifier = admin.NamedEntityIdentifier{
	Project: project,
	Domain:  domain,
	Name:    "",
}

func getMockRepositoryForNETest() repositories.RepositoryInterface {
	return repositoryMocks.NewMockRepository()
}

func getMockConfigForNETest() runtimeInterfaces.Configuration {
	mockConfig := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultDomains(), nil, nil, nil, nil, nil)
	return mockConfig
}

func TestNamedEntityManager_Get(t *testing.T) {
	repository := getMockRepositoryForNETest()
	manager := NewNamedEntityManager(repository, getMockConfigForNETest(), mockScope.NewTestScope())

	getFunction := func(input interfaces.GetNamedEntityInput) (models.NamedEntity, error) {
		return models.NamedEntity{
			NamedEntityKey: models.NamedEntityKey{
				ResourceType: input.ResourceType,
				Project:      input.Project,
				Domain:       input.Domain,
				Name:         input.Name,
			},
			NamedEntityMetadataFields: models.NamedEntityMetadataFields{
				Description: description,
			},
		}, nil
	}
	repository.NamedEntityRepo().(*repositoryMocks.MockNamedEntityRepo).SetGetCallback(getFunction)
	response, err := manager.GetNamedEntity(context.Background(), admin.NamedEntityGetRequest{
		ResourceType: resourceType,
		Id:           &namedEntityIdentifier,
	})
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestNamedEntityManager_Get_BadRequest(t *testing.T) {
	repository := getMockRepositoryForNETest()
	manager := NewNamedEntityManager(repository, getMockConfigForNETest(), mockScope.NewTestScope())

	response, err := manager.GetNamedEntity(context.Background(), admin.NamedEntityGetRequest{
		ResourceType: core.ResourceType_UNSPECIFIED,
		Id:           &namedEntityIdentifier,
	})
	assert.Error(t, err)
	assert.Nil(t, response)

	response, err = manager.GetNamedEntity(context.Background(), admin.NamedEntityGetRequest{
		ResourceType: resourceType,
		Id:           &badIdentifier,
	})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestNamedEntityManager_UpdateQueryFilters(t *testing.T) {
	identityFilter, err := common.NewSingleValueFilter(common.NamedEntityMetadata, common.Equal, "project", "proj")
	assert.NoError(t, err)

	repository := getMockRepositoryForNETest()
	manager := NewNamedEntityManager(repository, getMockConfigForNETest(), mockScope.NewTestScope())
	updatedFilters, err := manager.(*NamedEntityManager).updateQueryFilters([]common.InlineFilter{
		identityFilter,
	}, "eq(state, 0)")
	assert.NoError(t, err)
	assert.Len(t, updatedFilters, 2)

	assert.Equal(t, "project", updatedFilters[0].GetField())
	assert.Equal(t, "state", updatedFilters[1].GetField())
	queryExp, err := updatedFilters[1].GetGormQueryExpr()
	assert.NoError(t, err)
	assert.Equal(t, "COALESCE(state, 0) = ?", queryExp.Query)
	assert.Equal(t, "0", queryExp.Args)
}

func TestNamedEntityManager_Update(t *testing.T) {
	repository := getMockRepositoryForNETest()
	manager := NewNamedEntityManager(repository, getMockConfigForNETest(), mockScope.NewTestScope())
	updatedDescription := "updated description"
	var updateCalled bool

	updateFunction := func(input models.NamedEntity) error {
		updateCalled = true
		assert.Equal(t, input.Description, updatedDescription)
		assert.Equal(t, input.ResourceType, resourceType)
		assert.Equal(t, input.Project, project)
		assert.Equal(t, input.Domain, domain)
		assert.Equal(t, input.Name, name)
		return nil
	}
	repository.NamedEntityRepo().(*repositoryMocks.MockNamedEntityRepo).SetUpdateCallback(updateFunction)
	updatedMetadata := admin.NamedEntityMetadata{
		Description: updatedDescription,
	}
	response, err := manager.UpdateNamedEntity(context.Background(), admin.NamedEntityUpdateRequest{
		Metadata:     &updatedMetadata,
		ResourceType: resourceType,
		Id:           &namedEntityIdentifier,
	})
	assert.True(t, updateCalled)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestNamedEntityManager_Update_BadRequest(t *testing.T) {
	repository := getMockRepositoryForNETest()
	manager := NewNamedEntityManager(repository, getMockConfigForNETest(), mockScope.NewTestScope())
	updatedDescription := "updated description"

	updatedMetadata := admin.NamedEntityMetadata{
		Description: updatedDescription,
	}
	response, err := manager.UpdateNamedEntity(context.Background(), admin.NamedEntityUpdateRequest{
		Metadata:     &updatedMetadata,
		ResourceType: core.ResourceType_UNSPECIFIED,
		Id:           &namedEntityIdentifier,
	})
	assert.Error(t, err)
	assert.Nil(t, response)

	response, err = manager.UpdateNamedEntity(context.Background(), admin.NamedEntityUpdateRequest{
		Metadata:     &updatedMetadata,
		ResourceType: resourceType,
		Id:           &badIdentifier,
	})
	assert.Error(t, err)
	assert.Nil(t, response)
}
