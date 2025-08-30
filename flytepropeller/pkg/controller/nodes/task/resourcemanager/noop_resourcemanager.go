package resourcemanager

import (
	"context"

	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
)

const NoopResourceManagerID = "noopresourcemanager"

type NoopResourceManagerBuilder struct {
}

func (r *NoopResourceManagerBuilder) GetID() string {
	return NoopResourceManagerID
}

func (r *NoopResourceManagerBuilder) GetResourceRegistrar(namespacePrefix pluginCore.ResourceNamespace) pluginCore.ResourceRegistrar {
	return ResourceRegistrarProxy{
		ResourceRegistrar:       r,
		ResourceNamespacePrefix: namespacePrefix,
	}
}

func (r *NoopResourceManagerBuilder) RegisterResourceQuota(context.Context, pluginCore.ResourceNamespace, int) error {
	return nil
}

func (r *NoopResourceManagerBuilder) BuildResourceManager(context.Context) (BaseResourceManager, error) {
	return &NoopResourceManager{}, nil
}

type NoopResourceManager struct {
}

func (*NoopResourceManager) GetID() string {
	return NoopResourceManagerID
}

func (*NoopResourceManager) AllocateResource(context.Context, pluginCore.ResourceNamespace, Token, []FullyQualifiedResourceConstraint) (pluginCore.AllocationStatus, error) {

	return pluginCore.AllocationStatusGranted, nil
}

func (*NoopResourceManager) ReleaseResource(context.Context, pluginCore.ResourceNamespace, Token) error {
	return nil
}
