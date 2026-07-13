package impl

import (
	"context"
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestNewDBMetrics_NilScopeIsNoOp(t *testing.T) {
	// A nil scope must not panic and must produce a usable (no-op) metrics value
	// so existing callers/tests that don't wire a scope keep working.
	m := newDBMetrics(nil)
	assert.NotPanics(t, func() {
		err := m.observe(context.Background(), "get_task", func() error { return nil })
		assert.NoError(t, err)
	})
}

func TestDBMetrics_ObserveRecordsSuccess(t *testing.T) {
	scope := promutils.NewTestScope()
	m := newDBMetrics(scope)

	err := m.observe(context.Background(), "get_task", func() error { return nil })
	assert.NoError(t, err)

	// One successful call: call counter incremented, error counter untouched, latency observed.
	assert.Equal(t, float64(1), testutil.ToFloat64(m.calls.WithLabelValues("get_task")))
	assert.Equal(t, float64(0), testutil.ToFloat64(m.errors.WithLabelValues("get_task")))
	assert.Equal(t, 1, testutil.CollectAndCount(m.latency.SummaryVec))
}

func TestDBMetrics_ObserveRecordsError(t *testing.T) {
	scope := promutils.NewTestScope()
	m := newDBMetrics(scope)

	sentinel := errors.New("boom")
	err := m.observe(context.Background(), "create_task", func() error { return sentinel })
	assert.ErrorIs(t, err, sentinel)

	// One failed call: both call and error counters incremented.
	assert.Equal(t, float64(1), testutil.ToFloat64(m.calls.WithLabelValues("create_task")))
	assert.Equal(t, float64(1), testutil.ToFloat64(m.errors.WithLabelValues("create_task")))
}

func TestDBMetrics_LabeledByOperationOnly(t *testing.T) {
	scope := promutils.NewTestScope()
	m := newDBMetrics(scope)

	for i := 0; i < 3; i++ {
		assert.NoError(t, m.observe(context.Background(), "list_tasks", func() error { return nil }))
	}
	assert.NoError(t, m.observe(context.Background(), "get_task", func() error { return nil }))

	assert.Equal(t, float64(3), testutil.ToFloat64(m.calls.WithLabelValues("list_tasks")))
	assert.Equal(t, float64(1), testutil.ToFloat64(m.calls.WithLabelValues("get_task")))
}

// TestProjectRepo_InstrumentsCreateProject is an end-to-end check that a real
// repository operation increments its call counter and records latency through
// the wired-in metrics, against the embedded test database.
func TestProjectRepo_InstrumentsCreateProject(t *testing.T) {
	db := setupDB(t)
	t.Cleanup(func() { db.Exec("DELETE FROM projects") })

	scope := promutils.NewTestScope()
	repo := NewProjectRepo(db, scope)
	ctx := context.Background()
	state := int32(0)

	require.NoError(t, repo.CreateProject(ctx, &models.Project{
		Identifier: "metrics-proj",
		Name:       "metrics-proj",
		State:      &state,
	}))

	// Reach into the concrete type to read the instruments registered by the
	// constructor from the same scope.
	m := repo.(*projectRepo).metrics
	assert.Equal(t, float64(1), testutil.ToFloat64(m.calls.WithLabelValues("create_project")),
		"create_project call counter should be incremented")
	assert.Equal(t, float64(0), testutil.ToFloat64(m.errors.WithLabelValues("create_project")),
		"create_project error counter should be zero on success")
	assert.Equal(t, 1, testutil.CollectAndCount(m.latency.SummaryVec),
		"create_project latency should be recorded")

	// A duplicate create returns ErrProjectAlreadyExists and must bump the error counter.
	err := repo.CreateProject(ctx, &models.Project{
		Identifier: "metrics-proj",
		Name:       "metrics-proj",
		State:      &state,
	})
	require.Error(t, err)
	assert.Equal(t, float64(2), testutil.ToFloat64(m.calls.WithLabelValues("create_project")))
	assert.Equal(t, float64(1), testutil.ToFloat64(m.errors.WithLabelValues("create_project")))
}
