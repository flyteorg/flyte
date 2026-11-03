package clustered

import (
	"context"
	"fmt"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
	jobsetconstants "sigs.k8s.io/jobset/pkg/constants"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
)

const taskType = "clustered-task"

type clusteredResourceHandler struct{}

var _ k8s.Plugin = clusteredResourceHandler{}

// The JobSet's child pods are where a node daemon records what the hardware did, so the
// framework has to be able to find them from the JobSet this plugin tracks.
var _ k8s.ChildPodDiscovery = clusteredResourceHandler{}

func (clusteredResourceHandler) GetProperties() k8s.PluginProperties {
	// The plugin manager consumes this to stamp the JobSet name via
	// GetGeneratedNameWith(0, GeneratedNameMaxLength), bounding it so derived child
	// pod names fit the 63-char limit. See generatedNameMaxLength in util.go.
	return k8s.PluginProperties{GeneratedNameMaxLength: &generatedNameMaxLength}
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

// ChildPods implements k8s.ChildPodDiscovery. The pods that run the task are the ones the
// JobSet controller expands from the templates build.go put in the JobSet, which the
// framework tracks nothing of, since it tracks the JobSet.
//
// The selector is the attempt's own labels, which build.go merges onto every child pod
// template, narrowed by the JobSet the pods belong to and by the restart generation the
// pods were created in.
//
// The generation matters because a JobSet restart recreates every child pod under the same
// JobSet name, so name alone would pool the pods of every generation this attempt has been
// through and let a fault from a long-dead generation explain the failure of the current
// one. The controller stamps each generation's pods with its own restart-attempt number,
// and the JobSet's Status.Restarts is that number for the generation now running: the
// restart budget being exhausted is what ends the attempt, and it does not increment on the
// way out. While a restart is in flight the two disagree for an instant, and the selector
// then matches nothing rather than the wrong generation, which is the direction to fail in.
func (clusteredResourceHandler) ChildPods(
	_ context.Context,
	taskCtx pluginsCore.TaskExecutionMetadata,
	resource client.Object,
) (labels.Selector, error) {
	jobSet, ok := resource.(*jobsetv1alpha2.JobSet)
	if !ok {
		return nil, fmt.Errorf("unexpected resource type %T", resource)
	}

	selector := flytek8s.AttemptPodSelector(taskCtx)
	if selector == nil {
		return nil, nil
	}

	name, err := labels.NewRequirement(jobsetv1alpha2.JobSetNameKey, selection.Equals, []string{jobSet.Name})
	if err != nil {
		return nil, err
	}

	generation, err := labels.NewRequirement(
		jobsetconstants.RestartsKey, selection.Equals, []string{strconv.Itoa(int(jobSet.Status.Restarts))})
	if err != nil {
		return nil, err
	}

	return selector.Add(*name, *generation), nil
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

func (clusteredResourceHandler) BuildIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	// No name is set here: the plugin manager's addObjectMetadata stamps the object name
	// via GetGeneratedNameWith(0, GeneratedNameMaxLength) on both the create and lookup
	// paths, so both resolve the same object without the plugin naming it.
	return &jobsetv1alpha2.JobSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "JobSet",
			APIVersion: jobsetv1alpha2.SchemeGroupVersion.String(),
		},
	}, nil
}

func init() {
	if err := jobsetv1alpha2.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	pluginmachinery.PluginRegistry().RegisterScheme(taskType, jobsetv1alpha2.AddToScheme)

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  taskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{taskType},
			ResourceToWatch:     &jobsetv1alpha2.JobSet{},
			Plugin:              clusteredResourceHandler{},
			IsDefault:           false,
		})
}
