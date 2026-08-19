package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
)

func setupSettingsService(t *testing.T) *SettingsService {
	t.Cleanup(func() { testDB.Exec("DELETE FROM settings") })
	return NewSettingsService(impl.NewSettingsRepo(testDB))
}

func envSettings(key, value string) *settings.Settings {
	return &settings.Settings{
		EnvironmentVariables: &settings.StringMapSetting{
			State:    settings.SettingState_SETTING_STATE_VALUE,
			MapValue: &settings.StringMap{Entries: map[string]string{key: value}},
		},
	}
}

func TestSettingsCRUD(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	key := &settings.SettingsKey{Org: "acme", Domain: "development", Project: "recsys"}

	createResp, err := svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      key,
		Settings: envSettings("LOG_LEVEL", "debug"),
	}))
	require.NoError(t, err)
	assert.Equal(t, uint64(1), createResp.Msg.GetSettingsRecord().GetVersion())
	assert.True(t, proto.Equal(key, createResp.Msg.GetSettingsRecord().GetKey()))

	getResp, err := svc.GetSettingsForEdit(ctx, connect.NewRequest(&settings.GetSettingsForEditRequest{
		Key: key,
	}))
	require.NoError(t, err)
	require.Len(t, getResp.Msg.GetLevels(), 3)
	assert.Equal(t, uint64(1), getResp.Msg.GetLevels()[2].GetVersion())
	assert.True(t, proto.Equal(envSettings("LOG_LEVEL", "debug"), getResp.Msg.GetLevels()[2].GetSettings()))

	updateResp, err := svc.UpdateSettings(ctx, connect.NewRequest(&settings.UpdateSettingsRequest{
		Key:      key,
		Settings: envSettings("LOG_LEVEL", "warn"),
		Version:  1,
	}))
	require.NoError(t, err)
	assert.Equal(t, uint64(2), updateResp.Msg.GetSettingsRecord().GetVersion())

	getResp, err = svc.GetSettingsForEdit(ctx, connect.NewRequest(&settings.GetSettingsForEditRequest{
		Key: key,
	}))
	require.NoError(t, err)
	require.Len(t, getResp.Msg.GetLevels(), 3)
	assert.Equal(t, uint64(2), getResp.Msg.GetLevels()[2].GetVersion())
	assert.True(t, proto.Equal(envSettings("LOG_LEVEL", "warn"), getResp.Msg.GetLevels()[2].GetSettings()))
}

func TestCreateSettings_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	key := &settings.SettingsKey{Org: "acme", Domain: "development"}

	_, err := svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      key,
		Settings: envSettings("LOG_LEVEL", "debug"),
	}))
	require.NoError(t, err)

	resp, err := svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      key,
		Settings: envSettings("LOG_LEVEL", "warn"),
	}))
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err))

	getResp, err := svc.GetSettingsForEdit(ctx, connect.NewRequest(&settings.GetSettingsForEditRequest{
		Key: key,
	}))
	require.NoError(t, err)
	require.Len(t, getResp.Msg.GetLevels(), 2)
	assert.Equal(t, uint64(1), getResp.Msg.GetLevels()[1].GetVersion())
	assert.True(t, proto.Equal(envSettings("LOG_LEVEL", "debug"), getResp.Msg.GetLevels()[1].GetSettings()))
}

func TestUpdateSettings_StaleVersion(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	key := &settings.SettingsKey{Org: "acme", Domain: "development"}

	_, err := svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      key,
		Settings: envSettings("LOG_LEVEL", "debug"),
	}))
	require.NoError(t, err)

	// First writer wins, moving the row to version 2.
	_, err = svc.UpdateSettings(ctx, connect.NewRequest(&settings.UpdateSettingsRequest{
		Key:      key,
		Settings: envSettings("LOG_LEVEL", "info"),
		Version:  1,
	}))
	require.NoError(t, err)

	// Second writer still believes the row is at version 1.
	resp, err := svc.UpdateSettings(ctx, connect.NewRequest(&settings.UpdateSettingsRequest{
		Key:      key,
		Settings: envSettings("LOG_LEVEL", "warn"),
		Version:  1,
	}))
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))

	getResp, err := svc.GetSettingsForEdit(ctx, connect.NewRequest(&settings.GetSettingsForEditRequest{
		Key: key,
	}))
	require.NoError(t, err)
	require.Len(t, getResp.Msg.GetLevels(), 2)
	assert.Equal(t, uint64(2), getResp.Msg.GetLevels()[1].GetVersion())
	assert.True(t, proto.Equal(envSettings("LOG_LEVEL", "info"), getResp.Msg.GetLevels()[1].GetSettings()))
}

func TestUpdateSettings_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	resp, err := svc.UpdateSettings(ctx, connect.NewRequest(&settings.UpdateSettingsRequest{
		Key:      &settings.SettingsKey{Org: "acme", Domain: "nowhere"},
		Settings: envSettings("LOG_LEVEL", "debug"),
		Version:  1,
	}))
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestCreateSettings_Validation(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	tests := []struct {
		name string
		req  *settings.CreateSettingsRequest
	}{
		{
			name: "nil key",
			req:  &settings.CreateSettingsRequest{Settings: envSettings("LOG_LEVEL", "debug")},
		},
		{
			name: "nil settings",
			req:  &settings.CreateSettingsRequest{Key: &settings.SettingsKey{Org: "acme", Domain: "development"}},
		},
		{
			name: "project without domain",
			req: &settings.CreateSettingsRequest{
				Key:      &settings.SettingsKey{Org: "acme", Project: "recsys"},
				Settings: envSettings("LOG_LEVEL", "debug"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.CreateSettings(ctx, connect.NewRequest(tt.req))
			assert.Nil(t, resp)
			assert.Error(t, err)
			assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		})
	}
}

func TestUpdateSettings_Validation(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	tests := []struct {
		name string
		req  *settings.UpdateSettingsRequest
	}{
		{
			name: "nil key",
			req: &settings.UpdateSettingsRequest{
				Settings: envSettings("LOG_LEVEL", "debug"),
				Version:  1,
			},
		},
		{
			name: "nil settings",
			req: &settings.UpdateSettingsRequest{
				Key:     &settings.SettingsKey{Org: "acme", Domain: "development"},
				Version: 1,
			},
		},
		{
			name: "project without domain",
			req: &settings.UpdateSettingsRequest{
				Key:      &settings.SettingsKey{Org: "acme", Project: "recsys"},
				Settings: envSettings("LOG_LEVEL", "debug"),
				Version:  1,
			},
		},
		{
			name: "version zero",
			req: &settings.UpdateSettingsRequest{
				Key:      &settings.SettingsKey{Org: "acme", Domain: "development"},
				Settings: envSettings("LOG_LEVEL", "debug"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.UpdateSettings(ctx, connect.NewRequest(tt.req))
			assert.Nil(t, resp)
			assert.Error(t, err)
			assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		})
	}
}

func TestGetSettingsForEdit_Validation(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	tests := []struct {
		name string
		req  *settings.GetSettingsForEditRequest
	}{
		{
			name: "nil key",
			req:  &settings.GetSettingsForEditRequest{},
		},
		{
			name: "project without domain",
			req: &settings.GetSettingsForEditRequest{
				Key: &settings.SettingsKey{Org: "acme", Project: "recsys"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.GetSettingsForEdit(ctx, connect.NewRequest(tt.req))
			assert.Nil(t, resp)
			assert.Error(t, err)
			assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		})
	}
}

func TestGetSettingsForEdit_AllLevels(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	orgKey := &settings.SettingsKey{Org: "acme"}
	domainKey := &settings.SettingsKey{Org: "acme", Domain: "development"}
	projectKey := &settings.SettingsKey{Org: "acme", Domain: "development", Project: "recsys"}

	_, err := svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      orgKey,
		Settings: envSettings("LOG_LEVEL", "org"),
	}))
	require.NoError(t, err)

	_, err = svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      domainKey,
		Settings: envSettings("LOG_LEVEL", "domain"),
	}))
	require.NoError(t, err)

	_, err = svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      projectKey,
		Settings: envSettings("LOG_LEVEL", "project"),
	}))
	require.NoError(t, err)

	resp, err := svc.GetSettingsForEdit(ctx, connect.NewRequest(&settings.GetSettingsForEditRequest{
		Key: projectKey,
	}))
	require.NoError(t, err)
	assert.True(t, proto.Equal(projectKey, resp.Msg.GetRequestedKey()))

	levels := resp.Msg.GetLevels()
	require.Len(t, levels, 3)

	// Each level carries only what it stores. Nothing is merged, which is the job
	// of GetSettings, which is not implemented yet.
	assert.True(t, proto.Equal(orgKey, levels[0].GetKey()))
	assert.True(t, proto.Equal(envSettings("LOG_LEVEL", "org"), levels[0].GetSettings()))
	assert.Equal(t, uint64(1), levels[0].GetVersion())

	assert.True(t, proto.Equal(domainKey, levels[1].GetKey()))
	assert.True(t, proto.Equal(envSettings("LOG_LEVEL", "domain"), levels[1].GetSettings()))
	assert.Equal(t, uint64(1), levels[1].GetVersion())

	assert.True(t, proto.Equal(projectKey, levels[2].GetKey()))
	assert.True(t, proto.Equal(envSettings("LOG_LEVEL", "project"), levels[2].GetSettings()))
	assert.Equal(t, uint64(1), levels[2].GetVersion())
}

func TestGetSettingsForEdit_MissingLevelsAreEmpty(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	orgKey := &settings.SettingsKey{Org: "acme"}
	domainKey := &settings.SettingsKey{Org: "acme", Domain: "development"}
	projectKey := &settings.SettingsKey{Org: "acme", Domain: "development", Project: "recsys"}

	_, err := svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      orgKey,
		Settings: envSettings("LOG_LEVEL", "org"),
	}))
	require.NoError(t, err)

	resp, err := svc.GetSettingsForEdit(ctx, connect.NewRequest(&settings.GetSettingsForEditRequest{
		Key: projectKey,
	}))
	require.NoError(t, err)

	// Three levels because the request key has three fields set, not because
	// three rows exist. Only the org row was created.
	levels := resp.Msg.GetLevels()
	require.Len(t, levels, 3)

	assert.True(t, proto.Equal(orgKey, levels[0].GetKey()))
	assert.True(t, proto.Equal(envSettings("LOG_LEVEL", "org"), levels[0].GetSettings()))
	assert.Equal(t, uint64(1), levels[0].GetVersion())

	// Absent levels come back empty at version 0, which tells the client to call
	// CreateSettings here rather than UpdateSettings.
	assert.True(t, proto.Equal(domainKey, levels[1].GetKey()))
	assert.True(t, proto.Equal(&settings.Settings{}, levels[1].GetSettings()))
	assert.Equal(t, uint64(0), levels[1].GetVersion())

	assert.True(t, proto.Equal(projectKey, levels[2].GetKey()))
	assert.True(t, proto.Equal(&settings.Settings{}, levels[2].GetSettings()))
	assert.Equal(t, uint64(0), levels[2].GetVersion())
}

func TestGetSettingsForEdit_LevelsFollowKeyDepth(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	// No fixtures on purpose: with an empty table, the level count can only be
	// coming from the key.
	tests := []struct {
		name       string
		key        *settings.SettingsKey
		wantLevels int
	}{
		{
			name:       "org only",
			key:        &settings.SettingsKey{Org: "acme"},
			wantLevels: 1,
		},
		{
			name:       "org and domain",
			key:        &settings.SettingsKey{Org: "acme", Domain: "development"},
			wantLevels: 2,
		},
		{
			name:       "org domain and project",
			key:        &settings.SettingsKey{Org: "acme", Domain: "development", Project: "recsys"},
			wantLevels: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.GetSettingsForEdit(ctx, connect.NewRequest(&settings.GetSettingsForEditRequest{
				Key: tt.key,
			}))
			require.NoError(t, err)
			assert.Len(t, resp.Msg.GetLevels(), tt.wantLevels)
		})
	}
}
