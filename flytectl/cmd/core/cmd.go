package core

import (
	"context"
	"fmt"

	"github.com/lyft/flyteidl/clients/go/admin"
	"github.com/spf13/cobra"
)

func AddCommands(rootCmd *cobra.Command, cmdFuncs map[string]CommandFunc) {
	for resource, getFunc := range cmdFuncs {
		cmd := &cobra.Command{
			Use:   resource,
			Short: fmt.Sprintf("Retrieves %v resources.", resource),
			RunE:  generateCommandFunc(getFunc),
		}

		rootCmd.AddCommand(cmd)
	}
}

func generateCommandFunc(cmdFunc CommandFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		adminClient, err := admin.InitializeAdminClientFromConfig(ctx)
		if err != nil {
			return err
		}

		return cmdFunc(ctx, args, CommandContext{
			out:         cmd.OutOrStdout(),
			adminClient: adminClient,
		})
	}
}
