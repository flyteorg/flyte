package spark

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"

	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteplugins/go/tasks/logs"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"

	pluginIOMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"

	sj "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const sparkMainClass = "MainClass"
const sparkApplicationFile = "local:///spark_app.py"
const testImage = "image://"
const sparkUIAddress = "spark-ui.flyte"

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
	info, err := getEventInfoForSpark(dummySparkApplication(sj.RunningState))
	assert.NoError(t, err)
	assert.Len(t, info.Logs, 6)
	assert.Equal(t, fmt.Sprintf("https://%s", sparkUIAddress), info.CustomInfo.Fields[sparkDriverUI].GetStringValue())
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

	info, err = getEventInfoForSpark(dummySparkApplication(sj.SubmittedState))
	assert.NoError(t, err)
	assert.Len(t, info.Logs, 1)
	assert.Equal(t, "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.spark-app-name;streamFilter=typeLogStreamPrefix", info.Logs[0].Uri)

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

	info, err = getEventInfoForSpark(dummySparkApplication(sj.FailedState))
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

	ctx := context.TODO()
	taskPhase, err := sparkResourceHandler.GetTaskPhase(ctx, nil, dummySparkApplication(sj.NewState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseQueued)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, nil, dummySparkApplication(sj.SubmittedState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseInitializing)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, nil, dummySparkApplication(sj.RunningState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, nil, dummySparkApplication(sj.CompletedState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseSuccess)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, nil, dummySparkApplication(sj.InvalidatingState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, nil, dummySparkApplication(sj.FailingState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, nil, dummySparkApplication(sj.PendingRerunState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, nil, dummySparkApplication(sj.SucceedingState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, nil, dummySparkApplication(sj.FailedSubmissionState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRetryableFailure)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = sparkResourceHandler.GetTaskPhase(ctx, nil, dummySparkApplication(sj.FailedState))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRetryableFailure)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)
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

func dummySparkTaskTemplate(id string, sparkConf map[string]string) *core.TaskTemplate {

	sparkJob := dummySparkCustomObj(sparkConf)
	sparkJobJSON, err := utils.MarshalToString(sparkJob)
	if err != nil {
		panic(err)
	}

	structObj := structpb.Struct{}

	err = jsonpb.UnmarshalString(sparkJobJSON, &structObj)
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

func dummySparkTaskContext(taskTemplate *core.TaskTemplate, interruptible bool) pluginsCore.TaskExecutionContext {
	taskCtx := &mocks.TaskExecutionContext{}
	inputReader := &pluginIOMocks.InputReader{}
	inputReader.OnGetInputPrefixPath().Return(storage.DataReference("/input/prefix"))
	inputReader.OnGetInputPath().Return(storage.DataReference("/input"))
	inputReader.OnGetMatch(mock.Anything).Return(&core.LiteralMap{}, nil)
	taskCtx.OnInputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.On("GetOutputPath").Return(storage.DataReference("/data/outputs.pb"))
	outputReader.On("GetOutputPrefixPath").Return(storage.DataReference("/data/"))
	outputReader.On("GetRawOutputPrefix").Return(storage.DataReference(""))

	taskCtx.On("OutputWriter").Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.On("Read", mock.Anything).Return(taskTemplate, nil)
	taskCtx.On("TaskReader").Return(taskReader)

	tID := &mocks.TaskExecutionID{}
	tID.On("GetID").Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.On("GetGeneratedName").Return("some-acceptable-name")

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.On("GetTaskExecutionID").Return(tID)
	taskExecutionMetadata.On("GetNamespace").Return("test-namespace")
	taskExecutionMetadata.On("GetAnnotations").Return(map[string]string{"annotation-1": "val1"})
	taskExecutionMetadata.On("GetLabels").Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.On("GetOwnerReference").Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.On("GetSecurityContext").Return(core.SecurityContext{
		RunAs: &core.Identity{K8SServiceAccount: "new-val"},
	})
	taskExecutionMetadata.On("IsInterruptible").Return(interruptible)
	taskExecutionMetadata.On("GetMaxAttempts").Return(uint32(1))
	taskCtx.On("TaskExecutionMetadata").Return(taskExecutionMetadata)
	return taskCtx
}

func TestBuildResourceSpark(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}

	// Case1: Valid Spark Task-Template
	taskTemplate := dummySparkTaskTemplate("blah-1", dummySparkConf)

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

	// Set Interruptible Config
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		InterruptibleNodeSelector: map[string]string{
			"x/interruptible": "true",
		},
		InterruptibleTolerations: []corev1.Toleration{
			{
				Key:      "x/flyte",
				Value:    "interruptible",
				Operator: "Equal",
				Effect:   "NoSchedule",
			},
		}}),
	)
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

	// Validate Interruptible Toleration and NodeSelector set for Executor but not Driver.
	assert.Equal(t, 0, len(sparkApp.Spec.Driver.Tolerations))
	assert.Equal(t, 0, len(sparkApp.Spec.Driver.NodeSelector))

	assert.Equal(t, 1, len(sparkApp.Spec.Executor.Tolerations))
	assert.Equal(t, 1, len(sparkApp.Spec.Executor.NodeSelector))

	tol := sparkApp.Spec.Executor.Tolerations[0]
	assert.Equal(t, tol.Key, "x/flyte")
	assert.Equal(t, tol.Value, "interruptible")
	assert.Equal(t, tol.Operator, corev1.TolerationOperator("Equal"))
	assert.Equal(t, tol.Effect, corev1.TaintEffect("NoSchedule"))
	assert.Equal(t, "true", sparkApp.Spec.Executor.NodeSelector["x/interruptible"])

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

	assert.Equal(t, len(sparkApp.Spec.Driver.EnvVars["FLYTE_MAX_ATTEMPTS"]), 1)

	// Case 2: Driver/Executor request cores set.
	dummyConfWithRequest := make(map[string]string)

	for k, v := range dummySparkConf {
		dummyConfWithRequest[k] = v
	}

	dummyConfWithRequest["spark.kubernetes.driver.request.cores"] = "3"
	dummyConfWithRequest["spark.kubernetes.executor.request.cores"] = "4"

	taskTemplate = dummySparkTaskTemplate("blah-1", dummyConfWithRequest)
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
	assert.Equal(t, 0, len(sparkApp.Spec.Driver.Tolerations))
	assert.Equal(t, 0, len(sparkApp.Spec.Driver.NodeSelector))
	assert.Equal(t, 0, len(sparkApp.Spec.Executor.Tolerations))
	assert.Equal(t, 0, len(sparkApp.Spec.Executor.NodeSelector))

	// Case 4: Invalid Spark Task-Template
	taskTemplate.Custom = nil
	resource, err = sparkResourceHandler.BuildResource(context.TODO(), dummySparkTaskContext(taskTemplate, false))
	assert.NotNil(t, err)
	assert.Nil(t, resource)
}

func TestGetPropertiesSpark(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, sparkResourceHandler.GetProperties())
}
