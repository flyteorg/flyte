package database

import (
	"context"
	"testing"

	"github.com/XSAM/otelsql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestPostgresTelemetry(t *testing.T) {
	ctx := context.Background()
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, tp.Shutdown(ctx)) })
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, mp.Shutdown(ctx)) })
	db, err := openPostgres("host=127.0.0.1 port=1 dbname=telemetry_test user=test password=secret sslmode=disable connect_timeout=1",
		otelsql.WithTracerProvider(tp), otelsql.WithMeterProvider(mp))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	assert.Equal(t, "pgx", db.DriverName())
	assert.Equal(t, "SELECT $1", db.Rebind("SELECT ?"))

	parentCtx, parent := tp.Tracer("test").Start(ctx, "parent")
	// Exercise pgx connection errors without requiring a running database.
	require.Error(t, db.PingContext(parentCtx))
	parent.End()
	var found bool
	for _, span := range recorder.Ended() {
		if span.Name() != "sql.connector.connect" {
			continue
		}
		found = true
		assert.Equal(t, parent.SpanContext().SpanID(), span.Parent().SpanID())
		assert.Equal(t, codes.Error, span.Status().Code)
		assert.Contains(t, span.Attributes(), attribute.String("db.system.name", "postgresql"))
		assert.Contains(t, span.Attributes(), attribute.String("db.namespace", "telemetry_test"))
		for _, attr := range span.Attributes() {
			assert.NotContains(t, attr.Value.AsString(), "secret")
		}
	}
	require.True(t, found, "expected a connection span")
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &data))
	names := metricNames(data)
	assert.Contains(t, names, "db.client.operation.duration")
	assert.Contains(t, names, "db.sql.connection.max_open")
	require.NoError(t, db.Close())
	require.NoError(t, reader.Collect(ctx, &data))
	assert.NotContains(t, metricNames(data), "db.sql.connection.max_open", "closed pools must unregister metrics")
}

func metricNames(data metricdata.ResourceMetrics) []string {
	var names []string
	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			names = append(names, m.Name)
		}
	}
	return names
}

func TestOpenPostgresInvalidDSN(t *testing.T) {
	db, err := openPostgres("invalid dsn")
	require.Error(t, err)
	assert.Nil(t, db)
}
