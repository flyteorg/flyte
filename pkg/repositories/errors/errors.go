// Generic errors used across files in repositories/
package errors

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
)

const (
	notFound     = "missing entity of type %s with identifier %v"
	idNotFound   = "missing entity of type %s"
	invalidInput = "missing and/or invalid parameters: %s"
)

func GetMissingEntityError(entityType string, identifier proto.Message) errors.FlyteAdminError {
	return errors.NewFlyteAdminErrorf(codes.NotFound, notFound, entityType, identifier)
}

func GetMissingEntityByIDError(entityType string) errors.FlyteAdminError {
	return errors.NewFlyteAdminErrorf(codes.NotFound, idNotFound, entityType)
}

func GetInvalidInputError(input string) errors.FlyteAdminError {
	return errors.NewFlyteAdminErrorf(codes.InvalidArgument, invalidInput, input)
}
