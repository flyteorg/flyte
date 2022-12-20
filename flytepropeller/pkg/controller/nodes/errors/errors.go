package errors

import (
	"fmt"
	"strings"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type ErrorMessage = string

type NodeError struct {
	ErrCode ErrorCode
	Message ErrorMessage
	Node    v1alpha1.NodeID
}

func (n *NodeError) Code() ErrorCode {
	return n.ErrCode
}

func (n *NodeError) Error() string {
	return fmt.Sprintf("failed at Node[%s]. %v: %v", n.Node, n.ErrCode, n.Message)
}

func (n *NodeError) Is(target error) bool {
	t, ok := target.(*NodeError)
	if !ok {
		return false
	}
	if n == nil && t == nil {
		return true
	}
	if n == nil || t == nil {
		return false
	}
	return n.ErrCode == t.ErrCode && (n.Message == t.Message || t.Message == "") && (n.Node == t.Node || t.Node == "")
}

type NodeErrorWithCause struct {
	NodeError error
	cause     error
}

func (n *NodeErrorWithCause) Cause() error {
	return n.cause
}

func (n *NodeErrorWithCause) Code() ErrorCode {
	if asNodeErr, casted := n.NodeError.(*NodeError); casted {
		return asNodeErr.Code()
	}

	return ""
}

func (n *NodeErrorWithCause) Error() string {
	nodeError := ""
	if n.NodeError != nil {
		nodeError = n.NodeError.Error()
	}

	cause := ""
	if n.cause != nil {
		cause = n.cause.Error()
	}

	return fmt.Sprintf("%v, caused by: %v", nodeError, cause)
}

func (n *NodeErrorWithCause) Is(target error) bool {
	t, ok := target.(*NodeErrorWithCause)
	if !ok {
		return false
	}
	if n == nil && t == nil {
		return true
	}
	if n == nil || t == nil {
		return false
	}
	return n.Is(target) && (n.cause == t.cause || t.cause == nil)
}

func (n *NodeErrorWithCause) Unwrap() error {
	return n.cause
}

func errorf(c ErrorCode, n v1alpha1.NodeID, msgFmt string, args ...interface{}) error {
	return &NodeError{
		ErrCode: c,
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
		code = e.ErrCode
		isNodeError = true
		return
	}

	if e2, ok := err.(*NodeErrorWithCause); ok {
		if ne, ok := e2.NodeError.(*NodeError); ok {
			code = ne.ErrCode
			isNodeError = true
			return
		}
	}

	return
}

type ErrorCollection struct {
	Errors []error
}

func (e ErrorCollection) Error() string {
	sb := strings.Builder{}
	for idx, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("%v: %v\r\n", idx, err))
	}

	return sb.String()
}
