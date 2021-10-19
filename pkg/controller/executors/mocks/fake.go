package mocks

import (
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func NewFakeKubeClient() *Client {
	c := Client{}
	c.OnGetClient().Return(fake.NewClientBuilder().WithRuntimeObjects().Build())
	c.OnGetCache().Return(&informertest.FakeInformers{})
	return &c
}
