package get

import (
	"context"
	"encoding/json"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/printer"
)

type PrintableProject struct {
	ID          string `header:"Id"`
	Name        string `header:"Name"`
	Description string `header:"Description"`
}

var tableStructure = map[string]string{
	"ID":          "$.id",
	"Name":        "$.name",
	"Description": "$.description",
}

func transformProject(jsonbody []byte) (interface{}, error) {
	results := PrintableProject{}
	if err := json.Unmarshal(jsonbody, &results); err != nil {
		return results, err
	}
	return results, nil
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
				err := adminPrinter.Print(config.GetConfig().MustOutputFormat(), v, tableStructure, transformProject)
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
	return adminPrinter.Print(config.GetConfig().MustOutputFormat(), projects.Projects, tableStructure, transformProject)
}
