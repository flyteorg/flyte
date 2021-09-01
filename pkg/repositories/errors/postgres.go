package errors

import (
	"fmt"

	"github.com/jackc/pgconn"

	"github.com/flyteorg/datacatalog/pkg/errors"
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
	defaultPgError            = "failed database operation with %s"
	unsupportedTableOperation = "cannot query with specified table attributes: %s"
)

func (p *postgresErrorTransformer) fromGormError(err error) error {
	switch err.Error() {
	case gorm.ErrRecordNotFound.Error():
		return errors.NewDataCatalogErrorf(codes.NotFound, "entry not found")
	default:
		return errors.NewDataCatalogErrorf(codes.Internal, unexpectedType, err)
	}
}

func (p *postgresErrorTransformer) ToDataCatalogError(err error) error {
	cErr, ok := err.(ConnectError)
	if !ok {
		return p.fromGormError(err)
	}
	pqError := cErr.Unwrap().(*pgconn.PgError)
	switch pqError.Code {
	case uniqueConstraintViolationCode:
		return errors.NewDataCatalogErrorf(codes.AlreadyExists, uniqueConstraintViolation, pqError.Message)
	case undefinedTable:
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, unsupportedTableOperation, pqError.Message)
	default:
		return errors.NewDataCatalogErrorf(codes.Unknown, fmt.Sprintf(defaultPgError, pqError.Message))
	}
}

func NewPostgresErrorTransformer() ErrorTransformer {
	return &postgresErrorTransformer{}
}

type ConnectError interface {
	Unwrap() error
	Error() string
}
