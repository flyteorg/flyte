package errors

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestCompileErrorSet_List(t *testing.T) {
	set := compileErrorSet{}
	set.Put(*NewValueRequiredErr("node1", "param"))
	set.Put(*NewWorkflowHasNoEntryNodeErr("graph1"))
	set.Put(*NewWorkflowHasNoEntryNodeErr("graph1"))
	assert.Equal(t, len(set), 2)

	lst := set.List()
	assert.Equal(t, len(lst), 2)
	assert.Equal(t, lst[0].Code(), NoEntryNodeFound)
	assert.Equal(t, lst[1].Code(), ValueRequired)
}
