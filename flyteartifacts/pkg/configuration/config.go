package configuration

import (
	"github.com/flyteorg/flyte/flytestdlib/config"
	stdLibDb "github.com/flyteorg/flyte/flytestdlib/database"
	"time"
)

const artifactsServer = "artifactsServer"

type ApplicationConfiguration struct {
	ArtifactDatabaseConfig stdLibDb.DbConfig `json:"artifactDatabaseConfig" pflag:",Database configuration"`
}

var defaultApplicationConfiguration = ApplicationConfiguration{
	ArtifactDatabaseConfig: stdLibDb.DbConfig{
		EnableForeignKeyConstraintWhenMigrating: true,
		MaxIdleConnections:                      10,
		MaxOpenConnections:                      100,
		ConnMaxLifeTime:                         config.Duration{Duration: time.Hour},
		Postgres: stdLibDb.PostgresConfig{
			// These values are suitable for local sandbox development
			Host:         "localhost",
			Port:         30001,
			DbName:       "artifacts",
			User:         "postgres",
			Password:     "postgres",
			ExtraOptions: "sslmode=disable",
		},
	},
}

var ApplicationConfig = config.MustRegisterSection(artifactsServer, &defaultApplicationConfiguration)
