package update

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCommand(t *testing.T) {
	updateCommand := CreateUpdateCommand()
	assert.Equal(t, updateCommand.Use, updateUse)
	assert.Equal(t, updateCommand.Short, updateShort)
	assert.Equal(t, updateCommand.Long, updatecmdLong)
	assert.Equal(t, len(updateCommand.Commands()), 1)
	cmdNouns := updateCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	assert.Equal(t, cmdNouns[0].Use, "project")
	assert.Equal(t, cmdNouns[0].Aliases, []string{"projects"})
	assert.Equal(t, cmdNouns[0].Short, projectShort)
	assert.Equal(t, cmdNouns[0].Long, projectLong)
}
