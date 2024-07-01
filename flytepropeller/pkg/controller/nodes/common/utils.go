package common

import (
	"context"
	"strconv"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const maxUniqueIDLength = 20

// GenerateUniqueID is the UniqueId of a node is unique within a given workflow execution.
// In order to achieve that we track the lineage of the node.
// To compute the uniqueID of a node, we use the uniqueID and retry attempt of the parent node
// For nodes in level 0, there is no parent, and parentInfo is nil
func GenerateUniqueID(parentInfo executors.ImmutableParentInfo, nodeID string) (string, error) {
	var parentUniqueID v1alpha1.NodeID
	var parentRetryAttempt string

	if parentInfo != nil {
		parentUniqueID = parentInfo.GetUniqueID()
		parentRetryAttempt = strconv.Itoa(int(parentInfo.CurrentAttempt()))
	}

	return encoding.FixedLengthUniqueIDForParts(maxUniqueIDLength, []string{parentUniqueID, parentRetryAttempt, nodeID})
}

// CreateParentInfo creates a unique parent id, the unique id of parent is dependent on the unique id and the current
// attempt of the grandparent to track the lineage.
func CreateParentInfo(grandParentInfo executors.ImmutableParentInfo, nodeID string, parentAttempt uint32, nodeIsDynamic bool) (executors.ImmutableParentInfo, error) {
	uniqueID, err := GenerateUniqueID(grandParentInfo, nodeID)
	if err != nil {
		return nil, err
	}
	if nodeIsDynamic || (grandParentInfo != nil && grandParentInfo.IsInDynamicChain()) {
		return executors.NewParentInfo(uniqueID, parentAttempt, true), nil
	}

	return executors.NewParentInfo(uniqueID, parentAttempt, false), nil
}

func GetTargetEntity(ctx context.Context, nCtx interfaces.NodeExecutionContext) *core.Identifier {
	var targetEntity *core.Identifier
	if nCtx.Node().GetWorkflowNode() != nil {
		subRef := nCtx.Node().GetWorkflowNode().GetSubWorkflowRef()
		if subRef != nil && len(*subRef) > 0 {
			// todo: uncomment this if Support caching subworkflows and launchplans (v2) is upstreamed
			// for now, we can leave it empty
			//nCtx.ExecutionContext().FindSubWorkflow(*subRef)
			//targetEntity = subWorkflow.GetIdentifier()
		} else if nCtx.Node().GetWorkflowNode().GetLaunchPlanRefID() != nil {
			lpRef := nCtx.Node().GetWorkflowNode().GetLaunchPlanRefID()
			targetEntity = lpRef.Identifier
		}
	} else if taskIDStr := nCtx.Node().GetTaskID(); taskIDStr != nil && len(*taskIDStr) > 0 {
		taskID, err := nCtx.ExecutionContext().GetTask(*taskIDStr)
		if err != nil {
			// This doesn't feed a very important part of the node execution event, swallow it for now.
			logger.Errorf(ctx, "Failed to get task [%v] with error [%v]", taskID, err)
		}
		targetEntity = taskID.CoreTask().Id
	}
	return targetEntity
}
