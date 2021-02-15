package get

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytectl/pkg/adminutils"
	"github.com/lyft/flytectl/pkg/printer"

	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	taskShort = "Gets task resources"
	taskLong  = `
Retrieves all the task within project and domain.(task,tasks can be used interchangeably in these commands)
::

 bin/flytectl get task -p flytesnacks -d development

Retrieves task by name within project and domain.

::

 bin/flytectl task -p flytesnacks -d development square

Retrieves project by filters.
::

 Not yet implemented

Retrieves all the tasks within project and domain in yaml format.

::

 bin/flytectl get task -p flytesnacks -d development -o yaml

Retrieves all the tasks within project and domain in json format.

::

 bin/flytectl get task -p flytesnacks -d development -o json

Usage
`
)

var taskColumns = []printer.Column{
	{Header: "Version", JSONPath: "$.id.version"},
	{Header: "Name", JSONPath: "$.id.name"},
	{Header: "Type", JSONPath: "$.closure.compiledTask.template.type"},
	{Header: "Discoverable", JSONPath: "$.closure.compiledTask.template.metadata.discoverable"},
	{Header: "Discovery Version", JSONPath: "$.closure.compiledTask.template.metadata.discoveryVersion"},
	{Header: "Created At", JSONPath: "$.closure.createdAt"},
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
				Key:       "created_at",
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
