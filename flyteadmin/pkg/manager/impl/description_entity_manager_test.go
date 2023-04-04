package impl

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

var descriptionEntityIdentifier = core.Identifier{
	ResourceType: core.ResourceType_WORKFLOW,
	Project:      project,
	Domain:       domain,
	Name:         name,
	Version:      version,
}

var badDescriptionEntityIdentifier = core.Identifier{
	ResourceType: core.ResourceType_WORKFLOW,
	Project:      project,
	Domain:       domain,
	Name:         "",
	Version:      version,
}

func getMockRepositoryForDETest() interfaces.Repository {
	return repositoryMocks.NewMockRepository()
}

func getMockConfigForDETest() runtimeInterfaces.Configuration {
	mockConfig := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultDomains(), nil, nil, nil, nil, nil)
	return mockConfig
}

func TestDescriptionEntityManager_Get(t *testing.T) {
	repository := getMockRepositoryForDETest()
	manager := NewDescriptionEntityManager(repository, getMockConfigForDETest(), mockScope.NewTestScope())

	response, err := manager.GetDescriptionEntity(context.Background(), admin.ObjectGetRequest{
		Id: &descriptionEntityIdentifier,
	})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = manager.GetDescriptionEntity(context.Background(), admin.ObjectGetRequest{
		Id: &badDescriptionEntityIdentifier,
	})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestDescriptionEntityManager_List(t *testing.T) {
	repository := getMockRepositoryForDETest()
	manager := NewDescriptionEntityManager(repository, getMockConfigForDETest(), mockScope.NewTestScope())

	t.Run("failed to validate a request", func(t *testing.T) {
		response, err := manager.ListDescriptionEntity(context.Background(), admin.DescriptionEntityListRequest{
			Id: &admin.NamedEntityIdentifier{
				Name: "flyte",
			},
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("failed to sort description entity", func(t *testing.T) {
		response, err := manager.ListDescriptionEntity(context.Background(), admin.DescriptionEntityListRequest{
			ResourceType: core.ResourceType_TASK,
			Id: &admin.NamedEntityIdentifier{
				Name:    "flyte",
				Project: "project",
				Domain:  "domain",
			},
			Limit:  1,
			SortBy: &admin.Sort{Direction: 3},
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("failed to validate token", func(t *testing.T) {
		response, err := manager.ListDescriptionEntity(context.Background(), admin.DescriptionEntityListRequest{
			ResourceType: core.ResourceType_TASK,
			Id: &admin.NamedEntityIdentifier{
				Name:    "flyte",
				Project: "project",
				Domain:  "domain",
			},
			Limit: 1,
			Token: "hello",
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("list description entities in the task", func(t *testing.T) {
		response, err := manager.ListDescriptionEntity(context.Background(), admin.DescriptionEntityListRequest{
			ResourceType: core.ResourceType_TASK,
			Id: &admin.NamedEntityIdentifier{
				Name:    "flyte",
				Project: "project",
				Domain:  "domain",
			},
			Limit: 1,
		})
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("list description entities in the workflow", func(t *testing.T) {
		response, err := manager.ListDescriptionEntity(context.Background(), admin.DescriptionEntityListRequest{
			ResourceType: core.ResourceType_WORKFLOW,
			Id: &admin.NamedEntityIdentifier{
				Name:    "flyte",
				Project: "project",
				Domain:  "domain",
			},
			Limit: 1,
		})
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("failed to get filter", func(t *testing.T) {
		response, err := manager.ListDescriptionEntity(context.Background(), admin.DescriptionEntityListRequest{
			ResourceType: core.ResourceType_WORKFLOW,
			Id: &admin.NamedEntityIdentifier{
				Name:    "flyte",
				Project: "project",
				Domain:  "domain",
			},
			Filters: "wrong",
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}
