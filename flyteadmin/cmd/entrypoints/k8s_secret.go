package entrypoints

import (
	"github.com/spf13/cobra"

	"github.com/flyteorg/flyte/flyteadmin/auth"
)

var SecretsCmd = &cobra.Command{
	Use:     "secret",
	Aliases: []string{"secrets"},
}

func init() {
	SecretsCmd.AddCommand(auth.GetCreateSecretsCommand())
	SecretsCmd.AddCommand(auth.GetInitSecretsCommand())
	RootCmd.AddCommand(SecretsCmd)
}
