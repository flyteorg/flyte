package sandbox

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/avast/retry-go"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/enescakir/emoji"
	"github.com/flyteorg/flytectl/clierrors"
	sandboxCmdConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	"github.com/flyteorg/flytectl/pkg/configutil"
	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/github"
	"github.com/flyteorg/flytectl/pkg/k8s"
	"github.com/flyteorg/flytectl/pkg/util"
	"github.com/kataras/tablewriter"
	corev1api "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	flyteNamespace       = "flyte"
	diskPressureTaint    = "node.kubernetes.io/disk-pressure"
	taintEffect          = "NoSchedule"
	sandboxContextName   = "flyte-sandbox"
	sandboxDockerContext = "default"
	k8sEndpoint          = "https://127.0.0.1:30086"
	sandboxImageName     = "cr.flyte.org/flyteorg/flyte-sandbox"
	demoImageName        = "cr.flyte.org/flyteorg/flyte-sandbox-lite"
)

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

func WatchFlyteDeployment(ctx context.Context, appsClient corev1.CoreV1Interface) error {
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
		} else {
			table.Append([]string{"k8s: This might take a little bit", "Bootstrapping", ""})
			table.Render()
		}

		time.Sleep(40 * time.Second)
	}

	return nil
}

func MountVolume(file, destination string) (*mount.Mount, error) {
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

func UpdateLocalKubeContext(k8sCtxMgr k8s.ContextOps, dockerCtx string, contextName string) error {
	srcConfigAccess := &clientcmd.PathOptions{
		GlobalFile:   docker.Kubeconfig,
		LoadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
	}
	return k8sCtxMgr.CopyContext(srcConfigAccess, dockerCtx, contextName)
}

func startSandbox(ctx context.Context, cli docker.Docker, g github.GHRepoService, reader io.Reader, sandboxConfig *sandboxCmdConfig.Config, defaultImageName string, defaultImagePrefix string, exposedPorts map[nat.Port]struct{}, portBindings map[nat.Port][]nat.PortBinding, consolePort int) (*bufio.Scanner, error) {
	fmt.Printf("%v Bootstrapping a brand new flyte cluster... %v %v\n", emoji.FactoryWorker, emoji.Hammer, emoji.Wrench)

	if err := docker.RemoveSandbox(ctx, cli, reader); err != nil {
		if err.Error() != clierrors.ErrSandboxExists {
			return nil, err
		}
		fmt.Printf("Existing details of your sandbox")
		util.PrintSandboxMessage(consolePort)
		return nil, nil
	}

	if err := util.SetupFlyteDir(); err != nil {
		return nil, err
	}

	templateValues := configutil.ConfigTemplateSpec{
		Host:     "localhost:30081",
		Insecure: true,
		Console:  fmt.Sprintf("http://localhost:%d", consolePort),
	}
	if err := configutil.SetupConfig(configutil.FlytectlConfig, configutil.GetTemplate(), templateValues); err != nil {
		return nil, err
	}

	volumes := docker.Volumes
	if vol, err := MountVolume(sandboxConfig.Source, docker.Source); err != nil {
		return nil, err
	} else if vol != nil {
		volumes = append(volumes, *vol)
	}
	sandboxImage := sandboxConfig.Image
	if len(sandboxImage) == 0 {
		image, version, err := github.GetFullyQualifiedImageName(defaultImagePrefix, sandboxConfig.Version, defaultImageName, sandboxConfig.Prerelease, g)
		if err != nil {
			return nil, err
		}
		sandboxImage = image
		fmt.Printf("%s Fully Qualified image\n", image)
		fmt.Printf("%v Running Flyte %s release\n", emoji.Whale, version)
	}
	fmt.Printf("%v pulling docker image for release %s\n", emoji.Whale, sandboxImage)
	if err := docker.PullDockerImage(ctx, cli, sandboxImage, sandboxConfig.ImagePullPolicy, sandboxConfig.ImagePullOptions); err != nil {
		return nil, err
	}
	sandboxEnv := sandboxConfig.Env
	if sandboxConfig.Dev {
		sandboxEnv = append(sandboxEnv, "FLYTE_DEV=True")
	}

	fmt.Printf("%v booting Flyte-sandbox container\n", emoji.FactoryWorker)
	ID, err := docker.StartContainer(ctx, cli, volumes, exposedPorts, portBindings, docker.FlyteSandboxClusterName,
		sandboxImage, sandboxEnv)

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

func primeFlytekitPod(ctx context.Context, podService corev1.PodInterface) {
	_, err := podService.Create(ctx, &corev1api.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "py39-cacher",
		},
		Spec: corev1api.PodSpec{
			RestartPolicy: corev1api.RestartPolicyNever,
			Containers: []corev1api.Container{
				{

					Name:    "flytekit",
					Image:   "ghcr.io/flyteorg/flytekit:py3.9-latest",
					Command: []string{"echo"},
					Args:    []string{"Flyte"},
				},
			},
		},
	}, v1.CreateOptions{})
	if err != nil {
		fmt.Printf("Failed to create primer pod - %s", err)
	}
}

func StartCluster(ctx context.Context, args []string, sandboxConfig *sandboxCmdConfig.Config, primePod bool, defaultImageName string, defaultImagePrefix string, exposedPorts map[nat.Port]struct{}, portBindings map[nat.Port][]nat.PortBinding, consolePort int) error {
	k8sCtxMgr := k8s.NewK8sContextManager()
	err := k8sCtxMgr.CheckConfig()
	if err != nil {
		return err
	}

	cli, err := docker.GetDockerClient()
	if err != nil {
		return err
	}

	ghRepo := github.GetGHRepoService()

	reader, err := startSandbox(ctx, cli, ghRepo, os.Stdin, sandboxConfig, defaultImageName, defaultImagePrefix, exposedPorts, portBindings, consolePort)
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
		if err = UpdateLocalKubeContext(k8sCtxMgr, sandboxDockerContext, sandboxContextName); err != nil {
			return err
		}

		if err := WatchFlyteDeployment(ctx, k8sClient.CoreV1()); err != nil {
			return err
		}
		if primePod {
			primeFlytekitPod(ctx, k8sClient.CoreV1().Pods("default"))
		}

	}
	return nil
}

func StartDemoCluster(ctx context.Context, args []string, sandboxConfig *sandboxCmdConfig.Config) error {
	primePod := true
	sandboxImagePrefix := "sha"
	exposedPorts, portBindings, err := docker.GetDemoPorts()
	if sandboxConfig.Dev {
		exposedPorts, portBindings, err = docker.GetDevPorts()
	}
	if err != nil {
		return err
	}
	err = StartCluster(ctx, args, sandboxConfig, primePod, demoImageName, sandboxImagePrefix, exposedPorts, portBindings, util.DemoConsolePort)
	if err != nil {
		return err
	}
	util.PrintSandboxMessage(util.DemoConsolePort)
	return nil
}

func StartSandboxCluster(ctx context.Context, args []string, sandboxConfig *sandboxCmdConfig.Config) error {
	primePod := false
	demoImagePrefix := "dind"
	exposedPorts, portBindings, err := docker.GetSandboxPorts()
	if err != nil {
		return err
	}
	err = StartCluster(ctx, args, sandboxConfig, primePod, sandboxImageName, demoImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
	if err != nil {
		return err
	}
	util.PrintSandboxMessage(util.SandBoxConsolePort)
	return nil
}
