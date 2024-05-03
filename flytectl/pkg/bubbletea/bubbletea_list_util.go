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
	rootCmd *cobra.Command
	newArgs []string
	flags   []string
)
var (
	DOMAIN_NAME   = [3]string{"development", "staging", "production"}
	isCommand     = true
	nameToCommand = map[string]Command{}
)

// Generate a []list.Item of cmd's subcommands
func generateSubCmdItems(cmd *cobra.Command) []list.Item {
	items := []list.Item{}

	for _, subcmd := range cmd.Commands() {
		subCmdName := strings.Fields(subcmd.Use)[0]
		nameToCommand[subCmdName] = Command{
			Cmd:   subcmd,
			Name:  subCmdName,
			Short: subcmd.Short,
		}
		items = append(items, item(subCmdName))
	}

	return items
}

// Generate list.Model for domain names
func genDomainListModel(m listModel) (listModel, error) {
	items := []list.Item{}
	for _, domain := range DOMAIN_NAME {
		items = append(items, item(domain))
	}

	m.list = genList(items, "Please choose one of the domains")
	return m, nil
}

// Get the "get" "project" cobra.Command item
func extractGetProjectCmd() *cobra.Command {
	var getCmd *cobra.Command
	var getProjectCmd *cobra.Command

	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "get" {
			getCmd = cmd
			break
		}
	}
	for _, cmd := range getCmd.Commands() {
		if cmd.Use == "project" {
			getProjectCmd = cmd
			break
		}
	}
	return getProjectCmd
}

// Get all the project names from the configured endpoint
func getProjects(getProjectCmd *cobra.Command) ([]string, error) {
	ctx := context.Background()
	rootCmd.PersistentPreRunE(rootCmd, []string{})
	adminCfg := admin.GetConfig(ctx)

	clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).
		WithTokenCache(pkce.TokenCacheKeyringProvider{
			ServiceUser: fmt.Sprintf("%s:%s", adminCfg.Endpoint.String(), pkce.KeyRingServiceUser),
			ServiceName: pkce.KeyRingServiceName,
		}).Build(ctx)
	if err != nil {
		return nil, err
	}
	cmdCtx := cmdcore.NewCommandContext(clientSet, getProjectCmd.OutOrStdout())

	projects, err := cmdCtx.AdminFetcherExt().ListProjects(ctx, project.DefaultConfig.Filter)
	if err != nil {
		return nil, err
	}

	projectNames := []string{}
	for _, p := range projects.Projects {
		projectNames = append(projectNames, p.Id)
	}

	return projectNames, nil
}

// Generate list.Model for project names from the configured endpoint
func genProjectListModel(m listModel) (listModel, error) {
	getProjectCmd := extractGetProjectCmd()
	projects, err := getProjects(getProjectCmd)
	if err != nil {
		return m, err
	}

	items := []list.Item{}
	for _, project := range projects {
		items = append(items, item(project))
	}

	m.list = genList(items, "Please choose one of the projects")

	return m, nil
}

// Generate list.Model of options for different flags
func genFlagListModel(m listModel, f string) (listModel, error) {
	var err error

	switch f {
	case "-p":
		m, err = genProjectListModel(m)
	case "-d":
		m, err = genDomainListModel(m)
	}

	return m, err
}

// Generate list.Model of subcommands from a given command
func genCmdListModel(m listModel, c string) listModel {
	if len(nameToCommand[c].Cmd.Commands()) == 0 {
		return m
	}

	items := generateSubCmdItems(nameToCommand[c].Cmd)
	l := genList(items, "")
	m.list = l

	return m
}

// Generate list.Model after user chose one of the item
func genListModel(m listModel, item string) (listModel, error) {
	newArgs = append(newArgs, item)

	if isCommand {
		m = genCmdListModel(m, item)
		var ok bool
		if flags, ok = commandFlagMap[sliceToString(newArgs)]; ok { // If found in commandFlagMap means last command
			isCommand = false
		} else {
			return m, nil
		}
	}
	// TODO check if some flags are already input as arguments by user
	if len(flags) > 0 {
		nextFlag := flags[0]
		flags = flags[1:]
		newArgs = append(newArgs, nextFlag)
		var err error
		m, err = genFlagListModel(m, nextFlag)
		if err != nil {
			return m, err
		}
	} else {
		m.quitting = true
		return m, nil
	}
	return m, nil
}

func ifRunBubbleTea(_rootCmd cobra.Command) (*cobra.Command, bool, error) {
	cmd, flags, err := _rootCmd.Find(os.Args[1:])
	if err != nil {
		return cmd, false, err
	}

	tempCmd := cmd
	for tempCmd.HasParent() {
		newArgs = append([]string{tempCmd.Use}, newArgs...)
		tempCmd = tempCmd.Parent()
	}

	for _, flag := range flags {
		if flag == "-i" || flag == "--interactive" {
			return cmd, true, nil
		}
	}

	return cmd, false, nil
}

// func isValidCommand(curArg string, cmd *cobra.Command) (*cobra.Command, bool) {
// 	for _, subCmd := range cmd.Commands() {
// 		if subCmd.Use == curArg {
// 			return subCmd, true
// 		}
// 	}
// 	return nil, false
// }

// func findSubCmdItems(cmd *cobra.Command, inputArgs []string) ([]list.Item, error) {
// 	if len(inputArgs) == 0 {
// 		return generateSubCmdItems(cmd), nil
// 	}

// 	curArg := inputArgs[0]
// 	subCmd, isValid := isValidCommand(curArg, cmd)
// 	if !isValid {
// 		return nil, fmt.Errorf("not a valid argument: %v", curArg)
// 	}

// 	return findSubCmdItems(subCmd, inputArgs[1:])
// }
