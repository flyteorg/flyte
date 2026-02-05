package bundle

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	idlCore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

var testHandler = failFastHandler{}

func TestFailFastGetID(t *testing.T) {
	assert.Equal(t, "fail-fast", testHandler.GetID())
}

func TestGetProperties(t *testing.T) {
	assert.Empty(t, testHandler.GetProperties())
}

func TestHandleAlwaysFails(t *testing.T) {
	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetID().Return(idlCore.TaskExecutionIdentifier{
		NodeExecutionId: &idlCore.NodeExecutionIdentifier{
			ExecutionId: &idlCore.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.EXPECT().GetTaskExecutionID().Return(tID)

	taskCtx := &mocks.TaskExecutionContext{}
	taskCtx.EXPECT().TaskExecutionMetadata().Return(taskExecutionMetadata)
	taskReader := &mocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(&idlCore.TaskTemplate{
		Type: "unsupportedtype",
	}, nil)
	taskCtx.EXPECT().TaskReader().Return(taskReader)

	transition, err := testHandler.Handle(context.TODO(), taskCtx)
	assert.NoError(t, err)
	assert.Equal(t, core.PhasePermanentFailure, transition.Info().Phase())
	assert.Equal(t, "AlwaysFail", transition.Info().Err().Code)
	assert.Contains(t, transition.Info().Err().Message, "Task [unsupportedtype]")
}

func TestAbort(t *testing.T) {
	err := testHandler.Abort(context.TODO(), nil)
	assert.NoError(t, err)
}

func TestFinalize(t *testing.T) {
	err := testHandler.Finalize(context.TODO(), nil)
	assert.NoError(t, err)
}
