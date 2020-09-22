package get

import (
	"context"
	"github.com/lyft/flytectl/cmd/config"
	"encoding/json"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/pkg/printer"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/logger"
)

type PrintableProject struct {
	Id          string `header:"Id"`
	Name        string `header:"Name"`
	Description string `header:"Description"`
}

var tableStructure = map[string]string{
	"Id" : "$.id",
	"Name" : "$.name",
	"Description" : "$.description",
}


func getProjectsFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	adminPrinter := printer.Printer{}

	transformProject := func(jsonbody [] byte)(interface{},error){
		results := PrintableProject{}
		if err := json.Unmarshal(jsonbody, &results); err != nil {
			return results,err
		}
		return results,nil
	}
	if len(args) == 1 {
		projects, err := cmdCtx.AdminClient().ListProjects(ctx, &admin.ProjectListRequest{})
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved %v projects", len(projects.Projects))
		for _, v := range projects.Projects {
			if v.Name == args[0] {
				adminPrinter.Print(config.GetConfig().Output, projects.Projects,tableStructure,transformProject)
			}
		}
		return nil
	}
	projects, err := cmdCtx.AdminClient().ListProjects(ctx, &admin.ProjectListRequest{})
	if err != nil {
		return err
	}
	logger.Debugf(ctx, "Retrieved %v projects", len(projects.Projects))
	adminPrinter.Print(config.GetConfig().Output, projects.Projects,tableStructure,transformProject)
	return nil
}
