package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// newExpiredTaskAction builds a terminated TaskAction whose completed time is well past
// any sane maxTTL, so the GC will treat it as expired and delete it.
func newExpiredTaskAction() *flyteorgv1.TaskAction {
	expired := time.Now().UTC().Add(-2 * time.Hour).Format(labelTimeFormat)
	return &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gc-expired",
			Namespace: "default",
			Labels: map[string]string{
				LabelTerminationStatus: LabelValueTerminated,
				LabelCompletedTime:     expired,
			},
		},
	}
}

// newGCTestClient returns an in-memory fake client seeded with one expired TaskAction.
// If deleteErr is non-nil, every Delete fails with it, so we can exercise the error path.
func newGCTestClient(t *testing.T, deleteErr error) client.Client {
	scheme := runtime.NewScheme()
	require.NoError(t, flyteorgv1.AddToScheme(scheme))
	builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(newExpiredTaskAction())
	if deleteErr != nil {
		builder = builder.WithInterceptorFuncs(interceptor.Funcs{
			Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
				return deleteErr
			},
		})
	}
	return builder.Build()
}

// sweepObservations reports how many times the sweep_duration stopwatch has recorded.
// sweep_duration is a Summary, so testutil.ToFloat64 can not read it.
// Instead, we collect the metric and read its sample count directly.
func sweepObservations(t *testing.T, gc *GarbageCollector) uint64 {
	t.Helper()
	ch := make(chan prometheus.Metric, 1)
	gc.metrics.sweepTime.Observer.(prometheus.Collector).Collect(ch)
	close(ch)
	var m dto.Metric
	require.NoError(t, (<-ch).Write(&m))
	return m.GetSummary().GetSampleCount()
}

// TestGarbageCollectorDeletedMetric: a successful delete moves objects_deleted
// from 0 to 1, records one sweep, and leaves deletion_errors at 0.
func TestGarbageCollectorDeletedMetric(t *testing.T) {
	gc := NewGarbageCollector(newGCTestClient(t, nil), time.Minute, time.Hour, promutils.NewTestScope())

	require.Equal(t, 0.0, testutil.ToFloat64(gc.metrics.deleted))
	require.Equal(t, 0.0, testutil.ToFloat64(gc.metrics.errors))

	require.NoError(t, gc.collect(context.Background()))

	require.Equal(t, 1.0, testutil.ToFloat64(gc.metrics.deleted))
	require.Equal(t, 0.0, testutil.ToFloat64(gc.metrics.errors))
	require.GreaterOrEqual(t, sweepObservations(t, gc), uint64(1))
}

// TestGarbageCollectorDeletionErrorMetric: when a delete fails, deletion_errors moves
// 0 to 1 and objects_deleted stays at 0.
func TestGarbageCollectorDeletionErrorMetric(t *testing.T) {
	gc := NewGarbageCollector(
		newGCTestClient(t, errors.New("simulated delete failure")),
		time.Minute, time.Hour, promutils.NewTestScope(),
	)

	require.Equal(t, 0.0, testutil.ToFloat64(gc.metrics.errors))

	require.NoError(t, gc.collect(context.Background()))

	require.Equal(t, 1.0, testutil.ToFloat64(gc.metrics.errors))
	require.Equal(t, 0.0, testutil.ToFloat64(gc.metrics.deleted))
}
