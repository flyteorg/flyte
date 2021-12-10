package sandbox

import (
	sandboxConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdcore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	sandboxShort = `Used for sandbox interactions like start/teardown/status/exec.`
	sandboxLong  = `
The Flyte Sandbox is a fully standalone minimal environment for running Flyte. It provides a simplified way of running Flyte sandbox as a single Docker container locally.
	
Create sandbox cluster:
::

 flytectl sandbox start 
	
	
Remove sandbox cluster:
::

 flytectl sandbox teardown 	
	

Check status of sandbox container:
::

 flytectl sandbox status 	
	
Execute command inside sandbox container:
::

 flytectl sandbox exec -- pwd 	
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
			Long:  startLong, PFlagProvider: sandboxConfig.DefaultConfig},
		"teardown": {CmdFunc: teardownSandboxCluster, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: teardownShort,
			Long:  teardownLong},
		"status": {CmdFunc: sandboxClusterStatus, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: statusShort,
			Long:  statusLong},
		"exec": {CmdFunc: sandboxClusterExec, Aliases: []string{}, ProjectDomainNotRequired: true,
			Short: execShort,
			Long:  execLong},
	}

	cmdcore.AddCommands(sandbox, sandboxResourcesFuncs)

	return sandbox
}
