package demo

import (
	sandboxCmdConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdcore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/spf13/cobra"
)

const (
	flyteNs     = "flyte"
	K8sEndpoint = "https://127.0.0.1:6443"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	demoShort = `Helps with demo interactions like start, teardown, status, and exec.`
	demoLong  = `
Flyte Demo is a fully standalone minimal environment for running Flyte.
It provides a simplified way of running Flyte demo as a single Docker container locally.
	
To create a demo cluster, run:
::

 flytectl demo start 

To remove a demo cluster, run:
::

 flytectl demo teardown

To check the status of the demo container, run:
::

 flytectl demo status

To execute commands inside the demo container, use exec:
::

 flytectl demo exec -- pwd 	
`
)

// CreateDemoCommand will return demo command
func CreateDemoCommand() *cobra.Command {
	demo := &cobra.Command{
		Use:   "demo",
		Short: demoShort,
		Long:  demoLong,
	}

	demoResourcesFuncs := map[string]cmdcore.CommandEntry{
		"start": {CmdFunc: startDemoCluster, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: startShort,
			Long:  startLong, PFlagProvider: sandboxCmdConfig.DefaultConfig, DisableFlyteClient: true},
		"reload": {CmdFunc: reloadDemoCluster, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: reloadShort,
			Long:  reloadLong, PFlagProvider: sandboxCmdConfig.DefaultConfig, DisableFlyteClient: true},
		"teardown": {CmdFunc: teardownDemoCluster, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short:              teardownShort,
			Long:               teardownLong,
			PFlagProvider:      sandboxCmdConfig.DefaultTeardownFlags,
			DisableFlyteClient: true},
		"status": {CmdFunc: demoClusterStatus, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: statusShort,
			Long:  statusLong},
		"exec": {CmdFunc: demoClusterExec, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: execShort,
			Long:  execLong, DisableFlyteClient: true},
	}

	cmdcore.AddCommands(demo, demoResourcesFuncs)

	return demo
}
