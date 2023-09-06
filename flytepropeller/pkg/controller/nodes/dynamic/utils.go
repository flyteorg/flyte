package dynamic

import (
	"context"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
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
		iface.Outputs = t.GetInterface().Outputs
	}
	return iface, nil
}

func hierarchicalNodeID(parentNodeID, retryAttempt, nodeID string) (string, error) {
	return encoding.FixedLengthUniqueIDForParts(20, []string{parentNodeID, retryAttempt, nodeID})
}

func updateBindingNodeIDsWithLineage(parentNodeID, retryAttempt string, binding *core.BindingData) (err error) {
	switch b := binding.Value.(type) {
	case *core.BindingData_Promise:
		b.Promise.NodeId, err = hierarchicalNodeID(parentNodeID, retryAttempt, b.Promise.NodeId)
		if err != nil {
			return err
		}
	case *core.BindingData_Collection:
		for _, item := range b.Collection.Bindings {
			err = updateBindingNodeIDsWithLineage(parentNodeID, retryAttempt, item)
			if err != nil {
				return err
			}
		}
	case *core.BindingData_Map:
		for _, item := range b.Map.Bindings {
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
		if visitedTasks.Has(t.Id.String()) {
			continue
		}

		ct, err := compiler.CompileTask(t)
		if err != nil {
			return nil, err
		}

		compiledTasks = append(compiledTasks, ct)
		visitedTasks.Insert(t.Id.String())
	}

	return compiledTasks, nil
}
