package impl

import (
	"context"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestCreateProject_ReturnsAlreadyExists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Project{}))

	repo := NewProjectRepo(db)
	ctx := context.Background()
	state := int32(0)

	require.NoError(t, repo.CreateProject(ctx, &models.Project{
		Identifier: "flytesnacks",
		Name:       "flytesnacks",
		State:      &state,
	}))

	err = repo.CreateProject(ctx, &models.Project{
		Identifier: "flytesnacks",
		Name:       "flytesnacks",
		State:      &state,
	})
	require.ErrorIs(t, err, interfaces.ErrProjectAlreadyExists)
	require.False(t, strings.Contains(strings.ToUpper(err.Error()), "UNIQUE"))
}
