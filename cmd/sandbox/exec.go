package sandbox

import (
	"context"
	"fmt"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
)

const (
	execShort = "Execute any command in sandbox"
	execLong  = `
Execute command will Will run non-interactive commands and return immediately with the output.

::
 bin/flytectl sandbox exec -- ls -al 

Usage`
)

func sandboxClusterExec(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return Execute(ctx, cli, args)
	}
	return fmt.Errorf("missing argument. Please check usage examples by running flytectl sandbox exec --help")
}

func Execute(ctx context.Context, cli docker.Docker, args []string) error {
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
