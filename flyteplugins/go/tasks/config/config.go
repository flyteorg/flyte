package config

import (
	"github.com/flyteorg/flytestdlib/config"
)

const configSectionKey = "plugins"

var (
	// Root config section. If you are a plugin developer and your plugin needs a config, you should register
	// your config as a subsection for this root section.
	rootSection = config.MustRegisterSection(configSectionKey, &Config{})
)

// Top level plugins config.
type Config struct {
}

// Retrieves the current config value or default.
func GetConfig() *Config {
	return rootSection.GetConfig().(*Config)
}

func MustRegisterSubSection(subSectionKey string, section config.Config) config.Section {
	return rootSection.MustRegisterSection(subSectionKey, section)
}
