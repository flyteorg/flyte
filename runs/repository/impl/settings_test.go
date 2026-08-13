package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func setupSettingTest(t *testing.T) interfaces.SettingsRepo {
	db := setupDB(t)
	t.Cleanup(func() { db.Exec("DELETE FROM settings") })
	return NewSettingsRepo(db)
}

func TestCreateAndGetSettings_RoundTrip(t *testing.T) {
	repo := setupSettingTest(t)
	ctx := context.Background()

	key := models.EncodeSettingsKey("", "development", "")
	created := &models.Settings{
		Key:  key,
		Data: []byte(`{"environmentVariables":{"state":"VALUE", "values":{"LOG_LEVEL": "debug"}}}`),
	}
	require.NoError(t, repo.CreateSettings(ctx, created))
	require.Equal(t, uint64(1), created.Version)

	got, err := repo.GetSettings(ctx, key)
	require.NoError(t, err)
	require.Equal(t, key, got.Key)
	require.Equal(t, uint64(1), got.Version)
	require.JSONEq(t, string(created.Data), string(got.Data))
	require.False(t, got.CreatedAt.IsZero())
}

func TestCreateSettings_AlreadyExists(t *testing.T) {
	repo := setupSettingTest(t)
	ctx := context.Background()

	key := models.EncodeSettingsKey("", "development", "recsys")
	first := &models.Settings{Key: key, Data: []byte(`{}`)}
	require.NoError(t, repo.CreateSettings(ctx, first))

	dup := &models.Settings{Key: key, Data: []byte(`{"labels": {}}`)}
	err := repo.CreateSettings(ctx, dup)
	require.ErrorIs(t, err, interfaces.ErrSettingsAlreadyExists)

	got, err := repo.GetSettings(ctx, key)
	require.NoError(t, err)
	require.JSONEq(t, `{}`, string(got.Data))
}

func TestGetSettings_NotFound(t *testing.T) {
	repo := setupSettingTest(t)
	ctx := context.Background()

	got, err := repo.GetSettings(ctx, models.EncodeSettingsKey("", "nowhere", ""))
	require.ErrorIs(t, err, interfaces.ErrSettingsNotFound)
	require.Nil(t, got)
}

func TestGetSettingsByKeys_MissingRowsAreNotErrors(t *testing.T) {
	repo := setupSettingTest(t)
	ctx := context.Background()

	instanceKey := models.EncodeSettingsKey("", "", "")
	domainKey := models.EncodeSettingsKey("", "development", "")
	projectKey := models.EncodeSettingsKey("", "development", "recsys")

	require.NoError(t, repo.CreateSettings(ctx, &models.Settings{Key: instanceKey, Data: []byte(`{}`)}))
	require.NoError(t, repo.CreateSettings(ctx, &models.Settings{Key: domainKey, Data: []byte(`{}`)}))
	// deliberately no row for projectKey

	rows, err := repo.GetSettingsByKeys(ctx, []string{instanceKey, domainKey, projectKey})
	require.NoError(t, err)
	require.Len(t, rows, 2)

	foundKeys := make(map[string]bool)
	for _, row := range rows {
		foundKeys[row.Key] = true
	}
	require.True(t, foundKeys[instanceKey])
	require.True(t, foundKeys[domainKey])
	require.False(t, foundKeys[projectKey])
}

func TestUpdateSettings_VersionConflict(t *testing.T) {
	repo := setupSettingTest(t)
	ctx := context.Background()

	key := models.EncodeSettingsKey("", "development", "recsys")

	// The row is born at version 1.
	writer := &models.Settings{Key: key, Data: []byte(`{"run":{"defaultQueue":{"state":"VALUE","stringValue":"cpu-pool"}}}`)}
	require.NoError(t, repo.CreateSettings(ctx, writer))

	// A second caller reads the rwo while it's still at version 1.
	staleWriter, err := repo.GetSettings(ctx, key)
	require.NoError(t, err)
	require.Equal(t, uint64(1), staleWriter.Version)

	// First update wins, the writer's struct syncs to version 2.
	writer.Data = []byte(`{"run":{"defaultQueue":{"state":"VALUE","stringValue":"gpu-pool"}}}`)
	require.NoError(t, repo.UpdateSettings(ctx, writer))
	require.Equal(t, uint64(2), writer.Version)

	// The stale writer saves against a version that no longer exists.
	staleWriter.Data = []byte(`{"run":{"maxActionConcurrency":{"state":"VALUE","intValue":"64"}}}`)
	err = repo.UpdateSettings(ctx, staleWriter)
	require.ErrorIs(t, err, interfaces.ErrSettingsVersionConflict)
	require.Equal(t, uint64(1), staleWriter.Version)

	// The table keeps the winner's write, the stale write never landed.
	got, err := repo.GetSettings(ctx, key)
	require.NoError(t, err)
	require.Equal(t, uint64(2), got.Version)
	require.JSONEq(t, string(writer.Data), string(got.Data))
	require.True(t, got.UpdatedAt.After(got.CreatedAt))
}
