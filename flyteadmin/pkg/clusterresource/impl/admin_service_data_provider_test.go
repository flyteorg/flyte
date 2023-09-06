package impl

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceGetClusterResourceAttributes(t *testing.T) {
	ctx := context.TODO()
	project := "flytesnacks"
	domain := "development"
	t.Run("happy case", func(t *testing.T) {
		var attributes = map[string]string{
			"K1": "V1",
			"K2": "V2",
		}
		mockAdmin := mocks.AdminServiceClient{}
		mockAdmin.OnGetProjectDomainAttributesMatch(ctx, mock.MatchedBy(func(req *admin.ProjectDomainAttributesGetRequest) bool {
			return req.Project == project && req.Domain == domain && req.ResourceType == admin.MatchableResource_CLUSTER_RESOURCE
		})).Return(&admin.ProjectDomainAttributesGetResponse{
			Attributes: &admin.ProjectDomainAttributes{
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_ClusterResourceAttributes{
						ClusterResourceAttributes: &admin.ClusterResourceAttributes{
							Attributes: attributes,
						},
					},
				},
			},
		}, nil)

		provider := serviceAdminProvider{
			adminClient: &mockAdmin,
		}
		attrs, err := provider.GetClusterResourceAttributes(context.TODO(), project, domain)
		assert.NoError(t, err)
		assert.EqualValues(t, attrs.Attributes, attributes)
	})
	t.Run("admin service error", func(t *testing.T) {
		mockAdmin := mocks.AdminServiceClient{}
		mockAdmin.OnGetProjectDomainAttributesMatch(ctx, mock.MatchedBy(func(req *admin.ProjectDomainAttributesGetRequest) bool {
			return req.Project == project && req.Domain == domain && req.ResourceType == admin.MatchableResource_CLUSTER_RESOURCE
		})).Return(&admin.ProjectDomainAttributesGetResponse{}, errFoo)

		provider := serviceAdminProvider{
			adminClient: &mockAdmin,
		}
		_, err := provider.GetClusterResourceAttributes(context.TODO(), project, domain)
		assert.EqualError(t, err, errFoo.Error())
	})
	t.Run("wonky admin service response", func(t *testing.T) {
		mockAdmin := mocks.AdminServiceClient{}
		mockAdmin.OnGetProjectDomainAttributesMatch(ctx, mock.MatchedBy(func(req *admin.ProjectDomainAttributesGetRequest) bool {
			return req.Project == project && req.Domain == domain && req.ResourceType == admin.MatchableResource_CLUSTER_RESOURCE
		})).Return(&admin.ProjectDomainAttributesGetResponse{
			Attributes: &admin.ProjectDomainAttributes{
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
						ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
							Tags: []string{"foo", "bar", "baz"},
						},
					},
				},
			},
		}, nil)

		provider := serviceAdminProvider{
			adminClient: &mockAdmin,
		}
		attrs, err := provider.GetClusterResourceAttributes(context.TODO(), project, domain)
		assert.Nil(t, attrs)
		s, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, s.Code(), codes.NotFound)
	})
}

func TestServiceGetProjects(t *testing.T) {
	ctx := context.TODO()
	t.Run("happy case", func(t *testing.T) {
		mockAdmin := mocks.AdminServiceClient{}
		mockAdmin.OnListProjectsMatch(ctx, mock.MatchedBy(func(req *admin.ProjectListRequest) bool {
			return req.Limit == 100 && req.Filters == "ne(state,1)" && req.SortBy.Key == "created_at"
		})).Return(&admin.Projects{
			Projects: []*admin.Project{
				{
					Id: "flytesnacks",
				},
				{
					Id: "flyteexamples",
				},
			},
		}, nil)
		provider := serviceAdminProvider{
			adminClient: &mockAdmin,
		}
		projects, err := provider.GetProjects(ctx)
		assert.NoError(t, err)
		assert.Len(t, projects.Projects, 2)
	})
	t.Run("admin error", func(t *testing.T) {
		mockAdmin := mocks.AdminServiceClient{}
		mockAdmin.OnListProjectsMatch(ctx, mock.MatchedBy(func(req *admin.ProjectListRequest) bool {
			return req.Limit == 100 && req.Filters == "ne(state,1)" && req.SortBy.Key == "created_at"
		})).Return(nil, errFoo)
		provider := serviceAdminProvider{
			adminClient: &mockAdmin,
		}
		_, err := provider.GetProjects(ctx)
		assert.EqualError(t, err, errFoo.Error())
	})
}
