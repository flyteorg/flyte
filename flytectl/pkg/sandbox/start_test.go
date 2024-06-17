package sandbox

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	sandboxCmdConfig "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/sandbox"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/docker/mocks"
	f "github.com/flyteorg/flyte/flytectl/pkg/filesystemutils"
	ghutil "github.com/flyteorg/flyte/flytectl/pkg/github"
	ghMocks "github.com/flyteorg/flyte/flytectl/pkg/github/mocks"
	"github.com/flyteorg/flyte/flytectl/pkg/k8s"
	k8sMocks "github.com/flyteorg/flyte/flytectl/pkg/k8s/mocks"
	"github.com/flyteorg/flyte/flytectl/pkg/util"
	"github.com/google/go-github/v42/github"
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
		Phase: corev1.PodRunning,
		Conditions: []corev1.PodCondition{
			{
				Type:   corev1.PodReady,
				Status: corev1.ConditionTrue,
			},
			{
				Type:   corev1.ContainersReady,
				Status: corev1.ConditionTrue,
			},
		},
	},
}

var (
	githubMock *ghMocks.GHRepoService
	ctx        context.Context
	mockDocker *mocks.Docker
)

func sandboxSetup() {
	ctx = context.Background()
	mockDocker = &mocks.Docker{}
	errCh := make(chan error)
	sandboxCmdConfig.DefaultConfig.Version = "v0.19.1"
	bodyStatus := make(chan container.WaitResponse)
	githubMock = &ghMocks.GHRepoService{}
	sandboxCmdConfig.DefaultConfig.Image = "dummyimage"
	mockDocker.OnVolumeList(ctx, volume.ListOptions{Filters: filters.NewArgs(filters.KeyValuePair{Key: "name", Value: fmt.Sprintf("^%s$", docker.FlyteSandboxVolumeName)})}).Return(volume.ListResponse{Volumes: []*volume.Volume{}}, nil)
	mockDocker.OnVolumeCreate(ctx, volume.CreateOptions{Name: docker.FlyteSandboxVolumeName}).Return(volume.Volume{Name: docker.FlyteSandboxVolumeName}, nil)
	mockDocker.OnContainerCreateMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(container.CreateResponse{
		ID: "Hello",
	}, nil)

	mockDocker.OnContainerWaitMatch(ctx, mock.Anything, container.WaitConditionNotRunning).Return(bodyStatus, errCh)
}

func dummyReader() io.ReadCloser {
	return io.NopCloser(strings.NewReader(""))
}

func TestStartFunc(t *testing.T) {
	defaultImagePrefix := "dind"
	exposedPorts, portBindings, _ := docker.GetSandboxPorts()
	config := sandboxCmdConfig.DefaultConfig
	config.Image = "dummyimage"
	config.ImagePullOptions = docker.ImagePullOptions{
		RegistryAuth: "",
		Platform:     "",
	}
	config.Dev = true
	config.DisableAgent = true
	assert.Nil(t, util.SetupFlyteDir())
	assert.Nil(t, os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte", "state"), os.ModePerm))
	assert.Nil(t, ioutil.WriteFile(docker.Kubeconfig, []byte(content), os.ModePerm))

	fakePod.SetName("flyte")

	t.Run("Successfully run demo cluster", func(t *testing.T) {
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		mockDocker.OnVolumeList(ctx, volume.ListOptions{Filters: filters.NewArgs(filters.KeyValuePair{Key: mock.Anything, Value: mock.Anything})}).Return(volume.ListResponse{Volumes: []*volume.Volume{}}, nil)
		mockDocker.OnVolumeCreate(ctx, volume.CreateOptions{Name: mock.Anything}).Return(volume.Volume{}, nil)
		_, err := startSandbox(ctx, mockDocker, githubMock, dummyReader(), config, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.Nil(t, err)
	})
	t.Run("Successfully exit when demo cluster exist", func(t *testing.T) {
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{
			{
				ID: docker.FlyteSandboxClusterName,
				Names: []string{
					docker.FlyteSandboxClusterName,
				},
			},
		}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		reader, err := startSandbox(ctx, mockDocker, githubMock, strings.NewReader("n"), config, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.Nil(t, err)
		assert.Nil(t, reader)
	})
	t.Run("Successfully run demo cluster with source code", func(t *testing.T) {
		sandboxCmdConfig.DefaultConfig.DeprecatedSource = f.UserHomeDir()
		sandboxCmdConfig.DefaultConfig.Version = ""
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		_, err := startSandbox(ctx, mockDocker, githubMock, dummyReader(), config, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.Nil(t, err)
	})
	t.Run("Successfully run demo cluster with abs path of source code", func(t *testing.T) {
		sandboxCmdConfig.DefaultConfig.DeprecatedSource = "../"
		sandboxCmdConfig.DefaultConfig.Version = ""
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		_, err := startSandbox(ctx, mockDocker, githubMock, dummyReader(), config, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.Nil(t, err)
	})
	t.Run("Successfully run demo cluster with specific version", func(t *testing.T) {
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		sandboxCmdConfig.DefaultConfig.Image = ""
		tag := "v0.15.0"
		githubMock.OnGetReleaseByTagMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&github.RepositoryRelease{
			TagName: &tag,
		}, nil, nil)

		githubMock.OnGetCommitSHA1Match(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("dummySha", nil, nil)
		_, err := startSandbox(ctx, mockDocker, githubMock, dummyReader(), config, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.Nil(t, err)
	})
	t.Run("Failed run demo cluster with wrong version", func(t *testing.T) {
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		sandboxCmdConfig.DefaultConfig.Image = ""
		githubMock.OnGetReleaseByTagMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("non-existent-tag"))
		_, err := startSandbox(ctx, mockDocker, githubMock, dummyReader(), sandboxCmdConfig.DefaultConfig, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.NotNil(t, err)
		assert.Equal(t, "non-existent-tag", err.Error())
	})
	t.Run("Error in pulling image", func(t *testing.T) {
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), fmt.Errorf("failed to pull"))
		sandboxCmdConfig.DefaultConfig.Image = ""
		tag := "v0.15.0"
		githubMock.OnGetReleaseByTagMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&github.RepositoryRelease{
			TagName: &tag,
		}, nil, nil)

		githubMock.OnGetCommitSHA1Match(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("dummySha", nil, nil)
		_, err := startSandbox(ctx, mockDocker, githubMock, dummyReader(), sandboxCmdConfig.DefaultConfig, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.NotNil(t, err)
		assert.Equal(t, "failed to pull", err.Error())
	})
	t.Run("Error in  removing existing cluster", func(t *testing.T) {
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{
			{
				ID: docker.FlyteSandboxClusterName,
				Names: []string{
					docker.FlyteSandboxClusterName,
				},
			},
		}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(fmt.Errorf("failed to remove container"))
		_, err := startSandbox(ctx, mockDocker, githubMock, strings.NewReader("y"), sandboxCmdConfig.DefaultConfig, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.NotNil(t, err)
		assert.Equal(t, "failed to remove container", err.Error())
	})
	t.Run("Error in start container", func(t *testing.T) {
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(fmt.Errorf("failed to run container"))
		_, err := startSandbox(ctx, mockDocker, githubMock, dummyReader(), sandboxCmdConfig.DefaultConfig, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.NotNil(t, err)
		assert.Equal(t, "failed to run container", err.Error())
	})
	t.Run("Error in reading logs", func(t *testing.T) {
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, fmt.Errorf("failed to get container logs"))
		_, err := startSandbox(ctx, mockDocker, githubMock, dummyReader(), sandboxCmdConfig.DefaultConfig, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.NotNil(t, err)
		assert.Equal(t, "failed to get container logs", err.Error())
	})
	t.Run("Error in list container", func(t *testing.T) {
		sandboxSetup()
		mockDocker.OnContainerListMatch(mock.Anything, mock.Anything).Return([]types.Container{}, fmt.Errorf("failed to list containers"))
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		_, err := startSandbox(ctx, mockDocker, githubMock, dummyReader(), config, sandboxImageName, defaultImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
		assert.NotNil(t, err)
		assert.Equal(t, "failed to list containers", err.Error())
	})
	t.Run("Successfully run demo cluster command", func(t *testing.T) {
		//	mockOutStream := new(io.Writer)
		// cmdCtx := cmdCore.NewCommandContext(admin.InitializeMockClientset(), *mockOutStream)
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
		sandboxSetup()
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnImagePullMatch(mock.Anything, mock.Anything, mock.Anything).Return(dummyReader(), nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		stringReader := strings.NewReader(docker.SuccessMessage)
		reader := ioutil.NopCloser(stringReader)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(reader, nil)
		mockK8sContextMgr := &k8sMocks.ContextOps{}
		docker.Client = mockDocker
		sandboxCmdConfig.DefaultConfig.DeprecatedSource = ""
		sandboxCmdConfig.DefaultConfig.Version = ""
		k8s.ContextMgr = mockK8sContextMgr
		ghutil.Client = githubMock
		mockK8sContextMgr.OnCheckConfig().Return(nil)
		mockK8sContextMgr.OnCopyContextMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		err = StartSandboxCluster(context.Background(), []string{}, config)
		assert.Nil(t, err)
	})
	t.Run("Error in running demo cluster command", func(t *testing.T) {
		// mockOutStream := new(io.Writer)
		// cmdCtx := cmdCore.NewCommandContext(admin.InitializeMockClientset(), *mockOutStream)
		sandboxSetup()
		docker.Client = mockDocker
		mockDocker.OnContainerListMatch(mock.Anything, mock.Anything).Return([]types.Container{}, fmt.Errorf("failed to list containers"))
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		err := StartSandboxCluster(context.Background(), []string{}, config)
		assert.NotNil(t, err)
		err = StartDemoCluster(context.Background(), []string{}, config)
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

		err = WatchFlyteDeployment(ctx, client.CoreV1())
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

		err = WatchFlyteDeployment(ctx, client.CoreV1())
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
