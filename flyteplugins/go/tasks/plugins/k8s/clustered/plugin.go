package clustered

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
)

const taskType = "clustered-task"

type clusteredResourceHandler struct{}

var _ k8s.Plugin = clusteredResourceHandler{}

func (clusteredResourceHandler) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

func (clusteredResourceHandler) IsTerminal(_ context.Context, resource client.Object) (bool, error) {
	jobSet, ok := resource.(*jobsetv1alpha2.JobSet)
	if !ok {
		return false, fmt.Errorf("unexpected resource type %T", resource)
	}
	for _, cond := range jobSet.Status.Conditions {
		t := jobsetv1alpha2.JobSetConditionType(cond.Type)
		if (t == jobsetv1alpha2.JobSetCompleted || t == jobsetv1alpha2.JobSetFailed) &&
			cond.Status == metav1.ConditionTrue {
			return true, nil
		}
	}
	return false, nil
}

func (clusteredResourceHandler) GetCompletionTime(resource client.Object) (time.Time, error) {
	jobSet, ok := resource.(*jobsetv1alpha2.JobSet)
	if !ok {
		return time.Time{}, fmt.Errorf("unexpected resource type %T", resource)
	}
	for _, cond := range jobSet.Status.Conditions {
		t := jobsetv1alpha2.JobSetConditionType(cond.Type)
		if (t == jobsetv1alpha2.JobSetCompleted || t == jobsetv1alpha2.JobSetFailed) &&
			cond.Status == metav1.ConditionTrue {
			if !cond.LastTransitionTime.IsZero() {
				return cond.LastTransitionTime.Time, nil
			}
		}
	}
	return jobSet.CreationTimestamp.Time, nil
}

func (clusteredResourceHandler) BuildIdentityResource(_ context.Context, taskExecutionMetadata pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	// Name must match what BuildResource derives (see build.go) so the lookup/abort
	// path resolves the same object. buildJobSetName is deterministic from the generated
	// name alone, so no task template / replica count is needed here. The plugin manager's
	// addObjectMetadata leaves a non-empty name untouched.
	return &jobsetv1alpha2.JobSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "JobSet",
			APIVersion: jobsetv1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: buildJobSetName(taskExecutionMetadata.GetTaskExecutionID().GetGeneratedName()),
		},
	}, nil
}

func init() {
	if err := jobsetv1alpha2.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  taskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{taskType},
			ResourceToWatch:     &jobsetv1alpha2.JobSet{},
			Plugin:              clusteredResourceHandler{},
			IsDefault:           false,
		})
}
