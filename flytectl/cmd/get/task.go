package get

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/pkg/printer"
	"github.com/lyft/flytestdlib/logger"

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
	"Version" : "$.id.version",
	"Name" : "$.id.name",
	"Type" : "$.closure.compiledTask.template.type",
	"Discoverable" : "$.closure.compiledTask.template.metadata.discoverable",
	"DiscoveryVersion" : "$.closure.compiledTask.template.metadata.discovery_version",
}

var transformTask = func(jsonbody [] byte)(interface{},error){
	results := PrintableTask{}
	if err := json.Unmarshal(jsonbody, &results); err != nil {
		return results,err
	}
	return results,nil
}

func getTaskFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	if config.GetConfig().Project == "" {
		return fmt.Errorf("Please set project name to get domain")
	}
	if config.GetConfig().Domain == "" {
		return fmt.Errorf("Please set project name to get workflow")
	}
	taskPrinter := printer.Printer{
	}

	if len(args) == 1 {
		task, err := cmdCtx.AdminClient().ListTasks(ctx, &admin.ResourceListRequest{
			Id: &admin.NamedEntityIdentifier{
				Project: config.GetConfig().Project,
				Domain:  config.GetConfig().Domain,
				Name:    args[0],
			},
			Limit: 10,
		})
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved Task", task.Tasks)

		taskPrinter.Print(config.GetConfig().Output, task.Tasks,taskStructure,transformTask)
		return nil
	}

	tasks, err := cmdCtx.AdminClient().ListTaskIds(ctx, &admin.NamedEntityIdentifierListRequest{
		Project: config.GetConfig().Project,
		Domain:  config.GetConfig().Domain,
		Limit:   3,
	})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v Task", len(tasks.Entities))
	taskPrinter.Print(config.GetConfig().Output, tasks.Entities,entityStructure,transformTaskEntity)
	return nil
}
