package mocks

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeKubeClient struct {
	syncObj sync.RWMutex
	Cache   map[string]runtime.Object
}

func formatKey(name types.NamespacedName, kind schema.GroupVersionKind) string {
	key := fmt.Sprintf("%v:%v", name.String(), kind.String())
	return key
}

func (m *FakeKubeClient) Get(ctx context.Context, key client.ObjectKey, out client.Object) error {
	m.syncObj.RLock()
	defer m.syncObj.RUnlock()

	item, found := m.Cache[formatKey(key, out.GetObjectKind().GroupVersionKind())]
	if found {
		// deep copy to avoid mutating cache
		item = item.DeepCopyObject()
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

func (m *FakeKubeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	m.syncObj.RLock()
	defer m.syncObj.RUnlock()

	objs := make([]runtime.Object, 0, len(m.Cache))

	listOptions := &client.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(listOptions)
	}

	for _, val := range m.Cache {
		if listOptions.Raw != nil {
			if val.GetObjectKind().GroupVersionKind().Kind != listOptions.Raw.Kind {
				continue
			}

			if val.GetObjectKind().GroupVersionKind().GroupVersion().String() != listOptions.Raw.APIVersion {
				continue
			}
		}

		objs = append(objs, val.DeepCopyObject())
	}

	return meta.SetList(list, objs)
}

func (m *FakeKubeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) (err error) {
	m.syncObj.Lock()
	defer m.syncObj.Unlock()

	// if obj is a *v1.Pod then append a ContainerStatus for each Container
	pod, ok := obj.(*v1.Pod)
	if ok {
		for i := range pod.Spec.Containers {
			if len(pod.Status.ContainerStatuses) > i {
				continue
			}

			pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, v1.ContainerStatus{
				ContainerID: "docker://container-name",
			})
		}
	}

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	key := formatKey(types.NamespacedName{
		Name:      accessor.GetName(),
		Namespace: accessor.GetNamespace(),
	}, obj.GetObjectKind().GroupVersionKind())

	if _, exists := m.Cache[key]; !exists {
		m.Cache[key] = obj
		return nil
	}

	return errors.NewAlreadyExists(schema.GroupResource{}, accessor.GetName())
}

func (m *FakeKubeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	m.syncObj.Lock()
	defer m.syncObj.Unlock()

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	key := formatKey(types.NamespacedName{
		Name:      accessor.GetName(),
		Namespace: accessor.GetNamespace(),
	}, obj.GetObjectKind().GroupVersionKind())

	delete(m.Cache, key)

	return nil
}

func (m *FakeKubeClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	m.syncObj.Lock()
	defer m.syncObj.Unlock()

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	key := formatKey(types.NamespacedName{
		Name:      accessor.GetName(),
		Namespace: accessor.GetNamespace(),
	}, obj.GetObjectKind().GroupVersionKind())

	if _, exists := m.Cache[key]; exists {
		m.Cache[key] = obj
		return nil
	}

	return errors.NewNotFound(schema.GroupResource{}, accessor.GetName())
}

func (*FakeKubeClient) Status() client.StatusWriter {
	panic("implement me")
}

// Patch patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (*FakeKubeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	panic("implement me")

}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (*FakeKubeClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	panic("implement me")
}

func (*FakeKubeClient) Scheme() *runtime.Scheme {
	panic("implement me")
}

func (*FakeKubeClient) RESTMapper() meta.RESTMapper {
	panic("implement me")
}

func NewFakeKubeClient() *FakeKubeClient {
	return &FakeKubeClient{
		syncObj: sync.RWMutex{},
		Cache:   map[string]runtime.Object{},
	}
}
