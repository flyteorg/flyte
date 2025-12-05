package core

import (
	"context"
)

//go:generate enumer -type=AllocationStatus -trimprefix=AllocationStatus

type AllocationStatus int

const (
	// This is the enum returned when there's an error
	AllocationUndefined AllocationStatus = iota

	// Go for it
	AllocationStatusGranted

	// This means that no resources are available globally.  This is the only rejection message we use right now.
	AllocationStatusExhausted

	// We're not currently using this - but this would indicate that things globally are okay, but that your
	// own namespace is too busy
	AllocationStatusNamespaceQuotaExceeded
)

const namespaceSeparator = ":"

type ResourceNamespace string

func (r ResourceNamespace) CreateSubNamespace(namespace ResourceNamespace) ResourceNamespace {
	return r + namespaceSeparator + namespace
}

type ResourceRegistrar interface {
	RegisterResourceQuota(ctx context.Context, namespace ResourceNamespace, quota int) error
}

// ResourceManager Interface
// 1. Terms and definitions
//
//   - Resource: resource is an abstraction of anything that has a limited quota of units and can be claimed in a
//     single unit or multiple units at once. At Flyte's current state, a resource means a logical
//     separation (e.g., a cluster) of an external service that allows a limited number of outstanding
//     requests to be sent to.
//
//   - Token: Flyte uses a token to serve as the placeholder to represent a unit of resource. Flyte resource manager
//     manages resources by managing the tokens of the resources.
//
//     2. Description
//     ResourceManager provides a task-type-specific pooling system for Flyte Tasks. Plugin writers can optionally
//     request for resources in their tasks, in single quantity.
//
//     3. Usage
//     A Flyte plugin registers the resources and the desired quota of each resource with ResourceRegistrar at the
//     setup time of Flyte Propeller. At the end of the setup time, Flyte Propeller builds a ResourceManager based on
//     these registration requests.
//
//     During runtime, the ResourceManager does two simple things: allocating tokens and releasing tokens. When a Flyte
//     task execution wants to send a request to an external service, the plugin should claim a unit of the corresponding
//     resource. Specifically, an execution needs to generate a unique token, and register the token with ResourceManager
//     by calling ResourceManager's AllocateResource() function. ResourceManager will check its current utilization and
//     the allocation policy to decide whether or not to grant the request. Only when receiving the "AllocationGranted"
//     status shall this execution move forward and send out the request. The granted token will be recorded in a token
//     pool corresponding to the resource and managed by ResourceManager. When the request is done, the plugin will ask
//     the resource manager to release the token by calling ResourceManager's ReleaseResource(), and the token will be
//     erased from the corresponding pool.
//
//     4. Example
//     Flyte has a built-on Qubole plugin that allows Flyte tasks to send out Hive commands to Qubole.
//     In the plugin, a single Qubole cluster is a resource, and sending out a single Hive command to a Qubole cluster consumes
//     a token of the corresponding resource. The resource allocation is achieved by the Qubole plugin calling
//     status, err := AllocateResource(ctx, <cluster name>, <token string>, <constraint spec>)
//     and the de-allocation is achieved by the plugin calling
//     status, err := AllocateResource(ctx, <cluster name>, <token string>, <constraint spec>)
//
//     For example,
//     status, err := AllocateResource(ctx, "default_cluster", "flkgiwd13-akjdoe-0", ResourceConstraintsSpec{})
//     When the corresponding Hive command finishes, the plugin needs to make the following function call to release
//     the corresponding token
//     err := ReleaseResource(ctx, "default_cluster", "flkgiwd13-akjdoe-0")
type ResourceManager interface {
	GetID() string
	// During execution time, plugins can call AllocateResource() to register a token to the token pool associated with a resource with the resource manager.
	// If granted an allocation, the token will be recorded in the corresponding token pool until the same plugin releases it.
	// When calling AllocateResource, the plugin needs to specify a ResourceConstraintsSpec which contains resource capping constraints at different levels.
	// The ResourceConstraint pointers in ResourceConstraintsSpec, however, can be set to nil to present a non-constraint at that level
	AllocateResource(ctx context.Context, namespace ResourceNamespace, allocationToken string, constraintsSpec ResourceConstraintsSpec) (AllocationStatus, error)
	// During execution time, after an outstanding request is completed, the plugin need to use ReleaseResource() to release the allocation of the corresponding token
	// from the token pool in order to gain back the quota taken by the token
	ReleaseResource(ctx context.Context, namespace ResourceNamespace, allocationToken string) error
}

type ResourceConstraint struct {
	Value int64
}

// ResourceConstraintsSpec is a contract that a plugin can specify with ResourceManager to force runtime quota-allocation constraints
// at different levels.
//
// Setting constraints in a ResourceConstraintsSpec to nil objects is valid, meaning there's no constraint at the corresponding level.
// For example, a ResourceConstraintsSpec with nil ProjectScopeResourceConstraint and a non-nil NamespaceScopeResourceConstraint means
// that it only poses a cap at the namespace level. A zero-value ResourceConstraintsSpec means there's no constraints posed at any level.
type ResourceConstraintsSpec struct {
	ProjectScopeResourceConstraint   *ResourceConstraint
	NamespaceScopeResourceConstraint *ResourceConstraint
}
