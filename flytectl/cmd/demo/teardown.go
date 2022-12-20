package demo

import (
	"context"

	"github.com/flyteorg/flytectl/pkg/sandbox"

	"github.com/flyteorg/flytectl/pkg/docker"

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
	return sandbox.Teardown(ctx, cli)
}
