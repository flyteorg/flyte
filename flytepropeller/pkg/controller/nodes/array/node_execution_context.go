package array

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/pbhash"
	"github.com/flyteorg/flyte/flytestdlib/storage"
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

func constructLiteralMap(ctx context.Context, dataStore *storage.DataStore, arrayNode v1alpha1.ExecutableArrayNode, inputs *core.LiteralMap, index int) (*core.LiteralMap, error) {
	literals := make(map[string]*core.Literal)
	for name, literal := range inputs.Literals {
		if literalCollection := literal.GetCollection(); literalCollection != nil {
			if index >= len(literalCollection.Literals) {
				return nil, fmt.Errorf("index %v out of bounds for literal collection %v", index, name)
			}
			literals[name] = literalCollection.Literals[index]
		} else if literal.GetOffloadedMetadata().GetInferredType().GetCollectionType() != nil {
			// TODO - pvditt only do this if subNode is cacheable or DataMode is ArrayNode_INDIVIDUAL_INPUT_FILES
			downloadedLiteral := proto.Clone(literal).(*core.Literal)
			if err := common.ReadLargeLiteral(ctx, dataStore, downloadedLiteral); err != nil {
				return nil, err
			}
			if downloadedLiteral.GetCollection() == nil {
				return nil, fmt.Errorf("expected a collection literal")
			}
			if index >= len(downloadedLiteral.GetCollection().Literals) {
				return nil, fmt.Errorf("index %v out of bounds for literal collection %v", index, name)
			}
			literalIndexed := downloadedLiteral.GetCollection().Literals[index]

			switch arrayNode.GetDataMode() {
			case core.ArrayNode_SINGLE_INPUT_FILE:
				// set hash to the value of the literal at the given index
				literalDigest, err := pbhash.ComputeHash(ctx, literalIndexed)
				if err != nil {
					return nil, err
				}
				literal.Hash = base64.RawURLEncoding.EncodeToString(literalDigest)
				literals[name] = literal
			case core.ArrayNode_INDIVIDUAL_INPUT_FILES:
				// TODO - pvditt look into offloading if the literal is too large
				literals[name] = literalIndexed
			default:
				return nil, fmt.Errorf("unsupported data mode [%v]", arrayNode.GetDataMode())
			}
		} else {
			literals[name] = literal
		}
	}

	return &core.LiteralMap{
		Literals: literals,
	}, nil
}

func constructSubNodeInputs(ctx context.Context, nCtx interfaces.NodeExecutionContext, arrayNode v1alpha1.ExecutableArrayNode, subNodeIndex int, subDataDir storage.DataReference) (staticInputReader, []*v1alpha1.Binding, error) {
	var subNodeInputReader staticInputReader
	var subNodeInputBindings []*v1alpha1.Binding
	var err error

	// need to initialize the inputReader every time to ensure TaskHandler can access for cache lookups / population
	inputs, err := nCtx.InputReader().Get(ctx)
	if err != nil {
		return subNodeInputReader, subNodeInputBindings, err
	}

	inputLiteralMap, err := constructLiteralMap(ctx, nCtx.DataStore(), arrayNode, inputs, subNodeIndex)
	if err != nil {
		return subNodeInputReader, subNodeInputBindings, err
	}

	switch arrayNode.GetDataMode() {
	case core.ArrayNode_INDIVIDUAL_INPUT_FILES:
		inputFilePath := ioutils.NewInputFilePaths(
			ctx,
			nCtx.DataStore(),
			subDataDir,
		)

		subNodeInputReader = newStaticInputReader(inputFilePath, inputLiteralMap)
		subNodeInputBindings, err = constructInputBindings(arrayNode, inputLiteralMap)
		if err != nil {
			return subNodeInputReader, subNodeInputBindings, err
		}
	case core.ArrayNode_SINGLE_INPUT_FILE:
		subNodeInputReader = newStaticInputReader(nCtx.InputReader(), inputLiteralMap)
		// mock the input bindings for the subNode to nil to bypass input resolution in the
		// `nodeExecutor.preExecute` function. this is required because this function is the entrypoint
		// for initial cache lookups. an alternative solution would be to mock the datastore to bypass
		// writing the inputFile.
		subNodeInputBindings = nil
	default:
		return subNodeInputReader, subNodeInputBindings, fmt.Errorf("unsupported data mode [%v]", arrayNode.GetDataMode())
	}

	return subNodeInputReader, subNodeInputBindings, nil
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
	eventRecorder    ArrayEventRecorder
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
	eventRecorder ArrayEventRecorder, subNodeIndex int, nodeStatus *v1alpha1.NodeStatus) *arrayNodeExecutionContext {

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
