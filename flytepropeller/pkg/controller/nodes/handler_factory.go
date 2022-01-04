package nodes

import (
	"context"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/recovery"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/dynamic"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/pkg/errors"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/branch"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/end"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/start"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task"
)

//go:generate mockery -name HandlerFactory -case=underscore

type HandlerFactory interface {
	GetHandler(kind v1alpha1.NodeKind) (handler.Node, error)
	Setup(ctx context.Context, setup handler.SetupContext) error
}

type handlerFactory struct {
	handlers map[v1alpha1.NodeKind]handler.Node
}

func (f handlerFactory) GetHandler(kind v1alpha1.NodeKind) (handler.Node, error) {
	h, ok := f.handlers[kind]
	if !ok {
		return nil, errors.Errorf("Handler not registered for NodeKind [%v]", kind)
	}
	return h, nil
}

func (f handlerFactory) Setup(ctx context.Context, setup handler.SetupContext) error {
	for _, v := range f.handlers {
		if err := v.Setup(ctx, setup); err != nil {
			return err
		}
	}
	return nil
}

func NewHandlerFactory(ctx context.Context, executor executors.Node, workflowLauncher launchplan.Executor,
	launchPlanReader launchplan.Reader, kubeClient executors.Client, client catalog.Client, recoveryClient recovery.Client,
	eventConfig *config.EventConfig, clusterID string, scope promutils.Scope) (HandlerFactory, error) {

	t, err := task.New(ctx, kubeClient, client, eventConfig, clusterID, scope)
	if err != nil {
		return nil, err
	}

	f := &handlerFactory{
		handlers: map[v1alpha1.NodeKind]handler.Node{
			v1alpha1.NodeKindBranch:   branch.New(executor, eventConfig, scope),
			v1alpha1.NodeKindTask:     dynamic.New(t, executor, launchPlanReader, eventConfig, scope),
			v1alpha1.NodeKindWorkflow: subworkflow.New(executor, workflowLauncher, recoveryClient, eventConfig, scope),
			v1alpha1.NodeKindStart:    start.New(),
			v1alpha1.NodeKindEnd:      end.New(),
		},
	}

	return f, nil
}
