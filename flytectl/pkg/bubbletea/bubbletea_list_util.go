package bubbletea

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/project"
	cmdcore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/pkg/pkce"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin"

	"github.com/spf13/cobra"
)

type Command struct {
	Cmd   *cobra.Command
	Name  string
	Short string
}

var (
	cmdCtx        cmdcore.CommandContext
	rootCmd       *cobra.Command
	runnableCmd   *cobra.Command
	curArgs       []string
	unhandleFlags []string
	existingFlags []string
	commandMap          = map[string]Command{}
	isCommand           = true
	listErrMsg    error = nil
)

// Generate list.Model for domain names
func makeDomainListModel(m *listModel) error {
	ctx := context.Background()
	domains, err := cmdCtx.AdminFetcherExt().GetDomains(ctx)
	if err != nil {
		return err
	}

	items := []list.Item{}
	for _, domain := range domains.Domains {
		items = append(items, item(domain.Id))
	}
	m.list = makeList(items, "Please choose one of the domains")

	return nil
}

// Generate list.Model for project names from the configured endpoint
func makeProjectListModel(m *listModel) error {
	ctx := context.Background()
	projects, err := cmdCtx.AdminFetcherExt().ListProjects(ctx, project.DefaultConfig.Filter)
	if err != nil {
		return err
	}

	items := []list.Item{}
	for _, p := range projects.Projects {
		items = append(items, item(p.Id))
	}
	m.list = makeList(items, "Please choose one of the projects")

	return nil
}

// Generate list.Model of options for different flags
func makeFlagListModel(m *listModel) error {
	i := 0
	// If flag already specified by user, skip.
	for i < len(unhandleFlags) {
		if unhandleFlags[i][0:2] == "--" {
			if !runnableCmd.Flags().Lookup(strings.TrimPrefix(unhandleFlags[i], "--")).Changed {
				break
			}
		} else if unhandleFlags[i][0] == '-' {
			if unhandleFlags[i] == "-h" {
				curArgs = append(curArgs, "-h")
				m.quitting = true
				return nil
			}
			if !runnableCmd.Flags().ShorthandLookup(strings.TrimPrefix(unhandleFlags[i], "-")).Changed {
				break
			}
		}
		i++
	}
	if i == len(unhandleFlags) {
		m.quitting = true
		return nil
	}

	flag := unhandleFlags[i]
	unhandleFlags = unhandleFlags[i+1:]
	switch flag {
	case "-p":
		curArgs = append(curArgs, flag)
		err := makeProjectListModel(m)
		if err != nil {
			return err
		}
	case "-d":
		curArgs = append(curArgs, flag)
		makeDomainListModel(m)
	}

	return nil
}

// Generate a []list.Item of cmd's subcommands
func makeSubCmdItems(cmd *cobra.Command) []list.Item {
	items := []list.Item{}

	for _, subcmd := range cmd.Commands() {
		subCmdName := strings.Fields(subcmd.Use)[0]
		commandMap[subCmdName] = Command{
			Cmd:   subcmd,
			Name:  subCmdName,
			Short: subcmd.Short,
		}
		items = append(items, item(subCmdName))
	}

	return items
}

// Generate list.Model of subcommands from a given command
func makeCmdListModel(m *listModel, c *cobra.Command) error {
	items := makeSubCmdItems(c)
	// If no subcommands, but also not captured by commandFlagMap,
	// means this command is not supported in bubbletea list yet.
	if len(items) == 0 {
		listErrMsg = fmt.Errorf("this command is not supported in interactive mode yet")
		return listErrMsg
	}
	l := makeList(items, "")
	m.list = l
	return nil
}

// Generate list.Model after user chose one of the item
func makeListModel(m *listModel, item string) error {
	// Still in the stage of handling subcommands
	if isCommand {
		var ok bool
		// Check if we reach a runnable command
		if unhandleFlags, ok = commandFlagMap[sliceToString(curArgs)]; !ok {
			err := makeCmdListModel(m, commandMap[item].Cmd)
			if err != nil {
				return err
			}
			return nil
		}
		isCommand = false
		runnableCmd = commandMap[item].Cmd
		// Check if user input flags are valid
		if err := runnableCmd.ParseFlags(existingFlags); err != nil {
			return err
		}
	}

	// Handled all flags, quit.
	if len(unhandleFlags) == 0 {
		m.quitting = true
		return nil
	}

	// Still have flags to handle
	err := makeFlagListModel(m)
	if err != nil {
		return err
	}

	return nil
}

func initCmdCtx() error {
	ctx := context.Background()
	adminCfg := admin.GetConfig(ctx)
	clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).
		WithTokenCache(pkce.NewTokenCacheKeyringProvider(
			pkce.KeyRingServiceName,
			fmt.Sprintf("%s:%s", adminCfg.Endpoint.String(), pkce.KeyRingServiceUser),
		)).Build(ctx)
	if err != nil {
		return err
	}
	cmdCtx = cmdcore.NewCommandContext(clientSet, nil)
	return nil
}

func checkRunBubbleTea() (*cobra.Command, bool, error) {
	cmd, flags, err := rootCmd.Find(os.Args[1:])
	if err != nil {
		return cmd, false, err
	}
	existingFlags = flags

	tempCmd := cmd
	for tempCmd.HasParent() {
		curArgs = append([]string{strings.Fields(tempCmd.Use)[0]}, curArgs...)
		tempCmd = tempCmd.Parent()
	}

	for _, flag := range flags {
		if flag == "--interactive" || flag == "-i" {
			return cmd, true, nil
		}
	}

	return cmd, false, nil
}
