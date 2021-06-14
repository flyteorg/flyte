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
	assert.Equal(t, sandboxCommand.Short, "Used for testing flyte sandbox.")
	fmt.Println(sandboxCommand.Commands())
	assert.Equal(t, len(sandboxCommand.Commands()), 2)
	cmdNouns := sandboxCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})

	assert.Equal(t, cmdNouns[0].Use, "start")
	assert.Equal(t, cmdNouns[0].Short, startShort)
	assert.Equal(t, cmdNouns[0].Long, startLong)

	assert.Equal(t, cmdNouns[1].Use, "teardown")
	assert.Equal(t, cmdNouns[1].Short, teardownShort)
	assert.Equal(t, cmdNouns[1].Long, teardownLong)

}
