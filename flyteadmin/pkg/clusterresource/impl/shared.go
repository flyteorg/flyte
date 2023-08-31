package impl

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"
)

func NewMissingEntityError(entity string) error {
	return errors.NewFlyteAdminErrorf(codes.NotFound, "Failed to find [%s]", entity)
}
