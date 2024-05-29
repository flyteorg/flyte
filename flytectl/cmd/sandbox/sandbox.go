package sandbox

import (
	sandboxCmdConfig "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/sandbox"
	cmdcore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	sandboxShort = `Helps with sandbox interactions like start, teardown, status, and exec.`
	sandboxLong  = `
Flyte Sandbox is a fully standalone minimal environment for running Flyte.
It provides a simplified way of running Flyte sandbox as a single Docker container locally.
	
To create a sandbox cluster, run:
::

 flytectl sandbox start 

To remove a sandbox cluster, run:
::

 flytectl sandbox teardown

To check the status of the sandbox container, run:
::

 flytectl sandbox status

To execute commands inside the sandbox container, use exec:
::

 flytectl sandbox exec -- pwd 	

For just printing the docker commands for bringingup the demo container
::

 flytectl demo start --dryRun

`
)

// CreateSandboxCommand will return sandbox command
func CreateSandboxCommand() *cobra.Command {
	sandbox := &cobra.Command{
		Use:   "sandbox",
		Short: sandboxShort,
		Long:  sandboxLong,
	}

	sandboxResourcesFuncs := map[string]cmdcore.CommandEntry{
		"start": {CmdFunc: startSandboxCluster, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: startShort,
			Long:  startLong, PFlagProvider: sandboxCmdConfig.DefaultConfig, DisableFlyteClient: true},
		"teardown": {CmdFunc: teardownSandboxCluster, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: teardownShort,
			Long:  teardownLong, DisableFlyteClient: true},
		"status": {CmdFunc: sandboxClusterStatus, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: statusShort,
			Long:  statusLong},
		"exec": {CmdFunc: sandboxClusterExec, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: execShort,
			Long:  execLong, DisableFlyteClient: true},
	}

	cmdcore.AddCommands(sandbox, sandboxResourcesFuncs)

	return sandbox
}
