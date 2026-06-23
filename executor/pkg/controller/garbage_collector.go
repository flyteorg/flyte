package controller

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// gcMetrics holds the Prometheus instruments for the garbage collector. They are
// created once (in newGCMetrics) and only updated thereafter, never re-registered.
type gcMetrics struct {
	deleted   prometheus.Counter
	errors    prometheus.Counter
	sweepTime promutils.StopWatch
}

// newGCMetrics builds the garbage collector instruments under the given scope.
// Call exactly once per scope to avoid duplicate registration panics.
func newGCMetrics(scope promutils.Scope) gcMetrics {
	return gcMetrics{
		deleted:   scope.MustNewCounter("objects_deleted", "Total TaskActions deleted by the garbage collector"),
		errors:    scope.MustNewCounter("deletion_errors", "Total errors encountered while deleting expired TaskActions"),
		sweepTime: scope.MustNewStopWatch("sweep_duration", "Duration of a full garbage collection sweep", time.Millisecond),
	}
}

// GarbageCollector periodically deletes terminal TaskActions that have exceeded their TTL.
// It implements the controller-runtime manager.Runnable interface.
type GarbageCollector struct {
	client   client.Client
	interval time.Duration
	maxTTL   time.Duration
	metrics  gcMetrics
}

// NewGarbageCollector creates a new GarbageCollector.
func NewGarbageCollector(c client.Client, interval, maxTTL time.Duration, scope promutils.Scope) *GarbageCollector {
	return &GarbageCollector{
		client:   c,
		interval: interval,
		maxTTL:   maxTTL,
		metrics:  newGCMetrics(scope),
	}
}

// Start runs the garbage collection loop until the context is cancelled.
// It satisfies the manager.Runnable interface.
func (gc *GarbageCollector) Start(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("gc")
	logger.Info("starting TaskAction garbage collector", "interval", gc.interval, "maxTTL", gc.maxTTL)

	ticker := time.NewTicker(gc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("stopping TaskAction garbage collector")
			return nil
		case <-ticker.C:
			if err := gc.collect(ctx); err != nil {
				logger.Error(err, "garbage collection cycle failed")
			}
		}
	}
}

const gcPageSize = 500

// collect lists all terminated TaskActions (paginated) and deletes those whose completed-time has expired.
func (gc *GarbageCollector) collect(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("gc")

	// Time the full sweep, defer guarantees it records on every return path.
	timer := gc.metrics.sweepTime.Start()
	defer timer.Stop()

	cutoff := time.Now().UTC().Add(-gc.maxTTL).Format(labelTimeFormat)
	deleted := 0
	total := 0
	continueToken := ""

	for {
		var taskActions flyteorgv1.TaskActionList
		listOpts := []client.ListOption{
			client.MatchingLabels{LabelTerminationStatus: LabelValueTerminated},
			client.HasLabels{LabelCompletedTime},
			client.Limit(gcPageSize),
		}
		if continueToken != "" {
			listOpts = append(listOpts, client.Continue(continueToken))
		}

		if err := gc.client.List(ctx, &taskActions, listOpts...); err != nil {
			return err
		}

		total += len(taskActions.Items)

		for i := range taskActions.Items {
			ta := &taskActions.Items[i]
			completedTime := ta.GetLabels()[LabelCompletedTime]
			if completedTime == "" {
				continue
			}

			// The minute-precision format is lexicographically ordered, so string comparison works.
			if completedTime < cutoff {
				if err := gc.client.Delete(ctx, ta); err != nil {
					gc.metrics.errors.Inc()
					logger.Error(err, "failed to delete expired TaskAction",
						"name", ta.Name, "namespace", ta.Namespace, "completedTime", completedTime)
					continue
				}
				deleted++
				gc.metrics.deleted.Inc()
			}
		}

		continueToken = taskActions.GetContinue()
		if continueToken == "" {
			break
		}
	}

	if deleted > 0 {
		logger.Info("garbage collection completed", "deleted", deleted, "total", total)
	}

	return nil
}
