package configuration

import (
	"time"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration/shared"
	"github.com/flyteorg/flyte/flytestdlib/config"
	stdLibDb "github.com/flyteorg/flyte/flytestdlib/database"
	stdLibStorage "github.com/flyteorg/flyte/flytestdlib/storage"
)

const artifactsServer = "artifactsServer"

type ApplicationConfiguration struct {
	ArtifactDatabaseConfig  stdLibDb.DbConfig          `json:"artifactDatabaseConfig" pflag:",Database configuration"`
	ArtifactBlobStoreConfig stdLibStorage.Config       `json:"artifactBlobStoreConfig" pflag:",Blob store configuration"`
	ArtifactServerConfig    shared.ServerConfiguration `json:"artifactServerConfig" pflag:",Artifact server configuration"`
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
	ArtifactBlobStoreConfig: stdLibStorage.Config{
		InitContainer: "flyte-artifacts",
	},
	ArtifactServerConfig: shared.ServerConfiguration{
		Metrics: shared.Metrics{
			MetricsScope:    "service:",
			Port:            config.Port{Port: 10254},
			ProfilerEnabled: false,
		},
		Port:                       config.Port{Port: 50051},
		HttpPort:                   config.Port{Port: 50050},
		GrpcMaxResponseStatusBytes: 320000,
		GrpcServerReflection:       false,
		Security: shared.ServerSecurityOptions{
			Secure:  false,
			UseAuth: false,
		},
		MaxConcurrentStreams: 100,
	},
}

var ApplicationConfig = config.MustRegisterSection(artifactsServer, &defaultApplicationConfiguration)

func GetApplicationConfig() *ApplicationConfiguration {
	return ApplicationConfig.GetConfig().(*ApplicationConfiguration)
}
