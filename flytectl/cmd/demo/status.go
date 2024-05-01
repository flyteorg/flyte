package demo

import (
	"context"

	"github.com/flyteorg/flytectl/pkg/sandbox"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
)

const (
	statusShort = "Gets the status of the demo environment."
	statusLong  = `
Retrieves the status of the demo environment. Currently, Flyte demo runs as a local Docker container.

Usage
::

 flytectl demo status 

`
)

func demoClusterStatus(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}

	return sandbox.PrintStatus(ctx, cli)
}
