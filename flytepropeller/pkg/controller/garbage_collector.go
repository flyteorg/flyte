package controller

import (
	"context"
	"runtime/pprof"
	"time"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"

	"strings"

	"github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned/typed/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type gcMetrics struct {
	gcRoundSuccess labeled.Counter
	gcRoundFailure labeled.Counter
	gcTime         labeled.StopWatch
}

// GarbageCollector is an active background cleanup service, that deletes all workflows that are completed and older
// than the configured TTL
type GarbageCollector struct {
	wfClient                  v1alpha1.FlyteworkflowV1alpha1Interface
	namespaceClient           corev1.NamespaceInterface
	ttlHours                  int
	interval                  time.Duration
	clk                       clock.Clock
	metrics                   *gcMetrics
	namespace                 string
	labelSelectorRequirements []v1.LabelSelectorRequirement
}

// Issues a background deletion command with label selector for all completed workflows outside of the retention period
func (g *GarbageCollector) deleteWorkflows(ctx context.Context) error {
	s := CompletedWorkflowsSelectorOutsideRetentionPeriod(g.ttlHours, g.clk.Now())
	if len(g.labelSelectorRequirements) != 0 {
		s.MatchExpressions = append(s.MatchExpressions, g.labelSelectorRequirements...)
	}

	// Delete doesn't support 'all' namespaces. Let's fetch namespaces and loop over each.
	if g.namespace == "" || strings.ToLower(g.namespace) == "all" || strings.ToLower(g.namespace) == "all-namespaces" {
		namespaceList, err := g.namespaceClient.List(ctx, v1.ListOptions{})
		if err != nil {
			return err
		}
		for _, n := range namespaceList.Items {
			namespaceCtx := contextutils.WithNamespace(ctx, n.GetName())
			logger.Infof(namespaceCtx, "Triggering Workflow delete for namespace: [%s]", n.GetName())

			if err := g.deleteWorkflowsForNamespace(ctx, n.GetName(), s); err != nil {
				g.metrics.gcRoundFailure.Inc(namespaceCtx)
				logger.Errorf(namespaceCtx, "Garbage collection failed for for namespace: [%s]. Error : [%v]", n.GetName(), err)
			} else {
				g.metrics.gcRoundSuccess.Inc(namespaceCtx)
			}
		}
	} else {
		namespaceCtx := contextutils.WithNamespace(ctx, g.namespace)
		logger.Infof(namespaceCtx, "Triggering Workflow delete for namespace: [%s]", g.namespace)
		if err := g.deleteWorkflowsForNamespace(ctx, g.namespace, s); err != nil {
			g.metrics.gcRoundFailure.Inc(namespaceCtx)
			logger.Errorf(namespaceCtx, "Garbage collection failed for for namespace: [%s]. Error : [%v]", g.namespace, err)
		} else {
			g.metrics.gcRoundSuccess.Inc(namespaceCtx)
		}
	}
	return nil
}

// Deprecated: Please use deleteWorkflows instead
func (g *GarbageCollector) deprecatedDeleteWorkflows(ctx context.Context) error {
	s := DeprecatedCompletedWorkflowsSelectorOutsideRetentionPeriod(g.ttlHours, g.clk.Now())
	if len(g.labelSelectorRequirements) != 0 {
		s.MatchExpressions = append(s.MatchExpressions, g.labelSelectorRequirements...)
	}

	// Delete doesn't support 'all' namespaces. Let's fetch namespaces and loop over each.
	if g.namespace == "" || strings.ToLower(g.namespace) == "all" || strings.ToLower(g.namespace) == "all-namespaces" {
		namespaceList, err := g.namespaceClient.List(ctx, v1.ListOptions{})
		if err != nil {
			return err
		}
		for _, n := range namespaceList.Items {
			namespaceCtx := contextutils.WithNamespace(ctx, n.GetName())
			logger.Infof(namespaceCtx, "Triggering Workflow delete for namespace: [%s]", n.GetName())

			if err := g.deleteWorkflowsForNamespace(ctx, n.GetName(), s); err != nil {
				g.metrics.gcRoundFailure.Inc(namespaceCtx)
				logger.Errorf(namespaceCtx, "Garbage collection failed for for namespace: [%s]. Error : [%v]", n.GetName(), err)
			} else {
				g.metrics.gcRoundSuccess.Inc(namespaceCtx)
			}
		}
	} else {
		namespaceCtx := contextutils.WithNamespace(ctx, g.namespace)
		logger.Infof(namespaceCtx, "Triggering Workflow delete for namespace: [%s]", g.namespace)
		if err := g.deleteWorkflowsForNamespace(ctx, g.namespace, s); err != nil {
			g.metrics.gcRoundFailure.Inc(namespaceCtx)
			logger.Errorf(namespaceCtx, "Garbage collection failed for for namespace: [%s]. Error : [%v]", g.namespace, err)
		} else {
			g.metrics.gcRoundSuccess.Inc(namespaceCtx)
		}
	}
	return nil
}

func (g *GarbageCollector) deleteWorkflowsForNamespace(ctx context.Context, namespace string, labelSelector *v1.LabelSelector) error {
	gracePeriodZero := int64(0)
	propagation := v1.DeletePropagationBackground

	return g.wfClient.FlyteWorkflows(namespace).DeleteCollection(
		ctx,
		v1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodZero,
			PropagationPolicy:  &propagation,
		},
		v1.ListOptions{
			LabelSelector: v1.FormatLabelSelector(labelSelector),
		},
	)
}

// runGC runs GC periodically
func (g *GarbageCollector) runGC(ctx context.Context, ticker clock.Ticker) {
	logger.Infof(ctx, "Background workflow garbage collection started, with duration [%s], TTL [%d] hours", g.interval.String(), g.ttlHours)

	ctx = contextutils.WithGoroutineLabel(ctx, "gc-worker")
	pprof.SetGoroutineLabels(ctx)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C():
			logger.Infof(ctx, "Garbage collector running...")
			t := g.metrics.gcTime.Start(ctx)
			if err := g.deprecatedDeleteWorkflows(ctx); err != nil {
				logger.Errorf(ctx, "Garbage collection failed in this round.Error : [%v]", err)
			}

			if err := g.deleteWorkflows(ctx); err != nil {
				logger.Errorf(ctx, "Garbage collection failed in this round.Error : [%v]", err)
			}

			t.Stop()
		case <-ctx.Done():
			logger.Infof(ctx, "Garbage collector stopping")
			return

		}
	}
}

// StartGC starts a background garbage collection routine. Use the context to signal an exit signal
func (g *GarbageCollector) StartGC(ctx context.Context) error {
	if g.ttlHours <= 0 {
		logger.Warningf(ctx, "Garbage collector is disabled, as ttl [%d] is <=0", g.ttlHours)
		return nil
	}
	ticker := g.clk.NewTicker(g.interval)
	go g.runGC(ctx, ticker)
	return nil
}

func NewGarbageCollector(cfg *config.Config, scope promutils.Scope, clk clock.Clock, namespaceClient corev1.NamespaceInterface, wfClient v1alpha1.FlyteworkflowV1alpha1Interface) (*GarbageCollector, error) {
	ttl := 23
	if cfg.MaxTTLInHours < 23 {
		ttl = cfg.MaxTTLInHours
	} else {
		logger.Warningf(context.TODO(), "defaulting max ttl for workflows to 23 hours, since configured duration is larger than 23 [%d]", cfg.MaxTTLInHours)
	}
	labelSelectorRequirements := getShardedLabelSelectorRequirements(cfg)
	return &GarbageCollector{
		wfClient:        wfClient,
		ttlHours:        ttl,
		interval:        cfg.GCInterval.Duration,
		namespaceClient: namespaceClient,
		metrics: &gcMetrics{
			gcTime:         labeled.NewStopWatch("gc_latency", "time taken to issue a delete for TTL'ed workflows", time.Millisecond, scope),
			gcRoundSuccess: labeled.NewCounter("gc_success", "successful executions of delete request", scope),
			gcRoundFailure: labeled.NewCounter("gc_failure", "failure to delete workflows", scope),
		},
		clk:                       clk,
		namespace:                 cfg.LimitNamespace,
		labelSelectorRequirements: labelSelectorRequirements,
	}, nil
}
