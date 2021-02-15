package update

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCommand(t *testing.T) {
	updateCommand := CreateUpdateCommand()
	assert.Equal(t, updateCommand.Use, "update")
	assert.Equal(t, updateCommand.Short, "\nUsed for updating flyte resources eg: project.\n")
	assert.Equal(t, len(updateCommand.Commands()), 1)
	cmdNouns := updateCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	assert.Equal(t, cmdNouns[0].Use, "project")
	assert.Equal(t, cmdNouns[0].Aliases, []string{"projects"})
}
