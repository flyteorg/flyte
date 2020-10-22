package get

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/pkg/adminutils"
	"github.com/lyft/flytectl/pkg/printer"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

var workflowColumns = []printer.Column{
	{"Version", "$.id.version"},
	{"Name", "$.id.name"},
	{"Created At", "$.closure.createdAt"},
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
