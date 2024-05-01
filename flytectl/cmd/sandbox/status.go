package sandbox

import (
	"context"

	"github.com/flyteorg/flytectl/pkg/sandbox"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
)

const (
	statusShort = "Gets the status of the sandbox environment."
	statusLong  = `
Retrieves the status of the sandbox environment. Currently, Flyte sandbox runs as a local Docker container.

Usage
::

 flytectl sandbox status 

`
)

func sandboxClusterStatus(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}

	return sandbox.PrintStatus(ctx, cli)
}
