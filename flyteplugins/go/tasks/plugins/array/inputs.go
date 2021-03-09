package array

import (
	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flytestdlib/storage"
)

// A proxy inputreader that overrides the inputpath to be the inputpathprefix for array jobs
type arrayJobInputReader struct {
	io.InputReader
}

// We override the inputpath to return the prefix path for array jobs
func (i arrayJobInputReader) GetInputPath() storage.DataReference {
	return i.GetInputPrefixPath()
}

func GetInputReader(tCtx core.TaskExecutionContext, taskTemplate *idlCore.TaskTemplate) io.InputReader {
	var inputReader io.InputReader
	if taskTemplate.GetTaskTypeVersion() == 0 {
		// Prior to task type version == 1, dynamic type tasks (including array tasks) would write input files for each
		// individual array task instance. In this case we use a modified input reader to only pass in the parent input
		// directory.
		inputReader = arrayJobInputReader{tCtx.InputReader()}
	} else {
		inputReader = tCtx.InputReader()
	}
	return inputReader
}
