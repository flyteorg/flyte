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
	err := RegisterProvidersWithContext(ctx, serviceName, defaultConfig)
	assert.Nil(t, err)

	// validate no tracerProviders are registered
	assert.Len(t, tracerProviders, 0)
	assert.Len(t, meterProviders, 0)

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
	err = RegisterProvidersWithContext(ctx, serviceName, &fullConfig)
	assert.Nil(t, err)

	// validate tracerProvider is registered
	assert.Len(t, tracerProviders, 1)
	assert.Len(t, meterProviders, 1)
}

func TestNewSpan(t *testing.T) {
	ctx := context.TODO()
	_, span := NewSpan(ctx, "bar", "baz/bat")
	span.End()
}

func TestRegisterProvidersWithContextReusesProviders(t *testing.T) {
	ctx := context.Background()
	const serviceName = "ensure-providers-test"
	cfg := Config{
		ExporterType:  FileExporter,
		FileConfig:    FileConfig{Filename: t.TempDir() + "/telemetry.json"},
		SamplerConfig: SamplerConfig{ParentSampler: AlwaysSample},
	}
	t.Cleanup(func() {
		if tp := tracerProviders[serviceName]; tp != nil {
			assert.NoError(t, tp.Shutdown(ctx))
		}
		if mp := meterProviders[serviceName]; mp != nil {
			assert.NoError(t, mp.Shutdown(ctx))
		}
		delete(tracerProviders, serviceName)
		delete(meterProviders, serviceName)
	})
	if !assert.NoError(t, RegisterProvidersWithContext(ctx, serviceName, &cfg)) {
		return
	}
	tp, mp := GetTracerProvider(serviceName), GetMeterProvider(serviceName)
	assert.NoError(t, RegisterProvidersWithContext(ctx, serviceName, &cfg))
	assert.Same(t, tp, GetTracerProvider(serviceName))
	assert.Same(t, mp, GetMeterProvider(serviceName))
}
