package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type CanDo interface {
	MyDo() int
}

type Concrete struct {
	Doer CanDo
}

func (c *Concrete) MyDo() int {
	return 1
}

type Parent struct {
	Concrete *Concrete
}

func (p *Parent) GetDoer() CanDo {
	return p.Concrete
}

func (p *Parent) GetConcreteDoer() *Concrete {
	return p.Concrete
}

func TestPointersForNodeSpec(t *testing.T) {
	p := &Parent{
		Concrete: nil,
	}
	// GetDoer returns a fake nil because it carries type information
	// assert.NotNil(t, p.GetDoer()) funnily enough doesn't work, so use a regular if statement
	if p.GetDoer() == nil {
		assert.Fail(t, "GetDoer")
	}

	assert.Nil(t, p.GetConcreteDoer())
}

func TestNodeSpec(t *testing.T) {
	n := &NodeSpec{
		WorkflowNode: nil,
	}
	assert.Nil(t, n.GetWorkflowNode())
}
