package catalog

import (
	"github.com/lyft/flytestdlib/config"
)

//go:generate pflags Config --default-var defaultConfig

const ConfigSectionKey = "catalog-cache"

var (
	defaultConfig = &Config{
		Type: NoOpDiscoveryType,
	}

	configSection = config.MustRegisterSection(ConfigSectionKey, defaultConfig)
)

type DiscoveryType = string

const (
	NoOpDiscoveryType   DiscoveryType = "noop"
	LegacyDiscoveryType DiscoveryType = "legacy"
	DataCatalogType     DiscoveryType = "datacatalog"
)

type Config struct {
	Type     DiscoveryType `json:"type" pflag:"\"noop\", Catalog Implementation to use"`
	Endpoint string        `json:"endpoint" pflag:"\"\", Endpoint for catalog service"`
	Insecure bool          `json:"insecure" pflag:"false, Use insecure grpc connection"`
}

// Gets loaded config for Discovery
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
