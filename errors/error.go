// Contains utilities to use to create and consume simple errors.
package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

// A generic error code type.
type ErrorCode = string

type err struct {
	code    ErrorCode
	message string
}

func (e *err) Error() string {
	return fmt.Sprintf("[%v] %v", e.code, e.message)
}

func (e *err) Code() ErrorCode {
	return e.code
}

type errorWithCause struct {
	*err
	cause error
}

func (e *errorWithCause) Error() string {
	return fmt.Sprintf("%v, caused by: %v", e.err.Error(), errors.Cause(e))
}

func (e *errorWithCause) Cause() error {
	return e.cause
}

// Creates a new error using an error code and a message.
func Errorf(errorCode ErrorCode, msgFmt string, args ...interface{}) error {
	return &err{
		code:    errorCode,
		message: fmt.Sprintf(msgFmt, args...),
	}
}

// Wraps a root cause error with another. This is useful to unify an error type in a package.
func Wrapf(code ErrorCode, cause error, msgFmt string, args ...interface{}) error {
	return &errorWithCause{
		err: &err{
			code:    code,
			message: fmt.Sprintf(msgFmt, args...),
		},
		cause: cause,
	}
}

// Gets the error code of the passed error if it has one.
func GetErrorCode(e error) (code ErrorCode, found bool) {
	type coder interface {
		Code() ErrorCode
	}

	er, ok := e.(coder)
	if ok {
		return er.Code(), true
	}

	return
}

// Gets whether error is caused by another error with errCode.
func IsCausedBy(e error, errCode ErrorCode) bool {
	type causer interface {
		Cause() error
	}

	for e != nil {
		if code, found := GetErrorCode(e); found && code == errCode {
			return true
		}

		cause, ok := e.(causer)
		if !ok {
			break
		}

		e = cause.Cause()
	}

	return false
}

func IsCausedByError(e, e2 error) bool {
	type causer interface {
		Cause() error
	}

	for e != nil {
		if e == e2 {
			return true
		}

		cause, ok := e.(causer)
		if !ok {
			break
		}

		e = cause.Cause()
	}

	return false
}
