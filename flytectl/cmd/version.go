package cmd

import (
	"github.com/flyteorg/flytestdlib/version"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Displays version information for the client and server.",
		Run: func(cmd *cobra.Command, args []string) {
			version.LogBuildInformation("flytectl")
			// TODO: Log Admin version
		},
	}
)
