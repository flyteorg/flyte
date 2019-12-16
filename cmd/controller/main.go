package main

import (
	_ "github.com/lyft/flyteplugins/go/tasks/plugins/array/awsbatch"
	_ "github.com/lyft/flyteplugins/go/tasks/plugins/array/k8s"
	_ "github.com/lyft/flyteplugins/go/tasks/plugins/hive"
	_ "github.com/lyft/flyteplugins/go/tasks/plugins/k8s/container"
	_ "github.com/lyft/flyteplugins/go/tasks/plugins/k8s/sidecar"
	_ "github.com/lyft/flyteplugins/go/tasks/plugins/k8s/spark"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/lyft/flytepropeller/cmd/controller/cmd"
	"github.com/lyft/flytepropeller/pkg/controller"
)

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey,
		contextutils.TaskIDKey, controller.ResourceVersion)
}

func main() {
	cmd.Execute()
}
