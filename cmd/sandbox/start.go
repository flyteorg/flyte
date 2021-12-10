package sandbox

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/pkg/util/githubutil"

	"github.com/avast/retry-go"
	"github.com/olekukonko/tablewriter"
	corev1api "k8s.io/api/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/docker/docker/api/types/mount"
	"github.com/flyteorg/flytectl/pkg/configutil"
	"github.com/flyteorg/flytectl/pkg/k8s"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/enescakir/emoji"
	sandboxConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/util"
)

const (
	startShort = "Start the Flyte Sandbox cluster"
	startLong  = `
The Flyte Sandbox is a fully standalone minimal environment for running Flyte. It provides a simplified way of running Flyte sandbox as a single Docker container locally.  

Start sandbox cluster without any source code:
::

 flytectl sandbox start
	
Mount your source code repository inside sandbox:
::

 flytectl sandbox start --source=$HOME/flyteorg/flytesnacks 
	
Run specific version of Flyte. FlyteCTL sandbox only supports Flyte version available in the Github release, https://github.com/flyteorg/flyte/tags.
::

 flytectl sandbox start  --version=v0.14.0

Note: FlyteCTL sandbox is only supported for Flyte versions > v0.10.0

Specify a Flyte Sandbox compliant image with the registry. This is useful in case you want to use an image from your registry.
::

  flytectl sandbox start --image docker.io/my-override:latest

	
Specify a Flyte Sandbox image pull policy. Possible pull policy values are Always, IfNotPresent, or Never:
::

 flytectl sandbox start  --image docker.io/my-override:latest --imagePullPolicy Always
Usage
`
	k8sEndpoint             = "https://127.0.0.1:30086"
	flyteNamespace          = "flyte"
	flyteRepository         = "flyte"
	dind                    = "dind"
	sandboxSupportedVersion = "v0.10.0"
	diskPressureTaint       = "node.kubernetes.io/disk-pressure"
	taintEffect             = "NoSchedule"
)

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

func startSandboxCluster(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}

	reader, err := startSandbox(ctx, cli, os.Stdin)
	if err != nil {
		return err
	}
	if reader != nil {
		docker.WaitForSandbox(reader, docker.SuccessMessage)
	}

	if reader != nil {
		var k8sClient k8s.K8s
		err = retry.Do(
			func() error {
				k8sClient, err = k8s.GetK8sClient(docker.Kubeconfig, k8sEndpoint)
				return err
			},
			retry.Attempts(10),
		)
		if err != nil {
			return err
		}
		if err := watchFlyteDeployment(ctx, k8sClient.CoreV1()); err != nil {
			return err
		}
		util.PrintSandboxMessage()
	}
	return nil
}

func startSandbox(ctx context.Context, cli docker.Docker, reader io.Reader) (*bufio.Scanner, error) {
	fmt.Printf("%v Bootstrapping a brand new flyte cluster... %v %v\n", emoji.FactoryWorker, emoji.Hammer, emoji.Wrench)

	if err := docker.RemoveSandbox(ctx, cli, reader); err != nil {
		if err.Error() != clierrors.ErrSandboxExists {
			return nil, err
		}
		fmt.Printf("Existing details of your sandbox")
		util.PrintSandboxMessage()
		return nil, nil
	}

	if err := util.SetupFlyteDir(); err != nil {
		return nil, err
	}

	templateValues := configutil.ConfigTemplateSpec{
		Host:     "localhost:30081",
		Insecure: true,
	}
	if err := configutil.SetupConfig(configutil.FlytectlConfig, configutil.GetSandboxTemplate(), templateValues); err != nil {
		return nil, err
	}

	volumes := docker.Volumes
	if vol, err := mountVolume(sandboxConfig.DefaultConfig.Source, docker.Source); err != nil {
		return nil, err
	} else if vol != nil {
		volumes = append(volumes, *vol)
	}

	image, err := getSandboxImage(sandboxConfig.DefaultConfig.Version, sandboxConfig.DefaultConfig.Image)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v pulling docker image for release %s\n", emoji.Whale, image)

	if err := docker.PullDockerImage(ctx, cli, image, sandboxConfig.DefaultConfig.ImagePullPolicy); err != nil {
		return nil, err
	}

	fmt.Printf("%v booting Flyte-sandbox container\n", emoji.FactoryWorker)
	exposedPorts, portBindings, _ := docker.GetSandboxPorts()
	ID, err := docker.StartContainer(ctx, cli, volumes, exposedPorts, portBindings, docker.FlyteSandboxClusterName, image)
	if err != nil {
		fmt.Printf("%v Something went wrong: Failed to start Sandbox container %v, Please check your docker client and try again. \n", emoji.GrimacingFace, emoji.Whale)
		return nil, err
	}

	logReader, err := docker.ReadLogs(ctx, cli, ID)
	if err != nil {
		return nil, err
	}

	return logReader, nil
}

// Returns the alternate image if specified, else
// if no version is specified then the Latest release of cr.flyte.org/flyteorg/flyte-sandbox:dind is used
// else cr.flyte.org/flyteorg/flyte-sandbox:dind-{SHA}, where sha is derived from the version.
func getSandboxImage(version string, alternateImage string) (string, error) {

	if len(alternateImage) > 0 {
		return alternateImage, nil
	}

	var tag = dind
	if len(version) > 0 {
		isGreater, err := util.IsVersionGreaterThan(version, sandboxSupportedVersion)
		if err != nil {
			return "", err
		}
		if !isGreater {
			return "", fmt.Errorf("version flag only supported with flyte %s+ release", sandboxSupportedVersion)
		}
		sha, err := githubutil.GetSHAFromVersion(version, flyteRepository)
		if err != nil {
			return "", err
		}
		tag = fmt.Sprintf("%s-%s", dind, sha)
	}

	return docker.GetSandboxImage(tag), nil
}

func mountVolume(file, destination string) (*mount.Mount, error) {
	if len(file) > 0 {
		source, err := filepath.Abs(file)
		if err != nil {
			return nil, err
		}
		return &mount.Mount{
			Type:   mount.TypeBind,
			Source: source,
			Target: destination,
		}, nil
	}
	return nil, nil
}

func watchFlyteDeployment(ctx context.Context, appsClient corev1.CoreV1Interface) error {
	var data = os.Stdout
	table := tablewriter.NewWriter(data)
	table.SetHeader([]string{"Service", "Status", "Namespace"})
	table.SetRowLine(true)

	for {
		isTaint, err := isNodeTainted(ctx, appsClient)
		if err != nil {
			return err
		}
		if isTaint {
			return fmt.Errorf("docker sandbox doesn't have sufficient memory available. Please run docker system prune -a --volumes")
		}

		pods, err := getFlyteDeployment(ctx, appsClient)
		if err != nil {
			return err
		}
		table.ClearRows()
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)

		// Clear os.Stdout
		_, _ = data.WriteString("\x1b[3;J\x1b[H\x1b[2J")

		var total, ready int
		total = len(pods.Items)
		ready = 0
		if total != 0 {
			for _, v := range pods.Items {
				if isPodReady(v) {
					ready++
				}
				if len(v.Status.Conditions) > 0 {
					table.Append([]string{v.GetName(), string(v.Status.Phase), v.GetNamespace()})
				}
			}
			table.Render()
			if total == ready {
				break
			}
		}

		time.Sleep(40 * time.Second)
	}

	return nil
}

func isPodReady(v corev1api.Pod) bool {
	if (v.Status.Phase == corev1api.PodRunning) || (v.Status.Phase == corev1api.PodSucceeded) {
		return true
	}
	return false
}

func getFlyteDeployment(ctx context.Context, client corev1.CoreV1Interface) (*corev1api.PodList, error) {
	pods, err := client.Pods(flyteNamespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func isNodeTainted(ctx context.Context, client corev1.CoreV1Interface) (bool, error) {
	nodes, err := client.Nodes().List(ctx, v1.ListOptions{})
	if err != nil {
		return false, err
	}
	match := 0
	for _, node := range nodes.Items {
		for _, c := range node.Spec.Taints {
			if c.Key == diskPressureTaint && c.Effect == taintEffect {
				match++
			}
		}
	}
	if match > 0 {
		return true, nil
	}
	return false, nil
}
