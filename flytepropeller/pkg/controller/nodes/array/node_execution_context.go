package array

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
)

type arrayEventRecorder struct {
	nodeEvents []*event.NodeExecutionEvent
	taskEvents []*event.TaskExecutionEvent
}

func (a *arrayEventRecorder) RecordNodeEvent(ctx context.Context, event *event.NodeExecutionEvent, eventConfig *config.EventConfig) error {
	a.nodeEvents = append(a.nodeEvents, event)
	return nil
}

func (a *arrayEventRecorder) RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	a.taskEvents = append(a.taskEvents, event)
	return nil
}

func (a *arrayEventRecorder) NodeEvents() []*event.NodeExecutionEvent {
	return a.nodeEvents
}

func (a *arrayEventRecorder) TaskEvents() []*event.TaskExecutionEvent {
	return a.taskEvents
}

func newArrayEventRecorder() *arrayEventRecorder {
	return &arrayEventRecorder{
		nodeEvents: make([]*event.NodeExecutionEvent, 0),
		taskEvents: make([]*event.TaskExecutionEvent, 0),
	}
}

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

func constructLiteralMap(ctx context.Context, inputReader io.InputReader, index int) (*core.LiteralMap, error) {
	inputs, err := inputReader.Get(ctx)
	if err != nil {
		return nil, err
	}

	literals := make(map[string]*core.Literal)
	for name, literal := range inputs.Literals {
		if literalCollection := literal.GetCollection(); literalCollection != nil {
			literals[name] = literalCollection.Literals[index]
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
	taskTemplate, err := a.TaskReader.Read(ctx)
	if err != nil {
		return nil, err
	}

	// convert output list variable to singular
	outputVariables := make(map[string]*core.Variable)
	for key, value := range taskTemplate.Interface.Outputs.Variables {
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

	taskTemplate.Interface.Outputs = &core.VariableMap{
		Variables: outputVariables,
	}
	return taskTemplate, nil
}

type arrayNodeExecutionContext struct {
	interfaces.NodeExecutionContext
	eventRecorder    interfaces.EventRecorder
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

func newArrayNodeExecutionContext(nodeExecutionContext interfaces.NodeExecutionContext, inputReader io.InputReader, eventRecorder interfaces.EventRecorder, subNodeIndex int, nodeStatus *v1alpha1.NodeStatus, currentParallelism *uint32, maxParallelism uint32) *arrayNodeExecutionContext {
	arrayExecutionContext := newArrayExecutionContext(nodeExecutionContext.ExecutionContext(), subNodeIndex, currentParallelism, maxParallelism)
	return &arrayNodeExecutionContext{
		NodeExecutionContext: nodeExecutionContext,
		eventRecorder:        eventRecorder,
		executionContext:     arrayExecutionContext,
		inputReader:          inputReader,
		nodeStatus:           nodeStatus,
		taskReader:           &arrayTaskReader{nodeExecutionContext.TaskReader()},
	}
}
