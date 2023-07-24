package logger

import (
	"context"

	"github.com/flyteorg/flytestdlib/config"
)

//go:generate pflags Config --default-var defaultConfig

const configSectionKey = "Logger"

type FormatterType = string

const (
	FormatterJSON FormatterType = "json"
	FormatterText FormatterType = "text"
	FormatterGCP  FormatterType = "gcp"
)

const (
	jsonDataKey string = "json"
)

var (
	defaultConfig = &Config{
		Formatter: FormatterConfig{
			Type: FormatterJSON,
		},
		Level: WarnLevel,
	}

	configSection = config.MustRegisterSectionWithUpdates(configSectionKey, defaultConfig, func(ctx context.Context, newValue config.Config) {
		onConfigUpdated(*newValue.(*Config))
	})
)

// Global logger config.
type Config struct {
	// Determines whether to include source code location in logs. This might incurs a performance hit and is only
	// recommended on debug/development builds.
	IncludeSourceCode bool `json:"show-source" pflag:",Includes source code location in logs."`

	// Determines whether the logger should mute all logs (including panics)
	Mute bool `json:"mute" pflag:",Mutes all logs regardless of severity. Intended for benchmarks/tests only."`

	// Determines the minimum log level to log.
	Level Level `json:"level" pflag:",Sets the minimum logging level."`

	Formatter FormatterConfig `json:"formatter" pflag:",Sets logging format."`
}

type FormatterConfig struct {
	Type FormatterType `json:"type" pflag:",Sets logging format type."`
}

// Sets global logger config
func SetConfig(cfg *Config) error {
	if err := configSection.SetConfig(cfg); err != nil {
		return err
	}

	onConfigUpdated(*cfg)
	return nil
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

// Level type.
type Level = int

// These are the different logging levels.
const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)
