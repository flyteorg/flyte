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
		return nExecDetailsForView[i].NodeExec.Closure.CreatedAt.AsTime().Before(nExecDetailsForView[j].NodeExec.Closure.CreatedAt.AsTime())
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
	for _, nodeExec := range nExecDetails.NodeExecutions {
		nodeExecClosure := &NodeExecutionClosure{
			NodeExec: &NodeExecution{nodeExec},
		}
		nodeExecClosures = append(nodeExecClosures, nodeExecClosure)

		// Check if this is parent node. If yes do recursive call to get child nodes.
		if nodeExec.Metadata != nil && nodeExec.Metadata.IsParentNode {
			nodeExecClosure.ChildNodes, err = getNodeExecDetailsInt(ctx, project, domain, execName, nodeName, nodeExec.Id.NodeId, nodeExecDetailsMap, cmdCtx)
			if err != nil {
				return nil, err
			}
		} else {
			taskExecList, err := cmdCtx.AdminFetcherExt().FetchTaskExecutionsOnNode(ctx,
				nodeExec.Id.NodeId, execName, project, domain)
			if err != nil {
				return nil, err
			}
			for _, taskExec := range taskExecList.TaskExecutions {
				taskExecClosure := &TaskExecutionClosure{
					TaskExecution: &TaskExecution{taskExec},
				}
				nodeExecClosure.TaskExecutions = append(nodeExecClosure.TaskExecutions, taskExecClosure)
			}
			// Fetch the node inputs and outputs
			nExecDataResp, err := cmdCtx.AdminFetcherExt().FetchNodeExecutionData(ctx, nodeExec.Id.NodeId, execName, project, domain)
			if err != nil {
				return nil, err
			}
			// Extract the inputs from the literal map
			nodeExecClosure.Inputs, err = extractLiteralMap(nExecDataResp.FullInputs)
			if err != nil {
				return nil, err
			}
			// Extract the outputs from the literal map
			nodeExecClosure.Outputs, err = extractLiteralMap(nExecDataResp.FullOutputs)
			if err != nil {
				return nil, err
			}
		}
		nodeExecDetailsMap[nodeExec.Id.NodeId] = nodeExecClosure
		// Found the node
		if len(nodeName) > 0 && nodeName == nodeExec.Id.NodeId {
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
		return taskExecClosures[i].Id.RetryAttempt < taskExecClosures[j].Id.RetryAttempt
	})
	for _, taskExecClosure := range taskExecClosures {
		attemptView := rootView.Add(taskAttemptPrefix + strconv.Itoa(int(taskExecClosure.Id.RetryAttempt)))
		attemptView.Add(taskExecPrefix + taskExecClosure.Closure.Phase.String() +
			hyphenPrefix + taskExecClosure.Closure.CreatedAt.AsTime().String() +
			hyphenPrefix + taskExecClosure.Closure.UpdatedAt.AsTime().String())
		attemptView.Add(taskTypePrefix + taskExecClosure.Closure.TaskType)
		attemptView.Add(taskReasonPrefix + taskExecClosure.Closure.Reason)
		if taskExecClosure.Closure.Metadata != nil {
			metadata := attemptView.Add(taskMetadataPrefix)
			metadata.Add(taskGeneratedNamePrefix + taskExecClosure.Closure.Metadata.GeneratedName)
			metadata.Add(taskPluginIDPrefix + taskExecClosure.Closure.Metadata.PluginIdentifier)
			extResourcesView := metadata.Add(taskExtResourcesPrefix)
			for _, extResource := range taskExecClosure.Closure.Metadata.ExternalResources {
				extResourcesView.Add(taskExtResourcePrefix + extResource.ExternalId)
			}
			resourcePoolInfoView := metadata.Add(taskResourcePrefix)
			for _, rsPool := range taskExecClosure.Closure.Metadata.ResourcePoolInfo {
				resourcePoolInfoView.Add(taskExtResourcePrefix + rsPool.Namespace)
				resourcePoolInfoView.Add(taskExtResourceTokenPrefix + rsPool.AllocationToken)
			}
		}

		sort.Slice(taskExecClosure.Closure.Logs[:], func(i, j int) bool {
			return taskExecClosure.Closure.Logs[i].Name < taskExecClosure.Closure.Logs[j].Name
		})

		logsView := attemptView.Add(taskLogsPrefix)
		for _, logData := range taskExecClosure.Closure.Logs {
			logsView.Add(taskLogsNamePrefix + logData.Name)
			logsView.Add(taskLogURIPrefix + logData.Uri)
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
		return nodeExecutionClosures[i].NodeExec.Closure.CreatedAt.AsTime().Before(nodeExecutionClosures[j].NodeExec.Closure.CreatedAt.AsTime())
	})

	for _, nodeExecWrapper := range nodeExecutionClosures {
		nExecView := rootView.Add(nodeExecWrapper.NodeExec.Id.NodeId + hyphenPrefix + nodeExecWrapper.NodeExec.Closure.Phase.String() +
			hyphenPrefix + nodeExecWrapper.NodeExec.Closure.CreatedAt.AsTime().String() +
			hyphenPrefix + nodeExecWrapper.NodeExec.Closure.UpdatedAt.AsTime().String())
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
	for key, literalVal := range literalMap.Literals {
		extractedLiteralVal, err := coreutils.ExtractFromLiteral(literalVal)
		if err != nil {
			return nil, err
		}
		m[key] = extractedLiteralVal
	}
	return m, nil
}
