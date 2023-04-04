package demo

import (
	"context"

	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/sandbox"

	sandboxCmdConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
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
