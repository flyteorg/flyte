package sandbox

import (
	"context"

	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/sandbox"

	sandboxCmdConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	teardownShort = "Cleans up the sandbox environment"
	teardownLong  = `
Removes the Sandbox cluster and all the Flyte config created by 'sandbox start':
::

 flytectl sandbox teardown 
	

Usage
`
)

func teardownSandboxCluster(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}
	return sandbox.Teardown(ctx, cli, sandboxCmdConfig.DefaultTeardownFlags)
}
