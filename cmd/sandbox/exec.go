package sandbox

import (
	"context"
	"fmt"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
)

const (
	execShort = "Execute non-interactive command inside the sandbox container"
	execLong  = `
Runs non-interactive command inside the Sandbox container and immediately returns the output. By default, flytectl exec is present in /root directory inside the Sandbox container.

::
 flytectl sandbox exec -- ls -al 

Usage`
)

func sandboxClusterExec(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return execute(ctx, cli, args)
	}
	return fmt.Errorf("missing argument. Please check usage examples by running flytectl sandbox exec --help")
}

func execute(ctx context.Context, cli docker.Docker, args []string) error {
	c := docker.GetSandbox(ctx, cli)
	if c != nil {
		exec, err := docker.ExecCommend(ctx, cli, c.ID, args)
		if err != nil {
			return err
		}
		if err := docker.InspectExecResp(ctx, cli, exec.ID); err != nil {
			return err
		}
	}
	return nil
}
