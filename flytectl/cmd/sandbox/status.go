package sandbox

import (
	"context"

	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/sandbox"
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
