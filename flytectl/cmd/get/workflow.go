package get

import (
	"context"
	"encoding/json"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytectl/adminutils"
	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/printer"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

var workflowStructure = map[string]string{
	"Version": "$.id.version",
	"Name":    "$.id.name",
}

type PrintableWorkflow struct {
	Name    string `header:"Name"`
	Version string `header:"Version"`
}

var transformWorkflow = func(jsonbody []byte) (interface{}, error) {
	results := PrintableWorkflow{}
	if err := json.Unmarshal(jsonbody, &results); err != nil {
		return results, err
	}
	return results, nil
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
			Limit: 1,
		})
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved %v workflows", len(workflows.Workflows))

		return adminPrinter.Print(config.GetConfig().MustOutputFormat(), workflows.Workflows, workflowStructure, transformWorkflow)
	}

	workflows, err := adminutils.GetAllNamedEntities(ctx, cmdCtx.AdminClient().ListWorkflowIds, adminutils.ListRequest{Project: config.GetConfig().Project, Domain: config.GetConfig().Domain})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v workflows", len(workflows))
	return adminPrinter.Print(config.GetConfig().MustOutputFormat(), workflows, entityStructure, transformTaskEntity)
}
