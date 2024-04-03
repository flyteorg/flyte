package events

import (
	"context"

	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
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
	Type          EventReportingType `json:"type" pflag:",Sets the type of EventSink to configure [log/admin/file]."`
	FilePath      string             `json:"file-path" pflag:",For file types, specify where the file should be located."`
	Rate          int64              `json:"rate" pflag:",Max rate at which events can be recorded per second."`
	Capacity      int                `json:"capacity" pflag:",The max bucket size for event recording tokens."`
	MaxRetries    uint               `json:"max-retries" pflag:",The max number of retries for event recording."`
	BackoffScalar int                `json:"base-scalar" pflag:",The base/scalar backoff duration in milliseconds for event recording retries."`
	BackoffJitter float64            `json:"backoff-jitter" pflag:",The jitter factor for event recording retries."`
}

var (
	defaultConfig = Config{
		Rate:          int64(500),
		Capacity:      1000,
		Type:          EventSinkAdmin,
		MaxRetries:    5,
		BackoffScalar: 100,
		BackoffJitter: 0.1,
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
