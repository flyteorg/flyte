package spark

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	sparkOp "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	sparkOpConfig "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
)

const KindSparkApplication = "SparkApplication"
const sparkDriverUI = "sparkDriverUI"
const sparkHistoryUI = "sparkHistoryUI"

var featureRegex = regexp.MustCompile(`^spark.((flyteorg)|(flyte)).(.+).enabled$`)

var sparkTaskType = "spark"

type sparkResourceHandler struct {
}

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
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &sparkJob)
	if err != nil {
		return nil, errors.Wrapf(errors.BadTaskSpecification, err, "invalid TaskSpecification [%v], failed to unmarshal", taskTemplate.GetCustom())
	}

	if err = validateSparkJob(&sparkJob); err != nil {
		return nil, errors.Wrapf(errors.BadTaskSpecification, err, "invalid TaskSpecification [%v].", taskTemplate.GetCustom())
	}

	sparkConfig := getSparkConfig(taskCtx, &sparkJob)
	driverSpec, err := createDriverSpec(ctx, taskCtx, sparkConfig)
	if err != nil {
		return nil, err
	}
	executorSpec, err := createExecutorSpec(ctx, taskCtx, sparkConfig)
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
	if len(sparkConfig[sparkOpConfig.SparkDriverCoreLimitKey]) == 0 {
		// spark.kubernetes.driver.request.cores takes precedence over spark.driver.cores
		if len(sparkConfig[sparkOpConfig.SparkDriverCoreRequestKey]) != 0 {
			sparkConfig[sparkOpConfig.SparkDriverCoreLimitKey] = sparkConfig[sparkOpConfig.SparkDriverCoreRequestKey]
		} else if len(sparkConfig["spark.driver.cores"]) != 0 {
			sparkConfig[sparkOpConfig.SparkDriverCoreLimitKey] = sparkConfig["spark.driver.cores"]
		}
	}

	if len(sparkConfig[sparkOpConfig.SparkExecutorCoreLimitKey]) == 0 {
		// spark.kubernetes.executor.request.cores takes precedence over spark.executor.cores
		if len(sparkConfig[sparkOpConfig.SparkExecutorCoreRequestKey]) != 0 {
			sparkConfig[sparkOpConfig.SparkExecutorCoreLimitKey] = sparkConfig[sparkOpConfig.SparkExecutorCoreRequestKey]
		} else if len(sparkConfig["spark.executor.cores"]) != 0 {
			sparkConfig[sparkOpConfig.SparkExecutorCoreLimitKey] = sparkConfig["spark.executor.cores"]
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

func createSparkPodSpec(taskCtx pluginsCore.TaskExecutionContext, podSpec *v1.PodSpec, container *v1.Container) *sparkOp.SparkPodSpec {
	annotations := utils.UnionMaps(config.GetK8sPluginConfig().DefaultAnnotations, utils.CopyMap(taskCtx.TaskExecutionMetadata().GetAnnotations()))
	labels := utils.UnionMaps(config.GetK8sPluginConfig().DefaultLabels, utils.CopyMap(taskCtx.TaskExecutionMetadata().GetLabels()))

	sparkEnv := make([]v1.EnvVar, 0)
	for _, envVar := range container.Env {
		sparkEnv = append(sparkEnv, *envVar.DeepCopy())
	}
	sparkEnv = append(sparkEnv, v1.EnvVar{Name: "FLYTE_MAX_ATTEMPTS", Value: strconv.Itoa(int(taskCtx.TaskExecutionMetadata().GetMaxAttempts()))})

	spec := sparkOp.SparkPodSpec{
		Affinity:         podSpec.Affinity,
		Annotations:      annotations,
		Labels:           labels,
		Env:              sparkEnv,
		Image:            &container.Image,
		SecurityContenxt: podSpec.SecurityContext.DeepCopy(),
		DNSConfig:        podSpec.DNSConfig.DeepCopy(),
		Tolerations:      podSpec.Tolerations,
		SchedulerName:    &podSpec.SchedulerName,
		NodeSelector:     podSpec.NodeSelector,
		HostNetwork:      &podSpec.HostNetwork,
	}
	return &spec
}

type driverSpec struct {
	sparkSpec *sparkOp.DriverSpec
}

func createDriverSpec(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, sparkConfig map[string]string) (*driverSpec, error) {
	// Spark driver pods should always run as non-interruptible
	nonInterruptibleTaskCtx := flytek8s.NewPluginTaskExecutionContext(taskCtx, flytek8s.WithInterruptible(false))
	podSpec, _, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, nonInterruptibleTaskCtx)
	if err != nil {
		return nil, err
	}
	primaryContainer, err := flytek8s.GetContainer(podSpec, primaryContainerName)
	if err != nil {
		return nil, err
	}
	sparkPodSpec := createSparkPodSpec(nonInterruptibleTaskCtx, podSpec, primaryContainer)
	serviceAccountName := serviceAccountName(nonInterruptibleTaskCtx.TaskExecutionMetadata())
	spec := driverSpec{
		&sparkOp.DriverSpec{
			SparkPodSpec:   *sparkPodSpec,
			ServiceAccount: &serviceAccountName,
		},
	}
	if cores, err := strconv.ParseInt(sparkConfig["spark.driver.cores"], 10, 32); err == nil {
		spec.sparkSpec.Cores = intPtr(int32(cores))
	}
	spec.sparkSpec.Memory = strPtr(sparkConfig["spark.driver.memory"])
	return &spec, nil
}

type executorSpec struct {
	container          *v1.Container
	sparkSpec          *sparkOp.ExecutorSpec
	serviceAccountName string
}

func createExecutorSpec(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, sparkConfig map[string]string) (*executorSpec, error) {
	podSpec, _, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, err
	}
	primaryContainer, err := flytek8s.GetContainer(podSpec, primaryContainerName)
	if err != nil {
		return nil, err
	}
	sparkPodSpec := createSparkPodSpec(taskCtx, podSpec, primaryContainer)
	serviceAccountName := serviceAccountName(taskCtx.TaskExecutionMetadata())
	spec := executorSpec{
		primaryContainer,
		&sparkOp.ExecutorSpec{
			SparkPodSpec: *sparkPodSpec,
		},
		serviceAccountName,
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
	executorSpec *executorSpec) *sparkOp.SparkApplication {
	// Hack: Retry submit failures in-case of resource limits hit.
	submissionFailureRetries := int32(14)

	app := &sparkOp.SparkApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindSparkApplication,
			APIVersion: sparkOp.SchemeGroupVersion.String(),
		},
		Spec: sparkOp.SparkApplicationSpec{
			ServiceAccount: &executorSpec.serviceAccountName,
			Type:           getApplicationType(sparkJob.GetApplicationType()),
			Image:          &executorSpec.container.Image,
			Arguments:      executorSpec.container.Args,
			Driver:         *driverSpec.sparkSpec,
			Executor:       *executorSpec.sparkSpec,
			SparkConf:      sparkConfig,
			HadoopConf:     sparkJob.GetHadoopConf(),
			// SubmissionFailures handled here. Task Failures handled at Propeller/Job level.
			RestartPolicy: sparkOp.RestartPolicy{
				Type:                       sparkOp.OnFailure,
				OnSubmissionFailureRetries: &submissionFailureRetries,
			},
		},
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
		return sparkOp.PythonApplicationType
	case plugins.SparkApplication_JAVA:
		return sparkOp.JavaApplicationType
	case plugins.SparkApplication_SCALA:
		return sparkOp.ScalaApplicationType
	case plugins.SparkApplication_R:
		return sparkOp.RApplicationType
	}
	return sparkOp.PythonApplicationType
}

func (sparkResourceHandler) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &sparkOp.SparkApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindSparkApplication,
			APIVersion: sparkOp.SchemeGroupVersion.String(),
		},
	}, nil
}

func getEventInfoForSpark(pluginContext k8s.PluginContext, sj *sparkOp.SparkApplication) (*pluginsCore.TaskInfo, error) {
	sparkConfig := GetSparkConfig()
	taskLogs := make([]*core.TaskLog, 0, 3)
	taskExecID := pluginContext.TaskExecutionMetadata().GetTaskExecutionID()

	if sj.Status.DriverInfo.PodName != "" {
		p, err := logs.InitializeLogPlugins(&sparkConfig.LogConfig.Mixed)
		if err != nil {
			return nil, err
		}

		if p != nil {
			o, err := p.GetTaskLogs(tasklog.Input{
				PodName:         sj.Status.DriverInfo.PodName,
				Namespace:       sj.Namespace,
				LogName:         "(Driver Logs)",
				TaskExecutionID: taskExecID,
			})

			if err != nil {
				return nil, err
			}

			taskLogs = append(taskLogs, o.TaskLogs...)
		}
	}

	p, err := logs.InitializeLogPlugins(&sparkConfig.LogConfig.User)
	if err != nil {
		return nil, err
	}

	if p != nil {
		o, err := p.GetTaskLogs(tasklog.Input{
			PodName:         sj.Status.DriverInfo.PodName,
			Namespace:       sj.Namespace,
			LogName:         "(User Logs)",
			TaskExecutionID: taskExecID,
		})

		if err != nil {
			return nil, err
		}

		taskLogs = append(taskLogs, o.TaskLogs...)
	}

	p, err = logs.InitializeLogPlugins(&sparkConfig.LogConfig.System)
	if err != nil {
		return nil, err
	}

	if p != nil {
		o, err := p.GetTaskLogs(tasklog.Input{
			PodName:         sj.Name,
			Namespace:       sj.Namespace,
			LogName:         "(System Logs)",
			TaskExecutionID: taskExecID,
		})

		if err != nil {
			return nil, err
		}

		taskLogs = append(taskLogs, o.TaskLogs...)
	}

	p, err = logs.InitializeLogPlugins(&sparkConfig.LogConfig.AllUser)
	if err != nil {
		return nil, err
	}

	if p != nil {
		o, err := p.GetTaskLogs(tasklog.Input{
			PodName:         sj.Name,
			Namespace:       sj.Namespace,
			LogName:         "(Spark-Submit/All User Logs)",
			TaskExecutionID: taskExecID,
		})

		if err != nil {
			return nil, err
		}

		taskLogs = append(taskLogs, o.TaskLogs...)
	}

	customInfoMap := make(map[string]string)

	// Spark UI.
	if sj.Status.AppState.State == sparkOp.FailedState || sj.Status.AppState.State == sparkOp.CompletedState {
		if sj.Status.SparkApplicationID != "" && GetSparkConfig().SparkHistoryServerURL != "" {
			customInfoMap[sparkHistoryUI] = fmt.Sprintf("%s/history/%s", GetSparkConfig().SparkHistoryServerURL, sj.Status.SparkApplicationID)
			// Custom doesn't work unless the UI has a custom plugin to parse this, hence add to Logs as well.
			taskLogs = append(taskLogs, &core.TaskLog{
				Uri:           customInfoMap[sparkHistoryUI],
				Name:          "Spark History UI",
				MessageFormat: core.TaskLog_JSON,
			})
		}
	} else if sj.Status.AppState.State == sparkOp.RunningState && sj.Status.DriverInfo.WebUIIngressAddress != "" {
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
			MessageFormat: core.TaskLog_JSON,
		})
	}

	customInfo, err := utils.MarshalObjToStruct(customInfoMap)
	if err != nil {
		return nil, err
	}

	return &pluginsCore.TaskInfo{
		Logs:       taskLogs,
		CustomInfo: customInfo,
	}, nil
}

func (sparkResourceHandler) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {

	app := resource.(*sparkOp.SparkApplication)
	info, err := getEventInfoForSpark(pluginContext, app)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	occurredAt := time.Now()

	var phaseInfo pluginsCore.PhaseInfo

	switch app.Status.AppState.State {
	case sparkOp.NewState:
		phaseInfo = pluginsCore.PhaseInfoQueuedWithTaskInfo(occurredAt, pluginsCore.DefaultPhaseVersion, "job queued", info)
	case sparkOp.SubmittedState, sparkOp.PendingSubmissionState:
		phaseInfo = pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion, "job submitted", info)
	case sparkOp.FailedSubmissionState:
		reason := fmt.Sprintf("Spark Job  Submission Failed with Error: %s", app.Status.AppState.ErrorMessage)
		phaseInfo = pluginsCore.PhaseInfoRetryableFailure(errors.DownstreamSystemError, reason, info)
	case sparkOp.FailedState:
		reason := fmt.Sprintf("Spark Job Failed with Error: %s", app.Status.AppState.ErrorMessage)
		phaseInfo = pluginsCore.PhaseInfoRetryableFailure(errors.DownstreamSystemError, reason, info)
	case sparkOp.CompletedState:
		phaseInfo = pluginsCore.PhaseInfoSuccess(info)
	default:
		phaseInfo = pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, info)
	}

	phaseVersionUpdateErr := k8s.MaybeUpdatePhaseVersionFromPluginContext(&phaseInfo, &pluginContext)
	if phaseVersionUpdateErr != nil {
		return pluginsCore.PhaseInfoUndefined, phaseVersionUpdateErr
	}

	return phaseInfo, nil
}

func init() {
	if err := sparkOp.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

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
