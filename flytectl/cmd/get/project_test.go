package get

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytectl/pkg/filters"
	"github.com/stretchr/testify/assert"
)

var (
	resourceListRequestProject *admin.ProjectListRequest
	projectListResponse        *admin.Projects
	argsProject                = []string{"flyteexample"}
	project1                   *admin.Project
)

func getProjectSetup() {
	resourceListRequestProject = &admin.ProjectListRequest{}

	project1 = &admin.Project{
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

	projects := []*admin.Project{project1, project2}

	projectListResponse = &admin.Projects{
		Projects: projects,
	}
}

func TestListProjectFunc(t *testing.T) {
	s := testutils.SetupWithExt()
	getProjectSetup()
	project.DefaultConfig.Filter = filters.Filters{}
	s.MockAdminClient.OnListProjectsMatch(s.Ctx, resourceListRequestProject).Return(projectListResponse, nil)
	s.FetcherExt.OnListProjects(s.Ctx, filters.Filters{}).Return(projectListResponse, nil)
	err := getProjectsFunc(s.Ctx, argsProject, s.CmdCtx)

	assert.Nil(t, err)
	s.FetcherExt.AssertCalled(t, "ListProjects", s.Ctx, filters.Filters{})
}

func TestGetProjectFunc(t *testing.T) {
	s := testutils.SetupWithExt()
	getProjectSetup()
	argsProject = []string{}

	project.DefaultConfig.Filter = filters.Filters{}
	s.MockAdminClient.OnListProjectsMatch(s.Ctx, resourceListRequestProject).Return(projectListResponse, nil)
	s.FetcherExt.OnListProjects(s.Ctx, filters.Filters{}).Return(projectListResponse, nil)
	err := getProjectsFunc(s.Ctx, argsProject, s.CmdCtx)
	assert.Nil(t, err)
	s.FetcherExt.AssertCalled(t, "ListProjects", s.Ctx, filters.Filters{})
}

func TestGetProjectFuncError(t *testing.T) {
	s := testutils.SetupWithExt()
	getProjectSetup()
	project.DefaultConfig.Filter = filters.Filters{
		FieldSelector: "hello=",
	}
	s.MockAdminClient.OnListProjectsMatch(s.Ctx, resourceListRequestProject).Return(nil, fmt.Errorf("Please add a valid field selector"))
	s.FetcherExt.OnListProjects(s.Ctx, filters.Filters{
		FieldSelector: "hello=",
	}).Return(nil, fmt.Errorf("Please add a valid field selector"))
	err := getProjectsFunc(s.Ctx, argsProject, s.CmdCtx)
	assert.NotNil(t, err)
}
