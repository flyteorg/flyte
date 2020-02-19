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

func (t TokenPrefix) append(s string) string {
	return fmt.Sprintf("%s%s%s", t, tokenNamespaceSeparator, s)
}

func composeExecutionUrn(id *core.TaskExecutionIdentifier) string {
	return execUrnPrefix + execUrnSeparator + id.GetNodeExecutionId().GetExecutionId().GetProject() +
		execUrnSeparator + id.GetNodeExecutionId().GetExecutionId().GetDomain() + execUrnSeparator + id.GetNodeExecutionId().GetExecutionId().GetName()
}

func ComposeTokenPrefix(id *core.TaskExecutionIdentifier) TokenPrefix {
	execUrn := composeExecutionUrn(id) // This is for the ease of debugging. Doesn't necessarily need to have this
	return TokenPrefix(execUrn)
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
	BuildResourceManager(ctx context.Context) (pluginCore.ResourceManager, error)
}

// A proxy will be created for each TaskExecutionContext.
// The Proxy of an execution contains the resource namespace prefix (e.g., "qubole-hive-executor" for a hive task)
// and the token prefix (e.g., ex:<project>:<domain>:<exec_id>) of that execution.
// The plugins will only have access to a Proxy but not directly the underlying resource manager.
// The Proxy will prepend proper prefixes for the resource namespace and the allocation token.
type Proxy struct {
	pluginCore.ResourceManager
	ResourceNamespacePrefix pluginCore.ResourceNamespace
	TokenPrefix             TokenPrefix
}

func (p Proxy) AllocateResource(ctx context.Context, namespace pluginCore.ResourceNamespace,
	allocationToken string) (pluginCore.AllocationStatus, error) {
	status, err := p.ResourceManager.AllocateResource(ctx,
		p.ResourceNamespacePrefix.CreateSubNamespace(namespace),
		p.TokenPrefix.append(allocationToken))
	return status, err
}

func (p Proxy) ReleaseResource(ctx context.Context, namespace pluginCore.ResourceNamespace,
	allocationToken string) error {
	err := p.ResourceManager.ReleaseResource(ctx,
		p.ResourceNamespacePrefix.CreateSubNamespace(namespace),
		p.TokenPrefix.append(allocationToken))
	return err
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
