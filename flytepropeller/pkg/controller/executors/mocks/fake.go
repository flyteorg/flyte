package mocks

import (
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func NewFakeKubeClient() *Client {
	c := Client{}
	c.On("GetClient").Return(fake.NewClientBuilder().WithRuntimeObjects().Build())
	c.On("GetCache").Return(&informertest.FakeInformers{})
	return &c
}
