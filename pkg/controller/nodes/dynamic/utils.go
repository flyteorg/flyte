package dynamic

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/lyft/flytestdlib/storage"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/compiler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/utils"
)

// Constructs the expected interface of a given node.
func underlyingInterface(w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (*core.TypedInterface, error) {
	iface := &core.TypedInterface{}
	if node.GetTaskID() != nil {
		t, err := w.GetTask(*node.GetTaskID())
		if err != nil {
			// Should never happen
			return nil, err
		}

		iface.Outputs = t.CoreTask().GetInterface().Outputs
	} else if wfNode := node.GetWorkflowNode(); wfNode != nil {
		if wfRef := wfNode.GetSubWorkflowRef(); wfRef != nil {
			t := w.FindSubWorkflow(*wfRef)
			if t == nil {
				// Should never happen
				return nil, errors.Errorf(errors.IllegalStateError, node.GetID(), "Couldn't find subworkflow [%v].", wfRef)
			}

			iface.Outputs = t.GetOutputs().VariableMap
		} else {
			return nil, errors.Errorf(errors.IllegalStateError, node.GetID(), "Unknown interface")
		}
	} else if node.GetBranchNode() != nil {
		if ifBlock := node.GetBranchNode().GetIf(); ifBlock != nil && ifBlock.GetThenNode() != nil {
			bn, found := w.GetNode(*ifBlock.GetThenNode())
			if !found {
				return nil, errors.Errorf(errors.IllegalStateError, node.GetID(), "Couldn't find branch node [%v]",
					*ifBlock.GetThenNode())
			}

			return underlyingInterface(w, bn)
		}

		return nil, errors.Errorf(errors.IllegalStateError, node.GetID(), "Empty branch detected.")
	} else {
		return nil, errors.Errorf(errors.IllegalStateError, node.GetID(), "Unknown interface.")
	}

	return iface, nil
}

func hierarchicalNodeID(parentNodeID, nodeID string) (string, error) {
	return utils.FixedLengthUniqueIDForParts(20, parentNodeID, nodeID)
}

func updateBindingNodeIDsWithLineage(parentNodeID string, binding *core.BindingData) (err error) {
	switch b := binding.Value.(type) {
	case *core.BindingData_Promise:
		b.Promise.NodeId, err = hierarchicalNodeID(parentNodeID, b.Promise.NodeId)
		if err != nil {
			return err
		}
	case *core.BindingData_Collection:
		for _, item := range b.Collection.Bindings {
			err = updateBindingNodeIDsWithLineage(parentNodeID, item)
			if err != nil {
				return err
			}
		}
	case *core.BindingData_Map:
		for _, item := range b.Map.Bindings {
			err = updateBindingNodeIDsWithLineage(parentNodeID, item)
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

func cacheFlyteWorkflow(ctx context.Context, store storage.RawStore, wf *v1alpha1.FlyteWorkflow, target storage.DataReference) error {
	raw, err := json.Marshal(wf)
	if err != nil {
		return err
	}

	return store.WriteRaw(ctx, target, int64(len(raw)), storage.Options{}, bytes.NewReader(raw))
}

func loadCachedFlyteWorkflow(ctx context.Context, store storage.RawStore, source storage.DataReference) (
	*v1alpha1.FlyteWorkflow, error) {

	rawReader, err := store.ReadRaw(ctx, source)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(rawReader)
	if err != nil {
		return nil, err
	}

	err = rawReader.Close()
	if err != nil {
		return nil, err
	}

	wf := &v1alpha1.FlyteWorkflow{}
	return wf, json.Unmarshal(buf.Bytes(), wf)
}
