package get

import (
	"context"

	"github.com/flyteorg/flyte/flytectl/cmd/config"
	workflowconfig "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/workflow"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/pkg/ext"
	"github.com/flyteorg/flyte/flytectl/pkg/printer"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
)

const (
	workflowShort = "Gets workflow resources"
	workflowLong  = `
Retrieve all the workflows within project and domain (workflow/workflows can be used interchangeably in these commands):
::

 flytectl get workflow -p flytesnacks -d development

Retrieve all versions of a workflow by name within project and domain:

::

 flytectl get workflow -p flytesnacks -d development  core.basic.lp.go_greet

Retrieve latest version of workflow by name within project and domain:

::

 flytectl get workflow -p flytesnacks -d development  core.basic.lp.go_greet --latest

Retrieve particular version of workflow by name within project and domain:

::

 flytectl get workflow -p flytesnacks -d development  core.basic.lp.go_greet --version v2

Retrieve all the workflows with filters:
::

  flytectl get workflow -p flytesnacks -d development  --filter.fieldSelector="workflow.name=k8s_spark.dataframe_passing.my_smart_schema"

Retrieve specific workflow with filters:
::

  flytectl get workflow -p flytesnacks -d development k8s_spark.dataframe_passing.my_smart_schema --filter.fieldSelector="workflow.version=v1"

Retrieve all the workflows with limit and sorting:
::

  flytectl get -p flytesnacks -d development workflow  --filter.sortBy=created_at --filter.limit=1 --filter.asc

Retrieve workflows present in other pages by specifying the limit and page number:
::

  flytectl get -p flytesnacks -d development workflow --filter.limit=10 --filter.page 2

Retrieve all the workflows within project and domain in yaml format:

::

 flytectl get workflow -p flytesnacks -d development -o yaml

Retrieve all the workflow within project and domain in json format:

::

 flytectl get workflow -p flytesnacks -d development -o json

Visualize the graph for a workflow within project and domain in dot format:

::

 flytectl get workflow -p flytesnacks -d development  core.flyte_basics.basic_workflow.my_wf --latest -o dot

Visualize the graph for a workflow within project and domain in a dot content render:

::

 flytectl get workflow -p flytesnacks -d development  core.flyte_basics.basic_workflow.my_wf --latest -o doturl

Usage
`
)

var workflowColumns = []printer.Column{
	{Header: "Version", JSONPath: "$.id.version"},
	{Header: "Name", JSONPath: "$.id.name"},
	{Header: "Inputs", JSONPath: "$.closure.compiledWorkflow.primary.template.interface.inputs.variables." + printer.DefaultFormattedDescriptionsKey + ".description"},
	{Header: "Outputs", JSONPath: "$.closure.compiledWorkflow.primary.template.interface.outputs.variables." + printer.DefaultFormattedDescriptionsKey + ".description"},
	{Header: "Created At", JSONPath: "$.closure.createdAt"},
}

var listWorkflowColumns = []printer.Column{
	{Header: "Version", JSONPath: "$.id.version"},
	{Header: "Name", JSONPath: "$.id.name"},
	{Header: "Created At", JSONPath: "$.closure.createdAt"},
}

var namedEntityColumns = []printer.Column{
	{Header: "Project", JSONPath: "$.id.project"},
	{Header: "Domain", JSONPath: "$.id.domain"},
	{Header: "Name", JSONPath: "$.id.name"},
	{Header: "Description", JSONPath: "$.metadata.description"},
	{Header: "State", JSONPath: "$.metadata.state"},
}

func WorkflowToProtoMessages(l []*admin.Workflow) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func NamedEntityToProtoMessages(l []*admin.NamedEntity) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func WorkflowToTableProtoMessages(l []*admin.Workflow) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		m := proto.Clone(m).(*admin.Workflow)
		if m.Closure != nil && m.Closure.CompiledWorkflow != nil {
			if m.Closure.CompiledWorkflow.Primary != nil {
				if m.Closure.CompiledWorkflow.Primary.Template != nil {
					if m.Closure.CompiledWorkflow.Primary.Template.Interface != nil {
						if m.Closure.CompiledWorkflow.Primary.Template.Interface.Inputs != nil && m.Closure.CompiledWorkflow.Primary.Template.Interface.Inputs.Variables != nil {
							printer.FormatVariableDescriptions(m.Closure.CompiledWorkflow.Primary.Template.Interface.Inputs.Variables)
						}
						if m.Closure.CompiledWorkflow.Primary.Template.Interface.Outputs != nil && m.Closure.CompiledWorkflow.Primary.Template.Interface.Outputs.Variables != nil {
							printer.FormatVariableDescriptions(m.Closure.CompiledWorkflow.Primary.Template.Interface.Outputs.Variables)
						}
					}
				}
			}
		}
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
		var isList bool
		if workflows, isList, err = FetchWorkflowForName(ctx, cmdCtx.AdminFetcherExt(), name, config.GetConfig().Project, config.GetConfig().Domain); err != nil {
			return err
		}
		columns := workflowColumns
		if isList {
			columns = listWorkflowColumns
		}
		logger.Debugf(ctx, "Retrieved %v workflow", len(workflows))
		if config.GetConfig().MustOutputFormat() == printer.OutputFormatTABLE {
			return adminPrinter.Print(config.GetConfig().MustOutputFormat(), columns, WorkflowToTableProtoMessages(workflows)...)
		}
		return adminPrinter.Print(config.GetConfig().MustOutputFormat(), columns, WorkflowToProtoMessages(workflows)...)
	}

	nameEntities, err := cmdCtx.AdminFetcherExt().FetchAllWorkflows(ctx, config.GetConfig().Project, config.GetConfig().Domain, workflowconfig.DefaultConfig.Filter)
	if err != nil {
		return err
	}

	logger.Debugf(ctx, "Retrieved %v workflows", len(nameEntities))
	return adminPrinter.Print(config.GetConfig().MustOutputFormat(), namedEntityColumns, NamedEntityToProtoMessages(nameEntities)...)
}

// FetchWorkflowForName fetches the workflow give it name.
func FetchWorkflowForName(ctx context.Context, fetcher ext.AdminFetcherExtInterface, name, project,
	domain string) (workflows []*admin.Workflow, isList bool, err error) {
	var workflow *admin.Workflow
	if workflowconfig.DefaultConfig.Latest {
		if workflow, err = fetcher.FetchWorkflowLatestVersion(ctx, name, project, domain); err != nil {
			return nil, false, err
		}
		workflows = append(workflows, workflow)
	} else if workflowconfig.DefaultConfig.Version != "" {
		if workflow, err = fetcher.FetchWorkflowVersion(ctx, name, workflowconfig.DefaultConfig.Version, project, domain); err != nil {
			return nil, false, err
		}
		workflows = append(workflows, workflow)
	} else {
		workflows, err = fetcher.FetchAllVerOfWorkflow(ctx, name, project, domain, workflowconfig.DefaultConfig.Filter)
		if err != nil {
			return nil, false, err
		}
		isList = true
	}
	return workflows, isList, nil
}
