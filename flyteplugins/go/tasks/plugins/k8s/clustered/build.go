package clustered

import (
	"context"
	"math"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	flyteerr "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
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

	if spec.GetReplicas() < 1 {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "replicas must be >= 1, got %d", spec.GetReplicas())
	}
	if spec.GetNprocPerNode() < 1 {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "nproc_per_node must be >= 1, got %d", spec.GetNprocPerNode())
	}

	podSpec, objectMeta, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "failed to build pod spec: %v", err)
	}

	podSpec = applyInterconnect(ctx, spec.GetInterconnect(), podSpec)

	// Propagate the node-execution labels/annotations onto the pod template. The plugin
	// manager's addObjectMetadata only stamps these (incl. execution-id/node-id) on the
	// top-level JobSet, and the JobSet controller does not copy arbitrary parent labels
	// down to child pods. Without this, child pods lack execution-id/node-id and the
	// node-execution-scoped K8sReader.List in getLogContext returns nothing, so no
	// LogContext reaches the UI. Mirrors ray's buildWorkerPodTemplate.
	cfg := config.GetK8sPluginConfig()
	objectMeta.Labels = utils.UnionMaps(cfg.DefaultLabels, objectMeta.Labels,
		utils.CopyMap(taskCtx.TaskExecutionMetadata().GetLabels()))
	objectMeta.Annotations = utils.UnionMaps(cfg.DefaultAnnotations, objectMeta.Annotations,
		utils.CopyMap(taskCtx.TaskExecutionMetadata().GetAnnotations()))

	// The SDK is responsible for setting container.Command to the entrypoint module
	// (python -m flyte.distributed._entrypoint) at serde time. The plugin stays
	// module-path-agnostic so SDK renames do not require a backend release.
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

	injectTorchRunEnv(container, &spec)

	podSpec.RestartPolicy = corev1.RestartPolicyNever
	replicas := spec.GetReplicas()
	// JobSet name doubles as the headless service / pod subdomain; both must be
	// RFC 1123 subdomain-compatible. GeneratedName isn't guaranteed to be, and for
	// composed/nested tasks it can be long enough that JobSet's derived pod names
	// exceed the 63-char limit and the webhook rejects them. buildJobSetName both
	// sanitizes and bounds the name so the longest generated pod name stays valid.
	// BuildIdentityResource must derive the same name (see plugin.go) so the lookup
	// and abort paths resolve the object this create path produces.
	jobSetName := buildJobSetName(taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())
	if podSpec.Subdomain == "" {
		podSpec.Subdomain = jobSetName
	}

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

	failurePolicy, err := buildFailurePolicy(&spec)
	if err != nil {
		return nil, err
	}

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
				"flyte.org/execution": sanitizeLabelValue(taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().GetNodeExecutionId().GetExecutionId().GetName()),
			},
			Annotations: map[string]string{
				"flyte.org/task-type":      taskType,
				primaryContainerAnnotation: primaryContainerName,
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
					Name:     workersReplicatedJobName,
					Replicas: replicatedJobReplicas,
					Template: batchv1.JobTemplateSpec{
						Spec: jobSpec,
					},
				},
			},
		},
	}

	if ttl := spec.GetTtlSecondsAfterFinished(); ttl != nil {
		v := ttl.GetValue()
		if v > math.MaxInt32 {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "ttl_seconds_after_finished %d exceeds maximum allowed value", v)
		}
		ttlVal := int32(v)
		jobSet.Spec.TTLSecondsAfterFinished = &ttlVal
	}

	return jobSet, nil
}
