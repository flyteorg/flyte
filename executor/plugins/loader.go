package plugins

import (
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/core/sleep"
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/dask"
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/kfoperators/mpi"
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/kfoperators/pytorch"
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/kfoperators/tensorflow"
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/ray"
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/spark"
)
