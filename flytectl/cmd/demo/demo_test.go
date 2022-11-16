package demo

import (
	"fmt"
	"sort"
	"testing"

	"gotest.tools/assert"
)

func TestCreateDemoCommand(t *testing.T) {
	demoCommand := CreateDemoCommand()
	assert.Equal(t, demoCommand.Use, "demo")
	assert.Equal(t, demoCommand.Short, "Helps with demo interactions like start, teardown, status, and exec.")
	fmt.Println(demoCommand.Commands())

	assert.Equal(t, len(demoCommand.Commands()), 5)
	cmdNouns := demoCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})

	assert.Equal(t, cmdNouns[0].Use, "exec")
	assert.Equal(t, cmdNouns[0].Short, execShort)
	assert.Equal(t, cmdNouns[0].Long, execLong)

	assert.Equal(t, cmdNouns[1].Use, "reload")
	assert.Equal(t, cmdNouns[1].Short, reloadShort)
	assert.Equal(t, cmdNouns[1].Long, reloadLong)

	assert.Equal(t, cmdNouns[2].Use, "start")
	assert.Equal(t, cmdNouns[2].Short, startShort)
	assert.Equal(t, cmdNouns[2].Long, startLong)

	assert.Equal(t, cmdNouns[3].Use, "status")
	assert.Equal(t, cmdNouns[3].Short, statusShort)
	assert.Equal(t, cmdNouns[3].Long, statusLong)

	assert.Equal(t, cmdNouns[4].Use, "teardown")
	assert.Equal(t, cmdNouns[4].Short, teardownShort)
	assert.Equal(t, cmdNouns[4].Long, teardownLong)
}
