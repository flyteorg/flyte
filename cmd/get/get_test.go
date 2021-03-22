package get

import (
	"context"
	"fmt"
	"sort"
	"testing"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"

	"github.com/stretchr/testify/assert"
)

const projectValue = "dummyProject"
const domainValue = "dummyDomain"
const output = "json"
const executionNameValue = "e124"
const launchPlanNameValue = "lp_name"
const launchPlanVersionValue = "lp_version"
const workflowNameValue = "wf_name"
const workflowVersionValue = "wf_version"
const testDataFolder = "../testdata/"

var (
	err        error
	ctx        context.Context
	mockClient *mocks.AdminServiceClient
	cmdCtx     cmdCore.CommandContext
)
var setup = testutils.Setup
var tearDownAndVerify = testutils.TearDownAndVerify

func TestCreateGetCommand(t *testing.T) {
	getCommand := CreateGetCommand()
	assert.Equal(t, getCommand.Use, "get")
	assert.Equal(t, getCommand.Short, "Used for fetching various flyte resources including tasks/workflows/launchplans/executions/project.")
	fmt.Println(getCommand.Commands())
	assert.Equal(t, len(getCommand.Commands()), 5)
	cmdNouns := getCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	assert.Equal(t, cmdNouns[0].Use, "execution")
	assert.Equal(t, cmdNouns[0].Aliases, []string{"executions"})
	assert.Equal(t, cmdNouns[0].Short, "Gets execution resources")
	assert.Equal(t, cmdNouns[1].Use, "launchplan")
	assert.Equal(t, cmdNouns[1].Aliases, []string{"launchplans"})
	assert.Equal(t, cmdNouns[1].Short, "Gets launch plan resources")
	assert.Equal(t, cmdNouns[2].Use, "project")
	assert.Equal(t, cmdNouns[2].Aliases, []string{"projects"})
	assert.Equal(t, cmdNouns[2].Short, "Gets project resources")
	assert.Equal(t, cmdNouns[3].Use, "task")
	assert.Equal(t, cmdNouns[3].Aliases, []string{"tasks"})
	assert.Equal(t, cmdNouns[3].Short, "Gets task resources")
	assert.Equal(t, cmdNouns[4].Use, "workflow")
	assert.Equal(t, cmdNouns[4].Aliases, []string{"workflows"})
	assert.Equal(t, cmdNouns[4].Short, "Gets workflow resources")
}
