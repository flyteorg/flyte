package get

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/adminutils"
	"github.com/flyteorg/flytectl/pkg/printer"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	workflowShort = "Gets workflow resources"
	workflowLong  = `
Retrieves all the workflows within project and domain.(workflow,workflows can be used interchangeably in these commands)
::

 bin/flytectl get workflow -p flytesnacks -d development

Retrieves workflow by name within project and domain.

::

 bin/flytectl get workflow -p flytesnacks -d development  core.basic.lp.go_greet

Retrieves workflow by filters. 
::

 Not yet implemented

Retrieves all the workflow within project and domain in yaml format.

::

 bin/flytectl get workflow -p flytesnacks -d development -o yaml

Retrieves all the workflow within project and domain in json format.

::

 bin/flytectl get workflow -p flytesnacks -d development -o json

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
	if len(args) > 0 {
		workflows, err := cmdCtx.AdminClient().ListWorkflows(ctx, &admin.ResourceListRequest{
			Id: &admin.NamedEntityIdentifier{
				Project: config.GetConfig().Project,
				Domain:  config.GetConfig().Domain,
				Name:    args[0],
			},
			// TODO Sorting and limits should be parameters
			SortBy: &admin.Sort{
				Key:       "created_at",
				Direction: admin.Sort_DESCENDING,
			},
			Limit: 100,
		})
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved %v workflows", len(workflows.Workflows))

		return adminPrinter.Print(config.GetConfig().MustOutputFormat(), workflowColumns, WorkflowToProtoMessages(workflows.Workflows)...)
	}

	workflows, err := adminutils.GetAllNamedEntities(ctx, cmdCtx.AdminClient().ListWorkflowIds, adminutils.ListRequest{Project: config.GetConfig().Project, Domain: config.GetConfig().Domain})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v workflows", len(workflows))
	return adminPrinter.Print(config.GetConfig().MustOutputFormat(), entityColumns, adminutils.NamedEntityToProtoMessage(workflows)...)
}
