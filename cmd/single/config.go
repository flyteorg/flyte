package single

import (
	adminRepositoriesConfig "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/config"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

//go:generate pflags Config --default-var=DefaultConfig

var (
	DefaultConfig = &Config{}
	configSection = config.MustRegisterSection("flyte", DefaultConfig)
)

type Config struct {
	Propeller   Propeller   `json:"propeller" pflag:",Configuration to disable propeller or any of its components."`
	Admin       Admin       `json:"admin" pflag:",Configuration to disable FlyteAdmin or any of its components"`
	DataCatalog DataCatalog `json:"dataCatalog" pflag:",Configuration to disable DataCatalog or any of its components"`
}

type Propeller struct {
	Disabled       bool `json:"disabled" pflag:",Disables flytepropeller in the single binary mode"`
	DisableWebhook bool `json:"disableWebhook" pflag:",Disables webhook only"`
}

type Admin struct {
	Disabled                      bool                                  `json:"disabled" pflag:",Disables flyteadmin in the single binary mode"`
	DisableScheduler              bool                                  `json:"disableScheduler" pflag:",Disables Native scheduler in the single binary mode"`
	DisableClusterResourceManager bool                                  `json:"disableClusterResourceManager" pflag:",Disables Cluster resource manager"`
	SeedProjects                  []string                              `json:"seedProjects" pflag:",flyte projects to create by default."`
	SeedProjectsWithDetails       []adminRepositoriesConfig.SeedProject `json:"seedProjectsWithDetails" pflag:",,Detailed configuration for Flyte projects to be created by default."`
}

type DataCatalog struct {
	Disabled bool `json:"disabled" pflag:",Disables datacatalog in the single binary mode"`
}

// GetConfig returns a handle to the configuration for Flyte Single Binary
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
