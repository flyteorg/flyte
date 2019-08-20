package mocks

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockRuntimeClient struct {
	syncObj  sync.RWMutex
	Cache    map[string]runtime.Object
	CreateCb func(ctx context.Context, obj runtime.Object) (err error)
	GetCb    func(ctx context.Context, key client.ObjectKey, out runtime.Object) error
	ListCb   func(ctx context.Context, opts *client.ListOptions, list runtime.Object) error
}

func formatKey(name types.NamespacedName, kind schema.GroupVersionKind) string {
	key := fmt.Sprintf("%v:%v", name.String(), kind.String())
	return key
}

func (m MockRuntimeClient) Get(ctx context.Context, key client.ObjectKey, out runtime.Object) error {
	if m.GetCb != nil {
		return m.GetCb(ctx, key, out)
	}
	return m.defaultGet(ctx, key, out)
}

func (m MockRuntimeClient) Create(ctx context.Context, obj runtime.Object) (err error) {
	if m.CreateCb != nil {
		return m.CreateCb(ctx, obj)
	}
	return m.defaultCreate(ctx, obj)
}

func (m MockRuntimeClient) List(ctx context.Context, opts *client.ListOptions, list runtime.Object) error {
	if m.ListCb != nil {
		return m.ListCb(ctx, opts, list)
	}
	return m.defaultList(ctx, opts, list)
}

func (*MockRuntimeClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error {
	panic("implement me")
}

func (m *MockRuntimeClient) Update(ctx context.Context, obj runtime.Object) error {
	// TODO: split update/create, create should fail if already exists.
	return m.Create(ctx, obj)
}

func (*MockRuntimeClient) Status() client.StatusWriter {
	panic("implement me")
}

func (m MockRuntimeClient) defaultGet(ctx context.Context, key client.ObjectKey, out runtime.Object) error {
	m.syncObj.RLock()
	defer m.syncObj.RUnlock()

	item, found := m.Cache[formatKey(key, out.GetObjectKind().GroupVersionKind())]
	if found {
		// deep copy to avoid mutating cache
		item = item.(runtime.Object).DeepCopyObject()
		_, isUnstructured := out.(*unstructured.Unstructured)
		if isUnstructured {
			// Copy the value of the item in the cache to the returned value
			outVal := reflect.ValueOf(out)
			objVal := reflect.ValueOf(item)
			if !objVal.Type().AssignableTo(outVal.Type()) {
				return fmt.Errorf("cache had type %s, but %s was asked for", objVal.Type(), outVal.Type())
			}
			reflect.Indirect(outVal).Set(reflect.Indirect(objVal))
			return nil
		}

		p, err := runtime.DefaultUnstructuredConverter.ToUnstructured(item)
		if err != nil {
			return err
		}

		return runtime.DefaultUnstructuredConverter.FromUnstructured(p, out)
	}

	return errors.NewNotFound(schema.GroupResource{}, key.Name)
}

func (m MockRuntimeClient) defaultList(ctx context.Context, opts *client.ListOptions, list runtime.Object) error {
	m.syncObj.RLock()
	defer m.syncObj.RUnlock()

	objs := make([]runtime.Object, 0, len(m.Cache))

	for _, val := range m.Cache {
		if opts.Raw != nil {
			if val.GetObjectKind().GroupVersionKind().Kind != opts.Raw.Kind {
				continue
			}

			if val.GetObjectKind().GroupVersionKind().GroupVersion().String() != opts.Raw.APIVersion {
				continue
			}
		}

		objs = append(objs, val.(runtime.Object).DeepCopyObject())
	}

	return meta.SetList(list, objs)
}

func (m MockRuntimeClient) defaultCreate(ctx context.Context, obj runtime.Object) (err error) {
	m.syncObj.Lock()
	defer m.syncObj.Unlock()

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	m.Cache[formatKey(types.NamespacedName{
		Name:      accessor.GetName(),
		Namespace: accessor.GetNamespace(),
	}, obj.GetObjectKind().GroupVersionKind())] = obj

	return nil
}

func NewMockRuntimeClient() *MockRuntimeClient {
	return &MockRuntimeClient{
		syncObj: sync.RWMutex{},
		Cache:   map[string]runtime.Object{},
	}
}
