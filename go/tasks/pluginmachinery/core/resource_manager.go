package core

import (
	"context"
)

type AllocationStatus string

const (
	// This is the enum returned when there's an error
	AllocationUndefined AllocationStatus = "ResourceGranted"

	// Go for it
	AllocationStatusGranted AllocationStatus = "ResourceGranted"

	// This means that no resources are available globally.  This is the only rejection message we use right now.
	AllocationStatusExhausted AllocationStatus = "ResourceExhausted"

	// We're not currently using this - but this would indicate that things globally are okay, but that your
	// own namespace is too busy
	AllocationStatusNamespaceQuotaExceeded AllocationStatus = "NamespaceQuotaExceeded"
)

const namespaceSeparator = ":"

type ResourceNamespace string

func (r ResourceNamespace) CreateSubNamespace(namespace ResourceNamespace) ResourceNamespace {
	return r + namespaceSeparator + namespace
}

type ResourceRegistrar interface {
	RegisterResourceQuota(ctx context.Context, namespace ResourceNamespace, quota int) error
}

// Resource Manager manages type, and each allocation is of size one
// Flyte resource manager manages resources that Flyte components use. Generally speaking, a resource is an abstraction of anything that has a limited quota
// of units and can be claimed in a single unit or multiple units at once. At Flyte's current state, a resource means a logical separation (e.g., a cluster)
// of an external service that allows a limited number of outstanding requests to be sent to. Flyte uses a token to serve as the placeholder to represent a
// unit of resource. Flyte resource manager manages resources by managing the tokens of the resources. For example, Flyte has a Qubole plugin that allows Flyte
// tasks to send out Hive commands to Qubole. Here, a single Qubole cluster is a resource, and sending out a single Hive command to a Qubole cluster consumes a
// token of the corresponding resource.
//
// At the setup time of Flyte Propeller, a Flyte plugin needs to register the resources and the desired quota of each resource with Flyte resource manager. During
// runtime, Flyte resource manager does two simple things: allocating tokens and releasing tokens. When a Flyte task execution wants to send a  request to an external
// service, first it needs to claim a unit of the corresponding resource. Specifically, an execution needs to generate a unique token, and use the token to call the
// resource manager's AllocateResource() function. Flyte resource manager will check its current utilization and the allocation policy to decide whether or not to grant
// the request. Only when receiving the "Granted" return value shall this execution move forward and send out the request. The granted token will be recorded in a token pool
// corresponding to the resource and managed by the resource manager. When the request is done, the plugin will ask the resource manager to release the token, and the token
// will be erased from the corresponding pool.
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

// ResourceConstraintsSpec is a contract that a plugin can specify, to force quota-acquisition constraints at different levels.
// Setting constraints in a ResourceConstraintsSpec to nil objects is valid, meaning there's no constraint at the corresponding level.
// For example, a ResourceConstraintsSpec with nil ProjectScopeResourceConstraint and a non-nil NamespaceScopeResourceConstraint means
// that it only poses a cap at the namespace level. A zero-value ResourceConstraintsSpec means there's no constraints posed at any level.
type ResourceConstraintsSpec struct {
	ProjectScopeResourceConstraint *ResourceConstraint
	NamespaceScopeResourceConstraint *ResourceConstraint
}

