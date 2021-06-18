package get

import (
	"context"

	workflowconfig "github.com/flyteorg/flytectl/cmd/config/subcommand/workflow"
	"github.com/flyteorg/flytectl/pkg/ext"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/printer"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	workflowShort = "Gets workflow resources"
	workflowLong  = `
Retrieves all the workflows within project and domain.(workflow,workflows can be used interchangeably in these commands)
::

 flytectl get workflow -p flytesnacks -d development

Retrieves workflow by name within project and domain.

::

 flytectl get workflow -p flytesnacks -d development  core.basic.lp.go_greet

Retrieves latest version of workflow by name within project and domain.

::

 flytectl get workflow -p flytesnacks -d development  core.basic.lp.go_greet --latest

Retrieves particular version of workflow by name within project and domain.

::

 flytectl get workflow -p flytesnacks -d development  core.basic.lp.go_greet --version v2

Retrieves all the workflows with filters.
::
 
  bin/flytectl get workflow -p flytesnacks -d development  --filter.field-selector="workflow.name=k8s_spark.dataframe_passing.my_smart_schema"
 
Retrieve specific workflow with filters.
::
 
  bin/flytectl get workflow -p flytesnacks -d development k8s_spark.dataframe_passing.my_smart_schema --filter.field-selector="workflow.version=v1"
  
Retrieves all the workflows with limit and sorting.
::
  
  bin/flytectl get -p flytesnacks -d development workflow  --filter.sort-by=created_at --filter.limit=1 --filter.asc

Retrieves all the workflow within project and domain in yaml format.

::

 flytectl get workflow -p flytesnacks -d development -o yaml

Retrieves all the workflow within project and domain in json format.

::

 flytectl get workflow -p flytesnacks -d development -o json

Visualize the graph for a workflow within project and domain in dot format.

::

 flytectl get workflow -p flytesnacks -d development  core.flyte_basics.basic_workflow.my_wf --latest -o dot

Visualize the graph for a workflow within project and domain in a dot content render.

::

 flytectl get workflow -p flytesnacks -d development  core.flyte_basics.basic_workflow.my_wf --latest -o doturl

Usage
`
)

var workflowColumns = []printer.Column{
	{Header: "Version", JSONPath: "$.id.version"},
	{Header: "Name", JSONPath: "$.id.name"},
	{Header: "Created At", JSONPath: "$.closure.createdAt"},
}

func WorkflowToProtoMessages(l []*admin.Workflow) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func getWorkflowFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	adminPrinter := printer.Printer{}
	var workflows []*admin.Workflow
	var err error
	if len(args) > 0 {
		name := args[0]
		if workflows, err = FetchWorkflowForName(ctx, cmdCtx.AdminFetcherExt(), name, config.GetConfig().Project, config.GetConfig().Domain); err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved %v workflow", len(workflows))
		return adminPrinter.Print(config.GetConfig().MustOutputFormat(), workflowColumns, WorkflowToProtoMessages(workflows)...)
	}

	workflows, err = cmdCtx.AdminFetcherExt().FetchAllVerOfWorkflow(ctx, "", config.GetConfig().Project, config.GetConfig().Domain, workflowconfig.DefaultConfig.Filter)
	if err != nil {
		return err
	}

	logger.Debugf(ctx, "Retrieved %v workflows", len(workflows))
	return adminPrinter.Print(config.GetConfig().MustOutputFormat(), workflowColumns, WorkflowToProtoMessages(workflows)...)
}

// FetchWorkflowForName fetches the workflow give it name.
func FetchWorkflowForName(ctx context.Context, fetcher ext.AdminFetcherExtInterface, name, project,
	domain string) ([]*admin.Workflow, error) {
	var workflows []*admin.Workflow
	var workflow *admin.Workflow
	var err error
	if workflowconfig.DefaultConfig.Latest {
		if workflow, err = fetcher.FetchWorkflowLatestVersion(ctx, name, project, domain, workflowconfig.DefaultConfig.Filter); err != nil {
			return nil, err
		}
		workflows = append(workflows, workflow)
	} else if workflowconfig.DefaultConfig.Version != "" {
		if workflow, err = fetcher.FetchWorkflowVersion(ctx, name, workflowconfig.DefaultConfig.Version, project, domain); err != nil {
			return nil, err
		}
		workflows = append(workflows, workflow)
	} else {
		workflows, err = fetcher.FetchAllVerOfWorkflow(ctx, name, project, domain, workflowconfig.DefaultConfig.Filter)
		if err != nil {
			return nil, err
		}
	}
	return workflows, nil
}
