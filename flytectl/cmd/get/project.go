package get

import (
	"context"

	"github.com/flyteorg/flytectl/cmd/config/subcommand/project"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/printer"
)

const (
	projectShort = "Gets project resources"
	projectLong  = `
Retrieve all the projects:
::

 flytectl get project

.. note::
	  The terms project/projects are interchangeable in these commands.

Retrieve project by name:

::

 flytectl get project flytesnacks

Retrieve all the projects with filters:
::

  flytectl get project --filter.fieldSelector="project.name=flytesnacks"

Retrieve all the projects with limit and sorting:
::

  flytectl get project --filter.sortBy=created_at --filter.limit=1 --filter.asc

Retrieve projects present in other pages by specifying the limit and page number:
::

  flytectl get project --filter.limit=10 --filter.page=2

Retrieve all the projects in yaml format:

::

 flytectl get project -o yaml

Retrieve all the projects in json format:

::

 flytectl get project -o json

Usage
`
)

var projectColumns = []printer.Column{
	{Header: "ID", JSONPath: "$.id"},
	{Header: "Name", JSONPath: "$.name"},
	{Header: "Description", JSONPath: "$.description"},
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

	projects, err := cmdCtx.AdminFetcherExt().ListProjects(ctx, project.DefaultConfig.Filter)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		id := args[0]
		logger.Debugf(ctx, "Retrieved %v projects", len(projects.Projects))
		for _, v := range projects.Projects {
			if v.Id == id {
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
