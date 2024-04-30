package sandbox

import (
	"context"
	"fmt"

	"github.com/enescakir/emoji"
	"github.com/flyteorg/flytectl/pkg/docker"
)

func PrintStatus(ctx context.Context, cli docker.Docker) error {
	c, err := docker.GetSandbox(ctx, cli)
	if err != nil {
		return err
	}
	if c == nil {
		fmt.Printf("%v no Sandbox found \n", emoji.StopSign)
		return nil
	}
	fmt.Printf("Flyte local sandbox container image [%s] with status [%s] is in state [%s]", c.Image, c.Status, c.State)
	return nil
}
