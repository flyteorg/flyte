package mocks

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllertest"
)

// FakeInformers is a fake implementation of Informers.
// Implement a fake informer cache to avoid race condition in informertest.FakeInformers
type FakeInformers struct {
}

func (f *FakeInformers) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return nil
}

func (f *FakeInformers) WaitForCacheSync(ctx context.Context) bool {
	return true
}

func (f *FakeInformers) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return nil
}

func (f *FakeInformers) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

func (f *FakeInformers) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (cache.Informer, error) {
	return &controllertest.FakeInformer{}, nil
}

func (f *FakeInformers) Start(ctx context.Context) error {
	return nil
}

func (f *FakeInformers) GetInformer(ctx context.Context, obj client.Object) (cache.Informer, error) {
	return &controllertest.FakeInformer{}, nil
}

func NewFakeKubeClient() *Client {
	c := Client{}
	c.OnGetClient().Return(fake.NewClientBuilder().WithRuntimeObjects().Build())
	c.OnGetCache().Return(&FakeInformers{})
	return &c
}
