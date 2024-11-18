package get

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/disiqueira/gotree"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/pkg/printer"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

var nodeExecutionColumns = []printer.Column{
	{Header: "Name", JSONPath: "$.id.nodeID"},
	{Header: "Exec", JSONPath: "$.id.executionId.name"},
	{Header: "EndedAt", JSONPath: "$.endedAt"},
	{Header: "StartedAt", JSONPath: "$.startedAt"},
	{Header: "Phase", JSONPath: "$.phase"},
}

const (
	taskAttemptPrefix          = "Attempt :"
	taskExecPrefix             = "Task - "
	taskTypePrefix             = "Task Type - "
	taskReasonPrefix           = "Reason - "
	taskMetadataPrefix         = "Metadata"
	taskGeneratedNamePrefix    = "Generated Name : "
	taskPluginIDPrefix         = "Plugin Identifier : "
	taskExtResourcesPrefix     = "External Resources"
	taskExtResourcePrefix      = "Ext Resource : "
	taskExtResourceTokenPrefix = "Ext Resource Token : " //nolint
	taskResourcePrefix         = "Resource Pool Info"
	taskLogsPrefix             = "Logs :"
	outputsPrefix              = "Outputs :"
	taskLogsNamePrefix         = "Name :"
	taskLogURIPrefix           = "URI :"
	hyphenPrefix               = " - "
)

// TaskExecution wrapper around admin.TaskExecution
type TaskExecution struct {
	*admin.TaskExecution
}

// MarshalJSON overridden method to json marshalling to use jsonpb
func (in *TaskExecution) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.TaskExecution)
}

// UnmarshalJSON overridden method to json unmarshalling to use jsonpb
func (in *TaskExecution) UnmarshalJSON(b []byte) error {
	in.TaskExecution = &admin.TaskExecution{}
	return utils.UnmarshalBytesToPb(b, in.TaskExecution)
}

type NodeExecution struct {
	*admin.NodeExecution
}

// MarshalJSON overridden method to json marshalling to use jsonpb
func (in *NodeExecution) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.NodeExecution)
}

// UnmarshalJSON overridden method to json unmarshalling to use jsonpb
func (in *NodeExecution) UnmarshalJSON(b []byte) error {
	*in = NodeExecution{}
	return utils.UnmarshalBytesToPb(b, in.NodeExecution)
}

// NodeExecutionClosure forms a wrapper around admin.NodeExecution and also fetches the childnodes , task execs
// and input/output on the node executions from the admin api's.
type NodeExecutionClosure struct {
	NodeExec       *NodeExecution          `json:"node_exec,omitempty"`
	ChildNodes     []*NodeExecutionClosure `json:"child_nodes,omitempty"`
	TaskExecutions []*TaskExecutionClosure `json:"task_execs,omitempty"`
	// Inputs for the node
	Inputs map[string]interface{} `json:"inputs,omitempty"`
	// Outputs for the node
	Outputs map[string]interface{} `json:"outputs,omitempty"`
}

// TaskExecutionClosure wrapper around TaskExecution
type TaskExecutionClosure struct {
	*TaskExecution
}

func getExecutionDetails(ctx context.Context, project, domain, execName, nodeName string, cmdCtx cmdCore.CommandContext) ([]*NodeExecutionClosure, error) {
	// Fetching Node execution details
	nodeExecDetailsMap := map[string]*NodeExecutionClosure{}
	nExecDetails, err := getNodeExecDetailsInt(ctx, project, domain, execName, nodeName, "", nodeExecDetailsMap, cmdCtx)
	if err != nil {
		return nil, err
	}

	var nExecDetailsForView []*NodeExecutionClosure
	// Get the execution details only for the nodeId passed
	if len(nodeName) > 0 {
		// Fetch the last one which contains the nodeId details as previous ones are used to reach the nodeId
		if nodeExecDetailsMap[nodeName] != nil {
			nExecDetailsForView = append(nExecDetailsForView, nodeExecDetailsMap[nodeName])
		}
	} else {
		nExecDetailsForView = nExecDetails
	}

	sort.Slice(nExecDetailsForView[:], func(i, j int) bool {
		return nExecDetailsForView[i].NodeExec.Closure.GetCreatedAt().AsTime().Before(nExecDetailsForView[j].NodeExec.Closure.GetCreatedAt().AsTime())
	})

	return nExecDetailsForView, nil
}

func getNodeExecDetailsInt(ctx context.Context, project, domain, execName, nodeName, uniqueParentID string,
	nodeExecDetailsMap map[string]*NodeExecutionClosure, cmdCtx cmdCore.CommandContext) ([]*NodeExecutionClosure, error) {

	nExecDetails, err := cmdCtx.AdminFetcherExt().FetchNodeExecutionDetails(ctx, execName, project, domain, uniqueParentID)
	if err != nil {
		return nil, err
	}

	var nodeExecClosures []*NodeExecutionClosure
	for _, nodeExec := range nExecDetails.GetNodeExecutions() {
		nodeExecClosure := &NodeExecutionClosure{
			NodeExec: &NodeExecution{nodeExec},
		}
		nodeExecClosures = append(nodeExecClosures, nodeExecClosure)

		// Check if this is parent node. If yes do recursive call to get child nodes.
		if nodeExec.GetMetadata() != nil && nodeExec.GetMetadata().GetIsParentNode() {
			nodeExecClosure.ChildNodes, err = getNodeExecDetailsInt(ctx, project, domain, execName, nodeName, nodeExec.GetId().GetNodeId(), nodeExecDetailsMap, cmdCtx)
			if err != nil {
				return nil, err
			}
		} else {
			taskExecList, err := cmdCtx.AdminFetcherExt().FetchTaskExecutionsOnNode(ctx,
				nodeExec.GetId().GetNodeId(), execName, project, domain)
			if err != nil {
				return nil, err
			}
			for _, taskExec := range taskExecList.GetTaskExecutions() {
				taskExecClosure := &TaskExecutionClosure{
					TaskExecution: &TaskExecution{taskExec},
				}
				nodeExecClosure.TaskExecutions = append(nodeExecClosure.TaskExecutions, taskExecClosure)
			}
			// Fetch the node inputs and outputs
			nExecDataResp, err := cmdCtx.AdminFetcherExt().FetchNodeExecutionData(ctx, nodeExec.GetId().GetNodeId(), execName, project, domain)
			if err != nil {
				return nil, err
			}
			// Extract the inputs from the literal map
			nodeExecClosure.Inputs, err = extractLiteralMap(nExecDataResp.GetFullInputs())
			if err != nil {
				return nil, err
			}
			// Extract the outputs from the literal map
			nodeExecClosure.Outputs, err = extractLiteralMap(nExecDataResp.GetFullOutputs())
			if err != nil {
				return nil, err
			}
		}
		nodeExecDetailsMap[nodeExec.GetId().GetNodeId()] = nodeExecClosure
		// Found the node
		if len(nodeName) > 0 && nodeName == nodeExec.GetId().GetNodeId() {
			return nodeExecClosures, err
		}
	}
	return nodeExecClosures, nil
}

func createNodeTaskExecTreeView(rootView gotree.Tree, taskExecClosures []*TaskExecutionClosure) {
	if len(taskExecClosures) == 0 {
		return
	}
	if rootView == nil {
		rootView = gotree.New("")
	}
	// TODO: Replace this by filter to sort in the admin
	sort.Slice(taskExecClosures[:], func(i, j int) bool {
		return taskExecClosures[i].Id.GetRetryAttempt() < taskExecClosures[j].Id.GetRetryAttempt()
	})
	for _, taskExecClosure := range taskExecClosures {
		attemptView := rootView.Add(taskAttemptPrefix + strconv.Itoa(int(taskExecClosure.Id.GetRetryAttempt())))
		attemptView.Add(taskExecPrefix + taskExecClosure.Closure.GetPhase().String() +
			hyphenPrefix + taskExecClosure.Closure.GetCreatedAt().AsTime().String() +
			hyphenPrefix + taskExecClosure.Closure.GetUpdatedAt().AsTime().String())
		attemptView.Add(taskTypePrefix + taskExecClosure.Closure.GetTaskType())
		attemptView.Add(taskReasonPrefix + taskExecClosure.Closure.GetReason())
		if taskExecClosure.Closure.GetMetadata() != nil {
			metadata := attemptView.Add(taskMetadataPrefix)
			metadata.Add(taskGeneratedNamePrefix + taskExecClosure.Closure.GetMetadata().GetGeneratedName())
			metadata.Add(taskPluginIDPrefix + taskExecClosure.Closure.GetMetadata().GetPluginIdentifier())
			extResourcesView := metadata.Add(taskExtResourcesPrefix)
			for _, extResource := range taskExecClosure.Closure.GetMetadata().GetExternalResources() {
				extResourcesView.Add(taskExtResourcePrefix + extResource.GetExternalId())
			}
			resourcePoolInfoView := metadata.Add(taskResourcePrefix)
			for _, rsPool := range taskExecClosure.Closure.GetMetadata().GetResourcePoolInfo() {
				resourcePoolInfoView.Add(taskExtResourcePrefix + rsPool.GetNamespace())
				resourcePoolInfoView.Add(taskExtResourceTokenPrefix + rsPool.GetAllocationToken())
			}
		}

		sort.Slice(taskExecClosure.Closure.GetLogs()[:], func(i, j int) bool {
			return taskExecClosure.Closure.GetLogs()[i].GetName() < taskExecClosure.Closure.GetLogs()[j].GetName()
		})

		logsView := attemptView.Add(taskLogsPrefix)
		for _, logData := range taskExecClosure.Closure.GetLogs() {
			logsView.Add(taskLogsNamePrefix + logData.GetName())
			logsView.Add(taskLogURIPrefix + logData.GetUri())
		}
	}
}

func createNodeDetailsTreeView(rootView gotree.Tree, nodeExecutionClosures []*NodeExecutionClosure) gotree.Tree {
	if rootView == nil {
		rootView = gotree.New("")
	}
	if len(nodeExecutionClosures) == 0 {
		return rootView
	}
	// TODO : Move to sorting using filters.
	sort.Slice(nodeExecutionClosures[:], func(i, j int) bool {
		return nodeExecutionClosures[i].NodeExec.Closure.GetCreatedAt().AsTime().Before(nodeExecutionClosures[j].NodeExec.Closure.GetCreatedAt().AsTime())
	})

	for _, nodeExecWrapper := range nodeExecutionClosures {
		nExecView := rootView.Add(nodeExecWrapper.NodeExec.Id.GetNodeId() + hyphenPrefix + nodeExecWrapper.NodeExec.Closure.GetPhase().String() +
			hyphenPrefix + nodeExecWrapper.NodeExec.Closure.GetCreatedAt().AsTime().String() +
			hyphenPrefix + nodeExecWrapper.NodeExec.Closure.GetUpdatedAt().AsTime().String())
		if len(nodeExecWrapper.ChildNodes) > 0 {
			createNodeDetailsTreeView(nExecView, nodeExecWrapper.ChildNodes)
		}
		createNodeTaskExecTreeView(nExecView, nodeExecWrapper.TaskExecutions)
		if len(nodeExecWrapper.Outputs) > 0 {
			outputsView := nExecView.Add(outputsPrefix)
			for outputKey, outputVal := range nodeExecWrapper.Outputs {
				outputsView.Add(fmt.Sprintf("%s: %v", outputKey, outputVal))
			}
		}
	}
	return rootView
}

func extractLiteralMap(literalMap *core.LiteralMap) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	if literalMap == nil || literalMap.Literals == nil {
		return m, nil
	}
	for key, literalVal := range literalMap.GetLiterals() {
		extractedLiteralVal, err := coreutils.ExtractFromLiteral(literalVal)
		if err != nil {
			return nil, err
		}
		m[key] = extractedLiteralVal
	}
	return m, nil
}
