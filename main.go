package main

import (
	"fmt"
	"os"

	"github.com/flyteorg/flytecopilot/cmd"
)

func main() {
	rootCmd := cmd.NewDataCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
