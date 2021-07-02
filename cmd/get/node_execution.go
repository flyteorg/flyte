package get

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/execution"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/printer"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/disiqueira/gotree"
	"github.com/golang/protobuf/proto"
)

var nodeExecutionColumns = []printer.Column{
	{Header: "Name", JSONPath: "$.id.nodeID"},
	{Header: "Exec", JSONPath: "$.id.executionId.name"},
	{Header: "Duration", JSONPath: "$.closure.duration"},
	{Header: "StartedAt", JSONPath: "$.closure.startedAt"},
	{Header: "Phase", JSONPath: "$.closure.phase"},
}

var taskExecutionColumns = []printer.Column{
	{Header: "Name", JSONPath: "$.id.taskId.name"},
	{Header: "Node ID", JSONPath: "$.id.nodeExecutionId.nodeID"},
	{Header: "Execution ID", JSONPath: "$.closure.nodeExecutionId.executionId.name"},
	{Header: "Duration", JSONPath: "$.closure.duration"},
	{Header: "StartedAt", JSONPath: "$.closure.startedAt"},
	{Header: "Phase", JSONPath: "$.closure.phase"},
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
	taskLogsNamePrefix         = "Name :"
	taskLogURIPrefix           = "URI :"
	hyphenPrefix               = " - "
)

func NodeExecutionToProtoMessages(l []*admin.NodeExecution) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func NodeTaskExecutionToProtoMessages(l []*admin.TaskExecution) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func getExecutionDetails(ctx context.Context, project, domain, name string, cmdCtx cmdCore.CommandContext) error {
	adminPrinter := printer.Printer{}

	// Fetching Node execution details
	nExecDetails, nodeExecToTaskExec, err := getNodeExecDetailsWithTasks(ctx, project, domain, name, cmdCtx)
	if err != nil {
		return err
	}

	// o/p format of table is not supported on the details. TODO: Add tree format in printer
	if config.GetConfig().MustOutputFormat() == printer.OutputFormatTABLE {
		fmt.Println("TABLE format is not supported on detailed view and defaults to tree view. Choose either json/yaml")
		if len(execution.DefaultConfig.NodeID) == 0 {
			nodeExecTree := createNodeDetailsTreeView(nExecDetails, nodeExecToTaskExec)
			if nodeExecTree != nil {
				fmt.Println(nodeExecTree.Print())
			}
		} else {
			nodeTaskExecTree := createNodeTaskExecTreeView(nil, nodeExecToTaskExec[execution.DefaultConfig.NodeID])
			if nodeTaskExecTree != nil {
				fmt.Println(nodeTaskExecTree.Print())
			}
		}
		return nil
	}

	if len(execution.DefaultConfig.NodeID) == 0 {
		return adminPrinter.Print(config.GetConfig().MustOutputFormat(), nodeExecutionColumns,
			NodeExecutionToProtoMessages(nExecDetails)...)
	}

	taskExecList := nodeExecToTaskExec[execution.DefaultConfig.NodeID]
	if taskExecList != nil {
		return adminPrinter.Print(config.GetConfig().MustOutputFormat(), taskExecutionColumns,
			NodeTaskExecutionToProtoMessages(taskExecList.TaskExecutions)...)
	}
	return nil
}

func getNodeExecDetailsWithTasks(ctx context.Context, project, domain, name string, cmdCtx cmdCore.CommandContext) (
	[]*admin.NodeExecution, map[string]*admin.TaskExecutionList, error) {
	// Fetching Node execution details
	nExecDetails, err := cmdCtx.AdminFetcherExt().FetchNodeExecutionDetails(ctx, name, project, domain)
	if err != nil {
		return nil, nil, err
	}
	logger.Infof(ctx, "Retrieved %v node executions", len(nExecDetails.NodeExecutions))

	// Mapping node execution id to task list
	nodeExecToTaskExec := map[string]*admin.TaskExecutionList{}
	for _, nodeExec := range nExecDetails.NodeExecutions {
		nodeExecToTaskExec[nodeExec.Id.NodeId], err = cmdCtx.AdminFetcherExt().FetchTaskExecutionsOnNode(ctx,
			nodeExec.Id.NodeId, name, project, domain)
		if err != nil {
			return nil, nil, err
		}
	}
	return nExecDetails.NodeExecutions, nodeExecToTaskExec, nil
}

func createNodeTaskExecTreeView(rootView gotree.Tree, taskExecs *admin.TaskExecutionList) gotree.Tree {
	if taskExecs == nil || len(taskExecs.TaskExecutions) == 0 {
		return gotree.New("")
	}
	if rootView == nil {
		rootView = gotree.New("")
	}
	// TODO: Replace this by filter to sort in the admin
	sort.Slice(taskExecs.TaskExecutions[:], func(i, j int) bool {
		return taskExecs.TaskExecutions[i].Id.RetryAttempt < taskExecs.TaskExecutions[j].Id.RetryAttempt
	})
	for _, taskExec := range taskExecs.TaskExecutions {
		attemptView := rootView.Add(taskAttemptPrefix + strconv.Itoa(int(taskExec.Id.RetryAttempt)))
		attemptView.Add(taskExecPrefix + taskExec.Closure.Phase.String() +
			hyphenPrefix + taskExec.Closure.StartedAt.AsTime().String() +
			hyphenPrefix + taskExec.Closure.StartedAt.AsTime().
			Add(taskExec.Closure.Duration.AsDuration()).String())
		attemptView.Add(taskTypePrefix + taskExec.Closure.TaskType)
		attemptView.Add(taskReasonPrefix + taskExec.Closure.Reason)
		if taskExec.Closure.Metadata != nil {
			metadata := attemptView.Add(taskMetadataPrefix)
			metadata.Add(taskGeneratedNamePrefix + taskExec.Closure.Metadata.GeneratedName)
			metadata.Add(taskPluginIDPrefix + taskExec.Closure.Metadata.PluginIdentifier)
			extResourcesView := metadata.Add(taskExtResourcesPrefix)
			for _, extResource := range taskExec.Closure.Metadata.ExternalResources {
				extResourcesView.Add(taskExtResourcePrefix + extResource.ExternalId)
			}
			resourcePoolInfoView := metadata.Add(taskResourcePrefix)
			for _, rsPool := range taskExec.Closure.Metadata.ResourcePoolInfo {
				resourcePoolInfoView.Add(taskExtResourcePrefix + rsPool.Namespace)
				resourcePoolInfoView.Add(taskExtResourceTokenPrefix + rsPool.AllocationToken)
			}
		}

		sort.Slice(taskExec.Closure.Logs[:], func(i, j int) bool {
			return taskExec.Closure.Logs[i].Name < taskExec.Closure.Logs[j].Name
		})

		logsView := attemptView.Add(taskLogsPrefix)
		for _, logData := range taskExec.Closure.Logs {
			logsView.Add(taskLogsNamePrefix + logData.Name)
			logsView.Add(taskLogURIPrefix + logData.Uri)
		}
	}

	return rootView
}

func createNodeDetailsTreeView(nodeExecutions []*admin.NodeExecution, nodeExecToTaskExec map[string]*admin.TaskExecutionList) gotree.Tree {
	nodeDetailsTreeView := gotree.New("")
	if nodeExecutions == nil || nodeExecToTaskExec == nil {
		return nodeDetailsTreeView
	}
	// TODO : Move to sorting using filters.
	sort.Slice(nodeExecutions[:], func(i, j int) bool {
		// TODO : Remove this after fixing the StartedAt and Duration field not being populated for start-node and end-node in Admin
		if nodeExecutions[i].Closure.StartedAt == nil || nodeExecutions[j].Closure.StartedAt == nil {
			return true
		}
		return nodeExecutions[i].Closure.StartedAt.Nanos < nodeExecutions[j].Closure.StartedAt.Nanos
	})

	for _, nodeExec := range nodeExecutions {
		nExecView := nodeDetailsTreeView.Add(nodeExec.Id.NodeId + hyphenPrefix + nodeExec.Closure.Phase.String() +
			hyphenPrefix + nodeExec.Closure.StartedAt.AsTime().String() +
			hyphenPrefix + nodeExec.Closure.StartedAt.AsTime().
			Add(nodeExec.Closure.Duration.AsDuration()).String())
		taskExecs := nodeExecToTaskExec[nodeExec.Id.NodeId]
		createNodeTaskExecTreeView(nExecView, taskExecs)
	}
	return nodeDetailsTreeView
}
