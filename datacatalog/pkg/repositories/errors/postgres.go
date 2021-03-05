package errors

import (
	"fmt"

	"github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
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
	uniqueConstraintViolation = "value with matching %s already exists (%s)"
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
	pqError, ok := err.(*pq.Error)
	if !ok {
		return p.fromGormError(err)
	}
	switch pqError.Code {
	case uniqueConstraintViolationCode:
		return errors.NewDataCatalogErrorf(codes.AlreadyExists, uniqueConstraintViolation, pqError.Constraint, pqError.Message)
	case undefinedTable:
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, unsupportedTableOperation, pqError.Message)
	default:
		return errors.NewDataCatalogErrorf(codes.Unknown, fmt.Sprintf(defaultPgError, pqError.Message))
	}
}

func NewPostgresErrorTransformer() ErrorTransformer {
	return &postgresErrorTransformer{}
}
