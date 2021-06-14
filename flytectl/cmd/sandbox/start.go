package sandbox

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/enescakir/emoji"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
)

const (
	startShort = "Start the flyte sandbox"
	startLong  = `
Start will run the flyte sandbox cluster inside a docker container and setup the config that is required 
::

 bin/flytectl start

Usage
	`
)

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

func startSandboxCluster(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	fmt.Printf("%v It will take some time, We will start a fresh flyte cluster for you %v %v\n", emoji.ManTechnologist, emoji.Rocket, emoji.Rocket)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Please Check your docker client %v \n", emoji.ManTechnologist)
		return err
	}

	if err := setupFlytectlConfig(); err != nil {
		return err
	}

	if container := getSandbox(cli); container != nil {
		if cmdUtil.AskForConfirmation("delete existing sandbox cluster", os.Stdin) {
			if err := teardownSandboxCluster(ctx, []string{}, cmdCtx); err != nil {
				return err
			}
		}
	}

	ID, err := startContainer(cli)
	if err == nil {
		os.Setenv("KUBECONFIG", Kubeconfig)

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Something goes wrong with container status", r)
			}
		}()

		go watchError(cli, ID)
		if err := readLogs(cli, ID); err != nil {
			return err
		}

		fmt.Printf("Add (KUBECONFIG) to your environment variabl \n")
		fmt.Printf("export KUBECONFIG=%v \n", Kubeconfig)
		return nil
	}
	fmt.Println("Something goes wrong. We are not able to start sandbox container, Please check your docker client and try again \n", emoji.Rocket)
	fmt.Printf("error: %v", err)
	return nil
}
