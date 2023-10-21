package executors

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
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
			if got := IDFromObject(p, tt.args.op); !reflect.DeepEqual(got, []byte(tt.want)) {
				t.Errorf("IDFromObject() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

type singleInvokeClient struct {
	client.Client
	createCalled bool
	deleteCalled bool
}

func (f *singleInvokeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if f.createCalled {
		return fmt.Errorf("create called more than once")
	}
	f.createCalled = true
	return nil
}

func (f *singleInvokeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if f.deleteCalled {
		return fmt.Errorf("delete called more than once")
	}
	f.deleteCalled = true
	return nil
}

func TestWriteThroughCachingWriter_Create(t *testing.T) {
	ctx := context.TODO()
	c := &singleInvokeClient{}
	w, err := newWriteThroughCachingWriter(c, 1000, promutils.NewTestScope())
	assert.NoError(t, err)

	p := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "name",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "pod",
			APIVersion: "v1",
		},
	}

	err = w.Create(ctx, p)
	assert.NoError(t, err)

	assert.True(t, c.createCalled)

	err = w.Create(ctx, p)
	assert.NoError(t, err)
}

func TestWriteThroughCachingWriter_Delete(t *testing.T) {
	ctx := context.TODO()
	c := &singleInvokeClient{}
	w, err := newWriteThroughCachingWriter(c, 1000, promutils.NewTestScope())
	assert.NoError(t, err)

	p := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "name",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "pod",
			APIVersion: "v1",
		},
	}

	err = w.Delete(ctx, p)
	assert.NoError(t, err)

	assert.True(t, c.deleteCalled)

	err = w.Delete(ctx, p)
	assert.NoError(t, err)
}

func init() {
	labeled.SetMetricKeys(contextutils.ExecIDKey)
}

// FakeReader is a fake implementation of client.Reader for testing purposes.
type FakeReader struct {
	GetFunc  func(ctx context.Context, key client.ObjectKey, obj client.Object) error
	ListFunc func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
}

func (f *FakeReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return f.GetFunc(ctx, key, obj)
}

func (f *FakeReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return f.ListFunc(ctx, list, opts...)
}

func TestFallbackClientReader_Get(t *testing.T) {
	ctx := context.Background()
	key := client.ObjectKey{}
	var out client.Object

	client1 := &FakeReader{
		GetFunc: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
			return nil
		},
	}

	fallbackReader := fallbackClientReader{
		orderedClients: []client.Reader{client1},
	}

	err := fallbackReader.Get(ctx, key, out)
	assert.Nil(t, err)

	fallbackReader = fallbackClientReader{}
	err = fallbackReader.Get(ctx, key, out)
	assert.Nil(t, err)
}

func TestFallbackClientReader_List(t *testing.T) {
	ctx := context.Background()
	var list client.ObjectList

	client1 := &FakeReader{
		ListFunc: func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
			return nil
		},
	}

	fallbackReader := fallbackClientReader{
		orderedClients: []client.Reader{client1},
	}
	err := fallbackReader.List(ctx, list)
	assert.Nil(t, err)

	fallbackReader = fallbackClientReader{
		orderedClients: []client.Reader{},
	}
	err = fallbackReader.List(ctx, list)
	assert.Nil(t, err)
}

type FakeObject struct {
	metav1.ObjectMeta
	runtime.TypeMeta
}

func (f *FakeObject) DeepCopyObject() runtime.Object {
	return &FakeObject{
		ObjectMeta: *f.ObjectMeta.DeepCopy(),
		TypeMeta:   f.TypeMeta,
	}
}

func TestFallbackClientBuilder_WithUncached(t *testing.T) {
	builder := &FallbackClientBuilder{}
	assert.NotNil(t, builder)

	obj := &FakeObject{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
			Name:      "test-name1",
		},
	}

	returnedBuilder := builder.WithUncached(obj)

	assert.Equal(t, builder, returnedBuilder)

	assert.Contains(t, builder.uncached, obj)
	assert.Equal(t, 1, len(builder.uncached))
}

type FakeCache struct {
	FakeReader
}

func (f *FakeCache) GetInformer(ctx context.Context, obj client.Object) (cache.Informer, error) {
	return nil, nil
}

func (f *FakeCache) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (cache.Informer, error) {
	return nil, nil
}

func (f *FakeCache) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return nil
}

func (f *FakeCache) Start(ctx context.Context) error {
	return nil
}

func (f *FakeCache) WaitForCacheSync(ctx context.Context) bool {
	return true
}

func TestFallbackClientBuilder_Build(t *testing.T) {
	mockCache := &FakeCache{}
	mockConfig := &rest.Config{}
	mockOptions := client.Options{}

	builder := &FallbackClientBuilder{
		uncached: []client.Object{},
		scope:    promutils.NewTestScope(),
	}

	resultClient, err := builder.Build(mockCache, mockConfig, mockOptions)

	assert.Nil(t, resultClient)
	assert.NotNil(t, err)
}
