package otelutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterTracerProvider(t *testing.T) {
	serviceName := "foo"

	// register tracer provider with no exporters
	err := RegisterTracerProvider(serviceName, defaultConfig)
	assert.Nil(t, err)

	// validate no tracerProviders are registered
	assert.Len(t, tracerProviders, 0)

	// register tracer provider with all exporters
	fullConfig := Config{
		ExporterType: FileExporter,
		FileConfig: FileConfig{
			Filename: "/dev/null",
		},
		JaegerConfig: JaegerConfig{
		},
	}
	err = RegisterTracerProvider(serviceName, &fullConfig)
	assert.Nil(t, err)

	// validate tracerProvider is registered
	assert.Len(t, tracerProviders, 1)
}

func TestNewSpan(t *testing.T) {
	ctx := context.TODO()
	_, span := NewSpan(ctx, "bar", "baz/bat")
	span.End()
}
