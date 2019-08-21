package validators

import (
	"fmt"

	"github.com/lyft/datacatalog/pkg/errors"

	"google.golang.org/grpc/codes"
)

const missingFieldFormat = "missing %s"
const invalidArgFormat = "invalid value for %s, value:[%s]"

func NewMissingArgumentError(field string) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, fmt.Sprintf(missingFieldFormat, field))
}

func NewInvalidArgumentError(field string, value string) error {
	return errors.NewDataCatalogErrorf(codes.InvalidArgument, fmt.Sprintf(invalidArgFormat, field, value))
}
