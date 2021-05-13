package get

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/workflow"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flytectl/pkg/ext/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getWorkflowSetup() {
	ctx = u.Ctx
	mockClient = u.MockClient
	cmdCtx = u.CmdCtx
	workflow.DefaultConfig.Latest = false
	workflow.DefaultConfig.Version = ""
}

func TestGetWorkflowFuncWithError(t *testing.T) {
	t.Run("failure fetch latest", func(t *testing.T) {
		setup()
		getWorkflowSetup()
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		workflow.DefaultConfig.Latest = true
		mockFetcher.OnFetchWorkflowLatestVersionMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("error fetching latest version"))
		_, err = FetchWorkflowForName(ctx, mockFetcher, "workflowName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching version ", func(t *testing.T) {
		setup()
		getWorkflowSetup()
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		workflow.DefaultConfig.Version = "v1"
		mockFetcher.OnFetchWorkflowVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("error fetching version"))
		_, err = FetchWorkflowForName(ctx, mockFetcher, "workflowName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching all version ", func(t *testing.T) {
		setup()
		getWorkflowSetup()
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		mockFetcher.OnFetchAllVerOfWorkflowMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("error fetching all version"))
		_, err = FetchWorkflowForName(ctx, mockFetcher, "workflowName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching ", func(t *testing.T) {
		setup()
		getWorkflowSetup()
		workflow.DefaultConfig.Latest = true
		args := []string{"workflowName"}
		u.FetcherExt.OnFetchWorkflowLatestVersionMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("error fetching latest version"))
		err = getWorkflowFunc(ctx, args, cmdCtx)
		assert.NotNil(t, err)
	})
}
