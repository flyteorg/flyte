package launchplan

import (
	errors2 "github.com/flyteorg/flytestdlib/errors"
)

type ErrorCode = errors2.ErrorCode

const (
	RemoteErrorAlreadyExists ErrorCode = "AlreadyExists"
	RemoteErrorNotFound      ErrorCode = "NotFound"
	RemoteErrorSystem        ErrorCode = "SystemError" // timeouts, network error etc
	RemoteErrorUser          ErrorCode = "UserError"   // Incase of bad specification, invalid arguments, etc
)

// Checks if the error is of type RemoteError and the ErrorCode is of type RemoteErrorAlreadyExists
func IsAlreadyExists(err error) bool {
	return errors2.IsCausedBy(err, RemoteErrorAlreadyExists)
}

// Checks if the error is of type RemoteError and the ErrorCode is of type RemoteErrorUser
func IsUserError(err error) bool {
	return errors2.IsCausedBy(err, RemoteErrorUser)
}

// Checks if the error is of type RemoteError and the ErrorCode is of type RemoteErrorNotFound
func IsNotFound(err error) bool {
	return errors2.IsCausedBy(err, RemoteErrorNotFound)
}
