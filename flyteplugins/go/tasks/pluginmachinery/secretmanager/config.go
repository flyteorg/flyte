package secretmanager

import "github.com/flyteorg/flyte/v2/flytestdlib/config"

//go:generate pflags Config --default-var defaultConfig

type Type = string

const SectionKey = "secrets"

var (
	defaultConfig = &Config{
		Type:              "local",
		SecretFilePrefix:  "/etc/secrets",
		EnvironmentPrefix: "FLYTE_SECRET_",
	}

	section = config.MustRegisterSection(SectionKey, defaultConfig)
)

type Config struct {
	Type              Type   `json:"type" pflag:",Sets the type of storage to configure [local]."`
	SecretFilePrefix  string `json:"secrets-prefix" pflag:", Prefix where to look for secrets file"`
	EnvironmentPrefix string `json:"env-prefix" pflag:", Prefix for environment variables"`
}

func GetConfig() *Config {
	return section.GetConfig().(*Config)
}
