package workflowstore

import (
	"fmt"

	"github.com/pkg/errors"
)

var errStaleWorkflowError = fmt.Errorf("stale Workflow Found error")
var errWorkflowNotFound = fmt.Errorf("workflow not-found error")

func IsNotFound(err error) bool {
	return errors.Cause(err) == errWorkflowNotFound
}

func IsWorkflowStale(err error) bool {
	return errors.Cause(err) == errStaleWorkflowError
}
