package sandbox

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyteorg/flytectl/clierrors"

	"github.com/docker/docker/api/types/mount"

	"github.com/flyteorg/flytectl/pkg/configutil"

	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
	"github.com/flyteorg/flytectl/pkg/util"

	"github.com/flyteorg/flytectl/pkg/docker"

	"github.com/enescakir/emoji"
	sandboxConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

const (
	startShort = "Start the flyte sandbox cluster"
	startLong  = `
The Flyte Sandbox is a fully standalone minimal environment for running Flyte. provides a simplified way of running flyte-sandbox as a single Docker container running locally.  

Start sandbox cluster without any source code
::

 bin/flytectl sandbox start
	
Mount your source code repository inside sandbox 
::

 bin/flytectl sandbox start --source=$HOME/flyteorg/flytesnacks 
	
Run specific version of flyte, Only available after v0.14.0+
::

 bin/flytectl sandbox start  --version=v0.14.0

Usage
	`
	GeneratedManifest            = "/flyteorg/share/flyte_generated.yaml"
	FlyteReleaseURL              = "/flyteorg/flyte/releases/download/%v/flyte_sandbox_manifest.yaml"
	FlyteMinimumVersionSupported = "v0.14.0"
	GithubURL                    = "https://github.com"
)

var (
	FlyteManifest = f.FilePathJoin(f.UserHomeDir(), ".flyte", "flyte_generated.yaml")
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
	return nil
}

func startSandbox(ctx context.Context, cli docker.Docker, reader io.Reader) (*bufio.Scanner, error) {
	fmt.Printf("%v Bootstrapping a brand new flyte cluster... %v %v\n", emoji.FactoryWorker, emoji.Hammer, emoji.Wrench)

	if err := docker.RemoveSandbox(ctx, cli, reader); err != nil {
		if err.Error() != clierrors.ErrSandboxExists {
			return nil, err
		}
		printExistingSandboxMessage()
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

	if len(sandboxConfig.DefaultConfig.Version) > 0 {
		if err := downloadFlyteManifest(sandboxConfig.DefaultConfig.Version); err != nil {
			return nil, err
		}
		vol, err := mountVolume(FlyteManifest, GeneratedManifest)
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, *vol)
	}

	fmt.Printf("%v pulling docker image %s\n", emoji.Whale, docker.ImageName)
	if err := docker.PullDockerImage(ctx, cli, docker.ImageName); err != nil {
		return nil, err
	}

	fmt.Printf("%v booting Flyte-sandbox container\n", emoji.FactoryWorker)
	exposedPorts, portBindings, _ := docker.GetSandboxPorts()
	ID, err := docker.StartContainer(ctx, cli, volumes, exposedPorts, portBindings, docker.FlyteSandboxClusterName, docker.ImageName)
	if err != nil {
		fmt.Printf("%v Something went wrong: Failed to start Sandbox container %v, Please check your docker client and try again. \n", emoji.GrimacingFace, emoji.Whale)
		return nil, err
	}

	_, errCh := docker.WatchError(ctx, cli, ID)
	logReader, err := docker.ReadLogs(ctx, cli, ID)
	if err != nil {
		return nil, err
	}
	go func() {
		err := <-errCh
		if err != nil {
			fmt.Printf("err: %v", err)
			os.Exit(1)
		}
	}()

	return logReader, nil
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

func downloadFlyteManifest(version string) error {
	isGreater, err := util.IsVersionGreaterThan(version, FlyteMinimumVersionSupported)
	if err != nil {
		return err
	}
	if !isGreater {
		return fmt.Errorf("version flag only support %s+ flyte version", FlyteMinimumVersionSupported)
	}
	response, err := util.GetRequest(GithubURL, fmt.Sprintf(FlyteReleaseURL, version))
	if err != nil {
		return err
	}
	if err := util.WriteIntoFile(response, FlyteManifest); err != nil {
		return err
	}
	return nil
}

func printExistingSandboxMessage() {
	kubeconfig := strings.Join([]string{
		"$KUBECONFIG",
		f.FilePathJoin(f.UserHomeDir(), ".kube", "config"),
		docker.Kubeconfig,
	}, ":")

	fmt.Printf("Existing details of your sandbox:")
	fmt.Printf("%v %v %v %v %v \n", emoji.ManTechnologist, docker.SuccessMessage, emoji.Rocket, emoji.Rocket, emoji.PartyPopper)
	fmt.Printf("Add KUBECONFIG and FLYTECTL_CONFIG to your environment variable \n")
	fmt.Printf("export KUBECONFIG=%v \n", kubeconfig)
	fmt.Printf("export FLYTECTL_CONFIG=%v \n", configutil.FlytectlConfig)
}
