package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func setupProjectService(t *testing.T) *ProjectService {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Project{}))
	return NewProjectService(impl.NewProjectRepo(db), []*project.Domain{
		{Id: "development", Name: "Development"},
		{Id: "production", Name: "Production"},
	})
}

func TestProjectCRUD(t *testing.T) {
	ctx := context.Background()
	svc := setupProjectService(t)

	_, err := svc.CreateProject(ctx, connect.NewRequest(&project.CreateProjectRequest{
		Project: &project.Project{
			Id:          "p1",
			Name:        "Project 1",
			Description: "d1",
			Domains: []*project.Domain{
				{Id: "development", Name: "Development"},
			},
			State: project.ProjectState_PROJECT_STATE_ACTIVE,
			Org:   "org1",
		},
	}))
	require.NoError(t, err)

	getResp, err := svc.GetProject(ctx, connect.NewRequest(&project.GetProjectRequest{
		Id:  "p1",
		Org: "org1",
	}))
	require.NoError(t, err)
	assert.Equal(t, "Project 1", getResp.Msg.GetProject().GetName())
	assert.Equal(t, project.ProjectState_PROJECT_STATE_ACTIVE, getResp.Msg.GetProject().GetState())
	assert.Equal(t, "development", getResp.Msg.GetProject().GetDomains()[0].GetId())

	_, err = svc.UpdateProject(ctx, connect.NewRequest(&project.UpdateProjectRequest{
		Project: &project.Project{
			Id:          "p1",
			Name:        "Project 1 Updated",
			Description: "d2",
			Domains: []*project.Domain{
				{Id: "production", Name: "Production"},
			},
			State: project.ProjectState_PROJECT_STATE_ARCHIVED,
			Org:   "org1",
		},
	}))
	require.NoError(t, err)

	getResp, err = svc.GetProject(ctx, connect.NewRequest(&project.GetProjectRequest{
		Id:  "p1",
		Org: "org1",
	}))
	require.NoError(t, err)
	assert.Equal(t, "Project 1 Updated", getResp.Msg.GetProject().GetName())
	assert.Equal(t, project.ProjectState_PROJECT_STATE_ARCHIVED, getResp.Msg.GetProject().GetState())
	assert.Equal(t, "development", getResp.Msg.GetProject().GetDomains()[0].GetId())
}

func TestListProjects_DefaultExcludesArchived(t *testing.T) {
	ctx := context.Background()
	svc := setupProjectService(t)

	_, err := svc.CreateProject(ctx, connect.NewRequest(&project.CreateProjectRequest{
		Project: &project.Project{
			Id:    "active",
			Name:  "Active",
			State: project.ProjectState_PROJECT_STATE_ACTIVE,
			Org:   "org1",
		},
	}))
	require.NoError(t, err)

	_, err = svc.CreateProject(ctx, connect.NewRequest(&project.CreateProjectRequest{
		Project: &project.Project{
			Id:    "archived",
			Name:  "Archived",
			State: project.ProjectState_PROJECT_STATE_ARCHIVED,
			Org:   "org1",
		},
	}))
	require.NoError(t, err)

	resp, err := svc.ListProjects(ctx, connect.NewRequest(&project.ListProjectsRequest{
		Org:   "org1",
		Limit: 10,
	}))
	require.NoError(t, err)

	require.NotNil(t, resp.Msg.GetProjects())
	assert.Len(t, resp.Msg.GetProjects().GetProjects(), 1)
	assert.Equal(t, "active", resp.Msg.GetProjects().GetProjects()[0].GetId())
}

func TestListProjects_WithFiltersIncludesArchived(t *testing.T) {
	ctx := context.Background()
	svc := setupProjectService(t)

	_, err := svc.CreateProject(ctx, connect.NewRequest(&project.CreateProjectRequest{
		Project: &project.Project{
			Id:    "active",
			Name:  "Active",
			State: project.ProjectState_PROJECT_STATE_ACTIVE,
			Org:   "org1",
		},
	}))
	require.NoError(t, err)

	_, err = svc.CreateProject(ctx, connect.NewRequest(&project.CreateProjectRequest{
		Project: &project.Project{
			Id:    "archived",
			Name:  "Archived",
			State: project.ProjectState_PROJECT_STATE_ARCHIVED,
			Org:   "org1",
		},
	}))
	require.NoError(t, err)

	resp, err := svc.ListProjects(ctx, connect.NewRequest(&project.ListProjectsRequest{
		Org:     "org1",
		Limit:   10,
		Filters: "state == archived",
	}))
	require.NoError(t, err)

	require.NotNil(t, resp.Msg.GetProjects())
	assert.Len(t, resp.Msg.GetProjects().GetProjects(), 2)
}

func TestListProjects_ZeroLimitDoesNotApplyLimit(t *testing.T) {
	ctx := context.Background()
	svc := setupProjectService(t)

	_, err := svc.CreateProject(ctx, connect.NewRequest(&project.CreateProjectRequest{
		Project: &project.Project{
			Id:    "one",
			Name:  "One",
			State: project.ProjectState_PROJECT_STATE_ACTIVE,
			Org:   "org1",
		},
	}))
	require.NoError(t, err)

	_, err = svc.CreateProject(ctx, connect.NewRequest(&project.CreateProjectRequest{
		Project: &project.Project{
			Id:    "two",
			Name:  "Two",
			State: project.ProjectState_PROJECT_STATE_ACTIVE,
			Org:   "org1",
		},
	}))
	require.NoError(t, err)

	resp, err := svc.ListProjects(ctx, connect.NewRequest(&project.ListProjectsRequest{
		Org: "org1",
	}))
	require.NoError(t, err)
	require.NotNil(t, resp.Msg.GetProjects())
	assert.Len(t, resp.Msg.GetProjects().GetProjects(), 2)
	assert.Equal(t, "", resp.Msg.GetProjects().GetToken())
}
