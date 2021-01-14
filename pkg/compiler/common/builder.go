// This package defines the intermediate layer that the compiler builds and transformers accept.
package common

import (
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/compiler/errors"
)

const (
	StartNodeID = "start-node"
	EndNodeID   = "end-node"
)

//go:generate mockery -all -output=mocks -case=underscore

// A mutable workflow used during the build of the intermediate layer.
type WorkflowBuilder interface {
	Workflow
	StoreCompiledSubWorkflow(id WorkflowID, compiledWorkflow *core.CompiledWorkflow)
	AddExecutionEdge(nodeFrom, nodeTo NodeID)
	AddNode(n NodeBuilder, errs errors.CompileErrors) (node NodeBuilder, ok bool)
	ValidateWorkflow(fg *core.CompiledWorkflow, errs errors.CompileErrors) (Workflow, bool)
	NewNodeBuilder(n *core.Node) NodeBuilder
}

// A mutable node used during the build of the intermediate layer.
type NodeBuilder interface {
	Node
	SetID(id string)
	SetInterface(iface *core.TypedInterface)
	SetInputs(inputs []*core.Binding)
	SetSubWorkflow(wf Workflow)
	SetTask(task Task)
}
