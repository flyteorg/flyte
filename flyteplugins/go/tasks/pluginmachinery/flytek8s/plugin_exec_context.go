package flytek8s

import (
	"context"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
)

type pluginTaskOverrides struct {
	pluginsCore.TaskOverrides
	resources         *v1.ResourceRequirements
	extendedResources *core.ExtendedResources
}

func (to *pluginTaskOverrides) GetResources() *v1.ResourceRequirements {
	if to.resources != nil {
		return to.resources
	}
	return to.TaskOverrides.GetResources()
}

func (to *pluginTaskOverrides) GetExtendedResources() *core.ExtendedResources {
	if to.extendedResources != nil {
		return to.extendedResources
	}
	return to.TaskOverrides.GetExtendedResources()
}

func (to *pluginTaskOverrides) GetContainerImage() string {
	return to.TaskOverrides.GetContainerImage()
}

type pluginTaskExecutionMetadata struct {
	pluginsCore.TaskExecutionMetadata
	interruptible *bool
	overrides     *pluginTaskOverrides
}

func (tm *pluginTaskExecutionMetadata) IsInterruptible() bool {
	if tm.interruptible != nil {
		return *tm.interruptible
	}
	return tm.TaskExecutionMetadata.IsInterruptible()
}

func (tm *pluginTaskExecutionMetadata) GetOverrides() pluginsCore.TaskOverrides {
	if tm.overrides != nil {
		return tm.overrides
	}
	return tm.TaskExecutionMetadata.GetOverrides()
}

type pluginTaskExecutionContext struct {
	pluginsCore.TaskExecutionContext
	metadata *pluginTaskExecutionMetadata
}

func (tc *pluginTaskExecutionContext) TaskExecutionMetadata() pluginsCore.TaskExecutionMetadata {
	if tc.metadata != nil {
		return tc.metadata
	}
	return tc.TaskExecutionContext.TaskExecutionMetadata()
}

type PluginTaskExecutionContextOption func(*pluginTaskExecutionContext)

func WithInterruptible(v bool) PluginTaskExecutionContextOption {
	return func(tc *pluginTaskExecutionContext) {
		if tc.metadata == nil {
			tc.metadata = &pluginTaskExecutionMetadata{
				TaskExecutionMetadata: tc.TaskExecutionContext.TaskExecutionMetadata(),
			}
		}
		tc.metadata.interruptible = &v
	}
}

func WithResources(r *v1.ResourceRequirements) PluginTaskExecutionContextOption {
	return func(tc *pluginTaskExecutionContext) {
		if tc.metadata == nil {
			tc.metadata = &pluginTaskExecutionMetadata{
				TaskExecutionMetadata: tc.TaskExecutionContext.TaskExecutionMetadata(),
			}
		}
		if tc.metadata.overrides == nil {
			tc.metadata.overrides = &pluginTaskOverrides{
				TaskOverrides: tc.metadata.TaskExecutionMetadata.GetOverrides(),
			}
		}
		tc.metadata.overrides.resources = r
	}
}

func WithExtendedResources(er *core.ExtendedResources) PluginTaskExecutionContextOption {
	return func(tc *pluginTaskExecutionContext) {
		if tc.metadata == nil {
			tc.metadata = &pluginTaskExecutionMetadata{
				TaskExecutionMetadata: tc.TaskExecutionContext.TaskExecutionMetadata(),
			}
		}
		if tc.metadata.overrides == nil {
			tc.metadata.overrides = &pluginTaskOverrides{
				TaskOverrides: tc.metadata.TaskExecutionMetadata.GetOverrides(),
			}
		}
		tc.metadata.overrides.extendedResources = er
	}
}

func NewPluginTaskExecutionContext(tc pluginsCore.TaskExecutionContext, options ...PluginTaskExecutionContextOption) pluginsCore.TaskExecutionContext {
	tm := tc.TaskExecutionMetadata()
	to := tm.GetOverrides()
	ctx := &pluginTaskExecutionContext{
		TaskExecutionContext: tc,
		metadata: &pluginTaskExecutionMetadata{
			TaskExecutionMetadata: tm,
			overrides: &pluginTaskOverrides{
				TaskOverrides: to,
			},
		},
	}
	for _, o := range options {
		o(ctx)
	}
	return ctx
}

type NodeExecutionK8sReader struct {
	namespace   string
	executionID string
	nodeID      string
	client      client.Reader
}

func NewNodeExecutionK8sReader(meta pluginsCore.TaskExecutionMetadata, client pluginsCore.KubeClient) *NodeExecutionK8sReader {
	tID := meta.GetTaskExecutionID().GetID()
	executionID := tID.GetNodeExecutionId().GetExecutionId().GetName()
	nodeID := tID.GetNodeExecutionId().GetNodeId()
	namespace := meta.GetNamespace()
	return &NodeExecutionK8sReader{
		namespace:   namespace,
		executionID: executionID,
		nodeID:      nodeID,
		client:      client.GetCache(),
	}
}

func (n NodeExecutionK8sReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	key.Namespace = n.namespace
	err := n.client.Get(ctx, key, obj, opts...)
	if err != nil {
		return err
	}

	if obj.GetLabels()["node-id"] != n.nodeID || obj.GetLabels()["execution-id"] != n.executionID {
		// reset obj to default value, simulate not found
		p := reflect.ValueOf(obj).Elem()
		p.Set(reflect.Zero(p.Type()))
		kind := obj.GetObjectKind().GroupVersionKind()
		return errors.NewNotFound(schema.GroupResource{Group: kind.Group, Resource: kind.Kind}, obj.GetName())
	}
	return nil
}

func (n NodeExecutionK8sReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	opts = append(opts, client.InNamespace(n.namespace), client.MatchingLabels{
		"execution-id": n.executionID,
		"node-id":      n.nodeID,
	})
	return n.client.List(ctx, list, opts...)
}
