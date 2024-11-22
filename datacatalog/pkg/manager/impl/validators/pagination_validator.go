package validators

import (
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/datacatalog/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

// The token is a string that should be opaque to the client
// It represents the offset as an integer encoded as a string,
// but in the future it can be a string that encodes anything
func ValidateToken(token string) error {
	// if the token is empty, that is still valid input since it is optional
	if len(strings.Trim(token, " ")) == 0 {
		return nil
	}
	_, err := strconv.ParseUint(token, 10, 32)
	if err != nil {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid token value: %s", token)
	}
	return nil
}

// Validate the pagination options and set default limits
func ValidatePagination(options *datacatalog.PaginationOptions) error {
	err := ValidateToken(options.GetToken())
	if err != nil {
		return err
	}

	if options.GetSortKey() != datacatalog.PaginationOptions_CREATION_TIME {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid sort key %v", options.GetSortKey())
	}

	if options.GetSortOrder() != datacatalog.PaginationOptions_ASCENDING &&
		options.GetSortOrder() != datacatalog.PaginationOptions_DESCENDING {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid sort order %v", options.GetSortOrder())
	}

	return nil
}
