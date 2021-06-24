package sandbox

import (
	"context"
	"io"
	"testing"

	"github.com/docker/docker/api/types"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/docker/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSandboxStatus(t *testing.T) {
	t.Run("Sandbox status with zero result", func(t *testing.T) {
		ctx := context.Background()
		mockOutStream := new(io.Writer)
		cmdCtx := cmdCore.NewCommandContext(nil, *mockOutStream)
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		docker.Client = mockDocker
		err := sandboxClusterStatus(ctx, []string{}, cmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Sandbox status with running sandbox", func(t *testing.T) {
		ctx := context.Background()
		mockOutStream := new(io.Writer)
		cmdCtx := cmdCore.NewCommandContext(nil, *mockOutStream)
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{
			{
				ID: docker.FlyteSandboxClusterName,
				Names: []string{
					docker.FlyteSandboxClusterName,
				},
			},
		}, nil)
		docker.Client = mockDocker
		err := sandboxClusterStatus(ctx, []string{}, cmdCtx)
		assert.Nil(t, err)
	})
}
