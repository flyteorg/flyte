package array

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCoreMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	pluginsIOMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
)

func TestGetInputReader(t *testing.T) {

	inputReader := &pluginsIOMock.InputReader{}
	inputReader.OnGetInputPrefixPath().Return("test-data-prefix")
	inputReader.OnGetInputPathMatch(mock.Anything).Return("test-data-reference", nil)
	inputReader.OnGetInputDataPath().Return("test-data-reference")
	inputReader.OnGetMatch(mock.Anything).Return(&core.InputData{}, nil)

	t.Run("task_type_version == 0", func(t *testing.T) {
		taskCtx := &pluginsCoreMock.TaskExecutionContext{}
		taskCtx.On("InputReader").Return(inputReader)

		inputReader := GetInputReader(taskCtx, &core.TaskTemplate{
			TaskTypeVersion: 0,
		})
		assert.Equal(t, "test-data-prefix", inputReader.GetInputDataPath().String())
	})

	t.Run("task_type_version == 1", func(t *testing.T) {
		taskCtx := &pluginsCoreMock.TaskExecutionContext{}
		taskCtx.On("InputReader").Return(inputReader)

		inputReader := GetInputReader(taskCtx, &core.TaskTemplate{
			TaskTypeVersion: 1,
		})
		assert.Equal(t, "test-data-reference", inputReader.GetInputDataPath().String())
	})
}
