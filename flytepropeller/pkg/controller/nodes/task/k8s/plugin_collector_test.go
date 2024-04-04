package k8s

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"

	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// Set variable podList using v1.PodList
var pods = &corev1.PodList{
	Items: []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a",
				Namespace: "ns-a",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b",
				Namespace: "ns-a",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "c",
				Namespace: "ns-b",
			},
		},
	},
}

func TestNewResourceLevelMonitor(t *testing.T) {
	x := corev1.Pod{}
	x.GetObjectMeta()
	lm := ResourceLevelMonitor{}
	res := lm.countList(pods)
	assert.Equal(t, 2, res["ns-a"])
	assert.Equal(t, 1, res["ns-b"])
}

type MyFakeInformer struct {
	cache.SharedIndexInformer
	store cache.Store
}

func (m MyFakeInformer) GetStore() cache.Store {
	return m.store
}

func (m MyFakeInformer) HasSynced() bool {
	return true
}

type MyFakeStore struct {
	cache.Store
}

func TestResourceLevelMonitor_setLevels(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewScope("testscope")
	fakeKubeClient := mocks.NewFakeKubeClient()

	kinds, _, err := scheme.Scheme.ObjectKinds(&corev1.Pod{})
	assert.NoError(t, err)
	myInformer := MyFakeInformer{
		store: MyFakeStore{},
	}

	index := NewResourceMonitorIndex()
	rm := index.GetOrCreateResourceLevelMonitor(ctx, scope, myInformer, kinds[0], fakeKubeClient)
	counts := map[string]int{
		"ns-a": 2,
		"ns-b": 1,
	}
	rm.setLevels(ctx, counts)

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
	fakeKubeClient := mocks.NewFakeKubeClient()

	kinds, _, err := scheme.Scheme.ObjectKinds(&corev1.Pod{})
	assert.NoError(t, err)
	myInformer := MyFakeInformer{
		store: MyFakeStore{},
	}

	index := NewResourceMonitorIndex()
	rm := index.GetOrCreateResourceLevelMonitor(ctx, scope, myInformer, kinds[0], fakeKubeClient)
	fmt.Println(rm)
	//rm2 := index.GetOrCreateResourceLevelMonitor(ctx, scope, myInformer, kinds[0])

	//assert.Equal(t, rm, rm2)
}
