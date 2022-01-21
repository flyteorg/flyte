package impl

import (
	"context"
	"errors"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repoMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	configMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errFoo = errors.New("foo")

func TestGetClusterResourceAttributes(t *testing.T) {
	project := "flytesnacks"
	domain := "development"
	var attributes = map[string]string{
		"K1": "V1",
		"K2": "V2",
	}
	resourceManager := mocks.MockResourceManager{}
	t.Run("happy case", func(t *testing.T) {
		resourceManager.GetResourceFunc = func(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error) {
			return &interfaces.ResourceResponse{
				Project:      request.Project,
				Domain:       request.Domain,
				ResourceType: admin.MatchableResource_CLUSTER_RESOURCE.String(),
				Attributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_ClusterResourceAttributes{
						ClusterResourceAttributes: &admin.ClusterResourceAttributes{
							Attributes: attributes,
						},
					},
				},
			}, nil
		}
		provider := dbAdminProvider{
			resourceManager: &resourceManager,
		}
		attrs, err := provider.GetClusterResourceAttributes(context.TODO(), project, domain)
		assert.NoError(t, err)
		assert.EqualValues(t, attrs.Attributes, attributes)
	})
	t.Run("error", func(t *testing.T) {
		resourceManager.GetResourceFunc = func(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error) {
			return nil, errFoo
		}
		provider := dbAdminProvider{
			resourceManager: &resourceManager,
		}
		_, err := provider.GetClusterResourceAttributes(context.TODO(), project, domain)
		assert.EqualError(t, err, errFoo.Error())
	})
	t.Run("weird db response", func(t *testing.T) {
		resourceManager.GetResourceFunc = func(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error) {
			return &interfaces.ResourceResponse{
				Project:      request.Project,
				Domain:       request.Domain,
				ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
				Attributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
						ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
							Tags: []string{"foo", "bar", "baz"},
						},
					},
				},
			}, nil
		}
		provider := dbAdminProvider{
			resourceManager: &resourceManager,
		}
		attrs, err := provider.GetClusterResourceAttributes(context.TODO(), project, domain)
		assert.Nil(t, attrs)
		s, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, s.Code(), codes.NotFound)
	})
}

func TestGetProjects(t *testing.T) {
	mockApplicationConfig := configMocks.MockApplicationProvider{}
	mockApplicationConfig.SetDomainsConfig([]runtimeInterfaces.Domain{
		{
			Name: "development",
		},
		{
			Name: "production",
		},
	})
	mockConfig := configMocks.NewMockConfigurationProvider(&mockApplicationConfig, nil, nil, nil, nil, nil)

	t.Run("happy case", func(t *testing.T) {
		mockRepo := repoMocks.NewMockRepository()
		activeProjectState := int32(0)
		mockRepo.(*repoMocks.MockRepository).ProjectRepoIface = &repoMocks.MockProjectRepo{
			ListProjectsFunction: func(ctx context.Context, input repoInterfaces.ListResourceInput) ([]models.Project, error) {
				assert.Len(t, input.InlineFilters, 1)
				assert.Equal(t, input.SortParameter.GetGormOrderExpr(), "created_at desc")
				return []models.Project{
					{
						Identifier: "flytesnacks",
						State:      &activeProjectState,
					},
					{
						Identifier: "flyteexamples",
						State:      &activeProjectState,
					},
				}, nil
			},
		}

		provider := dbAdminProvider{
			db:     mockRepo,
			config: mockConfig,
		}
		projects, err := provider.GetProjects(context.TODO())
		assert.NoError(t, err)
		assert.Len(t, projects.Projects, 2)
	})
	t.Run("db error", func(t *testing.T) {
		mockRepo := repoMocks.NewMockRepository()
		mockRepo.(*repoMocks.MockRepository).ProjectRepoIface = &repoMocks.MockProjectRepo{
			ListProjectsFunction: func(ctx context.Context, input repoInterfaces.ListResourceInput) ([]models.Project, error) {
				return nil, errFoo
			},
		}
		provider := dbAdminProvider{
			db:     mockRepo,
			config: mockConfig,
		}
		_, err := provider.GetProjects(context.TODO())
		assert.EqualError(t, err, errFoo.Error())
	})
}
