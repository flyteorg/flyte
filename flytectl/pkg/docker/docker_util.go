package docker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/enescakir/emoji"
	"github.com/flyteorg/flyte/flytectl/clierrors"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/docker"
	cmdUtil "github.com/flyteorg/flyte/flytectl/pkg/commandutils"
	f "github.com/flyteorg/flyte/flytectl/pkg/filesystemutils"
	"github.com/moby/term"
)

var (
	FlyteSandboxConfigDir          = f.FilePathJoin(f.UserHomeDir(), ".flyte", "sandbox")
	Kubeconfig                     = f.FilePathJoin(FlyteSandboxConfigDir, "kubeconfig")
	SandboxKubeconfig              = f.FilePathJoin(f.UserHomeDir(), ".flyte", "k3s", "k3s.yaml")
	SuccessMessage                 = "Deploying Flyte..."
	FlyteSandboxClusterName        = "flyte-sandbox"
	FlyteSandboxVolumeName         = "flyte-sandbox"
	FlyteSandboxInternalDir        = "/var/lib/flyte"
	FlyteSandboxInternalConfigDir  = f.FilePathJoin(FlyteSandboxInternalDir, "config")
	FlyteSandboxInternalStorageDir = f.FilePathJoin(FlyteSandboxInternalDir, "storage")
	Environment                    = []string{"SANDBOX=1", "KUBERNETES_API_PORT=30086", "FLYTE_HOST=localhost:30081", "FLYTE_AWS_ENDPOINT=http://localhost:30084"}
	Source                         = "/root"
	K3sDir                         = "/etc/rancher/"
	Client                         Docker
	Volumes                        = []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: f.FilePathJoin(f.UserHomeDir(), ".flyte"),
			Target: K3sDir,
		},
	}
	ExecConfig = types.ExecConfig{
		AttachStderr: true,
		Tty:          true,
		WorkingDir:   "/",
		AttachStdout: true,
		Cmd:          []string{},
	}
	StdWriterPrefixLen = 8
	StartingBufLen     = 32*1024 + StdWriterPrefixLen + 1
	ExtraHosts         = []string{"host.docker.internal:host-gateway"}
)

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

// GetSandbox will return sandbox container if it exist
func GetSandbox(ctx context.Context, cli Docker) (*types.Container, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}
	for _, v := range containers {
		if strings.TrimLeft(v.Names[0], "/") == FlyteSandboxClusterName {
			return &v, nil
		}
	}
	return nil, nil
}

// RemoveSandbox will remove sandbox container if exist
func RemoveSandbox(ctx context.Context, cli Docker, reader io.Reader) error {
	c, err := GetSandbox(ctx, cli)
	if err != nil {
		return err
	}

	if c != nil {
		if docker.DefaultConfig.Force || cmdUtil.AskForConfirmation("delete existing sandbox cluster", reader) {
			err := cli.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{
				Force: true,
			})
			return err
		}
		return errors.New(clierrors.ErrSandboxExists)
	}
	return nil
}

// GetDevPorts will return dev cluster (minio + postgres) ports
func GetDevPorts() (map[nat.Port]struct{}, map[nat.Port][]nat.PortBinding, error) {
	return nat.ParsePortSpecs([]string{
		"0.0.0.0:30082:30082", // K8s Dashboard Port
		"0.0.0.0:30084:30084", // Minio API Port
		"0.0.0.0:30086:30086", // K8s cluster
		"0.0.0.0:30088:30088", // Minio Console Port
		"0.0.0.0:30089:30089", // Postgres Port
	})
}

// GetSandboxPorts will return sandbox ports
func GetSandboxPorts() (map[nat.Port]struct{}, map[nat.Port][]nat.PortBinding, error) {
	return nat.ParsePortSpecs([]string{
		// Notice that two host ports are mapped to the same container port in the case of Flyteconsole, this is done to
		// support the generated URLs produced by pyflyte run
		"0.0.0.0:30080:30081", // Flyteconsole Port.
		"0.0.0.0:30081:30081", // Flyteadmin Port
		"0.0.0.0:30082:30082", // K8s Dashboard Port
		"0.0.0.0:30084:30084", // Minio API Port
		"0.0.0.0:30086:30086", // K8s cluster
		"0.0.0.0:30088:30088", // Minio Console Port
		"0.0.0.0:30089:30089", // Postgres Port
	})
}

// GetDemoPorts will return demo ports
func GetDemoPorts() (map[nat.Port]struct{}, map[nat.Port][]nat.PortBinding, error) {
	return nat.ParsePortSpecs([]string{
		"0.0.0.0:6443:6443",   // K3s API Port
		"0.0.0.0:30080:30080", // HTTP Port
		"0.0.0.0:30000:30000", // Registry Port
		"0.0.0.0:30001:30001", // Postgres Port
		"0.0.0.0:30002:30002", // Minio API Port (use HTTP port for minio console)
		"0.0.0.0:30003:30003", // Buildkit Port
	})
}

// PullDockerImage will Pull docker image
func PullDockerImage(ctx context.Context, cli Docker, image string, pullPolicy ImagePullPolicy,
	imagePullOptions ImagePullOptions, dryRun bool) error {

	if dryRun {
		PrintPullImage(image, imagePullOptions)
		return nil
	}

	var needsPull bool
	if pullPolicy == ImagePullPolicyAlways {
		needsPull = true
	} else {
		imageSummary, err := cli.ImageList(ctx, types.ImageListOptions{})
		if err != nil {
			return err
		}
		found := false
		for _, img := range imageSummary {
			for _, tags := range img.RepoTags {
				if image == tags {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		needsPull = !found
	}

	// Image already exists, nothing to do.
	if !needsPull {
		return nil
	}

	// Image needs to be pulled but pull policy prevents it
	if pullPolicy == ImagePullPolicyNever {
		return fmt.Errorf("Image does not exist, but image pull policy prevents pulling it: %s", image)
	}

	fmt.Printf("%v Pulling image %s\n", emoji.Whale, image)
	r, err := cli.ImagePull(ctx, image, types.ImagePullOptions{
		RegistryAuth: imagePullOptions.RegistryAuth,
		Platform:     imagePullOptions.Platform,
	})
	if err != nil {
		return err
	}
	defer r.Close()
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	return jsonmessage.DisplayJSONMessagesStream(r, os.Stderr, termFd, isTerm, nil)
}

// PrintPullImage helper function to print the sandbox pull image command
func PrintPullImage(image string, pullOptions ImagePullOptions) {
	fmt.Printf("%v Run the following command to pull the sandbox image from registry.\n", emoji.Sparkle)
	var sb strings.Builder
	sb.WriteString("docker pull  ")
	if len(pullOptions.Platform) > 0 {
		sb.WriteString(fmt.Sprintf("--platform %v ", pullOptions.Platform))
	}
	sb.WriteString(fmt.Sprintf("%v", image))
	fmt.Printf("	%v \n", sb.String())
}

// PrintRemoveContainer helper function to remove sandbox container
func PrintRemoveContainer(name string) {
	fmt.Printf("%v Run the following command to remove the existing sandbox\n", emoji.Sparkle)
	fmt.Printf("	docker container rm %v --force\n", name)
}

// PrintCreateContainer helper function to print the docker command to run
func PrintCreateContainer(volumes []mount.Mount, portBindings map[nat.Port][]nat.PortBinding, name, image string, environment []string) {
	var sb strings.Builder
	fmt.Printf("%v Run the following command to create new sandbox container\n", emoji.Sparkle)
	sb.WriteString("	docker create --privileged ")
	for portProto, bindings := range portBindings {
		srcPort := portProto.Port()
		for _, binding := range bindings {
			sb.WriteString(fmt.Sprintf("-p %v:%v:%v ", binding.HostIP, srcPort, binding.HostPort))
		}
	}
	for _, env := range environment {
		sb.WriteString(fmt.Sprintf("--env %v ", env))
	}

	for _, volume := range volumes {
		sb.WriteString(fmt.Sprintf("--mount type=%v,source=%v,target=%v ", volume.Type, volume.Source, volume.Target))
	}
	sb.WriteString(fmt.Sprintf("--name %v ", name))
	sb.WriteString(fmt.Sprintf("%v", image))
	fmt.Printf("%v\n", sb.String())
	fmt.Printf("%v Run the following command to start the sandbox container\n", emoji.Sparkle)
	fmt.Printf("	docker start %v\n", name)
	fmt.Printf("%v Run the following command to check the logs and monitor the sandbox container and make sure there are no error during startup and then visit flyteconsole\n", emoji.EightSpokedAsterisk)
	fmt.Printf("	docker logs -f %v\n", name)
}

// StartContainer will create and start docker container
func StartContainer(ctx context.Context, cli Docker, volumes []mount.Mount, exposedPorts map[nat.Port]struct{},
	portBindings map[nat.Port][]nat.PortBinding, name, image string, additionalEnvVars []string, dryRun bool) (string, error) {

	// Append the additional env variables to the default list of env
	Environment = append(Environment, additionalEnvVars...)
	if dryRun {
		PrintCreateContainer(volumes, portBindings, name, image, Environment)
		return "", nil
	}
	fmt.Printf("%v Starting container... %v %v\n", emoji.FactoryWorker, emoji.Hammer, emoji.Wrench)
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Env:          Environment,
		Image:        image,
		Tty:          false,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		Mounts:       volumes,
		PortBindings: portBindings,
		Privileged:   true,
		ExtraHosts:   ExtraHosts, // add it because linux machine doesn't have this host name by default
	}, nil,
		nil, name)

	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}
	return resp.ID, nil
}

// CopyContainerFile try to create the container, see if the source file is there, copy it to the destination
func CopyContainerFile(ctx context.Context, cli Docker, source, destination, name, image string) error {
	resp, err := cli.ContainerCreate(ctx, &container.Config{Image: image}, &container.HostConfig{}, nil, nil, name)
	if err != nil {
		return err
	}
	var removeErr error
	defer func() {
		removeErr = cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{
			Force: true,
		})
	}()
	_, err = cli.ContainerStatPath(ctx, resp.ID, source)
	if err != nil {
		return err
	}
	reader, _, err := cli.CopyFromContainer(ctx, resp.ID, source)
	if err != nil {
		return err
	}
	tarFile := destination + ".tar"
	outFile, err := os.Create(tarFile)
	if err != nil {
		return err
	}
	defer outFile.Close()
	defer reader.Close()
	_, err = io.Copy(outFile, reader)
	if err != nil {
		return err
	}
	r, _ := os.Open(tarFile)
	err = f.ExtractTar(r, destination)
	if err != nil {
		return err
	}
	return removeErr
}

// ReadLogs will return io scanner for reading the logs of a container
func ReadLogs(ctx context.Context, cli Docker, id string) (*bufio.Scanner, error) {
	reader, err := cli.ContainerLogs(ctx, id, container.LogsOptions{
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
			return true
		}
		fmt.Println(reader.Text())
	}
	return false
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
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, resp.Reader)
	if err != nil {
		return err
	}
	return nil
}

func PrintCreateVolume(name string) {
	fmt.Printf("%v Run the following command to create a volume\n", emoji.Sparkle)
	fmt.Printf("	docker volume create %v\n", name)
}

func GetOrCreateVolume(
	ctx context.Context, cli Docker, volumeName string, dryRun bool,
) (*volume.Volume, error) {
	if dryRun {
		PrintCreateVolume(volumeName)
		return nil, nil
	}

	lo := volume.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "name", Value: fmt.Sprintf("^%s$", volumeName)},
		),
	}
	resp, err := cli.VolumeList(ctx, lo)
	if err != nil {
		return nil, err
	}
	switch len(resp.Volumes) {
	case 0:
		v, err := cli.VolumeCreate(ctx, volume.CreateOptions{Name: volumeName})
		if err != nil {
			return nil, err
		}
		return &v, nil
	case 1:
		return resp.Volumes[0], nil
	default:
		// We don't expect to ever arrive at this point
		return nil, fmt.Errorf("unexpected error - found multiple volumes with name: %s", volumeName)
	}
}
