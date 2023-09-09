package bundle

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/stretchr/testify/assert"
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
	tID.On("GetID").Return(idlCore.TaskExecutionIdentifier{
		NodeExecutionId: &idlCore.NodeExecutionIdentifier{
			ExecutionId: &idlCore.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.On("GetTaskExecutionID").Return(tID)

	taskCtx := &mocks.TaskExecutionContext{}
	taskCtx.On("TaskExecutionMetadata").Return(taskExecutionMetadata)
	taskReader := &mocks.TaskReader{}
	taskReader.On("Read", mock.Anything).Return(&idlCore.TaskTemplate{
		Type: "unsupportedtype",
	}, nil)
	taskCtx.On("TaskReader").Return(taskReader)

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
