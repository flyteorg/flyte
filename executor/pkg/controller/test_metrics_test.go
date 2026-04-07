package controller

import (
	"sync"

	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
)

var metricsKeysOnce sync.Once

func ensureTestMetricKeys() {
	metricsKeysOnce.Do(func() {
		labeled.SetMetricKeys(
			contextutils.ProjectKey,
			contextutils.DomainKey,
			contextutils.WorkflowIDKey,
			contextutils.TaskIDKey,
		)
	})
}
