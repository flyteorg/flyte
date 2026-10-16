package retention

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// NewMeterProvider uses the shared otel configuration. V1's otelutils factory
// initializes tracing only, so this one-shot binary owns its metric exporter.
func NewMeterProvider(ctx context.Context, cfg *otelutils.Config) (metric.MeterProvider, func(context.Context) error, error) {
	if cfg == nil || cfg.ExporterType == otelutils.NoopExporter {
		return noop.NewMeterProvider(), func(context.Context) error { return nil }, nil
	}
	var exporter sdkmetric.Exporter
	var file *os.File
	var err error
	switch cfg.ExporterType {
	case otelutils.OtlpGrpcExporter:
		exporter, err = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpointURL(cfg.OtlpGrpcConfig.Endpoint))
	case otelutils.OtlpHttpExporter:
		endpoint, endpointErr := metricsHTTPEndpoint(cfg.OtlpHttpConfig.Endpoint)
		if endpointErr != nil {
			return nil, nil, endpointErr
		}
		exporter, err = otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(endpoint))
	case otelutils.FileExporter:
		file, err = os.Create(cfg.FileConfig.Filename)
		if err == nil {
			exporter, err = stdoutmetric.New(stdoutmetric.WithWriter(file))
		}
	default:
		err = fmt.Errorf("retention metrics require otel.type noop, file, otlpgrpc, or otlphttp; got %q", cfg.ExporterType)
	}
	if err != nil {
		if file != nil {
			_ = file.Close()
		}
		return nil, nil, err
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)), sdkmetric.WithResource(resource.NewWithAttributes("", attribute.String("service.name", "flyteadmin-retention"))))
	return provider, func(ctx context.Context) error {
		err := provider.Shutdown(ctx)
		if file != nil {
			closeErr := file.Close()
			if err == nil {
				err = closeErr
			}
		}
		return err
	}, nil
}

// V1's shared OTEL configuration may explicitly name the trace signal path.
// Retention exports metrics to the corresponding metric signal endpoint.
func metricsHTTPEndpoint(endpoint string) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	if strings.HasSuffix(parsed.Path, "/v1/traces") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/v1/traces") + "/v1/metrics"
	}
	return parsed.String(), nil
}
