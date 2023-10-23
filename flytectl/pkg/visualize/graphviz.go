package visualize

import (
	"fmt"
	"strings"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"

	graphviz "github.com/awalterschulze/gographviz"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const (
	// node identifiers
	StartNode string = "start-node"
	EndNode   string = "end-node"

	// subgraph attributes
	SubgraphPrefix string = "cluster_"

	// shape attributes
	DoubleCircleShape string = "doublecircle"
	BoxShape          string = "box"
	DiamondShape      string = "diamond"
	ShapeType         string = "shape"

	// color attributes
	ColorAttr string = "color"
	Red       string = "red"
	Green     string = "green"

	// structural attributes
	LabelAttr string = "label"
	LHeadAttr string = "lhead"
	LTailAttr string = "ltail"

	// conditional
	ElseFail string = "orElse - Fail"
	Else     string = "orElse"
)

func operandToString(op *core.Operand) string {
	if op.GetPrimitive() != nil {
		l, err := coreutils.ExtractFromLiteral(&core.Literal{Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: op.GetPrimitive(),
				},
			},
		}})
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("%v", l)
	}
	return op.GetVar()
}

func comparisonToString(expr *core.ComparisonExpression) string {
	return fmt.Sprintf("%s %s %s", operandToString(expr.LeftValue), expr.Operator.String(), operandToString(expr.RightValue))
}

func conjunctionToString(expr *core.ConjunctionExpression) string {
	return fmt.Sprintf("(%s) %s (%s)", booleanExprToString(expr.LeftExpression), expr.Operator.String(), booleanExprToString(expr.RightExpression))
}

func booleanExprToString(expr *core.BooleanExpression) string {
	if expr.GetConjunction() != nil {
		return conjunctionToString(expr.GetConjunction())
	}
	return comparisonToString(expr.GetComparison())
}

func constructStartNode(parentGraph string, n string, graph Graphvizer) (*graphviz.Node, error) {
	attrs := map[string]string{ShapeType: DoubleCircleShape, ColorAttr: Green}
	attrs[LabelAttr] = "start"
	err := graph.AddNode(parentGraph, n, attrs)
	return graph.GetNode(n), err
}

func constructEndNode(parentGraph string, n string, graph Graphvizer) (*graphviz.Node, error) {
	attrs := map[string]string{ShapeType: DoubleCircleShape, ColorAttr: Red}
	attrs[LabelAttr] = "end"
	err := graph.AddNode(parentGraph, n, attrs)
	return graph.GetNode(n), err
}

func constructTaskNode(parentGraph string, name string, graph Graphvizer, n *core.Node, t *core.CompiledTask) (*graphviz.Node, error) {
	attrs := map[string]string{ShapeType: BoxShape}
	if n.Metadata != nil && n.Metadata.Name != "" {
		v := strings.LastIndexAny(n.Metadata.Name, ".")
		attrs[LabelAttr] = fmt.Sprintf("\"%s [%s]\"", n.Metadata.Name[v+1:], t.Template.Type)
	}
	tName := strings.ReplaceAll(name, "-", "_")
	err := graph.AddNode(parentGraph, tName, attrs)
	return graph.GetNode(tName), err
}

func constructErrorNode(parentGraph string, name string, graph Graphvizer, m string) (*graphviz.Node, error) {
	attrs := map[string]string{ShapeType: BoxShape, ColorAttr: Red, LabelAttr: fmt.Sprintf("\"%s\"", m)}
	eName := strings.ReplaceAll(name, "-", "_")
	err := graph.AddNode(parentGraph, eName, attrs)
	return graph.GetNode(eName), err
}

func constructBranchConditionNode(parentGraph string, name string, graph Graphvizer, n *core.Node) (*graphviz.Node, error) {
	attrs := map[string]string{ShapeType: DiamondShape}
	if n.Metadata != nil && n.Metadata.Name != "" {
		attrs[LabelAttr] = fmt.Sprintf("\"[%s]\"", n.Metadata.Name)
	}
	cName := strings.ReplaceAll(name, "-", "_")
	err := graph.AddNode(parentGraph, cName, attrs)
	return graph.GetNode(cName), err
}

func getName(prefix, id string) string {
	if prefix != "" {
		return prefix + "_" + id
	}
	return id
}

type graphBuilder struct {
	// Mutated as graph is built
	graphNodes map[string]*graphviz.Node
	// Mutated as graph is built. lookup table for all graphviz compiled edges.
	graphEdges map[string]*graphviz.Edge
	// lookup table for all graphviz compiled subgraphs
	subWf map[string]*core.CompiledWorkflow
	// a lookup table for all tasks in the graph
	tasks map[string]*core.CompiledTask
	// a lookup for all node clusters. This is to remap the edges to the cluster itself (instead of the node)
	// this is useful in the case of branchNodes and subworkflow nodes
	nodeClusters map[string]string
}

func (gb *graphBuilder) addBranchSubNodeEdge(graph Graphvizer, parentNode, n *graphviz.Node, label string) error {
	edgeName := fmt.Sprintf("%s-%s", parentNode.Name, n.Name)
	if _, ok := gb.graphEdges[edgeName]; !ok {
		attrs := map[string]string{}
		if c, ok := gb.nodeClusters[n.Name]; ok {
			attrs[LHeadAttr] = fmt.Sprintf("\"%s\"", c)
		}
		attrs[LabelAttr] = fmt.Sprintf("\"%s\"", label)
		err := graph.AddEdge(parentNode.Name, n.Name, true, attrs)
		if err != nil {
			return err
		}
		gb.graphEdges[edgeName] = graph.GetEdge(parentNode.Name, n.Name)
	}
	return nil
}

func (gb *graphBuilder) constructBranchNode(parentGraph string, prefix string, graph Graphvizer, n *core.Node) (*graphviz.Node, error) {
	parentBranchNode, err := constructBranchConditionNode(parentGraph, getName(prefix, n.Id), graph, n)
	if err != nil {
		return nil, err
	}
	gb.graphNodes[n.Id] = parentBranchNode

	if n.GetBranchNode().GetIfElse() == nil {
		return parentBranchNode, nil
	}

	subNode, err := gb.constructNode(parentGraph, prefix, graph, n.GetBranchNode().GetIfElse().Case.ThenNode)
	if err != nil {
		return nil, err
	}
	if err := gb.addBranchSubNodeEdge(graph, parentBranchNode, subNode, booleanExprToString(n.GetBranchNode().GetIfElse().Case.Condition)); err != nil {
		return nil, err
	}

	if n.GetBranchNode().GetIfElse().GetError() != nil {
		name := fmt.Sprintf("%s-error", parentBranchNode.Name)
		subNode, err := constructErrorNode(prefix, name, graph, n.GetBranchNode().GetIfElse().GetError().Message)
		if err != nil {
			return nil, err
		}
		gb.graphNodes[name] = subNode
		if err := gb.addBranchSubNodeEdge(graph, parentBranchNode, subNode, ElseFail); err != nil {
			return nil, err
		}
	} else {
		subNode, err := gb.constructNode(parentGraph, prefix, graph, n.GetBranchNode().GetIfElse().GetElseNode())
		if err != nil {
			return nil, err
		}
		if err := gb.addBranchSubNodeEdge(graph, parentBranchNode, subNode, Else); err != nil {
			return nil, err
		}
	}

	if n.GetBranchNode().GetIfElse().GetOther() != nil {
		for _, c := range n.GetBranchNode().GetIfElse().GetOther() {
			subNode, err := gb.constructNode(parentGraph, prefix, graph, c.ThenNode)
			if err != nil {
				return nil, err
			}
			if err := gb.addBranchSubNodeEdge(graph, parentBranchNode, subNode, booleanExprToString(c.Condition)); err != nil {
				return nil, err
			}
		}
	}
	return parentBranchNode, nil
}

func (gb *graphBuilder) constructNode(parentGraphName string, prefix string, graph Graphvizer, n *core.Node) (*graphviz.Node, error) {
	name := getName(prefix, n.Id)
	var err error
	var gn *graphviz.Node

	if n.Id == StartNode {
		gn, err = constructStartNode(parentGraphName, strings.ReplaceAll(name, "-", "_"), graph)
		gb.nodeClusters[name] = parentGraphName
	} else if n.Id == EndNode {
		gn, err = constructEndNode(parentGraphName, strings.ReplaceAll(name, "-", "_"), graph)
		gb.nodeClusters[name] = parentGraphName
	} else {
		switch n.Target.(type) {
		case *core.Node_TaskNode:
			tID := n.GetTaskNode().GetReferenceId().String()
			t, ok := gb.tasks[tID]
			if !ok {
				return nil, fmt.Errorf("failed to find task [%s] in closure", tID)
			}
			gn, err = constructTaskNode(parentGraphName, name, graph, n, t)
			if err != nil {
				return nil, err
			}
			gb.nodeClusters[name] = parentGraphName
		case *core.Node_BranchNode:
			sanitizedName := strings.ReplaceAll(n.Metadata.Name, "-", "_")
			branchSubGraphName := SubgraphPrefix + sanitizedName
			err := graph.AddSubGraph(parentGraphName, branchSubGraphName, map[string]string{LabelAttr: sanitizedName})
			if err != nil {
				return nil, err
			}
			gn, err = gb.constructBranchNode(branchSubGraphName, prefix, graph, n)
			if err != nil {
				return nil, err
			}
			gb.nodeClusters[name] = branchSubGraphName
		case *core.Node_WorkflowNode:
			if n.GetWorkflowNode().GetLaunchplanRef() != nil {
				attrs := map[string]string{}
				err := graph.AddNode(parentGraphName, name, attrs)
				if err != nil {
					return nil, err
				}
			} else {
				sanitizedName := strings.ReplaceAll(name, "-", "_")
				subGraphName := SubgraphPrefix + sanitizedName
				err := graph.AddSubGraph(parentGraphName, subGraphName, map[string]string{LabelAttr: sanitizedName})
				if err != nil {
					return nil, err
				}
				subGB := graphBuilderFromParent(gb)
				swf, ok := gb.subWf[n.GetWorkflowNode().GetSubWorkflowRef().String()]
				if !ok {
					return nil, fmt.Errorf("subworkfow [%s] not found", n.GetWorkflowNode().GetSubWorkflowRef().String())
				}
				if err := subGB.constructGraph(subGraphName, name, graph, swf); err != nil {
					return nil, err
				}
				gn = subGB.graphNodes[StartNode]
				gb.nodeClusters[gn.Name] = subGraphName
			}
		}
	}
	if err != nil {
		return nil, err
	}
	gb.graphNodes[n.Id] = gn
	return gn, nil
}

func (gb *graphBuilder) addEdge(fromNodeName, toNodeName string, graph Graphvizer) error {
	toNode, toOk := gb.graphNodes[toNodeName]
	fromNode, fromOk := gb.graphNodes[fromNodeName]
	if !toOk || !fromOk {
		return fmt.Errorf("nodes[%s] -> [%s] referenced before creation", fromNodeName, toNodeName)
	}
	if !graph.DoesEdgeExist(fromNode.Name, toNode.Name) {
		attrs := map[string]string{}
		// Now lets check that the toNode or the fromNode is a cluster. If so then following this thread,
		// https://stackoverflow.com/questions/2012036/graphviz-how-to-connect-subgraphs, we will connect the cluster
		if c, ok := gb.nodeClusters[fromNode.Name]; ok {
			attrs[LTailAttr] = fmt.Sprintf("\"%s\"", c)
		}
		if c, ok := gb.nodeClusters[toNode.Name]; ok {
			attrs[LHeadAttr] = fmt.Sprintf("\"%s\"", c)
		}
		err := graph.AddEdge(fromNode.Name, toNode.Name, true, attrs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gb *graphBuilder) constructGraph(parentGraphName string, prefix string, graph Graphvizer, w *core.CompiledWorkflow) error {
	if w == nil || w.Template == nil {
		return nil
	}
	for _, n := range w.Template.Nodes {
		if _, err := gb.constructNode(parentGraphName, prefix, graph, n); err != nil {
			return err
		}
	}

	for name := range gb.graphNodes {
		upstreamNodes := w.Connections.Upstream[name]
		downstreamNodes := w.Connections.Downstream[name]
		if downstreamNodes != nil {
			for _, n := range downstreamNodes.Ids {
				if err := gb.addEdge(name, n, graph); err != nil {
					return err
				}
			}
		}
		if upstreamNodes != nil {
			for _, n := range upstreamNodes.Ids {
				if err := gb.addEdge(n, name, graph); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (gb *graphBuilder) CompiledWorkflowClosureToGraph(w *core.CompiledWorkflowClosure) (FlyteGraph, error) {
	dotGraph := FlyteGraph{graphviz.NewGraph()}
	_ = dotGraph.SetDir(true)
	_ = dotGraph.SetStrict(true)

	tLookup := make(map[string]*core.CompiledTask)
	for _, t := range w.Tasks {
		if t.Template == nil || t.Template.Id == nil {
			return FlyteGraph{}, fmt.Errorf("no template found in the workflow task %v", t)
		}
		tLookup[t.Template.Id.String()] = t
	}
	gb.tasks = tLookup
	wLookup := make(map[string]*core.CompiledWorkflow)
	for _, swf := range w.SubWorkflows {
		if swf.Template == nil || swf.Template.Id == nil {
			return FlyteGraph{}, fmt.Errorf("no template found in the sub workflow %v", swf)
		}
		wLookup[swf.Template.Id.String()] = swf
	}
	gb.subWf = wLookup

	return dotGraph, gb.constructGraph("", "", dotGraph, w.Primary)
}

func newGraphBuilder() *graphBuilder {
	return &graphBuilder{
		graphNodes:   make(map[string]*graphviz.Node),
		graphEdges:   make(map[string]*graphviz.Edge),
		nodeClusters: make(map[string]string),
	}
}

func graphBuilderFromParent(gb *graphBuilder) *graphBuilder {
	newGB := newGraphBuilder()
	newGB.subWf = gb.subWf
	newGB.tasks = gb.tasks
	return newGB
}

// RenderWorkflow Renders the workflow graph on the console
func RenderWorkflow(w *core.CompiledWorkflowClosure) (string, error) {
	if w == nil {
		return "", fmt.Errorf("empty workflow closure")
	}
	gb := newGraphBuilder()
	graph, err := gb.CompiledWorkflowClosureToGraph(w)
	if err != nil {
		return "", err
	}
	return graph.String(), nil
}
