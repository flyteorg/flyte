package sandbox

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/flyteorg/flytectl/pkg/configutil"
	"github.com/flyteorg/flytectl/pkg/util"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"

	"github.com/docker/docker/api/types"
	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/docker/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var containers []types.Container

func TestTearDownFunc(t *testing.T) {
	container1 := types.Container{
		ID: "FlyteSandboxClusterName",
		Names: []string{
			docker.FlyteSandboxClusterName,
		},
	}
	containers = append(containers, container1)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return(containers, nil)
		mockDocker.OnContainerRemove(ctx, mock.Anything, types.ContainerRemoveOptions{Force: true}).Return(nil)

		err := tearDownSandbox(ctx, mockDocker)
		assert.Nil(t, err)
	})
	t.Run("Error", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return(containers, nil)
		mockDocker.OnContainerRemove(ctx, mock.Anything, types.ContainerRemoveOptions{Force: true}).Return(fmt.Errorf("err"))
		err := tearDownSandbox(ctx, mockDocker)
		assert.NotNil(t, err)
	})

}

func TestTearDownClusterFunc(t *testing.T) {
	_ = util.SetupFlyteDir()
	_ = util.WriteIntoFile([]byte("data"), configutil.FlytectlConfig)
	mockOutStream := new(io.Writer)
	ctx := context.Background()
	cmdCtx := cmdCore.NewCommandContext(nil, *mockOutStream)
	mockDocker := &mocks.Docker{}
	mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return(containers, nil)
	mockDocker.OnContainerRemove(ctx, mock.Anything, types.ContainerRemoveOptions{Force: true}).Return(nil)
	docker.Client = mockDocker
	err := teardownSandboxCluster(ctx, []string{}, cmdCtx)
	assert.Nil(t, err)
}
