package get

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/execution"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/printer"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/golang/protobuf/proto"
)

const (
	executionShort = "Gets execution resources"
	executionLong  = `
Retrieves all the executions within project and domain.(execution,executions can be used interchangeably in these commands)
::

 bin/flytectl get execution -p flytesnacks -d development

Retrieves execution by name within project and domain.

::

 bin/flytectl get execution -p flytesnacks -d development oeh94k9r2r

Retrieves all the executions with filters.
::
 
  bin/flytectl get execution -p flytesnacks -d development --filter.fieldSelector="execution.phase in (FAILED;SUCCEEDED),execution.duration<200" 

 
Retrieves all the execution with limit and sorting.
::
  
   bin/flytectl get execution -p flytesnacks -d development --filter.sortBy=created_at --filter.limit=1 --filter.asc
   

Retrieves all the execution within project and domain in yaml format

::

 bin/flytectl get execution -p flytesnacks -d development -o yaml

Retrieves all the execution within project and domain in json format.

::

 bin/flytectl get execution -p flytesnacks -d development -o json


Get more details for the execution using --details flag which shows node executions along with task executions on them. Default view is tree view and TABLE format is not supported on this view

::

 bin/flytectl get execution -p flytesnacks -d development oeh94k9r2r --details

Using yaml view for the details. In this view only node details are available. For task details pass --nodeID flag

::

 bin/flytectl get execution -p flytesnacks -d development oeh94k9r2r --details -o yaml

Using --nodeID flag to get task executions on a specific node. Use the nodeID attribute from node details view

::

 bin/flytectl get execution -p flytesnacks -d development oeh94k9r2r --nodID n0

Task execution view is also available in yaml/json format. Below example shows yaml. This also contains inputs/outputs data for each node

::

 bin/flytectl get execution -p flytesnacks -d development oeh94k9r2r --nodID n0 -o yaml

Usage
`
)

var hundredChars = 100

var executionColumns = []printer.Column{
	{Header: "Name", JSONPath: "$.id.name"},
	{Header: "Launch Plan Name", JSONPath: "$.spec.launchPlan.name"},
	{Header: "Type", JSONPath: "$.spec.launchPlan.resourceType"},
	{Header: "Phase", JSONPath: "$.closure.phase"},
	{Header: "Scheduled Time", JSONPath: "$.spec.metadata.scheduledAt"},
	{Header: "Started", JSONPath: "$.closure.startedAt"},
	{Header: "Elapsed Time", JSONPath: "$.closure.duration"},
	{Header: "Abort data (Trunc)", JSONPath: "$.closure.abortMetadata[\"cause\"]", TruncateTo: &hundredChars},
	{Header: "Error data (Trunc)", JSONPath: "$.closure.error[\"message\"]", TruncateTo: &hundredChars},
}

func ExecutionToProtoMessages(l []*admin.Execution) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func getExecutionFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	adminPrinter := printer.Printer{}
	var executions []*admin.Execution
	if len(args) > 0 {
		name := args[0]
		exec, err := cmdCtx.AdminFetcherExt().FetchExecution(ctx, name, config.GetConfig().Project, config.GetConfig().Domain)
		if err != nil {
			return err
		}
		executions = append(executions, exec)
		logger.Infof(ctx, "Retrieved %v executions", len(executions))

		if execution.DefaultConfig.Details || len(execution.DefaultConfig.NodeID) > 0 {
			// Fetching Node execution details
			nExecDetailsForView, err := getExecutionDetails(ctx, config.GetConfig().Project, config.GetConfig().Domain, name, execution.DefaultConfig.NodeID, cmdCtx)
			if err != nil {
				return err
			}
			// o/p format of table is not supported on the details. TODO: Add tree format in printer
			if config.GetConfig().MustOutputFormat() == printer.OutputFormatTABLE {
				fmt.Println("TABLE format is not supported on detailed view and defaults to tree view. Choose either json/yaml")
				nodeExecTree := createNodeDetailsTreeView(nil, nExecDetailsForView)
				fmt.Println(nodeExecTree.Print())
				return nil
			}
			return adminPrinter.PrintInterface(config.GetConfig().MustOutputFormat(), nodeExecutionColumns, nExecDetailsForView)
		}
		return adminPrinter.Print(config.GetConfig().MustOutputFormat(), executionColumns,
			ExecutionToProtoMessages(executions)...)
	}
	executionList, err := cmdCtx.AdminFetcherExt().ListExecution(ctx, config.GetConfig().Project, config.GetConfig().Domain, execution.DefaultConfig.Filter)
	if err != nil {
		return err
	}
	logger.Infof(ctx, "Retrieved %v executions", len(executionList.Executions))
	return adminPrinter.Print(config.GetConfig().MustOutputFormat(), executionColumns,
		ExecutionToProtoMessages(executionList.Executions)...)
}
