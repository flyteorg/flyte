package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flyteorg/flytectl/pkg/configutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/enescakir/emoji"
	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
)

var (
	Kubeconfig              = f.FilePathJoin(f.UserHomeDir(), ".flyte", "k3s", "k3s.yaml")
	SuccessMessage          = "Flyte is ready! Flyte UI is available at http://localhost:30081/console"
	ImageName               = "cr.flyte.org/flyteorg/flyte-sandbox:dind"
	FlyteSandboxClusterName = "flyte-sandbox"
	Environment             = []string{"SANDBOX=1", "KUBERNETES_API_PORT=30086", "FLYTE_HOST=localhost:30081", "FLYTE_AWS_ENDPOINT=http://localhost:30084"}
	Source                  = "/root"
	K3sDir                  = "/etc/rancher/"
	Client                  Docker
	Volumes                 = []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: f.FilePathJoin(f.UserHomeDir(), ".flyte"),
			Target: K3sDir,
		},
	}
	ExecConfig = types.ExecConfig{
		AttachStderr: true,
		Tty:          true,
		WorkingDir:   Source,
		AttachStdout: true,
		Cmd:          []string{},
	}
	StdWriterPrefixLen = 8
	StartingBufLen     = 32*1024 + StdWriterPrefixLen + 1
)

// GetSandbox will return sandbox container if it exist
func GetSandbox(ctx context.Context, cli Docker) *types.Container {
	containers, _ := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	for _, v := range containers {
		if strings.Contains(v.Names[0], FlyteSandboxClusterName) {
			return &v
		}
	}
	return nil
}

// RemoveSandbox will remove sandbox container if exist
func RemoveSandbox(ctx context.Context, cli Docker, reader io.Reader) error {
	if c := GetSandbox(ctx, cli); c != nil {
		if cmdUtil.AskForConfirmation("delete existing sandbox cluster", reader) {
			err := cli.ContainerRemove(context.Background(), c.ID, types.ContainerRemoveOptions{
				Force: true,
			})
			return err
		}
		return nil
	}
	return nil
}

// GetSandboxPorts will return sandbox ports
func GetSandboxPorts() (map[nat.Port]struct{}, map[nat.Port][]nat.PortBinding, error) {
	return nat.ParsePortSpecs([]string{
		"127.0.0.1:30086:30086",
		"127.0.0.1:30081:30081",
		"127.0.0.1:30082:30082",
		"127.0.0.1:30084:30084",
	})
}

// PullDockerImage will Pull docker image
func PullDockerImage(ctx context.Context, cli Docker, image string) error {
	r, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, r)
	return err
}

//StartContainer will create and start docker container
func StartContainer(ctx context.Context, cli Docker, volumes []mount.Mount, exposedPorts map[nat.Port]struct{}, portBindings map[nat.Port][]nat.PortBinding, name, image string) (string, error) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Env:          Environment,
		Image:        image,
		Tty:          false,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		Mounts:       volumes,
		PortBindings: portBindings,
		Privileged:   true,
	}, nil,
		nil, name)

	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	return resp.ID, nil
}

// WatchError will return channel for watching errors of a container
func WatchError(ctx context.Context, cli Docker, id string) (<-chan container.ContainerWaitOKBody, <-chan error) {
	return cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
}

// ReadLogs will return io scanner for reading the logs of a container
func ReadLogs(ctx context.Context, cli Docker, id string) (*bufio.Scanner, error) {
	reader, err := cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: true,
		Follow:     true,
	})
	if err != nil {
		return nil, err
	}
	return bufio.NewScanner(reader), nil
}

// WaitForSandbox will wait until it doesn't get success message
func WaitForSandbox(reader *bufio.Scanner, message string) bool {
	for reader.Scan() {
		if strings.Contains(reader.Text(), message) {
			kubeconfig := strings.Join([]string{
				"$KUBECONFIG",
				f.FilePathJoin(f.UserHomeDir(), ".kube", "config"),
				Kubeconfig,
			}, ":")

			fmt.Printf("%v %v %v %v %v \n", emoji.ManTechnologist, message, emoji.Rocket, emoji.Rocket, emoji.PartyPopper)
			fmt.Printf("Please visit https://github.com/flyteorg/flytesnacks for more example %v \n", emoji.Rocket)
			fmt.Printf("Register all flytesnacks example by running 'flytectl register examples  -d development  -p flytesnacks' \n")

			fmt.Printf("Add KUBECONFIG and FLYTECTL_CONFIG to your environment variable \n")
			fmt.Printf("export KUBECONFIG=%v \n", kubeconfig)
			fmt.Printf("export FLYTECTL_CONFIG=%v \n", configutil.FlytectlConfig)
			return true
		}
		fmt.Println(reader.Text())
	}
	return false
}

// GetDockerClient will returns the docker client
func GetDockerClient() (Docker, error) {
	if Client == nil {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("%v Please Check your docker client %v \n", emoji.GrimacingFace, emoji.Whale)
			return nil, err
		}
		return cli, nil
	}
	return Client, nil
}

// ExecCommend will execute a command in container and returns an execution id
func ExecCommend(ctx context.Context, cli Docker, containerID string, command []string) (types.IDResponse, error) {
	ExecConfig.Cmd = command
	r, err := cli.ContainerExecCreate(ctx, containerID, ExecConfig)
	if err != nil {
		return types.IDResponse{}, err
	}
	return r, err
}

func InspectExecResp(ctx context.Context, cli Docker, containerID string) error {
	resp, err := cli.ContainerExecAttach(ctx, containerID, types.ExecStartCheck{})
	if err != nil {
		return err
	}
	s := bufio.NewScanner(resp.Reader)
	for s.Scan() {
		fmt.Println(s.Text())
	}
	return nil
}
