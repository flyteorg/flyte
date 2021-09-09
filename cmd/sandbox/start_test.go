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

	"github.com/flyteorg/flytectl/pkg/util/githubutil"

	"github.com/flyteorg/flytectl/pkg/k8s"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	sandboxConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/docker/mocks"
	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
	"github.com/flyteorg/flytectl/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

var content = `
apiVersion: v1
clusters:
- cluster:
    server: https://localhost:8080
    extensions:
    - name: client.authentication.k8s.io/exec
      extension:
        audience: foo
        other: bar
  name: foo-cluster
contexts:
- context:
    cluster: foo-cluster
    user: foo-user
    namespace: bar
  name: foo-context
current-context: foo-context
kind: Config
users:
- name: foo-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      args:
      - arg-1
      - arg-2
      command: foo-command
      provideClusterInfo: true
`

var fakeNode = &corev1.Node{
	Spec: corev1.NodeSpec{
		Taints: []corev1.Taint{},
	},
}

var fakePod = corev1.Pod{
	Status: corev1.PodStatus{
		Phase:      corev1.PodRunning,
		Conditions: []corev1.PodCondition{},
	},
}

func TestStartSandboxFunc(t *testing.T) {
	p1, p2, _ := docker.GetSandboxPorts()
	assert.Nil(t, util.SetupFlyteDir())
	assert.Nil(t, os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte", "k3s"), os.ModePerm))
	assert.Nil(t, ioutil.WriteFile(docker.Kubeconfig, []byte(content), os.ModePerm))

	fakePod.SetName("flyte")
	fakePod.SetName("flyte")

	t.Run("Successfully run sandbox cluster", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		errCh := make(chan error)
		sandboxConfig.DefaultConfig.Version = ""
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.GetSandboxImage(dind),
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
	t.Run("Successfully exit when sandbox cluster exist", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.GetSandboxImage(dind),
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
		mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
		reader, err := startSandbox(ctx, mockDocker, strings.NewReader("n"))
		assert.Nil(t, err)
		assert.Nil(t, reader)
	})
	t.Run("Successfully run sandbox cluster with source code", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Source = f.UserHomeDir()
		sandboxConfig.DefaultConfig.Version = ""
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: sandboxConfig.DefaultConfig.Source,
			Target: docker.Source,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.GetSandboxImage(dind),
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
		sandboxConfig.DefaultConfig.Version = ""
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
			Image:        docker.GetSandboxImage(dind),
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

		sha, err := githubutil.GetSHAFromVersion(sandboxConfig.DefaultConfig.Version, "flyte")
		assert.Nil(t, err)

		volumes := docker.Volumes
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.GetSandboxImage(fmt.Sprintf("%s-%s", dind, sha)),
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
	t.Run("Failed run sandbox cluster with wrong version", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Version = "v0.1444.0"
		sandboxConfig.DefaultConfig.Source = ""
		volumes := docker.Volumes
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.GetSandboxImage(dind),
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
			Image:        docker.GetSandboxImage(dind),
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
			Image:        docker.GetSandboxImage(dind),
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
			Image:        docker.GetSandboxImage(dind),
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
			Image:        docker.GetSandboxImage(dind),
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
		sandboxConfig.DefaultConfig.Version = ""
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: sandboxConfig.DefaultConfig.Source,
			Target: docker.Source,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.GetSandboxImage(dind),
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
		client := testclient.NewSimpleClientset()
		k8s.Client = client
		_, err := client.CoreV1().Pods("flyte").Create(ctx, &fakePod, v1.CreateOptions{})
		if err != nil {
			t.Error(err)
		}
		fakeNode.SetName("master")
		_, err = client.CoreV1().Nodes().Create(ctx, fakeNode, v1.CreateOptions{})
		if err != nil {
			t.Error(err)
		}
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        docker.GetSandboxImage(dind),
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
		sandboxConfig.DefaultConfig.Version = ""
		err = startSandboxCluster(ctx, []string{}, cmdCtx)
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
			Image:        docker.GetSandboxImage(dind),
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

func TestMonitorFlyteDeployment(t *testing.T) {
	t.Run("Monitor k8s deployment fail because of storage", func(t *testing.T) {
		ctx := context.Background()
		client := testclient.NewSimpleClientset()
		k8s.Client = client
		fakePod.SetName("flyte")
		fakePod.SetName("flyte")

		_, err := client.CoreV1().Pods("flyte").Create(ctx, &fakePod, v1.CreateOptions{})
		if err != nil {
			t.Error(err)
		}
		fakeNode.SetName("master")
		fakeNode.Spec.Taints = append(fakeNode.Spec.Taints, corev1.Taint{
			Effect: "NoSchedule",
			Key:    "node.kubernetes.io/disk-pressure",
		})
		_, err = client.CoreV1().Nodes().Create(ctx, fakeNode, v1.CreateOptions{})
		if err != nil {
			t.Error(err)
		}

		err = watchFlyteDeployment(ctx, client.CoreV1())
		assert.NotNil(t, err)

	})

	t.Run("Monitor k8s deployment success", func(t *testing.T) {
		ctx := context.Background()
		client := testclient.NewSimpleClientset()
		k8s.Client = client
		fakePod.SetName("flyte")
		fakePod.SetName("flyte")

		_, err := client.CoreV1().Pods("flyte").Create(ctx, &fakePod, v1.CreateOptions{})
		if err != nil {
			t.Error(err)
		}
		fakeNode.SetName("master")
		fakeNode.Spec.Taints = []corev1.Taint{}
		_, err = client.CoreV1().Nodes().Create(ctx, fakeNode, v1.CreateOptions{})
		if err != nil {
			t.Error(err)
		}

		err = watchFlyteDeployment(ctx, client.CoreV1())
		assert.Nil(t, err)

	})

}

func TestGetFlyteDeploymentCount(t *testing.T) {

	ctx := context.Background()
	client := testclient.NewSimpleClientset()
	c, err := getFlyteDeployment(ctx, client.CoreV1())
	assert.Nil(t, err)
	assert.Equal(t, 0, len(c.Items))
}

func TestGetNodeTaintStatus(t *testing.T) {
	t.Run("Check node taint with success", func(t *testing.T) {
		ctx := context.Background()
		client := testclient.NewSimpleClientset()
		fakeNode.SetName("master")
		_, err := client.CoreV1().Nodes().Create(ctx, fakeNode, v1.CreateOptions{})
		if err != nil {
			t.Error(err)
		}
		c, err := isNodeTainted(ctx, client.CoreV1())
		assert.Nil(t, err)
		assert.Equal(t, false, c)
	})
	t.Run("Check node taint with fail", func(t *testing.T) {
		ctx := context.Background()
		client := testclient.NewSimpleClientset()
		fakeNode.SetName("master")
		_, err := client.CoreV1().Nodes().Create(ctx, fakeNode, v1.CreateOptions{})
		if err != nil {
			t.Error(err)
		}
		node, err := client.CoreV1().Nodes().Get(ctx, "master", v1.GetOptions{})
		if err != nil {
			t.Error(err)
		}
		node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
			Effect: taintEffect,
			Key:    diskPressureTaint,
		})
		_, err = client.CoreV1().Nodes().Update(ctx, node, v1.UpdateOptions{})
		if err != nil {
			t.Error(err)
		}
		c, err := isNodeTainted(ctx, client.CoreV1())
		assert.Nil(t, err)
		assert.Equal(t, true, c)
	})
}

func TestGetSandboxImage(t *testing.T) {
	t.Run("Get Latest sandbox", func(t *testing.T) {
		image, err := getSandboxImage("")
		assert.Nil(t, err)
		assert.Equal(t, docker.GetSandboxImage(dind), image)
	})

	t.Run("Get sandbox image with version ", func(t *testing.T) {
		image, err := getSandboxImage("v0.14.0")
		assert.Nil(t, err)
		assert.Equal(t, true, strings.HasPrefix(image, docker.ImageName))
	})
	t.Run("Get sandbox image with wrong version ", func(t *testing.T) {
		_, err := getSandboxImage("v100.1.0")
		assert.NotNil(t, err)
	})
	t.Run("Get sandbox image with wrong version ", func(t *testing.T) {
		_, err := getSandboxImage("aaaaaa")
		assert.NotNil(t, err)
	})
	t.Run("Get sandbox image with version that is not supported", func(t *testing.T) {
		_, err := getSandboxImage("v0.10.0")
		assert.NotNil(t, err)
	})
}
