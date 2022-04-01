package cmdcore

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/pkg/pkce"
	"github.com/flyteorg/flyteidl/clients/go/admin"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type PFlagProvider interface {
	GetPFlagSet(prefix string) *pflag.FlagSet
}

type CommandEntry struct {
	ProjectDomainNotRequired bool
	CmdFunc                  CommandFunc
	Aliases                  []string
	Short                    string
	Long                     string
	PFlagProvider            PFlagProvider
}

func AddCommands(rootCmd *cobra.Command, cmdFuncs map[string]CommandEntry) {
	for resource, cmdEntry := range cmdFuncs {
		cmd := &cobra.Command{
			Use:          resource,
			Short:        cmdEntry.Short,
			Long:         cmdEntry.Long,
			Aliases:      cmdEntry.Aliases,
			RunE:         generateCommandFunc(cmdEntry),
			SilenceUsage: true,
		}

		if cmdEntry.PFlagProvider != nil {
			cmd.Flags().AddFlagSet(cmdEntry.PFlagProvider.GetPFlagSet(""))
		}

		rootCmd.AddCommand(cmd)
	}
}

func generateCommandFunc(cmdEntry CommandEntry) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		if !cmdEntry.ProjectDomainNotRequired {
			if config.GetConfig().Project == "" {
				return fmt.Errorf("project and domain are required parameters")
			}
			if config.GetConfig().Domain == "" {
				return fmt.Errorf("project and domain are required parameters")
			}
		}
		if _, err := config.GetConfig().OutputFormat(); err != nil {
			return err
		}

		adminCfg := admin.GetConfig(ctx)
		clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).
			WithTokenCache(pkce.TokenCacheKeyringProvider{
				ServiceUser: fmt.Sprintf("%s:%s", adminCfg.Endpoint.String(), pkce.KeyRingServiceUser),
				ServiceName: pkce.KeyRingServiceName,
			}).Build(ctx)
		if err != nil {
			return err
		}
		return cmdEntry.CmdFunc(ctx, args, NewCommandContext(clientSet, cmd.OutOrStdout()))
	}
}
