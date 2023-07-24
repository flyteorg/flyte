// Package plugins facilitates all the plugins that should be loaded by FlytePropeller
package plugins

import (
	// Common place to import all plugins, so that it can be imported by Singlebinary (flytelite) or by propeller main
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/k8s"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/hive"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/dask"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/mpi"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/pytorch"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/tensorflow"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/pod"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/ray"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sagemaker"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/spark"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/webapi/athena"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/webapi/bigquery"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/webapi/databricks"
	_ "github.com/flyteorg/flyteplugins/go/tasks/plugins/webapi/snowflake"
)
