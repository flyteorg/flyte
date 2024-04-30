package demo

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/flytestdlib/logger"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/k8s"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	internalBootstrapAgent = "flyte-sandbox-bootstrap"
	labelSelector          = "app.kubernetes.io/name=flyte-binary"
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
		return legacyReloadDemoCluster(ctx)
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

// legacyReloadDemoCluster will kill the flyte binary pod so the new one can pick up a new config file
func legacyReloadDemoCluster(ctx context.Context) error {
	k8sClient, err := k8s.GetK8sClient(docker.Kubeconfig, K8sEndpoint)
	if err != nil {
		fmt.Println("Could not get K8s client")
		return err
	}
	pi := k8sClient.CoreV1().Pods(flyteNs)
	podList, err := pi.List(ctx, v1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		fmt.Println("could not list pods")
		return err
	}
	if len(podList.Items) != 1 {
		return fmt.Errorf("should only have one pod running, %d found, %v", len(podList.Items), podList.Items)
	}
	logger.Debugf(ctx, "Found %d pods\n", len(podList.Items))
	var grace = int64(0)
	err = pi.Delete(ctx, podList.Items[0].Name, v1.DeleteOptions{
		GracePeriodSeconds: &grace,
	})
	if err != nil {
		fmt.Printf("Could not delete Flyte pod, old configuration may still be in effect. Err: %s\n", err)
		return err
	}

	return nil
}
