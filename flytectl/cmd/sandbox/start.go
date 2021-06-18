package sandbox

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/enescakir/emoji"
	sandboxConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/sandbox"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
)

const (
	startShort = "Start the flyte sandbox"
	startLong  = `
Start will run the flyte sandbox cluster inside a docker container and setup the config that is required 
::

 bin/flytectl sandbox start
	
Mount your flytesnacks repository code inside sandbox 
::

 bin/flytectl sandbox start --flytesnacks=$HOME/flyteorg/flytesnacks 
Usage
	`
)

var volumes = []mount.Mount{
	{
		Type:   mount.TypeBind,
		Source: f.FilePathJoin(f.UserHomeDir(), ".flyte"),
		Target: K3sDir,
	},
}

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

func startSandboxCluster(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	fmt.Printf("%v Bootstrapping a brand new flyte cluster... %v %v\n", emoji.FactoryWorker, emoji.Hammer, emoji.Wrench)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("%v Please Check your docker client %v \n", emoji.GrimacingFace, emoji.Whale)
		return err
	}

	if err := setupFlytectlConfig(); err != nil {
		return err
	}

	if err := removeSandboxIfExist(cli, os.Stdin); err != nil {
		return err
	}

	if len(sandboxConfig.DefaultConfig.SnacksRepo) > 0 {
		volumes = append(volumes, mount.Mount{
			Type:   mount.TypeBind,
			Source: sandboxConfig.DefaultConfig.SnacksRepo,
			Target: flyteSnackDir,
		})
	}

	os.Setenv("KUBECONFIG", Kubeconfig)
	os.Setenv("FLYTECTL_CONFIG", FlytectlConfig)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%v Something went horribly wrong! %s\n", emoji.GrimacingFace, r)
		}
	}()

	ID, err := startContainer(cli, volumes)
	if err != nil {
		fmt.Printf("%v Something went horribly wrong: Failed to start Sandbox container %v, Please check your docker client and try again. \n", emoji.GrimacingFace, emoji.Whale)
		return fmt.Errorf("error: %v", err)
	}

	_ = readLogs(cli, ID, SuccessMessage)
	fmt.Printf("Add KUBECONFIG and FLYTECTL_CONFIG to your environment variable \n")
	fmt.Printf("export KUBECONFIG=%v \n", Kubeconfig)
	fmt.Printf("export FLYTECTL_CONFIG=%v \n", FlytectlConfig)
	return nil
}
