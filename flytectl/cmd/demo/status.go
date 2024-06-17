package demo

import (
	"context"

	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/sandbox"
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
