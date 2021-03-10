package secretmanager

import "github.com/flyteorg/flytestdlib/config"

//go:generate pflags Config --default-var defaultConfig

const SectionKey = "secrets"

var (
	defaultConfig = &Config{
		SecretFilePrefix:  "/etc/secrets",
		EnvironmentPrefix: "FLYTE_SECRET_",
	}

	section = config.MustRegisterSection(SectionKey, defaultConfig)
)

type Config struct {
	SecretFilePrefix  string `json:"secrets-prefix" pflag:", Prefix where to look for secrets file"`
	EnvironmentPrefix string `json:"env-prefix" pflag:", Prefix for environment variables"`
}

func GetConfig() *Config {
	return section.GetConfig().(*Config)
}
