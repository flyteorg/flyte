package validators

import (
	"fmt"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/datacatalog/pkg/common"
	"github.com/flyteorg/flyte/datacatalog/pkg/errors"
)

const missingFieldFormat = "missing %s"
const invalidArgFormat = "invalid value for %s, value:[%s]"
const invalidFilterFormat = "%s cannot be filtered by %s properties"

func NewMissingArgumentError(field string) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, "%s", fmt.Sprintf(missingFieldFormat, field))
}

func NewInvalidArgumentError(field string, value string) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, "%s", fmt.Sprintf(invalidArgFormat, field, value))
}

func NewInvalidFilterError(entity common.Entity, propertyEntity common.Entity) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, "%s", fmt.Sprintf(invalidFilterFormat, entity, propertyEntity))
}
