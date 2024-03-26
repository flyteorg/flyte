package executors

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

func TestIdFromObject(t *testing.T) {
	type args struct {
		ns   string
		name string
		kind string
		op   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"default", args{"default", "name", "pod", "c"}, "/v1, Kind=pod:default:name:c"},
		{"no-cluster", args{"my-ns", "name", "pod", "c"}, "/v1, Kind=pod:my-ns:name:c"},
		{"differ-oper", args{"default", "name", "pod", "d"}, "/v1, Kind=pod:default:name:d"},
		{"withcluster", args{"default", "name", "pod", "d"}, "/v1, Kind=pod:default:name:d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tt.args.ns,
					Name:      tt.args.name,
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       tt.args.kind,
					APIVersion: "v1",
				},
			}
			if got := idFromObject(p, tt.args.op); !reflect.DeepEqual(got, []byte(tt.want)) {
				t.Errorf("idFromObject() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

type mockKubeClient struct {
	client.Client
	createCalledCount int
	deleteCalledCount int
	getCalledCount    int
	getMissCount      int
}

func (m *mockKubeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	m.createCalledCount++
	return nil
}

func (m *mockKubeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	m.deleteCalledCount++
	return nil
}

func (m *mockKubeClient) Get(ctx context.Context, objectKey types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	if m.getCalledCount < m.getMissCount {
		m.getMissCount--
		return k8serrors.NewNotFound(v1.Resource("pod"), "name")
	}

	m.getCalledCount++
	return nil
}

func TestFlyteK8sClient(t *testing.T) {
	ctx := context.TODO()
	scope := promutils.NewTestScope()

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "name",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "pod",
			APIVersion: "v1",
		},
	}

	objectKey := types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	}

	// test cache reader
	tests := []struct {
		name                   string
		initCacheReader        bool
		cacheMissCount         int
		expectedClientGetCount int
	}{
		{"no-cache", false, 0, 2},
		{"with-cache-one-miss", true, 1, 1},
		{"with-cache-no-misses", true, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cacheReader client.Reader
			if tt.initCacheReader {
				cacheReader = &mockKubeClient{
					getMissCount: tt.cacheMissCount,
				}
			}

			kubeClient := &mockKubeClient{}

			flyteK8sClient, err := newFlyteK8sClient(kubeClient, cacheReader, scope.NewSubScope(tt.name))
			assert.NoError(t, err)

			for i := 0; i < 2; i++ {
				err := flyteK8sClient.Get(ctx, objectKey, pod)
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedClientGetCount, kubeClient.getCalledCount)
		})
	}

	// test create
	t.Run("create", func(t *testing.T) {
		kubeClient := &mockKubeClient{}
		flyteK8sClient, err := newFlyteK8sClient(kubeClient, nil, scope.NewSubScope("create"))
		assert.NoError(t, err)

		for i := 0; i < 5; i++ {
			err = flyteK8sClient.Create(ctx, pod)
			assert.NoError(t, err)
		}

		assert.Equal(t, 1, kubeClient.createCalledCount)
	})

	// test delete
	t.Run("delete", func(t *testing.T) {
		kubeClient := &mockKubeClient{}
		flyteK8sClient, err := newFlyteK8sClient(kubeClient, nil, scope.NewSubScope("delete"))
		assert.NoError(t, err)

		for i := 0; i < 5; i++ {
			err = flyteK8sClient.Delete(ctx, pod)
			assert.NoError(t, err)
		}

		assert.Equal(t, 1, kubeClient.deleteCalledCount)
	})
}

func init() {
	labeled.SetMetricKeys(contextutils.ExecIDKey)
}
