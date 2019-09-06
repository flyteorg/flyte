package k8s

import (
	"math"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

func refInt(i int) *int {
	return &i
}

func refStr(s string) *string {
	return &s
}

func computeRetryStrategy(n *core.Node, t *core.TaskTemplate) *v1alpha1.RetryStrategy {
	if n.GetMetadata() != nil && n.GetMetadata().GetRetries() != nil {
		return &v1alpha1.RetryStrategy{
			MinAttempts: refInt(int(n.GetMetadata().GetRetries().Retries + 1)),
		}
	}

	if t != nil && t.GetMetadata() != nil && t.GetMetadata().GetRetries() != nil {
		return &v1alpha1.RetryStrategy{
			MinAttempts: refInt(int(t.GetMetadata().GetRetries().Retries + 1)),
		}
	}

	return nil
}

func computeActiveDeadlineSeconds(n *core.Node, t *core.TaskTemplate) *int64 {
	if n.GetMetadata() != nil && n.GetMetadata().Timeout != nil {
		return &n.GetMetadata().Timeout.Seconds
	}

	if t != nil && t.GetMetadata() != nil && t.GetMetadata().Timeout != nil {
		return &t.GetMetadata().Timeout.Seconds
	}

	return nil
}

func getResources(task *core.TaskTemplate) *core.Resources {
	if task == nil {
		return nil
	}

	if task.GetContainer() == nil {
		return nil
	}

	return task.GetContainer().Resources
}

func toAliasValueArray(aliases []*core.Alias) []v1alpha1.Alias {
	if aliases == nil {
		return nil
	}

	res := make([]v1alpha1.Alias, 0, len(aliases))
	for _, alias := range aliases {
		res = append(res, v1alpha1.Alias{Alias: *alias})
	}

	return res
}

func toBindingValueArray(bindings []*core.Binding) []*v1alpha1.Binding {
	if bindings == nil {
		return nil
	}

	res := make([]*v1alpha1.Binding, 0, len(bindings))
	for _, binding := range bindings {
		res = append(res, &v1alpha1.Binding{Binding: binding})
	}

	return res
}

func minInt(i, j int) int {
	return int(math.Min(float64(i), float64(j)))
}
