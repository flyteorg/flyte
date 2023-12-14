package entrypoints

import (
	"github.com/spf13/cobra"

	"github.com/flyteorg/flyte/flyteadmin/auth"
)

var secretsCmd = &cobra.Command{
	Use:     "secret",
	Aliases: []string{"secrets"},
}

func init() {
	secretsCmd.AddCommand(auth.GetCreateSecretsCommand())
	secretsCmd.AddCommand(auth.GetInitSecretsCommand())
	RootCmd.AddCommand(secretsCmd)
}
