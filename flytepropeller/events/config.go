package events

import (
	"context"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

//go:generate pflags Config

type EventReportingType = string

// The reserved config section key for storage.
const configSectionKey = "Event"

const (
	EventSinkLog   EventReportingType = "log"
	EventSinkFile  EventReportingType = "file"
	EventSinkAdmin EventReportingType = "admin"
)

type Config struct {
	Type     EventReportingType `json:"type" pflag:",Sets the type of EventSink to configure [log/admin/file]."`
	FilePath string             `json:"file-path" pflag:",For file types, specify where the file should be located."`
	Rate     int64              `json:"rate" pflag:",Max rate at which events can be recorded per second."`
	Capacity int                `json:"capacity" pflag:",The max bucket size for event recording tokens."`
}

var (
	defaultConfig = Config{
		Rate:     int64(500),
		Capacity: 1000,
		Type:     EventSinkAdmin,
	}

	configSection = config.MustRegisterSection(configSectionKey, &defaultConfig)
)

// GetConfig Retrieves current global config for storage.
func GetConfig(ctx context.Context) *Config {
	if c, ok := configSection.GetConfig().(*Config); ok {
		return c
	}

	logger.Warnf(ctx, "Failed to retrieve config section [%v].", configSectionKey)
	return nil
}
