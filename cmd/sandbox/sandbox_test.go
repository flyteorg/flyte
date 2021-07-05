package sandbox

import (
	"fmt"
	"sort"
	"testing"

	"gotest.tools/assert"
)

func TestCreateSandboxCommand(t *testing.T) {
	sandboxCommand := CreateSandboxCommand()
	assert.Equal(t, sandboxCommand.Use, "sandbox")
	assert.Equal(t, sandboxCommand.Short, "Used for sandbox interactions like start/teardown/status/exec.")
	fmt.Println(sandboxCommand.Commands())
	assert.Equal(t, len(sandboxCommand.Commands()), 4)
	cmdNouns := sandboxCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})

	assert.Equal(t, cmdNouns[0].Use, "exec")
	assert.Equal(t, cmdNouns[0].Short, execShort)
	assert.Equal(t, cmdNouns[0].Long, execLong)

	assert.Equal(t, cmdNouns[1].Use, "start")
	assert.Equal(t, cmdNouns[1].Short, startShort)
	assert.Equal(t, cmdNouns[1].Long, startLong)

	assert.Equal(t, cmdNouns[2].Use, "status")
	assert.Equal(t, cmdNouns[2].Short, statusShort)
	assert.Equal(t, cmdNouns[2].Long, statusLong)

	assert.Equal(t, cmdNouns[3].Use, "teardown")
	assert.Equal(t, cmdNouns[3].Short, teardownShort)
	assert.Equal(t, cmdNouns[3].Long, teardownLong)

}
