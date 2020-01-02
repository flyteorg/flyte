package resourcemanager

import (
	"context"

	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
)

type NoopResourceManagerBuilder struct {
}

func (r *NoopResourceManagerBuilder) ResourceRegistrar(namespacePrefix pluginCore.ResourceNamespace) pluginCore.ResourceRegistrar {
	return ResourceRegistrarProxy{
		ResourceRegistrar:       r,
		ResourceNamespacePrefix: namespacePrefix,
	}
}

func (r *NoopResourceManagerBuilder) RegisterResourceQuota(ctx context.Context, namespace pluginCore.ResourceNamespace, quota int) error {
	return nil
}

func (r *NoopResourceManagerBuilder) BuildResourceManager(ctx context.Context) (pluginCore.ResourceManager, error) {
	return &NoopResourceManager{}, nil
}

type NoopResourceManager struct {
}

func (*NoopResourceManager) AllocateResource(ctx context.Context, namespace pluginCore.ResourceNamespace, allocationToken string) (
	pluginCore.AllocationStatus, error) {

	return pluginCore.AllocationStatusGranted, nil
}

func (*NoopResourceManager) ReleaseResource(ctx context.Context, namespace pluginCore.ResourceNamespace, allocationToken string) error {
	return nil
}
