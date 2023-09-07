// Postgres-specific implementation of an ErrorTransformer.
// This errors utility translates postgres application error codes into internal error types.
// The go postgres driver defines possible error codes here: https://github.com/lib/pq/blob/master/error.go
// And the postgres standard defines error responses here:
//
//	https://www.postgresql.org/docs/current/static/protocol-error-fields.html
//
// Inspired by https://www.codementor.io/tamizhvendan/managing-data-in-golang-using-gorm-part-1-a9cdjb8nb
package errors

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/jackc/pgconn"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

// Postgres error codes
const (
	uniqueConstraintViolationCode = "23505"
	undefinedTable                = "42P01"
)

// Error message format strings
const (
	unexpectedType            = "unexpected error type for: %v"
	uniqueConstraintViolation = "value with matching already exists (%s)"
	defaultPgError            = "failed database operation with %s"
	unsupportedTableOperation = "cannot query with specified table attributes: %s"
)

type postgresErrorTransformerMetrics struct {
	Scope              promutils.Scope
	NotFound           prometheus.Counter
	GormError          prometheus.Counter
	AlreadyExistsError prometheus.Counter
	UndefinedTable     prometheus.Counter
	PostgresError      prometheus.Counter
}

type postgresErrorTransformer struct {
	metrics postgresErrorTransformerMetrics
}

func (p *postgresErrorTransformer) fromGormError(err error) flyteAdminErrors.FlyteAdminError {
	switch err.Error() {
	case gorm.ErrRecordNotFound.Error():
		p.metrics.NotFound.Inc()
		return flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "entry not found")
		// If we want to intercept other gorm errors, add additional case statements here.
	default:
		p.metrics.GormError.Inc()
		return flyteAdminErrors.NewFlyteAdminErrorf(codes.Internal, unexpectedType, err)
	}
}

func (p *postgresErrorTransformer) ToFlyteAdminError(err error) flyteAdminErrors.FlyteAdminError {
	if unwrappedErr := errors.Unwrap(err); unwrappedErr != nil {
		err = unwrappedErr
	}

	pqError, ok := err.(*pgconn.PgError)
	if !ok {
		logger.Debugf(context.Background(), "Unable to cast to pgconn.PgError. Error type: [%v]",
			reflect.TypeOf(err))
		return p.fromGormError(err)
	}

	switch pqError.Code {
	case uniqueConstraintViolationCode:
		p.metrics.AlreadyExistsError.Inc()
		return flyteAdminErrors.NewFlyteAdminErrorf(codes.AlreadyExists, uniqueConstraintViolation, pqError.Message)
	case undefinedTable:
		p.metrics.UndefinedTable.Inc()
		return flyteAdminErrors.NewFlyteAdminErrorf(codes.InvalidArgument, unsupportedTableOperation, pqError.Message)
	default:
		p.metrics.PostgresError.Inc()
		return flyteAdminErrors.NewFlyteAdminError(codes.Unknown, fmt.Sprintf(defaultPgError, pqError.Message))
	}
}

func NewPostgresErrorTransformer(scope promutils.Scope) ErrorTransformer {
	metrics := postgresErrorTransformerMetrics{
		Scope: scope,
		NotFound: scope.MustNewCounter("not_found",
			"count of all queries for entities not found in the database"),
		GormError: scope.MustNewCounter("gorm_error",
			"unspecified gorm error returned by database operation"),
		AlreadyExistsError: scope.MustNewCounter("already_exists",
			"counts for when a unique constraint was violated in a database operation"),
		UndefinedTable: scope.MustNewCounter("undefined_table",
			"database operations referencing an undefined table"),
		PostgresError: scope.MustNewCounter("postgres_error",
			"unspecified postgres error returned in a database operation"),
	}
	return &postgresErrorTransformer{
		metrics: metrics,
	}
}
