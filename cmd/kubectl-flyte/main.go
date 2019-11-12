package main

import (
	"fmt"
	"os"

	"github.com/lyft/flytepropeller/cmd/kubectl-flyte/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {

	rootCmd := cmd.NewFlyteCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
