package dynamic

import (
	"context"

	"github.com/Masterminds/semver"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/lyft/flytepropeller/pkg/compiler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/utils"
)

// Constructs the expected interface of a given node.
func underlyingInterface(ctx context.Context, taskReader handler.TaskReader) (*core.TypedInterface, error) {
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

func isFlyteKitVersionBelow(runtime *core.RuntimeMetadata, ver *semver.Version) (bool, error) {
	if runtime == nil {
		return false, nil
	}

	if runtime.Type != core.RuntimeMetadata_FLYTE_SDK {
		return false, nil
	}

	v, err := semver.NewVersion(runtime.Version)
	if err != nil {
		return false, err
	}

	return v.LessThan(ver), nil
}

func makeArrayInterface(varMap *core.VariableMap) *core.VariableMap {
	if varMap == nil || len(varMap.Variables) == 0 {
		return varMap
	}

	for _, val := range varMap.Variables {
		val.Type = &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: val.Type,
			},
		}
	}

	return varMap
}
