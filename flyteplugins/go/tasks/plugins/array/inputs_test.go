package array

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCoreMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	pluginsIOMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestGetInputReader(t *testing.T) {

	inputReader := &pluginsIOMock.InputReader{}
	inputReader.On("GetInputPrefixPath").Return(storage.DataReference("test-data-prefix"))
	inputReader.On("GetInputPath").Return(storage.DataReference("test-data-reference"))
	inputReader.On("Get", mock.Anything).Return(&core.LiteralMap{}, nil)

	t.Run("task_type_version == 0", func(t *testing.T) {
		taskCtx := &pluginsCoreMock.TaskExecutionContext{}
		taskCtx.On("InputReader").Return(inputReader)

		inputReader := GetInputReader(taskCtx, &core.TaskTemplate{
			TaskTypeVersion: 0,
		})
		assert.Equal(t, inputReader.GetInputPath().String(), "test-data-prefix")
	})

	t.Run("task_type_version == 1", func(t *testing.T) {
		taskCtx := &pluginsCoreMock.TaskExecutionContext{}
		taskCtx.On("InputReader").Return(inputReader)

		inputReader := GetInputReader(taskCtx, &core.TaskTemplate{
			TaskTypeVersion: 1,
		})
		assert.Equal(t, inputReader.GetInputPath().String(), "test-data-reference")
	})
}
