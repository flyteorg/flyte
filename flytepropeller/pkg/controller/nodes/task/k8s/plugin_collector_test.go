package k8s

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var pods = []interface{}{
	&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "a",
			Namespace: "ns-a",
		},
	},
	&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "b",
			Namespace: "ns-a",
		},
	},
	&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "c",
			Namespace: "ns-b",
		},
	},
}

type MyFakeCache struct {
	cache.Cache
}

func (m MyFakeCache) List(_ context.Context, list client.ObjectList, _ ...client.ListOption) error {
	objectMetadataList, ok := list.(*metav1.PartialObjectMetadataList)
	if !ok {
		return fmt.Errorf("unexpected type %T", list)
	}

	objectMetadataList.Items = make([]metav1.PartialObjectMetadata, 0)
	for _, pod := range pods {
		objectMetadataList.Items = append(objectMetadataList.Items, metav1.PartialObjectMetadata{
			TypeMeta:   objectMetadataList.TypeMeta,
			ObjectMeta: pod.(*v1.Pod).ObjectMeta,
		})
	}

	return nil
}

func (m MyFakeCache) WaitForCacheSync(_ context.Context) bool {
	return true
}

func TestResourceLevelMonitor_collect(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewScope("testscope")

	kinds, _, err := scheme.Scheme.ObjectKinds(&v1.Pod{})
	assert.NoError(t, err)
	myCache := MyFakeCache{}

	index := NewResourceMonitorIndex()
	rm := index.GetOrCreateResourceLevelMonitor(ctx, scope, myCache, kinds[0])
	rm.collect(ctx)

	var expected = `
		# HELP testscope:k8s_resources Current levels of K8s objects as seen from their informer caches
		# TYPE testscope:k8s_resources gauge
		testscope:k8s_resources{kind="",ns="ns-a",project=""} 2
		testscope:k8s_resources{kind="",ns="ns-b",project=""} 1
	`

	err = testutil.CollectAndCompare(rm.Levels.GaugeVec, strings.NewReader(expected))
	assert.NoError(t, err)
}

func TestResourceLevelMonitorSingletonness(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewScope("testscope")

	kinds, _, err := scheme.Scheme.ObjectKinds(&v1.Pod{})
	assert.NoError(t, err)
	myCache := MyFakeCache{}

	index := NewResourceMonitorIndex()
	rm := index.GetOrCreateResourceLevelMonitor(ctx, scope, myCache, kinds[0])
	rm2 := index.GetOrCreateResourceLevelMonitor(ctx, scope, myCache, kinds[0])

	assert.Equal(t, rm, rm2)
}
