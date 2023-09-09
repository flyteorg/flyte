package entrypoints

import (
	"github.com/flyteorg/flyteadmin/auth"
	"github.com/spf13/cobra"
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
