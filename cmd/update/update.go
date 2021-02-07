package update

import (
	cmdcore "github.com/lyft/flytectl/cmd/core"

	"github.com/spf13/cobra"
)

// CreateUpdateCommand will return update command
func CreateUpdateCommand() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update various resources.",
	}

	updateResourcesFuncs := map[string]cmdcore.CommandEntry{
		"project":    {CmdFunc: updateProjectsFunc, Aliases: []string{"projects"}, ProjectDomainNotRequired: true, PFlagProvider: projectConfig},
	}

	cmdcore.AddCommands(updateCmd, updateResourcesFuncs)
	return updateCmd
}
