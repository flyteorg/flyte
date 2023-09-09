package cmd

import (
	"github.com/flyteorg/flytestdlib/version"
	"github.com/spf13/cobra"
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
