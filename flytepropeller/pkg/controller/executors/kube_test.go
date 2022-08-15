package executors

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
