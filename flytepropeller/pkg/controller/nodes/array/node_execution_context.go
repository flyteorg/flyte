package array

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"

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

// SubNodeInputResolver caches parent inputs and offloaded literals to avoid redundant reads
// when constructing inputs for each subnode in an array node.
type SubNodeInputResolver struct {
	nCtx                 interfaces.NodeExecutionContext
	arrayNode            v1alpha1.ExecutableArrayNode
	parentInputs         *core.LiteralMap
	offloadedCollections map[string]*core.Literal
}

func newSubNodeInputResolver(nCtx interfaces.NodeExecutionContext, arrayNode v1alpha1.ExecutableArrayNode) *SubNodeInputResolver {
	return &SubNodeInputResolver{
		nCtx:      nCtx,
		arrayNode: arrayNode,
	}
}

// Initialize reads parent inputs and downloads any offloaded collection literals once.
func (r *SubNodeInputResolver) Initialize(ctx context.Context) error {
	inputs, err := r.nCtx.InputReader().Get(ctx)
	if err != nil {
		return err
	}
	r.parentInputs = inputs

	r.offloadedCollections = make(map[string]*core.Literal)
	for name, literal := range inputs.Literals {
		if slices.Contains(r.arrayNode.GetBoundInputs(), name) {
			continue
		}
		if literal.GetOffloadedMetadata().GetInferredType().GetCollectionType() != nil {
			downloadedLiteral := proto.Clone(literal).(*core.Literal)
			if err := common.ReadLargeLiteral(ctx, r.nCtx.DataStore(), downloadedLiteral); err != nil {
				return err
			}
			if downloadedLiteral.GetCollection() == nil {
				return fmt.Errorf("expected a collection literal for %s", name)
			}
			r.offloadedCollections[name] = downloadedLiteral
		}
	}
	return nil
}

// GetSubNodeInputs returns the input reader and bindings for a specific subnode index.
func (r *SubNodeInputResolver) GetSubNodeInputs(ctx context.Context, index int, subDataDir storage.DataReference) (staticInputReader, []*v1alpha1.Binding, error) {
	inputLiteralMap, err := r.constructSubNodeLiteralMap(ctx, index)
	if err != nil {
		return staticInputReader{}, nil, err
	}

	var subNodeInputReader staticInputReader
	var subNodeInputBindings []*v1alpha1.Binding

	switch r.arrayNode.GetDataMode() {
	case core.ArrayNode_INDIVIDUAL_INPUT_FILES:
		inputFilePath := ioutils.NewInputFilePaths(ctx, r.nCtx.DataStore(), subDataDir)
		subNodeInputReader = newStaticInputReader(inputFilePath, inputLiteralMap)
		subNodeInputBindings, err = constructInputBindings(r.arrayNode, inputLiteralMap)
		if err != nil {
			return staticInputReader{}, nil, err
		}
	case core.ArrayNode_SINGLE_INPUT_FILE:
		subNodeInputReader = newStaticInputReader(r.nCtx.InputReader(), inputLiteralMap)
		// mock the input bindings for the subNode to nil to bypass input resolution in the
		// `nodeExecutor.preExecute` function. this is required because this function is the entrypoint
		// for initial cache lookups.
		subNodeInputBindings = nil
	default:
		return staticInputReader{}, nil, fmt.Errorf("unsupported data mode [%v]", r.arrayNode.GetDataMode())
	}

	return subNodeInputReader, subNodeInputBindings, nil
}

// constructSubNodeLiteralMap builds the literal map for a specific subnode by indexing into cached data.
func (r *SubNodeInputResolver) constructSubNodeLiteralMap(ctx context.Context, index int) (*core.LiteralMap, error) {
	literals := make(map[string]*core.Literal)
	for name, literal := range r.parentInputs.Literals {
		if slices.Contains(r.arrayNode.GetBoundInputs(), name) {
			literals[name] = literal
		} else if literalCollection := literal.GetCollection(); literalCollection != nil {
			if index >= len(literalCollection.Literals) {
				return nil, fmt.Errorf("index %v out of bounds for literal collection %v", index, name)
			}
			literals[name] = literalCollection.Literals[index]
		} else if downloadedLiteral, ok := r.offloadedCollections[name]; ok {
			if index >= len(downloadedLiteral.GetCollection().Literals) {
				return nil, fmt.Errorf("index %v out of bounds for literal collection %v", index, name)
			}
			literalIndexed := downloadedLiteral.GetCollection().Literals[index]

			switch r.arrayNode.GetDataMode() {
			case core.ArrayNode_SINGLE_INPUT_FILE:
				literalDigest, err := pbhash.ComputeHash(ctx, literalIndexed)
				if err != nil {
					return nil, err
				}
				literal.Hash = base64.RawURLEncoding.EncodeToString(literalDigest)
				literals[name] = literal
			case core.ArrayNode_INDIVIDUAL_INPUT_FILES:
				literals[name] = literalIndexed
			default:
				return nil, fmt.Errorf("unsupported data mode [%v]", r.arrayNode.GetDataMode())
			}
		} else {
			literals[name] = literal
		}
	}

	return &core.LiteralMap{Literals: literals}, nil
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
