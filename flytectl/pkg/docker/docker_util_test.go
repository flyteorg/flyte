package docker

import (
	"archive/tar"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/docker/mocks"
	f "github.com/flyteorg/flyte/flytectl/pkg/filesystemutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	containers []types.Container
	imageName  = "cr.flyte.org/flyteorg/flyte-sandbox"
)

func setupSandbox() {
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

func dummyReader() io.ReadCloser {
	return io.NopCloser(strings.NewReader(""))
}

func TestGetSandbox(t *testing.T) {
	setupSandbox()
	t.Run("Successfully get sandbox container", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()

		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(containers, nil)
		c, err := GetSandbox(ctx, mockDocker)
		assert.Equal(t, c.Names[0], FlyteSandboxClusterName)
		assert.Nil(t, err)
	})

	t.Run("Successfully get sandbox container with zero result", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()

		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		c, err := GetSandbox(ctx, mockDocker)
		assert.Nil(t, c)
		assert.Nil(t, err)
	})

	t.Run("Error in get sandbox container", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()

		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(containers, nil)
		mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(nil)
		err := RemoveSandbox(ctx, mockDocker, strings.NewReader("y"))
		assert.Nil(t, err)
	})

}

func TestRemoveSandboxWithNoReply(t *testing.T) {
	setupSandbox()
	t.Run("Successfully remove sandbox container", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()

		// Verify the attributes
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(containers, nil)
		mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(nil)
		err := RemoveSandbox(ctx, mockDocker, strings.NewReader("n"))
		assert.NotNil(t, err)

		docker.DefaultConfig.Force = true
		err = RemoveSandbox(ctx, mockDocker, strings.NewReader(""))
		assert.Nil(t, err)
	})

	t.Run("Successfully remove sandbox container with zero sandbox containers are running", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()

		// Verify the attributes
		mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{}, nil)
		mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(nil)
		err := RemoveSandbox(ctx, mockDocker, strings.NewReader("n"))
		assert.Nil(t, err)
	})

}

func TestPullDockerImage(t *testing.T) {
	t.Run("Successful pull existing image with ImagePullPolicyAlways", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()
		// Verify the attributes
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnImageListMatch(ctx, types.ImageListOptions{}).Return([]image.Summary{{RepoTags: []string{"nginx:latest"}}}, nil)
		err := PullDockerImage(ctx, mockDocker, "nginx:latest", ImagePullPolicyAlways, ImagePullOptions{}, false)
		assert.Nil(t, err)
	})

	t.Run("Successful pull non-existent image with ImagePullPolicyAlways", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()
		// Verify the attributes
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnImageListMatch(ctx, types.ImageListOptions{}).Return([]image.Summary{}, nil)
		err := PullDockerImage(ctx, mockDocker, "nginx:latest", ImagePullPolicyAlways, ImagePullOptions{}, false)
		assert.Nil(t, err)
	})

	t.Run("Error in pull image", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()
		// Verify the attributes
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), fmt.Errorf("error"))
		err := PullDockerImage(ctx, mockDocker, "nginx:latest", ImagePullPolicyAlways, ImagePullOptions{}, false)
		assert.NotNil(t, err)
	})

	t.Run("Success pull non-existent image with ImagePullPolicyIfNotPresent", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()
		// Verify the attributes
		mockDocker.OnImagePullMatch(ctx, mock.Anything, types.ImagePullOptions{}).Return(dummyReader(), nil)
		mockDocker.OnImageListMatch(ctx, types.ImageListOptions{}).Return([]image.Summary{}, nil)
		err := PullDockerImage(ctx, mockDocker, "nginx:latest", ImagePullPolicyIfNotPresent, ImagePullOptions{}, false)
		assert.Nil(t, err)
	})

	t.Run("Success skip existing image with ImagePullPolicyIfNotPresent", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()
		mockDocker.OnImageListMatch(ctx, types.ImageListOptions{}).Return([]image.Summary{{RepoTags: []string{"nginx:latest"}}}, nil)
		err := PullDockerImage(ctx, mockDocker, "nginx:latest", ImagePullPolicyIfNotPresent, ImagePullOptions{}, false)
		assert.Nil(t, err)
	})

	t.Run("Success skip existing image with ImagePullPolicyNever", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()
		mockDocker.OnImageListMatch(ctx, types.ImageListOptions{}).Return([]image.Summary{{RepoTags: []string{"nginx:latest"}}}, nil)
		err := PullDockerImage(ctx, mockDocker, "nginx:latest", ImagePullPolicyNever, ImagePullOptions{}, false)
		assert.Nil(t, err)
	})

	t.Run("Error non-existent image with ImagePullPolicyNever", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()
		mockDocker.OnImageListMatch(ctx, types.ImageListOptions{}).Return([]image.Summary{}, nil)
		err := PullDockerImage(ctx, mockDocker, "nginx:latest", ImagePullPolicyNever, ImagePullOptions{}, false)
		assert.ErrorContains(t, err, "Image does not exist, but image pull policy prevents pulling it")
	})
}

func TestStartContainer(t *testing.T) {
	p1, p2, _ := GetSandboxPorts()

	t.Run("Successfully create a container", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		ctx := context.Background()

		// Verify the attributes
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          Environment,
			Image:        imageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       Volumes,
			PortBindings: p2,
			Privileged:   true,
			ExtraHosts:   ExtraHosts,
		}, nil, nil, mock.Anything).Return(container.CreateResponse{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		id, err := StartContainer(ctx, mockDocker, Volumes, p1, p2, "nginx", imageName, nil, false)
		assert.Nil(t, err)
		assert.Greater(t, len(id), 0)
		assert.Equal(t, id, "Hello")
	})

	t.Run("Successfully create a container with Env", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		ctx := context.Background()
		// Setup additional env
		additionalEnv := []string{"a=1", "b=2"}
		expectedEnv := append(Environment, "a=1")
		expectedEnv = append(expectedEnv, "b=2")

		// Verify the attributes
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          expectedEnv,
			Image:        imageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       Volumes,
			PortBindings: p2,
			Privileged:   true,
			ExtraHosts:   ExtraHosts,
		}, nil, nil, mock.Anything).Return(container.CreateResponse{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		id, err := StartContainer(ctx, mockDocker, Volumes, p1, p2, "nginx", imageName, additionalEnv, false)
		assert.Nil(t, err)
		assert.Greater(t, len(id), 0)
		assert.Equal(t, id, "Hello")
		assert.Equal(t, expectedEnv, Environment)
	})

	t.Run("Error in creating container", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		ctx := context.Background()

		// Verify the attributes
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          Environment,
			Image:        imageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       Volumes,
			PortBindings: p2,
			Privileged:   true,
			ExtraHosts:   ExtraHosts,
		}, nil, nil, mock.Anything).Return(container.CreateResponse{
			ID: "",
		}, fmt.Errorf("error"))
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(nil)
		id, err := StartContainer(ctx, mockDocker, Volumes, p1, p2, "nginx", imageName, nil, false)
		assert.NotNil(t, err)
		assert.Equal(t, len(id), 0)
		assert.Equal(t, id, "")
	})

	t.Run("Error in start of a container", func(t *testing.T) {
		setupSandbox()
		mockDocker := &mocks.Docker{}
		ctx := context.Background()

		// Verify the attributes
		mockDocker.OnContainerCreate(ctx, &container.Config{
			Env:          Environment,
			Image:        imageName,
			Tty:          false,
			ExposedPorts: p1,
		}, &container.HostConfig{
			Mounts:       Volumes,
			PortBindings: p2,
			Privileged:   true,
			ExtraHosts:   ExtraHosts,
		}, nil, nil, mock.Anything).Return(container.CreateResponse{
			ID: "Hello",
		}, nil)
		mockDocker.OnContainerStart(ctx, "Hello", container.StartOptions{}).Return(fmt.Errorf("error"))
		id, err := StartContainer(ctx, mockDocker, Volumes, p1, p2, "nginx", imageName, nil, false)
		assert.NotNil(t, err)
		assert.Equal(t, len(id), 0)
		assert.Equal(t, id, "")
	})
}

func TestReadLogs(t *testing.T) {
	setupSandbox()

	t.Run("Successfully read logs", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, nil)
		_, err := ReadLogs(ctx, mockDocker, "test")
		assert.Nil(t, err)
	})

	t.Run("Error in reading logs", func(t *testing.T) {
		mockDocker := &mocks.Docker{}
		ctx := context.Background()
		mockDocker.OnContainerLogsMatch(ctx, mock.Anything, container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Timestamps: true,
			Follow:     true,
		}).Return(nil, fmt.Errorf("error"))
		_, err := ReadLogs(ctx, mockDocker, "test")
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

func TestGetOrCreateVolume(t *testing.T) {
	t.Run("VolumeExists", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		expected := &volume.Volume{Name: "test"}

		mockDocker.OnVolumeList(ctx, volume.ListOptions{Filters: filters.NewArgs(filters.KeyValuePair{Key: "name", Value: "^test$"})}).Return(volume.ListResponse{Volumes: []*volume.Volume{expected}}, nil)
		actual, err := GetOrCreateVolume(ctx, mockDocker, "test", false)
		assert.Equal(t, expected, actual, "volumes should match")
		assert.Nil(t, err)
	})
	t.Run("VolumeDoesNotExist", func(t *testing.T) {
		ctx := context.Background()
		mockDocker := &mocks.Docker{}
		expected := volume.Volume{Name: "test"}

		mockDocker.OnVolumeList(ctx, volume.ListOptions{Filters: filters.NewArgs(filters.KeyValuePair{Key: "name", Value: "^test$"})}).Return(volume.ListResponse{Volumes: []*volume.Volume{}}, nil)
		mockDocker.OnVolumeCreate(ctx, volume.CreateOptions{Name: "test"}).Return(expected, nil)
		actual, err := GetOrCreateVolume(ctx, mockDocker, "test", false)
		assert.Equal(t, expected, *actual, "volumes should match")
		assert.Nil(t, err)
	})

}

func TestDemoPorts(t *testing.T) {
	_, ports, _ := GetDemoPorts()
	assert.Equal(t, 6, len(ports))
}

func TestCopyFile(t *testing.T) {
	ctx := context.Background()
	// Create a fake tar file in tmp.
	fo, err := os.CreateTemp("", "sampledata")
	assert.NoError(t, err)
	tarWriter := tar.NewWriter(fo)
	err = tarWriter.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "flyte.yaml",
		Size:     4,
		Mode:     0640,
		ModTime:  time.Unix(1245206587, 0),
	})
	assert.NoError(t, err)
	cnt, err := tarWriter.Write([]byte("a: b"))
	assert.NoError(t, err)
	assert.Equal(t, 4, cnt)
	tarWriter.Close()
	fo.Close()

	image := "some:image"
	containerName := "my-container"

	t.Run("No errors", func(t *testing.T) {
		// Create reader of the tar file
		reader, err := os.Open(fo.Name())
		assert.NoError(t, err)
		// Create destination file name
		destDir, err := os.MkdirTemp("", "dest")
		assert.NoError(t, err)
		destination := filepath.Join(destDir, "destfile")

		// Mocks
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerCreate(
			ctx, &container.Config{Image: image}, &container.HostConfig{}, nil, nil, containerName).Return(
			container.CreateResponse{ID: containerName}, nil)
		mockDocker.OnContainerStatPath(ctx, containerName, "some source").Return(types.ContainerPathStat{}, nil)
		mockDocker.OnCopyFromContainer(ctx, containerName, "some source").Return(reader, types.ContainerPathStat{}, nil)
		mockDocker.OnContainerRemove(ctx, containerName, container.RemoveOptions{Force: true}).Return(nil)
		assert.Nil(t, err)

		// Run
		err = CopyContainerFile(ctx, mockDocker, "some source", destination, containerName, image)
		assert.NoError(t, err)

		// Read the file and make sure it's correct
		strBytes, err := os.ReadFile(destination)
		assert.NoError(t, err)
		assert.Equal(t, "a: b", string(strBytes))
	})

	t.Run("Erroring on stat", func(t *testing.T) {
		myErr := fmt.Errorf("erroring on stat")

		// Mocks
		mockDocker := &mocks.Docker{}
		mockDocker.OnContainerCreate(
			ctx, &container.Config{Image: image}, &container.HostConfig{}, nil, nil, containerName).Return(
			container.CreateResponse{ID: containerName}, nil)
		mockDocker.OnContainerStatPath(ctx, containerName, "some source").Return(types.ContainerPathStat{}, myErr)
		mockDocker.OnContainerRemove(ctx, containerName, container.RemoveOptions{Force: true}).Return(nil)
		assert.Nil(t, err)

		// Run
		err = CopyContainerFile(ctx, mockDocker, "some source", "", containerName, image)
		assert.Equal(t, myErr, err)
	})
}
