package delete

import (
	"context"
	"sort"
	"testing"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"

	"github.com/stretchr/testify/assert"
)

var (
	err        error
	ctx        context.Context
	mockClient *mocks.AdminServiceClient
	cmdCtx     cmdCore.CommandContext
)
var setup = testutils.Setup
var tearDownAndVerify = testutils.TearDownAndVerify

func TestDeleteCommand(t *testing.T) {
	deleteCommand := RemoteDeleteCommand()
	assert.Equal(t, deleteCommand.Use, "delete")
	assert.Equal(t, deleteCommand.Short, deleteCmdShort)
	assert.Equal(t, deleteCommand.Long, deleteCmdLong)
	assert.Equal(t, len(deleteCommand.Commands()), 3)
	cmdNouns := deleteCommand.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})
	useArray := []string{"cluster-resource-attribute", "execution", "task-resource-attribute"}
	aliases := [][]string{{"cluster-resource-attributes"}, {"executions"}, {"task-resource-attributes"}}
	shortArray := []string{clusterResourceAttributesShort, execCmdShort, taskResourceAttributesShort}
	longArray := []string{clusterResourceAttributesLong, execCmdLong, taskResourceAttributesLong}
	for i := range cmdNouns {
		assert.Equal(t, cmdNouns[i].Use, useArray[i])
		assert.Equal(t, cmdNouns[i].Aliases, aliases[i])
		assert.Equal(t, cmdNouns[i].Short, shortArray[i])
		assert.Equal(t, cmdNouns[i].Long, longArray[i])
	}
}
