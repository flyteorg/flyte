package demo

import (
	"context"
	"fmt"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/k8s"
	"github.com/flyteorg/flytestdlib/logger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	labelSelector = "app=flyte"
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

// reloadDemoCluster will kill the flyte binary pod so the new one can pick up a new config file
func reloadDemoCluster(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
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
