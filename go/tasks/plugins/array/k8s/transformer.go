package k8s

import (
	"context"
	"regexp"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/template"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	idlPlugins "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	core2 "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PodKind = "pod"
const primaryContainerKey = "primary_container_name"

var namespaceRegex = regexp.MustCompile("(?i){{.namespace}}(?i)")

type arrayTaskContext struct {
	core.TaskExecutionContext
	arrayInputReader io.InputReader
}

// Overrides the TaskExecutionContext from base and returns a specialized context for Array
func (a *arrayTaskContext) InputReader() io.InputReader {
	return a.arrayInputReader
}

func GetNamespaceForExecution(tCtx core.TaskExecutionContext, namespaceTemplate string) string {

	// Default to parent namespace
	namespace := tCtx.TaskExecutionMetadata().GetNamespace()
	if namespaceTemplate != "" {
		if namespaceRegex.MatchString(namespaceTemplate) {
			namespace = namespaceRegex.ReplaceAllString(namespaceTemplate, namespace)
		} else {
			namespace = namespaceTemplate
		}
	}
	return namespace
}

// Initializes a pod from an array job task template with a K8sPod set as the task target.
// TODO: This should be removed by end of 2021 (it duplicates the pod plugin logic) once we improve array job handling
// and move it to the node level.  See https://github.com/flyteorg/flyte/issues/1131
func buildPodMapTask(task *idlCore.TaskTemplate, metadata core.TaskExecutionMetadata) (v1.Pod, error) {
	if task.GetK8SPod() == nil || task.GetK8SPod().PodSpec == nil {
		return v1.Pod{}, errors.Errorf(errors.BadTaskSpecification, "Missing pod spec for task")
	}
	var podSpec = &v1.PodSpec{}
	err := utils.UnmarshalStructToObj(task.GetK8SPod().PodSpec, &podSpec)
	if err != nil {
		return v1.Pod{}, errors.Errorf(errors.BadTaskSpecification,
			"Unable to unmarshal task custom [%v], Err: [%v]", task.GetCustom(), err.Error())
	}
	primaryContainerName, ok := task.GetConfig()[primaryContainerKey]
	if !ok {
		return v1.Pod{}, errors.Errorf(errors.BadTaskSpecification,
			"invalid TaskSpecification, config missing [%s] key in [%v]", primaryContainerKey, task.GetConfig())
	}

	var pod = v1.Pod{
		Spec: *podSpec,
	}
	if task.GetK8SPod().Metadata != nil {
		if task.GetK8SPod().Metadata.Annotations != nil {
			pod.Annotations = task.GetK8SPod().Metadata.Annotations
		}
		if task.GetK8SPod().Metadata.Labels != nil {
			pod.Labels = task.GetK8SPod().Metadata.Labels
		}
	}
	if len(pod.Annotations) == 0 {
		pod.Annotations = make(map[string]string)
	}
	pod.Annotations[primaryContainerKey] = primaryContainerName

	// Set the restart policy to *not* inherit from the default so that a completed pod doesn't get caught in a
	// CrashLoopBackoff after the initial job completion.
	pod.Spec.RestartPolicy = v1.RestartPolicyNever
	flytek8s.GetServiceAccountNameFromTaskExecutionMetadata(metadata)
	return pod, nil
}

// Note that Name is not set on the result object.
// It's up to the caller to set the Name before creating the object in K8s.
func FlyteArrayJobToK8sPodTemplate(ctx context.Context, tCtx core.TaskExecutionContext, namespaceTemplate string) (
	podTemplate v1.Pod, job *idlPlugins.ArrayJob, err error) {

	// Check that the taskTemplate is valid
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return v1.Pod{}, nil, err
	} else if taskTemplate == nil {
		return v1.Pod{}, nil, errors.Errorf(errors.BadTaskSpecification, "Required value not set, taskTemplate is nil")
	}

	if taskTemplate.GetContainer() == nil && taskTemplate.GetK8SPod() == nil {
		return v1.Pod{}, nil, errors.Errorf(errors.BadTaskSpecification,
			"Required value not set, taskTemplate Container or K8sPod")
	}

	arrTCtx := &arrayTaskContext{
		TaskExecutionContext: tCtx,
		arrayInputReader:     array.GetInputReader(tCtx, taskTemplate),
	}
	var arrayJob *idlPlugins.ArrayJob
	if taskTemplate.GetCustom() != nil {
		arrayJob, err = core2.ToArrayJob(taskTemplate.GetCustom(), taskTemplate.TaskTypeVersion)
		if err != nil {
			return v1.Pod{}, nil, err
		}
	}

	annotations := utils.UnionMaps(config.GetK8sPluginConfig().DefaultAnnotations, tCtx.TaskExecutionMetadata().GetAnnotations())
	labels := utils.UnionMaps(config.GetK8sPluginConfig().DefaultLabels, tCtx.TaskExecutionMetadata().GetLabels())

	var pod = v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       PodKind,
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			// Note that name is missing here
			Namespace:       GetNamespaceForExecution(tCtx, namespaceTemplate),
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{tCtx.TaskExecutionMetadata().GetOwnerReference()},
		},
	}
	if taskTemplate.GetContainer() != nil {
		podSpec, err := flytek8s.ToK8sPodSpec(ctx, arrTCtx)
		if err != nil {
			return v1.Pod{}, nil, err
		}
		pod.Spec = *podSpec
	} else if taskTemplate.GetK8SPod() != nil {
		k8sPod, err := buildPodMapTask(taskTemplate, tCtx.TaskExecutionMetadata())
		if err != nil {
			return v1.Pod{}, nil, err
		}
		pod.Labels = utils.UnionMaps(pod.Labels, k8sPod.Labels)
		pod.Annotations = utils.UnionMaps(pod.Annotations, k8sPod.Annotations)
		pod.Spec = k8sPod.Spec

		containerIndex, err := getTaskContainerIndex(&pod)
		if err != nil {
			return v1.Pod{}, nil, err
		}
		templateParameters := template.Parameters{
			TaskExecMetadata: tCtx.TaskExecutionMetadata(),
			Inputs:           arrTCtx.arrayInputReader,
			OutputPath:       tCtx.OutputWriter(),
			Task:             tCtx.TaskReader(),
		}
		err = flytek8s.AddFlyteCustomizationsToContainer(
			ctx, templateParameters, flytek8s.MergeExistingResources, &pod.Spec.Containers[containerIndex])
		if err != nil {
			return v1.Pod{}, nil, err
		}
	}

	return pod, arrayJob, nil
}
