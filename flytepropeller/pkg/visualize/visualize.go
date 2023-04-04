package visualize

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"k8s.io/apimachinery/pkg/util/sets"
)

const executionEdgeLabel = "execution"

type edgeStyle = string

const (
	styleSolid  edgeStyle = "solid"
	styleDashed edgeStyle = "dashed"
)

const staticNodeID = "static"

func flatten(binding *core.BindingData, flatMap map[common.NodeID]sets.String) {
	switch binding.GetValue().(type) {
	case *core.BindingData_Collection:
		for _, v := range binding.GetCollection().GetBindings() {
			flatten(v, flatMap)
		}
	case *core.BindingData_Map:
		for _, v := range binding.GetMap().GetBindings() {
			flatten(v, flatMap)
		}
	case *core.BindingData_Promise:
		if _, ok := flatMap[binding.GetPromise().NodeId]; !ok {
			flatMap[binding.GetPromise().NodeId] = sets.String{}
		}

		flatMap[binding.GetPromise().NodeId].Insert(binding.GetPromise().GetVar())
	case *core.BindingData_Scalar:
		if _, ok := flatMap[staticNodeID]; !ok {
			flatMap[staticNodeID] = sets.NewString()
		}
	}
}

// Returns GraphViz https://www.graphviz.org/ representation of the current state of the state machine.
func WorkflowToGraphViz(g *v1alpha1.FlyteWorkflow) string {
	res := fmt.Sprintf("digraph G {rankdir=TB;workflow[label=\"Workflow Id: %v\"];node[style=filled];",
		g.ID)

	nodeFinder := func(nodeId common.NodeID) *v1alpha1.NodeSpec {
		for _, n := range g.Nodes {
			if n.ID == nodeId {
				return n
			}
		}

		return nil
	}

	nodeLabel := func(nodeId common.NodeID) string {
		node := nodeFinder(nodeId)
		return fmt.Sprintf("%v(%v)", node.ID, node.Kind)
	}

	edgeLabel := func(nodeFromId, nodeToId common.NodeID) string {
		flatMap := make(map[common.NodeID]sets.String)
		nodeFrom := nodeFinder(nodeFromId)
		nodeTo := nodeFinder(nodeToId)
		for _, binding := range nodeTo.GetInputBindings() {
			flatten(binding.GetBinding(), flatMap)
		}

		if vars, found := flatMap[nodeFrom.ID]; found {
			return strings.Join(vars.List(), ",")
		} else if vars, found := flatMap[""]; found && nodeFromId == common.StartNodeID {
			return strings.Join(vars.List(), ",")
		} else {
			return executionEdgeLabel
		}
	}

	style := func(edgeLabel string) string {
		if edgeLabel == executionEdgeLabel {
			return styleDashed
		}

		return styleSolid
	}

	start := nodeFinder(common.StartNodeID)
	res += fmt.Sprintf("\"%v\" [shape=Msquare];", nodeLabel(start.ID))
	visitedNodes := sets.NewString(start.ID)
	createdEdges := sets.NewString()

	for nodesToVisit := NewNodeNameQ(start.ID); nodesToVisit.HasNext(); {
		node := nodesToVisit.Deque()
		nodes, found := g.GetConnections().Downstream[node]
		if found {
			nodesToVisit.Enqueue(nodes...)

			for _, child := range nodes {
				label := edgeLabel(node, child)
				edge := fmt.Sprintf("\"%v\" -> \"%v\" [label=\"%v\",style=\"%v\"];",
					nodeLabel(node),
					nodeLabel(child),
					label,
					style(label),
				)

				if !createdEdges.Has(edge) {
					res += edge
					createdEdges.Insert(edge)
				}
			}
		}

		// add static bindings' links
		flatMap := make(common.StringAdjacencyList)
		n := nodeFinder(node)
		for _, binding := range n.GetInputBindings() {
			flatten(binding.GetBinding(), flatMap)
		}

		if vars, found := flatMap[staticNodeID]; found {
			res += fmt.Sprintf("\"static\" -> \"%v\" [label=\"%v\"];",
				nodeLabel(node),
				strings.Join(vars.List(), ","),
			)
		}

		visitedNodes.Insert(node)
	}

	res += "}"

	return res
}

func ToGraphViz(g *core.CompiledWorkflow) string {
	res := fmt.Sprintf("digraph G {rankdir=TB;workflow[label=\"Workflow Id: %v\"];node[style=filled];",
		g.Template.GetId())

	nodeFinder := func(nodeId common.NodeID) *core.Node {
		for _, n := range g.Template.Nodes {
			if n.Id == nodeId {
				return n
			}
		}

		return nil
	}

	nodeKind := func(nodeId common.NodeID) string {
		node := nodeFinder(nodeId)
		if nodeId == common.StartNodeID {
			return "start"
		} else if nodeId == common.EndNodeID {
			return "end"
		} else {
			return reflect.TypeOf(node.GetTarget()).Name()
		}
	}

	nodeLabel := func(nodeId common.NodeID) string {
		node := nodeFinder(nodeId)
		return fmt.Sprintf("%v(%v)", node.GetId(), nodeKind(nodeId))
	}

	edgeLabel := func(nodeFromId, nodeToId common.NodeID) string {
		flatMap := make(map[common.NodeID]sets.String)
		nodeFrom := nodeFinder(nodeFromId)
		nodeTo := nodeFinder(nodeToId)
		for _, binding := range nodeTo.GetInputs() {
			flatten(binding.GetBinding(), flatMap)
		}

		if vars, found := flatMap[nodeFrom.GetId()]; found {
			return strings.Join(vars.List(), ",")
		} else if vars, found := flatMap[""]; found && nodeFromId == common.StartNodeID {
			return strings.Join(vars.List(), ",")
		} else {
			return executionEdgeLabel
		}
	}

	style := func(edgeLabel string) string {
		if edgeLabel == executionEdgeLabel {
			return styleDashed
		}

		return styleSolid
	}

	start := nodeFinder(common.StartNodeID)
	res += fmt.Sprintf("\"%v\" [shape=Msquare];", nodeLabel(start.GetId()))
	visitedNodes := sets.NewString(start.GetId())
	createdEdges := sets.NewString()

	for nodesToVisit := NewNodeNameQ(start.GetId()); nodesToVisit.HasNext(); {
		node := nodesToVisit.Deque()
		nodes, found := g.GetConnections().GetDownstream()[node]
		if found {
			nodesToVisit.Enqueue(nodes.Ids...)

			for _, child := range nodes.Ids {
				label := edgeLabel(node, child)
				edge := fmt.Sprintf("\"%v\" -> \"%v\" [label=\"%v\",style=\"%v\"];",
					nodeLabel(node),
					nodeLabel(child),
					label,
					style(label),
				)

				if !createdEdges.Has(edge) {
					res += edge
					createdEdges.Insert(edge)
				}
			}
		}

		// add static bindings' links
		flatMap := make(common.StringAdjacencyList)
		n := nodeFinder(node)
		for _, binding := range n.GetInputs() {
			flatten(binding.GetBinding(), flatMap)
		}

		if vars, found := flatMap[staticNodeID]; found {
			res += fmt.Sprintf("\"static\" -> \"%v\" [label=\"%v\"];",
				nodeLabel(node),
				strings.Join(vars.List(), ","),
			)
		}

		visitedNodes.Insert(node)
	}

	res += "}"

	return res
}
