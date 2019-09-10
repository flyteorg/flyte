package k8splugins

import (
	"context"
	"fmt"

	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s/config"
	"github.com/lyft/flyteplugins/go/tasks/v1/logs"

	v1 "github.com/lyft/flyteplugins/go/tasks/v1"
	"github.com/lyft/flyteplugins/go/tasks/v1/events"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"

	"k8s.io/client-go/kubernetes/scheme"

	sparkOp "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta1"
	logUtils "github.com/lyft/flyteidl/clients/go/coreutils/logs"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flyteplugins/go/tasks/v1/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"

	pluginsConfig "github.com/lyft/flyteplugins/go/tasks/v1/config"
)

const KindSparkApplication = "SparkApplication"
const sparkDriverUI = "sparkDriverUI"
const sparkHistoryUI = "sparkHistoryUI"

var sparkTaskType = "spark"

// Spark-specific configs
type SparkConfig struct {
	DefaultSparkConfig    map[string]string `json:"spark-config-default" pflag:",Key value pairs of default spark configuration that should be applied to every SparkJob"`
	SparkHistoryServerURL string            `json:"spark-history-server-url" pflag:",URL for SparkHistory Server that each job will publish the execution history to."`
}

var (
	sparkConfigSection = pluginsConfig.MustRegisterSubSection("spark", &SparkConfig{})
)

func GetSparkConfig() *SparkConfig {
	return sparkConfigSection.GetConfig().(*SparkConfig)
}

// This method should be used for unit testing only
func setSparkConfig(cfg *SparkConfig) error {
	return sparkConfigSection.SetConfig(cfg)
}

type sparkResourceHandler struct {
}

func (sparkResourceHandler) GetProperties() types.ExecutorProperties {
	return types.ExecutorProperties{}
}

// Creates a new Job that will execute the main container as well as any generated types the result from the execution.
func (sparkResourceHandler) BuildResource(ctx context.Context, taskCtx types.TaskContext, task *core.TaskTemplate, inputs *core.LiteralMap) (flytek8s.K8sResource, error) {

	sparkJob := plugins.SparkJob{}
	err := utils.UnmarshalStruct(task.GetCustom(), &sparkJob)
	if err != nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", task.GetCustom(), err.Error())
	}

	annotations := flytek8s.UnionMaps(config.GetK8sPluginConfig().DefaultAnnotations, utils.CopyMap(taskCtx.GetAnnotations()))
	labels := flytek8s.UnionMaps(config.GetK8sPluginConfig().DefaultLabels, utils.CopyMap(taskCtx.GetLabels()))
	container := task.GetContainer()

	envVars := flytek8s.DecorateEnvVars(ctx, flytek8s.ToK8sEnvVar(task.GetContainer().GetEnv()), taskCtx.GetTaskExecutionID())

	sparkEnvVars := make(map[string]string)
	for _, envVar := range envVars {
		sparkEnvVars[envVar.Name] = envVar.Value
	}

	driverSpec := sparkOp.DriverSpec{
		SparkPodSpec: sparkOp.SparkPodSpec{
			Annotations: annotations,
			Labels:      labels,
			EnvVars:     sparkEnvVars,
			Image:       &container.Image,
		},
		ServiceAccount: &sparkTaskType,
	}

	executorSpec := sparkOp.ExecutorSpec{
		SparkPodSpec: sparkOp.SparkPodSpec{
			Annotations: annotations,
			Labels:      labels,
			Image:       &container.Image,
			EnvVars:     sparkEnvVars,
		},
	}

	modifiedArgs, err := utils.ReplaceTemplateCommandArgs(context.TODO(),
		task.GetContainer().GetArgs(),
		utils.CommandLineTemplateArgs{
			Input:        taskCtx.GetInputsFile().String(),
			OutputPrefix: taskCtx.GetDataDir().String(),
			Inputs:       utils.LiteralMapToTemplateArgs(context.TODO(), inputs),
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
		sparkConfig[k] = v
	}

	// Set pod limits.
	if sparkConfig["spark.kubernetes.driver.limit.cores"] == "" && sparkConfig["spark.driver.cores"] != "" {
		sparkConfig["spark.kubernetes.driver.limit.cores"] = sparkConfig["spark.driver.cores"]
	}
	if sparkConfig["spark.kubernetes.executor.limit.cores"] == "" && sparkConfig["spark.executor.cores"] != "" {
		sparkConfig["spark.kubernetes.executor.limit.cores"] = sparkConfig["spark.executor.cores"]
	}

	j := &sparkOp.SparkApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindSparkApplication,
			APIVersion: sparkOp.SchemeGroupVersion.String(),
		},
		Spec: sparkOp.SparkApplicationSpec{
			ServiceAccount: &sparkTaskType,
			Type:           getApplicationType(sparkJob.GetApplicationType()),
			Mode:           sparkOp.ClusterMode,
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

	if sparkJob.MainApplicationFile != "" {
		j.Spec.MainApplicationFile = &sparkJob.MainApplicationFile
	} else if sparkJob.MainClass != "" {
		j.Spec.MainClass = &sparkJob.MainClass
	}

	return j, nil
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

func (sparkResourceHandler) BuildIdentityResource(ctx context.Context, taskCtx types.TaskContext) (flytek8s.K8sResource, error) {
	return &sparkOp.SparkApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindSparkApplication,
			APIVersion: sparkOp.SchemeGroupVersion.String(),
		},
	}, nil
}

func getEventInfoForSpark(sj *sparkOp.SparkApplication) (*events.TaskEventInfo, error) {
	var taskLogs []*core.TaskLog
	customInfoMap := make(map[string]string)

	logConfig := logs.GetLogConfig()
	if logConfig.IsKubernetesEnabled && sj.Status.DriverInfo.PodName != "" {
		k8sLog, err := logUtils.NewKubernetesLogPlugin(logConfig.KubernetesURL).GetTaskLog(
			sj.Status.DriverInfo.PodName,
			sj.Namespace,
			"",
			"",
			"Driver Logs (via Kubernetes)")

		if err != nil {
			return nil, err
		}
		taskLogs = append(taskLogs, &k8sLog)
	}

	if logConfig.IsCloudwatchEnabled {
		cwUserLogs := core.TaskLog{
			Uri: fmt.Sprintf(
				"https://console.aws.amazon.com/cloudwatch/home?region=%s#logStream:group=%s;prefix=var.log.containers.%s;streamFilter=typeLogStreamPrefix",
				logConfig.CloudwatchRegion,
				logConfig.CloudwatchLogGroup,
				sj.Status.DriverInfo.PodName),
			Name:          "User Driver Logs (via Cloudwatch)",
			MessageFormat: core.TaskLog_JSON,
		}
		cwSystemLogs := core.TaskLog{
			Uri: fmt.Sprintf(
				"https://console.aws.amazon.com/cloudwatch/home?region=%s#logStream:group=%s;prefix=system_log.var.log.containers.%s;streamFilter=typeLogStreamPrefix",
				logConfig.CloudwatchRegion,
				logConfig.CloudwatchLogGroup,
				sj.Name),
			Name:          "System Logs (via Cloudwatch)",
			MessageFormat: core.TaskLog_JSON,
		}
		taskLogs = append(taskLogs, &cwUserLogs)
		taskLogs = append(taskLogs, &cwSystemLogs)

	}

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
		// Append https as the operator doesn't currently.
		customInfoMap[sparkDriverUI] = fmt.Sprintf("https://%s", sj.Status.DriverInfo.WebUIIngressAddress)
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

	return &events.TaskEventInfo{
		Logs:       taskLogs,
		CustomInfo: customInfo,
	}, nil
}

func (sparkResourceHandler) GetTaskStatus(_ context.Context, _ types.TaskContext, r flytek8s.K8sResource) (
	types.TaskStatus, *events.TaskEventInfo, error) {

	app := r.(*sparkOp.SparkApplication)
	var status types.TaskStatus
	switch app.Status.AppState.State {
	case sparkOp.NewState, sparkOp.SubmittedState, sparkOp.PendingSubmissionState:
		status = types.TaskStatusQueued
	case sparkOp.FailedSubmissionState:
		status = types.TaskStatusRetryableFailure(errors.Errorf(errors.DownstreamSystemError, "Spark Job  Submission Failed with Error: %s", app.Status.AppState.ErrorMessage))
	case sparkOp.FailedState:
		status = types.TaskStatusRetryableFailure(errors.Errorf(errors.DownstreamSystemError, "Spark Job Failed with Error: %s", app.Status.AppState.ErrorMessage))
	case sparkOp.CompletedState:
		status = types.TaskStatusSucceeded
	default:
		status = types.TaskStatusRunning
	}

	info, err := getEventInfoForSpark(app)
	if err != nil {
		return types.TaskStatusUndefined, nil, err
	}

	return status, info, nil
}

func (sparkResourceHandler) PopulateTaskEventInfo(taskCtx types.TaskContext, resource flytek8s.K8sResource) error {
	return nil
}

func init() {
	if err := sparkOp.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	v1.RegisterLoader(func(ctx context.Context) error {
		return v1.K8sRegisterForTaskTypes(sparkTaskType, &sparkOp.SparkApplication{},
			flytek8s.DefaultInformerResyncDuration, sparkResourceHandler{}, sparkTaskType)
	})
}
