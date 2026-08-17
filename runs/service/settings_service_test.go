package service

import (
	"testing"

	"github.com/flyteorg/flyte/v2/runs/repository/impl"
)

func setupSettingsService(t *testing.T) *SettingsService {
	t.Cleanup(func() { testDB.Exec("DELETE FROM settings") })
	return NewSettingsService(impl.NewSettingsRepo(testDB))
}

func envSettings(key, value string) *settings.Settings {

}
