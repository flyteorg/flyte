package array

import (
	"bytes"
	"context"
	"fmt"
	"time"

	idlcore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/codex"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/k8s"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/golang/protobuf/ptypes"
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

func buildTaskExecutionEvent(_ context.Context, nCtx interfaces.NodeExecutionContext, taskPhase idlcore.TaskExecution_Phase, taskPhaseVersion uint32, externalResources []*event.ExternalResourceInfo) (*event.TaskExecutionEvent, error) {
	occurredAt, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, err
	}

	nodeExecutionID := nCtx.NodeExecutionMetadata().GetNodeExecutionID()
	workflowExecutionID := nodeExecutionID.ExecutionId
	return &event.TaskExecutionEvent{
		TaskId: &idlcore.Identifier{
			ResourceType: idlcore.ResourceType_TASK,
			Project:      workflowExecutionID.Project,
			Domain:       workflowExecutionID.Domain,
			Name:         nCtx.NodeID(),
			Version:      "v1", // this value is irrelevant but necessary for the identifier to be valid
		},
		ParentNodeExecutionId: nCtx.NodeExecutionMetadata().GetNodeExecutionID(),
		RetryAttempt:          0, // ArrayNode will never retry
		Phase:                 taskPhase,
		PhaseVersion:          taskPhaseVersion,
		OccurredAt:            occurredAt,
		Metadata: &event.TaskExecutionMetadata{
			ExternalResources: externalResources,
		},
		TaskType:     "k8s-array",
		EventVersion: 1,
	}, nil
}

func buildSubNodeID(nCtx interfaces.NodeExecutionContext, index int, retryAttempt uint32) string {
	return fmt.Sprintf("%s-n%d-%d", nCtx.NodeID(), index, retryAttempt)
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

func isTerminalNodePhase(nodePhase v1alpha1.NodePhase) bool {
	return nodePhase == v1alpha1.NodePhaseSucceeded || nodePhase == v1alpha1.NodePhaseFailed || nodePhase == v1alpha1.NodePhaseTimedOut ||
		nodePhase == v1alpha1.NodePhaseSkipped || nodePhase == v1alpha1.NodePhaseRecovered
}
