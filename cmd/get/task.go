package get

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytectl/adminutils"
	"github.com/lyft/flytectl/printer"

	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

var taskColumns = []printer.Column{
	{"Version", "$.id.version"},
	{"Name", "$.id.name"},
	{"Type", "$.closure.compiledTask.template.type"},
	{"Discoverable", "$.closure.compiledTask.template.metadata.discoverable"},
	{"Discovery Version", "$.closure.compiledTask.template.metadata.discoveryVersion"},
	{"Created At", "$.closure.createdAt"},
}

func TaskToProtoMessages(l []*admin.Task) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func getTaskFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {

	taskPrinter := printer.Printer{}

	if len(args) == 1 {
		task, err := cmdCtx.AdminClient().ListTasks(ctx, &admin.ResourceListRequest{
			Id: &admin.NamedEntityIdentifier{
				Project: config.GetConfig().Project,
				Domain:  config.GetConfig().Domain,
				Name:    args[0],
			},
			// TODO Sorting and limits should be parameters
			SortBy: &admin.Sort{
				Key: "created_at",
				Direction: admin.Sort_DESCENDING,
			},
			Limit: 100,
		})
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved Task", task.Tasks)

		return taskPrinter.Print(config.GetConfig().MustOutputFormat(), taskColumns, TaskToProtoMessages(task.Tasks)...)
	}
	tasks, err := adminutils.GetAllNamedEntities(ctx, cmdCtx.AdminClient().ListTaskIds, adminutils.ListRequest{Project: config.GetConfig().Project, Domain: config.GetConfig().Domain})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v Task", len(tasks))
	return taskPrinter.Print(config.GetConfig().MustOutputFormat(), entityColumns, adminutils.NamedEntityToProtoMessage(tasks)...)
}
