package errors

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type ErrorMessage = string

type WorkflowError struct {
	errors.StackTrace
	Code     ErrorCode
	Message  ErrorMessage
	Workflow v1alpha1.WorkflowID
}

func (w *WorkflowError) Error() string {
	return fmt.Sprintf("Workflow[%s] failed. %v: %v", w.Workflow, w.Code, w.Message)
}

func (w *WorkflowError) Is(target error) bool {
	t, ok := target.(*WorkflowError)
	if !ok {
		return false
	}
	if w == nil && t == nil {
		return true
	}
	if w == nil || t == nil {
		return false
	}
	return w.Code == t.Code
}

type WorkflowErrorWithCause struct {
	*WorkflowError
	cause error
}

func (w *WorkflowErrorWithCause) Cause() error {
	return w.cause
}

func (w *WorkflowErrorWithCause) Error() string {
	return fmt.Sprintf("%v, caused by: %v", w.WorkflowError.Error(), w.cause)
}

func (w *WorkflowErrorWithCause) Is(target error) bool {
	t, ok := target.(*WorkflowErrorWithCause)
	if !ok {
		return false
	}
	if w == nil && t == nil {
		return true
	}
	if w == nil || t == nil {
		return false
	}
	return w.Code == t.Code && (w.cause == t.cause || t.cause == nil) && (w.Message == t.Message || t.Message == "") && (w.Workflow == t.Workflow || t.Workflow == "")
}

func (w *WorkflowErrorWithCause) Unwrap() error {
	return w.cause
}

func errorf(c ErrorCode, w v1alpha1.WorkflowID, msgFmt string, args ...interface{}) *WorkflowError {
	return &WorkflowError{
		Code:     c,
		Workflow: w,
		Message:  fmt.Sprintf(msgFmt, args...),
	}
}

func Errorf(c ErrorCode, w v1alpha1.WorkflowID, msgFmt string, args ...interface{}) error {
	return errorf(c, w, msgFmt, args...)
}

func Wrapf(c ErrorCode, w v1alpha1.WorkflowID, cause error, msgFmt string, args ...interface{}) error {
	return &WorkflowErrorWithCause{
		WorkflowError: errorf(c, w, msgFmt, args...),
		cause:         cause,
	}
}

func Matches(err error, code ErrorCode) bool {
	errCode, isWorkflowError := GetErrorCode(err)
	if isWorkflowError {
		return code == errCode
	}
	return false
}

func GetErrorCode(err error) (code ErrorCode, isWorkflowError bool) {
	isWorkflowError = false
	e, ok := err.(*WorkflowError)
	if ok {
		code = e.Code
		isWorkflowError = true
		return
	}

	e2, ok := err.(*WorkflowErrorWithCause)
	if ok {
		code = e2.Code
		isWorkflowError = true
		return
	}
	return
}
