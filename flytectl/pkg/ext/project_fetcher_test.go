package ext

import (
	"testing"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdminFetcherExtClient_ListProjects(t *testing.T) {

	project1 := &admin.Project{
		Id:   "flyteexample",
		Name: "flyteexample",
		Domains: []*admin.Domain{
			{
				Id:   "development",
				Name: "development",
			},
		},
	}

	project2 := &admin.Project{
		Id:   "flytesnacks",
		Name: "flytesnacks",
		Domains: []*admin.Domain{
			{
				Id:   "development",
				Name: "development",
			},
		},
	}

	adminClient = new(mocks.AdminServiceClient)
	adminFetcherExt = AdminFetcherExtClient{AdminClient: adminClient}

	projects := &admin.Projects{
		Projects: []*admin.Project{project1, project2},
	}
	adminClient.OnListProjectsMatch(mock.Anything, mock.Anything).Return(projects, nil)
	_, err := adminFetcherExt.ListProjects(ctx, taskFilter)
	assert.Nil(t, err)
}
