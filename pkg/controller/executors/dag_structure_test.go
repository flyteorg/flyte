package executors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeafNodeDAGStructure(t *testing.T) {
	lds := NewLeafNodeDAGStructure("n1", "np1", "np2")
	t.Run("ToNodeHappy", func(t *testing.T) {
		nl, err := lds.ToNode("n1")
		assert.NoError(t, err)
		assert.Equal(t, []string{"np1", "np2"}, nl)
	})

	t.Run("ToNodeSad", func(t *testing.T) {
		_, err := lds.ToNode("np1")
		assert.Error(t, err)
	})

	t.Run("FromNodeAny", func(t *testing.T) {
		nl, err := lds.FromNode("np1")
		assert.NoError(t, err)
		assert.Equal(t, []string{}, nl)
	})

	t.Run("FromNodeLeaf", func(t *testing.T) {
		nl, err := lds.FromNode("n1")
		assert.NoError(t, err)
		assert.Equal(t, []string{}, nl)
	})

}
