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
	"github.com/flyteorg/flytestdlib/logger"
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
	K8sEndpoint          = "https://127.0.0.1:6443"
	sandboxK8sEndpoint   = "https://127.0.0.1:30086"
	sandboxImageName     = "cr.flyte.org/flyteorg/flyte-sandbox"
	demoImageName        = "cr.flyte.org/flyteorg/flyte-sandbox-bundled"
	DefaultFlyteConfig   = "/opt/flyte/defaults.flyte.yaml"
	k3sKubeConfigEnvVar  = "K3S_KUBECONFIG_OUTPUT=/var/lib/flyte/config/kubeconfig"
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

func UpdateLocalKubeContext(k8sCtxMgr k8s.ContextOps, dockerCtx string, contextName string, kubeConfigPath string) error {
	srcConfigAccess := &clientcmd.PathOptions{
		GlobalFile:   kubeConfigPath,
		LoadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
	}
	return k8sCtxMgr.CopyContext(srcConfigAccess, dockerCtx, contextName)
}

func startSandbox(ctx context.Context, cli docker.Docker, g github.GHRepoService, reader io.Reader, sandboxConfig *sandboxCmdConfig.Config, defaultImageName string, defaultImagePrefix string, exposedPorts map[nat.Port]struct{}, portBindings map[nat.Port][]nat.PortBinding, consolePort int) (*bufio.Scanner, error) {
	fmt.Printf("%v Bootstrapping a brand new flyte cluster... %v %v\n", emoji.FactoryWorker, emoji.Hammer, emoji.Wrench)
	if sandboxConfig.DryRun {
		docker.PrintRemoveContainer(docker.FlyteSandboxClusterName)
	} else {
		if err := docker.RemoveSandbox(ctx, cli, reader); err != nil {
			if err.Error() != clierrors.ErrSandboxExists {
				return nil, err
			}
			fmt.Printf("Existing details of your sandbox")
			util.PrintSandboxStartMessage(consolePort, docker.Kubeconfig, sandboxConfig.DryRun)
			return nil, nil
		}
	}

	templateValues := configutil.ConfigTemplateSpec{
		Host:     "localhost:30080",
		Insecure: true,
		Console:  fmt.Sprintf("http://localhost:%d", consolePort),
	}
	if err := configutil.SetupConfig(configutil.FlytectlConfig, configutil.GetTemplate(), templateValues); err != nil {
		return nil, err
	}

	volumes := docker.Volumes
	// Mount this even though it should no longer be necessary. This is for user code
	if vol, err := MountVolume(sandboxConfig.DeprecatedSource, docker.Source); err != nil {
		return nil, err
	} else if vol != nil {
		volumes = append(volumes, *vol)
	}

	// This is the sandbox configuration directory mount, flyte will write the kubeconfig here. May hold more in future releases
	// To be interoperable with the old sandbox, only mount if the directory exists, should've created by StartCluster
	if fileInfo, err := os.Stat(docker.FlyteSandboxConfigDir); err == nil {
		if fileInfo.IsDir() {
			if vol, err := MountVolume(docker.FlyteSandboxConfigDir, docker.FlyteSandboxInternalConfigDir); err != nil {
				return nil, err
			} else if vol != nil {
				volumes = append(volumes, *vol)
			}
		}
	}

	// Create and mount a docker volume that will be used to persist data
	// across sandbox clusters
	if _, err := docker.GetOrCreateVolume(
		ctx,
		cli,
		docker.FlyteSandboxVolumeName,
		sandboxConfig.DryRun,
	); err != nil {
		return nil, err
	}
	volumes = append(volumes, mount.Mount{
		Type:   mount.TypeVolume,
		Source: docker.FlyteSandboxVolumeName,
		Target: docker.FlyteSandboxInternalStorageDir,
	})

	sandboxImage := sandboxConfig.Image
	if len(sandboxImage) == 0 {
		image, version, err := github.GetFullyQualifiedImageName(defaultImagePrefix, sandboxConfig.Version, defaultImageName, sandboxConfig.Prerelease, g)
		if err != nil {
			return nil, err
		}
		sandboxImage = image
		fmt.Printf("%v Going to use Flyte %s release with image %s \n", emoji.Whale, version, image)
	}
	if err := docker.PullDockerImage(ctx, cli, sandboxImage, sandboxConfig.ImagePullPolicy, sandboxConfig.ImagePullOptions, sandboxConfig.DryRun); err != nil {
		return nil, err
	}
	sandboxEnv := sandboxConfig.Env
	if sandboxConfig.Dev {
		sandboxEnv = append(sandboxEnv, "FLYTE_DEV=True")
	}

	if sandboxConfig.DisableAgent {
		sandboxEnv = append(sandboxEnv, "DISABLE_AGENT=True")
	}

	ID, err := docker.StartContainer(ctx, cli, volumes, exposedPorts, portBindings, docker.FlyteSandboxClusterName,
		sandboxImage, sandboxEnv, sandboxConfig.DryRun)

	if err != nil {
		fmt.Printf("%v Something went wrong: Failed to start Sandbox container %v, Please check your docker client and try again. \n", emoji.GrimacingFace, emoji.Whale)
		return nil, err
	}

	var logReader *bufio.Scanner
	if !sandboxConfig.DryRun {
		logReader, err = docker.ReadLogs(ctx, cli, ID)
		if err != nil {
			return nil, err
		}
	}

	return logReader, nil
}

func StartCluster(ctx context.Context, args []string, sandboxConfig *sandboxCmdConfig.Config, defaultImageName string, defaultImagePrefix string, exposedPorts map[nat.Port]struct{}, portBindings map[nat.Port][]nat.PortBinding, consolePort int) error {
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
	if err := util.CreatePathAndFile(docker.Kubeconfig); err != nil {
		return err
	}

	reader, err := startSandbox(ctx, cli, ghRepo, os.Stdin, sandboxConfig, defaultImageName, defaultImagePrefix, exposedPorts, portBindings, consolePort)
	if err != nil {
		return err
	}

	if reader != nil {
		var k8sClient k8s.K8s
		err = retry.Do(
			func() error {
				// This should wait for the kubeconfig file being there.
				k8sClient, err = k8s.GetK8sClient(docker.Kubeconfig, K8sEndpoint)
				return err
			},
			retry.Attempts(10),
		)
		if err != nil {
			return err
		}

		// Live-ness check
		err = retry.Do(
			func() error {
				// Have to get a new client every time because you run into x509 errors if not
				fmt.Println("Waiting for cluster to come up...")
				k8sClient, err = k8s.GetK8sClient(docker.Kubeconfig, K8sEndpoint)
				if err != nil {
					logger.Debugf(ctx, "Error getting K8s client in liveness check %s", err)
					return err
				}
				req := k8sClient.CoreV1().RESTClient().Get()
				req = req.RequestURI("livez")
				res := req.Do(ctx)
				return res.Error()
			},
			retry.Attempts(15),
		)
		if err != nil {
			return err
		}

		// Readiness check
		err = retry.Do(
			func() error {
				// No need to refresh client here
				req := k8sClient.CoreV1().RESTClient().Get()
				req = req.RequestURI("readyz")
				res := req.Do(ctx)
				return res.Error()
			},
			retry.Attempts(10),
		)
		if err != nil {
			return err
		}

		// This will copy the kubeconfig from where k3s writes it () to the main file.
		// This code is located after the waits above since it appears that k3s goes through at least a couple versions
		// of the config keys/certs. If this copy is done too early, the copied credentials won't work.
		if err = UpdateLocalKubeContext(k8sCtxMgr, sandboxDockerContext, sandboxContextName, docker.Kubeconfig); err != nil {
			return err
		}

		// Watch for Flyte Deployment
		if err := WatchFlyteDeployment(ctx, k8sClient.CoreV1()); err != nil {
			return err
		}
	}
	return nil
}

// StartClusterForSandbox is the code for the original multi deploy version of sandbox, should be removed once we
// document the new development experience for plugins.
func StartClusterForSandbox(ctx context.Context, args []string, sandboxConfig *sandboxCmdConfig.Config, defaultImageName string, defaultImagePrefix string, exposedPorts map[nat.Port]struct{}, portBindings map[nat.Port][]nat.PortBinding, consolePort int) error {
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

	if err := util.CreatePathAndFile(docker.SandboxKubeconfig); err != nil {
		return err
	}

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
				k8sClient, err = k8s.GetK8sClient(docker.SandboxKubeconfig, sandboxK8sEndpoint)
				return err
			},
			retry.Attempts(10),
		)
		if err != nil {
			return err
		}
		if err = UpdateLocalKubeContext(k8sCtxMgr, sandboxDockerContext, sandboxContextName, docker.SandboxKubeconfig); err != nil {
			return err
		}

		// TODO: This doesn't appear to correctly watch for the Flyte deployment but doesn't do so on master either.
		if err := WatchFlyteDeployment(ctx, k8sClient.CoreV1()); err != nil {
			return err
		}
	}
	return nil
}

func StartDemoCluster(ctx context.Context, args []string, sandboxConfig *sandboxCmdConfig.Config) error {
	sandboxImagePrefix := "sha"
	exposedPorts, portBindings, err := docker.GetDemoPorts()
	if err != nil {
		return err
	}
	// K3s will automatically write the file specified by this var, which is mounted from user's local state dir.

	sandboxConfig.Env = append(sandboxConfig.Env, k3sKubeConfigEnvVar)
	err = StartCluster(ctx, args, sandboxConfig, demoImageName, sandboxImagePrefix, exposedPorts, portBindings, util.DemoConsolePort)
	if err != nil {
		return err
	}
	util.PrintDemoStartMessage(util.DemoConsolePort, docker.Kubeconfig, sandboxConfig.DryRun)
	return nil
}

func StartSandboxCluster(ctx context.Context, args []string, sandboxConfig *sandboxCmdConfig.Config) error {
	demoImagePrefix := "dind"
	exposedPorts, portBindings, err := docker.GetSandboxPorts()
	if err != nil {
		return err
	}
	err = StartClusterForSandbox(ctx, args, sandboxConfig, sandboxImageName, demoImagePrefix, exposedPorts, portBindings, util.SandBoxConsolePort)
	if err != nil {
		return err
	}
	util.PrintSandboxStartMessage(util.SandBoxConsolePort, docker.SandboxKubeconfig, sandboxConfig.DryRun)
	return nil
}
