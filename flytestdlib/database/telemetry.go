package database

import (
	"database/sql/driver"
	"io"

	"github.com/XSAM/otelsql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
)

// TelemetryOptions uses Flyte's configured providers instead of OTel's global
// providers. Register the service's providers before opening the database.
func TelemetryOptions(serviceName string) []otelsql.Option {
	return []otelsql.Option{
		otelsql.WithTracerProvider(otelutils.GetTracerProvider(serviceName)),
		otelsql.WithMeterProvider(otelutils.GetMeterProvider(serviceName)),
	}
}

// metricsConnector ties the pool metrics callback to sql.DB.Close, including
// temporary connections used during database creation and failed connection attempts.
type metricsConnector struct {
	driver.Connector
	registration metric.Registration
}

func (c *metricsConnector) Close() error {
	var err error
	if c.registration != nil {
		err = c.registration.Unregister()
	}
	if closer, ok := c.Connector.(io.Closer); ok {
		if closeErr := closer.Close(); closeErr != nil {
			return closeErr
		}
	}
	return err
}

func openPostgres(dsn string, options ...otelsql.Option) (*sqlx.DB, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	opts := []otelsql.Option{
		otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL, semconv.DBNamespace(cfg.Database)),
	}
	opts = append(opts, options...)
	connector := &metricsConnector{Connector: stdlib.GetConnector(*cfg)}
	db := otelsql.OpenDB(connector, opts...)
	connector.registration, err = otelsql.RegisterDBStatsMetrics(db, opts...)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	// Preserve pgx's bind type for sqlx named queries and Rebind.
	return sqlx.NewDb(db, "pgx"), nil
}
