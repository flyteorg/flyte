package controller

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

// GarbageCollector periodically deletes terminal TaskActions that have exceeded their TTL.
// It implements the controller-runtime manager.Runnable interface.
type GarbageCollector struct {
	client   client.Client
	reader   client.Reader
	interval time.Duration
	maxTTL   time.Duration
	metrics  *gcMetrics
}

// NewGarbageCollector creates a new GarbageCollector. reader must be an
// uncached reader (e.g. mgr.GetAPIReader()): collect() lists with Continue
// pagination, which the controller-runtime cache does not support
// ("continue list option is not supported by the cache"). client is used for
// the deletes. provider is the executor's OTel meter provider
// (otelutils.GetMeterProvider(otelServiceName) in executor/setup.go).
func NewGarbageCollector(c client.Client, reader client.Reader, interval, maxTTL time.Duration, provider metric.MeterProvider) *GarbageCollector {
	metrics, err := newGCMetrics(provider)
	if err != nil {
		// Non-fatal: run unmetered rather than fail startup.
		log.Log.Error(err, "failed to register garbage collector metrics")
	}
	return &GarbageCollector{
		client:   c,
		reader:   reader,
		interval: interval,
		maxTTL:   maxTTL,
		metrics:  metrics,
	}
}

// Start runs the garbage collection loop until the context is cancelled.
// It satisfies the manager.Runnable interface.
func (gc *GarbageCollector) Start(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("gc")
	logger.Info("starting TaskAction garbage collector", "interval", gc.interval, "maxTTL", gc.maxTTL)

	ticker := time.NewTicker(gc.interval)
	defer ticker.Stop()

	// Sweep once immediately on startup. time.Ticker only fires its first tick
	// after a full interval, so without this a restart defers the first
	// collection by the whole interval.
	if err := gc.collect(ctx); err != nil {
		logger.Error(err, "initial garbage collection cycle failed")
	}

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

// gcPageSize bounds each List page. It's a var (not const) so tests can lower
// it to exercise the Continue pagination path without creating 500 objects.
var gcPageSize = 500

func shouldDeleteTerminalTaskAction(completedTime string, maxTTL time.Duration, now time.Time) bool {
	if completedTime == "" {
		return false
	}
	if maxTTL <= 0 {
		return true
	}

	// The minute-precision format is lexicographically ordered, so string comparison works.
	cutoff := now.UTC().Add(-maxTTL).Format(labelTimeFormat)
	return completedTime < cutoff
}

// collect lists all terminated TaskActions (paginated) and deletes those whose completed-time has expired.
// A non-positive maxTTL means terminal TaskActions are deleted on the next GC cycle.
func (gc *GarbageCollector) collect(ctx context.Context) (err error) {
	start := time.Now()
	defer func() { gc.metrics.recordSweep(ctx, start, err) }()
	logger := log.FromContext(ctx).WithName("gc")

	now := time.Now().UTC()
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

		if err := gc.reader.List(ctx, &taskActions, listOpts...); err != nil {
			return err
		}

		total += len(taskActions.Items)

		for i := range taskActions.Items {
			ta := &taskActions.Items[i]
			completedTime := ta.GetLabels()[LabelCompletedTime]
			if completedTime == "" {
				continue
			}

			if shouldDeleteTerminalTaskAction(completedTime, gc.maxTTL, now) {
				if err := gc.client.Delete(ctx, ta); err != nil {
					// Already gone is the desired state: a child TaskAction is
					// often cascade-deleted (via OwnerReferences) when its parent
					// is deleted earlier in this same pass, so the explicit delete
					// races with the cascade and returns NotFound. Not an error.
					if apierrors.IsNotFound(err) {
						gc.metrics.recordDeletion(ctx, gcOutcomeAlreadyGone)
						continue
					}
					logger.Error(err, "failed to delete expired TaskAction",
						"name", ta.Name, "namespace", ta.Namespace, "completedTime", completedTime)
					gc.metrics.recordDeletion(ctx, gcOutcomeFailed)
					continue
				}
				deleted++
				gc.metrics.recordDeletion(ctx, gcOutcomeDeleted)
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
