package config

import (
	"fmt"
	"strings"

	"github.com/lyft/flytestdlib/config"

	"github.com/lyft/flytectl/printer"
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

func (cfg Config) OutputFormat() (printer.OutputFormat, error) {
	return printer.OutputFormatString(strings.ToUpper(cfg.Output))
}

func (cfg Config) MustOutputFormat() printer.OutputFormat {
	f, err := cfg.OutputFormat()
	if err != nil {
		panic(fmt.Sprintf("unsupported output format [%s], supported types %s", cfg.Output, printer.OutputFormats()))
	}
	return f
}

func GetConfig() *Config {
	return section.GetConfig().(*Config)
}
