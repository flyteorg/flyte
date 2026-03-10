package transformers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestProjectModelRoundTrip(t *testing.T) {
	p := &project.Project{
		Id:          "p1",
		Name:        "Project 1",
		Description: "desc",
		Domains: []*project.Domain{
			{Id: "development", Name: "Development"},
		},
		Labels: &task.Labels{
			Values: map[string]string{"team": "infra"},
		},
		State: project.ProjectState_PROJECT_STATE_ACTIVE,
		Org:   "org1",
	}

	model, err := NewProjectModel(p)
	require.NoError(t, err)
	require.NotNil(t, model)
	assert.Equal(t, "p1", model.Identifier)
	assert.NotEmpty(t, model.Labels)
	require.NotNil(t, model.State)
	assert.Equal(t, int32(project.ProjectState_PROJECT_STATE_ACTIVE), *model.State)

	domains := []*project.Domain{{Id: "production", Name: "Production"}}
	restored, err := ProjectModelToProject(model, domains)
	require.NoError(t, err)
	assert.Equal(t, p.Id, restored.Id)
	assert.Equal(t, p.Name, restored.Name)
	assert.Equal(t, p.Description, restored.Description)
	assert.Equal(t, p.State, restored.State)
	require.NotNil(t, restored.GetLabels())
	assert.Equal(t, "infra", restored.GetLabels().GetValues()["team"])
	require.Len(t, restored.Domains, 1)
	assert.Equal(t, "production", restored.Domains[0].Id)
}

func TestProjectModelToProject_InvalidPayload(t *testing.T) {
	_, err := ProjectModelToProject(&models.Project{
		Labels: []byte("invalid-proto"),
	}, nil)
	require.Error(t, err)
}
