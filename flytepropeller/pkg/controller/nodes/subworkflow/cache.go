package subworkflow

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/validators"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
)

type injectedInputReader struct {
	io.InputReader
	injectedLiterals *core.LiteralMap
}

func (i *injectedInputReader) Get(ctx context.Context) (*core.LiteralMap, error) {
	literalMap, err := i.InputReader.Get(ctx)
	if err != nil {
		return nil, err
	}

	// create LiteralMap with injected literals
	size := 0
	if literalMap != nil {
		size += len(literalMap.Literals)
	}
	if i.injectedLiterals != nil {
		size += len(i.injectedLiterals.Literals)
	}

	literals := make(map[string]*core.Literal, size)
	if literalMap != nil {
		for name, literal := range literalMap.Literals {
			literals[name] = literal
		}
	}
	if i.injectedLiterals != nil {
		for name, literal := range i.injectedLiterals.Literals {
			literals[name] = literal
		}
	}

	return &core.LiteralMap{
		Literals: literals,
	}, nil
}

func (w *workflowNodeHandler) GetCatalogKey(ctx context.Context, nCtx interfaces.NodeExecutionContext) (catalog.Key, error) {
	wfNode := nCtx.Node().GetWorkflowNode()

	// identify the subworkflow / launch plan identifier and input / output interface
	var identifier *core.Identifier
	var typedInterface *core.TypedInterface
	var inputReader io.InputReader
	if wfNode.GetSubWorkflowRef() != nil {
		// retrieve the subworkflow spec
		subWorkflow := nCtx.ExecutionContext().FindSubWorkflow(*wfNode.GetSubWorkflowRef())
		if subWorkflow == nil {
			return catalog.Key{}, errors.Errorf(errors.BadSpecificationError, nCtx.NodeID(), "subworkflow not found [%v]", wfNode.GetSubWorkflowRef())
		}

		identifier = subWorkflow.GetIdentifier()
		typedInterface = subWorkflow.GetInterface()
		inputReader = nCtx.InputReader()
	} else if wfNode.GetLaunchPlanRefID() != nil {
		// retrieve the launchplan spec
		launchPlanRefID := nCtx.Node().GetWorkflowNode().GetLaunchPlanRefID()
		launchPlan := nCtx.ExecutionContext().FindLaunchPlan(*launchPlanRefID)
		if launchPlan == nil {
			return catalog.Key{}, errors.Errorf(errors.BadSpecificationError, nCtx.NodeID(), "launch plan not found [%v]", launchPlanRefID)
		}

		// compile the input interface from the launchplan interface and fixed inputs
		inputInterface := make(map[string]*core.Variable, len(launchPlan.GetInterface().Inputs.Variables))
		for inputName, inputVariable := range launchPlan.GetInterface().Inputs.Variables {
			inputInterface[inputName] = inputVariable
		}
		if launchPlan.GetFixedInputs() != nil {
			for inputName, inputLiteral := range launchPlan.GetFixedInputs().Literals {
				inputInterface[inputName] = &core.Variable{
					Type: validators.LiteralTypeForLiteral(inputLiteral),
				}
			}
		}

		identifier = launchPlanRefID.Identifier
		typedInterface = &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: inputInterface,
			},
			Outputs: launchPlan.GetInterface().Outputs,
		}
		inputReader = &injectedInputReader{
			InputReader:      nCtx.InputReader(),
			injectedLiterals: launchPlan.GetFixedInputs(),
		}
	} else {
		return catalog.Key{}, errors.Errorf(errors.BadSpecificationError, nCtx.NodeID(), notExistsErrMsg)
	}

	return catalog.Key{
		Identifier:           *identifier,
		CacheVersion:         "",
		CacheIgnoreInputVars: nil,
		TypedInterface:       *typedInterface,
		InputReader:          inputReader,
	}, nil
}

func (w *workflowNodeHandler) IsCacheable(ctx context.Context, nCtx interfaces.NodeExecutionContext) (bool, bool, error) {
	// subworkflow / launch plan cachability is controlled using NodeMetadata overrides. therefore,
	// we always return false.
	return false, false, nil
}
