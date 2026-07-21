package executor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
)

func TestWatchPodTemplatesPopulatesDefaultStore(t *testing.T) {
	podTemplate := &v1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{Name: "flyte-template", Namespace: "flyte"},
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{Containers: []v1.Container{{Name: "default", Image: "img"}}},
		},
	}
	client := fake.NewSimpleClientset(podTemplate)
	informerFactory := informers.NewSharedInformerFactory(client, 0)

	require.NoError(t, watchPodTemplates(informerFactory, "flyte"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	informerFactory.Start(ctx.Done())
	cache.WaitForCacheSync(ctx.Done(), informerFactory.Core().V1().PodTemplates().Informer().HasSynced)

	// direct hit in the pod's namespace
	assert.NotNil(t, flytek8s.DefaultPodTemplateStore.LoadOrDefault("flyte", "flyte-template"))
	// fallback to the default namespace for pods running elsewhere
	assert.NotNil(t, flytek8s.DefaultPodTemplateStore.LoadOrDefault("other-ns", "flyte-template"))
	assert.Nil(t, flytek8s.DefaultPodTemplateStore.LoadOrDefault("flyte", "missing"))
}
