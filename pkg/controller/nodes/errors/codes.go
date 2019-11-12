package errors

type ErrorCode string

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
	RemoteChildWorkflowExecutionFailed ErrorCode = "RemoteChildWorkflowExecutionFailed"
	NoBranchTakenError                 ErrorCode = "NoBranchTakenError"
	OutputsNotFoundError               ErrorCode = "OutputsNotFoundError"
	StorageError                       ErrorCode = "StorageError"
	EventRecordingFailed               ErrorCode = "EventRecordingFailed"
	CatalogCallFailed                  ErrorCode = "CatalogCallFailed"
)

func (e ErrorCode) String() string {
	return string(e)
}
