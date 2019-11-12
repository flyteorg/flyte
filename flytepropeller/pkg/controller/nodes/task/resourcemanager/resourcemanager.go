package resourcemanager

import (
	"context"
	"sync"

	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flytestdlib/promutils"
)

//go:generate mockery -name ResourceManager -case=underscore

// This struct is designed to serve as the identifier of an user of resource manager
type Resource struct {
	quota          int
	metrics        Metrics
	rejectedTokens sync.Map
}

type Metrics interface {
	GetScope() promutils.Scope
}

type Builder interface {
	ResourceRegistrar(namespacePrefix pluginCore.ResourceNamespace) pluginCore.ResourceRegistrar
	BuildResourceManager(ctx context.Context) (pluginCore.ResourceManager, error)
}

type Proxy struct {
	pluginCore.ResourceManager
	NamespacePrefix pluginCore.ResourceNamespace
}

func (p Proxy) getPrefixedNamespace(namespace pluginCore.ResourceNamespace) pluginCore.ResourceNamespace {
	return p.NamespacePrefix.CreateSubNamespace(namespace)
}

func (p Proxy) AllocateResource(ctx context.Context, namespace pluginCore.ResourceNamespace,
	allocationToken string) (pluginCore.AllocationStatus, error) {
	status, err := p.ResourceManager.AllocateResource(ctx, p.getPrefixedNamespace(namespace), allocationToken)
	return status, err
}

func (p Proxy) ReleaseResource(ctx context.Context, namespace pluginCore.ResourceNamespace,
	allocationToken string) error {
	err := p.ResourceManager.ReleaseResource(ctx, p.getPrefixedNamespace(namespace), allocationToken)
	return err
}

type ResourceRegistrarProxy struct {
	pluginCore.ResourceRegistrar
	NamespacePrefix pluginCore.ResourceNamespace
}

func (p ResourceRegistrarProxy) getPrefixedNamespace(namespace pluginCore.ResourceNamespace) pluginCore.ResourceNamespace {
	return p.NamespacePrefix.CreateSubNamespace(namespace)
}

func (p ResourceRegistrarProxy) RegisterResourceQuota(ctx context.Context, namespace pluginCore.ResourceNamespace, quota int) error {
	return p.ResourceRegistrar.RegisterResourceQuota(ctx, p.getPrefixedNamespace(namespace), quota)
}
