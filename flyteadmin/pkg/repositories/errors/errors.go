// Generic errors used across files in repositories/
package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
)

const (
	singletonNotFound = "missing singleton entity of type %s"
	notFound          = "missing entity of type %s with identifier %v"
	idNotFound        = "missing entity of type %s"
	invalidInput      = "missing and/or invalid parameters: %s"
)

func GetMissingEntityError(entityType string, identifier proto.Message) errors.FlyteAdminError {
	return errors.NewFlyteAdminErrorf(codes.NotFound, notFound, entityType, identifier)
}

func GetSingletonMissingEntityError(entityType string) errors.FlyteAdminError {
	return errors.NewFlyteAdminErrorf(codes.NotFound, singletonNotFound, entityType)
}

func GetMissingEntityByIDError(entityType string) errors.FlyteAdminError {
	return errors.NewFlyteAdminErrorf(codes.NotFound, idNotFound, entityType)
}

func GetInvalidInputError(input string) errors.FlyteAdminError {
	return errors.NewFlyteAdminErrorf(codes.InvalidArgument, invalidInput, input)
}
