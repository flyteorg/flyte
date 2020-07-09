package get

import (
	"context"

	"github.com/lyft/flytectl/cmd/core"

	"github.com/lyft/flytestdlib/logger"

	"github.com/landoop/tableprinter"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/spf13/cobra"
)

func CreateGetCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve various resource.",
	}

	getResourcesFuncs := map[string]core.CommandFunc{
		"projects": getProjectsFunc,
	}

	core.AddCommands(getCmd, getResourcesFuncs)

	return getCmd
}

func getProjectsFunc(ctx context.Context, args []string, cmdCtx core.CommandContext) error {
	projects, err := cmdCtx.AdminClient().ListProjects(ctx, &admin.ProjectListRequest{})
	if err != nil {
		return err
	}

	logger.Debugf(ctx, "Retrieved %v projects", len(projects.Projects))
	printer := tableprinter.New(cmdCtx.OutputPipe())
	printer.Print(toPrintableProjects(projects.Projects))
	return nil
}

func toPrintableProjects(projects []*admin.Project) []interface{} {
	type printableProject struct {
		Id          string `header:"Id"`
		Name        string `header:"Name"`
		Description string `header:"Description"`
	}

	res := make([]interface{}, 0, len(projects))
	for _, p := range projects {
		res = append(res, printableProject{
			Id:          p.Id,
			Name:        p.Name,
			Description: p.Description,
		})
	}

	return res
}
