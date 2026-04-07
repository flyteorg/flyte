package plugin

import (
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

var _ pluginsCore.SetupContext = &setupContext{}

type setupContext struct {
	kubeClient        pluginsCore.KubeClient
	secretManager     pluginsCore.SecretManager
	resourceRegistrar pluginsCore.ResourceRegistrar
	enqueueOwner      pluginsCore.EnqueueOwner
	enqueueLabels     []string
	ownerKind         string
	metricsScope      promutils.Scope
}

// NewSetupContext creates a SetupContext implementation for the executor.
func NewSetupContext(
	kubeClient pluginsCore.KubeClient,
	secretManager pluginsCore.SecretManager,
	resourceRegistrar pluginsCore.ResourceRegistrar,
	enqueueOwner pluginsCore.EnqueueOwner,
	enqueueLabels []string,
	ownerKind string,
	metricsScope promutils.Scope,
) pluginsCore.SetupContext {
	return &setupContext{
		kubeClient:        kubeClient,
		secretManager:     secretManager,
		resourceRegistrar: resourceRegistrar,
		enqueueOwner:      enqueueOwner,
		enqueueLabels:     enqueueLabels,
		ownerKind:         ownerKind,
		metricsScope:      metricsScope,
	}
}

func (s *setupContext) EnqueueOwner() pluginsCore.EnqueueOwner           { return s.enqueueOwner }
func (s *setupContext) IncludeEnqueueLabels() []string                   { return s.enqueueLabels }
func (s *setupContext) OwnerKind() string                                { return s.ownerKind }
func (s *setupContext) MetricsScope() promutils.Scope                    { return s.metricsScope }
func (s *setupContext) KubeClient() pluginsCore.KubeClient               { return s.kubeClient }
func (s *setupContext) SecretManager() pluginsCore.SecretManager         { return s.secretManager }
func (s *setupContext) ResourceRegistrar() pluginsCore.ResourceRegistrar { return s.resourceRegistrar }
