package sandbox

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/docker/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSandboxStatus(t *testing.T) {
	t.Run("Sandbox status with zero result", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		s := testutils.Setup()
		mockDocker.OnContainerList(s.Ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		docker.Client = mockDocker
		err := sandboxClusterStatus(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Sandbox status with running sandbox", func(t *testing.T) {
		s := testutils.Setup()
		ctx := s.Ctx
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{
			{
				ID: docker.FlyteSandboxClusterName,
				Names: []string{
					docker.FlyteSandboxClusterName,
				},
			},
		}, nil)
		docker.Client = mockDocker
		err := sandboxClusterStatus(ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
	})
}
