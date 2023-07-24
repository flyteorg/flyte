package errors

import (
	admin_errors "github.com/flyteorg/flyteadmin/pkg/errors"
)

// Defines the basic error transformer interface that all database types must implement.
type ErrorTransformer interface {
	ToFlyteAdminError(err error) admin_errors.FlyteAdminError
}
