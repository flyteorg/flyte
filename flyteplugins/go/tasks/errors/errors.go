package errors

import (
	"fmt"
	"github.com/flyteorg/flyte/flytestdlib/errors"
)

// Error codes
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

// Errorf wraps the error message with a given error code.
func Errorf(errorCode errors.ErrorCode, msgFmt string, args ...interface{}) error {
	return errors.Errorf(errorCode, msgFmt, args...)
}

// Wrapf wraps an existing error with additional context and a given error code.
func Wrapf(errorCode errors.ErrorCode, err error, msgFmt string, args ...interface{}) error {
	return errors.Wrapf(errorCode, err, msgFmt, args...)
}

// Task represents a task that can be updated.
type Task struct {
	ID     string
	Status string
	Meta   string // Example metadata for the task
}

// UpdateTask updates the task's status and metadata, handling errors during the process.
func UpdateTask(task *Task, newStatus string, newMeta string) error {
	// Validate input
	if task == nil {
		return Errorf(BadTaskSpecification, "Task cannot be nil")
	}
	if newStatus == "" {
		return Errorf(BadTaskSpecification, "New status cannot be empty")
	}

	// Simulate task update process
	fmt.Printf("Updating task %s with new status: %s\n", task.ID, newStatus)

	// Simulating potential failure in metadata access
	if len(newMeta) > 100 { // Assuming metadata size limit is 100 characters
		return Errorf(MetadataTooLarge, "Metadata size exceeds limit for task %s", task.ID)
	}

	// Simulate a downstream system error
	if task.ID == "task-123" { // Example condition where a specific task might fail
		err := fmt.Errorf("failed to communicate with downstream system")
		return Wrapf(DownstreamSystemError, err, "Failed to update task %s", task.ID)
	}

	// Simulate successful update
	task.Status = newStatus
	task.Meta = newMeta

	// Simulating a runtime failure
	if task.Status == "failed" {
		return Errorf(RuntimeFailure, "Task %s failed during update", task.ID)
	}

	fmt.Printf("Task %s successfully updated.\n", task.ID)
	return nil
}
