package workflowstore

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrStaleWorkflowError signals that the local copy of workflow is Stale, i.e., a new version was written to the datastore,
// But the informer cache has not yet synced to the latest copy
var ErrStaleWorkflowError = fmt.Errorf("stale Workflow Found error")

// ErrWorkflowNotFound indicates that the workflow does not exist and it is safe to ignore the event
var ErrWorkflowNotFound = fmt.Errorf("workflow not-found error")

// ErrWorkflowTerminated indicates that the workflow being operated on has previously been stored in a terminal state.
var ErrWorkflowTerminated = fmt.Errorf("workflow has already been terminated")

// ErrWorkflowToLarge is returned in cased an update operation fails because the Workflow object (CRD) has surpassed the Datastores
// supported limit.
var ErrWorkflowToLarge = fmt.Errorf("workflow too large")

// IsNotFound returns true if the error is caused by ErrWorkflowNotFound
func IsNotFound(err error) bool {
	return errors.Cause(err) == ErrWorkflowNotFound
}

// IsWorkflowStale returns true if the error is caused by ErrStaleWorkflowError
func IsWorkflowStale(err error) bool {
	return errors.Cause(err) == ErrStaleWorkflowError
}

// IsWorkflowTerminated returns true if the error is caused by ErrWorkflowTerminated
func IsWorkflowTerminated(err error) bool {
	return errors.Cause(err) == ErrWorkflowTerminated
}

// IsWorkflowTooLarge returns true if the error is caused by ErrWorkflowToLarge
func IsWorkflowTooLarge(err error) bool {
	return errors.Cause(err) == ErrWorkflowToLarge
}
