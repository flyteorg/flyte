package errors

type ErrorCode string

const (
	IllegalStateError     ErrorCode = "IllegalStateError"
	BadSpecificationError ErrorCode = "BadSpecificationError"
	CausedByError         ErrorCode = "CausedByError"
	RuntimeExecutionError ErrorCode = "RuntimeExecutionError"
	EventRecordingError   ErrorCode = "ErrorRecordingError"
)

func (e ErrorCode) String() string {
	return string(e)
}
