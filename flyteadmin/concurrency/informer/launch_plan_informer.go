package informer

import (
	"context"
	"sync"
	"time"

	"github.com/flyteorg/flyte/flyteadmin/concurrency/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
)

// LaunchPlanInformer caches and periodically refreshes launch plan concurrency information
type LaunchPlanInformer struct {
	repo            interfaces.ConcurrencyRepoInterface
	refreshInterval time.Duration
	launchPlans     map[string]*admin.SchedulerPolicy
	mutex           sync.RWMutex
	scope           promutils.Scope
	metrics         *launchPlanInformerMetrics
}

type launchPlanInformerMetrics struct {
	refreshSuccessCount prometheus.Counter
	refreshFailureCount prometheus.Counter
	refreshDuration     promutils.StopWatch
	cachedLaunchPlans   prometheus.Gauge
}

// NewLaunchPlanInformer creates a new LaunchPlanInformer
func NewLaunchPlanInformer(
	repo interfaces.ConcurrencyRepoInterface,
	refreshInterval time.Duration,
	scope promutils.Scope,
) *LaunchPlanInformer {
	informerScope := scope.NewSubScope("launch_plan_informer")
	return &LaunchPlanInformer{
		repo:            repo,
		refreshInterval: refreshInterval,
		launchPlans:     make(map[string]*admin.SchedulerPolicy),
		scope:           informerScope,
		metrics: &launchPlanInformerMetrics{
			refreshSuccessCount: informerScope.MustNewCounter("refresh_success_count", "Count of successful refreshes"),
			refreshFailureCount: informerScope.MustNewCounter("refresh_failure_count", "Count of failed refreshes"),
			refreshDuration:     informerScope.MustNewStopWatch("refresh_duration", "Duration of refresh operations", time.Millisecond),
			cachedLaunchPlans:   informerScope.MustNewGauge("cached_launch_plans", "Number of cached launch plans"),
		},
	}
}

// Start begins the periodic refresh of launch plan information
func (l *LaunchPlanInformer) Start(ctx context.Context) {
	ticker := time.NewTicker(l.refreshInterval)
	go func() {
		// Perform an initial refresh
		err := l.Refresh(ctx)
		if err != nil {
			logger.Errorf(ctx, "Initial launch plan informer refresh failed: %v", err)
		}

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				err := l.Refresh(ctx)
				if err != nil {
					logger.Errorf(ctx, "Launch plan informer refresh failed: %v", err)
				}
			}
		}
	}()
}

// Refresh updates the cached launch plan information
func (l *LaunchPlanInformer) Refresh(ctx context.Context) error {
	timer := l.metrics.refreshDuration.Start()
	defer timer.Stop()

	launchPlans, err := l.repo.GetAllLaunchPlansWithConcurrency(ctx)
	if err != nil {
		l.metrics.refreshFailureCount.Inc()
		return err
	}

	newLaunchPlans := make(map[string]*admin.SchedulerPolicy)
	for _, lp := range launchPlans {
		if lp.SchedulerPolicy != nil {
			lpID := getLaunchPlanKey(lp.Project, lp.Domain, lp.Name)
			policy := &admin.SchedulerPolicy{
				Max:    lp.SchedulerPolicy.MaxConcurrent,
				Policy: admin.ConcurrencyPolicy(lp.SchedulerPolicy.Policy),
				Level:  admin.ConcurrencyLevel(lp.SchedulerPolicy.Level),
			}
			newLaunchPlans[lpID] = policy
		}
	}

	l.mutex.Lock()
	l.launchPlans = newLaunchPlans
	l.mutex.Unlock()

	l.metrics.cachedLaunchPlans.Set(float64(len(newLaunchPlans)))
	l.metrics.refreshSuccessCount.Inc()
	return nil
}

// GetPolicy returns the cached concurrency policy for a launch plan
func (l *LaunchPlanInformer) GetPolicy(ctx context.Context, id core.Identifier) (*admin.SchedulerPolicy, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	lpKey := getLaunchPlanKey(id.Project, id.Domain, id.Name)
	policy, exists := l.launchPlans[lpKey]
	if !exists {
		// If not in cache, try to fetch directly
		lp, err := l.repo.GetActiveLaunchPlanWithConcurrency(ctx, id)
		if err != nil {
			return nil, err
		}

		if lp != nil && lp.SchedulerPolicy != nil {
			return &admin.SchedulerPolicy{
				Max:    lp.SchedulerPolicy.MaxConcurrent,
				Policy: admin.ConcurrencyPolicy(lp.SchedulerPolicy.Policy),
				Level:  admin.ConcurrencyLevel(lp.SchedulerPolicy.Level),
			}, nil
		}
		return nil, nil
	}
	return policy, nil
}

// getLaunchPlanKey generates a unique string key for a launch plan
func getLaunchPlanKey(project, domain, name string) string {
	return project + ":" + domain + ":" + name
}
