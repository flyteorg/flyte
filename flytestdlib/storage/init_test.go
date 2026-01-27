package storage

import (
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
)

var metrics *dataStoreMetrics

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
	scope := promutils.NewTestScope()
	metrics = newDataStoreMetrics(scope)
}
