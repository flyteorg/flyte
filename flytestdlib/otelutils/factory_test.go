package otelutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterTracerProvider(t *testing.T) {
	serviceName := "foo"
	ctx := context.Background()

	// register tracer provider with no exporters
	err := RegisterTracerProvider(ctx, serviceName, defaultConfig)
	assert.Nil(t, err)

	// validate no tracerProviders are registered
	assert.Len(t, tracerProviders, 0)

	// register tracer provider with all exporters
	fullConfig := Config{
		ExporterType: FileExporter,
		FileConfig: FileConfig{
			Filename: "/dev/null",
		},
		JaegerConfig: JaegerConfig{},
	}
	err = RegisterTracerProvider(ctx, serviceName, &fullConfig)
	assert.Nil(t, err)

	// validate tracerProvider is registered
	assert.Len(t, tracerProviders, 1)

	// register tracer provider with otel grpc exporter
	fullConfig = Config{
		ExporterType: OtlpGrpcExporter,
	}

	err = RegisterTracerProvider(ctx, serviceName, &fullConfig)
	assert.Nil(t, err)

	// validate tracerProvider is registered
	assert.Len(t, tracerProviders, 1)

	// register tracer provider with otel http exporter
	fullConfig = Config{
		ExporterType: OtlpHTTPExporter,
	}

	err = RegisterTracerProvider(ctx, serviceName, &fullConfig)
	assert.Nil(t, err)

	// validate tracerProvider is registered
	assert.Len(t, tracerProviders, 1)

}

func TestNewSpan(t *testing.T) {
	ctx := context.TODO()
	_, span := NewSpan(ctx, "bar", "baz/bat")
	span.End()
}
