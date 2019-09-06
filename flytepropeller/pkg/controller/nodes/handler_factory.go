package nodes

import (
	"context"
	"time"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/dynamic"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/catalog"
	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/branch"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/end"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/start"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task"
	"github.com/lyft/flytestdlib/storage"
	"github.com/pkg/errors"
)

//go:generate mockery -name HandlerFactory

type HandlerFactory interface {
	GetHandler(kind v1alpha1.NodeKind) (handler.IFace, error)
}

type handlerFactory struct {
	handlers map[v1alpha1.NodeKind]handler.IFace
}

func (f handlerFactory) GetHandler(kind v1alpha1.NodeKind) (handler.IFace, error) {
	h, ok := f.handlers[kind]
	if !ok {
		return nil, errors.Errorf("Handler not registered for NodeKind [%v]", kind)
	}
	return h, nil
}

func NewHandlerFactory(ctx context.Context,
	executor executors.Node,
	eventSink events.EventSink,
	workflowLauncher launchplan.Executor,
	enQWorkflow v1alpha1.EnqueueWorkflow,
	revalPeriod time.Duration,
	store *storage.DataStore,
	catalogClient catalog.Client,
	kubeClient executors.Client,
	scope promutils.Scope,
) (HandlerFactory, error) {

	f := &handlerFactory{
		handlers: map[v1alpha1.NodeKind]handler.IFace{
			v1alpha1.NodeKindBranch: branch.New(executor, eventSink, scope),
			v1alpha1.NodeKindTask: dynamic.New(
				task.New(eventSink, store, enQWorkflow, revalPeriod, catalogClient, kubeClient, scope),
				executor,
				enQWorkflow,
				store,
				scope),
			v1alpha1.NodeKindWorkflow: subworkflow.New(executor, eventSink, workflowLauncher, enQWorkflow, store, scope),
			v1alpha1.NodeKindStart:    start.New(store),
			v1alpha1.NodeKindEnd:      end.New(store),
		},
	}
	for _, v := range f.handlers {
		if err := v.Initialize(ctx); err != nil {
			return nil, err
		}
	}
	return f, nil
}
