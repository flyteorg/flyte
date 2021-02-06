package get

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestCreateGetCommand(t *testing.T) {
	getCommand := CreateGetCommand()
	assert.Equal(t, getCommand.Use , "get")
	assert.Equal(t, getCommand.Short , "Retrieve various resource.")
	fmt.Println(getCommand.Commands())
	assert.Equal(t, len(getCommand.Commands()), 5)
	cmdNouns := getCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	assert.Equal(t, cmdNouns[0].Use, "execution")
	assert.Equal(t, cmdNouns[0].Aliases, []string{"executions"})
	assert.Equal(t, cmdNouns[0].Short, "Retrieves execution resources.")
	assert.Equal(t, cmdNouns[1].Use, "launchplan")
	assert.Equal(t, cmdNouns[1].Aliases, []string{"launchplans"})
	assert.Equal(t, cmdNouns[1].Short, "Retrieves launchplan resources.")
	assert.Equal(t, cmdNouns[2].Use, "project")
	assert.Equal(t, cmdNouns[2].Aliases, []string{"projects"})
	assert.Equal(t, cmdNouns[2].Short, "Retrieves project resources.")
	assert.Equal(t, cmdNouns[3].Use, "task")
	assert.Equal(t, cmdNouns[3].Aliases, []string{"tasks"})
	assert.Equal(t, cmdNouns[3].Short, "Retrieves task resources.")
	assert.Equal(t, cmdNouns[4].Use, "workflow")
	assert.Equal(t, cmdNouns[4].Aliases, []string{"workflows"})
	assert.Equal(t, cmdNouns[4].Short, "Retrieves workflow resources.")
}
