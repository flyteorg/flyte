package runtime

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/flyteorg/flytestdlib/config"

	"github.com/flyteorg/datacatalog/pkg/runtime/configs"
	"github.com/flyteorg/flytestdlib/database"
	"github.com/flyteorg/flytestdlib/logger"
)

const datacatalog = "datacatalog"

var datacatalogConfig = config.MustRegisterSection(datacatalog, &configs.DataCatalogConfig{})

// Defines the interface to return top-level config structs necessary to start up a datacatalog application.
type ApplicationConfiguration interface {
	GetDbConfig() *database.DbConfig
	GetDataCatalogConfig() configs.DataCatalogConfig
}

type ApplicationConfigurationProvider struct{}

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

func (p *ApplicationConfigurationProvider) GetDataCatalogConfig() configs.DataCatalogConfig {
	return *datacatalogConfig.GetConfig().(*configs.DataCatalogConfig)
}

func NewApplicationConfigurationProvider() ApplicationConfiguration {
	return &ApplicationConfigurationProvider{}
}
