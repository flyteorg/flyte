package main

import (
	"github.com/flyteorg/flyte/flytepropeller/cmd/controller/cmd"
	_ "github.com/flyteorg/flyte/flytepropeller/plugins"
)

func main() {
	cmd.Execute()
}
