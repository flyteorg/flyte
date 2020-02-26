package resourcemanager

import (
	"context"
	"fmt"
	"sync"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flytestdlib/promutils"
)

type TokenPrefix string

const execUrnPrefix = "ex"
const execUrnSeparator = ":"
const tokenNamespaceSeparator = "-"

func (t TokenPrefix) append(s string) TokenPrefix {
	return TokenPrefix(fmt.Sprintf("%s%s%s", t, tokenNamespaceSeparator, s))
}

func composeProjectScopePrefix(id *core.TaskExecutionIdentifier) TokenPrefix {
	return TokenPrefix(execUrnPrefix + execUrnSeparator + id.GetNodeExecutionId().GetExecutionId().GetProject())
}

func composeNamespaceScopePrefix(id *core.TaskExecutionIdentifier) TokenPrefix {
	return composeProjectScopePrefix(id).append(execUrnSeparator + id.GetNodeExecutionId().GetExecutionId().GetDomain())
}

func composeExecutionScopePrefix(id *core.TaskExecutionIdentifier) TokenPrefix {
	return composeNamespaceScopePrefix(id).append(execUrnSeparator + id.GetNodeExecutionId().GetExecutionId().GetProject())
}

func ComposeTokenPrefix(id *core.TaskExecutionIdentifier) TokenPrefix {
	return composeExecutionScopePrefix(id) // Token prefix is a required part of the token. We leverage the prefix to achieve project-level and namespace-level capping
}

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
	GetID() string
	GetResourceRegistrar(namespacePrefix pluginCore.ResourceNamespace) pluginCore.ResourceRegistrar
	BuildResourceManager(ctx context.Context) (BaseResourceManager, error)
}

// A proxy will be created for each TaskExecutionContext.
// The Proxy of an execution contains the resource namespace prefix (e.g., "qubole-hive-executor" for a hive task)
// and the token prefix (e.g., ex:<project>:<domain>:<exec_id>) of that execution.
// The plugins will only have access to a Proxy but not directly the underlying resource manager.
// The Proxy will prepend proper prefixes for the resource namespace and the allocation token.
type Proxy struct {
	// pluginCore.ResourceManager
	BaseResourceManager
	ResourceNamespacePrefix pluginCore.ResourceNamespace
	ExecutionIdentifier     *core.TaskExecutionIdentifier
}

func (p Proxy) ComposeResourceConstraint(spec pluginCore.ResourceConstraintsSpec) []ComposedResourceConstraint {
	composedResourceConstraintList := make([]ComposedResourceConstraint, 0)
	if spec.ProjectScopeResourceConstraint != nil {
		composedResourceConstraintList = append(composedResourceConstraintList, composeProjectScopeResourceConstraint(spec, p.ExecutionIdentifier))
	}
	if spec.NamespaceScopeResourceConstraint != nil {
		composedResourceConstraintList = append(composedResourceConstraintList, composeNamespaceScopeResourceConstraint(spec, p.ExecutionIdentifier))
	}
	return composedResourceConstraintList
}

func (p Proxy) AllocateResource(ctx context.Context, namespace pluginCore.ResourceNamespace,
	allocationToken string, constraintsSpec pluginCore.ResourceConstraintsSpec) (pluginCore.AllocationStatus, error) {
	composedResourceConstraintList := p.ComposeResourceConstraint(constraintsSpec)
	status, err := p.BaseResourceManager.AllocateResource(ctx,
		p.ResourceNamespacePrefix.CreateSubNamespace(namespace),
		string(ComposeTokenPrefix(p.ExecutionIdentifier).append(allocationToken)),
		composedResourceConstraintList)
	return status, err
}

func (p Proxy) ReleaseResource(ctx context.Context, namespace pluginCore.ResourceNamespace,
	allocationToken string) error {
	err := p.BaseResourceManager.ReleaseResource(ctx,
		p.ResourceNamespacePrefix.CreateSubNamespace(namespace),
		string(ComposeTokenPrefix(p.ExecutionIdentifier).append(allocationToken)))
	return err
}

func GetTaskResourceManager(r BaseResourceManager, resourceNamespacePrefix pluginCore.ResourceNamespace,
	id *core.TaskExecutionIdentifier) pluginCore.ResourceManager {
	return Proxy{
		BaseResourceManager:     r,
		ResourceNamespacePrefix: resourceNamespacePrefix,
		ExecutionIdentifier:     id,
	}
}

// The Proxy will prepend a proper prefix for the resource namespace.
type ResourceRegistrarProxy struct {
	pluginCore.ResourceRegistrar
	ResourceNamespacePrefix pluginCore.ResourceNamespace
}

func (p ResourceRegistrarProxy) RegisterResourceQuota(ctx context.Context, namespace pluginCore.ResourceNamespace, quota int) error {
	return p.ResourceRegistrar.RegisterResourceQuota(ctx,
		p.ResourceNamespacePrefix.CreateSubNamespace(namespace), quota)
}

type BaseResourceManager interface {
	GetID() string
	AllocateResource(ctx context.Context, namespace pluginCore.ResourceNamespace, allocationToken string, constraints []ComposedResourceConstraint) (pluginCore.AllocationStatus, error)
	ReleaseResource(ctx context.Context, namespace pluginCore.ResourceNamespace, allocationToken string) error
}
