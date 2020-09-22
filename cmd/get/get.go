package get

import (
	"github.com/lyft/flytectl/cmd/core"

	"github.com/spf13/cobra"
)

func CreateGetCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve various resource.",
	}

	getResourcesFuncs := map[string]cmdcore.CommandFunc{
		"projects":  getProjectsFunc,
		"tasks":     getTaskFunc,
		"workflows": getWorkflowFunc,
	}

	cmdcore.AddCommands(getCmd, getResourcesFuncs)

	return getCmd
}
