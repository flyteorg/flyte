package flytek8s

import (
	v1 "k8s.io/api/core/v1"

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

func (to *pluginTaskOverrides) GetPodTemplate() *core.K8SPod {
	return to.TaskOverrides.GetPodTemplate()
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
