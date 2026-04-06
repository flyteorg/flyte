package impl

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestCreateProject_ReturnsAlreadyExists(t *testing.T) {
	db := setupDB(t)
	t.Cleanup(func() { db.Exec("DELETE FROM projects") })

	repo := NewProjectRepo(db)
	ctx := context.Background()
	state := int32(0)

	require.NoError(t, repo.CreateProject(ctx, &models.Project{
		Identifier: "flytesnacks",
		Name:       "flytesnacks",
		State:      &state,
	}))

	err := repo.CreateProject(ctx, &models.Project{
		Identifier: "flytesnacks",
		Name:       "flytesnacks",
		State:      &state,
	})
	require.ErrorIs(t, err, interfaces.ErrProjectAlreadyExists)
	require.False(t, strings.Contains(strings.ToUpper(err.Error()), "UNIQUE"))
}
