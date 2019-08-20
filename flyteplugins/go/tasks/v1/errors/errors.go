package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

type ErrorCode = string

const (
	TaskFailedWithError        ErrorCode = "TaskFailedWithError"
	DownstreamSystemError      ErrorCode = "DownstreamSystemError"
	TaskFailedUnknownError     ErrorCode = "TaskFailedUnknownError"
	BadTaskSpecification       ErrorCode = "BadTaskSpecification"
	TaskEventRecordingFailed   ErrorCode = "TaskEventRecordingFailed"
	MetadataAccessFailed       ErrorCode = "MetadataAccessFailed"
	MetadataTooLarge           ErrorCode = "MetadataTooLarge"
	PluginInitializationFailed ErrorCode = "PluginInitializationFailed"
	CacheFailed                ErrorCode = "AutoRefreshCacheFailed"
	RuntimeFailure             ErrorCode = "RuntimeFailure"
)

type TaskError struct {
	Code    string
	Message string
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("task failed, %v: %v", e.Code, e.Message)
}

type TaskErrorWithCause struct {
	*TaskError
	cause error
}

func (e *TaskErrorWithCause) Error() string {
	return fmt.Sprintf("%v, caused by: %v", e.TaskError.Error(), errors.Cause(e))
}

func (e *TaskErrorWithCause) Cause() error {
	return e.cause
}

func Errorf(errorCode ErrorCode, msgFmt string, args ...interface{}) *TaskError {
	return &TaskError{
		Code:    errorCode,
		Message: fmt.Sprintf(msgFmt, args...),
	}
}

func Wrapf(errorCode ErrorCode, err error, msgFmt string, args ...interface{}) *TaskErrorWithCause {
	return &TaskErrorWithCause{
		TaskError: Errorf(errorCode, msgFmt, args...),
		cause:     err,
	}
}

func GetErrorCode(err error) (code ErrorCode, isTaskError bool) {
	isTaskError = false
	e, ok := err.(*TaskError)
	if ok {
		code = e.Code
		isTaskError = true
		return
	}

	e2, ok := err.(*TaskError)
	if ok {
		code = e2.Code
		isTaskError = true
		return
	}
	return
}
