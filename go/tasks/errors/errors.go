package errors

import (
	"github.com/flyteorg/flytestdlib/errors"
)

const (
	TaskFailedWithError        errors.ErrorCode = "TaskFailedWithError"
	DownstreamSystemError      errors.ErrorCode = "DownstreamSystemError"
	TaskFailedUnknownError     errors.ErrorCode = "TaskFailedUnknownError"
	BadTaskSpecification       errors.ErrorCode = "BadTaskSpecification"
	TaskEventRecordingFailed   errors.ErrorCode = "TaskEventRecordingFailed"
	MetadataAccessFailed       errors.ErrorCode = "MetadataAccessFailed"
	MetadataTooLarge           errors.ErrorCode = "MetadataTooLarge"
	PluginInitializationFailed errors.ErrorCode = "PluginInitializationFailed"
	CacheFailed                errors.ErrorCode = "AutoRefreshCacheFailed"
	RuntimeFailure             errors.ErrorCode = "RuntimeFailure"
	CorruptedPluginState       errors.ErrorCode = "CorruptedPluginState"
	ResourceManagerFailure     errors.ErrorCode = "ResourceManagerFailure"
	BackOffError               errors.ErrorCode = "BackOffError"
)

func Errorf(errorCode errors.ErrorCode, msgFmt string, args ...interface{}) error {
	return errors.Errorf(errorCode, msgFmt, args...)
}

func Wrapf(errorCode errors.ErrorCode, err error, msgFmt string, args ...interface{}) error {
	return errors.Wrapf(errorCode, err, msgFmt, args...)
}
