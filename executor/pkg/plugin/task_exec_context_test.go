package plugin

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

func TestMain(m *testing.M) {
	labeled.SetMetricKeys(
		contextutils.ProjectKey,
		contextutils.DomainKey,
		contextutils.WorkflowIDKey,
		contextutils.TaskIDKey,
	)
	os.Exit(m.Run())
}

func newTestDataStore(t *testing.T) *storage.DataStore {
	t.Helper()
	ds, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	require.NoError(t, err)
	return ds
}

func minimalTaskAction(namespace, name, runOutputBase, actionName string, attempts uint32) *flyteorgv1.TaskAction {
	return &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			RunOutputBase: runOutputBase,
			ActionName:    actionName,
			InputURI:      runOutputBase + "/inputs.pb",
		},
		Status: flyteorgv1.TaskActionStatus{
			Attempts: attempts,
		},
	}
}

// TestNewTaskExecutionContext_FirstAttempt_NoPreviousCheckpoint verifies that
// when Attempts is unset (0), no previous checkpoint path is set — there is
// nothing to resume from on the very first execution.
func TestNewTaskExecutionContext_FirstAttempt_NoPreviousCheckpoint(t *testing.T) {
	ds := newTestDataStore(t)
	ta := minimalTaskAction("flyte", "my-action-abc", "s3://bucket/run", "action-0", 0)

	tCtx, err := NewTaskExecutionContext(ta, ds, nil, nil, nil, nil)
	require.NoError(t, err)

	prev := string(tCtx.OutputWriter().GetPreviousCheckpointsPrefix())
	assert.Empty(t, prev, "first attempt must have no previous checkpoint path")
}

// TestNewTaskExecutionContext_Retry_PreviousCheckpointSet verifies that on a
// retry (Attempts == 2), the previous checkpoint path points to the attempt-1
// output directory and ends with the checkpoint suffix.
func TestNewTaskExecutionContext_Retry_PreviousCheckpointSet(t *testing.T) {
	ds := newTestDataStore(t)
	ta := minimalTaskAction("flyte", "my-action-abc", "s3://bucket/run", "action-0", 2)

	tCtx, err := NewTaskExecutionContext(ta, ds, nil, nil, nil, nil)
	require.NoError(t, err)

	prev := string(tCtx.OutputWriter().GetPreviousCheckpointsPrefix())
	require.NotEmpty(t, prev, "retry must have a previous checkpoint path")
	assert.True(t, strings.HasSuffix(prev, "/_flytecheckpoints"),
		"previous checkpoint path must end with /_flytecheckpoints; got %q", prev)
	assert.True(t, strings.HasPrefix(prev, "s3://bucket/"),
		"previous checkpoint path must be under the RunOutputBase bucket; got %q", prev)
}

// TestNewTaskExecutionContext_Retry_PreviousAndCurrentDontCollide verifies that
// the previous and current checkpoint paths are distinct and both live under
// the same RunOutputBase — retries must not overwrite prior checkpoints.
func TestNewTaskExecutionContext_Retry_PreviousAndCurrentDontCollide(t *testing.T) {
	ds := newTestDataStore(t)
	ta := minimalTaskAction("flyte", "my-action-abc", "s3://bucket/run", "action-0", 2)

	tCtx, err := NewTaskExecutionContext(ta, ds, nil, nil, nil, nil)
	require.NoError(t, err)

	ow := tCtx.OutputWriter()
	prev := string(ow.GetPreviousCheckpointsPrefix())
	curr := string(ow.GetCheckpointPrefix())

	assert.NotEqual(t, prev, curr, "previous and current checkpoint paths must not collide")
	assert.True(t, strings.HasPrefix(prev, "s3://bucket/"),
		"previous checkpoint must live under RunOutputBase bucket; got %q", prev)
	assert.True(t, strings.HasPrefix(curr, "s3://bucket/"),
		"current checkpoint must live under RunOutputBase bucket; got %q", curr)
}
