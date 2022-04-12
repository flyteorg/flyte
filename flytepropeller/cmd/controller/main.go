package main

import (
	_ "github.com/flyteorg/flytepropeller/plugins"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flytepropeller/cmd/controller/cmd"
)

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey,
		contextutils.TaskIDKey)
}

func main() {
	cmd.Execute()
}
