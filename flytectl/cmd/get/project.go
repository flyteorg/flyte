package get

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flytectl/pkg/printer"
)

const(
projectShort = "Gets project resources"
projectLong  = `
Retrieves all the projects.(project,projects can be used interchangeably in these commands)
::

 bin/flytectl get project

Retrieves project by name

::

 bin/flytectl get project flytesnacks

Retrieves project by filters
::

 Not yet implemented

Retrieves all the projects in yaml format

::

 bin/flytectl get project -o yaml

Retrieves all the projects in json format

::

 bin/flytectl get project -o json

Usage
`
)

var projectColumns = []printer.Column{
	{"ID", "$.id"},
	{"Name", "$.name"},
	{"Description", "$.description"},
}

func ProjectToProtoMessages(l []*admin.Project) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func getProjectsFunc(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	adminPrinter := printer.Printer{}
	projects, err := cmdCtx.AdminClient().ListProjects(ctx, &admin.ProjectListRequest{})
	if err != nil {
		return err
	}
	if len(args) == 1 {
		name := args[0]
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "Retrieved %v projects", len(projects.Projects))
		for _, v := range projects.Projects {
			if v.Name == name {
				err := adminPrinter.Print(config.GetConfig().MustOutputFormat(), projectColumns, v)
				if err != nil {
					return err
				}
				return nil
			}
		}
		return nil
	}
	logger.Debugf(ctx, "Retrieved %v projects", len(projects.Projects))
	return adminPrinter.Print(config.GetConfig().MustOutputFormat(), projectColumns, ProjectToProtoMessages(projects.Projects)...)
}
