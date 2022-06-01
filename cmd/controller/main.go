package main

import (
	_ "github.com/flyteorg/flytepropeller/plugins"

	"github.com/flyteorg/flytepropeller/cmd/controller/cmd"
)

func main() {
	cmd.Execute()
}
