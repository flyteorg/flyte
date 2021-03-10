package visualize

import "github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"

type NodeQ []v1alpha1.NodeID

func (s *NodeQ) Enqueue(items ...v1alpha1.NodeID) {
	*s = append(*s, items...)
}

func (s NodeQ) HasNext() bool {
	return len(s) > 0
}

func (s NodeQ) Remaining() int {
	return len(s)
}

func (s *NodeQ) Peek() v1alpha1.NodeID {
	if s.HasNext() {
		return (*s)[0]
	}

	return ""
}

func (s *NodeQ) Deque() v1alpha1.NodeID {
	item := s.Peek()
	*s = (*s)[1:]
	return item
}

func NewNodeNameQ(items ...v1alpha1.NodeID) NodeQ {
	return NodeQ(items)
}
