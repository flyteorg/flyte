package visualize

import (
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/pkg/errors"
)

type VisitStatus int8

const (
	NotVisited VisitStatus = iota
	Visited
	Completed
)

type NodeVisitor map[v1alpha1.NodeID]VisitStatus

func NewNodeVisitor(nodes []v1alpha1.NodeID) NodeVisitor {
	v := make(NodeVisitor, len(nodes))
	for _, n := range nodes {
		v[n] = NotVisited
	}
	return v
}

func tsortHelper(g v1alpha1.ExecutableSubWorkflow, currentNode v1alpha1.ExecutableNode, visited NodeVisitor, reverseSortedNodes *[]v1alpha1.ExecutableNode) error {
	if visited[currentNode.GetID()] == NotVisited {
		visited[currentNode.GetID()] = Visited
		defer func() {
			visited[currentNode.GetID()] = Completed
		}()
		nodes, err := g.FromNode(currentNode.GetID())
		if err != nil {
			return err
		}
		for _, childID := range nodes {
			child, ok := g.GetNode(childID)
			if !ok {
				return errors.Errorf("Unable to find Node [%s] in Workflow [%s]", childID, g.GetID())
			}
			if err := tsortHelper(g, child, visited, reverseSortedNodes); err != nil {
				return err
			}
		}

		*reverseSortedNodes = append(*reverseSortedNodes, currentNode)
		return nil
	}
	// Node was successfully visited previously
	if visited[currentNode.GetID()] == Completed {
		return nil
	}
	// Node was partially visited and we are in the subgraph and that reached back to the parent
	return errors.Errorf("Cycle detected. Node [%v]", currentNode.GetID())
}

func reverseSlice(sl []v1alpha1.ExecutableNode) []v1alpha1.ExecutableNode {
	for i := len(sl)/2 - 1; i >= 0; i-- {
		opp := len(sl) - 1 - i
		sl[i], sl[opp] = sl[opp], sl[i]
	}
	return sl
}

func TopologicalSort(g v1alpha1.ExecutableSubWorkflow) ([]v1alpha1.ExecutableNode, error) {
	reverseSortedNodes := make([]v1alpha1.ExecutableNode, 0, 25)
	visited := NewNodeVisitor(g.GetNodes())
	if err := tsortHelper(g, g.StartNode(), visited, &reverseSortedNodes); err != nil {
		return nil, err
	}
	return reverseSlice(reverseSortedNodes), nil
}
