package factory

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/array"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/branch"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/dynamic"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/end"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/gate"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/recovery"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/start"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type handlerFactory struct {
	handlers map[v1alpha1.NodeKind]interfaces.NodeHandler

	workflowLauncher        launchplan.Executor
	launchPlanReader        launchplan.Reader
	kubeClient              executors.Client
	kubeClientset           kubernetes.Interface
	catalogClient           catalog.Client
	recoveryClient          recovery.Client
	eventConfig             *config.EventConfig
	literalOffloadingConfig config.LiteralOffloadingConfig
	clusterID               string
	signalClient            service.SignalServiceClient
	scope                   promutils.Scope
}

func (f *handlerFactory) GetHandler(kind v1alpha1.NodeKind) (interfaces.NodeHandler, error) {
	h, ok := f.handlers[kind]
	if !ok {
		return nil, errors.Errorf("Handler not registered for NodeKind [%v]", kind)
	}
	return h, nil
}

func (f *handlerFactory) Setup(ctx context.Context, executor interfaces.Node, setup interfaces.SetupContext) error {
	t, err := task.New(ctx, f.kubeClient, f.kubeClientset, f.catalogClient, f.eventConfig, f.clusterID, f.scope)
	if err != nil {
		return err
	}

	arrayHandler, err := array.New(executor, f.eventConfig, f.literalOffloadingConfig, f.scope)
	if err != nil {
		return err
	}

	f.handlers = map[v1alpha1.NodeKind]interfaces.NodeHandler{
		v1alpha1.NodeKindBranch:   branch.New(executor, f.eventConfig, f.scope),
		v1alpha1.NodeKindTask:     dynamic.New(t, executor, f.launchPlanReader, f.eventConfig, f.scope),
		v1alpha1.NodeKindWorkflow: subworkflow.New(executor, f.workflowLauncher, f.recoveryClient, f.eventConfig, f.scope),
		v1alpha1.NodeKindGate:     gate.New(f.eventConfig, f.signalClient, f.scope),
		v1alpha1.NodeKindArray:    arrayHandler,
		v1alpha1.NodeKindStart:    start.New(),
		v1alpha1.NodeKindEnd:      end.New(),
	}

	for _, v := range f.handlers {
		if err := v.Setup(ctx, setup); err != nil {
			return err
		}
	}
	return nil
}

func NewHandlerFactory(ctx context.Context, workflowLauncher launchplan.Executor, launchPlanReader launchplan.Reader,
	kubeClient executors.Client, kubeClientset kubernetes.Interface, catalogClient catalog.Client, recoveryClient recovery.Client, eventConfig *config.EventConfig,
	literalOffloadingConfig config.LiteralOffloadingConfig,
	clusterID string, signalClient service.SignalServiceClient, scope promutils.Scope) (interfaces.HandlerFactory, error) {

	return &handlerFactory{
		workflowLauncher:        workflowLauncher,
		launchPlanReader:        launchPlanReader,
		kubeClient:              kubeClient,
		kubeClientset:           kubeClientset,
		catalogClient:           catalogClient,
		recoveryClient:          recoveryClient,
		eventConfig:             eventConfig,
		literalOffloadingConfig: literalOffloadingConfig,
		clusterID:               clusterID,
		signalClient:            signalClient,
		scope:                   scope,
	}, nil
}
