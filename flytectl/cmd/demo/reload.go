package demo

import (
	"context"
	"fmt"

	sandboxCmdConfig "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/sandbox"
)

const (
	internalBootstrapAgent = "flyte-sandbox-bootstrap"
)
const (
	reloadShort = "Power cycle the Flyte executable pod, effectively picking up an updated config."
	reloadLong  = `
If you've changed the ~/.flyte/state/flyte.yaml file, run this command to restart the Flyte binary pod, effectively
picking up the new settings:

Usage
::

 flytectl demo reload

`
)

func isLegacySandbox(ctx context.Context, cli docker.Docker, containerID string) (bool, error) {
	var result bool

	// Check if sandbox is compatible with new bootstrap mechanism
	exec, err := docker.ExecCommend(
		ctx,
		cli,
		containerID,
		[]string{"sh", "-c", fmt.Sprintf("which %s > /dev/null", internalBootstrapAgent)},
	)
	if err != nil {
		return result, err
	}
	if err = docker.InspectExecResp(ctx, cli, exec.ID); err != nil {
		return result, err
	}
	res, err := cli.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return result, err
	}

	result = res.ExitCode != 0
	return result, nil
}

func reloadDemoCluster(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}
	c, err := docker.GetSandbox(ctx, cli)
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("reload failed - could not find an active sandbox")
	}

	// Working with a legacy sandbox - fallback to legacy reload mechanism
	useLegacyMethod, err := isLegacySandbox(ctx, cli, c.ID)
	if err != nil {
		return err
	}
	if useLegacyMethod {
		return sandbox.LegacyReloadDemoCluster(ctx, sandboxCmdConfig.DefaultConfig)
	}

	// At this point we know that we are on a modern sandbox, and we can use the
	// internal bootstrap agent to reload the cluster
	exec, err := docker.ExecCommend(ctx, cli, c.ID, []string{internalBootstrapAgent})
	if err != nil {
		return err
	}
	if err = docker.InspectExecResp(ctx, cli, exec.ID); err != nil {
		return err
	}

	return nil
}
