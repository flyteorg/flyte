package clustered

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	flyteerr "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	stdutils "github.com/flyteorg/flyte/v2/flytestdlib/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
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
	if err = stdutils.UnmarshalStructToPb(taskTemplate.GetCustom(), &spec); err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid ClusteredTaskSpec: %v", err)
	}

	if spec.GetReplicas() < 1 {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "replicas must be >= 1, got %d", spec.GetReplicas())
	}
	// generatedNameMaxLength reserves pod-index digits up to maxReplicasForNaming so the
	// derived pod names stay within the 63-char limit. Beyond that the reservation is exceeded
	// and the JobSet webhook would reject the pods, so fail fast with a clear spec error instead.
	if spec.GetReplicas() > maxReplicasForNaming {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "replicas must be <= %d, got %d", maxReplicasForNaming, spec.GetReplicas())
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
	injectStartupTimeoutEnv(container, GetConfig().StartupTimeout.Duration)

	podSpec.RestartPolicy = corev1.RestartPolicyNever
	replicas := spec.GetReplicas()
	// The plugin manager stamps the object name via GetGeneratedNameWith(0,
	// GeneratedNameMaxLength) — sanitized to a DNS-1035 label and bounded so JobSet's
	// derived child pod names stay within the 63-char limit (see generatedNameMaxLength
	// in util.go). Derive the same name here for the headless service / pod subdomain,
	// which must equal the JobSet name for pod DNS to resolve.
	jobSetName, err := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedNameWith(0, generatedNameMaxLength)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "failed to derive JobSet name: %v", err)
	}
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

	if spec.GetFailurePolicy().GetRestartOnHostMaintenance() {
		// Fail the Job with reason PodFailurePolicy when a pod is terminated by an
		// involuntary disruption (DisruptionTarget: drain, eviction API, preemption,
		// taint eviction). The net effect matches backoffLimit=0, but the distinct
		// failure reason lets the JobSet failurePolicy rule (failure.go) restart the
		// set without charging maxRestarts. Requires RestartPolicyNever (set above).
		jobSpec.PodFailurePolicy = &batchv1.PodFailurePolicy{
			Rules: []batchv1.PodFailurePolicyRule{{
				Action: batchv1.PodFailurePolicyActionFailJob,
				OnPodConditions: []batchv1.PodFailurePolicyOnPodConditionsPattern{{
					Type:   corev1.DisruptionTarget,
					Status: corev1.ConditionTrue,
				}},
			}},
		}
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
			Labels: utils.UnionMaps(objectMeta.Labels, map[string]string{
				"flyte.org/execution": sanitizeLabelValue(taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().GetNodeExecutionId().GetExecutionId().GetName()),
			}),
			Annotations: utils.UnionMaps(objectMeta.Annotations, map[string]string{
				"flyte.org/task-type":      taskType,
				primaryContainerAnnotation: primaryContainerName,
			}),
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

	if err := applyKueue(jobSet, &GetConfig().Kueue, userLabelSources(taskTemplate, taskCtx)); err != nil {
		return nil, err
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

// kueueQueueNameLabel is the label Kueue reads to pick the LocalQueue a job is submitted to.
const kueueQueueNameLabel = "kueue.x-k8s.io/queue-name"

// startupTimeoutEnv carries the per-worker startup budget, in whole seconds, to the launcher.
const startupTimeoutEnv = "FLYTE_CLUSTERED_STARTUP_TIMEOUT"

// labelSource is a set of labels a user controls, named for error messages.
type labelSource struct {
	name   string
	labels map[string]string
}

// userLabelSources are the labels a user can set on a task: the task's pod template, a
// pod-template override, and the execution's labels. Platform-owned labels (default labels
// and named base pod templates) are not included.
func userLabelSources(taskTemplate *core.TaskTemplate, taskCtx pluginsCore.TaskExecutionContext) []labelSource {
	meta := taskCtx.TaskExecutionMetadata()
	return []labelSource{
		{name: "the task's pod template", labels: taskTemplate.GetK8SPod().GetMetadata().GetLabels()},
		{name: "the pod template override", labels: meta.GetOverrides().GetPodTemplate().GetMetadata().GetLabels()},
		{name: "the execution labels", labels: meta.GetLabels()},
	}
}

// applyKueue submits the JobSet to the configured Kueue queue: it is created suspended so the
// gang is admitted as a whole, and labelled with the queue. The queue is set by the platform
// only. A task that names a different queue through userSources is rejected rather than
// silently moved, so the user learns that the label has no effect; naming the configured
// queue is accepted. The label is kept off the pod template so pods carry no queue of their
// own. A disabled config leaves the JobSet as is.
func applyKueue(jobSet *jobsetv1alpha2.JobSet, cfg *KueueConfig, userSources []labelSource) error {
	if !cfg.Enabled {
		return nil
	}
	queue := strings.TrimSpace(cfg.QueueName)
	if queue == "" {
		return flyteerr.Errorf(flyteerr.BadTaskSpecification,
			"plugins.clustered.kueue.queue-name must be set when Kueue is enabled")
	}
	if errs := validation.IsValidLabelValue(queue); len(errs) > 0 {
		return flyteerr.Errorf(flyteerr.BadTaskSpecification,
			"invalid plugins.clustered.kueue.queue-name %q: %s", queue, strings.Join(errs, "; "))
	}

	if jobSet.Labels == nil {
		jobSet.Labels = map[string]string{}
	}
	for _, source := range userSources {
		if chosen, ok := source.labels[kueueQueueNameLabel]; ok && chosen != queue {
			return flyteerr.Errorf(flyteerr.BadTaskSpecification,
				"%s set %s=%q, but the Kueue queue is set by the platform (%q); remove the label",
				source.name, kueueQueueNameLabel, chosen, queue)
		}
	}

	jobSet.Labels[kueueQueueNameLabel] = queue
	// on pod template we delete the label
	for i := range jobSet.Spec.ReplicatedJobs {
		delete(jobSet.Spec.ReplicatedJobs[i].Template.Spec.Template.Labels, kueueQueueNameLabel)
	}
	jobSet.Spec.Suspend = ptr.To(true)
	return nil
}

// injectStartupTimeoutEnv tells every worker how long to wait for its peers. A zero budget
// leaves the container untouched so the launcher keeps its default.
func injectStartupTimeoutEnv(container *corev1.Container, budget time.Duration) {
	seconds := int64(budget / time.Second)
	if seconds <= 0 {
		return
	}
	container.Env = upsertEnv(container.Env, []corev1.EnvVar{
		{Name: startupTimeoutEnv, Value: strconv.FormatInt(seconds, 10)},
	})
}
