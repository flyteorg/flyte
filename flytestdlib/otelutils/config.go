package otelutils

import (
	"context"

	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

//go:generate pflags Config --default-var=defaultConfig

const configSectionKey = "otel"

type ExporterType = string

const (
	NoopExporter   ExporterType = "noop"
	FileExporter   ExporterType = "file"
	JaegerExporter ExporterType = "jaeger"
)

var (
	ConfigSection = config.MustRegisterSection(configSectionKey, defaultConfig)
	defaultConfig = &Config{
		ExporterType: NoopExporter,
		FileConfig: FileConfig{
			Filename: "/tmp/trace.txt",
		},
		JaegerConfig: JaegerConfig{
			Endpoint: "http://localhost:14268/api/traces",
		},
	}
)

type Config struct {
	ExporterType ExporterType `json:"type" pflag:",Sets the type of exporter to configure [noop/file/jaeger]."`
	FileConfig   FileConfig   `json:"file" pflag:",Configuration for exporting telemetry traces to a file"`
	JaegerConfig JaegerConfig `json:"jaeger" pflag:",Configuration for exporting telemetry traces to a jaeger"`
}

type FileConfig struct {
	Filename string `json:"filename" pflag:",Filename to store exported telemetry traces"`
}

type JaegerConfig struct {
	Endpoint string `json:"endpoint" pflag:",Endpoint for the jaeger telemetry trace ingestor"`
}

func GetConfig() *Config {
	if c, ok := ConfigSection.GetConfig().(*Config); ok {
		return c
	}

	logger.Warnf(context.TODO(), "Failed to retrieve config section [%v].", configSectionKey)
	return nil
}
