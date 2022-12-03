package storage

import (
	"github.com/flyteorg/flyte/stdlib/contextutils"
	"github.com/flyteorg/flyte/stdlib/promutils"
	"github.com/flyteorg/flyte/stdlib/promutils/labeled"
)

var metrics *dataStoreMetrics

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
	scope := promutils.NewTestScope()
	metrics = newDataStoreMetrics(scope)
}
