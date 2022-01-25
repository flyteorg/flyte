package docker

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	sandboxConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"

	f "github.com/flyteorg/flytectl/pkg/filesystemutils"

	"github.com/docker/docker/api/types/container"
	"github.com/flyteorg/flytectl/pkg/docker/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/docker/docker/api/types"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	u "github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/stretchr/testify/assert"
)

var (
	cmdCtx     cmdCore.CommandContext
	containers []types.Container
	imageName  = "cr.flyte.org/flyteorg/flyte-sandbox"
)

func setupSandbox() {
	mockAdminClient := u.MockClient
	cmdCtx = cmdCore.NewCommandContext(mockAdminClient, u.MockOutStream)
	err := os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte"), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	container1 := types.Container{
		ID: "FlyteSandboxClusterName",
		Names: []string{
			FlyteSandboxClusterName,
		},
	}
	containers = append(containers, container1)
}

func TestGetSandbox(t *testing.T) {
	setupSandbox()
	t.Run("Successfully get sandbox container", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		context := context.Background()

		mockDocker.OnContainerList(context, types.ContainerListOptions{All: true}).Return(containers, nil)
		c := GetSandbox(context, mockDocker)
		assert.Equal(t, c.Names[0], FlyteSandboxClusterName)
	})

	t.Run("Successfully get sandbox container with zero result", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		context := context.Background()

		mockDocker.OnContainerList(context, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		c := GetSandbox(context, mockDocker)
		assert.Nil(t, c)
	})

	t.Run("Error in get sandbox container", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		context := context.Background()

		mockDocker.OnContainerList(context, types.ContainerListOptions{All: true}).Return(containers, nil)
		mockDocker.OnContainerRemove(context, mock.Anything, types.ContainerRemoveOptions{Force: true}).Return(nil)
		err := RemoveSandbox(context, mockDocker, strings.NewReader("y"))
		assert.Nil(t, err)
	})

}

func TestRemoveSandboxWithNoReply(t *testing.T) {
	setupSandbox()
	t.Run("Successfully remove sandbox container", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		context := context.Background()

		// Verify the attributes
		mockDocker.OnContainerList(context, types.ContainerListOptions{All: true}).Return(containers, nil)
		mockDocker.OnContainerRemove(context, mock.Anything, types.ContainerRemoveOptions{Force: true}).Return(nil)
		err := RemoveSandbox(context, mockDocker, strings.NewReader("n"))
		assert.NotNil(t, err)
	})

	t.Run("Successfully remove sandbox container with zero sandbox containers are running", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		context := context.Background()

		// Verify the attributes
		mockDocker.OnContainerList(context, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnContainerRemove(context, mock.Anything, types.ContainerRemoveOptions{Force: true}).Return(nil)
		err := RemoveSandbox(context, mockDocker, strings.NewReader("n"))
		assert.Nil(t, err)
	})

}

func TestPullDockerImage(t *testing.T) {
	t.Run("Successfully pull image Always", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		context := context.Background()
		// Verify the attributes
		mockDocker.OnImagePullMatch(context, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		err := PullDockerImage(context, mockDocker, "nginx:latest", sandboxConfig.ImagePullPolicyAlways)
		assert.Nil(t, err)
	})

	t.Run("Error in pull image", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		context := context.Background()
		// Verify the attributes
		mockDocker.OnImagePullMatch(context, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, fmt.Errorf("error"))
		err := PullDockerImage(context, mockDocker, "nginx:latest", sandboxConfig.ImagePullPolicyAlways)
		assert.NotNil(t, err)
	})

	t.Run("Successfully pull image IfNotPresent", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		context := context.Background()
		// Verify the attributes
		mockDocker.OnImagePullMatch(context, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		mockDocker.OnImageListMatch(context, types.ImageListOptions{}).Return([]types.ImageSummary{}, nil)
		err := PullDockerImage(context, mockDocker, "nginx:latest", sandboxConfig.ImagePullPolicyIfNotPresent)
		assert.Nil(t, err)
	})

	t.Run("Successfully pull image Never", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		context := context.Background()
		err := PullDockerImage(context, mockDocker, "nginx:latest", sandboxConfig.ImagePullPolicyNever)
		assert.Nil(t, err)
	})
}

func TestStartContainer(t *testing.T) {
	p1, p2, _ := GetSandboxPorts()

	t.Run("Successfully create a container", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		context := context.Background()

		// Verify the attributes
		mockDocker.OnContainerCreate(context, &container.Config{
			Env:          Environment,
			Image:        imageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       Volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(context, "Hello", types.ContainerStartOptions{}).Return(nil)
		id, err := StartContainer(context, mockDocker, Volumes, p1, p2, "nginx", imageName)
		assert.Nil(t, err)
		assert.Greater(t, len(id), 0)
		assert.Equal(t, id, "Hello")
	})

	t.Run("Error in creating container", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		context := context.Background()

		// Verify the attributes
		mockDocker.OnContainerCreate(context, &container.Config{
			Env:          Environment,
			Image:        imageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       Volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "",
		}, fmt.Errorf("error"))
		mockDocker.OnContainerStart(context, "Hello", types.ContainerStartOptions{}).Return(nil)
		id, err := StartContainer(context, mockDocker, Volumes, p1, p2, "nginx", imageName)
		assert.NotNil(t, err)
		assert.Equal(t, len(id), 0)
		assert.Equal(t, id, "")
	})

	t.Run("Error in start of a container", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		context := context.Background()

		// Verify the attributes
		mockDocker.OnContainerCreate(context, &container.Config{
			Env:          Environment,
			Image:        imageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       Volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(context, "Hello", types.ContainerStartOptions{}).Return(fmt.Errorf("error"))
		id, err := StartContainer(context, mockDocker, Volumes, p1, p2, "nginx", imageName)
		assert.NotNil(t, err)
		assert.Equal(t, len(id), 0)
		assert.Equal(t, id, "")
	})
}

func TestReadLogs(t *testing.T) {
	setupSandbox()

	t.Run("Successfully read logs", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		context := context.Background()
		mockDocker.OnContainerLogsMatch(context, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		_, err := ReadLogs(context, mockDocker, "test")
		assert.Nil(t, err)
	})

	t.Run("Error in reading logs", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		context := context.Background()
		mockDocker.OnContainerLogsMatch(context, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, fmt.Errorf("error"))
		_, err := ReadLogs(context, mockDocker, "test")
		assert.NotNil(t, err)
	})
}

func TestWaitForSandbox(t *testing.T) {
	setupSandbox()
	t.Run("Successfully read logs ", func(t *testing.T) {
		reader := bufio.NewScanner(strings.NewReader("hello \n Flyte"))

		check := WaitForSandbox(reader, "Flyte")
		assert.Equal(t, true, check)
	})

	t.Run("Error in reading logs ", func(t *testing.T) {
		reader := bufio.NewScanner(strings.NewReader(""))
		check := WaitForSandbox(reader, "Flyte")
		assert.Equal(t, false, check)
	})
}

func TestDockerClient(t *testing.T) {
	t.Run("Successfully get docker mock client", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		Client = mockDocker
		cli, err := GetDockerClient()
		assert.Nil(t, err)
		assert.NotNil(t, cli)
	})
	t.Run("Successfully get docker client", func(t *testing.T) {
		Client = nil
		cli, err := GetDockerClient()
		assert.Nil(t, err)
		assert.NotNil(t, cli)
	})
}

func TestDockerExec(t *testing.T) {
	t.Run("Successfully exec command in container", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		Client = mockDocker
		c := ExecConfig
		c.Cmd = []string{"ls"}
		mockDocker.OnContainerExecCreateMatch(ctx, mock.Anything, c).Return(types.IDResponse{}, nil)
		_, err := ExecCommend(ctx, mockDocker, "test", []string{"ls"})
		assert.Nil(t, err)
	})
	t.Run("Failed exec command in container", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		Client = mockDocker
		c := ExecConfig
		c.Cmd = []string{"ls"}
		mockDocker.OnContainerExecCreateMatch(ctx, mock.Anything, c).Return(types.IDResponse{}, fmt.Errorf("test"))
		_, err := ExecCommend(ctx, mockDocker, "test", []string{"ls"})
		assert.NotNil(t, err)
	})
}

func TestInspectExecResp(t *testing.T) {
	t.Run("Failed exec command in container", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		Client = mockDocker
		c := ExecConfig
		c.Cmd = []string{"ls"}
		reader := bufio.NewReader(strings.NewReader("test"))

		mockDocker.OnContainerExecInspectMatch(ctx, mock.Anything).Return(types.ContainerExecInspect{}, nil)
		mockDocker.OnContainerExecAttachMatch(ctx, mock.Anything, types.ExecStartCheck{}).Return(types.HijackedResponse{
			Reader: reader,
		}, fmt.Errorf("err"))

		err := InspectExecResp(ctx, mockDocker, "test")
		assert.NotNil(t, err)
	})
	t.Run("Successfully exec command in container", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		Client = mockDocker
		c := ExecConfig
		c.Cmd = []string{"ls"}
		reader := bufio.NewReader(strings.NewReader("test"))

		mockDocker.OnContainerExecAttachMatch(ctx, mock.Anything, types.ExecStartCheck{}).Return(types.HijackedResponse{
			Reader: reader,
		}, nil)

		err := InspectExecResp(ctx, mockDocker, "test")
		assert.Nil(t, err)
	})

}
