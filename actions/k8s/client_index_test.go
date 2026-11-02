package k8s

import (
	"context"
	"testing"
	"time"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	toolscache "k8s.io/client-go/tools/cache"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Only cache reads are available; the ActionsClient has no API client to fall back to.
type indexedTestCache struct {
	ctrlcache.Cache
	reader client.Reader
}

func (c *indexedTestCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return c.reader.Get(ctx, key, obj, opts...)
}
func (c *indexedTestCache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return c.reader.List(ctx, list, opts...)
}

func TestListChildActions_IndexedCache(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	require.NoError(t, executorv1.AddToScheme(scheme))
	store := fake.NewClientBuilder().WithScheme(scheme).
		WithIndex(&executorv1.TaskAction{}, parentActionIndex, indexTaskActionParent).Build()
	c := &ActionsClient{namespace: "flyte", sharedCache: &indexedTestCache{reader: store}}
	parentID := &common.ActionIdentifier{Run: &common.RunIdentifier{Project: "proj", Domain: "dev", Name: "run1"}, Name: "a0"}
	create := func(name, parent, project, domain, run, namespace string) *executorv1.TaskAction {
		ta := &executorv1.TaskAction{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Spec:       executorv1.TaskActionSpec{Project: project, Domain: domain, RunName: run, ActionName: name},
		}
		if parent != "" {
			ta.Spec.ParentActionName = &parent
		}
		require.NoError(t, store.Create(ctx, ta))
		return ta
	}
	parent := create("run1-a0", "", "proj", "dev", "run1", "flyte")
	child := create("child", "a0", "proj", "dev", "run1", "flyte")
	second := create("second", "a0", "proj", "dev", "run1", "flyte")
	create("other-run", "a0", "proj", "dev", "run2", "flyte")
	create("other-project", "a0", "other", "dev", "run1", "flyte")
	create("other-domain", "a0", "proj", "other", "run1", "flyte")
	create("other-namespace", "a0", "proj", "dev", "run1", "other")
	create("grandchild", "child", "proj", "dev", "run1", "flyte")
	create("unrelated-root", "", "proj", "dev", "run1", "flyte")
	assertNames := func(expected ...string) {
		t.Helper()
		got, err := c.ListChildActions(ctx, parentID)
		require.NoError(t, err)
		var names []string
		for _, ta := range got {
			names = append(names, ta.Name)
		}
		assert.ElementsMatch(t, expected, names)
	}
	assertNames(parent.Name, child.Name, second.Name)
	// Returned objects must be safe for callers to mutate.
	got, err := c.ListChildActions(ctx, parentID)
	require.NoError(t, err)
	got[0].Name = "mutated"
	assertNames(parent.Name, child.Name, second.Name)
	other := "other-parent"
	child.Spec.ParentActionName = &other
	require.NoError(t, store.Update(ctx, child))
	assertNames(parent.Name, second.Name)
	require.NoError(t, store.Delete(ctx, parent))
	assertNames(second.Name)
	require.NoError(t, store.Delete(ctx, second))
	assertNames()
}

// This source supports list/watch, not streaming initial events.
type startupListWatch struct{ *toolscache.ListWatch }

func (startupListWatch) IsWatchListSemanticsUnSupported() bool { return true }

type watcherStartupCache struct {
	ctrlcache.Cache
	informer   toolscache.SharedIndexInformer
	indexCalls int
}

func (c *watcherStartupCache) IndexField(_ context.Context, _ client.Object, field string, extract client.IndexerFunc) error {
	c.indexCalls++
	return c.informer.AddIndexers(toolscache.Indexers{
		field: func(obj interface{}) ([]string, error) { return extract(obj.(client.Object)), nil },
	})
}
func (c *watcherStartupCache) GetInformer(context.Context, client.Object, ...ctrlcache.InformerGetOption) (ctrlcache.Informer, error) {
	return c.informer, nil
}

func TestStartWatching_AfterInformerStarted(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w := watch.NewRaceFreeFake()
	defer w.Stop()
	informer := toolscache.NewSharedIndexInformer(startupListWatch{&toolscache.ListWatch{
		ListFunc:  func(metav1.ListOptions) (runtime.Object, error) { return &executorv1.TaskActionList{}, nil },
		WatchFunc: func(metav1.ListOptions) (watch.Interface, error) { return w, nil },
	}}, &executorv1.TaskAction{}, 0, toolscache.Indexers{})
	cache := &watcherStartupCache{informer: informer}
	require.NoError(t, RegisterTaskActionIndexes(ctx, cache))
	go informer.Run(ctx.Done())
	require.True(t, toolscache.WaitForCacheSync(ctx.Done(), informer.HasSynced))
	c := &ActionsClient{sharedCache: cache, namespace: "flyte", numWorkers: 1, bufferSize: 1}
	require.NoError(t, c.StartWatching(ctx))
	defer c.StopWatching()
	assert.Equal(t, 1, cache.indexCalls, "StartWatching must not register indexes on a running informer")
	assert.Contains(t, informer.GetIndexer().GetIndexers(), parentActionIndex)
}
