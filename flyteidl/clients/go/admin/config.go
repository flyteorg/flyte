package admin

import (
	"context"
	"time"

	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/logger"
)

//go:generate pflags Config

const configSectionKey = "admin"

type Config struct {
	Endpoint              config.URL      `json:"endpoint" pflag:",For admin types, specify where the uri of the service is located."`
	UseInsecureConnection bool            `json:"insecure" pflag:",Use insecure connection."`
	MaxBackoffDelay       config.Duration `json:"maxBackoffDelay" pflag:",Max delay for grpc backoff"`
	PerRetryTimeout       config.Duration `json:"perRetryTimeout" pflag:",gRPC per retry timeout"`
	MaxRetries            int             `json:"maxRetries" pflag:",Max number of gRPC retries"`
}

var (
	defaultConfig = Config{
		MaxBackoffDelay: config.Duration{8 * time.Second},
		PerRetryTimeout: config.Duration{15 * time.Second},
		MaxRetries:      4,
	}
	configSection = config.MustRegisterSectionWithUpdates(configSectionKey, &defaultConfig, func(ctx context.Context, newValue config.Config) {
		if newValue.(*Config).MaxRetries < 0 {
			logger.Panicf(ctx, "Admin configuration given with negative gRPC retry value.")
		}
	})
)

func GetConfig(ctx context.Context) *Config {
	if c, ok := configSection.GetConfig().(*Config); ok {
		return c
	}

	logger.Warnf(ctx, "Failed to retrieve config section [%v].", configSectionKey)
	return nil
}
