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

func (f *FakeInformers) Get(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	return nil
}

func (f *FakeInformers) WaitForCacheSync(context.Context) bool {
	return true
}

func (f *FakeInformers) IndexField(context.Context, client.Object, string, client.IndexerFunc) error {
	return nil
}

func (f *FakeInformers) List(context.Context, client.ObjectList, ...client.ListOption) error {
	return nil
}

func (f *FakeInformers) GetInformerForKind(context.Context, schema.GroupVersionKind, ...cache.InformerGetOption) (cache.Informer, error) {
	return &controllertest.FakeInformer{}, nil
}

func (f *FakeInformers) Start(context.Context) error {
	return nil
}

func (f *FakeInformers) GetInformer(context.Context, client.Object, ...cache.InformerGetOption) (cache.Informer, error) {
	return &controllertest.FakeInformer{}, nil
}

func NewFakeKubeClient() *Client {
	c := Client{}
	c.EXPECT().GetClient().Return(fake.NewClientBuilder().WithRuntimeObjects().Build())
	c.EXPECT().GetCache().Return(&FakeInformers{})
	return &c
}
