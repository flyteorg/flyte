package demo

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	sandboxCmdConfig "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/sandbox"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/flyteorg/flyte/flytectl/pkg/configutil"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/docker/mocks"
	"github.com/flyteorg/flyte/flytectl/pkg/k8s"
	k8sMocks "github.com/flyteorg/flyte/flytectl/pkg/k8s/mocks"
	"github.com/flyteorg/flyte/flytectl/pkg/sandbox"
	"github.com/flyteorg/flyte/flytectl/pkg/util"
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

	t.Run("SuccessKeepVolume", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(containers, nil)
		mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(nil)
		mockK8sContextMgr := &k8sMocks.ContextOps{}
		k8s.ContextMgr = mockK8sContextMgr
		mockK8sContextMgr.OnRemoveContextMatch(mock.Anything).Return(nil)
		err := sandbox.Teardown(ctx, mockDocker, sandboxCmdConfig.DefaultTeardownFlags)
		assert.Nil(t, err)
	})
	t.Run("SuccessRemoveVolume", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(containers, nil)
		mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(nil)
		mockDocker.OnVolumeRemove(ctx, docker.FlyteSandboxVolumeName, true).Return(nil)
		mockK8sContextMgr := &k8sMocks.ContextOps{}
		k8s.ContextMgr = mockK8sContextMgr
		mockK8sContextMgr.OnRemoveContextMatch(mock.Anything).Return(nil)
		err := sandbox.Teardown(
			ctx,
			mockDocker,
			&sandboxCmdConfig.TeardownFlags{Volume: true},
		)
		assert.Nil(t, err)
	})
	t.Run("ErrorOnContainerRemove", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(containers, nil)
		mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(fmt.Errorf("err"))
		err := sandbox.Teardown(ctx, mockDocker, sandboxCmdConfig.DefaultTeardownFlags)
		assert.NotNil(t, err)
	})

	t.Run("ErrorOnContainerList", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(nil, fmt.Errorf("err"))
		err := sandbox.Teardown(ctx, mockDocker, sandboxCmdConfig.DefaultTeardownFlags)
		assert.NotNil(t, err)
	})

}

func TestTearDownClusterFunc(t *testing.T) {
	_ = util.SetupFlyteDir()
	_ = util.WriteIntoFile([]byte("data"), configutil.FlytectlConfig)
	s := testutils.Setup()
	defer s.TearDown()

	ctx := s.Ctx
	mockDocker := &mocks.Docker{}
	mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(containers, nil)
	mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(nil)
	docker.Client = mockDocker
	err := teardownDemoCluster(ctx, []string{}, s.CmdCtx)
	assert.Nil(t, err)
}
