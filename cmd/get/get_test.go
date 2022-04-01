package get

import (
	"fmt"
	"sort"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
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

var setup = testutils.Setup
var tearDownAndVerify = testutils.TearDownAndVerify

const (
	testDataTempFile            = "temp-output-file"
	testDataNotExistentTempFile = "non-existent-dir/temp-output-file"
)

func TestCreateGetCommand(t *testing.T) {
	getCommand := CreateGetCommand()
	assert.Equal(t, getCommand.Use, "get")
	assert.Equal(t, getCommand.Short, "Fetches various Flyte resources such as tasks, workflows, launch plans, executions, and projects.")
	fmt.Println(getCommand.Commands())
	assert.Equal(t, len(getCommand.Commands()), 11)
	cmdNouns := getCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	useArray := []string{"cluster-resource-attribute", "execution", "execution-cluster-label",
		"execution-queue-attribute", "launchplan", "plugin-override", "project", "task", "task-resource-attribute", "workflow", "workflow-execution-config"}
	aliases := [][]string{{"cluster-resource-attributes"}, {"executions"}, {"execution-cluster-labels"},
		{"execution-queue-attributes"}, {"launchplans"}, {"plugin-overrides"}, {"projects"}, {"tasks"}, {"task-resource-attributes"}, {"workflows"}, {"workflow-execution-config"}}
	shortArray := []string{clusterResourceAttributesShort, executionShort, executionClusterLabelShort, executionQueueAttributesShort, launchPlanShort,
		pluginOverrideShort, projectShort, taskShort, taskResourceAttributesShort, workflowShort, workflowExecutionConfigShort}
	longArray := []string{clusterResourceAttributesLong, executionLong, executionClusterLabelLong, executionQueueAttributesLong, launchPlanLong,
		pluginOverrideLong, projectLong, taskLong, taskResourceAttributesLong, workflowLong, workflowExecutionConfigLong}
	for i := range cmdNouns {
		assert.Equal(t, cmdNouns[i].Use, useArray[i])
		assert.Equal(t, cmdNouns[i].Aliases, aliases[i])
		assert.Equal(t, cmdNouns[i].Short, shortArray[i])
		assert.Equal(t, cmdNouns[i].Long, longArray[i])
	}
}
