package spark

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	sparkOp "github.com/kubeflow/spark-operator/v2/api/v1beta2"
	sparkOpCommon "github.com/kubeflow/spark-operator/v2/pkg/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

const (
	KindSparkApplication                = "SparkApplication"
	sparkDriverUI                       = "sparkDriverUI"
	sparkHistoryUI                      = "sparkHistoryUI"
	defaultDriverPrimaryContainerName   = sparkOpCommon.SparkDriverContainerName
	defaultExecutorPrimaryContainerName = sparkOpCommon.Spark3DefaultExecutorContainerName
)

var featureRegex = regexp.MustCompile(`^spark.((flyteorg)|(flyte)).(.+).enabled$`)

var sparkTaskType = "spark"

// applicationStatePendingSubmission was written by pre-2.x operators; the kubeflow 2.x client
// no longer defines it but a status carrying it must still map to "submitted".
const applicationStatePendingSubmission = sparkOp.ApplicationStateType("PENDING_SUBMISSION")

type sparkResourceHandler struct{}

func validateSparkJob(sparkJob *plugins.SparkJob) error {
	if sparkJob == nil {
		return fmt.Errorf("empty sparkJob")
	}

	if len(sparkJob.MainApplicationFile) == 0 && len(sparkJob.MainClass) == 0 {
		return fmt.Errorf("either MainApplicationFile or MainClass must be set")
	}

	return nil
}

func (sparkResourceHandler) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

// Creates a new Job that will execute the main container as well as any generated types the result from the execution.
func (sparkResourceHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "unable to fetch task specification [%v]", err.Error())
	} else if taskTemplate == nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "nil task specification")
	}

	sparkJob := plugins.SparkJob{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &sparkJob) //nolint: staticcheck
	if err != nil {
		return nil, errors.Wrapf(errors.BadTaskSpecification, err, "invalid TaskSpecification [%v], failed to unmarshal", taskTemplate.GetCustom())
	}

	if err = validateSparkJob(&sparkJob); err != nil {
		return nil, errors.Wrapf(errors.BadTaskSpecification, err, "invalid TaskSpecification [%v].", taskTemplate.GetCustom())
	}

	sparkConfig := getSparkConfig(taskCtx, &sparkJob)
	driverSpec, err := createDriverSpec(ctx, taskCtx, sparkConfig, &sparkJob)
	if err != nil {
		return nil, err
	}
	executorSpec, err := createExecutorSpec(ctx, taskCtx, sparkConfig, &sparkJob)
	if err != nil {
		return nil, err
	}
	app := createSparkApplication(&sparkJob, sparkConfig, driverSpec, executorSpec)
	return app, nil
}

func getSparkConfig(taskCtx pluginsCore.TaskExecutionContext, sparkJob *plugins.SparkJob) map[string]string {
	// Start with default config values.
	sparkConfig := make(map[string]string)
	for k, v := range GetSparkConfig().DefaultSparkConfig {
		sparkConfig[k] = v
	}

	if sparkJob.GetExecutorPath() != "" {
		sparkConfig["spark.pyspark.python"] = sparkJob.GetExecutorPath()
		sparkConfig["spark.pyspark.driver.python"] = sparkJob.GetExecutorPath()
	}

	for k, v := range sparkJob.GetSparkConf() {
		// Add optional features if present.
		if featureRegex.MatchString(k) {
			addConfig(sparkConfig, k, v)
		} else {
			sparkConfig[k] = v
		}
	}

	// Set pod limits.
	if len(sparkConfig[sparkOpCommon.SparkKubernetesDriverLimitCores]) == 0 {
		// spark.kubernetes.driver.request.cores takes precedence over spark.driver.cores
		if len(sparkConfig[sparkOpCommon.SparkKubernetesDriverRequestCores]) != 0 {
			sparkConfig[sparkOpCommon.SparkKubernetesDriverLimitCores] = sparkConfig[sparkOpCommon.SparkKubernetesDriverRequestCores]
		} else if len(sparkConfig["spark.driver.cores"]) != 0 {
			sparkConfig[sparkOpCommon.SparkKubernetesDriverLimitCores] = sparkConfig["spark.driver.cores"]
		}
	}

	if len(sparkConfig[sparkOpCommon.SparkKubernetesExecutorLimitCores]) == 0 {
		// spark.kubernetes.executor.request.cores takes precedence over spark.executor.cores
		if len(sparkConfig[sparkOpCommon.SparkKubernetesExecutorRequestCores]) != 0 {
			sparkConfig[sparkOpCommon.SparkKubernetesExecutorLimitCores] = sparkConfig[sparkOpCommon.SparkKubernetesExecutorRequestCores]
		} else if len(sparkConfig["spark.executor.cores"]) != 0 {
			sparkConfig[sparkOpCommon.SparkKubernetesExecutorLimitCores] = sparkConfig["spark.executor.cores"]
		}
	}

	sparkConfig["spark.kubernetes.executor.podNamePrefix"] = taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	sparkConfig["spark.kubernetes.driverEnv.FLYTE_START_TIME"] = strconv.FormatInt(time.Now().UnixNano()/1000000, 10)

	return sparkConfig
}

func serviceAccountName(metadata pluginsCore.TaskExecutionMetadata) string {
	name := flytek8s.GetServiceAccountNameFromTaskExecutionMetadata(metadata)
	if len(name) == 0 {
		name = sparkTaskType
	}
	return name
}

func createSparkPodSpec(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, podSpec, customPodSpec *v1.PodSpec, container *v1.Container, sparkContainerName string) *sparkOp.SparkPodSpec {
	annotations := utils.UnionMaps(config.GetK8sPluginConfig().DefaultAnnotations, utils.CopyMap(taskCtx.TaskExecutionMetadata().GetAnnotations()))
	labels := utils.UnionMaps(config.GetK8sPluginConfig().DefaultLabels, utils.CopyMap(taskCtx.TaskExecutionMetadata().GetLabels()))

	sparkEnv := make([]v1.EnvVar, 0)
	for _, envVar := range container.Env {
		sparkEnv = append(sparkEnv, *envVar.DeepCopy())
	}
	sparkEnv = append(sparkEnv, v1.EnvVar{Name: "FLYTE_MAX_ATTEMPTS", Value: strconv.Itoa(int(taskCtx.TaskExecutionMetadata().GetMaxAttempts()))})

	spec := sparkOp.SparkPodSpec{
		Affinity:    podSpec.Affinity,
		Annotations: annotations,
		Labels:      labels,
		Env:         sparkEnv,
		Image:       &container.Image,

		// In pre-2.x SparkApplication CRDs, this field was called SecurityContenxt (sic) and serialized to `securityContext`.
		// This new field serializes to `podSecurityContext`, which is invisible to the old CRD.
		// Users who set this field (whether via platform defaults, pod templates, or pod templates in user code)
		// will see this setting getting dropped. Once they upgrade the CRD, things will start working again.
		PodSecurityContext: podSpec.SecurityContext.DeepCopy(),

		DNSConfig:     podSpec.DNSConfig.DeepCopy(),
		Tolerations:   podSpec.Tolerations,
		SchedulerName: &podSpec.SchedulerName,
		NodeSelector:  podSpec.NodeSelector,
		HostNetwork:   &podSpec.HostNetwork,
	}

	// The legacy fields above are always populated so the object stays valid on clusters whose
	// CRD/operator predate pod-template support (unknown fields are pruned by the API server
	// there). Where the CRD accepts it, additionally pass a pod spec through as the pod template;
	// the operator treats explicit fields as overrides of the template. The user's driver/executor
	// pod spec is passed through verbatim so the operator can patch it onto the container it
	// generates (`spark-kubernetes-driver`/`spark-kubernetes-executor`); Flyte's own container
	// names never match those, so merging it here would drop it on the floor.
	if GetSparkConfig().EnablePodTemplate && podTemplateSupported(ctx) {
		templatePodSpec := podSpec.DeepCopy()
		if customPodSpec != nil {
			templatePodSpec = customPodSpec.DeepCopy()
			// Preserve Flyte defaults when the user doesn't specify them in the custom pod spec.
			if templatePodSpec.EnableServiceLinks == nil {
				templatePodSpec.EnableServiceLinks = podSpec.EnableServiceLinks
			}
		} else {
			for i := range templatePodSpec.Containers {
				if templatePodSpec.Containers[i].Name != container.Name {
					continue
				}
				// The operator's mutating webhook patches the container it finds by name --
				// spark-kubernetes-driver, or executor/spark-kubernetes-executor.
				templatePodSpec.Containers[i].Name = sparkContainerName
				// SparkApplication already set the command and args, so we don't need to set it in the primary pod again.
				templatePodSpec.Containers[i].Command = nil
				templatePodSpec.Containers[i].Args = nil
				break
			}
		}

		spec.Template = &v1.PodTemplateSpec{Spec: *templatePodSpec}
		sa := serviceAccountName(taskCtx.TaskExecutionMetadata())
		spec.ServiceAccount = &sa
	}
	return &spec
}

type driverSpec struct {
	sparkSpec *sparkOp.DriverSpec
}

func createDriverSpec(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, sparkConfig map[string]string, sparkJob *plugins.SparkJob) (*driverSpec, error) {
	// Spark driver pods should always run as non-interruptible
	nonInterruptibleTaskCtx := flytek8s.NewPluginTaskExecutionContext(taskCtx, flytek8s.WithInterruptible(false))
	podSpec, _, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, nonInterruptibleTaskCtx)
	if err != nil {
		return nil, err
	}

	driverPod := sparkJob.GetDriverPod()
	var customPodSpec *v1.PodSpec
	if driverPod != nil {
		if driverPod.GetPodSpec() != nil {
			err = utils.UnmarshalStructToObj(driverPod.GetPodSpec(), &customPodSpec) //nolint: staticcheck
			if err != nil {
				return nil, errors.Errorf(errors.BadTaskSpecification,
					"Unable to unmarshal driver pod spec [%v], Err: [%v]", driverPod.GetPodSpec(), err.Error())
			}
		}
	}

	// Re-apply platform scheduling after the custom driver pod merge: the merge can
	// append OR'd node selector terms that would otherwise escape the default-affinity
	// and (forced) non-interruptible requirements. Idempotent.
	flytek8s.ApplyPlatformSchedulingConstraints(nonInterruptibleTaskCtx.TaskExecutionMetadata().IsInterruptible(), podSpec)

	primaryContainer, err := flytek8s.GetContainer(podSpec, primaryContainerName)
	if err != nil {
		return nil, err
	}
	sparkPodSpec := createSparkPodSpec(ctx, nonInterruptibleTaskCtx, podSpec, customPodSpec, primaryContainer, defaultDriverPrimaryContainerName)
	if sparkPodSpec.Template != nil {
		sparkConfig[sparkOpCommon.SparkKubernetesDriverPodTemplateContainerName] = defaultDriverPrimaryContainerName
	}
	serviceAccountName := serviceAccountName(nonInterruptibleTaskCtx.TaskExecutionMetadata())
	sparkPodSpec.ServiceAccount = &serviceAccountName
	spec := driverSpec{
		&sparkOp.DriverSpec{
			SparkPodSpec: *sparkPodSpec,
		},
	}
	if cores, err := strconv.ParseInt(sparkConfig["spark.driver.cores"], 10, 32); err == nil {
		spec.sparkSpec.Cores = intPtr(int32(cores))
	}
	spec.sparkSpec.Memory = strPtr(sparkConfig["spark.driver.memory"])
	return &spec, nil
}

type executorSpec struct {
	container *v1.Container
	sparkSpec *sparkOp.ExecutorSpec
}

func createExecutorSpec(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, sparkConfig map[string]string, sparkJob *plugins.SparkJob) (*executorSpec, error) {
	podSpec, _, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, err
	}

	executorPod := sparkJob.GetExecutorPod()
	var customPodSpec *v1.PodSpec
	if executorPod != nil {
		if executorPod.GetPodSpec() != nil {
			err = utils.UnmarshalStructToObj(executorPod.GetPodSpec(), &customPodSpec) //nolint: staticcheck
			if err != nil {
				return nil, errors.Errorf(errors.BadTaskSpecification,
					"Unable to unmarshal executor pod spec [%v], Err: [%v]", executorPod.GetPodSpec(), err.Error())
			}
		}
	}

	// Re-apply platform scheduling after the custom executor pod merge: the merge can
	// append OR'd node selector terms that would otherwise escape the default-affinity
	// and (non)interruptible requirements. Idempotent.
	flytek8s.ApplyPlatformSchedulingConstraints(taskCtx.TaskExecutionMetadata().IsInterruptible(), podSpec)

	primaryContainer, err := flytek8s.GetContainer(podSpec, primaryContainerName)
	if err != nil {
		return nil, err
	}
	sparkPodSpec := createSparkPodSpec(ctx, taskCtx, podSpec, customPodSpec, primaryContainer, defaultExecutorPrimaryContainerName)
	if sparkPodSpec.Template != nil {
		sparkConfig[sparkOpCommon.SparkKubernetesExecutorPodTemplateContainerName] = defaultExecutorPrimaryContainerName
	}
	spec := executorSpec{
		primaryContainer,
		&sparkOp.ExecutorSpec{
			SparkPodSpec: *sparkPodSpec,
		},
	}
	if execCores, err := strconv.ParseInt(sparkConfig["spark.executor.cores"], 10, 32); err == nil {
		spec.sparkSpec.Cores = intPtr(int32(execCores))
	}
	if execCount, err := strconv.ParseInt(sparkConfig["spark.executor.instances"], 10, 32); err == nil {
		spec.sparkSpec.Instances = intPtr(int32(execCount))
	}
	spec.sparkSpec.Memory = strPtr(sparkConfig["spark.executor.memory"])
	return &spec, nil
}

func createSparkApplication(sparkJob *plugins.SparkJob, sparkConfig map[string]string, driverSpec *driverSpec,
	executorSpec *executorSpec,
) *sparkOp.SparkApplication {
	// Hack: Retry submit failures in-case of resource limits hit.
	submissionFailureRetries := int32(14)

	app := &sparkOp.SparkApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindSparkApplication,
			APIVersion: sparkOp.SchemeGroupVersion.String(),
		},
		Spec: sparkOp.SparkApplicationSpec{
			Type:       getApplicationType(sparkJob.GetApplicationType()),
			Image:      &executorSpec.container.Image,
			Arguments:  executorSpec.container.Args,
			Driver:     *driverSpec.sparkSpec,
			Executor:   *executorSpec.sparkSpec,
			SparkConf:  sparkConfig,
			HadoopConf: sparkJob.GetHadoopConf(),
			// SubmissionFailures handled here. Task Failures handled at Propeller/Job level.
			RestartPolicy: sparkOp.RestartPolicy{
				Type:                       sparkOp.RestartPolicyOnFailure,
				OnSubmissionFailureRetries: &submissionFailureRetries,
			},
		},
	}

	// The operator reads spec.sparkVersion to decide how pod templates are handled: its
	// validating webhook rejects an application that carries a driver/executor template while
	// declaring less than 3.0.0, and its submission path labels the pod for webhook mutation
	// when the declared version is below 3.0.0 (internal/webhook/sparkapplication_validator.go
	// and internal/controller/sparkapplication/submission.go in kubeflow/spark-operator).
	// Left empty an application with a template is denied at admission, because an unset
	// version parses as invalid semver and sorts below every real one.
	if driverSpec.sparkSpec.Template != nil || executorSpec.sparkSpec.Template != nil {
		app.Spec.SparkVersion = minPodTemplateSparkVersion
	}

	if val, ok := sparkConfig["spark.batchScheduler"]; ok {
		app.Spec.BatchScheduler = &val
	}

	if sparkJob.MainApplicationFile != "" {
		app.Spec.MainApplicationFile = &sparkJob.MainApplicationFile
	}
	if sparkJob.MainClass != "" {
		app.Spec.MainClass = &sparkJob.MainClass
	}
	return app
}

func addConfig(sparkConfig map[string]string, key string, value string) {
	if strings.ToLower(strings.TrimSpace(value)) != "true" {
		sparkConfig[key] = value
		return
	}

	matches := featureRegex.FindAllStringSubmatch(key, -1)
	if len(matches) == 0 || len(matches[0]) == 0 {
		sparkConfig[key] = value
		return
	}
	featureName := matches[0][len(matches[0])-1]

	// Use the first matching feature in-case of duplicates.
	for _, feature := range GetSparkConfig().Features {
		if feature.Name == featureName {
			for k, v := range feature.SparkConfig {
				sparkConfig[k] = v
			}
			return
		}
	}
	sparkConfig[key] = value
}

// Convert SparkJob ApplicationType to Operator CRD ApplicationType
func getApplicationType(applicationType plugins.SparkApplication_Type) sparkOp.SparkApplicationType {
	switch applicationType {
	case plugins.SparkApplication_PYTHON:
		return sparkOp.SparkApplicationTypePython
	case plugins.SparkApplication_JAVA:
		return sparkOp.SparkApplicationTypeJava
	case plugins.SparkApplication_SCALA:
		return sparkOp.SparkApplicationTypeScala
	case plugins.SparkApplication_R:
		return sparkOp.SparkApplicationTypeR
	}
	return sparkOp.SparkApplicationTypePython
}

func (sparkResourceHandler) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &sparkOp.SparkApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindSparkApplication,
			APIVersion: sparkOp.SchemeGroupVersion.String(),
		},
	}, nil
}

func getEventInfoForSpark(ctx context.Context, pluginContext k8s.PluginContext, sj *sparkOp.SparkApplication) (*pluginsCore.TaskInfo, error) {
	taskTemplate, err := pluginContext.TaskReader().Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read task template: %w", err)
	}

	sparkConfig := GetSparkConfig()
	taskLogs := make([]*core.TaskLog, 0, 3)
	var logCtx *core.LogContext
	taskExecID := pluginContext.TaskExecutionMetadata().GetTaskExecutionID()

	if sj.Status.DriverInfo.PodName != "" {
		p, err := logs.InitializeLogPlugins(&sparkConfig.LogConfig.Mixed, taskTemplate)
		if err != nil {
			return nil, err
		}

		if p != nil {
			o, err := p.GetTaskLogs(tasklog.Input{
				PodName:         sj.Status.DriverInfo.PodName,
				Namespace:       sj.Namespace,
				TaskExecutionID: taskExecID,
				EnableVscode:    flytek8s.IsVscodeEnabled(ctx, sj.Spec.Driver.Env),
			})
			if err != nil {
				return nil, err
			}

			taskLogs = append(taskLogs, o.TaskLogs...)
		}
	}

	p, err := logs.InitializeLogPlugins(&sparkConfig.LogConfig.User, taskTemplate)
	if err != nil {
		return nil, err
	}

	if p != nil {
		o, err := p.GetTaskLogs(tasklog.Input{
			PodName:         sj.Status.DriverInfo.PodName,
			Namespace:       sj.Namespace,
			TaskExecutionID: taskExecID,
		})
		if err != nil {
			return nil, err
		}

		taskLogs = append(taskLogs, o.TaskLogs...)
	}

	logCtx = &core.LogContext{
		PrimaryPodName: sj.Status.DriverInfo.PodName,
	}
	logCtx.Pods = append(logCtx.Pods, &core.PodLogContext{
		Namespace:            sj.Namespace,
		PodName:              sj.Status.DriverInfo.PodName,
		PrimaryContainerName: defaultDriverPrimaryContainerName,
		Containers: []*core.ContainerContext{
			{ContainerName: defaultDriverPrimaryContainerName},
		},
	})

	for executorPodName, executorState := range sj.Status.ExecutorState {
		if executorState != sparkOp.ExecutorStatePending && executorState != sparkOp.ExecutorStateUnknown {
			logCtx.Pods = append(logCtx.Pods, &core.PodLogContext{
				Namespace:            sj.Namespace,
				PodName:              executorPodName,
				PrimaryContainerName: "spark-kubernetes-executor",
				Containers: []*core.ContainerContext{
					{ContainerName: "spark-kubernetes-executor"},
				},
			})
		}
	}

	p, err = logs.InitializeLogPlugins(&sparkConfig.LogConfig.System, taskTemplate)
	if err != nil {
		return nil, err
	}

	if p != nil {
		o, err := p.GetTaskLogs(tasklog.Input{
			PodName:         sj.Name,
			Namespace:       sj.Namespace,
			TaskExecutionID: taskExecID,
		})
		if err != nil {
			return nil, err
		}

		taskLogs = append(taskLogs, o.TaskLogs...)
	}

	p, err = logs.InitializeLogPlugins(&sparkConfig.LogConfig.AllUser, taskTemplate)
	if err != nil {
		return nil, err
	}

	if p != nil {
		o, err := p.GetTaskLogs(tasklog.Input{
			PodName:         sj.Name,
			Namespace:       sj.Namespace,
			TaskExecutionID: taskExecID,
		})
		if err != nil {
			return nil, err
		}

		// "All user" logs are shown already in the queuing and initializing phase.
		for _, log := range o.TaskLogs {
			log.ShowWhilePending = true
		}

		taskLogs = append(taskLogs, o.TaskLogs...)
	}

	customInfoMap := make(map[string]string)

	// Spark UI.
	if sj.Status.AppState.State == sparkOp.ApplicationStateFailed || sj.Status.AppState.State == sparkOp.ApplicationStateCompleted {
		if sj.Status.SparkApplicationID != "" && GetSparkConfig().SparkHistoryServerURL != "" {
			customInfoMap[sparkHistoryUI] = fmt.Sprintf("%s/history/%s", GetSparkConfig().SparkHistoryServerURL, sj.Status.SparkApplicationID)
			// Custom doesn't work unless the UI has a custom plugin to parse this, hence add to Logs as well.
			taskLogs = append(taskLogs, &core.TaskLog{
				Uri:           customInfoMap[sparkHistoryUI],
				Name:          "Spark History UI",
				Ready:         true,
				MessageFormat: core.TaskLog_JSON,
				LinkType:      core.TaskLog_DASHBOARD,
			})
		}
	} else if sj.Status.AppState.State == sparkOp.ApplicationStateRunning && sj.Status.DriverInfo.WebUIIngressAddress != "" {
		// Older versions of spark-operator does not append http:// but newer versions do.
		uri := sj.Status.DriverInfo.WebUIIngressAddress
		if !strings.HasPrefix(uri, "https://") && !strings.HasPrefix(uri, "http://") {
			uri = fmt.Sprintf("https://%s", uri)
		}
		customInfoMap[sparkDriverUI] = uri

		// Custom doesn't work unless the UI has a custom plugin to parse this, hence add to Logs as well.
		taskLogs = append(taskLogs, &core.TaskLog{
			Uri:           customInfoMap[sparkDriverUI],
			Name:          "Spark Driver UI",
			Ready:         true,
			MessageFormat: core.TaskLog_JSON,
			LinkType:      core.TaskLog_DASHBOARD,
		})
	}

	customInfo, err := utils.MarshalObjToStruct(customInfoMap) //nolint: staticcheck
	if err != nil {
		return nil, err
	}

	return &pluginsCore.TaskInfo{
		Logs:       taskLogs,
		LogContext: logCtx,
		CustomInfo: customInfo,
	}, nil
}

func (sparkResourceHandler) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	app := resource.(*sparkOp.SparkApplication)
	info, err := getEventInfoForSpark(ctx, pluginContext, app)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	phaseInfo, err := flytek8s.DemystifyFailedOrPendingPod(ctx, pluginContext, *info, app.Namespace, app.Status.DriverInfo.PodName, defaultDriverPrimaryContainerName)
	if err != nil {
		logger.Errorf(ctx, "Failed to demystify pod status for spark driver. Error: %v", err)
	}
	if phaseInfo.Phase().IsFailure() {
		// If the spark driver pod is in a failure state, we can fail fast without checking the SparkJob status.
		return phaseInfo, nil
	}
	occurredAt := time.Now()
	switch app.Status.AppState.State {
	case sparkOp.ApplicationStateNew:
		phaseInfo = pluginsCore.PhaseInfoQueuedWithTaskInfo(occurredAt, pluginsCore.DefaultPhaseVersion, "job queued", info)
	case sparkOp.ApplicationStateSubmitted, applicationStatePendingSubmission:
		phaseInfo = pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion, "job submitted", info)
	case sparkOp.ApplicationStateFailedSubmission:
		reason := fmt.Sprintf("Spark Job  Submission Failed with Error: %s", app.Status.AppState.ErrorMessage)
		phaseInfo = pluginsCore.PhaseInfoRetryableFailure(errors.DownstreamSystemError, reason, info)
	case sparkOp.ApplicationStateFailed:
		reason := fmt.Sprintf("Spark Job Failed with Error: %s", app.Status.AppState.ErrorMessage)
		phaseInfo = pluginsCore.PhaseInfoRetryableFailure(errors.DownstreamSystemError, reason, info)
	case sparkOp.ApplicationStateCompleted:
		phaseInfo = pluginsCore.PhaseInfoSuccess(info)
	default:
		phaseInfo = pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, info)
	}

	for _, tl := range info.Logs {
		// TODO: Add readiness probe for spark driver pod. Need to upgrade spark-operator client version.
		if tl != nil && tl.LinkType == core.TaskLog_DASHBOARD && strings.Contains(tl.Name, "Spark Driver UI") {
			if phaseInfo.Phase() != pluginsCore.PhaseRunning {
				tl.Ready = false
				phaseInfo.WithReason("Spark driver UI is not ready")
			} else {
				tl.Ready = true
				phaseInfo.WithReason("Spark driver UI is ready")
			}
		} else if tl != nil && tl.LinkType == core.TaskLog_IDE {
			if phaseInfo.Phase() != pluginsCore.PhaseRunning {
				phaseInfo.WithReason("Vscode server is not ready")
			} else {
				phaseInfo.WithReason("Vscode server is ready")
			}
		}
	}

	phaseVersionUpdateErr := k8s.MaybeUpdatePhaseVersionFromPluginContext(&phaseInfo, &pluginContext)
	if phaseVersionUpdateErr != nil {
		return phaseInfo, phaseVersionUpdateErr
	}

	return phaseInfo, nil
}

// IsTerminal returns true if the SparkApplication is in a terminal state (Completed, Failed, or FailedSubmission)
func (sparkResourceHandler) IsTerminal(_ context.Context, resource client.Object) (bool, error) {
	app, ok := resource.(*sparkOp.SparkApplication)
	if !ok {
		return false, fmt.Errorf("unexpected resource type: expected *SparkApplication, got %T", resource)
	}
	state := app.Status.AppState.State
	return state == sparkOp.ApplicationStateCompleted || state == sparkOp.ApplicationStateFailed || state == sparkOp.ApplicationStateFailedSubmission, nil
}

// GetCompletionTime returns the termination time of the SparkApplication
func (sparkResourceHandler) GetCompletionTime(resource client.Object) (time.Time, error) {
	app, ok := resource.(*sparkOp.SparkApplication)
	if !ok {
		return time.Time{}, fmt.Errorf("unexpected resource type: expected *SparkApplication, got %T", resource)
	}

	if !app.Status.TerminationTime.IsZero() {
		return app.Status.TerminationTime.Time, nil
	}

	// Fallback to submission time or creation time
	if !app.Status.LastSubmissionAttemptTime.IsZero() {
		return app.Status.LastSubmissionAttemptTime.Time, nil
	}

	return app.CreationTimestamp.Time, nil
}

func init() {
	// Do not add the types to client-go's global scheme.Scheme: binaries that also link the legacy
	// GoogleCloudPlatform spark-on-k8s-operator client would panic on registering a second Go type for
	// the same GVK. The executor builds its scheme from the plugin registry entry below.
	pluginmachinery.PluginRegistry().RegisterScheme(sparkTaskType, sparkOp.AddToScheme)

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  sparkTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{sparkTaskType},
			ResourceToWatch:     &sparkOp.SparkApplication{},
			Plugin:              sparkResourceHandler{},
			IsDefault:           false,
		})
}

func strPtr(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

func intPtr(val int32) *int32 {
	if val == 0 {
		return nil
	}
	return &val
}
