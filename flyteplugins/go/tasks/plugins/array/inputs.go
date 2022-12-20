package array

import (
	"context"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flytestdlib/storage"
)

// arrayJobInputReader is a proxy inputreader that overrides the inputpath to be the inputpathprefix for array jobs
type arrayJobInputReader struct {
	io.InputReader
}

// GetInputPath overrides the inputpath to return the prefix path for array jobs
func (i arrayJobInputReader) GetInputPath() storage.DataReference {
	return i.GetInputPrefixPath()
}

func GetInputReader(tCtx core.TaskExecutionContext, taskTemplate *idlCore.TaskTemplate) io.InputReader {
	if taskTemplate.GetTaskTypeVersion() == 0 && taskTemplate.Type != AwsBatchTaskType {
		// Prior to task type version == 1, dynamic type tasks (including array tasks) would write input files for each
		// individual array task instance. In this case we use a modified input reader to only pass in the parent input
		// directory.
		return arrayJobInputReader{tCtx.InputReader()}
	}

	return tCtx.InputReader()
}

// StaticInputReader complies with the io.InputReader interface but has the input already populated.
type StaticInputReader struct {
	io.InputFilePaths
	input *idlCore.LiteralMap
}

func NewStaticInputReader(inputPaths io.InputFilePaths, input *idlCore.LiteralMap) StaticInputReader {
	return StaticInputReader{
		InputFilePaths: inputPaths,
		input:          input,
	}
}

func (i StaticInputReader) Get(_ context.Context) (*idlCore.LiteralMap, error) {
	return i.input, nil
}
