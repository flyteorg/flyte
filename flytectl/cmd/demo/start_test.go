package demo

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin"

	"github.com/flyteorg/flytectl/pkg/githubutil"

	"github.com/flyteorg/flytectl/pkg/k8s"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	sandboxConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/docker/mocks"
	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
	k8sMocks "github.com/flyteorg/flytectl/pkg/k8s/mocks"
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
  name: default
contexts:
- context:
    cluster: default
    user: default
    namespace: bar
  name: default
current-context: default
kind: Config
users:
- name: default
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

func TestStartDemoFunc(t *testing.T) {
	p1, p2, _ := docker.GetSandboxPorts()
	assert.Nil(t, util.SetupFlyteDir())
	assert.Nil(t, os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte", "k3s"), os.ModePerm))
	assert.Nil(t, ioutil.WriteFile(docker.Kubeconfig, []byte(content), os.ModePerm))

	fakePod.SetName("flyte")

	t.Run("Successfully run demo cluster", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		errCh := make(chan error)
		sandboxConfig.DefaultConfig.Version = "v0.19.1"
		bodyStatus := make(chan container.ContainerWaitOKBody)
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", sandboxConfig.DefaultConfig.Version, demoImageName, false)
		assert.Nil(t, err)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		_, err = startDemo(ctx, mockDocker, os.Stdin)
		assert.Nil(t, err)
	})
	t.Run("Successfully exit when demo cluster exist", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		errCh := make(chan error)
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		reader, err := startDemo(ctx, mockDocker, strings.NewReader("n"))
		assert.Nil(t, err)
		assert.Nil(t, reader)
	})
	t.Run("Successfully run demo cluster with source code", func(t *testing.T) {
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
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		_, err = startDemo(ctx, mockDocker, os.Stdin)
		assert.Nil(t, err)
	})
	t.Run("Successfully run demo cluster with abs path of source code", func(t *testing.T) {
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
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		_, err = startDemo(ctx, mockDocker, os.Stdin)
		assert.Nil(t, err)
	})
	t.Run("Successfully run demo cluster with specific version", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Version = "v0.18.0"
		sandboxConfig.DefaultConfig.Source = ""

		image, _, err := githubutil.GetFullyQualifiedImageName("sha", sandboxConfig.DefaultConfig.Version, demoImageName, false)
		assert.Nil(t, err)
		volumes := docker.Volumes
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		_, err = startDemo(ctx, mockDocker, os.Stdin)
		assert.Nil(t, err)
	})
	t.Run("Failed run demo cluster with wrong version", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Version = "v0.1444.0"
		sandboxConfig.DefaultConfig.Source = ""
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		volumes := docker.Volumes
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		_, err = startDemo(ctx, mockDocker, os.Stdin)
		assert.NotNil(t, err)
	})
	t.Run("Error in pulling image", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		sandboxConfig.DefaultConfig.Source = f.UserHomeDir()
		volumes := docker.Volumes
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: sandboxConfig.DefaultConfig.Source,
			Target: docker.Source,
		})
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		_, err = startDemo(ctx, mockDocker, os.Stdin)
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
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		_, err = startDemo(ctx, mockDocker, strings.NewReader("y"))
		assert.NotNil(t, err)
	})
	t.Run("Error in start container", func(t *testing.T) {
		ctx := context.Background()
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker := &mocks.Docker{}
		sandboxConfig.DefaultConfig.Source = ""
		sandboxConfig.DefaultConfig.Version = ""
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		_, err = startDemo(ctx, mockDocker, os.Stdin)
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
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		_, err = startDemo(ctx, mockDocker, os.Stdin)
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
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		_, err = startDemo(ctx, mockDocker, os.Stdin)
		assert.Nil(t, err)
	})
	t.Run("Successfully run demo cluster command", func(t *testing.T) {
		mockOutStream := new(io.Writer)
		ctx := context.Background()
		cmdCtx := cmdCore.NewCommandContext(admin.InitializeMockClientset(), *mockOutStream)
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
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		mockK8sContextMgr := &k8sMocks.ContextOps{}
		docker.Client = mockDocker
		sandboxConfig.DefaultConfig.Source = ""
		sandboxConfig.DefaultConfig.Version = ""
		k8s.ContextMgr = mockK8sContextMgr
		mockK8sContextMgr.OnCopyContextMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		err = startDemoCluster(ctx, []string{}, cmdCtx)
		assert.Nil(t, err)
	})
	t.Run("Error in running demo cluster command", func(t *testing.T) {
		mockOutStream := new(io.Writer)
		ctx := context.Background()
		cmdCtx := cmdCore.NewCommandContext(admin.InitializeMockClientset(), *mockOutStream)
		mockDocker := &mocks.Docker{}
		errCh := make(chan error)
		bodyStatus := make(chan container.ContainerWaitOKBody)
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          docker.Environment,
			Image:        image,
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
		err = startDemoCluster(ctx, []string{}, cmdCtx)
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

func TestGetDemoImage(t *testing.T) {
	t.Run("Get Latest demo cluster", func(t *testing.T) {
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "", demoImageName, false)
		assert.Nil(t, err)
		assert.Equal(t, true, strings.HasPrefix(image, "cr.flyte.org/flyteorg/flyte-sandbox-lite:sha-"))
	})

	t.Run("Get demo image with version ", func(t *testing.T) {
		image, _, err := githubutil.GetFullyQualifiedImageName("sha", "v0.14.0", demoImageName, false)
		assert.Nil(t, err)
		assert.Equal(t, true, strings.HasPrefix(image, demoImageName))
	})
	t.Run("Get demo image with wrong version ", func(t *testing.T) {
		_, _, err := githubutil.GetFullyQualifiedImageName("sha", "v100.1.0", demoImageName, false)
		assert.NotNil(t, err)
	})
	t.Run("Get demo image with wrong version ", func(t *testing.T) {
		_, _, err := githubutil.GetFullyQualifiedImageName("sha", "aaaaaa", demoImageName, false)
		assert.NotNil(t, err)
	})
	t.Run("Get demo image with version that is not supported", func(t *testing.T) {
		_, _, err := githubutil.GetFullyQualifiedImageName("sha", "v0.10.0", demoImageName, false)
		assert.NotNil(t, err)
	})

}
