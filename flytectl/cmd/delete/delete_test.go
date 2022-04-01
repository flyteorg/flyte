package delete

import (
	"sort"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/stretchr/testify/assert"
)

const (
	testDataNonExistentFile = "testdata/non-existent-file"
	testDataInvalidAttrFile = "testdata/invalid_attribute.yaml"
)

var setup = testutils.Setup
var tearDownAndVerify = testutils.TearDownAndVerify

func TestDeleteCommand(t *testing.T) {
	deleteCommand := RemoteDeleteCommand()
	assert.Equal(t, deleteCommand.Use, "delete")
	assert.Equal(t, deleteCommand.Short, deleteCmdShort)
	assert.Equal(t, deleteCommand.Long, deleteCmdLong)
	assert.Equal(t, len(deleteCommand.Commands()), 7)
	cmdNouns := deleteCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	useArray := []string{"cluster-resource-attribute", "execution", "execution-cluster-label", "execution-queue-attribute", "plugin-override", "task-resource-attribute", "workflow-execution-config"}
	aliases := [][]string{{"cluster-resource-attributes"}, {"executions"}, {"execution-cluster-labels"}, {"execution-queue-attributes"}, {"plugin-overrides"}, {"task-resource-attributes"}, {"workflow-execution-config"}}
	shortArray := []string{clusterResourceAttributesShort, execCmdShort, executionClusterLabelShort, executionQueueAttributesShort, pluginOverrideShort, taskResourceAttributesShort, workflowExecutionConfigShort}
	longArray := []string{clusterResourceAttributesLong, execCmdLong, executionClusterLabelLong, executionQueueAttributesLong, pluginOverrideLong, taskResourceAttributesLong, workflowExecutionConfigLong}
	for i := range cmdNouns {
		assert.Equal(t, cmdNouns[i].Use, useArray[i])
		assert.Equal(t, cmdNouns[i].Aliases, aliases[i])
		assert.Equal(t, cmdNouns[i].Short, shortArray[i])
		assert.Equal(t, cmdNouns[i].Long, longArray[i])
	}
}
