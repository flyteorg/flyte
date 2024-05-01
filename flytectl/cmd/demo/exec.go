package demo

import (
	"context"
	"fmt"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
)

const (
	execShort = "Executes non-interactive command inside the demo container"
	execLong  = `
Run non-interactive commands inside the demo container and immediately return the output.
By default, "flytectl exec" is present in the /root directory inside the demo container.

::

 flytectl demo exec -- ls -al 

Usage`
)

func demoClusterExec(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return execute(ctx, cli, args)
	}
	return fmt.Errorf("missing argument. Please check usage examples by running flytectl demo exec --help")
}

func execute(ctx context.Context, cli docker.Docker, args []string) error {
	c, err := docker.GetSandbox(ctx, cli)
	if err != nil {
		return err
	}
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
