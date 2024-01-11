package otelutils

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	rawtrace "go.opentelemetry.io/otel/trace"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/version"
)

const (
	AdminClientTracer       = "admin-client"
	AdminGormTracer         = "admin-gorm"
	AdminServerTracer       = "admin-server"
	BlobstoreClientTracer   = "blobstore-client"
	DataCatalogClientTracer = "datacatalog-client"
	DataCatalogGormTracer   = "datacatalog-gorm"
	DataCatalogServerTracer = "datacatalog-server"
	FlytePropellerTracer    = "flytepropeller"
	K8sClientTracer         = "k8s-client"
)

var (
	fileExporter       trace.SpanExporter
	noopTracerProvider = rawtrace.NewNoopTracerProvider()
	tracerProviders    = make(map[string]*trace.TracerProvider)
)

func RegisterTracerProvider(serviceName string, config *Config) error {
	if config == nil {
		return nil
	}

	var opts []trace.TracerProviderOption
	switch config.ExporterType {
	case NoopExporter:
		return nil
	case FileExporter:
		if fileExporter == nil {
			// initialize file exporter and reuse between all services
			f, err := os.Create(config.FileConfig.Filename)
			if err != nil {
				return err
			}

			var spanFormatter spanFormatter
			switch config.FileConfig.SpanFormat {
			case OtelSpanFormat:
				spanFormatter = NewOTELSpanFormatter()
			case OtelCollectorSpanFormat:
				spanFormatter, err = NewOTELCollectorSpanFormatter()
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown span format [%v]", config.FileConfig.SpanFormat)
			}

			fileExporter, err = NewExporter(spanFormatter, f)
			if err != nil {
				return err
			}
		}

		opts = append(opts, trace.WithBatcher(fileExporter))
	case JaegerExporter:
		// configure jaeger exporter
		exporter, err := jaeger.New(
			jaeger.WithCollectorEndpoint(
				jaeger.WithEndpoint(config.JaegerConfig.Endpoint),
			),
		)
		if err != nil {
			return err
		}

		opts = append(opts, trace.WithBatcher(exporter))
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

	opts = append(opts, trace.WithResource(telemetryResource))
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
