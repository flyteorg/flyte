package launchplan

import "fmt"

type ErrorCode string

const (
	RemoteErrorAlreadyExists ErrorCode = "AlreadyExists"
	RemoteErrorNotFound      ErrorCode = "NotFound"
	RemoteErrorSystem                  = "SystemError" // timeouts, network error etc
	RemoteErrorUser                    = "UserError"   // Incase of bad specification, invalid arguments, etc
)

type RemoteError struct {
	Code    ErrorCode
	Cause   error
	Message string
}

func (r RemoteError) Error() string {
	return fmt.Sprintf("%s: %s, caused by [%s]", r.Code, r.Message, r.Cause.Error())
}

func Wrapf(code ErrorCode, cause error, msg string, args ...interface{}) error {
	return &RemoteError{
		Code:    code,
		Cause:   cause,
		Message: fmt.Sprintf(msg, args...),
	}
}

// Checks if the error is of type RemoteError and the ErrorCode is of type RemoteErrorAlreadyExists
func IsAlreadyExists(err error) bool {
	e, ok := err.(*RemoteError)
	if ok {
		return e.Code == RemoteErrorAlreadyExists
	}
	return false
}

// Checks if the error is of type RemoteError and the ErrorCode is of type RemoteErrorUser
func IsUserError(err error) bool {
	e, ok := err.(*RemoteError)
	if ok {
		return e.Code == RemoteErrorUser
	}
	return false
}

// Checks if the error is of type RemoteError and the ErrorCode is of type RemoteErrorNotFound
func IsNotFound(err error) bool {
	e, ok := err.(*RemoteError)
	if ok {
		return e.Code == RemoteErrorNotFound
	}
	return false
}
