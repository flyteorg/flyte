package core

import (
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// When a change is observed, the owning entity can be triggered for re-validation
type EnqueueOwner func(labels map[string]string) error

// Passed to the Loader function when setting up a plugin
type SetupContext interface {
	// returns a callback mechanism that indicates that (workflow, task) is ready to be re-evaluated
	EnqueueOwner() EnqueueOwner
	// returns a list of labels that should be used to enqueue the owner
	IncludeEnqueueLabels() []string
	// provides a k8s specific owner kind
	OwnerKind() string
	// a metrics scope to publish stats under
	MetricsScope() promutils.Scope
	// A kubernetes client to the bound cluster
	KubeClient() KubeClient
	// Returns a secret manager that can retrieve configured secrets for this plugin
	SecretManager() SecretManager
	// Returns a resource negotiator that the plugin can register resource quota against
	ResourceRegistrar() ResourceRegistrar
}
