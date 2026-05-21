package clustered

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	flyteerr "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

func (clusteredResourceHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "unable to fetch task template: %v", err)
	}
	if taskTemplate == nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "nil task template")
	}

	var spec clusteredpb.ClusteredTaskSpec
	if err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &spec); err != nil { //nolint:staticcheck
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid ClusteredTaskSpec: %v", err)
	}

	podSpec, objectMeta, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "failed to build pod spec: %v", err)
	}

	podSpec = applyInterconnect(ctx, spec.GetInterconnect(), podSpec)

	// Find the primary container and splice in the entrypoint + runtime env vars.
	primaryIdx := -1
	for i, c := range podSpec.Containers {
		if c.Name == primaryContainerName {
			primaryIdx = i
			break
		}
	}
	if primaryIdx == -1 {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "primary container %q not found in pod spec", primaryContainerName)
	}

	container := &podSpec.Containers[primaryIdx]

	// Move original command + args into args-only (passed through to torchrun → a0).
	originalArgs := append(container.Command, container.Args...)
	container.Command = []string{"python", "-m", "flyte.distributed._entrypoint"}
	container.Args = originalArgs

	injectTorchRunEnv(container, &spec)

	podSpec.RestartPolicy = corev1.RestartPolicyNever
	if podSpec.Subdomain == "" {
		podSpec.Subdomain = taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	}

	replicas := spec.GetReplicas()
	completionMode := batchv1.IndexedCompletion
	backoffLimit := int32(0)
	jobSpec := batchv1.JobSpec{
		Parallelism:    &replicas,
		Completions:    &replicas,
		CompletionMode: &completionMode,
		BackoffLimit:   &backoffLimit,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: *objectMeta,
			Spec:       *podSpec,
		},
	}

	failurePolicy := buildFailurePolicy(&spec)

	jobSetName := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	enableDNSHostnames := true
	replicatedJobReplicas := int32(1)

	jobSet := &jobsetv1alpha2.JobSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "JobSet",
			APIVersion: jobsetv1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobSetName,
			Namespace: taskCtx.TaskExecutionMetadata().GetNamespace(),
			Labels: map[string]string{
				"flyte.org/execution": taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().GetNodeExecutionId().GetExecutionId().GetName(),
			},
			Annotations: map[string]string{
				"flyte.org/task-type": taskType,
			},
		},
		Spec: jobsetv1alpha2.JobSetSpec{
			Network: &jobsetv1alpha2.Network{
				EnableDNSHostnames: &enableDNSHostnames,
			},
			SuccessPolicy: &jobsetv1alpha2.SuccessPolicy{
				Operator: jobsetv1alpha2.OperatorAll,
			},
			FailurePolicy: failurePolicy,
			ReplicatedJobs: []jobsetv1alpha2.ReplicatedJob{
				{
					Name:     "workers",
					Replicas: replicatedJobReplicas,
					Template: batchv1.JobTemplateSpec{
						Spec: jobSpec,
					},
				},
			},
		},
	}

	if ttl := spec.GetTtlSecondsAfterFinished(); ttl != nil {
		ttlVal := int32(ttl.GetValue())
		jobSet.Spec.TTLSecondsAfterFinished = &ttlVal
	}

	return jobSet, nil
}
