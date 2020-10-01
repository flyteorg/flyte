package get

import (
	cmdcore "github.com/lyft/flytectl/cmd/core"

	"github.com/spf13/cobra"
)

func CreateGetCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve various resource.",
	}

	getResourcesFuncs := map[string]cmdcore.CommandEntry{
		"projects":  {CmdFunc: getProjectsFunc, ProjectDomainNotRequired: true},
		"tasks":     {CmdFunc: getTaskFunc},
		"workflows": {CmdFunc: getWorkflowFunc},
	}

	cmdcore.AddCommands(getCmd, getResourcesFuncs)

	return getCmd
}
