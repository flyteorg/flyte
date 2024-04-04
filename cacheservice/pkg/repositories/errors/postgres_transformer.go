package errors

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/jackc/pgconn"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"

	cacheErr "github.com/flyteorg/flyte/cacheservice/pkg/errors"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// Postgres error codes
const (
	uniqueConstraintViolationCode = "23505"
	undefinedTable                = "42P01"
)

type postgresErrorTransformer struct {
}

const (
	unexpectedType            = "unexpected error type for: %v"
	uniqueConstraintViolation = "value with matching already exists (%s)"
	defaultPgError            = "failed database operation with code [%s] and msg [%s]"
	unsupportedTableOperation = "cannot query with specified table attributes: %s"
)

func (p *postgresErrorTransformer) fromGormError(err error) error {
	switch err.Error() {
	case gorm.ErrRecordNotFound.Error():
		return cacheErr.NewCacheServiceErrorf(codes.NotFound, "entry not found")
	default:
		return cacheErr.NewCacheServiceErrorf(codes.Internal, unexpectedType, err)
	}
}

func (p *postgresErrorTransformer) ToCacheServiceError(err error) error {
	// First try the stdlib error handling
	if database.IsPgErrorWithCode(err, uniqueConstraintViolationCode) {
		return cacheErr.NewCacheServiceErrorf(codes.AlreadyExists, uniqueConstraintViolation, err.Error())
	}

	if unwrappedErr := errors.Unwrap(err); unwrappedErr != nil {
		err = unwrappedErr
	}

	pqError, ok := err.(*pgconn.PgError)
	if !ok {
		logger.InfofNoCtx("Unable to cast to pgconn.PgError. Error type: [%v]",
			reflect.TypeOf(err))
		return p.fromGormError(err)
	}

	switch pqError.Code {
	case uniqueConstraintViolationCode:
		return cacheErr.NewCacheServiceErrorf(codes.AlreadyExists, uniqueConstraintViolation, pqError.Message)
	case undefinedTable:
		return cacheErr.NewCacheServiceErrorf(codes.InvalidArgument, unsupportedTableOperation, pqError.Message)
	default:
		return cacheErr.NewCacheServiceErrorf(codes.Unknown, fmt.Sprintf(defaultPgError, pqError.Code, pqError.Message))
	}
}

func NewPostgresErrorTransformer() ErrorTransformer {
	return &postgresErrorTransformer{}
}
