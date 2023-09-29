package main

import (
	_ "github.com/flyteorg/flyte/flytepropeller/plugins"

	"github.com/flyteorg/flyte/flytepropeller/cmd/controller/cmd"
)

func main() {
	cmd.Execute()
}
