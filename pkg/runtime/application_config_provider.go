package runtime

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

const database = "database"
const flyteAdmin = "flyteadmin"
const scheduler = "scheduler"
const remoteData = "remoteData"
const notifications = "notifications"
const domains = "domains"
const externalEvents = "externalEvents"

const postgres = "postgres"

const KB = 1024
const MB = KB * KB

var databaseConfig = config.MustRegisterSection(database, &interfaces.DbConfigSection{
	Port:         5432,
	User:         postgres,
	Host:         postgres,
	DbName:       postgres,
	ExtraOptions: "sslmode=disable",
})
var flyteAdminConfig = config.MustRegisterSection(flyteAdmin, &interfaces.ApplicationConfig{
	ProfilerPort:          10254,
	MetricsScope:          "flyte:",
	MetadataStoragePrefix: []string{"metadata", "admin"},
	EventVersion:          2,
	AsyncEventsBufferSize: 100,
	MaxParallelism:        25,
})

var schedulerConfig = config.MustRegisterSection(scheduler, &interfaces.SchedulerConfig{
	ProfilerPort: config.Port{Port: 10253},
	EventSchedulerConfig: interfaces.EventSchedulerConfig{
		Scheme:               common.Local,
		FlyteSchedulerConfig: &interfaces.FlyteSchedulerConfig{},
	},
	WorkflowExecutorConfig: interfaces.WorkflowExecutorConfig{
		Scheme: common.Local,
		FlyteWorkflowExecutorConfig: &interfaces.FlyteWorkflowExecutorConfig{
			AdminRateLimit: &interfaces.AdminRateLimit{
				Tps:   100,
				Burst: 10,
			},
		},
	},
})
var remoteDataConfig = config.MustRegisterSection(remoteData, &interfaces.RemoteDataConfig{
	Scheme:         common.None,
	MaxSizeInBytes: 2 * MB,
})
var notificationsConfig = config.MustRegisterSection(notifications, &interfaces.NotificationsConfig{
	Type: common.Local,
})
var domainsConfig = config.MustRegisterSection(domains, &interfaces.DomainsConfig{
	{
		ID:   "development",
		Name: "development",
	},
	{
		ID:   "staging",
		Name: "staging",
	},
	{
		ID:   "production",
		Name: "production",
	},
})
var externalEventsConfig = config.MustRegisterSection(externalEvents, &interfaces.ExternalEventsConfig{
	Type: common.Local,
})

// Implementation of an interfaces.ApplicationConfiguration
type ApplicationConfigurationProvider struct{}

func (p *ApplicationConfigurationProvider) GetDbConfig() interfaces.DbConfig {
	dbConfigSection := databaseConfig.GetConfig().(*interfaces.DbConfigSection)
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
	return interfaces.DbConfig{
		Host:         dbConfigSection.Host,
		Port:         dbConfigSection.Port,
		DbName:       dbConfigSection.DbName,
		User:         dbConfigSection.User,
		Password:     password,
		ExtraOptions: dbConfigSection.ExtraOptions,
		Debug:        dbConfigSection.Debug,
	}
}

func (p *ApplicationConfigurationProvider) GetTopLevelConfig() *interfaces.ApplicationConfig {
	return flyteAdminConfig.GetConfig().(*interfaces.ApplicationConfig)
}

func (p *ApplicationConfigurationProvider) GetSchedulerConfig() *interfaces.SchedulerConfig {
	return schedulerConfig.GetConfig().(*interfaces.SchedulerConfig)
}

func (p *ApplicationConfigurationProvider) GetRemoteDataConfig() *interfaces.RemoteDataConfig {
	return remoteDataConfig.GetConfig().(*interfaces.RemoteDataConfig)
}

func (p *ApplicationConfigurationProvider) GetNotificationsConfig() *interfaces.NotificationsConfig {
	return notificationsConfig.GetConfig().(*interfaces.NotificationsConfig)
}

func (p *ApplicationConfigurationProvider) GetDomainsConfig() *interfaces.DomainsConfig {
	return domainsConfig.GetConfig().(*interfaces.DomainsConfig)
}

func (p *ApplicationConfigurationProvider) GetExternalEventsConfig() *interfaces.ExternalEventsConfig {
	return externalEventsConfig.GetConfig().(*interfaces.ExternalEventsConfig)
}

func NewApplicationConfigurationProvider() interfaces.ApplicationConfiguration {
	return &ApplicationConfigurationProvider{}
}
