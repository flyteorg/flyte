package demo

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/docker/mocks"
	admin2 "github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDemoClusterExec(t *testing.T) {
	mockDocker := &mocks.Docker{}
	mockOutStream := new(io.Writer)
	ctx := context.Background()
	mockClient := admin2.InitializeMockClientset()
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	reader := bufio.NewReader(strings.NewReader("test"))

	mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{
		{
			ID: docker.FlyteSandboxClusterName,
			Names: []string{
				docker.FlyteSandboxClusterName,
			},
		},
	}, nil)
	docker.ExecConfig.Cmd = []string{"ls -al"}
	mockDocker.OnContainerExecCreateMatch(ctx, mock.Anything, docker.ExecConfig).Return(types.IDResponse{}, nil)
	mockDocker.OnContainerExecInspectMatch(ctx, mock.Anything).Return(types.ContainerExecInspect{}, nil)
	mockDocker.OnContainerExecAttachMatch(ctx, mock.Anything, types.ExecStartCheck{}).Return(types.HijackedResponse{
		Reader: reader,
	}, fmt.Errorf("Test"))
	docker.Client = mockDocker
	err := demoClusterExec(ctx, []string{"ls -al"}, cmdCtx)

	assert.NotNil(t, err)
}

func TestSandboxClusterExecWithoutCmd(t *testing.T) {
	mockDocker := &mocks.Docker{}
	reader := bufio.NewReader(strings.NewReader("test"))
	s := testutils.Setup()
	defer s.TearDown()

	ctx := s.Ctx

	mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{
		{
			ID: docker.FlyteSandboxClusterName,
			Names: []string{
				docker.FlyteSandboxClusterName,
			},
		},
	}, nil)
	docker.ExecConfig.Cmd = []string{}
	mockDocker.OnContainerExecCreateMatch(ctx, mock.Anything, docker.ExecConfig).Return(types.IDResponse{}, nil)
	mockDocker.OnContainerExecInspectMatch(ctx, mock.Anything).Return(types.ContainerExecInspect{}, nil)
	mockDocker.OnContainerExecAttachMatch(ctx, mock.Anything, types.ExecStartCheck{}).Return(types.HijackedResponse{
		Reader: reader,
	}, fmt.Errorf("Test"))
	docker.Client = mockDocker
	err := demoClusterExec(ctx, []string{}, s.CmdCtx)

	assert.NotNil(t, err)
}
