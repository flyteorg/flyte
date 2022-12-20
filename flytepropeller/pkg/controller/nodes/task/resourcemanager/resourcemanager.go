package resourcemanager

import (
	"context"
	"fmt"
	"sync"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flytestdlib/promutils"
)

type TokenPrefix string
type Token string

const execUrnPrefix = "ex"
const execUrnSeparator = ":"
const tokenNamespaceSeparator = "-"

// Prepending the prefix to the actual token string using '-' separator.
// The output is of type Token
func (t Token) prepend(prefix TokenPrefix) Token {
	return Token(fmt.Sprintf("%s%s%s", prefix, tokenNamespaceSeparator, t))
}

// Extending the prefix using ':' separator. The output is still of type TokenPrefix
func (t TokenPrefix) extend(prefixPart string) TokenPrefix {
	return TokenPrefix(fmt.Sprintf("%s%s%s", t, execUrnSeparator, prefixPart))
}

func composeProjectScopePrefix(id *core.TaskExecutionIdentifier) TokenPrefix {
	return TokenPrefix(execUrnPrefix).extend(id.GetNodeExecutionId().GetExecutionId().GetProject())
}

func composeNamespaceScopePrefix(id *core.TaskExecutionIdentifier) TokenPrefix {
	return composeProjectScopePrefix(id).extend(id.GetNodeExecutionId().GetExecutionId().GetDomain())
}

func composeExecutionScopePrefix(id *core.TaskExecutionIdentifier) TokenPrefix {
	return composeNamespaceScopePrefix(id).extend(id.GetNodeExecutionId().GetExecutionId().GetName())
}

func ComposeTokenPrefix(id *core.TaskExecutionIdentifier) TokenPrefix {
	return composeExecutionScopePrefix(id) // Token prefix is a required part of the token. We leverage the prefix to achieve project-level and namespace-level capping
}

// This struct is designed to serve as the identifier of an user of resource manager
type Resource struct {
	quota          BaseResourceConstraint
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
	ResourcePoolInfo        map[string]*event.ResourcePoolInfo
}

type TaskResourceManager interface {
	pluginCore.ResourceManager
	GetResourcePoolInfo() []*event.ResourcePoolInfo
}

func (p Proxy) ComposeResourceConstraint(spec pluginCore.ResourceConstraintsSpec) []FullyQualifiedResourceConstraint {
	composedResourceConstraintList := make([]FullyQualifiedResourceConstraint, 0)
	if spec.ProjectScopeResourceConstraint != nil {
		composedResourceConstraintList = append(composedResourceConstraintList, composeFullyQualifiedProjectScopeResourceConstraint(spec, p.ExecutionIdentifier))
	}
	if spec.NamespaceScopeResourceConstraint != nil {
		composedResourceConstraintList = append(composedResourceConstraintList, composeFullyQualifiedNamespaceScopeResourceConstraint(spec, p.ExecutionIdentifier))
	}
	return composedResourceConstraintList
}

func (p Proxy) AllocateResource(ctx context.Context, namespace pluginCore.ResourceNamespace,
	allocationToken string, constraintsSpec pluginCore.ResourceConstraintsSpec) (pluginCore.AllocationStatus, error) {
	composedResourceConstraintList := p.ComposeResourceConstraint(constraintsSpec)
	status, err := p.BaseResourceManager.AllocateResource(ctx,
		p.ResourceNamespacePrefix.CreateSubNamespace(namespace),
		Token(allocationToken).prepend(ComposeTokenPrefix(p.ExecutionIdentifier)),
		composedResourceConstraintList)
	if err != nil {
		return status, err
	}
	p.ResourcePoolInfo[allocationToken] = &event.ResourcePoolInfo{
		AllocationToken: allocationToken,
		Namespace:       string(namespace),
	}
	return status, err
}

func (p Proxy) ReleaseResource(ctx context.Context, namespace pluginCore.ResourceNamespace,
	allocationToken string) error {
	err := p.BaseResourceManager.ReleaseResource(ctx,
		p.ResourceNamespacePrefix.CreateSubNamespace(namespace),
		Token(allocationToken).prepend(ComposeTokenPrefix(p.ExecutionIdentifier)))
	return err
}

func (p Proxy) GetResourcePoolInfo() []*event.ResourcePoolInfo {
	response := make([]*event.ResourcePoolInfo, 0, len(p.ResourcePoolInfo))
	for _, resourcePoolInfo := range p.ResourcePoolInfo {
		response = append(response, resourcePoolInfo)
	}
	return response
}

func GetTaskResourceManager(r BaseResourceManager, resourceNamespacePrefix pluginCore.ResourceNamespace,
	id *core.TaskExecutionIdentifier) TaskResourceManager {
	return Proxy{
		BaseResourceManager:     r,
		ResourceNamespacePrefix: resourceNamespacePrefix,
		ExecutionIdentifier:     id,
		ResourcePoolInfo:        make(map[string]*event.ResourcePoolInfo),
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
	AllocateResource(ctx context.Context, namespace pluginCore.ResourceNamespace, allocationToken Token, constraints []FullyQualifiedResourceConstraint) (pluginCore.AllocationStatus, error)
	ReleaseResource(ctx context.Context, namespace pluginCore.ResourceNamespace, allocationToken Token) error
}
