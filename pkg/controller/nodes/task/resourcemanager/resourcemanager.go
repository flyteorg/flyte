package resourcemanager

import (
	"context"
	"fmt"
	"sync"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flytestdlib/promutils"
)

//go:generate mockery -name ResourceManager -case=underscore

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
	ResourceRegistrar(namespacePrefix pluginCore.ResourceNamespace) pluginCore.ResourceRegistrar
	BuildResourceManager(ctx context.Context) (pluginCore.ResourceManager, error)
}

// A proxy will be created for each TaskExecutionContext
type Proxy struct {
	pluginCore.ResourceManager
	ResourceNamespacePrefix pluginCore.ResourceNamespace
	TokenPrefix             TokenPrefix
}

func (p Proxy) getPrefixedNamespace(namespace pluginCore.ResourceNamespace) pluginCore.ResourceNamespace {
	return p.ResourceNamespacePrefix.CreateSubNamespace(namespace)
}

func (p Proxy) AllocateResource(ctx context.Context, namespace pluginCore.ResourceNamespace,
	allocationToken string) (pluginCore.AllocationStatus, error) {

	namespacedAllocationToken := p.TokenPrefix.append(allocationToken)
	status, err := p.ResourceManager.AllocateResource(ctx, p.getPrefixedNamespace(namespace), namespacedAllocationToken)
	return status, err
}

func (p Proxy) ReleaseResource(ctx context.Context, namespace pluginCore.ResourceNamespace,
	allocationToken string) error {
	err := p.ResourceManager.ReleaseResource(ctx, p.getPrefixedNamespace(namespace), allocationToken)
	return err
}

type ResourceRegistrarProxy struct {
	pluginCore.ResourceRegistrar
	ResourceNamespacePrefix pluginCore.ResourceNamespace
	TokenPrefix             TokenPrefix
}

func (p ResourceRegistrarProxy) getPrefixedNamespace(namespace pluginCore.ResourceNamespace) pluginCore.ResourceNamespace {
	return p.ResourceNamespacePrefix.CreateSubNamespace(namespace)
}

func (p ResourceRegistrarProxy) RegisterResourceQuota(ctx context.Context, namespace pluginCore.ResourceNamespace, quota int) error {
	return p.ResourceRegistrar.RegisterResourceQuota(ctx, p.getPrefixedNamespace(namespace), quota)
}
