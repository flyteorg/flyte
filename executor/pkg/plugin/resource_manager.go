package plugin

import (
	"context"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
)

// The webapi plugin machinery requires the host to supply a ResourceRegistrar at plugin setup and a
// ResourceManager at task execution, for its allocation-token quota feature. FlytePropeller (v1)
// always supplied them, defaulting to a no-op manager when no quota backend is configured. The v2
// executor never reimplemented this, so a webapi plugin that declares ResourceQuotas dereferences a
// nil registrar or manager. These no-op types restore the contract and grant every allocation. Swap
// in a real ResourceManager to enforce quotas without touching any plugin.

// noopResourceRegistrar accepts quota registrations without recording them.
type noopResourceRegistrar struct{}

var _ pluginsCore.ResourceRegistrar = noopResourceRegistrar{}

func (noopResourceRegistrar) RegisterResourceQuota(_ context.Context, _ pluginsCore.ResourceNamespace, _ int) error {
	return nil
}

// noopResourceManager grants every allocation request and ignores releases.
type noopResourceManager struct{}

var _ pluginsCore.ResourceManager = noopResourceManager{}

func (noopResourceManager) GetID() string { return "executor-noop-resource-manager" }

func (noopResourceManager) AllocateResource(_ context.Context, _ pluginsCore.ResourceNamespace, _ string, _ pluginsCore.ResourceConstraintsSpec) (pluginsCore.AllocationStatus, error) {
	return pluginsCore.AllocationStatusGranted, nil
}

func (noopResourceManager) ReleaseResource(_ context.Context, _ pluginsCore.ResourceNamespace, _ string) error {
	return nil
}

// NewNoopResourceRegistrar returns a ResourceRegistrar that accepts quota declarations without
// enforcing them.
func NewNoopResourceRegistrar() pluginsCore.ResourceRegistrar { return noopResourceRegistrar{} }

// NewNoopResourceManager returns a ResourceManager that grants every allocation.
func NewNoopResourceManager() pluginsCore.ResourceManager { return noopResourceManager{} }
