package visualize

import graphviz "github.com/awalterschulze/gographviz"

//go:generate mockery -all -case=underscore

type Graphvizer interface {
	AddEdge(src, dst string, directed bool, attrs map[string]string) error
	AddNode(parentGraph string, name string, attrs map[string]string) error
	AddSubGraph(parentGraph string, name string, attrs map[string]string) error
	AddAttr(parentGraph string, field string, value string) error
	SetName(name string) error
	GetEdge(src, dest string) *graphviz.Edge
	GetNode(key string) *graphviz.Node
	DoesEdgeExist(src, dest string) bool
}

type FlyteGraph struct {
	*graphviz.Graph
}

// GetNode given the key to the node
func (g FlyteGraph) GetNode(key string) *graphviz.Node {
	return g.Nodes.Lookup[key]
}

// GetEdge gets the edge in the graph from src to dest
func (g FlyteGraph) GetEdge(src, dest string) *graphviz.Edge {
	return g.Edges.SrcToDsts[src][dest][0]
}

// DoesEdgeExist checks if an edge exists in the graph from src to dest
func (g FlyteGraph) DoesEdgeExist(src, dest string) bool {
	return g.Edges.SrcToDsts[src][dest] != nil
}
