package service

import (
	"context"
	"math"
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

const (
	stateValue   = settings.SettingState_SETTING_STATE_VALUE
	stateInherit = settings.SettingState_SETTING_STATE_INHERIT
	stateUnset   = settings.SettingState_SETTING_STATE_UNSET
)

func quantity(state settings.SettingState, value string) *settings.QuantitySetting {
	return &settings.QuantitySetting{State: state, QuantityValue: value}
}

func concurrency(state settings.SettingState, value int64) *settings.Int64Setting {
	return &settings.Int64Setting{State: state, IntValue: value}
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
		{
			name: "invalid quantity",
			req: &settings.CreateSettingsRequest{
				Key: &settings.SettingsKey{Org: "acme", Domain: "development"},
				Settings: &settings.Settings{TaskResource: &settings.TaskResourceSettings{
					Max: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "apple")},
				}},
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
		{
			name: "invalid quantity",
			req: &settings.UpdateSettingsRequest{
				Key: &settings.SettingsKey{Org: "acme", Domain: "development"},
				Settings: &settings.Settings{TaskResource: &settings.TaskResourceSettings{
					Max: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "apple")},
				}},
				Version: 1,
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
	// of GetSettings.
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

func TestValidateMaxActionConcurrency(t *testing.T) {
	tests := []struct {
		name    string
		setting *settings.Int64Setting
		wantErr bool
	}{
		{"nil leaf", nil, false},
		{"inherit ignores its value", concurrency(stateInherit, 1), false},
		{"unset ignores its value", concurrency(stateUnset, 1), false},
		{"zero means unlimited", concurrency(stateValue, 0), false},
		{"one is rejected", concurrency(stateValue, 1), true},
		{"two is allowed", concurrency(stateValue, 2), false},
		{"negative is rejected", concurrency(stateValue, -1), true},
		{"max uint32 is allowed", concurrency(stateValue, math.MaxUint32), false},
		{"above max uint32 is rejected", concurrency(stateValue, math.MaxUint32+1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMaxActionConcurrency(tt.setting)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateQuantity(t *testing.T) {
	tests := []struct {
		name    string
		setting *settings.QuantitySetting
		wantErr bool
	}{
		{"nil leaf", nil, false},
		{"inherit ignores its value", quantity(stateInherit, "apple"), false},
		{"unset ignores its value", quantity(stateUnset, "apple"), false},
		{"millicore", quantity(stateValue, "500m"), false},
		{"whole cpus", quantity(stateValue, "16"), false},
		{"binary memory", quantity(stateValue, "64Gi"), false},
		{"garbage value is rejected", quantity(stateValue, "apple"), true},
		{"empty value is rejected", quantity(stateValue, ""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQuantity("task_resource.max.cpu", tt.setting)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateTaskResourceDefaults(t *testing.T) {
	t.Run("nil defaults has nothing to check", func(t *testing.T) {
		assert.NoError(t, validateTaskResourceDefaults("task_resource.max", nil))
	})

	t.Run("all four valid pass", func(t *testing.T) {
		assert.NoError(t, validateTaskResourceDefaults("task_resource.max",
			&settings.TaskResourceDefaults{
				Cpu:     quantity(stateValue, "16"),
				Gpu:     quantity(stateValue, "1"),
				Memory:  quantity(stateValue, "64Gi"),
				Storage: quantity(stateValue, "10Gi"),
			}))
	})

	tests := []struct {
		name     string
		defaults *settings.TaskResourceDefaults
		wantPath string
	}{
		{"cpu", &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "apple")}, "task_resource.max.cpu"},
		{"gpu", &settings.TaskResourceDefaults{Gpu: quantity(stateValue, "apple")}, "task_resource.max.gpu"},
		{"memory", &settings.TaskResourceDefaults{Memory: quantity(stateValue, "apple")}, "task_resource.max.memory"},
		{"storage", &settings.TaskResourceDefaults{Storage: quantity(stateValue, "apple")}, "task_resource.max.storage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTaskResourceDefaults("task_resource.max", tt.defaults)
			require.Error(t, err)
			assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
			assert.Contains(t, err.Error(), tt.wantPath)
		})
	}
}

func TestValidateSettings(t *testing.T) {
	t.Run("nil settings has nothing to check", func(t *testing.T) {
		assert.NoError(t, validateSettings(nil))
	})

	t.Run("empty settings has nothing to check", func(t *testing.T) {
		assert.NoError(t, validateSettings(&settings.Settings{}))
	})

	t.Run("valid settings pass", func(t *testing.T) {
		assert.NoError(t, validateSettings(&settings.Settings{
			Run: &settings.RunSettings{MaxActionConcurrency: concurrency(stateValue, 64)},
			TaskResource: &settings.TaskResourceSettings{
				Min: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "500m")},
				Max: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "16")},
			},
		}))
	})

	tests := []struct {
		name     string
		input    *settings.Settings
		wantPath string
	}{
		{
			name: "bad min quantity names the min path",
			input: &settings.Settings{TaskResource: &settings.TaskResourceSettings{
				Min: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "apple")},
			}},
			wantPath: "task_resource.min.cpu",
		},
		{
			name: "bad max quantity names the max path",
			input: &settings.Settings{TaskResource: &settings.TaskResourceSettings{
				Max: &settings.TaskResourceDefaults{Memory: quantity(stateValue, "apple")},
			}},
			wantPath: "task_resource.max.memory",
		},
		{
			name: "bad concurrency is reported",
			input: &settings.Settings{Run: &settings.RunSettings{
				MaxActionConcurrency: concurrency(stateValue, 1),
			}},
			wantPath: "max_action_concurrency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSettings(tt.input)
			require.Error(t, err)
			assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
			assert.Contains(t, err.Error(), tt.wantPath)
		})
	}
}

func TestGetSettings_MergesAcrossLevels(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	orgKey := &settings.SettingsKey{Org: "acme"}
	projectKey := &settings.SettingsKey{Org: "acme", Domain: "development", Project: "recsys"}

	_, err := svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      orgKey,
		Settings: envSettings("LOG_LEVEL", "info"),
	}))
	require.NoError(t, err)

	_, err = svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      projectKey,
		Settings: envSettings("TEAM", "ml"),
	}))
	require.NoError(t, err)

	resp, err := svc.GetSettings(ctx, connect.NewRequest(&settings.GetSettingsRequest{
		Key: projectKey,
	}))
	require.NoError(t, err)

	record := resp.Msg.GetSettingsRecord()
	assert.True(t, proto.Equal(projectKey, record.GetKey()))

	// Entries from both levels, with no domain row in between.
	assert.Equal(t, map[string]string{"LOG_LEVEL": "info", "TEAM": "ml"},
		record.GetSettings().GetEnvironmentVariables().GetMapValue().GetEntries())
	assert.Equal(t, levelProject, record.GetSettings().GetEnvironmentVariables().GetScopeLevel())

	// Merged settings are not a stored row, so no version comes back even though
	// both stored rows are at version 1.
	assert.Equal(t, uint64(0), record.GetVersion())
}

func TestGetSettings_NothingStored(t *testing.T) {
	ctx := context.Background()
	svc := setupSettingsService(t)

	resp, err := svc.GetSettings(ctx, connect.NewRequest(&settings.GetSettingsRequest{
		Key: &settings.SettingsKey{Org: "acme", Domain: "development", Project: "recsys"},
	}))
	require.NoError(t, err)

	record := resp.Msg.GetSettingsRecord()
	require.NotNil(t, record)
	require.NotNil(t, record.GetSettings())
	assert.Nil(t, record.GetSettings().GetEnvironmentVariables())
	assert.Equal(t, uint64(0), record.GetVersion())
}
