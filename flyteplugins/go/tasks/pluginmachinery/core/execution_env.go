package core

import (
	"context"

	_struct "github.com/golang/protobuf/ptypes/struct"
)

type ExecutionEnvID struct {
	// Org is the organization that the execution environment belongs to. This is an optional field
	// so usage will need to support the case where it is empty.
	Org string

	// Project is the project that the execution environment belongs to.
	Project string

	// Domain is the domain that the execution environment belongs to.
	Domain string

	// Name is the name of the execution environment. This is designed to be a human-readable
	// identifier for the execution environment.
	Name string

	// Version is the version of the execution environment. This value is determined by the
	// specific execution environment implementation. For example, some may auto-generate based on
	// environment attributes and others set user-provided values.
	Version string
}

// String returns a string representation of the execution environment ID. This adheres to k8s Pod
// naming conventions and is used to generate unique string identifiers for execution environments.
func (e ExecutionEnvID) String() string {
	if len(e.Org) == 0 {
		return e.Project + "_" + e.Domain + "_" + e.Name + "_" + e.Version
	}
	return e.Org + "_" + e.Project + "_" + e.Domain + "_" + e.Name + "_" + e.Version
}

// ExecutionEnvClient is an interface that defines the methods to interact with an execution
// environment.
type ExecutionEnvClient interface {
	Get(ctx context.Context, executionEnvID ExecutionEnvID) *_struct.Struct
	Create(ctx context.Context, executionEnvID ExecutionEnvID, executionEnvSpec *_struct.Struct) (*_struct.Struct, error)
	Status(ctx context.Context, executionEnvID ExecutionEnvID) (interface{}, error)
}
