package config

import "github.com/flyteorg/flytestdlib/config"

//go:generate pflags ConsoleConfig --default-var DefaultConsoleConfig --bind-default-var

var (
	DefaultConsoleConfig = &ConsoleConfig{}

	cfg = config.MustRegisterSection("console", DefaultConsoleConfig)
)

// FilesConfig containing flags used for registration
type ConsoleConfig struct {
	Endpoint string `json:"endpoint" pflag:",Endpoint of console, if different than flyte admin"`
}

func GetConfig() *ConsoleConfig {
	return cfg.GetConfig().(*ConsoleConfig)
}
