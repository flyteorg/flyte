package main

import (
	"fmt"
	"os"

	"github.com/lyft/flytepropeller/pkg/controller"

	"github.com/lyft/flytepropeller/cmd/kubectl-flyte/cmd"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey,
		contextutils.TaskIDKey, controller.ResourceVersion)
}

func main() {

	rootCmd := cmd.NewFlyteCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
