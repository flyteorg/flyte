package otelutils

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger" // nolint:staticcheck
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	rawmetric "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	rawtrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/version"
)

const (
	BlobstoreClientTracer = "blobstore-client"
	K8sClientTracer       = "k8s-client"
)

var tracerProviders = make(map[string]*trace.TracerProvider)
var meterProviders = make(map[string]*sdkmetric.MeterProvider)
var noopTracerProvider = noop.NewTracerProvider()
var noopMeterProvider = metricnoop.NewMeterProvider()

// RegisterProvidersWithContext registers a tracer & metric provider for the given service name.
func RegisterProvidersWithContext(ctx context.Context, serviceName string, config *Config) error {
	if config == nil {
		return nil
	}

	var exporter trace.SpanExporter
	var metricExporter sdkmetric.Exporter
	var err error
	switch config.ExporterType {
	case NoopExporter:
		return nil
	case FileExporter:
		// configure file exporter
		f, err := os.Create(config.FileConfig.Filename)
		if err != nil {
			return err
		}

		exporter, err = stdouttrace.New(
			stdouttrace.WithWriter(f),
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return err
		}

		metricExporter, err = stdoutmetric.New(
			stdoutmetric.WithWriter(f),
			stdoutmetric.WithPrettyPrint(),
		)
		if err != nil {
			return err
		}
	case JaegerExporter:
		// configure jaeger exporter
		exporter, err = jaeger.New(
			jaeger.WithCollectorEndpoint(
				jaeger.WithEndpoint(config.JaegerConfig.Endpoint),
			),
		)
		if err != nil {
			return err
		}
	case OtlpGrpcExporter:
		exporter, err = otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpointURL(config.OtlpGrpcConfig.Endpoint),
		)
		if err != nil {
			return err
		}

		metricExporter, err = otlpmetricgrpc.New(
			ctx,
			otlpmetricgrpc.WithEndpointURL(config.OtlpGrpcConfig.Endpoint),
		)
		if err != nil {
			return err
		}
	case OtlpHttpExporter:
		exporter, err = otlptracehttp.New(
			ctx,
			otlptracehttp.WithEndpointURL(config.OtlpHttpConfig.Endpoint),
		)
		if err != nil {
			return err
		}

		metricExporter, err = otlpmetrichttp.New(
			ctx,
			otlpmetrichttp.WithEndpointURL(config.OtlpHttpConfig.Endpoint),
		)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown otel exporter type [%v]", config.ExporterType)
	}

	telemetryResource := newTelemetryResource(serviceName)

	tracerProvider, err := newTracerProvider(config, exporter, telemetryResource)
	if err != nil {
		return err
	}
	tracerProviders[serviceName] = tracerProvider

	meterProvider := newMeterProvider(metricExporter, telemetryResource)
	meterProviders[serviceName] = meterProvider
	return nil
}

func newTelemetryResource(serviceName string) *resource.Resource {
	// Carry over the attributes from resource.Default() (telemetry.sdk.*,
	// OTEL_RESOURCE_ATTRIBUTES / OTEL_SERVICE_NAME env vars, host info) but
	// stamp them with semconv.SchemaURL instead of calling resource.Merge.
	// Merge panics when the two sides declare different schema URLs, which
	// happens whenever the installed otel/sdk's internal semconv version
	// drifts apart from the semconv version imported above.
	defaultAttrs := resource.Default().Attributes()
	attrs := make([]attribute.KeyValue, 0, len(defaultAttrs)+2)
	attrs = append(attrs, defaultAttrs...)
	attrs = append(attrs,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(version.Version),
	)
	return resource.NewWithAttributes(semconv.SchemaURL, attrs...)
}

func newTracerProvider(config *Config, exporter trace.SpanExporter, telemetryResource *resource.Resource) (*trace.TracerProvider, error) {
	var sampler trace.Sampler
	switch config.SamplerConfig.ParentSampler {
	case AlwaysSample:
		sampler = trace.ParentBased(trace.AlwaysSample())
	case TraceIDRatioBased:
		sampler = trace.ParentBased(trace.TraceIDRatioBased(config.SamplerConfig.TraceIDRatio))
	default:
		return nil, fmt.Errorf("unknown otel sampler type [%v]", config.SamplerConfig.ParentSampler)
	}

	opts := []trace.TracerProviderOption{
		trace.WithBatcher(exporter),
		trace.WithResource(telemetryResource),
		trace.WithSampler(sampler),
	}
	tracerProvider := trace.NewTracerProvider(opts...)
	return tracerProvider, nil
}

func newMeterProvider(exporter sdkmetric.Exporter, telemetryResource *resource.Resource) *sdkmetric.MeterProvider {
	opts := []sdkmetric.Option{
		sdkmetric.WithResource(telemetryResource),
	}
	if exporter != nil {
		opts = append(opts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)))
	}

	return sdkmetric.NewMeterProvider(opts...)
}

// GetTracerProvider returns the tracer provider for the given service name.
func GetTracerProvider(serviceName string) rawtrace.TracerProvider {
	if t, ok := tracerProviders[serviceName]; ok {
		return t
	}

	return noopTracerProvider
}

// GetMeterProvider returns the meter provider for the given service name.
func GetMeterProvider(serviceName string) rawmetric.MeterProvider {
	if m, ok := meterProviders[serviceName]; ok {
		return m
	}

	return noopMeterProvider
}

// NewSpan creates a new span with the given service name and span name.
func NewSpan(ctx context.Context, serviceName string, spanName string) (context.Context, rawtrace.Span) {
	var attributes []attribute.KeyValue
	for key, value := range contextutils.GetLogFields(ctx) {
		if value, ok := value.(string); ok {
			attributes = append(attributes, attribute.String(key, value))
		}
	}

	tracerProvider := GetTracerProvider(serviceName)
	return tracerProvider.Tracer("default").Start(ctx, spanName, rawtrace.WithAttributes(attributes...))
}
