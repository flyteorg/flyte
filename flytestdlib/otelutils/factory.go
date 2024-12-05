package otelutils

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger" // nolint:staticcheck
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	rawtrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/version"
)

const (
	AdminClientTracer        = "admin-client"
	AdminGormTracer          = "admin-gorm"
	AdminServerTracer        = "admin-server"
	BlobstoreClientTracer    = "blobstore-client"
	DataCatalogClientTracer  = "datacatalog-client"
	DataCatalogGormTracer    = "datacatalog-gorm"
	DataCatalogServerTracer  = "datacatalog-server"
	CacheServiceClientTracer = "cacheservice-client"
	CacheServiceGormTracer   = "cacheservice-gorm"
	CacheServiceServerTracer = "cacheservice-server"
	FlytePropellerTracer     = "flytepropeller"
	K8sClientTracer          = "k8s-client"
)

var tracerProviders = make(map[string]*trace.TracerProvider)
var noopTracerProvider = noop.NewTracerProvider()

func RegisterTracerProvider(serviceName string, config *Config) error {
	if config == nil {
		return nil
	}

	var exporter trace.SpanExporter
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
			context.Background(),
			otlptracegrpc.WithEndpointURL(config.OtlpGrpcConfig.Endpoint),
		)
		if err != nil {
			return err
		}
	case OtlpHttpExporter:
		exporter, err = otlptracehttp.New(
			context.Background(),
			otlptracehttp.WithEndpointURL(config.OtlpHttpConfig.Endpoint),
		)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown otel exporter type [%v]", config.ExporterType)
	}

	telemetryResource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version.Version),
		),
	)
	if err != nil {
		return err
	}

	var sampler trace.Sampler
	switch config.SamplerConfig.ParentSampler {
	case AlwaysSample:
		sampler = trace.ParentBased(trace.AlwaysSample())
	case TraceIDRatioBased:
		sampler = trace.ParentBased(trace.TraceIDRatioBased(config.SamplerConfig.TraceIDRatio))
	default:
		return fmt.Errorf("unknown otel sampler type [%v]", config.SamplerConfig.ParentSampler)
	}

	opts := []trace.TracerProviderOption{
		trace.WithBatcher(exporter),
		trace.WithResource(telemetryResource),
		trace.WithSampler(sampler),
	}
	tracerProvider := trace.NewTracerProvider(opts...)
	tracerProviders[serviceName] = tracerProvider
	return nil
}

func GetTracerProvider(serviceName string) rawtrace.TracerProvider {
	if t, ok := tracerProviders[serviceName]; ok {
		return t
	}

	return noopTracerProvider
}

type SpanWrapper struct {
	rawtrace.Span
}

func (s SpanWrapper) EndErr(err error) {
	if err != nil && err != io.EOF {
		s.Span.SetStatus(codes.Error, err.Error())
	}
	s.Span.End()
}

func NewSpan(ctx context.Context, serviceName string, spanName string) (context.Context, SpanWrapper) {
	var attributes []attribute.KeyValue
	for key, value := range contextutils.GetLogFields(ctx) {
		if value, ok := value.(string); ok {
			attributes = append(attributes, attribute.String(key, value))
		}
	}

	tracerProvider := GetTracerProvider(serviceName)
	ctx, span := tracerProvider.Tracer("default").Start(ctx, spanName, rawtrace.WithAttributes(attributes...))
	return ctx, SpanWrapper{Span: span}
}
