package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/plugin"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func TestServiceGetClusterResourceAttributes(t *testing.T) {
	ctx := context.TODO()
	org := "testorg"
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
		attrs, err := provider.GetClusterResourceAttributes(context.TODO(), org, project, domain)
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
		_, err := provider.GetClusterResourceAttributes(context.TODO(), org, project, domain)
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
		attrs, err := provider.GetClusterResourceAttributes(context.TODO(), org, project, domain)
		assert.Nil(t, attrs)
		s, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, s.Code(), codes.NotFound)
	})
}

func TestServiceGetProjects(t *testing.T) {
	ctx := context.TODO()
	clusterResourcePlugin := plugin.NewDefaultClusterResourcePlugin()

	t.Run("happy case", func(t *testing.T) {
		mockAdmin := mocks.AdminServiceClient{}
		mockAdmin.OnListProjectsMatch(ctx, mock.MatchedBy(func(req *admin.ProjectListRequest) bool {
			res := req.Limit == 100 && req.Filters == "value_not_in(state,1;3)" && req.SortBy.Key == "updated_at"
			return res
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
		projects, err := provider.GetProjects(ctx, clusterResourcePlugin)
		assert.NoError(t, err)
		assert.Len(t, projects.Projects, 2)
	})
	t.Run("admin error", func(t *testing.T) {
		mockAdmin := mocks.AdminServiceClient{}
		mockAdmin.OnListProjectsMatch(ctx, mock.MatchedBy(func(req *admin.ProjectListRequest) bool {
			return req.Limit == 100 && req.Filters == "value_not_in(state,1;3)" && req.SortBy.Key == "updated_at"
		})).Return(nil, errFoo)
		provider := serviceAdminProvider{
			adminClient: &mockAdmin,
		}
		_, err := provider.GetProjects(ctx, clusterResourcePlugin)
		assert.EqualError(t, err, errFoo.Error())
	})
}
