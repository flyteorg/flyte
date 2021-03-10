package k8s

import (
	"context"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
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

	// This informer will be used to get a list of the underlying objects that we want a tally of
	sharedInformer cache.SharedIndexInformer

	// The kind here will be used to differentiate all the metrics, we'll leave out group and version for now
	gvk schema.GroupVersionKind

	// We only want to create one of these ResourceLevelMonitor objects per K8s resource type. If one of these objects
	// already exists because another plugin created it, we pass along the pointer to that older object. However, we don't
	// want the collection goroutine to be kicked off multiple times.
	once sync.Once
}

// The reason that we use namespace as the one and only thing to cut by is because it's the feature that we are sure that any
// K8s resource created by a plugin will have (as yet, Flyte doesn't have any plugins that create cluster level resources and
// it probably won't for a long time). We can't assume that all the operators and CRDs that Flyte will ever work with will have
// the exact same set of labels or annotations or owner references. The only thing we can really count on is namespace.
func (r *ResourceLevelMonitor) countList(ctx context.Context, objects []interface{}) map[string]int {
	// Map of namespace to counts
	counts := map[string]int{}

	// Collect the object counts by namespace
	for _, v := range objects {
		metadata, err := meta.Accessor(v)
		if err != nil {
			logger.Errorf(ctx, "Error converting obj %v to an Accessor %s\n", v, err)
			continue
		}
		counts[metadata.GetNamespace()]++
	}

	return counts
}

// The context here is expected to already have a value for the KindKey
func (r *ResourceLevelMonitor) collect(ctx context.Context) {
	// Emit gauges at the namespace layer - since these are just K8s resources, we cannot be guaranteed to have the necessary
	// information to derive project/domain
	objects := r.sharedInformer.GetStore().List()
	counts := r.countList(ctx, objects)

	for ns, count := range counts {
		withNamespaceCtx := contextutils.WithNamespace(ctx, ns)
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
		r.sharedInformer.HasSynced()
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

func (r *ResourceMonitorIndex) GetOrCreateResourceLevelMonitor(ctx context.Context, scope promutils.Scope, si cache.SharedIndexInformer,
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
		sharedInformer: si,
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
