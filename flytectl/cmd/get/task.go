package get

import (
	"context"
	"encoding/json"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytectl/adminutils"
	"github.com/lyft/flytectl/printer"

	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

type PrintableTask struct {
	Version          string `header:"Version"`
	Name             string `header:"Name"`
	Type             string `header:"Type"`
	Discoverable     bool   `header:"Discoverable"`
	DiscoveryVersion string `header:"DiscoveryVersion"`
}

var taskStructure = map[string]string{
	"Version":          "$.id.version",
	"Name":             "$.id.name",
	"Type":             "$.closure.compiledTask.template.type",
	"Discoverable":     "$.closure.compiledTask.template.metadata.discoverable",
	"DiscoveryVersion": "$.closure.compiledTask.template.metadata.discovery_version",
}

var transformTask = func(jsonbody []byte) (interface{}, error) {
	results := PrintableTask{}
	if err := json.Unmarshal(jsonbody, &results); err != nil {
		return results, err
	}
	return results, nil
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
			Limit: 1,
		})
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved Task", task.Tasks)

		return taskPrinter.Print(config.GetConfig().MustOutputFormat(), task.Tasks, taskStructure, transformTask)
	}
	tasks, err := adminutils.GetAllNamedEntities(ctx, cmdCtx.AdminClient().ListTaskIds, adminutils.ListRequest{Project: config.GetConfig().Project, Domain: config.GetConfig().Domain})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v Task", len(tasks))
	return taskPrinter.Print(config.GetConfig().MustOutputFormat(), tasks, entityStructure, transformTaskEntity)
}
