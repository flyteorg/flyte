package validators

import (
	"strconv"
	"strings"

	"github.com/lyft/datacatalog/pkg/errors"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"google.golang.org/grpc/codes"
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
func ValidatePagination(options datacatalog.PaginationOptions) error {
	err := ValidateToken(options.Token)
	if err != nil {
		return err
	}

	if options.SortKey != datacatalog.PaginationOptions_CREATION_TIME {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid sort key %v", options.SortKey)
	}

	if options.SortOrder != datacatalog.PaginationOptions_ASCENDING &&
		options.SortOrder != datacatalog.PaginationOptions_DESCENDING {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Invalid sort order %v", options.SortOrder)
	}

	return nil
}
