package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
)

// TestResolveSettings resolves a chain without going through a connect handler,
// which is what run creation will do in phase 3.
func TestResolveSettings(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testDB.Exec("DELETE FROM settings") })

	repo := impl.NewSettingsRepo(testDB)
	svc := NewSettingsService(repo)

	_, err := svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      &settings.SettingsKey{Org: "acme"},
		Settings: envSettings("LOG_LEVEL", "info"),
	}))
	require.NoError(t, err)

	_, err = svc.CreateSettings(ctx, connect.NewRequest(&settings.CreateSettingsRequest{
		Key:      &settings.SettingsKey{Org: "acme", Domain: "production", Project: "analytics"},
		Settings: envSettings("LOG_LEVEL", "debug"),
	}))
	require.NoError(t, err)

	resolved, err := resolveSettings(ctx, repo, &settings.SettingsKey{
		Org: "acme", Domain: "production", Project: "analytics",
	})
	require.NoError(t, err)

	want := &settings.StringMapSetting{
		State:      stateValue,
		MapValue:   &settings.StringMap{Entries: map[string]string{"LOG_LEVEL": "debug"}},
		ScopeLevel: settings.ScopeLevel_SCOPE_LEVEL_PROJECT,
	}

	require.True(t, proto.Equal(want, resolved.GetEnvironmentVariables()),
		"got %s", protojson.Format(resolved.GetEnvironmentVariables()))
}
