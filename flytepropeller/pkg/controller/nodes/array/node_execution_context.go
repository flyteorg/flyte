package array

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
)

type staticInputReader struct {
	io.InputFilePaths
	input *core.LiteralMap
}

func (i staticInputReader) Get(_ context.Context) (*core.LiteralMap, error) {
	return i.input, nil
}

func newStaticInputReader(inputPaths io.InputFilePaths, input *core.LiteralMap) staticInputReader {
	return staticInputReader{
		InputFilePaths: inputPaths,
		input:          input,
	}
}

func constructLiteralMap(inputs *core.LiteralMap, index int) (*core.LiteralMap, error) {
	literals := make(map[string]*core.Literal)
	for name, literal := range inputs.Literals {
		if literalCollection := literal.GetCollection(); literalCollection != nil {
			if index >= len(literalCollection.Literals) {
				return nil, fmt.Errorf("index %v out of bounds for literal collection %v", index, name)
			}
			literals[name] = literalCollection.Literals[index]
		} else {
			literals[name] = literal
		}
	}

	return &core.LiteralMap{
		Literals: literals,
	}, nil
}

type arrayTaskReader struct {
	interfaces.TaskReader
}

func (a *arrayTaskReader) Read(ctx context.Context) (*core.TaskTemplate, error) {
	originalTaskTemplate, err := a.TaskReader.Read(ctx)
	if err != nil {
		return nil, err
	}

	// convert output list variable to singular
	outputVariables := make(map[string]*core.Variable)
	for key, value := range originalTaskTemplate.Interface.Outputs.Variables {
		switch v := value.Type.Type.(type) {
		case *core.LiteralType_CollectionType:
			outputVariables[key] = &core.Variable{
				Type:        v.CollectionType,
				Description: value.Description,
			}
		default:
			outputVariables[key] = value
		}
	}

	taskTemplate := *originalTaskTemplate
	taskTemplate.Interface = &core.TypedInterface{
		Inputs: originalTaskTemplate.Interface.Inputs,
		Outputs: &core.VariableMap{
			Variables: outputVariables,
		},
	}
	return &taskTemplate, nil
}

type arrayNodeExecutionContext struct {
	interfaces.NodeExecutionContext
	eventRecorder    arrayEventRecorder
	executionContext executors.ExecutionContext
	inputReader      io.InputReader
	nodeStatus       *v1alpha1.NodeStatus
	taskReader       interfaces.TaskReader
}

func (a *arrayNodeExecutionContext) EventsRecorder() interfaces.EventRecorder {
	return a.eventRecorder
}

func (a *arrayNodeExecutionContext) ExecutionContext() executors.ExecutionContext {
	return a.executionContext
}

func (a *arrayNodeExecutionContext) InputReader() io.InputReader {
	return a.inputReader
}

func (a *arrayNodeExecutionContext) NodeStatus() v1alpha1.ExecutableNodeStatus {
	return a.nodeStatus
}

func (a *arrayNodeExecutionContext) TaskReader() interfaces.TaskReader {
	return a.taskReader
}

func newArrayNodeExecutionContext(nodeExecutionContext interfaces.NodeExecutionContext, inputReader io.InputReader,
	eventRecorder arrayEventRecorder, subNodeIndex int, nodeStatus *v1alpha1.NodeStatus) *arrayNodeExecutionContext {

	arrayExecutionContext := newArrayExecutionContext(nodeExecutionContext.ExecutionContext(), subNodeIndex)
	return &arrayNodeExecutionContext{
		NodeExecutionContext: nodeExecutionContext,
		eventRecorder:        eventRecorder,
		executionContext:     arrayExecutionContext,
		inputReader:          inputReader,
		nodeStatus:           nodeStatus,
		taskReader:           &arrayTaskReader{nodeExecutionContext.TaskReader()},
	}
}
