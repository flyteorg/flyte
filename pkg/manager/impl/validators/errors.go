package validators

import (
	"fmt"

	"github.com/flyteorg/datacatalog/pkg/errors"

	"github.com/flyteorg/datacatalog/pkg/common"
	"google.golang.org/grpc/codes"
)

const missingFieldFormat = "missing %s"
const invalidArgFormat = "invalid value for %s, value:[%s]"
const invalidFilterFormat = "%s cannot be filtered by %s properties"

func NewMissingArgumentError(field string) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, fmt.Sprintf(missingFieldFormat, field))
}

func NewInvalidArgumentError(field string, value string) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, fmt.Sprintf(invalidArgFormat, field, value))
}

func NewInvalidFilterError(entity common.Entity, propertyEntity common.Entity) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, fmt.Sprintf(invalidFilterFormat, entity, propertyEntity))
}
