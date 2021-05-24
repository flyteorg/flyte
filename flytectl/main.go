package main

import (
	"context"
	"os"

	"github.com/flyteorg/flytectl/cmd"
	"github.com/flyteorg/flytestdlib/logger"
)

func main() {
	if err := cmd.ExecuteCmd(); err != nil {
		logger.Error(context.TODO(), err)
		os.Exit(1)
	}
}
