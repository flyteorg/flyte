package flytek8s

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/v1/events"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate mockery -name K8sResourceHandler

type Handler interface {
	Handle(context.Context, runtime.Object) error
}

// Defines an interface that deals with k8s resources. Combined with K8sTaskExecutor, this provides an easier and more
// consistent way to write TaskExecutors that create k8s resources.
type K8sResourceHandler interface {
	// Defines a func to create the full resource object that will be posted to k8s.
	BuildResource(ctx context.Context, taskCtx types.TaskContext, task *core.TaskTemplate, inputs *core.LiteralMap) (K8sResource, error)

	// Defines a func to create a query object (typically just object and type meta portions) that's used to query k8s
	// resources.
	BuildIdentityResource(ctx context.Context, taskCtx types.TaskContext) (K8sResource, error)

	// Analyses the k8s resource and reports the status as TaskPhase.
	GetTaskStatus(ctx context.Context, taskCtx types.TaskContext, resource K8sResource) (types.TaskStatus, *events.TaskEventInfo, error)
}

type K8sResource interface {
	runtime.Object
	metav1.Object
	schema.ObjectKind
}
