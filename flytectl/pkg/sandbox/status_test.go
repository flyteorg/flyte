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
		defer s.TearDown()
		mockDocker.OnContainerList(s.Ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		err := PrintStatus(s.Ctx, mockDocker)
		assert.Nil(t, err)
	})
	t.Run("Sandbox status with running sandbox", func(t *testing.T) {
		s := testutils.Setup()
		defer s.TearDown()
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
		err := PrintStatus(ctx, mockDocker)
		assert.Nil(t, err)
	})
}
