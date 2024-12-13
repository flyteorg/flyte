package sandbox

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	sandboxCmdConfig "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/sandbox"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/docker/mocks"
	"github.com/flyteorg/flyte/flytectl/pkg/k8s"
	k8sMocks "github.com/flyteorg/flyte/flytectl/pkg/k8s/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTearDownFunc(t *testing.T) {
	var containers []types.Container
	container1 := types.Container{
		ID: "FlyteSandboxClusterName",
		Names: []string{
			docker.FlyteSandboxClusterName,
		},
	}
	containers = append(containers, container1)
	ctx := context.Background()

	mockDocker := &mocks.Docker{}
	mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(containers, nil)
	mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(fmt.Errorf("err"))
	err := Teardown(ctx, mockDocker, sandboxCmdConfig.DefaultTeardownFlags)
	assert.NotNil(t, err)

	mockDocker = &mocks.Docker{}
	mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(nil, fmt.Errorf("err"))
	err = Teardown(ctx, mockDocker, sandboxCmdConfig.DefaultTeardownFlags)
	assert.NotNil(t, err)

	mockDocker = &mocks.Docker{}
	mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(containers, nil)
	mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(nil)
	mockK8sContextMgr := &k8sMocks.ContextOps{}
	mockK8sContextMgr.OnRemoveContext(mock.Anything).Return(nil)
	k8s.ContextMgr = mockK8sContextMgr
	err = Teardown(ctx, mockDocker, sandboxCmdConfig.DefaultTeardownFlags)
	assert.Nil(t, err)

}
