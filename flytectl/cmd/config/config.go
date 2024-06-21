package config

import (
	"fmt"
	"strings"

	"github.com/flyteorg/flyte/flytectl/pkg/printer"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

var (
	defaultConfig = &Config{
		Output: printer.OutputFormatTABLE.String(),
	}

	section = config.MustRegisterSection("root", defaultConfig)
)

// Config hold configuration for flytectl flag
type Config struct {
	Project     string `json:"project" pflag:",Specifies the project to work on."`
	Domain      string `json:"domain" pflag:",Specifies the domain to work on."`
	Output      string `json:"output" pflag:",Specifies the output type."`
	Interactive bool   `json:"interactive" pflag:",Set this to trigger bubbletea interface."`
}

// OutputFormat will return output format
func (cfg Config) OutputFormat() (printer.OutputFormat, error) {
	return printer.OutputFormatString(strings.ToUpper(cfg.Output))
}

// MustOutputFormat will validate the supported output format and return output format
func (cfg Config) MustOutputFormat() printer.OutputFormat {
	f, err := cfg.OutputFormat()
	if err != nil {
		panic(fmt.Sprintf("unsupported output format [%s], supported types %s", cfg.Output, printer.OutputFormats()))
	}
	return f
}

// GetConfig will return the config
func GetConfig() *Config {
	return section.GetConfig().(*Config)
}
