package config

import (
	"github.com/lyft/flytestdlib/config"
)

//go:generate pflags Config

var (
	defaultConfig = &Config{}
	section       = config.MustRegisterSection("root", defaultConfig)
)

type Config struct {
	Project string `json:"project" pflag:",Specifies the project to work on."`
	Domain  string `json:"domain" pflag:",Specified the domain to work on."`
	Output  string `json:"output" pflag:",Specified the output type."`
}

func GetConfig() *Config {
	return section.GetConfig().(*Config)
}
