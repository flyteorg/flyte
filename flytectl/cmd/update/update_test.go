package update

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDataNonExistentFile = "testdata/non-existent-file"
	testDataInvalidAttrFile = "testdata/invalid_attribute.yaml"
)

func TestUpdateCommand(t *testing.T) {
	updateCommand := CreateUpdateCommand()
	assert.Equal(t, updateCommand.Use, updateUse)
	assert.Equal(t, updateCommand.Short, updateShort)
	assert.Equal(t, updateCommand.Long, updatecmdLong)
	assert.Equal(t, len(updateCommand.Commands()), 12)
	cmdNouns := updateCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	useArray := []string{"cluster-resource-attribute", "execution", "execution-cluster-label", "execution-queue-attribute", "launchplan",
		"launchplan-meta", "plugin-override", "project", "task-meta", "task-resource-attribute", "workflow-execution-config", "workflow-meta"}
	aliases := [][]string{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}}
	shortArray := []string{clusterResourceAttributesShort, updateExecutionShort, executionClusterLabelShort, executionQueueAttributesShort, updateLPShort, updateLPMetaShort,
		pluginOverrideShort, projectShort, updateTaskShort, taskResourceAttributesShort, workflowExecutionConfigShort, updateWorkflowShort}
	longArray := []string{clusterResourceAttributesLong, updateExecutionLong, executionClusterLabelLong, executionQueueAttributesLong, updateLPLong, updateLPMetaLong,
		pluginOverrideLong, projectLong, updateTaskLong, taskResourceAttributesLong, workflowExecutionConfigLong, updateWorkflowLong}
	for i := range cmdNouns {
		assert.Equal(t, cmdNouns[i].Use, useArray[i])
		assert.Equal(t, cmdNouns[i].Aliases, aliases[i])
		assert.Equal(t, cmdNouns[i].Short, shortArray[i])
		assert.Equal(t, cmdNouns[i].Long, longArray[i])
	}
}
