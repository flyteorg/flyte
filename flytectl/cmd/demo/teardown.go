package demo

import (
	"context"

	sandboxCmdConfig "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/sandbox"
)

const (
	teardownShort = "Cleans up the demo environment"
	teardownLong  = `
Removes the demo cluster and all the Flyte config created by 'demo start':
::

 flytectl demo teardown 
	

Usage
`
)

func teardownDemoCluster(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}
	return sandbox.Teardown(ctx, cli, sandboxCmdConfig.DefaultTeardownFlags)
}
