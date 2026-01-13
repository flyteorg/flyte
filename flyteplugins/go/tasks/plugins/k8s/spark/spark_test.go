package spark

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	sj "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	sparkOp "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginIOMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8smocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	stdlibUtils "github.com/flyteorg/flyte/v2/flytestdlib/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

const sparkMainClass = "MainClass"
const sparkApplicationFile = "local:///spark_app.py"
const testImage = "image://"
const sparkUIAddress = "https://spark-ui.flyte"

var (
	dummySparkConf = map[string]string{
		"spark.driver.memory":             "200M",
		"spark.driver.cores":              "1",
		"spark.executor.cores":            "2",
		"spark.executor.instances":        "3",
		"spark.executor.memory":           "500M",
		"spark.flyte.feature1.enabled":    "true",
		"spark.flyteorg.feature2.enabled": "true",
		"spark.flyteorg.feature3.enabled": "true",
		"spark.batchScheduler":            "volcano",
	}

	dummyEnvVars = []*core.KeyValuePair{
		{Key: "Env_Var", Value: "Env_Val"},
	}

	dummyEnvVarsWithSecretRef = []corev1.EnvVar{
		{Name: "Env_Var", Value: "Env_Val"},
		{Name: "SECRET", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: "key",
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "secret-name",
				},
			},
		}},
	}

	testArgs = []string{
		"execute-spark-task",
	}
)

func TestGetApplicationType(t *testing.T) {
	assert.Equal(t, getApplicationType(plugins.SparkApplication_PYTHON), sj.PythonApplicationType)
	assert.Equal(t, getApplicationType(plugins.SparkApplication_R), sj.RApplicationType)
	assert.Equal(t, getApplicationType(plugins.SparkApplication_JAVA), sj.JavaApplicationType)
	assert.Equal(t, getApplicationType(plugins.SparkApplication_SCALA), sj.ScalaApplicationType)
}

func TestGetEventInfo(t *testing.T) {
	assert.NoError(t, setSparkConfig(&Config{
		LogConfig: LogConfig{
			User: logs.LogConfig{
				IsCloudwatchEnabled:   true,
				CloudwatchTemplateURI: "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.{{ .podName }};streamFilter=typeLogStreamPrefix",
				IsKubernetesEnabled:   true,
				KubernetesURL:         "k8s.com",
			},
			System: logs.LogConfig{
				IsCloudwatchEnabled:   true,
				CloudwatchTemplateURI: "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=system_log.var.log.containers.{{ .podName }};streamFilter=typeLogStreamPrefix",
			},
			AllUser: logs.LogConfig{
				IsCloudwatchEnabled:   true,
				CloudwatchTemplateURI: "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.{{ .podName }};streamFilter=typeLogStreamPrefix",
			},
			Mixed: logs.LogConfig{
				IsKubernetesEnabled: true,
				KubernetesURL:       "k8s.com",
			},
		},
	}))
	pluginContext := dummySparkPluginContext(dummySparkTaskTemplateContainer("blah-1", dummySparkConf), k8s.PluginState{})
	info, err := getEventInfoForSpark(context.TODO(), pluginContext, dummySparkApplication(sj.RunningState))
	assert.NoError(t, err)
	assert.Len(t, info.Logs, 6)
	assert.Equal(t, "https://spark-ui.flyte", info.CustomInfo.Fields[sparkDriverUI].GetStringValue())
	generatedLinks := make([]string, 0, len(info.Logs))
	for _, l := range info.Logs {
		generatedLinks = append(generatedLinks, l.Uri)
	}

	expectedLinks := []string{
		"k8s.com/#!/log/spark-namespace/spark-pod/pod?namespace=spark-namespace",
		"k8s.com/#!/log/spark-namespace/spark-pod/pod?namespace=spark-namespace",
		"https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.spark-pod;streamFilter=typeLogStreamPrefix",
		"https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=system_log.var.log.containers.spark-app-name;streamFilter=typeLogStreamPrefix",
		"https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.spark-app-name;streamFilter=typeLogStreamPrefix",
		"https://spark-ui.flyte",
	}

	assert.Equal(t, expectedLinks, generatedLinks)

	info, err = getEventInfoForSpark(context.TODO(), pluginContext, dummySparkApplication(sj.SubmittedState))
	generatedLinks = make([]string, 0, len(info.Logs))
	for _, l := range info.Logs {
		generatedLinks = append(generatedLinks, l.Uri)
	}
	assert.NoError(t, err)
	assert.Len(t, info.Logs, 5)
	assert.Equal(t, expectedLinks[:5], generatedLinks) // No Spark Driver UI for Submitted state
	assert.True(t, info.Logs[4].ShowWhilePending)      // All User Logs should be shown while pending
	generatedLinks = make([]string, 0, len(info.Logs))
	for _, l := range info.Logs {
		generatedLinks = append(generatedLinks, l.Uri)
	}
	assert.NoError(t, err)
	assert.Len(t, info.Logs, 5)
	assert.Equal(t, expectedLinks[:5], generatedLinks) // No Spark Driver UI for Submitted state
	assert.True(t, info.Logs[4].ShowWhilePending)      // All User Logs should be shown while pending

	assert.NoError(t, setSparkConfig(&Config{
		SparkHistoryServerURL: "spark-history.flyte",
		LogConfig: LogConfig{
			User: logs.LogConfig{
				IsCloudwatchEnabled:   true,
				CloudwatchTemplateURI: "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.{{ .podName }};streamFilter=typeLogStreamPrefix",
				IsKubernetesEnabled:   true,
				KubernetesURL:         "k8s.com",
			},
			System: logs.LogConfig{
				IsCloudwatchEnabled:   true,
				CloudwatchTemplateURI: "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=system_log.var.log.containers.{{ .podName }};streamFilter=typeLogStreamPrefix",
			},
			AllUser: logs.LogConfig{
				IsCloudwatchEnabled:   true,
				CloudwatchTemplateURI: "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.{{ .podName }};streamFilter=typeLogStreamPrefix",
			},
		},
	}))

	info, err = getEventInfoForSpark(context.TODO(), pluginContext, dummySparkApplication(sj.FailedState))
	assert.NoError(t, err)
	assert.Len(t, info.Logs, 5)
	assert.Equal(t, "spark-history.flyte/history/app-id", info.CustomInfo.Fields[sparkHistoryUI].GetStringValue())
	generatedLinks = make([]string, 0, len(info.Logs))
	for _, l := range info.Logs {
		generatedLinks = append(generatedLinks, l.Uri)
	}

	expectedLinks = []string{
		"k8s.com/#!/log/spark-namespace/spark-pod/pod?namespace=spark-namespace",
		"https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.spark-pod;streamFilter=typeLogStreamPrefix",
		"https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=system_log.var.log.containers.spark-app-name;streamFilter=typeLogStreamPrefix",
		"https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.spark-app-name;streamFilter=typeLogStreamPrefix",
		"spark-history.flyte/history/app-id",
	}

	assert.Equal(t, expectedLinks, generatedLinks)
}

func TestGetTaskPhase(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}
	expectedLogCtx := &core.LogContext{
		PrimaryPodName: "spark-pod",
		Pods: []*core.PodLogContext{
			{
				Namespace:            "spark-namespace",
				PodName:              "spark-pod",
				PrimaryContainerName: "spark-kubernetes-driver",
				Containers: []*core.ContainerContext{
					{
						ContainerName: "spark-kubernetes-driver",
					},
				},
			},
			{
				Namespace:            "spark-namespace",
				PodName:              "exec-pod-2",
				PrimaryContainerName: "spark-kubernetes-executor",
				Containers: []*core.ContainerContext{
					{
						ContainerName: "spark-kubernetes-executor",
					},
				},
			},
		},
	}

	ctx := context.TODO()
	pluginCtx := dummySparkPluginContext(dummySparkTaskTemplateContainer("", dummySparkConf), k8s.PluginState{})
	taskPhase, err := sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.NewState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseQueued)
	assert.NotNil(t, taskPhase.Info())
	assert.NotNil(t, taskPhase.Info().LogContext)
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.SubmittedState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseInitializing)
	assert.NotNil(t, taskPhase.Info())
	assert.NotNil(t, taskPhase.Info().LogContext)
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.RunningState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.Equal(t, expectedLogCtx, taskPhase.Info().LogContext)
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.CompletedState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseSuccess)
	assert.NotNil(t, taskPhase.Info())
	assert.Equal(t, expectedLogCtx, taskPhase.Info().LogContext)
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.InvalidatingState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.Equal(t, expectedLogCtx, taskPhase.Info().LogContext)
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.FailingState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.Equal(t, expectedLogCtx, taskPhase.Info().LogContext)
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.PendingRerunState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.Equal(t, expectedLogCtx, taskPhase.Info().LogContext)
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.SucceedingState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.Equal(t, expectedLogCtx, taskPhase.Info().LogContext)
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.FailedSubmissionState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRetryableFailure)
	assert.NotNil(t, taskPhase.Info())
	assert.Equal(t, expectedLogCtx, taskPhase.Info().LogContext)
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.FailedState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRetryableFailure)
	assert.NotNil(t, taskPhase.Info())
	assert.Equal(t, expectedLogCtx, taskPhase.Info().LogContext)
	assert.Nil(t, err)
}

func TestGetTaskPhaseIncreasePhaseVersion(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}
	ctx := context.TODO()

	pluginState := k8s.PluginState{
		Phase:        pluginsCore.PhaseInitializing,
		PhaseVersion: pluginsCore.DefaultPhaseVersion,
		Reason:       "task submitted to K8s",
	}

	pluginCtx := dummySparkPluginContext(dummySparkTaskTemplateContainer("", dummySparkConf), pluginState)
	taskPhase, err := sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.SubmittedState))

	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Version(), pluginsCore.DefaultPhaseVersion+1)
}

func dummySparkApplication(state sj.ApplicationStateType) *sj.SparkApplication {

	return &sj.SparkApplication{
		ObjectMeta: v1.ObjectMeta{
			Name:      "spark-app-name",
			Namespace: "spark-namespace",
		},
		Status: sj.SparkApplicationStatus{
			SparkApplicationID: "app-id",
			AppState: sj.ApplicationState{
				State: state,
			},
			DriverInfo: sj.DriverInfo{
				PodName:             "spark-pod",
				WebUIIngressAddress: sparkUIAddress,
			},
			ExecutionAttempts: 1,
			ExecutorState: map[string]sparkOp.ExecutorState{
				"exec-pod-1": sparkOp.ExecutorPendingState,
				"exec-pod-2": sparkOp.ExecutorRunningState,
			},
		},
	}
}

func dummySparkCustomObj(sparkConf map[string]string) *plugins.SparkJob {
	sparkJob := plugins.SparkJob{}

	sparkJob.MainClass = sparkMainClass
	sparkJob.MainApplicationFile = sparkApplicationFile
	sparkJob.SparkConf = sparkConf
	sparkJob.ApplicationType = plugins.SparkApplication_PYTHON
	return &sparkJob
}

func dummyPodSpec() *corev1.PodSpec {
	return &corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:  "init",
				Image: testImage,
				Args:  testArgs,
			},
		},
		Containers: []corev1.Container{
			{
				Name:  "primary",
				Image: testImage,
				Args:  testArgs,
				Env:   dummyEnvVarsWithSecretRef,
			},
			{
				Name:  "secondary",
				Image: testImage,
				Args:  testArgs,
				Env:   flytek8s.ToK8sEnvVar(dummyEnvVars),
			},
		},
	}
}

func dummySparkTaskTemplateContainer(id string, sparkConf map[string]string) *core.TaskTemplate {
	sparkJob := dummySparkCustomObj(sparkConf)
	sparkJobJSON, err := utils.MarshalToString(sparkJob)
	if err != nil {
		panic(err)
	}

	structObj := structpb.Struct{}

	err = stdlibUtils.UnmarshalStringToPb(sparkJobJSON, &structObj)
	if err != nil {
		panic(err)
	}

	return &core.TaskTemplate{
		Id:   &core.Identifier{Name: id},
		Type: "container",
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image: testImage,
				Args:  testArgs,
				Env:   dummyEnvVars,
			},
		},
		Custom: &structObj,
	}
}

func dummySparkTaskTemplatePod(id string, sparkConf map[string]string, podSpec *corev1.PodSpec) *core.TaskTemplate {
	sparkJob := dummySparkCustomObj(sparkConf)
	sparkJobJSON, err := utils.MarshalToString(sparkJob)
	if err != nil {
		panic(err)
	}

	structObj := structpb.Struct{}

	err = stdlibUtils.UnmarshalStringToPb(sparkJobJSON, &structObj)
	if err != nil {
		panic(err)
	}

	podSpecPb, err := utils.MarshalObjToStruct(podSpec)
	if err != nil {
		panic(err)
	}

	return &core.TaskTemplate{
		Id:   &core.Identifier{Name: id},
		Type: "k8s_pod",
		Target: &core.TaskTemplate_K8SPod{
			K8SPod: &core.K8SPod{
				PodSpec: podSpecPb,
			},
		},
		Config: map[string]string{
			flytek8s.PrimaryContainerKey: "primary",
		},
		Custom: &structObj,
	}
}

func dummySparkTaskContext(taskTemplate *core.TaskTemplate, interruptible bool) pluginsCore.TaskExecutionContext {
	taskCtx := &mocks.TaskExecutionContext{}
	inputReader := &pluginIOMocks.InputReader{}
	inputReader.EXPECT().GetInputPrefixPath().Return("/input/prefix")
	inputReader.EXPECT().GetInputPath().Return("/input")
	inputReader.EXPECT().Get(mock.Anything).Return(&core.LiteralMap{}, nil)
	taskCtx.EXPECT().InputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.EXPECT().GetOutputPath().Return("/data/outputs.pb")
	outputReader.EXPECT().GetOutputPrefixPath().Return("/data/")
	outputReader.EXPECT().GetRawOutputPrefix().Return("")
	outputReader.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	outputReader.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")

	taskCtx.EXPECT().OutputWriter().Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(taskTemplate, nil)
	taskCtx.EXPECT().TaskReader().Return(taskReader)

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetID().Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.EXPECT().GetGeneratedName().Return("some-acceptable-name")
	tID.EXPECT().GetUniqueNodeID().Return("an-unique-id")

	overrides := &mocks.TaskOverrides{}
	overrides.EXPECT().GetResources().Return(&corev1.ResourceRequirements{})
	// No support for GPUs, and consequently, ExtendedResources on Spark plugin.
	overrides.EXPECT().GetExtendedResources().Return(nil)
	overrides.EXPECT().GetPodTemplate().Return(nil)
	overrides.EXPECT().GetContainerImage().Return("")

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.EXPECT().GetTaskExecutionID().Return(tID)
	taskExecutionMetadata.EXPECT().GetNamespace().Return("test-namespace")
	taskExecutionMetadata.EXPECT().GetAnnotations().Return(map[string]string{"annotation-1": "val1"})
	taskExecutionMetadata.EXPECT().GetLabels().Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.EXPECT().GetOwnerReference().Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.EXPECT().GetSecurityContext().Return(core.SecurityContext{
		RunAs: &core.Identity{K8SServiceAccount: "new-val"},
	})
	taskExecutionMetadata.EXPECT().IsInterruptible().Return(interruptible)
	taskExecutionMetadata.EXPECT().GetMaxAttempts().Return(uint32(1))
	taskExecutionMetadata.EXPECT().GetEnvironmentVariables().Return(nil)
	taskExecutionMetadata.EXPECT().GetPlatformResources().Return(nil)
	taskExecutionMetadata.EXPECT().GetOverrides().Return(overrides)
	taskExecutionMetadata.EXPECT().GetK8sServiceAccount().Return("new-val")
	taskExecutionMetadata.EXPECT().GetConsoleURL().Return("")
	taskCtx.EXPECT().TaskExecutionMetadata().Return(taskExecutionMetadata)
	pluginStateReaderMock := mocks.PluginStateReader{}
	pluginStateReaderMock.EXPECT().Get(mock.AnythingOfType(reflect.TypeOf(&k8s.PluginState{}).String())).RunAndReturn(
		func(v interface{}) (uint8, error) {
			*(v.(*k8s.PluginState)) = k8s.PluginState{}
			return 0, nil
		})

	taskCtx.EXPECT().PluginStateReader().Return(&pluginStateReaderMock)
	return taskCtx
}

func dummySparkPluginContext(taskTemplate *core.TaskTemplate, pluginState k8s.PluginState) k8s.PluginContext {
	return dummySparkPluginContextWithPods(taskTemplate, pluginState)
}

func dummySparkPluginContextWithPods(taskTemplate *core.TaskTemplate, pluginState k8s.PluginState, pods ...client.Object) k8s.PluginContext {
	pCtx := &k8smocks.PluginContext{}
	inputReader := &pluginIOMocks.InputReader{}
	inputReader.EXPECT().GetInputPrefixPath().Return("/input/prefix")
	inputReader.EXPECT().GetInputPath().Return("/input")
	inputReader.EXPECT().Get(mock.Anything).Return(&core.LiteralMap{}, nil)
	pCtx.EXPECT().InputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.EXPECT().GetOutputPath().Return("/data/outputs.pb")
	outputReader.EXPECT().GetOutputPrefixPath().Return("/data/")
	outputReader.EXPECT().GetRawOutputPrefix().Return("")
	outputReader.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	outputReader.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")

	pCtx.EXPECT().OutputWriter().Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(taskTemplate, nil)
	pCtx.EXPECT().TaskReader().Return(taskReader)

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetID().Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.EXPECT().GetGeneratedName().Return("some-acceptable-name")
	tID.EXPECT().GetUniqueNodeID().Return("an-unique-id")

	overrides := &mocks.TaskOverrides{}
	overrides.EXPECT().GetResources().Return(&corev1.ResourceRequirements{})
	// No support for GPUs, and consequently, ExtendedResources on Spark plugin.
	overrides.EXPECT().GetExtendedResources().Return(nil)
	overrides.EXPECT().GetContainerImage().Return("")

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.EXPECT().GetTaskExecutionID().Return(tID)
	taskExecutionMetadata.EXPECT().GetNamespace().Return("test-namespace")
	taskExecutionMetadata.EXPECT().GetAnnotations().Return(map[string]string{"annotation-1": "val1"})
	taskExecutionMetadata.EXPECT().GetLabels().Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.EXPECT().GetOwnerReference().Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.EXPECT().GetSecurityContext().Return(core.SecurityContext{
		RunAs: &core.Identity{K8SServiceAccount: "new-val"},
	})
	taskExecutionMetadata.EXPECT().IsInterruptible().Return(false)
	taskExecutionMetadata.EXPECT().GetMaxAttempts().Return(uint32(1))
	taskExecutionMetadata.EXPECT().GetEnvironmentVariables().Return(nil)
	taskExecutionMetadata.EXPECT().GetPlatformResources().Return(nil)
	taskExecutionMetadata.EXPECT().GetOverrides().Return(overrides)
	taskExecutionMetadata.EXPECT().GetK8sServiceAccount().Return("new-val")
	taskExecutionMetadata.EXPECT().GetConsoleURL().Return("")
	pCtx.EXPECT().TaskExecutionMetadata().Return(taskExecutionMetadata)

	pluginStateReaderMock := mocks.PluginStateReader{}
	pluginStateReaderMock.EXPECT().Get(mock.AnythingOfType(reflect.TypeOf(&pluginState).String())).RunAndReturn(
		func(v interface{}) (uint8, error) {
			*(v.(*k8s.PluginState)) = pluginState
			return 0, nil
		})

	// Add K8sReader mock for pods
	objs := make([]client.Object, len(pods))
	copy(objs, pods)
	reader := fake.NewClientBuilder().WithObjects(objs...).Build()
	pCtx.EXPECT().K8sReader().Return(reader)

	pCtx.EXPECT().PluginStateReader().Return(&pluginStateReaderMock)
	return pCtx
}

func defaultPluginConfig() *config.K8sPluginConfig {
	// Set Interruptible Config
	runAsUser := int64(1000)
	dnsOptVal1 := "1"
	dnsOptVal2 := "1"
	dnsOptVal3 := "3"

	// Set scheduler
	schedulerName := "custom-scheduler"

	// Node selectors
	defaultNodeSelector := map[string]string{
		"x/default": "true",
	}
	interruptibleNodeSelector := map[string]string{
		"x/interruptible": "true",
	}

	defaultPodHostNetwork := true

	// Default env vars passed explicitly and default env vars derived from environment
	defaultEnvVars := make(map[string]string)
	defaultEnvVars["foo"] = "bar"

	defaultEnvVarsFromEnv := make(map[string]string)
	targetKeyFromEnv := "TEST_VAR_FROM_ENV_KEY"
	targetValueFromEnv := "TEST_VAR_FROM_ENV_VALUE"
	os.Setenv(targetKeyFromEnv, targetValueFromEnv)
	defer os.Unsetenv(targetKeyFromEnv)
	defaultEnvVarsFromEnv["fooEnv"] = targetKeyFromEnv

	// Default affinity/anti-affinity
	defaultAffinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "x/default",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"true"},
							},
						},
					},
				},
			},
		},
	}

	// Interruptible/non-interruptible nodeselector requirement
	interruptibleNodeSelectorRequirement := &corev1.NodeSelectorRequirement{
		Key:      "x/interruptible",
		Operator: corev1.NodeSelectorOpIn,
		Values:   []string{"true"},
	}

	nonInterruptibleNodeSelectorRequirement := &corev1.NodeSelectorRequirement{
		Key:      "x/non-interruptible",
		Operator: corev1.NodeSelectorOpIn,
		Values:   []string{"true"},
	}

	config := &config.K8sPluginConfig{
		DefaultAffinity: defaultAffinity,
		DefaultPodSecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &runAsUser,
		},
		DefaultPodDNSConfig: &corev1.PodDNSConfig{
			Nameservers: []string{"8.8.8.8", "8.8.4.4"},
			Options: []corev1.PodDNSConfigOption{
				{
					Name:  "ndots",
					Value: &dnsOptVal1,
				},
				{
					Name: "single-request-reopen",
				},
				{
					Name:  "timeout",
					Value: &dnsOptVal2,
				},
				{
					Name:  "attempts",
					Value: &dnsOptVal3,
				},
			},
			Searches: []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"},
		},
		DefaultTolerations: []corev1.Toleration{
			{
				Key:      "x/flyte",
				Value:    "default",
				Operator: "Equal",
				Effect:   "NoSchedule",
			},
		},
		DefaultNodeSelector:       defaultNodeSelector,
		InterruptibleNodeSelector: interruptibleNodeSelector,
		InterruptibleTolerations: []corev1.Toleration{
			{
				Key:      "x/flyte",
				Value:    "interruptible",
				Operator: "Equal",
				Effect:   "NoSchedule",
			},
		},
		InterruptibleNodeSelectorRequirement:    interruptibleNodeSelectorRequirement,
		NonInterruptibleNodeSelectorRequirement: nonInterruptibleNodeSelectorRequirement,
		SchedulerName:                           schedulerName,
		EnableHostNetworkingPod:                 &defaultPodHostNetwork,
		DefaultEnvVars:                          defaultEnvVars,
		DefaultEnvVarsFromEnv:                   defaultEnvVarsFromEnv,
	}
	return config
}

func findEnvVarByName(envVars []corev1.EnvVar, name string) *corev1.EnvVar {
	for _, envVar := range envVars {
		if envVar.Name == name {
			return &envVar
		}
	}
	return nil
}

func TestBuildResourceContainer(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}

	// Case1: Valid Spark Task-Template
	taskTemplate := dummySparkTaskTemplateContainer("blah-1", dummySparkConf)

	// Set spark custom feature config.
	assert.NoError(t, setSparkConfig(&Config{
		Features: []Feature{
			{
				Name:        "feature1",
				SparkConfig: map[string]string{"spark.hadoop.feature1": "true"},
			},
			{
				Name:        "feature2",
				SparkConfig: map[string]string{"spark.hadoop.feature2": "true"},
			},
		},
	}))

	defaultConfig := defaultPluginConfig()
	assert.NoError(t, config.SetK8sPluginConfig(defaultConfig))
	resource, err := sparkResourceHandler.BuildResource(context.TODO(), dummySparkTaskContext(taskTemplate, true))
	assert.Nil(t, err)

	assert.NotNil(t, resource)
	sparkApp, ok := resource.(*sj.SparkApplication)
	assert.True(t, ok)
	assert.Equal(t, sparkMainClass, *sparkApp.Spec.MainClass)
	assert.Equal(t, sparkApplicationFile, *sparkApp.Spec.MainApplicationFile)
	assert.Equal(t, sj.PythonApplicationType, sparkApp.Spec.Type)
	assert.Equal(t, testArgs, sparkApp.Spec.Arguments)
	assert.Equal(t, testImage, *sparkApp.Spec.Image)
	assert.NotNil(t, sparkApp.Spec.Driver.SparkPodSpec.SecurityContenxt)
	assert.Equal(t, *sparkApp.Spec.Driver.SparkPodSpec.SecurityContenxt.RunAsUser, *defaultConfig.DefaultPodSecurityContext.RunAsUser)
	assert.NotNil(t, sparkApp.Spec.Driver.DNSConfig)
	assert.Equal(t, []string{"8.8.8.8", "8.8.4.4"}, sparkApp.Spec.Driver.DNSConfig.Nameservers)
	assert.ElementsMatch(t, defaultConfig.DefaultPodDNSConfig.Options, sparkApp.Spec.Driver.DNSConfig.Options)
	assert.Equal(t, []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"}, sparkApp.Spec.Driver.DNSConfig.Searches)
	assert.NotNil(t, sparkApp.Spec.Executor.SparkPodSpec.SecurityContenxt)
	assert.Equal(t, *sparkApp.Spec.Executor.SparkPodSpec.SecurityContenxt.RunAsUser, *defaultConfig.DefaultPodSecurityContext.RunAsUser)
	assert.NotNil(t, sparkApp.Spec.Executor.DNSConfig)
	assert.NotNil(t, sparkApp.Spec.Executor.DNSConfig)
	assert.ElementsMatch(t, defaultConfig.DefaultPodDNSConfig.Options, sparkApp.Spec.Executor.DNSConfig.Options)
	assert.Equal(t, []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"}, sparkApp.Spec.Executor.DNSConfig.Searches)

	//Validate Driver/Executor Spec.
	driverCores, _ := strconv.ParseInt(dummySparkConf["spark.driver.cores"], 10, 32)
	execCores, _ := strconv.ParseInt(dummySparkConf["spark.executor.cores"], 10, 32)
	execInstances, _ := strconv.ParseInt(dummySparkConf["spark.executor.instances"], 10, 32)

	assert.Equal(t, "new-val", *sparkApp.Spec.ServiceAccount)
	assert.Equal(t, int32(driverCores), *sparkApp.Spec.Driver.Cores)
	assert.Equal(t, int32(execCores), *sparkApp.Spec.Executor.Cores)
	assert.Equal(t, int32(execInstances), *sparkApp.Spec.Executor.Instances)
	assert.Equal(t, dummySparkConf["spark.driver.memory"], *sparkApp.Spec.Driver.Memory)
	assert.Equal(t, dummySparkConf["spark.executor.memory"], *sparkApp.Spec.Executor.Memory)
	assert.Equal(t, dummySparkConf["spark.batchScheduler"], *sparkApp.Spec.BatchScheduler)
	assert.Equal(t, defaultConfig.SchedulerName, *sparkApp.Spec.Executor.SchedulerName)
	assert.Equal(t, defaultConfig.SchedulerName, *sparkApp.Spec.Driver.SchedulerName)
	assert.Equal(t, *defaultConfig.EnableHostNetworkingPod, *sparkApp.Spec.Executor.HostNetwork)
	assert.Equal(t, *defaultConfig.EnableHostNetworkingPod, *sparkApp.Spec.Driver.HostNetwork)

	// Validate
	// * Default tolerations set for both Driver and Executor.
	// * Interruptible tolerations and node selector set for Executor but not Driver.
	// * Default node selector set for both Driver and Executor.
	// * Interruptible node selector requirements set for Executor Affinity, non-interruptiblefir Driver Affinity.
	assert.Equal(t, 1, len(sparkApp.Spec.Driver.Tolerations))
	assert.Equal(t, 1, len(sparkApp.Spec.Driver.NodeSelector))
	assert.Equal(t, defaultConfig.DefaultNodeSelector, sparkApp.Spec.Driver.NodeSelector)
	tolDriverDefault := sparkApp.Spec.Driver.Tolerations[0]
	assert.Equal(t, tolDriverDefault.Key, "x/flyte")
	assert.Equal(t, tolDriverDefault.Value, "default")
	assert.Equal(t, tolDriverDefault.Operator, corev1.TolerationOperator("Equal"))
	assert.Equal(t, tolDriverDefault.Effect, corev1.TaintEffect("NoSchedule"))

	assert.Equal(t, 2, len(sparkApp.Spec.Executor.Tolerations))
	assert.Equal(t, 2, len(sparkApp.Spec.Executor.NodeSelector))
	assert.Equal(t, map[string]string{
		"x/default":       "true",
		"x/interruptible": "true",
	}, sparkApp.Spec.Executor.NodeSelector)

	tolExecInterrupt := sparkApp.Spec.Executor.Tolerations[0]
	assert.Equal(t, tolExecInterrupt.Key, "x/flyte")
	assert.Equal(t, tolExecInterrupt.Value, "interruptible")
	assert.Equal(t, tolExecInterrupt.Operator, corev1.TolerationOperator("Equal"))
	assert.Equal(t, tolExecInterrupt.Effect, corev1.TaintEffect("NoSchedule"))

	tolExecDefault := sparkApp.Spec.Executor.Tolerations[1]
	assert.Equal(t, tolExecDefault.Key, "x/flyte")
	assert.Equal(t, tolExecDefault.Value, "default")
	assert.Equal(t, tolExecDefault.Operator, corev1.TolerationOperator("Equal"))
	assert.Equal(t, tolExecDefault.Effect, corev1.TaintEffect("NoSchedule"))

	for confKey, confVal := range dummySparkConf {
		exists := false

		if featureRegex.MatchString(confKey) && confKey != "spark.flyteorg.feature3.enabled" {
			match := featureRegex.FindAllStringSubmatch(confKey, -1)
			feature := match[0][len(match[0])-1]
			assert.True(t, feature == "feature1" || feature == "feature2")
			for k, v := range sparkApp.Spec.SparkConf {
				key := "spark.hadoop." + feature
				if k == key {
					assert.Equal(t, v, "true")
					exists = true
				}
			}
		} else {
			for k, v := range sparkApp.Spec.SparkConf {

				if k == confKey {
					assert.Equal(t, v, confVal)
					exists = true
				}
			}
		}
		assert.True(t, exists)
	}

	assert.Equal(t, dummySparkConf["spark.driver.cores"], sparkApp.Spec.SparkConf["spark.kubernetes.driver.limit.cores"])
	assert.Equal(t, dummySparkConf["spark.executor.cores"], sparkApp.Spec.SparkConf["spark.kubernetes.executor.limit.cores"])
	assert.Greater(t, len(sparkApp.Spec.SparkConf["spark.kubernetes.driverEnv.FLYTE_START_TIME"]), 1)
	assert.Equal(t, dummySparkConf["spark.flyteorg.feature3.enabled"], sparkApp.Spec.SparkConf["spark.flyteorg.feature3.enabled"])

	assert.Equal(t, len(findEnvVarByName(sparkApp.Spec.Driver.Env, "FLYTE_MAX_ATTEMPTS").Value), 1)
	assert.Equal(t, defaultConfig.DefaultEnvVars["foo"], findEnvVarByName(sparkApp.Spec.Driver.Env, "foo").Value)
	assert.Equal(t, defaultConfig.DefaultEnvVars["foo"], findEnvVarByName(sparkApp.Spec.Executor.Env, "foo").Value)
	assert.Equal(t, defaultConfig.DefaultEnvVars["fooEnv"], findEnvVarByName(sparkApp.Spec.Driver.Env, "fooEnv").Value)
	assert.Equal(t, defaultConfig.DefaultEnvVars["fooEnv"], findEnvVarByName(sparkApp.Spec.Executor.Env, "fooEnv").Value)

	assert.Equal(t, &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						defaultConfig.DefaultAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0],
						*defaultConfig.NonInterruptibleNodeSelectorRequirement,
					},
				},
			},
		},
	}, sparkApp.Spec.Driver.Affinity.NodeAffinity)

	assert.Equal(t, &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						defaultConfig.DefaultAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0],
						*defaultConfig.InterruptibleNodeSelectorRequirement,
					},
				},
			},
		},
	}, sparkApp.Spec.Executor.Affinity.NodeAffinity)

	// Case 2: Driver/Executor request cores set.
	dummyConfWithRequest := make(map[string]string)

	for k, v := range dummySparkConf {
		dummyConfWithRequest[k] = v
	}

	dummyConfWithRequest["spark.kubernetes.driver.request.cores"] = "3"
	dummyConfWithRequest["spark.kubernetes.executor.request.cores"] = "4"

	taskTemplate = dummySparkTaskTemplateContainer("blah-1", dummyConfWithRequest)
	resource, err = sparkResourceHandler.BuildResource(context.TODO(), dummySparkTaskContext(taskTemplate, false))
	assert.Nil(t, err)
	assert.NotNil(t, resource)
	sparkApp, ok = resource.(*sj.SparkApplication)
	assert.True(t, ok)

	assert.Equal(t, dummyConfWithRequest["spark.kubernetes.driver.request.cores"], sparkApp.Spec.SparkConf["spark.kubernetes.driver.limit.cores"])
	assert.Equal(t, dummyConfWithRequest["spark.kubernetes.executor.request.cores"], sparkApp.Spec.SparkConf["spark.kubernetes.executor.limit.cores"])

	// Case 3: Interruptible False
	resource, err = sparkResourceHandler.BuildResource(context.TODO(), dummySparkTaskContext(taskTemplate, false))
	assert.Nil(t, err)
	assert.NotNil(t, resource)
	sparkApp, ok = resource.(*sj.SparkApplication)
	assert.True(t, ok)

	// Validate Interruptible Toleration and NodeSelector not set  for both Driver and Executors.
	// Validate that the default Toleration and NodeSelector are set for both Driver and Executors.
	assert.Equal(t, 1, len(sparkApp.Spec.Driver.Tolerations))
	assert.Equal(t, 1, len(sparkApp.Spec.Driver.NodeSelector))
	assert.Equal(t, defaultConfig.DefaultNodeSelector, sparkApp.Spec.Driver.NodeSelector)
	assert.Equal(t, 1, len(sparkApp.Spec.Executor.Tolerations))
	assert.Equal(t, 1, len(sparkApp.Spec.Executor.NodeSelector))
	assert.Equal(t, defaultConfig.DefaultNodeSelector, sparkApp.Spec.Executor.NodeSelector)
	assert.Equal(t, sparkApp.Spec.Executor.Tolerations[0].Key, "x/flyte")
	assert.Equal(t, sparkApp.Spec.Executor.Tolerations[0].Value, "default")
	assert.Equal(t, sparkApp.Spec.Driver.Tolerations[0].Key, "x/flyte")
	assert.Equal(t, sparkApp.Spec.Driver.Tolerations[0].Value, "default")

	// Validate correct affinity and nodeselector requirements are set for both Driver and Executors.
	assert.Equal(t, &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						defaultConfig.DefaultAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0],
						*defaultConfig.NonInterruptibleNodeSelectorRequirement,
					},
				},
			},
		},
	}, sparkApp.Spec.Driver.Affinity.NodeAffinity)

	assert.Equal(t, &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						defaultConfig.DefaultAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0],
						*defaultConfig.NonInterruptibleNodeSelectorRequirement,
					},
				},
			},
		},
	}, sparkApp.Spec.Executor.Affinity.NodeAffinity)

	// Case 4: Invalid Spark Task-Template
	taskTemplate.Custom = nil
	resource, err = sparkResourceHandler.BuildResource(context.TODO(), dummySparkTaskContext(taskTemplate, false))
	assert.NotNil(t, err)
	assert.Nil(t, resource)
}

func TestBuildResourcePodTemplate(t *testing.T) {
	defaultConfig := defaultPluginConfig()
	assert.NoError(t, config.SetK8sPluginConfig(defaultConfig))
	extraToleration := corev1.Toleration{
		Key:      "x/flyte",
		Value:    "extra",
		Operator: "Equal",
	}
	podSpec := dummyPodSpec()
	podSpec.Tolerations = append(podSpec.Tolerations, extraToleration)
	podSpec.NodeSelector = map[string]string{"x/custom": "foo"}
	taskTemplate := dummySparkTaskTemplatePod("blah-1", dummySparkConf, podSpec)
	taskTemplate.GetK8SPod()
	sparkResourceHandler := sparkResourceHandler{}

	taskCtx := dummySparkTaskContext(taskTemplate, true)
	resource, err := sparkResourceHandler.BuildResource(context.TODO(), taskCtx)

	assert.Nil(t, err)
	assert.NotNil(t, resource)
	sparkApp, ok := resource.(*sj.SparkApplication)
	assert.True(t, ok)

	// Application
	assert.Equal(t, v1.TypeMeta{
		Kind:       KindSparkApplication,
		APIVersion: sparkOp.SchemeGroupVersion.String(),
	}, sparkApp.TypeMeta)

	// Application spec
	assert.Equal(t, flytek8s.GetServiceAccountNameFromTaskExecutionMetadata(taskCtx.TaskExecutionMetadata()), *sparkApp.Spec.ServiceAccount)
	assert.Equal(t, sparkOp.PythonApplicationType, sparkApp.Spec.Type)
	assert.Equal(t, testImage, *sparkApp.Spec.Image)
	assert.Equal(t, testArgs, sparkApp.Spec.Arguments)
	assert.Equal(t, sparkOp.RestartPolicy{
		Type:                       sparkOp.OnFailure,
		OnSubmissionFailureRetries: intPtr(int32(14)),
	}, sparkApp.Spec.RestartPolicy)
	assert.Equal(t, sparkMainClass, *sparkApp.Spec.MainClass)
	assert.Equal(t, sparkApplicationFile, *sparkApp.Spec.MainApplicationFile)

	// Driver
	assert.Equal(t, utils.UnionMaps(defaultConfig.DefaultAnnotations, map[string]string{"annotation-1": "val1"}), sparkApp.Spec.Driver.Annotations)
	assert.Equal(t, utils.UnionMaps(defaultConfig.DefaultLabels, map[string]string{"label-1": "val1"}), sparkApp.Spec.Driver.Labels)
	assert.Equal(t, len(findEnvVarByName(sparkApp.Spec.Driver.Env, "FLYTE_MAX_ATTEMPTS").Value), 1)
	assert.Equal(t, defaultConfig.DefaultEnvVars["foo"], findEnvVarByName(sparkApp.Spec.Driver.Env, "foo").Value)
	assert.Equal(t, defaultConfig.DefaultEnvVars["fooEnv"], findEnvVarByName(sparkApp.Spec.Driver.Env, "fooEnv").Value)
	assert.Equal(t, findEnvVarByName(dummyEnvVarsWithSecretRef, "SECRET"), findEnvVarByName(sparkApp.Spec.Driver.Env, "SECRET"))
	assert.Equal(t, 10, len(sparkApp.Spec.Driver.Env))
	assert.Equal(t, testImage, *sparkApp.Spec.Driver.Image)
	assert.Equal(t, flytek8s.GetServiceAccountNameFromTaskExecutionMetadata(taskCtx.TaskExecutionMetadata()), *sparkApp.Spec.Driver.ServiceAccount)
	assert.Equal(t, defaultConfig.DefaultPodSecurityContext, sparkApp.Spec.Driver.SecurityContenxt)
	assert.Equal(t, defaultConfig.DefaultPodDNSConfig, sparkApp.Spec.Driver.DNSConfig)
	assert.Equal(t, defaultConfig.EnableHostNetworkingPod, sparkApp.Spec.Driver.HostNetwork)
	assert.Equal(t, defaultConfig.SchedulerName, *sparkApp.Spec.Driver.SchedulerName)
	assert.Equal(t, []corev1.Toleration{
		defaultConfig.DefaultTolerations[0],
		extraToleration,
	}, sparkApp.Spec.Driver.Tolerations)
	assert.Equal(t, map[string]string{
		"x/default": "true",
		"x/custom":  "foo",
	}, sparkApp.Spec.Driver.NodeSelector)
	assert.Equal(t, &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						defaultConfig.DefaultAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0],
						*defaultConfig.NonInterruptibleNodeSelectorRequirement,
					},
				},
			},
		},
	}, sparkApp.Spec.Driver.Affinity.NodeAffinity)
	cores, _ := strconv.ParseInt(dummySparkConf["spark.driver.cores"], 10, 32)
	assert.Equal(t, intPtr(int32(cores)), sparkApp.Spec.Driver.Cores)
	assert.Equal(t, dummySparkConf["spark.driver.memory"], *sparkApp.Spec.Driver.Memory)

	// Executor
	assert.Equal(t, utils.UnionMaps(defaultConfig.DefaultAnnotations, map[string]string{"annotation-1": "val1"}), sparkApp.Spec.Executor.Annotations)
	assert.Equal(t, utils.UnionMaps(defaultConfig.DefaultLabels, map[string]string{"label-1": "val1"}), sparkApp.Spec.Executor.Labels)
	assert.Equal(t, defaultConfig.DefaultEnvVars["foo"], findEnvVarByName(sparkApp.Spec.Executor.Env, "foo").Value)
	assert.Equal(t, defaultConfig.DefaultEnvVars["fooEnv"], findEnvVarByName(sparkApp.Spec.Executor.Env, "fooEnv").Value)
	assert.Equal(t, findEnvVarByName(dummyEnvVarsWithSecretRef, "SECRET"), findEnvVarByName(sparkApp.Spec.Executor.Env, "SECRET"))
	assert.Equal(t, 10, len(sparkApp.Spec.Executor.Env))
	assert.Equal(t, testImage, *sparkApp.Spec.Executor.Image)
	assert.Equal(t, defaultConfig.DefaultPodSecurityContext, sparkApp.Spec.Executor.SecurityContenxt)
	assert.Equal(t, defaultConfig.DefaultPodDNSConfig, sparkApp.Spec.Executor.DNSConfig)
	assert.Equal(t, defaultConfig.EnableHostNetworkingPod, sparkApp.Spec.Executor.HostNetwork)
	assert.Equal(t, defaultConfig.SchedulerName, *sparkApp.Spec.Executor.SchedulerName)
	assert.ElementsMatch(t, []corev1.Toleration{
		defaultConfig.DefaultTolerations[0],
		extraToleration,
		defaultConfig.InterruptibleTolerations[0],
	}, sparkApp.Spec.Executor.Tolerations)
	assert.Equal(t, map[string]string{
		"x/default":       "true",
		"x/custom":        "foo",
		"x/interruptible": "true",
	}, sparkApp.Spec.Executor.NodeSelector)
	assert.Equal(t, &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						defaultConfig.DefaultAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0],
						*defaultConfig.InterruptibleNodeSelectorRequirement,
					},
				},
			},
		},
	}, sparkApp.Spec.Executor.Affinity.NodeAffinity)
	cores, _ = strconv.ParseInt(dummySparkConf["spark.executor.cores"], 10, 32)
	instances, _ := strconv.ParseInt(dummySparkConf["spark.executor.instances"], 10, 32)
	assert.Equal(t, intPtr(int32(instances)), sparkApp.Spec.Executor.Instances)
	assert.Equal(t, intPtr(int32(cores)), sparkApp.Spec.Executor.Cores)
	assert.Equal(t, dummySparkConf["spark.executor.memory"], *sparkApp.Spec.Executor.Memory)
}

func TestGetPropertiesSpark(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, sparkResourceHandler.GetProperties())
}

func TestGetTaskPhaseWithNamespaceInLogContext(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}
	ctx := context.TODO()

	pluginCtx := dummySparkPluginContext(dummySparkTaskTemplateContainer("", dummySparkConf), k8s.PluginState{})
	taskPhase, err := sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.RunningState))
	assert.NoError(t, err)
	assert.NotNil(t, taskPhase.Info())
	assert.NotNil(t, taskPhase.Info().LogContext)
	assert.Equal(t, 2, len(taskPhase.Info().LogContext.Pods))

	// Verify namespace is set in the driver pod log context
	driverPodLogContext := taskPhase.Info().LogContext.Pods[0]
	assert.Equal(t, "spark-namespace", driverPodLogContext.Namespace)
	assert.Equal(t, "spark-pod", driverPodLogContext.PodName)
	assert.Equal(t, defaultDriverPrimaryContainerName, driverPodLogContext.PrimaryContainerName)
}

func TestGetTaskPhaseWithFailedPod(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}
	ctx := context.TODO()

	// Create a failed driver pod
	pod := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "spark-pod",
			Namespace: "spark-namespace",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: defaultDriverPrimaryContainerName,
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 1,
							Reason:   "Error",
							Message:  "Container failed",
						},
					},
				},
			},
		},
	}

	pluginCtx := dummySparkPluginContextWithPods(dummySparkTaskTemplateContainer("", dummySparkConf), k8s.PluginState{}, pod)

	// Even though SparkApplication status is running, should return failure due to pod status
	taskPhase, err := sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.RunningState))
	assert.NoError(t, err)
	assert.True(t, taskPhase.Phase().IsFailure())
}

func TestGetTaskPhaseWithPendingPodInvalidImage(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}
	ctx := context.TODO()

	// Create a pending driver pod with InvalidImageName - this should fail immediately
	pod := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "spark-pod",
			Namespace: "spark-namespace",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			Conditions: []corev1.PodCondition{
				{
					Type:               corev1.PodReady,
					Status:             corev1.ConditionFalse,
					LastTransitionTime: v1.Time{Time: time.Now()},
					Reason:             "ContainersNotReady",
					Message:            "containers with unready status: [spark-kubernetes-driver]",
				},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  defaultDriverPrimaryContainerName,
					Ready: false,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "InvalidImageName",
							Message: "Invalid image name",
						},
					},
				},
			},
		},
	}

	pluginCtx := dummySparkPluginContextWithPods(dummySparkTaskTemplateContainer("", dummySparkConf), k8s.PluginState{}, pod)

	taskPhase, err := sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.SubmittedState))
	assert.NoError(t, err)
	// Should detect the InvalidImageName and return a failure phase
	assert.True(t, taskPhase.Phase().IsFailure())
}

func TestGetTaskPhaseContainerNameConstant(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}
	ctx := context.TODO()

	pluginCtx := dummySparkPluginContext(dummySparkTaskTemplateContainer("", dummySparkConf), k8s.PluginState{})

	taskPhase, err := sparkResourceHandler.GetTaskPhase(ctx, pluginCtx, dummySparkApplication(sj.CompletedState))
	assert.NoError(t, err)
	assert.NotNil(t, taskPhase.Info())
	assert.NotNil(t, taskPhase.Info().LogContext)

	// Verify the constant is used for driver container names
	driverPodLogContext := taskPhase.Info().LogContext.Pods[0]
	assert.Equal(t, defaultDriverPrimaryContainerName, driverPodLogContext.PrimaryContainerName)
	assert.Equal(t, 1, len(driverPodLogContext.Containers))
	assert.Equal(t, defaultDriverPrimaryContainerName, driverPodLogContext.Containers[0].ContainerName)
}

func TestIsTerminal(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}
	ctx := context.Background()

	tests := []struct {
		name           string
		state          sj.ApplicationStateType
		expectedResult bool
	}{
		{"Completed", sj.CompletedState, true},
		{"Failed", sj.FailedState, true},
		{"FailedSubmission", sj.FailedSubmissionState, true},
		{"New", sj.NewState, false},
		{"Submitted", sj.SubmittedState, false},
		{"Running", sj.RunningState, false},
		{"PendingRerun", sj.PendingRerunState, false},
		{"Invalidating", sj.InvalidatingState, false},
		{"Succeeding", sj.SucceedingState, false},
		{"Failing", sj.FailingState, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := dummySparkApplication(tt.state)
			result, err := sparkResourceHandler.IsTerminal(ctx, app)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestIsTerminal_WrongResourceType(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}
	ctx := context.Background()

	var wrongResource client.Object = &corev1.ConfigMap{}
	result, err := sparkResourceHandler.IsTerminal(ctx, wrongResource)
	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "unexpected resource type")
}

func TestGetCompletionTime(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}

	now := time.Now().Truncate(time.Second)
	earlier := now.Add(-1 * time.Hour)
	evenEarlier := now.Add(-2 * time.Hour)

	tests := []struct {
		name         string
		app          *sj.SparkApplication
		expectedTime time.Time
	}{
		{
			name: "uses TerminationTime",
			app: &sj.SparkApplication{
				ObjectMeta: v1.ObjectMeta{
					CreationTimestamp: v1.NewTime(evenEarlier),
				},
				Status: sj.SparkApplicationStatus{
					TerminationTime: v1.NewTime(now),
					SubmissionTime:  v1.NewTime(earlier),
				},
			},
			expectedTime: now,
		},
		{
			name: "falls back to SubmissionTime",
			app: &sj.SparkApplication{
				ObjectMeta: v1.ObjectMeta{
					CreationTimestamp: v1.NewTime(evenEarlier),
				},
				Status: sj.SparkApplicationStatus{
					SubmissionTime: v1.NewTime(now),
				},
			},
			expectedTime: now,
		},
		{
			name: "falls back to CreationTimestamp",
			app: &sj.SparkApplication{
				ObjectMeta: v1.ObjectMeta{
					CreationTimestamp: v1.NewTime(now),
				},
				Status: sj.SparkApplicationStatus{},
			},
			expectedTime: now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sparkResourceHandler.GetCompletionTime(tt.app)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTime.Unix(), result.Unix())
		})
	}
}

func TestGetCompletionTime_WrongResourceType(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}

	var wrongResource client.Object = &corev1.ConfigMap{}
	result, err := sparkResourceHandler.GetCompletionTime(wrongResource)
	assert.Error(t, err)
	assert.True(t, result.IsZero())
	assert.Contains(t, err.Error(), "unexpected resource type")
}
