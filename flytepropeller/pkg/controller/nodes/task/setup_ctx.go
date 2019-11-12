package task

import (
	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flytestdlib/promutils"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

type setupContext struct {
	handler.SetupContext
	kubeClient    pluginCore.KubeClient
	secretManager pluginCore.SecretManager
}

func (s setupContext) SecretManager() pluginCore.SecretManager {
	return s.secretManager
}

func (s setupContext) MetricsScope() promutils.Scope {
	return s.SetupContext.MetricsScope()
}

func (s setupContext) KubeClient() pluginCore.KubeClient {
	return s.kubeClient
}

func (s setupContext) EnqueueOwner() pluginCore.EnqueueOwner {
	return func(ownerId types.NamespacedName) error {
		s.SetupContext.EnqueueOwner()(ownerId.String())
		return nil
	}
}

func (t *Handler) newSetupContext(sCtx handler.SetupContext) *setupContext {

	return &setupContext{
		SetupContext:  sCtx,
		kubeClient:    t.kubeClient,
		secretManager: t.secretManager,
	}
}

type nameSpacedSetupCtx struct {
	*setupContext
	rn pluginCore.ResourceRegistrar
}

func (n nameSpacedSetupCtx) ResourceRegistrar() pluginCore.ResourceRegistrar {
	return n.rn
}

func newNameSpacedSetupCtx(sCtx *setupContext, rn pluginCore.ResourceRegistrar) nameSpacedSetupCtx {
	return nameSpacedSetupCtx{
		setupContext: sCtx,
		rn:           rn,
	}
}
