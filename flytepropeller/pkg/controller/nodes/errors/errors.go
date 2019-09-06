package errors

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type ErrorMessage = string

type NodeError struct {
	errors.StackTrace
	Code    ErrorCode
	Message ErrorMessage
	Node    v1alpha1.NodeID
}

func (n *NodeError) Error() string {
	return fmt.Sprintf("failed at Node[%s]. %v: %v", n.Node, n.Code, n.Message)
}

type NodeErrorWithCause struct {
	*NodeError
	cause error
}

func (n *NodeErrorWithCause) Error() string {
	return fmt.Sprintf("%v, caused by: %v", n.NodeError.Error(), errors.Cause(n))
}

func (n *NodeErrorWithCause) Cause() error {
	return n.cause
}

func errorf(c ErrorCode, n v1alpha1.NodeID, msgFmt string, args ...interface{}) *NodeError {
	return &NodeError{
		Code:    c,
		Node:    n,
		Message: fmt.Sprintf(msgFmt, args...),
	}
}

func Errorf(c ErrorCode, n v1alpha1.NodeID, msgFmt string, args ...interface{}) error {
	return errorf(c, n, msgFmt, args...)
}

func Wrapf(c ErrorCode, n v1alpha1.NodeID, cause error, msgFmt string, args ...interface{}) error {
	return &NodeErrorWithCause{
		NodeError: errorf(c, n, msgFmt, args...),
		cause:     cause,
	}
}

func Matches(err error, code ErrorCode) bool {
	errCode, isNodeError := GetErrorCode(err)
	if isNodeError {
		return code == errCode
	}
	return false
}

func GetErrorCode(err error) (code ErrorCode, isNodeError bool) {
	isNodeError = false
	e, ok := err.(*NodeError)
	if ok {
		code = e.Code
		isNodeError = true
		return
	}

	e2, ok := err.(*NodeErrorWithCause)
	if ok {
		code = e2.Code
		isNodeError = true
		return
	}
	return
}
