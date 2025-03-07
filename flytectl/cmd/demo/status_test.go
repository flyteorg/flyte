package demo

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/docker/mocks"
	"github.com/stretchr/testify/assert"
)

func TestDemoStatus(t *testing.T) {
	t.Run("Demo status with zero result", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		s := testutils.Setup(t)
		mockDocker.EXPECT().ContainerList(s.Ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		docker.Client = mockDocker
		err := demoClusterStatus(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Demo status with running", func(t *testing.T) {
		s := testutils.Setup(t)
		ctx := s.Ctx
		mockDocker := &mocks.Docker{}
		mockDocker.EXPECT().ContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{
			{
				ID: docker.FlyteSandboxClusterName,
				Names: []string{
					docker.FlyteSandboxClusterName,
				},
			},
		}, nil)
		docker.Client = mockDocker
		err := demoClusterStatus(ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
	})
}
