package runtime

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/flyteorg/flyte/cacheservice/pkg/runtime/configs"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const cacheservice = "cacheservice"

var cacheserviceConfig = config.MustRegisterSection(cacheservice, &configs.CacheServiceConfig{})

// ApplicationConfiguration defines the interface to return top-level config structs necessary to start up a cacheservice application.
type ApplicationConfiguration interface {
	GetCacheServiceConfig() configs.CacheServiceConfig
	GetDbConfig() *database.DbConfig
}

type ApplicationConfigurationProvider struct{}

func (p *ApplicationConfigurationProvider) GetCacheServiceConfig() configs.CacheServiceConfig {
	return *cacheserviceConfig.GetConfig().(*configs.CacheServiceConfig)
}

func (p *ApplicationConfigurationProvider) GetDbConfig() *database.DbConfig {
	dbConfigSection := database.GetConfig()
	if len(dbConfigSection.Postgres.PasswordPath) > 0 {
		if _, err := os.Stat(dbConfigSection.Postgres.PasswordPath); os.IsNotExist(err) {
			logger.Fatalf(context.Background(),
				"missing database password at specified path [%s]", dbConfigSection.Postgres.PasswordPath)
		}
		passwordVal, err := ioutil.ReadFile(dbConfigSection.Postgres.PasswordPath)
		if err != nil {
			logger.Fatalf(context.Background(), "failed to read database password from path [%s] with err: %v",
				dbConfigSection.Postgres.PasswordPath, err)
		}
		// Passwords can contain special characters as long as they are percent encoded
		// https://www.postgresql.org/docs/current/libpq-connect.html
		dbConfigSection.Postgres.Password = strings.TrimSpace(string(passwordVal))
	}

	return dbConfigSection
}

func NewApplicationConfigurationProvider() ApplicationConfiguration {
	return &ApplicationConfigurationProvider{}
}
