package array

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/codex"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/k8s"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func appendLiteral(name string, literal *idlcore.Literal, outputLiterals map[string]*idlcore.Literal, length int) {
	outputLiteral, exists := outputLiterals[name]
	if !exists {
		outputLiteral = &idlcore.Literal{
			Value: &idlcore.Literal_Collection{
				Collection: &idlcore.LiteralCollection{
					Literals: make([]*idlcore.Literal, 0, length),
				},
			},
		}

		outputLiterals[name] = outputLiteral
	}

	collection := outputLiteral.GetCollection()
	collection.Literals = append(collection.Literals, literal)
}

func buildSubNodeID(nCtx interfaces.NodeExecutionContext, index int) string {
	return fmt.Sprintf("%s-n%d", nCtx.NodeID(), index)
}

func bytesFromK8sPluginState(pluginState k8s.PluginState) ([]byte, error) {
	buffer := make([]byte, 0, task.MaxPluginStateSizeBytes)
	bufferWriter := bytes.NewBuffer(buffer)

	codec := codex.GobStateCodec{}
	if err := codec.Encode(pluginState, bufferWriter); err != nil {
		return nil, err
	}

	return bufferWriter.Bytes(), nil
}

func constructOutputReferences(ctx context.Context, nCtx interfaces.NodeExecutionContext, subNodeIndex int, currentAttempt uint32) (storage.DataReference, storage.DataReference, error) {
	subDataDir, err := nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetDataDir(), []string{strconv.Itoa(subNodeIndex)}...)
	if err != nil {
		return "", "", err
	}

	subOutputDir, err := nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetOutputDir(), []string{strconv.Itoa(subNodeIndex), strconv.Itoa(int(currentAttempt))}...)
	if err != nil {
		return "", "", err
	}

	return subDataDir, subOutputDir, nil
}

func convertLiteralToBindingData(literal *idlcore.Literal) (*idlcore.BindingData, error) {
	bindingData := &idlcore.BindingData{}

	if literal.GetValue() == nil {
		return bindingData, nil
	}

	switch val := literal.GetValue().(type) {
	case *idlcore.Literal_Scalar:
		bindingData.Value = &idlcore.BindingData_Scalar{Scalar: val.Scalar}
	case *idlcore.Literal_Collection:
		convertedItems := make([]*idlcore.BindingData, len(val.Collection.Literals))
		for i, item := range val.Collection.Literals {
			bd, err := convertLiteralToBindingData(item)
			if err != nil {
				return nil, err
			}
			convertedItems[i] = bd
		}
		bindingData.Value = &idlcore.BindingData_Collection{
			Collection: &idlcore.BindingDataCollection{Bindings: convertedItems},
		}
	case *idlcore.Literal_Map:
		bdMap := &idlcore.BindingDataMap{Bindings: make(map[string]*idlcore.BindingData)}
		for key, literal := range val.Map.Literals {
			bd, err := convertLiteralToBindingData(literal)
			if err != nil {
				return nil, err
			}
			bdMap.Bindings[key] = bd
		}
		bindingData.Value = &idlcore.BindingData_Map{Map: bdMap}
	case *idlcore.Literal_OffloadedMetadata:
		bindingData.Value = &idlcore.BindingData_OffloadedMetadata{
			OffloadedMetadata: val.OffloadedMetadata,
		}
	default:
		return nil, fmt.Errorf("unsupported Literal value type: %T", val)
	}

	return bindingData, nil
}

func constructInputBindings(arrayNode v1alpha1.ExecutableArrayNode, inputLiteralMap *idlcore.LiteralMap) ([]*v1alpha1.Binding, error) {
	var inputBindings []*v1alpha1.Binding
	for _, binding := range arrayNode.GetSubNodeSpec().GetInputBindings() {
		key := inputLiteralMap.GetLiterals()[binding.GetVar()]
		if key == nil {
			return nil, fmt.Errorf("binding [%s] not found in input literal map", binding.GetVar())
		}
		bindingData, err := convertLiteralToBindingData(key)
		if err != nil {
			return nil, err
		}
		inputBinding := &v1alpha1.Binding{
			Binding: &idlcore.Binding{
				Var:     binding.GetVar(),
				Binding: bindingData,
			},
		}
		inputBindings = append(inputBindings, inputBinding)
	}
	return inputBindings, nil
}

func inferParallelism(ctx context.Context, parallelism *uint32, parallelismBehavior string, remainingWorkflowParallelism, arrayNodeSize int) (bool, int) {
	if parallelism != nil && *parallelism > 0 {
		// if parallelism is not defaulted - use it
		return false, int(*parallelism)
	} else if parallelismBehavior == config.ParallelismBehaviorWorkflow || (parallelism == nil && parallelismBehavior == config.ParallelismBehaviorHybrid) {
		// if workflow level parallelism
		return true, remainingWorkflowParallelism
	} else if parallelismBehavior == config.ParallelismBehaviorUnlimited ||
		(parallelism != nil && *parallelism == 0 && parallelismBehavior == config.ParallelismBehaviorHybrid) {
		// if unlimited parallelism
		return false, arrayNodeSize
	}

	logger.Warnf(ctx, "unable to infer ArrayNode parallelism configuration for parallelism:%v behavior:%v, defaulting to unlimited parallelism",
		parallelism, parallelismBehavior)
	return false, arrayNodeSize
}

func isTerminalNodePhase(nodePhase v1alpha1.NodePhase) bool {
	return nodePhase == v1alpha1.NodePhaseSucceeded || nodePhase == v1alpha1.NodePhaseFailed || nodePhase == v1alpha1.NodePhaseTimedOut ||
		nodePhase == v1alpha1.NodePhaseSkipped || nodePhase == v1alpha1.NodePhaseRecovered
}

func shouldIncrementTaskPhaseVersion(subNodeStatus *v1alpha1.NodeStatus, previousNodePhase v1alpha1.NodePhase, previousTaskPhase int) bool {
	if subNodeStatus.GetPhase() != previousNodePhase {
		return true
	} else if subNodeStatus.GetTaskNodeStatus() != nil &&
		subNodeStatus.GetTaskNodeStatus().GetPhase() != previousTaskPhase {
		return true
	}
	return subNodeStatus.GetWorkflowNodeStatus() != nil &&
		subNodeStatus.GetWorkflowNodeStatus().GetWorkflowNodePhase() != v1alpha1.WorkflowNodePhase(previousTaskPhase)
}
