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

var workflowStructure = map[string]string{
	"Version" : "$.id.version",
	"Name" : "$.id.name",
}

type PrintableWorkflow struct {
	Name    string `header:"Name"`
	Version string `header:"Version"`
}

var transformWorkflow = func(jsonbody [] byte)(interface{},error){
	results := PrintableWorkflow{}
	if err := json.Unmarshal(jsonbody, &results); err != nil {
		return results,err
	}
	return results,nil
}


func getWorkflowFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	if config.GetConfig().Project == "" {
		return fmt.Errorf("Please set project name to get domain")
	}
	if config.GetConfig().Domain == "" {
		return fmt.Errorf("Please set project name to get workflow")
	}
	adminPrinter := printer.Printer{
	}
	if len(args) > 0 {
		workflows, err := cmdCtx.AdminClient().ListWorkflows(ctx, &admin.ResourceListRequest{
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
		logger.Debugf(ctx, "Retrieved %v workflows", len(workflows.Workflows))

		adminPrinter.Print(config.GetConfig().Output, workflows.Workflows,workflowStructure,transformWorkflow)
		return nil
	}
	workflows, err := cmdCtx.AdminClient().ListWorkflowIds(ctx, &admin.NamedEntityIdentifierListRequest{
		Project: config.GetConfig().Project,
		Domain:  config.GetConfig().Domain,
		Limit:   10,
	})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v workflows", len(workflows.Entities))

	adminPrinter.Print(config.GetConfig().Output, workflows.Entities,entityStructure,transformTaskEntity)
	return nil
}
