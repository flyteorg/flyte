package update

import (
	"bytes"
	"context"
	"os"
	"sort"
	"testing"

	"github.com/flyteorg/flyte/flytectl/cmd/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testDataNonExistentFile = "testdata/non-existent-file"
	testDataInvalidAttrFile = "testdata/invalid_attribute.yaml"
)

type MockSubCommand struct {
	mock.Mock
}

func (m *MockSubCommand) RunE(cmd *cobra.Command, args []string) error {
	m.Called(cmd, args).Get(0)
	return nil
}

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

var testYAML = `
---
kind: workflow-execution-config
domain: production
max_parallelism: 5
project: flytesnacks
raw_output_data_config:
  output_location_prefix: s3://example-data
security_context:
  run_as:
    k8s_service_account: not-demo-1
---
kind: project
name: flytesnacks-new
id: flytesnacks
---
kind: task-resource-attribute
domain: development
project: flytesnacks
defaults:
  cpu: "1"
  memory: "150Mi"
limits:
  cpu: "2"
  memory: "450Mi"
---
kind: cluster-resource-attribute
domain: development
project: flytesnacks
attributes:
  foo: "bar"
  buzz: "lightyear"
---
kind: execution-queue-attribute
domain: development
project: flytesnacks
tags:
  - foo
  - bar
  - buzz
  - lightyear
---
kind: execution-cluster-label
domain: development
project: flytesnacks
value: foo
---
kind: plugin-override
domain: development
project: flytesnacks
overrides:
   - task_type: python_task # Task type for which to apply plugin implementation overrides
     plugin_id:             # Plugin id(s) to be used in place of the default for the task type.
       - plugin_override1
       - plugin_override2
     missing_plugin_behavior: 1 # Behavior when no specified plugin_id has an associated handler. 0 : FAIL , 1: DEFAULT
---
# this is not supported and should raise warning
kind: execution
name: foo
activate: True
`

func TestRunUpdateFiles(t *testing.T) {
	// Create a test YAML file with multiple kinds
	tmpFile, err := os.CreateTemp("", "test*.yaml")
	assert.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	_, err = tmpFile.Write([]byte(testYAML))
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	// Create mock subcommands for each kind
	CreateUpdateCommand()
	mockCmds := make(map[string]*MockSubCommand)
	for kind := range configStructMap {
		mockCmd := &MockSubCommand{}
		mockCmd.On("RunE", mock.Anything, mock.Anything).Return(nil)
		subCmdMap[kind].RunE = mockCmd.RunE
		mockCmds[kind] = mockCmd
	}

	// Update the config to use the temporary file
	config.DefaultUpdateConfig.AttrFile = []string{tmpFile.Name()}

	// Capture logs
	var logBuffer bytes.Buffer
    logrus.SetOutput(&logBuffer)

	// Run the function
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	err = runUpdateFiles(cmd, []string{})
	assert.NoError(t, err)

    logrus.SetOutput(os.Stderr)  // Restore logger output

	// Verify that the subcommand functions were called
	for _, mockCmd := range mockCmds {
		mockCmd.AssertCalled(t, "RunE", mock.Anything, mock.Anything)
	}

	// Check for the warning log message
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Cannot pass config to subcommand 'execution' with file")
}
