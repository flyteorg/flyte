package update

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestUpdateCommand(t *testing.T) {
	updateCommand := CreateUpdateCommand()
	assert.Equal(t, updateCommand.Use , "update")
	assert.Equal(t, updateCommand.Short , "Update various resources.")
	assert.Equal(t, len(updateCommand.Commands()), 1)
	cmdNouns := updateCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	assert.Equal(t, cmdNouns[0].Use, "project")
	assert.Equal(t, cmdNouns[0].Aliases, []string{"projects"})
}
