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
	NoopExporter     ExporterType = "noop"
	FileExporter     ExporterType = "file"
	JaegerExporter   ExporterType = "jaeger"
	OtlpGrpcExporter ExporterType = "otlpgrpc"
	OtlpHttpExporter ExporterType = "otlphttp" // nolint:golint
)

type SamplerType = string

const (
	AlwaysSample      SamplerType = "always"
	TraceIDRatioBased SamplerType = "traceid"
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
		OtlpGrpcConfig: OtlpGrpcConfig{
			Endpoint: "http://localhost:4317",
		},
		OtlpHttpConfig: OtlpHttpConfig{
			Endpoint: "http://localhost:4318/v1/traces",
		},
		SamplerConfig: SamplerConfig{
			ParentSampler: AlwaysSample,
			TraceIDRatio:  0.01,
		},
	}
)

type Config struct {
	ExporterType   ExporterType   `json:"type" pflag:",Sets the type of exporter to configure [noop/file/jaeger/otlpgrpc/otlphttp]."`
	FileConfig     FileConfig     `json:"file" pflag:",Configuration for exporting telemetry traces to a file"`
	JaegerConfig   JaegerConfig   `json:"jaeger" pflag:",Configuration for exporting telemetry traces to a jaeger"`
	OtlpGrpcConfig OtlpGrpcConfig `json:"otlpgrpc" pflag:",Configuration for exporting telemetry traces to an OTLP gRPC collector"`
	OtlpHttpConfig OtlpHttpConfig `json:"otlphttp" pflag:",Configuration for exporting telemetry traces to an OTLP HTTP collector"` // nolint:golint
	SamplerConfig  SamplerConfig  `json:"sampler" pflag:",Configuration for the sampler to use for the tracer"`
}

type FileConfig struct {
	Filename string `json:"filename" pflag:",Filename to store exported telemetry traces"`
}

type JaegerConfig struct {
	Endpoint string `json:"endpoint" pflag:",Endpoint for the jaeger telemetry trace ingestor"`
}

type OtlpGrpcConfig struct {
	Endpoint string `json:"endpoint" pflag:",Endpoint for the OTLP telemetry trace collector"`
}

type OtlpHttpConfig struct { // nolint:golint
	Endpoint string `json:"endpoint" pflag:",Endpoint for the OTLP telemetry trace collector"`
}

type SamplerConfig struct {
	ParentSampler SamplerType `json:"parentSampler" pflag:",Sets the parent sampler to use for the tracer"`
	TraceIDRatio  float64     `json:"traceIdRatio" pflag:"-,Sets the trace id ratio for the TraceIdRatioBased sampler"`
}

func GetConfig() *Config {
	if c, ok := ConfigSection.GetConfig().(*Config); ok {
		return c
	}

	logger.Warnf(context.TODO(), "Failed to retrieve config section [%v].", configSectionKey)
	return nil
}
