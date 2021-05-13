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
	assert.Equal(t, len(getCommand.Commands()), 7)
	cmdNouns := getCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	useArray := []string{"cluster-resource-attribute", "execution", "launchplan", "project",
		"task", "task-resource-attribute", "workflow"}
	aliases := [][]string{{"cluster-resource-attributes"}, {"executions"}, {"launchplans"}, {"projects"}, {"tasks"},
		{"task-resource-attributes"}, {"workflows"}}
	shortArray := []string{clusterResourceAttributesShort, executionShort, launchPlanShort, projectShort,
		taskShort, taskResourceAttributesShort, workflowShort}
	longArray := []string{clusterResourceAttributesLong, executionLong, launchPlanLong, projectLong, taskLong,
		taskResourceAttributesLong, workflowLong}
	for i := range cmdNouns {
		assert.Equal(t, cmdNouns[i].Use, useArray[i])
		assert.Equal(t, cmdNouns[i].Aliases, aliases[i])
		assert.Equal(t, cmdNouns[i].Short, shortArray[i])
		assert.Equal(t, cmdNouns[i].Long, longArray[i])
	}
}
