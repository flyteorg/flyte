package compiler

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"k8s.io/apimachinery/pkg/util/sets"
)

func toInterfaceProviderMap(tasks []common.InterfaceProvider) map[string]common.InterfaceProvider {
	res := make(map[string]common.InterfaceProvider, len(tasks))
	for _, task := range tasks {
		res[task.GetID().String()] = task
	}

	return res
}

func toSlice(s sets.String) []string {
	res := make([]string, 0, len(s))
	for str := range s {
		res = append(res, str)
	}

	return res
}

func toNodeIdsSet(nodes common.NodeIndex) sets.String {
	res := sets.NewString()
	for nodeID := range nodes {
		res.Insert(nodeID)
	}

	return res
}

// Runs a depth-first coreWorkflow traversal to detect any cycles in the coreWorkflow. It produces the first cycle found, as well as
// all visited nodes and a boolean indicating whether or not it found a cycle.
func detectCycle(startNode string, neighbors func(nodeId string) sets.String) (cycle []common.NodeID, visited sets.String,
	detected bool) {

	// This is a set of nodes that were ever visited.
	visited = sets.NewString()
	// This is a set of in-progress visiting nodes.
	visiting := sets.NewString()
	var detector func(nodeId string) ([]common.NodeID, bool)
	detector = func(nodeId string) ([]common.NodeID, bool) {
		if visiting.Has(nodeId) {
			return []common.NodeID{}, true
		}

		visiting.Insert(nodeId)
		visited.Insert(nodeId)

		for nextID := range neighbors(nodeId) {
			if path, detected := detector(nextID); detected {
				return append([]common.NodeID{nextID}, path...), true
			}
		}

		visiting.Delete(nodeId)

		return []common.NodeID{}, false
	}

	if path, detected := detector(startNode); detected {
		return append([]common.NodeID{startNode}, path...), visiting, true
	}

	return []common.NodeID{}, visited, false
}

func toCompiledWorkflows(wfs ...*core.WorkflowTemplate) []*core.CompiledWorkflow {
	compiledSubWfs := make([]*core.CompiledWorkflow, 0, len(wfs))
	for _, wf := range wfs {
		compiledSubWfs = append(compiledSubWfs, &core.CompiledWorkflow{Template: wf})
	}

	return compiledSubWfs
}
