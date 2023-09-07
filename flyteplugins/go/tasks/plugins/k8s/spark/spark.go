package spark

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"strconv"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/template"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"k8s.io/client-go/kubernetes/scheme"

	sparkOp "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"regexp"
	"strings"
	"time"
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

	annotations := utils.UnionMaps(config.GetK8sPluginConfig().DefaultAnnotations, utils.CopyMap(taskCtx.TaskExecutionMetadata().GetAnnotations()))
	labels := utils.UnionMaps(config.GetK8sPluginConfig().DefaultLabels, utils.CopyMap(taskCtx.TaskExecutionMetadata().GetLabels()))
	container := taskTemplate.GetContainer()

	envVars := flytek8s.DecorateEnvVars(ctx, flytek8s.ToK8sEnvVar(container.GetEnv()),
		taskCtx.TaskExecutionMetadata().GetEnvironmentVariables(), taskCtx.TaskExecutionMetadata().GetTaskExecutionID())

	sparkEnvVars := make(map[string]string)
	for _, envVar := range envVars {
		sparkEnvVars[envVar.Name] = envVar.Value
	}

	sparkEnvVars["FLYTE_MAX_ATTEMPTS"] = strconv.Itoa(int(taskCtx.TaskExecutionMetadata().GetMaxAttempts()))

	serviceAccountName := flytek8s.GetServiceAccountNameFromTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())

	if len(serviceAccountName) == 0 {
		serviceAccountName = sparkTaskType
	}
	driverSpec := sparkOp.DriverSpec{
		SparkPodSpec: sparkOp.SparkPodSpec{
			Affinity:         config.GetK8sPluginConfig().DefaultAffinity,
			Annotations:      annotations,
			Labels:           labels,
			EnvVars:          sparkEnvVars,
			Image:            &container.Image,
			SecurityContenxt: config.GetK8sPluginConfig().DefaultPodSecurityContext.DeepCopy(),
			DNSConfig:        config.GetK8sPluginConfig().DefaultPodDNSConfig.DeepCopy(),
			Tolerations:      config.GetK8sPluginConfig().DefaultTolerations,
			SchedulerName:    &config.GetK8sPluginConfig().SchedulerName,
			NodeSelector:     config.GetK8sPluginConfig().DefaultNodeSelector,
			HostNetwork:      config.GetK8sPluginConfig().EnableHostNetworkingPod,
		},
		ServiceAccount: &serviceAccountName,
	}

	executorSpec := sparkOp.ExecutorSpec{
		SparkPodSpec: sparkOp.SparkPodSpec{
			Affinity:         config.GetK8sPluginConfig().DefaultAffinity.DeepCopy(),
			Annotations:      annotations,
			Labels:           labels,
			Image:            &container.Image,
			EnvVars:          sparkEnvVars,
			SecurityContenxt: config.GetK8sPluginConfig().DefaultPodSecurityContext.DeepCopy(),
			DNSConfig:        config.GetK8sPluginConfig().DefaultPodDNSConfig.DeepCopy(),
			Tolerations:      config.GetK8sPluginConfig().DefaultTolerations,
			SchedulerName:    &config.GetK8sPluginConfig().SchedulerName,
			NodeSelector:     config.GetK8sPluginConfig().DefaultNodeSelector,
			HostNetwork:      config.GetK8sPluginConfig().EnableHostNetworkingPod,
		},
	}

	modifiedArgs, err := template.Render(ctx, container.GetArgs(), template.Parameters{
		TaskExecMetadata: taskCtx.TaskExecutionMetadata(),
		Inputs:           taskCtx.InputReader(),
		OutputPath:       taskCtx.OutputWriter(),
		Task:             taskCtx.TaskReader(),
	})
	if err != nil {
		return nil, err
	}

	// Hack: Retry submit failures in-case of resource limits hit.
	submissionFailureRetries := int32(14)
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
	if len(sparkConfig["spark.kubernetes.driver.limit.cores"]) == 0 {
		// spark.kubernetes.driver.request.cores takes precedence over spark.driver.cores
		if len(sparkConfig["spark.kubernetes.driver.request.cores"]) != 0 {
			sparkConfig["spark.kubernetes.driver.limit.cores"] = sparkConfig["spark.kubernetes.driver.request.cores"]
		} else if len(sparkConfig["spark.driver.cores"]) != 0 {
			sparkConfig["spark.kubernetes.driver.limit.cores"] = sparkConfig["spark.driver.cores"]
		}
	}

	if len(sparkConfig["spark.kubernetes.executor.limit.cores"]) == 0 {
		// spark.kubernetes.executor.request.cores takes precedence over spark.executor.cores
		if len(sparkConfig["spark.kubernetes.executor.request.cores"]) != 0 {
			sparkConfig["spark.kubernetes.executor.limit.cores"] = sparkConfig["spark.kubernetes.executor.request.cores"]
		} else if len(sparkConfig["spark.executor.cores"]) != 0 {
			sparkConfig["spark.kubernetes.executor.limit.cores"] = sparkConfig["spark.executor.cores"]
		}
	}

	sparkConfig["spark.kubernetes.executor.podNamePrefix"] = taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	sparkConfig["spark.kubernetes.driverEnv.FLYTE_START_TIME"] = strconv.FormatInt(time.Now().UnixNano()/1000000, 10)

	// Add driver/executor defaults to CRD Driver/Executor Spec as well.
	cores, err := strconv.ParseInt(sparkConfig["spark.driver.cores"], 10, 32)
	if err == nil {
		driverSpec.Cores = intPtr(int32(cores))
	}
	driverSpec.Memory = strPtr(sparkConfig["spark.driver.memory"])

	execCores, err := strconv.ParseInt(sparkConfig["spark.executor.cores"], 10, 32)
	if err == nil {
		executorSpec.Cores = intPtr(int32(execCores))
	}

	execCount, err := strconv.ParseInt(sparkConfig["spark.executor.instances"], 10, 32)
	if err == nil {
		executorSpec.Instances = intPtr(int32(execCount))
	}
	executorSpec.Memory = strPtr(sparkConfig["spark.executor.memory"])

	j := &sparkOp.SparkApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindSparkApplication,
			APIVersion: sparkOp.SchemeGroupVersion.String(),
		},
		Spec: sparkOp.SparkApplicationSpec{
			ServiceAccount: &serviceAccountName,
			Type:           getApplicationType(sparkJob.GetApplicationType()),
			Image:          &container.Image,
			Arguments:      modifiedArgs,
			Driver:         driverSpec,
			Executor:       executorSpec,
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
		j.Spec.BatchScheduler = &val
	}

	if sparkJob.MainApplicationFile != "" {
		j.Spec.MainApplicationFile = &sparkJob.MainApplicationFile
	}
	if sparkJob.MainClass != "" {
		j.Spec.MainClass = &sparkJob.MainClass
	}

	// Spark driver pods should always run as non-interruptible. As such, we hardcode
	// `interruptible=false` to explicitly add non-interruptible node selector
	// requirements to the driver pods
	flytek8s.ApplyInterruptibleNodeSelectorRequirement(false, j.Spec.Driver.Affinity)

	// Add Interruptible Tolerations/NodeSelector to only Executor pods.
	// The Interruptible NodeSelector takes precedence over the DefaultNodeSelector
	if taskCtx.TaskExecutionMetadata().IsInterruptible() {
		j.Spec.Executor.Tolerations = append(j.Spec.Executor.Tolerations, config.GetK8sPluginConfig().InterruptibleTolerations...)
		j.Spec.Executor.NodeSelector = config.GetK8sPluginConfig().InterruptibleNodeSelector
	}

	// Add interruptible/non-interruptible node selector requirements to executor pod
	flytek8s.ApplyInterruptibleNodeSelectorRequirement(taskCtx.TaskExecutionMetadata().IsInterruptible(), j.Spec.Executor.Affinity)

	return j, nil
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
	state := sj.Status.AppState.State
	isQueued := state == sparkOp.NewState ||
		state == sparkOp.PendingSubmissionState ||
		state == sparkOp.SubmittedState

	sparkConfig := GetSparkConfig()
	taskLogs := make([]*core.TaskLog, 0, 3)
	taskExecID := pluginContext.TaskExecutionMetadata().GetTaskExecutionID().GetID()

	if !isQueued {
		if sj.Status.DriverInfo.PodName != "" {
			p, err := logs.InitializeLogPlugins(&sparkConfig.LogConfig.Mixed)
			if err != nil {
				return nil, err
			}

			if p != nil {
				o, err := p.GetTaskLogs(tasklog.Input{
					PodName:                 sj.Status.DriverInfo.PodName,
					Namespace:               sj.Namespace,
					LogName:                 "(Driver Logs)",
					TaskExecutionIdentifier: &taskExecID,
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
				PodName:                 sj.Status.DriverInfo.PodName,
				Namespace:               sj.Namespace,
				LogName:                 "(User Logs)",
				TaskExecutionIdentifier: &taskExecID,
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
				PodName:                 sj.Name,
				Namespace:               sj.Namespace,
				LogName:                 "(System Logs)",
				TaskExecutionIdentifier: &taskExecID,
			})

			if err != nil {
				return nil, err
			}

			taskLogs = append(taskLogs, o.TaskLogs...)
		}
	}

	p, err := logs.InitializeLogPlugins(&sparkConfig.LogConfig.AllUser)
	if err != nil {
		return nil, err
	}

	if p != nil {
		o, err := p.GetTaskLogs(tasklog.Input{
			PodName:                 sj.Name,
			Namespace:               sj.Namespace,
			LogName:                 "(Spark-Submit/All User Logs)",
			TaskExecutionIdentifier: &taskExecID,
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
	switch app.Status.AppState.State {
	case sparkOp.NewState:
		return pluginsCore.PhaseInfoQueued(occurredAt, pluginsCore.DefaultPhaseVersion, "job queued"), nil
	case sparkOp.SubmittedState, sparkOp.PendingSubmissionState:
		return pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion, "job submitted", info), nil
	case sparkOp.FailedSubmissionState:
		reason := fmt.Sprintf("Spark Job  Submission Failed with Error: %s", app.Status.AppState.ErrorMessage)
		return pluginsCore.PhaseInfoRetryableFailure(errors.DownstreamSystemError, reason, info), nil
	case sparkOp.FailedState:
		reason := fmt.Sprintf("Spark Job Failed with Error: %s", app.Status.AppState.ErrorMessage)
		return pluginsCore.PhaseInfoRetryableFailure(errors.DownstreamSystemError, reason, info), nil
	case sparkOp.CompletedState:
		return pluginsCore.PhaseInfoSuccess(info), nil
	}
	return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, info), nil
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
