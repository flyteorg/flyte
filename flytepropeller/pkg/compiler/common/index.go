package common

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Defines an index of nodebuilders based on the id.
type NodeIndex map[NodeID]NodeBuilder

// Defines an index of tasks based on the id.
type TaskIndex map[TaskIDKey]Task

type WorkflowIndex map[WorkflowIDKey]*core.CompiledWorkflow

// Defines a string adjacency list.
type AdjacencyList map[string]IdentifierSet

type StringAdjacencyList map[string]sets.String

// Converts the sets in the adjacency list to sorted arrays.
func (l AdjacencyList) ToMapOfLists() map[string][]Identifier {
	res := make(map[string][]Identifier, len(l))
	for key, set := range l {
		res[key] = set.List()
	}

	return res
}

// Creates a new TaskIndex.
func NewTaskIndex(tasks ...Task) TaskIndex {
	res := make(TaskIndex, len(tasks))
	for _, task := range tasks {
		id := task.GetID()
		res[(&id).String()] = task
	}

	return res
}

// Creates a new NodeIndex
func NewNodeIndex(nodes ...NodeBuilder) NodeIndex {
	res := make(NodeIndex, len(nodes))
	for _, task := range nodes {
		res[task.GetId()] = task
	}

	return res
}

func NewWorkflowIndex(workflows []*core.CompiledWorkflow, errs errors.CompileErrors) (index WorkflowIndex, ok bool) {
	ok = true
	index = make(WorkflowIndex, len(workflows))
	for _, wf := range workflows {
		if wf.Template.Id == nil {
			// TODO: Log/Return error
			return nil, false
		}

		if _, found := index[wf.Template.Id.String()]; found {
			errs.Collect(errors.NewDuplicateIDFoundErr(wf.Template.Id.String()))
			ok = false
		} else {
			index[wf.Template.Id.String()] = wf
		}
	}

	return
}
