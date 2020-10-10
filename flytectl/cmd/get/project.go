package get

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/printer"
)

var projectColumns = []printer.Column{
	{"ID", "$.id"},
	{"Name", "$.name"},
	{"Description", "$.description"},
}

func getProjectsFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	adminPrinter := printer.Printer{}

	if len(args) == 1 {
		name := args[0]
		projects, err := cmdCtx.AdminClient().ListProjects(ctx, &admin.ProjectListRequest{})
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved %v projects", len(projects.Projects))
		for _, v := range projects.Projects {
			if v.Name == name {
				err := adminPrinter.Print(config.GetConfig().MustOutputFormat(), v, projectColumns)
				if err != nil {
					return err
				}
				return nil
			}
		}
		return nil
	}
	projects, err := cmdCtx.AdminClient().ListProjects(ctx, &admin.ProjectListRequest{})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v projects", len(projects.Projects))
	return adminPrinter.Print(config.GetConfig().MustOutputFormat(), projects.Projects, projectColumns)
}
