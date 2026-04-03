package otelutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterTracerProviderWithContext(t *testing.T) {
	ctx := context.Background()
	serviceName := "foo"

	// register tracer provider with no exporters
	err := RegisterTracerProviderWithContext(ctx, serviceName, defaultConfig)
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
		SamplerConfig: SamplerConfig{
			ParentSampler: AlwaysSample,
		},
	}
	err = RegisterTracerProviderWithContext(ctx, serviceName, &fullConfig)
	assert.Nil(t, err)

	// validate tracerProvider is registered
	assert.Len(t, tracerProviders, 1)
}

func TestNewSpan(t *testing.T) {
	ctx := context.TODO()
	_, span := NewSpan(ctx, "bar", "baz/bat")
	span.End()
}
