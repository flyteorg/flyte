package main

import (
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/k8s"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/hive"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/container"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/mpi"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/pytorch"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sagemaker"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sidecar"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/spark"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/webapi/athena"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/webapi/snowflake"
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
