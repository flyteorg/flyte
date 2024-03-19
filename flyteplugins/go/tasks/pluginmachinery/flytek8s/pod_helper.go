package flytek8s

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/imdario/mergo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginserrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const PodKind = "pod"
const OOMKilled = "OOMKilled"
const Interrupted = "Interrupted"
const PrimaryContainerNotFound = "PrimaryContainerNotFound"
const SIGKILL = 137

const defaultContainerTemplateName = "default"
const primaryContainerTemplateName = "primary"
const PrimaryContainerKey = "primary_container_name"

// AddRequiredNodeSelectorRequirements adds the provided v1.NodeSelectorRequirement
// objects to an existing v1.Affinity object. If there are no existing required
// node selectors, the new v1.NodeSelectorRequirement will be added as-is.
// However, if there are existing required node selectors, we iterate over all existing
// node selector terms and append the node selector requirement. Note that multiple node
// selector terms are OR'd, and match expressions within a single node selector term
// are AND'd during scheduling.
// See: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity
func AddRequiredNodeSelectorRequirements(base *v1.Affinity, new ...v1.NodeSelectorRequirement) {
	if base.NodeAffinity == nil {
		base.NodeAffinity = &v1.NodeAffinity{}
	}
	if base.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		base.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{}
	}
	if len(base.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) > 0 {
		nodeSelectorTerms := base.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		for i := range nodeSelectorTerms {
			nst := &nodeSelectorTerms[i]
			nst.MatchExpressions = append(nst.MatchExpressions, new...)
		}
	} else {
		base.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []v1.NodeSelectorTerm{v1.NodeSelectorTerm{MatchExpressions: new}}
	}
}

// AddPreferredNodeSelectorRequirements appends the provided v1.NodeSelectorRequirement
// objects to an existing v1.Affinity object's list of preferred scheduling terms.
// See: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity-weight
// for how weights are used during scheduling.
func AddPreferredNodeSelectorRequirements(base *v1.Affinity, weight int32, new ...v1.NodeSelectorRequirement) {
	if base.NodeAffinity == nil {
		base.NodeAffinity = &v1.NodeAffinity{}
	}
	base.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
		base.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
		v1.PreferredSchedulingTerm{
			Weight: weight,
			Preference: v1.NodeSelectorTerm{
				MatchExpressions: new,
			},
		},
	)
}

// ApplyInterruptibleNodeSelectorRequirement configures the node selector requirement of the node-affinity using the configuration specified.
func ApplyInterruptibleNodeSelectorRequirement(interruptible bool, affinity *v1.Affinity) {
	// Determine node selector terms to add to node affinity
	var nodeSelectorRequirement v1.NodeSelectorRequirement
	if interruptible {
		if config.GetK8sPluginConfig().InterruptibleNodeSelectorRequirement == nil {
			return
		}
		nodeSelectorRequirement = *config.GetK8sPluginConfig().InterruptibleNodeSelectorRequirement
	} else {
		if config.GetK8sPluginConfig().NonInterruptibleNodeSelectorRequirement == nil {
			return
		}
		nodeSelectorRequirement = *config.GetK8sPluginConfig().NonInterruptibleNodeSelectorRequirement
	}

	AddRequiredNodeSelectorRequirements(affinity, nodeSelectorRequirement)
}

// ApplyInterruptibleNodeAffinity configures the node-affinity for the pod using the configuration specified.
func ApplyInterruptibleNodeAffinity(interruptible bool, podSpec *v1.PodSpec) {
	if podSpec.Affinity == nil {
		podSpec.Affinity = &v1.Affinity{}
	}

	ApplyInterruptibleNodeSelectorRequirement(interruptible, podSpec.Affinity)
}

// Specialized merging of overrides into a base *core.ExtendedResources object. Note
// that doing a nested merge may not be the intended behavior all the time, so we
// handle each field separately here.
func applyExtendedResourcesOverrides(base, overrides *core.ExtendedResources) *core.ExtendedResources {
	// Handle case where base might be nil
	var new *core.ExtendedResources
	if base == nil {
		new = &core.ExtendedResources{}
	} else {
		new = proto.Clone(base).(*core.ExtendedResources)
	}

	// No overrides found
	if overrides == nil {
		return new
	}

	// GPU Accelerator
	if overrides.GetGpuAccelerator() != nil {
		new.GpuAccelerator = overrides.GetGpuAccelerator()
	}

	return new
}

func ApplyGPUNodeSelectors(podSpec *v1.PodSpec, gpuAccelerator *core.GPUAccelerator) {
	// Short circuit if pod spec does not contain any containers that use GPUs
	gpuResourceName := config.GetK8sPluginConfig().GpuResourceName
	requiresGPUs := false
	for _, cnt := range podSpec.Containers {
		if _, ok := cnt.Resources.Limits[gpuResourceName]; ok {
			requiresGPUs = true
			break
		}
	}
	if !requiresGPUs {
		return
	}

	if podSpec.Affinity == nil {
		podSpec.Affinity = &v1.Affinity{}
	}

	// Apply changes for GPU device
	device := gpuAccelerator.GetDevice()
	if len(device) > 0 {
		// Add node selector requirement for GPU device
		deviceNsr := v1.NodeSelectorRequirement{
			Key:      config.GetK8sPluginConfig().GpuDeviceNodeLabel,
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{device},
		}
		AddRequiredNodeSelectorRequirements(podSpec.Affinity, deviceNsr)
		// Add toleration for GPU device
		deviceTol := v1.Toleration{
			Key:      config.GetK8sPluginConfig().GpuDeviceNodeLabel,
			Value:    device,
			Operator: v1.TolerationOpEqual,
			Effect:   v1.TaintEffectNoSchedule,
		}
		podSpec.Tolerations = append(podSpec.Tolerations, deviceTol)
	}

	// Short circuit if a partition size preference is not specified
	partitionSizeValue := gpuAccelerator.GetPartitionSizeValue()
	if partitionSizeValue == nil {
		return
	}

	// Apply changes for GPU partition size, if applicable
	var partitionSizeNsr *v1.NodeSelectorRequirement
	var partitionSizeTol *v1.Toleration
	switch p := partitionSizeValue.(type) {
	case *core.GPUAccelerator_Unpartitioned:
		if !p.Unpartitioned {
			break
		}
		if config.GetK8sPluginConfig().GpuUnpartitionedNodeSelectorRequirement != nil {
			partitionSizeNsr = config.GetK8sPluginConfig().GpuUnpartitionedNodeSelectorRequirement
		} else {
			partitionSizeNsr = &v1.NodeSelectorRequirement{
				Key:      config.GetK8sPluginConfig().GpuPartitionSizeNodeLabel,
				Operator: v1.NodeSelectorOpDoesNotExist,
			}
		}
		if config.GetK8sPluginConfig().GpuUnpartitionedToleration != nil {
			partitionSizeTol = config.GetK8sPluginConfig().GpuUnpartitionedToleration
		}
	case *core.GPUAccelerator_PartitionSize:
		partitionSizeNsr = &v1.NodeSelectorRequirement{
			Key:      config.GetK8sPluginConfig().GpuPartitionSizeNodeLabel,
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{p.PartitionSize},
		}
		partitionSizeTol = &v1.Toleration{
			Key:      config.GetK8sPluginConfig().GpuPartitionSizeNodeLabel,
			Value:    p.PartitionSize,
			Operator: v1.TolerationOpEqual,
			Effect:   v1.TaintEffectNoSchedule,
		}
	}
	if partitionSizeNsr != nil {
		AddRequiredNodeSelectorRequirements(podSpec.Affinity, *partitionSizeNsr)
	}
	if partitionSizeTol != nil {
		podSpec.Tolerations = append(podSpec.Tolerations, *partitionSizeTol)
	}
}

// UpdatePod updates the base pod spec used to execute tasks. This is configured with plugins and task metadata-specific options
func UpdatePod(taskExecutionMetadata pluginsCore.TaskExecutionMetadata,
	resourceRequirements []v1.ResourceRequirements, podSpec *v1.PodSpec) {
	if len(podSpec.RestartPolicy) == 0 {
		podSpec.RestartPolicy = v1.RestartPolicyNever
	}
	podSpec.Tolerations = append(
		GetPodTolerations(taskExecutionMetadata.IsInterruptible(), resourceRequirements...), podSpec.Tolerations...)

	if len(podSpec.ServiceAccountName) == 0 {
		podSpec.ServiceAccountName = taskExecutionMetadata.GetK8sServiceAccount()
	}
	if len(podSpec.SchedulerName) == 0 {
		podSpec.SchedulerName = config.GetK8sPluginConfig().SchedulerName
	}
	podSpec.NodeSelector = utils.UnionMaps(config.GetK8sPluginConfig().DefaultNodeSelector, podSpec.NodeSelector)
	if taskExecutionMetadata.IsInterruptible() {
		podSpec.NodeSelector = utils.UnionMaps(podSpec.NodeSelector, config.GetK8sPluginConfig().InterruptibleNodeSelector)
	}
	if podSpec.Affinity == nil && config.GetK8sPluginConfig().DefaultAffinity != nil {
		podSpec.Affinity = config.GetK8sPluginConfig().DefaultAffinity.DeepCopy()
	}
	if podSpec.SecurityContext == nil && config.GetK8sPluginConfig().DefaultPodSecurityContext != nil {
		podSpec.SecurityContext = config.GetK8sPluginConfig().DefaultPodSecurityContext.DeepCopy()
	}
	if config.GetK8sPluginConfig().EnableHostNetworkingPod != nil {
		podSpec.HostNetwork = *config.GetK8sPluginConfig().EnableHostNetworkingPod
	}
	if podSpec.DNSConfig == nil && config.GetK8sPluginConfig().DefaultPodDNSConfig != nil {
		podSpec.DNSConfig = config.GetK8sPluginConfig().DefaultPodDNSConfig.DeepCopy()
	}
	ApplyInterruptibleNodeAffinity(taskExecutionMetadata.IsInterruptible(), podSpec)
}

func mergeMapInto(src map[string]string, dst map[string]string) {
	for key, value := range src {
		dst[key] = value
	}
}

// BuildRawPod constructs a PodSpec and ObjectMeta based on the definition passed by the TaskExecutionContext. This
// definition does not include any configuration injected by Flyte.
func BuildRawPod(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (*v1.PodSpec, *metav1.ObjectMeta, string, error) {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		logger.Warnf(ctx, "failed to read task information when trying to construct Pod, err: %s", err.Error())
		return nil, nil, "", err
	}

	var podSpec *v1.PodSpec
	objectMeta := metav1.ObjectMeta{
		Annotations: make(map[string]string),
		Labels:      make(map[string]string),
	}
	primaryContainerName := ""

	switch target := taskTemplate.GetTarget().(type) {
	case *core.TaskTemplate_Container:
		// handles tasks defined by a single container
		c, err := BuildRawContainer(ctx, tCtx)
		if err != nil {
			return nil, nil, "", err
		}

		primaryContainerName = c.Name
		podSpec = &v1.PodSpec{
			Containers: []v1.Container{
				*c,
			},
		}
	case *core.TaskTemplate_K8SPod:
		// handles pod tasks that marshal the pod spec to the k8s_pod task target.
		if target.K8SPod.PodSpec == nil {
			return nil, nil, "", pluginserrors.Errorf(pluginserrors.BadTaskSpecification,
				"Pod tasks with task type version > 1 should specify their target as a K8sPod with a defined pod spec")
		}

		err := utils.UnmarshalStructToObj(target.K8SPod.PodSpec, &podSpec)
		if err != nil {
			return nil, nil, "", pluginserrors.Errorf(pluginserrors.BadTaskSpecification,
				"Unable to unmarshal task k8s pod [%v], Err: [%v]", target.K8SPod.PodSpec, err.Error())
		}

		// get primary container name
		var ok bool
		if primaryContainerName, ok = taskTemplate.GetConfig()[PrimaryContainerKey]; !ok {
			return nil, nil, "", pluginserrors.Errorf(pluginserrors.BadTaskSpecification,
				"invalid TaskSpecification, config missing [%s] key in [%v]", PrimaryContainerKey, taskTemplate.GetConfig())
		}

		// update annotations and labels
		if taskTemplate.GetK8SPod().Metadata != nil {
			mergeMapInto(target.K8SPod.Metadata.Annotations, objectMeta.Annotations)
			mergeMapInto(target.K8SPod.Metadata.Labels, objectMeta.Labels)
		}
	default:
		return nil, nil, "", pluginserrors.Errorf(pluginserrors.BadTaskSpecification,
			"invalid TaskSpecification, unable to determine Pod configuration")
	}

	return podSpec, &objectMeta, primaryContainerName, nil
}

// ApplyFlytePodConfiguration updates the PodSpec and ObjectMeta with various Flyte configuration. This includes
// applying default k8s configuration, applying overrides (resources etc.), injecting copilot containers, and merging with the
// configuration PodTemplate (if exists).
func ApplyFlytePodConfiguration(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, podSpec *v1.PodSpec, objectMeta *metav1.ObjectMeta, primaryContainerName string) (*v1.PodSpec, *metav1.ObjectMeta, error) {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		logger.Warnf(ctx, "failed to read task information when trying to construct Pod, err: %s", err.Error())
		return nil, nil, err
	}

	// add flyte resource customizations to containers
	templateParameters := template.Parameters{
		Inputs:           tCtx.InputReader(),
		OutputPath:       tCtx.OutputWriter(),
		Task:             tCtx.TaskReader(),
		TaskExecMetadata: tCtx.TaskExecutionMetadata(),
	}

	resourceRequests := make([]v1.ResourceRequirements, 0, len(podSpec.Containers))
	var primaryContainer *v1.Container
	for index, container := range podSpec.Containers {
		var resourceMode = ResourceCustomizationModeEnsureExistingResourcesInRange
		if container.Name == primaryContainerName {
			resourceMode = ResourceCustomizationModeMergeExistingResources
		}

		if err := AddFlyteCustomizationsToContainer(ctx, templateParameters, resourceMode, &podSpec.Containers[index]); err != nil {
			return nil, nil, err
		}

		resourceRequests = append(resourceRequests, podSpec.Containers[index].Resources)
		if container.Name == primaryContainerName {
			primaryContainer = &podSpec.Containers[index]
		}
	}

	if primaryContainer == nil {
		return nil, nil, pluginserrors.Errorf(pluginserrors.BadTaskSpecification, "invalid TaskSpecification, primary container [%s] not defined", primaryContainerName)
	}

	// add copilot configuration to primaryContainer and PodSpec (if necessary)
	var dataLoadingConfig *core.DataLoadingConfig
	if container := taskTemplate.GetContainer(); container != nil {
		dataLoadingConfig = container.GetDataConfig()
	} else if pod := taskTemplate.GetK8SPod(); pod != nil {
		dataLoadingConfig = pod.GetDataConfig()
	}

	if dataLoadingConfig != nil {
		if err := AddCoPilotToContainer(ctx, config.GetK8sPluginConfig().CoPilot,
			primaryContainer, taskTemplate.Interface, dataLoadingConfig); err != nil {
			return nil, nil, err
		}

		if err := AddCoPilotToPod(ctx, config.GetK8sPluginConfig().CoPilot, podSpec, taskTemplate.GetInterface(),
			tCtx.TaskExecutionMetadata(), tCtx.InputReader(), tCtx.OutputWriter(), dataLoadingConfig); err != nil {
			return nil, nil, err
		}
	}

	// update primaryContainer and PodSpec with k8s plugin configuration, etc
	UpdatePod(tCtx.TaskExecutionMetadata(), resourceRequests, podSpec)
	if primaryContainer.SecurityContext == nil && config.GetK8sPluginConfig().DefaultSecurityContext != nil {
		primaryContainer.SecurityContext = config.GetK8sPluginConfig().DefaultSecurityContext.DeepCopy()
	}

	// merge PodSpec and ObjectMeta with configuration pod template (if exists)
	podSpec, objectMeta, err = MergeWithBasePodTemplate(ctx, tCtx, podSpec, objectMeta, primaryContainerName)
	if err != nil {
		return nil, nil, err
	}

	// handling for extended resources
	// Merge overrides with base extended resources
	extendedResources := applyExtendedResourcesOverrides(
		taskTemplate.GetExtendedResources(),
		tCtx.TaskExecutionMetadata().GetOverrides().GetExtendedResources(),
	)

	// GPU accelerator
	if extendedResources.GetGpuAccelerator() != nil {
		ApplyGPUNodeSelectors(podSpec, extendedResources.GetGpuAccelerator())
	}

	// Override container image if necessary
	if len(tCtx.TaskExecutionMetadata().GetOverrides().GetContainerImage()) > 0 {
		ApplyContainerImageOverride(podSpec, tCtx.TaskExecutionMetadata().GetOverrides().GetContainerImage(), primaryContainerName)
	}

	return podSpec, objectMeta, nil
}

func ApplyContainerImageOverride(podSpec *v1.PodSpec, containerImage string, primaryContainerName string) {
	for i, c := range podSpec.Containers {
		if c.Name == primaryContainerName {
			podSpec.Containers[i].Image = containerImage
			return
		}
	}
}

// ToK8sPodSpec builds a PodSpec and ObjectMeta based on the definition passed by the TaskExecutionContext. This
// involves parsing the raw PodSpec definition and applying all Flyte configuration options.
func ToK8sPodSpec(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (*v1.PodSpec, *metav1.ObjectMeta, string, error) {
	// build raw PodSpec and ObjectMeta
	podSpec, objectMeta, primaryContainerName, err := BuildRawPod(ctx, tCtx)
	if err != nil {
		return nil, nil, "", err
	}

	// add flyte configuration
	podSpec, objectMeta, err = ApplyFlytePodConfiguration(ctx, tCtx, podSpec, objectMeta, primaryContainerName)
	if err != nil {
		return nil, nil, "", err
	}

	return podSpec, objectMeta, primaryContainerName, nil
}

func GetContainer(podSpec *v1.PodSpec, name string) (*v1.Container, error) {
	for _, container := range podSpec.Containers {
		if container.Name == name {
			return &container, nil
		}
	}
	return nil, pluginserrors.Errorf(pluginserrors.BadTaskSpecification, "invalid TaskSpecification, container [%s] not defined", name)
}

// getBasePodTemplate attempts to retrieve the PodTemplate to use as the base for k8s Pod configuration. This value can
// come from one of the following:
// (1) PodTemplate name in the TaskMetadata: This name is then looked up in the PodTemplateStore.
// (2) Default PodTemplate name from configuration: This name is then looked up in the PodTemplateStore.
func getBasePodTemplate(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, podTemplateStore PodTemplateStore) (*v1.PodTemplate, error) {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, pluginserrors.Errorf(pluginserrors.BadTaskSpecification, "TaskSpecification cannot be read, Err: [%v]", err.Error())
	}

	var podTemplate *v1.PodTemplate
	if taskTemplate.Metadata != nil && len(taskTemplate.Metadata.PodTemplateName) > 0 {
		// retrieve PodTemplate by name from PodTemplateStore
		podTemplate = podTemplateStore.LoadOrDefault(tCtx.TaskExecutionMetadata().GetNamespace(), taskTemplate.Metadata.PodTemplateName)
		if podTemplate == nil {
			return nil, pluginserrors.Errorf(pluginserrors.BadTaskSpecification, "PodTemplate '%s' does not exist", taskTemplate.Metadata.PodTemplateName)
		}
	} else {
		// check for default PodTemplate
		podTemplate = podTemplateStore.LoadOrDefault(tCtx.TaskExecutionMetadata().GetNamespace(), config.GetK8sPluginConfig().DefaultPodTemplateName)
	}

	return podTemplate, nil
}

// MergeWithBasePodTemplate attempts to merge the provided PodSpec and ObjectMeta with the configuration PodTemplate for
// this task.
func MergeWithBasePodTemplate(ctx context.Context, tCtx pluginsCore.TaskExecutionContext,
	podSpec *v1.PodSpec, objectMeta *metav1.ObjectMeta, primaryContainerName string) (*v1.PodSpec, *metav1.ObjectMeta, error) {

	// attempt to retrieve base PodTemplate
	podTemplate, err := getBasePodTemplate(ctx, tCtx, DefaultPodTemplateStore)
	if err != nil {
		return nil, nil, err
	} else if podTemplate == nil {
		// if no PodTemplate to merge as base -> return
		return podSpec, objectMeta, nil
	}

	// merge podSpec with podTemplate
	mergedPodSpec, err := mergePodSpecs(&podTemplate.Template.Spec, podSpec, primaryContainerName)
	if err != nil {
		return nil, nil, err
	}

	// merge PodTemplate PodSpec with podSpec
	var mergedObjectMeta *metav1.ObjectMeta = podTemplate.Template.ObjectMeta.DeepCopy()
	if err := mergo.Merge(mergedObjectMeta, objectMeta, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
		return nil, nil, err
	}

	return mergedPodSpec, mergedObjectMeta, nil
}

// mergePodSpecs merges the two provided PodSpecs. This process uses the first as the base configuration, where values
// set by the first PodSpec are overwritten by the second in the return value. Additionally, this function applies
// container-level configuration from the basePodSpec.
func mergePodSpecs(basePodSpec *v1.PodSpec, podSpec *v1.PodSpec, primaryContainerName string) (*v1.PodSpec, error) {
	if basePodSpec == nil || podSpec == nil {
		return nil, errors.New("neither the basePodSpec or the podSpec can be nil")
	}

	// extract defaultContainerTemplate and primaryContainerTemplate
	var defaultContainerTemplate, primaryContainerTemplate *v1.Container
	for i := 0; i < len(basePodSpec.Containers); i++ {
		if basePodSpec.Containers[i].Name == defaultContainerTemplateName {
			defaultContainerTemplate = &basePodSpec.Containers[i]
		} else if basePodSpec.Containers[i].Name == primaryContainerTemplateName {
			primaryContainerTemplate = &basePodSpec.Containers[i]
		}
	}

	// merge PodTemplate PodSpec with podSpec
	var mergedPodSpec *v1.PodSpec = basePodSpec.DeepCopy()
	if err := mergo.Merge(mergedPodSpec, podSpec, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
		return nil, err
	}

	// merge PodTemplate containers
	var mergedContainers []v1.Container
	for _, container := range podSpec.Containers {
		// if applicable start with defaultContainerTemplate
		var mergedContainer *v1.Container
		if defaultContainerTemplate != nil {
			mergedContainer = defaultContainerTemplate.DeepCopy()
		}

		// if applicable merge with primaryContainerTemplate
		if container.Name == primaryContainerName && primaryContainerTemplate != nil {
			if mergedContainer == nil {
				mergedContainer = primaryContainerTemplate.DeepCopy()
			} else {
				err := mergo.Merge(mergedContainer, primaryContainerTemplate, mergo.WithOverride, mergo.WithAppendSlice)
				if err != nil {
					return nil, err
				}
			}
		}

		// if applicable merge with existing container
		if mergedContainer == nil {
			mergedContainers = append(mergedContainers, container)
		} else {
			err := mergo.Merge(mergedContainer, container, mergo.WithOverride, mergo.WithAppendSlice)
			if err != nil {
				return nil, err
			}

			mergedContainers = append(mergedContainers, *mergedContainer)
		}
	}

	mergedPodSpec.Containers = mergedContainers
	return mergedPodSpec, nil
}

func BuildIdentityPod() *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       PodKind,
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}
}

// DemystifyPending is one the core functions, that helps FlytePropeller determine if a pending pod is indeed pending,
// or it is actually stuck in a un-reparable state. In such a case the pod should be marked as dead and the task should
// be retried. This has to be handled sadly, as K8s is still largely designed for long running services that should
// recover from failures, but Flyte pods are completely automated and should either run or fail
// Important considerations.
// Pending Status in Pod could be for various reasons and sometimes could signal a problem
// Case I: Pending because the Image pull is failing and it is backing off
//
//	This could be transient. So we can actually rely on the failure reason.
//	The failure transitions from ErrImagePull -> ImagePullBackoff
//
// Case II: Not enough resources are available. This is tricky. It could be that the total number of
//
//	resources requested is beyond the capability of the system. for this we will rely on configuration
//	and hence input gates. We should not allow bad requests that Request for large number of resource through.
//	In the case it makes through, we will fail after timeout
func DemystifyPending(status v1.PodStatus, info pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, error) {
	phaseInfo, t := demystifyPendingHelper(status, info)

	if phaseInfo.Phase().IsTerminal() {
		return phaseInfo, nil
	}

	podPendingTimeout := config.GetK8sPluginConfig().PodPendingTimeout.Duration
	if podPendingTimeout > 0 && time.Since(t) >= podPendingTimeout {
		return pluginsCore.PhaseInfoRetryableFailureWithCleanup("PodPendingTimeout", phaseInfo.Reason(), &pluginsCore.TaskInfo{
			OccurredAt: &t,
		}), nil
	}

	if phaseInfo.Phase() != pluginsCore.PhaseUndefined {
		return phaseInfo, nil
	}

	return pluginsCore.PhaseInfoQueuedWithTaskInfo(time.Now(), pluginsCore.DefaultPhaseVersion, "Scheduling", phaseInfo.Info()), nil
}

func demystifyPendingHelper(status v1.PodStatus, info pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, time.Time) {
	// Search over the difference conditions in the status object.  Note that the 'Pending' this function is
	// demystifying is the 'phase' of the pod status. This is different than the PodReady condition type also used below
	phaseInfo := pluginsCore.PhaseInfoQueuedWithTaskInfo(time.Now(), pluginsCore.DefaultPhaseVersion, "Demistify Pending", &info)

	t := time.Now()
	for _, c := range status.Conditions {
		t = c.LastTransitionTime.Time
		switch c.Type {
		case v1.PodScheduled:
			if c.Status == v1.ConditionFalse {
				// Waiting to be scheduled. This usually refers to inability to acquire resources.
				return pluginsCore.PhaseInfoQueuedWithTaskInfo(t, pluginsCore.DefaultPhaseVersion, fmt.Sprintf("%s:%s", c.Reason, c.Message), phaseInfo.Info()), t
			}

		case v1.PodReasonUnschedulable:
			// We Ignore case in which we are unable to find resources on the cluster. This is because
			// - The resources may be not available at the moment, but may become available eventually
			//   The pod scheduler will keep on looking at this pod and trying to satisfy it.
			//
			//  Pod status looks like this:
			// 	message: '0/1 nodes are available: 1 Insufficient memory.'
			//  reason: Unschedulable
			// 	status: "False"
			// 	type: PodScheduled
			return pluginsCore.PhaseInfoQueuedWithTaskInfo(t, pluginsCore.DefaultPhaseVersion, fmt.Sprintf("%s:%s", c.Reason, c.Message), phaseInfo.Info()), t

		case v1.PodReady:
			if c.Status == v1.ConditionFalse {
				// This happens in the case the image is having some problems. In the following example, K8s is having
				// problems downloading an image. To ensure that, we will have to iterate over all the container statuses and
				// find if some container has imagepull failure
				// e.g.
				//     - lastProbeTime: null
				//      lastTransitionTime: 2018-12-18T00:57:30Z
				//      message: 'containers with unready status: [myapp-container]'
				//      reason: ContainersNotReady
				//      status: "False"
				//      type: Ready
				//
				// e.g. Container status
				//     - image: blah
				//      imageID: ""
				//      lastState: {}
				//      name: myapp-container
				//      ready: false
				//      restartCount: 0
				//      state:
				//        waiting:
				//          message: Back-off pulling image "blah"
				//          reason: ImagePullBackOff
				for _, containerStatus := range status.ContainerStatuses {
					if !containerStatus.Ready {
						if containerStatus.State.Waiting != nil {
							// There are a variety of reasons that can cause a pod to be in this waiting state.
							// Waiting state may be legitimate when the container is being downloaded, started or init containers are running
							reason := containerStatus.State.Waiting.Reason
							finalReason := fmt.Sprintf("%s|%s", c.Reason, reason)
							finalMessage := fmt.Sprintf("%s|%s", c.Message, containerStatus.State.Waiting.Message)
							switch reason {
							case "ErrImagePull", "ContainerCreating", "PodInitializing":
								// But, there are only two "reasons" when a pod is successfully being created and hence it is in
								// waiting state
								// Refer to https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kubelet_pods.go
								// and look for the default waiting states
								// We also want to allow Image pulls to be retried, so ErrImagePull will be ignored
								// as it eventually enters into ImagePullBackOff
								// ErrImagePull -> Transitionary phase to ImagePullBackOff
								// ContainerCreating -> Image is being downloaded
								// PodInitializing -> Init containers are running
								return pluginsCore.PhaseInfoInitializing(t, pluginsCore.DefaultPhaseVersion, fmt.Sprintf("[%s]: %s", finalReason, finalMessage), &pluginsCore.TaskInfo{OccurredAt: &t}), t

							case "CreateContainerError":
								// This may consist of:
								// 1. Transient errors: e.g. failed to reserve
								// container name, container name [...] already in use
								// by container
								// 2. Permanent errors: e.g. no command specified
								// To handle both types of errors gracefully without
								// arbitrary pattern matching in the message, we simply
								// allow a grace period for kubelet to resolve
								// transient issues with the container runtime. If the
								// error persists beyond this time, the corresponding
								// task is marked as failed.
								// NOTE: The current implementation checks for a timeout
								// by comparing the condition's LastTransitionTime
								// based on the corresponding kubelet's clock with the
								// current time based on FlytePropeller's clock. This
								// is not ideal given that these 2 clocks are NOT
								// synced, and therefore, only provides an
								// approximation of the elapsed time since the last
								// transition.

								gracePeriod := config.GetK8sPluginConfig().CreateContainerErrorGracePeriod.Duration
								if time.Since(t) >= gracePeriod {
									return pluginsCore.PhaseInfoFailureWithCleanup(finalReason, GetMessageAfterGracePeriod(finalMessage, gracePeriod), &pluginsCore.TaskInfo{
										OccurredAt: &t,
									}), t
								}
								return pluginsCore.PhaseInfoInitializing(
									t,
									pluginsCore.DefaultPhaseVersion,
									fmt.Sprintf("[%s]: %s", finalReason, finalMessage),
									&pluginsCore.TaskInfo{OccurredAt: &t},
								), t

							case "CreateContainerConfigError":
								gracePeriod := config.GetK8sPluginConfig().CreateContainerConfigErrorGracePeriod.Duration
								if time.Since(t) >= gracePeriod {
									return pluginsCore.PhaseInfoFailureWithCleanup(finalReason, GetMessageAfterGracePeriod(finalMessage, gracePeriod), &pluginsCore.TaskInfo{
										OccurredAt: &t,
									}), t
								}
								return pluginsCore.PhaseInfoInitializing(
									t,
									pluginsCore.DefaultPhaseVersion,
									fmt.Sprintf("[%s]: %s", finalReason, finalMessage),
									&pluginsCore.TaskInfo{OccurredAt: &t},
								), t

							case "InvalidImageName":
								return pluginsCore.PhaseInfoFailureWithCleanup(finalReason, finalMessage, &pluginsCore.TaskInfo{
									OccurredAt: &t,
								}), t

							case "ImagePullBackOff":
								gracePeriod := config.GetK8sPluginConfig().ImagePullBackoffGracePeriod.Duration
								if time.Since(t) >= gracePeriod {
									return pluginsCore.PhaseInfoRetryableFailureWithCleanup(finalReason, GetMessageAfterGracePeriod(finalMessage, gracePeriod), &pluginsCore.TaskInfo{
										OccurredAt: &t,
									}), t
								}

								return pluginsCore.PhaseInfoInitializing(
									t,
									pluginsCore.DefaultPhaseVersion,
									fmt.Sprintf("[%s]: %s", finalReason, finalMessage),
									&pluginsCore.TaskInfo{OccurredAt: &t},
								), t

							default:
								// Since we are not checking for all error states, we may end up perpetually
								// in the queued state returned at the bottom of this function, until the Pod is reaped
								// by K8s and we get elusive 'pod not found' errors
								// So be default if the container is not waiting with the PodInitializing/ContainerCreating
								// reasons, then we will assume a failure reason, and fail instantly
								return pluginsCore.PhaseInfoSystemRetryableFailureWithCleanup(finalReason, finalMessage, &pluginsCore.TaskInfo{
									OccurredAt: &t,
								}), t
							}
						}
					}
				}
			}
		}
	}

	return phaseInfo, t
}

func GetMessageAfterGracePeriod(message string, gracePeriod time.Duration) string {
	return fmt.Sprintf("Grace period [%s] exceeded|%s", gracePeriod, message)
}

func DemystifySuccess(status v1.PodStatus, info pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, error) {
	for _, status := range append(
		append(status.InitContainerStatuses, status.ContainerStatuses...), status.EphemeralContainerStatuses...) {
		if status.State.Terminated != nil && strings.Contains(status.State.Terminated.Reason, OOMKilled) {
			return pluginsCore.PhaseInfoRetryableFailure(OOMKilled,
				"Pod reported success despite being OOMKilled", &info), nil
		}
	}
	return pluginsCore.PhaseInfoSuccess(&info), nil
}

// DeterminePrimaryContainerPhase as the name suggests, given all the containers, will return a pluginsCore.PhaseInfo object
// corresponding to the phase of the primaryContainer which is identified using the provided name.
// This is useful in case of sidecars or pod jobs, where Flyte will monitor successful exit of a single container.
func DeterminePrimaryContainerPhase(primaryContainerName string, statuses []v1.ContainerStatus, info *pluginsCore.TaskInfo) pluginsCore.PhaseInfo {
	for _, s := range statuses {
		if s.Name == primaryContainerName {
			if s.State.Waiting != nil || s.State.Running != nil {
				return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, info)
			}

			if s.State.Terminated != nil {
				if s.State.Terminated.ExitCode != 0 || strings.Contains(s.State.Terminated.Reason, OOMKilled) {
					message := fmt.Sprintf("\r\n[%v] terminated with exit code (%v). Reason [%v]. Message: \n%v.",
						s.Name,
						s.State.Terminated.ExitCode,
						s.State.Terminated.Reason,
						s.State.Terminated.Message)
					return pluginsCore.PhaseInfoRetryableFailure(
						s.State.Terminated.Reason, message, info)
				}
				return pluginsCore.PhaseInfoSuccess(info)
			}
		}
	}

	// If for some reason we can't find the primary container, always just return a permanent failure
	return pluginsCore.PhaseInfoFailure(PrimaryContainerNotFound,
		fmt.Sprintf("Primary container [%s] not found in pod's container statuses", primaryContainerName), info)
}

// DemystifyFailure resolves the various Kubernetes pod failure modes to determine
// the most appropriate course of action
func DemystifyFailure(status v1.PodStatus, info pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, error) {
	code := "UnknownError"
	message := "Pod failed. No message received from kubernetes."
	if len(status.Reason) > 0 {
		code = status.Reason
	}

	if len(status.Message) > 0 {
		message = status.Message
	}

	//
	// Handle known pod statuses
	//
	// This is useful for handling node interruption events
	// which can be different between providers and versions of Kubernetes. Given that
	// we don't have a consistent way of detecting interruption events, we will be
	// documenting all possibilities as follows. We will also be handling these as
	// system retryable failures that do not count towards user-specified task retries,
	// for now. This is required for FlytePropeller to correctly transition
	// interruptible nodes to non-interruptible ones after the
	// `interruptible-failure-threshold` is exceeded. See:
	// https://github.com/flyteorg/flytepropeller/blob/a3c6e91f19c19601a957b29891437112868845de/pkg/controller/nodes/node_exec_context.go#L213

	// GKE (>= v1.20) Kubelet graceful node shutdown
	// See: https://cloud.google.com/kubernetes-engine/docs/how-to/preemptible-vms#graceful-shutdown
	// Cloud audit log for patch of Pod object during graceful node shutdown:
	// request: {
	//     @type: "k8s.io/Patch"
	//     status: {
	//         conditions: null
	//         message: "Pod Node is in progress of shutting down, not admitting any new pods"
	//         phase: "Failed"
	//         qosClass: null
	//         reason: "Shutdown"
	//         startTime: "2022-01-30T14:24:07Z"
	//     }
	// }
	//
	// In some versions of GKE the reason can also be "Terminated"
	if code == "Shutdown" || code == "Terminated" {
		return pluginsCore.PhaseInfoSystemRetryableFailure(Interrupted, message, &info), nil
	}

	//
	// Handle known container statuses
	//
	for _, c := range append(
		append(status.InitContainerStatuses, status.ContainerStatuses...), status.EphemeralContainerStatuses...) {
		var containerState v1.ContainerState
		if c.LastTerminationState.Terminated != nil {
			containerState = c.LastTerminationState
		} else if c.State.Terminated != nil {
			containerState = c.State
		}
		if containerState.Terminated != nil {
			if strings.Contains(containerState.Terminated.Reason, OOMKilled) {
				code = OOMKilled
			} else if containerState.Terminated.ExitCode == SIGKILL {
				// in some setups, node termination sends SIGKILL to all the containers running on that node. Capturing and
				// tagging that correctly.
				code = Interrupted
			}

			if containerState.Terminated.ExitCode == 0 {
				message += fmt.Sprintf("\r\n[%v] terminated with ExitCode 0.", c.Name)
			} else {
				message += fmt.Sprintf("\r\n[%v] terminated with exit code (%v). Reason [%v]. Message: \n%v.",
					c.Name,
					containerState.Terminated.ExitCode,
					containerState.Terminated.Reason,
					containerState.Terminated.Message)
			}
		}
	}
	return pluginsCore.PhaseInfoRetryableFailure(code, message, &info), nil
}

func GetLastTransitionOccurredAt(pod *v1.Pod) metav1.Time {
	var lastTransitionTime metav1.Time
	containerStatuses := append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...)
	for _, containerStatus := range containerStatuses {
		if r := containerStatus.State.Running; r != nil {
			if r.StartedAt.Unix() > lastTransitionTime.Unix() {
				lastTransitionTime = r.StartedAt
			}
		} else if r := containerStatus.State.Terminated; r != nil {
			if r.FinishedAt.Unix() > lastTransitionTime.Unix() {
				lastTransitionTime = r.FinishedAt
			}
		}
	}

	if lastTransitionTime.IsZero() {
		lastTransitionTime = metav1.NewTime(time.Now())
	}

	return lastTransitionTime
}

func GetReportedAt(pod *v1.Pod) metav1.Time {
	var reportedAt metav1.Time
	for _, condition := range pod.Status.Conditions {
		if condition.Reason == "PodCompleted" && condition.Type == v1.PodReady && condition.Status == v1.ConditionFalse {
			if condition.LastTransitionTime.Unix() > reportedAt.Unix() {
				reportedAt = condition.LastTransitionTime
			}
		}
	}

	return reportedAt
}
