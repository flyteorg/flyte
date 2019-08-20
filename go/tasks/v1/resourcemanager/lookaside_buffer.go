package resourcemanager

import (
	"context"
	"fmt"
)

// This error will be returned when a key is not found in the buffer
var ExecutionNotFoundError = fmt.Errorf("Execution not found")

//go:generate mockery -name ExecutionLooksideBuffer -case=underscore

// Remembers an execution key to a value.  Specifically for example in the Qubole case, the key will be the same
// key that's used in the AutoRefreshCache
type ExecutionLooksideBuffer interface {
	ConfirmExecution(ctx context.Context, executionKey string, executionValue string) error
	RetrieveExecution(ctx context.Context, executionKey string) (string, error)
}
