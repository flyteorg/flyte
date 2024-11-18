package dynamic

import (
	"context"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
)

// Constructs the expected interface of a given node.
func underlyingInterface(ctx context.Context, taskReader interfaces.TaskReader) (*core.TypedInterface, error) {
	t, err := taskReader.Read(ctx)
	iface := &core.TypedInterface{}
	if err != nil {
		// Should never happen
		return nil, err
	}

	if t.GetInterface() != nil {
		iface.Outputs = t.GetInterface().GetOutputs()
	}
	return iface, nil
}

func hierarchicalNodeID(parentNodeID, retryAttempt, nodeID string) (string, error) {
	return encoding.FixedLengthUniqueIDForParts(20, []string{parentNodeID, retryAttempt, nodeID})
}

func updateBindingNodeIDsWithLineage(parentNodeID, retryAttempt string, binding *core.BindingData) (err error) {
	switch b := binding.GetValue().(type) {
	case *core.BindingData_Promise:
		b.Promise.NodeId, err = hierarchicalNodeID(parentNodeID, retryAttempt, b.Promise.GetNodeId())
		if err != nil {
			return err
		}
	case *core.BindingData_Collection:
		for _, item := range b.Collection.GetBindings() {
			err = updateBindingNodeIDsWithLineage(parentNodeID, retryAttempt, item)
			if err != nil {
				return err
			}
		}
	case *core.BindingData_Map:
		for _, item := range b.Map.GetBindings() {
			err = updateBindingNodeIDsWithLineage(parentNodeID, retryAttempt, item)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func compileTasks(_ context.Context, tasks []*core.TaskTemplate) ([]*core.CompiledTask, error) {
	compiledTasks := make([]*core.CompiledTask, 0, len(tasks))
	visitedTasks := sets.NewString()
	for _, t := range tasks {
		if visitedTasks.Has(t.GetId().String()) {
			continue
		}

		ct, err := compiler.CompileTask(t)
		if err != nil {
			return nil, err
		}

		compiledTasks = append(compiledTasks, ct)
		visitedTasks.Insert(t.GetId().String())
	}

	return compiledTasks, nil
}
