package cmd

import (
	"github.com/lyft/flytestdlib/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Aliases: []string{"version", "ver"},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("Version: %s\nBuildSHA: %s\nBuildTS: %s\n", version.Version, version.Build, version.BuildTime.String())
	},
}

func init() {
	root.AddCommand(versionCmd)
}
