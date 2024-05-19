package events

import (
	"context"
	"strconv"

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
	MaxRetries    int                `json:"max-retries" pflag:",The max number of retries for event recording."`
	BackoffScalar int                `json:"base-scalar" pflag:",The base/scalar backoff duration in milliseconds for event recording retries."`
	BackoffJitter string             `json:"backoff-jitter" pflag:",A string representation of a floating point number between 0 and 1 specifying the jitter factor for event recording retries."`
}

var (
	defaultConfig = Config{
		Rate:          int64(500),
		Capacity:      1000,
		Type:          EventSinkAdmin,
		MaxRetries:    5,
		BackoffScalar: 100,
		BackoffJitter: "0.1",
	}

	configSection = config.MustRegisterSectionWithUpdates(configSectionKey, &defaultConfig, func(ctx context.Context, newValue config.Config) {
		if newValue.(*Config).MaxRetries < 0 {
			logger.Panicf(ctx, "Admin configuration given with negative gRPC retry value.")
		}

		if jitter, err := strconv.ParseFloat(newValue.(*Config).BackoffJitter, 64); err != nil || jitter < 0 || jitter > 1 {
			logger.Panicf(ctx, "Invalid jitter value [%v]. Must be between 0 and 1.", jitter)
		}
	})
)

func (c Config) GetBackoffJitter(ctx context.Context) float64 {
	jitter, err := strconv.ParseFloat(c.BackoffJitter, 64)
	if err != nil {
		logger.Warnf(ctx, "Failed to parse backoff jitter [%v]. Error: %v", c.BackoffJitter, err)
		return 0.1
	}

	return jitter
}

// GetConfig Retrieves current global config for storage.
func GetConfig(ctx context.Context) *Config {
	if c, ok := configSection.GetConfig().(*Config); ok {
		return c
	}

	logger.Warnf(ctx, "Failed to retrieve config section [%v].", configSectionKey)
	return nil
}
