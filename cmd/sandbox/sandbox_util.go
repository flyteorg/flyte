package sandbox

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/enescakir/emoji"
	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
)

var (
	Kubeconfig         = f.FilePathJoin(f.UserHomeDir(), ".flyte", "k3s", "k3s.yaml")
	FlytectlConfig     = f.FilePathJoin(f.UserHomeDir(), ".flyte", "config.yaml")
	SuccessMessage     = "Flyte is ready! Flyte UI is available at http://localhost:30081/console"
	ImageName          = "ghcr.io/flyteorg/flyte-sandbox:dind"
	SandboxClusterName = "flyte-sandbox"
	Environment        = []string{"SANDBOX=1", "KUBERNETES_API_PORT=30086", "FLYTE_HOST=localhost:30081", "FLYTE_AWS_ENDPOINT=http://localhost:30084"}
)

func setupFlytectlConfig() error {
	response, err := http.Get("https://raw.githubusercontent.com/flyteorg/flytectl/master/config.yaml")
	if err != nil {
		return err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(FlytectlConfig, data, 0600)
	if err != nil {
		fmt.Printf("Please create ~/.flyte dir %v \n", emoji.ManTechnologist)
		return err
	}
	return nil
}

func configCleanup() error {
	err := os.Remove(FlytectlConfig)
	if err != nil {
		return err
	}
	err = os.RemoveAll(f.FilePathJoin(f.UserHomeDir(), ".flyte", "k3s"))
	if err != nil {
		return err
	}
	return nil
}

func getSandbox(cli *client.Client) *types.Container {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil
	}
	for _, v := range containers {
		if strings.Contains(v.Names[0], SandboxClusterName) {
			return &v
		}
	}
	return nil
}

func startContainer(cli *client.Client) (string, error) {
	ExposedPorts, PortBindings, _ := nat.ParsePortSpecs([]string{
		"127.0.0.1:30086:30086",
		"127.0.0.1:30081:30081",
		"127.0.0.1:30082:30082",
		"127.0.0.1:30084:30084",
	})
	r, err := cli.ImagePull(context.Background(), ImageName, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(os.Stdout, r); err != nil {
		return "", err
	}

	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Env:          Environment,
		Image:        ImageName,
		Tty:          false,
		ExposedPorts: ExposedPorts,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: f.FilePathJoin(f.UserHomeDir(), ".flyte"),
				Target: "/etc/rancher/",
			},
			// TODO (Yuvraj) Add flytectl config in sandbox and mount with host file system
			//{
			//	Type:   mount.TypeBind,
			//	Source: f.FilePathJoin(f.UserHomeDir(), ".flyte", "config.yaml"),
			//	Target: "/.flyte/",
			//},
		},
		PortBindings: PortBindings,
		Privileged:   true,
	}, nil,
		nil, SandboxClusterName)
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func watchError(cli *client.Client, id string) {
	statusCh, errCh := cli.ContainerWait(context.Background(), id, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}
}

func readLogs(cli *client.Client, id string) error {
	reader, err := cli.ContainerLogs(context.Background(), id, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: true,
		Follow:     true,
	})
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), SuccessMessage) {
			fmt.Printf("%v %v %v %v %v \n", emoji.ManTechnologist, SuccessMessage, emoji.Rocket, emoji.Rocket, emoji.PartyPopper)
			fmt.Printf("Please visit https://github.com/flyteorg/flytesnacks for more example %v \n", emoji.Rocket)
			fmt.Printf("Register all flytesnacks example by running 'flytectl register examples  -d development  -p flytesnacks' \n")
			break
		}
		fmt.Println(scanner.Text())
	}
	return nil
}
