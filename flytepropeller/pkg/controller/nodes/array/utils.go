package array

import (
	"bytes"
	"context"
	"fmt"

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

func constructOutputReferences(ctx context.Context, nCtx interfaces.NodeExecutionContext, postfix ...string) (storage.DataReference, storage.DataReference, error) {
	subDataDir, err := nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetDataDir(), postfix...)
	if err != nil {
		return "", "", err
	}

	subOutputDir, err := nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetOutputDir(), postfix...)
	if err != nil {
		return "", "", err
	}

	return subDataDir, subOutputDir, nil
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
