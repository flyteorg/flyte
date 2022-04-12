package demo

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/pkg/configutil"

	"github.com/flyteorg/flytectl/pkg/docker"

	"github.com/docker/docker/api/types"
	"github.com/enescakir/emoji"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/k8s"
)

const (
	teardownShort = "Cleans up the demo environment"
	teardownLong  = `
Removes the demo cluster and all the Flyte config created by 'demo start':
::

 flytectl demo teardown 
	

Usage
`
)

func teardownDemoCluster(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}

	return tearDownDemo(ctx, cli)
}

func tearDownDemo(ctx context.Context, cli docker.Docker) error {
	c := docker.GetSandbox(ctx, cli)
	if c != nil {
		if err := cli.ContainerRemove(context.Background(), c.ID, types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			return err
		}
	}
	if err := configutil.ConfigCleanup(); err != nil {
		fmt.Printf("Config cleanup failed. Which Failed due to %v \n ", err)
	}
	if err := removeDemoKubeContext(); err != nil {
		fmt.Printf("Kubecontext cleanup failed. Which Failed due to %v \n ", err)
	}
	fmt.Printf("%v %v Demo cluster is removed successfully. \n", emoji.Broom, emoji.Broom)
	return nil
}

func removeDemoKubeContext() error {
	k8sCtxMgr := k8s.NewK8sContextManager()
	return k8sCtxMgr.RemoveContext(demoContextName)
}
