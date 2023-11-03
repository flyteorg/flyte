package otelutils

import (
	"context"

	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

//go:generate pflags Config --default-var=defaultConfig

type Type = string

const configSectionKey = "otel"

var (
	ConfigSection = config.MustRegisterSection(configSectionKey, defaultConfig)
	defaultConfig = &Config{
		FileConfig: FileConfig{
			Enabled:  false,
			Filename: "/tmp/trace.txt",
		},
		JaegerConfig: JaegerConfig{
			Enabled:  false,
			Endpoint: "http://localhost:14268/api/traces",
		},
	}
)

type Config struct {
	FileConfig   FileConfig   `json:"file" pflag:",Configuration for exporting telemetry traces to a file"`
	JaegerConfig JaegerConfig `json:"jaeger" pflag:",Configuration for exporting telemetry traces to a jaeger"`
}

type FileConfig struct {
	Enabled  bool   `json:"enabled" pflag:",Set to true to enable the file exporter"`
	Filename string `json:"filename" pflag:",Filename to store exported telemetry traces"`
}

type JaegerConfig struct {
	Enabled  bool   `json:"enabled" pflag:",Set to true to enable the jaeger exporter"`
	Endpoint string `json:"endpoint" pflag:",Endpoint for the jaeger telemtry trace ingestor"`
}

func GetConfig() *Config {
	if c, ok := ConfigSection.GetConfig().(*Config); ok {
		return c
	}

	logger.Warnf(context.TODO(), "Failed to retrieve config section [%v].", configSectionKey)
	return nil
}
