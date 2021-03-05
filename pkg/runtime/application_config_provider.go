package runtime

import (
	"context"
	"io/ioutil"
	"os"

	dbconfig "github.com/flyteorg/datacatalog/pkg/repositories/config"
	"github.com/flyteorg/datacatalog/pkg/runtime/configs"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

const database = "database"
const datacatalog = "datacatalog"

var databaseConfig = config.MustRegisterSection(database, &dbconfig.DbConfigSection{})
var datacatalogConfig = config.MustRegisterSection(datacatalog, &configs.DataCatalogConfig{})

// Defines the interface to return top-level config structs necessary to start up a datacatalog application.
type ApplicationConfiguration interface {
	GetDbConfig() dbconfig.DbConfig
	GetDataCatalogConfig() configs.DataCatalogConfig
}

type ApplicationConfigurationProvider struct{}

func (p *ApplicationConfigurationProvider) GetDbConfig() dbconfig.DbConfig {
	dbConfigSection := databaseConfig.GetConfig().(*dbconfig.DbConfigSection)
	password := dbConfigSection.Password
	if len(dbConfigSection.PasswordPath) > 0 {
		if _, err := os.Stat(dbConfigSection.PasswordPath); os.IsNotExist(err) {
			logger.Fatalf(context.Background(),
				"missing database password at specified path [%s]", dbConfigSection.PasswordPath)
		}
		passwordVal, err := ioutil.ReadFile(dbConfigSection.PasswordPath)
		if err != nil {
			logger.Fatalf(context.Background(), "failed to read database password from path [%s] with err: %v",
				dbConfigSection.PasswordPath, err)
		}
		password = string(passwordVal)
	}
	return dbconfig.DbConfig{
		Host:         dbConfigSection.Host,
		Port:         dbConfigSection.Port,
		DbName:       dbConfigSection.DbName,
		User:         dbConfigSection.User,
		Password:     password,
		ExtraOptions: dbConfigSection.ExtraOptions,
	}
}

func (p *ApplicationConfigurationProvider) GetDataCatalogConfig() configs.DataCatalogConfig {
	return *datacatalogConfig.GetConfig().(*configs.DataCatalogConfig)
}

func NewApplicationConfigurationProvider() ApplicationConfiguration {
	return &ApplicationConfigurationProvider{}
}
