package sandbox

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"

	sandboxConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/docker/mocks"
	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStartSandboxFunc(t *testing.T) {
	p1, p2, _ := docker.GetSandboxPorts()

	t.Run("Successfully run sandbox cluster", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       docker.Volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(nil)
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		_, err := startSandbox(ctx, mockDocker, os.Stdin)
		assert.Nil(t, err)
	})
	t.Run("Successfully run sandbox cluster with source code", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Source = f.UserHomeDir()
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: sandboxConfig.DefaultConfig.Source,
			Target: docker.Source,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(nil)
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		_, err := startSandbox(ctx, mockDocker, os.Stdin)
		assert.Nil(t, err)
	})
	t.Run("Successfully run sandbox cluster with abs path of source code", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Source = "../"
		absPath, err := filepath.Abs(sandboxConfig.DefaultConfig.Source)
		assert.Nil(t, err)
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: absPath,
			Target: docker.Source,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(nil)
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		_, err = startSandbox(ctx, mockDocker, os.Stdin)
		assert.Nil(t, err)
	})
	t.Run("Successfully run sandbox cluster with specific version", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Version = "v0.15.0"
		sandboxConfig.DefaultConfig.Source = ""
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: FlyteManifest,
			Target: GeneratedManifest,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(nil)
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		_, err := startSandbox(ctx, mockDocker, os.Stdin)
		assert.Nil(t, err)
	})
	t.Run("Failed run sandbox cluster with wrong version", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Version = "v0.13.0"
		sandboxConfig.DefaultConfig.Source = ""
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: FlyteManifest,
			Target: GeneratedManifest,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(nil)
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		_, err := startSandbox(ctx, mockDocker, os.Stdin)
		assert.NotNil(t, err)
	})
	t.Run("Error in pulling image", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Source = f.UserHomeDir()
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: sandboxConfig.DefaultConfig.Source,
			Target: docker.Source,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(nil)
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, fmt.Errorf("error"))
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		_, err := startSandbox(ctx, mockDocker, os.Stdin)
		assert.NotNil(t, err)
	})
	t.Run("Error in  removing existing cluster", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Source = f.UserHomeDir()
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: sandboxConfig.DefaultConfig.Source,
			Target: docker.Source,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(nil)
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{
			{
				ID: docker.FlyteSandboxClusterName,
				Names: []string{
					docker.FlyteSandboxClusterName,
				},
			},
		}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		mockDocker.OnContainerRemove(ctx, mock.Anything, types.ContainerRemoveOptions{Force: true}).Return(fmt.Errorf("error"))
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		_, err := startSandbox(ctx, mockDocker, strings.NewReader("y"))
		assert.NotNil(t, err)
	})
	t.Run("Error in start container", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Source = ""
		sandboxConfig.DefaultConfig.Version = ""
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       docker.Volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, fmt.Errorf("error"))
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(fmt.Errorf("error"))
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		_, err := startSandbox(ctx, mockDocker, os.Stdin)
		assert.NotNil(t, err)
	})
	t.Run("Failed manifest", func(t *testing.T) {
		err := downloadFlyteManifest("v100.9.9")
		assert.NotNil(t, err)
	})
	t.Run("Error in reading logs", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Source = f.UserHomeDir()
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: sandboxConfig.DefaultConfig.Source,
			Target: docker.Source,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(nil)
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, fmt.Errorf("error"))
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		_, err := startSandbox(ctx, mockDocker, os.Stdin)
		assert.NotNil(t, err)
	})
	t.Run("Error in list container", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Source = f.UserHomeDir()
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: sandboxConfig.DefaultConfig.Source,
			Target: docker.Source,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(nil)
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, fmt.Errorf("error"))
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		_, err := startSandbox(ctx, mockDocker, os.Stdin)
		assert.Nil(t, err)
	})
	t.Run("Successfully run sandbox cluster command", func(t *testing.T) {
		mockOutStream := new(io.Writer)
		ctx := context.Background()
		cmdCtx := cmdCore.NewCommandContext(nil, *mockOutStream)
		mockDocker := &mocks.Docker{}
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       docker.Volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(nil)
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		stringReader := strings.NewReader(docker.SuccessMessage)
		reader := ioutil.NopCloser(stringReader)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(reader, nil)
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		docker.Client = mockDocker
		sandboxConfig.DefaultConfig.Source = ""
		err := startSandboxCluster(ctx, []string{}, cmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Error in running sandbox cluster command", func(t *testing.T) {
		mockOutStream := new(io.Writer)
		ctx := context.Background()
		cmdCtx := cmdCore.NewCommandContext(nil, *mockOutStream)
		mockDocker := &mocks.Docker{}
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.ImageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       docker.Volumes,
			PortBindings: p2,
			Privileged:   true,
		}, nil, nil, mock.Anything).Return(container.ContainerCreateCreatedBody{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", types.ContainerStartOptions{}).Return(fmt.Errorf("error"))
		mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return([]types.Container{}, fmt.Errorf("error"))
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(os.Stdin, nil)
		stringReader := strings.NewReader(docker.SuccessMessage)
		reader := ioutil.NopCloser(stringReader)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, types.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(reader, nil)
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		docker.Client = mockDocker
		sandboxConfig.DefaultConfig.Source = ""
		err := startSandboxCluster(ctx, []string{}, cmdCtx)
		assert.NotNil(t, err)
	})
}
