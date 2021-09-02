package errors

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/jackc/pgconn"

	catalogErrors "github.com/flyteorg/datacatalog/pkg/errors"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
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
		return catalogErrors.NewDataCatalogErrorf(codes.NotFound, "entry not found")
	default:
		return catalogErrors.NewDataCatalogErrorf(codes.Internal, unexpectedType, err)
	}
}

func (p *postgresErrorTransformer) ToDataCatalogError(err error) error {
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
		return catalogErrors.NewDataCatalogErrorf(codes.AlreadyExists, uniqueConstraintViolation, pqError.Message)
	case undefinedTable:
		return catalogErrors.NewDataCatalogErrorf(codes.InvalidArgument, unsupportedTableOperation, pqError.Message)
	default:
		return catalogErrors.NewDataCatalogErrorf(codes.Unknown, fmt.Sprintf(defaultPgError, pqError.Code, pqError.Message))
	}
}

func NewPostgresErrorTransformer() ErrorTransformer {
	return &postgresErrorTransformer{}
}

type ConnectError interface {
	Unwrap() error
	Error() string
}
