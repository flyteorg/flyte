package k8s

import (
	"context"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

const resourceLevelMonitorCycleDuration = 10 * time.Second
const KindKey contextutils.Key = "kind"

// This object is responsible for emitting metrics that show the current number of a given K8s resource kind, cut by namespace.
// It needs to be kicked off. The periodicity is not currently configurable because it seems unnecessary. It will also
// a timer measuring how long it takes to run each measurement cycle.
type ResourceLevelMonitor struct {
	Scope promutils.Scope

	// Meta timer - this times each collection cycle to measure how long it takes to collect the levels GaugeVec below
	CollectorTimer *labeled.StopWatch

	// System Observability: This is a labeled gauge that emits the current number of objects in the informer. It is used
	// to monitor current levels.
	Levels *labeled.Gauge

	// cache is used to retrieve the object lists
	cache cache.Cache

	// The kind here will be used to differentiate all the metrics, we'll leave out group and version for now
	gvk schema.GroupVersionKind

	// We only want to create one of these ResourceLevelMonitor objects per K8s resource type. If one of these objects
	// already exists because another plugin created it, we pass along the pointer to that older object. However, we don't
	// want the collection goroutine to be kicked off multiple times.
	once sync.Once
}

// The context here is expected to already have a value for the KindKey
func (r *ResourceLevelMonitor) collect(ctx context.Context) {
	// Emit gauges at the namespace layer - since these are just K8s resources, we cannot be guaranteed to have the necessary
	// information to derive project/domain
	list := metav1.PartialObjectMetadataList{
		TypeMeta: metav1.TypeMeta{
			Kind:       r.gvk.Kind,
			APIVersion: r.gvk.GroupVersion().String(),
		},
	}
	if err := r.cache.List(ctx, &list); err != nil {
		logger.Warnf(ctx, "failed to list objects for %s.%s: %v", r.gvk.Kind, r.gvk.Version, err)
		return
	}

	// aggregate the object counts by namespace
	namespaceCounts := map[string]int{}
	for _, item := range list.Items {
		namespaceCounts[item.GetNamespace()]++
	}

	// emit namespace object count metrics
	for namespace, count := range namespaceCounts {
		withNamespaceCtx := contextutils.WithNamespace(ctx, namespace)
		r.Levels.Set(withNamespaceCtx, float64(count))
	}
}

func (r *ResourceLevelMonitor) RunCollector(ctx context.Context) {
	ticker := time.NewTicker(resourceLevelMonitorCycleDuration)
	collectorCtx := contextutils.WithGoroutineLabel(ctx, "k8s-resource-level-monitor")
	// Since the idea is that one of these objects is always only responsible for exactly one type of K8s resource, we
	// can safely set the context here for that kind for all downstream usage
	collectorCtx = context.WithValue(collectorCtx, KindKey, strings.ToLower(r.gvk.Kind))

	go func() {
		defer ticker.Stop()
		pprof.SetGoroutineLabels(collectorCtx)
		r.cache.WaitForCacheSync(collectorCtx)
		logger.Infof(ctx, "K8s resource collector %s has synced", r.gvk.Kind)
		for {
			select {
			case <-collectorCtx.Done():
				return
			case <-ticker.C:
				t := r.CollectorTimer.Start(collectorCtx)
				r.collect(collectorCtx)
				t.Stop()
			}
		}
	}()
}

func (r *ResourceLevelMonitor) RunCollectorOnce(ctx context.Context) {
	r.once.Do(func() {
		r.RunCollector(ctx)
	})
}

// This struct is here to ensure that we do not create more than one of these monitors for a given GVK. It wouldn't necessarily break
// anything, but it's a waste of compute cycles to compute counts multiple times. This can happen if multiple plugins create the same
// underlying K8s resource type. If two plugins both created Pods (ie sidecar and container), without this we would launch two
// ResourceLevelMonitor's, have two goroutines spinning, etc.
type ResourceMonitorIndex struct {
	lock     *sync.Mutex
	monitors map[schema.GroupVersionKind]*ResourceLevelMonitor

	// These are declared here because this constructor will be called more than once, by different K8s resource types (Pods, SparkApps, OtherCrd, etc.)
	// and metric names have to be unique. It felt more reasonable at time of writing to have one metric and have each resource type just be a label
	// rather than one metric per type, but can revisit this down the road.
	gauges      map[promutils.Scope]*labeled.Gauge
	stopwatches map[promutils.Scope]*labeled.StopWatch
}

func (r *ResourceMonitorIndex) GetOrCreateResourceLevelMonitor(ctx context.Context, scope promutils.Scope, cache cache.Cache,
	gvk schema.GroupVersionKind) *ResourceLevelMonitor {

	logger.Infof(ctx, "Attempting to create K8s gauge emitter for kind %s/%s", gvk.Version, gvk.Kind)

	r.lock.Lock()
	defer r.lock.Unlock()

	if r.monitors[gvk] != nil {
		logger.Infof(ctx, "Monitor for resource type %s already created, skipping...", gvk.Kind)
		return r.monitors[gvk]
	}
	logger.Infof(ctx, "Creating monitor for resource type %s...", gvk.Kind)

	// Refer to the existing labels in main.go of this repo. For these guys, we need to add namespace and kind (the K8s resource name, pod, sparkapp, etc.)
	additionalLabels := labeled.AdditionalLabelsOption{
		Labels: []string{contextutils.NamespaceKey.String(), KindKey.String()},
	}
	if r.gauges[scope] == nil {
		x := labeled.NewGauge("k8s_resources", "Current levels of K8s objects as seen from their informer caches", scope, additionalLabels)
		r.gauges[scope] = &x
	}
	if r.stopwatches[scope] == nil {
		x := labeled.NewStopWatch("k8s_collection_cycle", "Measures how long it takes to run a collection",
			time.Millisecond, scope, additionalLabels)
		r.stopwatches[scope] = &x
	}

	rm := &ResourceLevelMonitor{
		Scope:          scope,
		CollectorTimer: r.stopwatches[scope],
		Levels:         r.gauges[scope],
		cache:          cache,
		gvk:            gvk,
	}
	r.monitors[gvk] = rm

	return rm
}

// This is a global variable to this file. At runtime, the NewResourceMonitorIndex() function should only be called once
// but can be called multiple times in unit tests.
var index *ResourceMonitorIndex

func NewResourceMonitorIndex() *ResourceMonitorIndex {
	if index == nil {
		index = &ResourceMonitorIndex{
			lock:        &sync.Mutex{},
			monitors:    make(map[schema.GroupVersionKind]*ResourceLevelMonitor),
			gauges:      make(map[promutils.Scope]*labeled.Gauge),
			stopwatches: make(map[promutils.Scope]*labeled.StopWatch),
		}
	}
	return index
}
