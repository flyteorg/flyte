package core

import (
	"context"

	_struct "github.com/golang/protobuf/ptypes/struct"
)

// ExecutionEnvClient is an interface that defines the methods to interact with an execution
// environment.
type ExecutionEnvClient interface {
	Get(ctx context.Context, executionEnvID string) *_struct.Struct
	Create(ctx context.Context, executionEnvID string, executionEnvSpec *_struct.Struct) (*_struct.Struct, error)
	Status(ctx context.Context, executionEnvID string) (interface{}, error)
}
