package errors

import "github.com/flyteorg/flyte/flytestdlib/errors"

type ErrorCode = errors.ErrorCode

const (
	UnknownError                       ErrorCode = "UnknownError"
	DownstreamNodeNotFoundError        ErrorCode = "DownstreamNodeNotFound"
	UserProvidedError                  ErrorCode = "UserProvidedError"
	IllegalStateError                  ErrorCode = "IllegalStateError"
	BadSpecificationError              ErrorCode = "BadSpecificationError"
	UnsupportedTaskTypeError           ErrorCode = "UnsupportedTaskType"
	BindingResolutionError             ErrorCode = "BindingResolutionError"
	CausedByError                      ErrorCode = "CausedByError"
	RuntimeExecutionError              ErrorCode = "RuntimeExecutionError"
	SubWorkflowExecutionFailed         ErrorCode = "SubWorkflowExecutionFailed"
	RemoteChildWorkflowExecutionFailed ErrorCode = "RemoteChildWorkflowExecutionFailed"
	OutputsNotFoundError               ErrorCode = "OutputsNotFoundError"
	InputsNotFoundError                ErrorCode = "InputsNotFoundError"
	StorageError                       ErrorCode = "StorageError"
	EventRecordingFailed               ErrorCode = "EventRecordingFailed"
	InvalidArrayLength                 ErrorCode = "InvalidArrayLength"
	PromiseAttributeResolveError       ErrorCode = "PromiseAttributeResolveError"
	IDLNotFoundErr                     ErrorCode = "IDLNotFoundErr"
	InvalidPrimitiveType               ErrorCode = "InvalidPrimitiveType"
)
