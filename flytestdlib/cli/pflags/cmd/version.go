package cmd

import (
	"github.com/spf13/cobra"

	"github.com/flyteorg/flyte/v2/flytestdlib/version"
)

var versionCmd = &cobra.Command{
	Aliases: []string{"version", "ver"},
	Run: func(cmd *cobra.Command, args []string) {
		version.LogBuildInformation("pflags")
	},
}

func init() {
	root.AddCommand(versionCmd)
}
