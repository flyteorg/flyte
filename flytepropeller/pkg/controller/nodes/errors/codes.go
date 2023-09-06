package errors

import "github.com/flyteorg/flytestdlib/errors"

type ErrorCode = errors.ErrorCode

const (
	UnknownError                       ErrorCode = "UnknownError"
	InitializationError                ErrorCode = "InitializationError"
	NotYetImplementedError             ErrorCode = "NotYetImplementedError"
	DownstreamNodeNotFoundError        ErrorCode = "DownstreamNodeNotFound"
	UserProvidedError                  ErrorCode = "UserProvidedError"
	IllegalStateError                  ErrorCode = "IllegalStateError"
	BadSpecificationError              ErrorCode = "BadSpecificationError"
	UnsupportedTaskTypeError           ErrorCode = "UnsupportedTaskType"
	BindingResolutionError             ErrorCode = "BindingResolutionError"
	CausedByError                      ErrorCode = "CausedByError"
	RuntimeExecutionError              ErrorCode = "RuntimeExecutionError"
	SubWorkflowExecutionFailed         ErrorCode = "SubWorkflowExecutionFailed"
	SubWorkflowExecutionFailing        ErrorCode = "SubWorkflowExecutionFailing"
	RemoteChildWorkflowExecutionFailed ErrorCode = "RemoteChildWorkflowExecutionFailed"
	NoBranchTakenError                 ErrorCode = "NoBranchTakenError"
	OutputsNotFoundError               ErrorCode = "OutputsNotFoundError"
	InputsNotFoundError                ErrorCode = "InputsNotFoundError"
	StorageError                       ErrorCode = "StorageError"
	EventRecordingFailed               ErrorCode = "EventRecordingFailed"
	CatalogCallFailed                  ErrorCode = "CatalogCallFailed"
	InvalidArrayLength                 ErrorCode = "InvalidArrayLength"
)
