package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

func TestParseDatabaseConfig(t *testing.T) {
	assert.NoError(t, logger.SetConfig(&logger.Config{IncludeSourceCode: true}))

	accessor := viper.NewAccessor(config.Options{
		RootSection: configSection,
		SearchPaths: []string{filepath.Join("testdata", "config.yaml")},
	})

	assert.NoError(t, accessor.UpdateConfig(context.Background()))

	assert.Equal(t, 100, GetConfig().MaxOpenConnections)
	assert.Equal(t, 10, GetConfig().MaxIdleConnections)
	assert.Equal(t, config.Duration{Duration: 3600000000000}, GetConfig().ConnMaxLifeTime)

	assert.Equal(t, false, GetConfig().Postgres.IsEmpty())
	assert.Equal(t, 5432, GetConfig().Postgres.Port)
	assert.Equal(t, "postgres", GetConfig().Postgres.User)
	assert.Equal(t, "postgres", GetConfig().Postgres.Host)
	assert.Equal(t, "postgres", GetConfig().Postgres.DbName)
	assert.Equal(t, "sslmode=disable", GetConfig().Postgres.ExtraOptions)
	assert.Equal(t, "password", GetConfig().Postgres.Password)
	assert.Equal(t, "/etc/secret", GetConfig().Postgres.PasswordPath)
	assert.Equal(t, true, GetConfig().Postgres.Debug)

}

func TestPostgresConfigValidate(t *testing.T) {
	iam := PostgresConfig{AuthType: AuthTypeIAM, ExtraOptions: "sslmode=require"}
	withPassword := iam
	withPassword.Password = "secret"
	withPasswordPath := iam
	withPasswordPath.PasswordPath = "/etc/db/password"
	withSSLDisabled := iam
	withSSLDisabled.ExtraOptions = "sslmode=disable"
	withSSLDisabledSpaced := iam
	withSSLDisabledSpaced.ExtraOptions = "connect_timeout=5 sslmode = disable"
	withSSLAllow := iam
	withSSLAllow.ExtraOptions = "sslmode=allow"
	withSSLPrefer := iam
	withSSLPrefer.ExtraOptions = "sslmode=prefer"
	withBadOptions := iam
	withBadOptions.ExtraOptions = "sslmode=bogus"

	for _, tc := range []struct {
		name    string
		cfg     PostgresConfig
		wantErr string
	}{
		{name: "empty auth type is password auth", cfg: PostgresConfig{Password: "secret", ExtraOptions: "sslmode=disable"}},
		{name: "password auth", cfg: PostgresConfig{AuthType: AuthTypePassword, PasswordPath: "/etc/db/password"}},
		{name: "iam auth", cfg: iam},
		{name: "iam auth with password", cfg: withPassword, wantErr: "cannot be combined with password or passwordPath"},
		{name: "iam auth with password path", cfg: withPasswordPath, wantErr: "cannot be combined with password or passwordPath"},
		{name: "iam auth with ssl disabled", cfg: withSSLDisabled, wantErr: "requires SSL"},
		{name: "iam auth with ssl disabled and spaces", cfg: withSSLDisabledSpaced, wantErr: "requires SSL"},
		{name: "iam auth with ssl allow", cfg: withSSLAllow, wantErr: "requires SSL"},
		{name: "iam auth with ssl prefer", cfg: withSSLPrefer},
		{name: "iam auth with invalid options", cfg: withBadOptions, wantErr: "parse database postgres options"},
		{name: "unknown auth type", cfg: PostgresConfig{AuthType: "iamm"}, wantErr: `unknown database postgres authType "iamm"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

// IAM config loaded through the real config loader, as the flyte-core chart renders it.
func TestParseIAMDatabaseConfig(t *testing.T) {
	load := func(t *testing.T, file string) PostgresConfig {
		t.Helper()
		// Sections are process-wide; start from the defaults as a fresh process would.
		defaults := *defaultConfig
		assert.NoError(t, configSection.SetConfig(&defaults))
		accessor := viper.NewAccessor(config.Options{
			RootSection: configSection,
			SearchPaths: []string{filepath.Join("testdata", file)},
		})
		assert.NoError(t, accessor.UpdateConfig(context.Background()))
		return GetConfig().Postgres
	}

	t.Run("explicitly empty password", func(t *testing.T) {
		pg := load(t, "iam_config.yaml")
		assert.True(t, pg.IsIAMAuth())
		assert.Empty(t, pg.Password)
		assert.NoError(t, pg.Validate())
	})

	t.Run("omitted password keeps the built-in default and is rejected", func(t *testing.T) {
		pg := load(t, "iam_config_default_password.yaml")
		assert.Equal(t, defaultConfig.Postgres.Password, pg.Password)
		assert.ErrorContains(t, pg.Validate(), "cannot be combined with password or passwordPath")
	})
}
