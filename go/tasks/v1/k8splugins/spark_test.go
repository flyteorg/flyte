package k8splugins

import (
	"context"
	"fmt"
	"testing"

	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flyteplugins/go/tasks/v1/logs"
	"github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"

	sj "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta1"
	"github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flyteplugins/go/tasks/v1/types"
)

const sparkApplicationFile = "local:///spark_app.py"
const testImage = "image://"
const sparkUIAddress = "spark-ui.flyte"

var (
	dummySparkConf = map[string]string{
		"spark.driver.memory":      "500M",
		"spark.driver.cores":       "1",
		"spark.executor.cores":     "1",
		"spark.executor.instances": "3",
		"spark.executor.memory":    "500M",
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
	assert.NoError(t, logs.SetLogConfig(&logs.LogConfig{
		IsCloudwatchEnabled: true,
		CloudwatchRegion:    "us-east-1",
		CloudwatchLogGroup:  "/kubernetes/flyte",
		IsKubernetesEnabled: true,
		KubernetesURL:       "k8s.com",
	}))
	info, err := getEventInfoForSpark(dummySparkApplication(sj.RunningState))
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("https://%s", sparkUIAddress), info.CustomInfo.Fields[sparkDriverUI].GetStringValue())
	assert.Equal(t, "k8s.com/#!/log/spark-namespace/spark-pod/pod?namespace=spark-namespace", info.Logs[0].Uri)
	assert.Equal(t, "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.spark-pod;streamFilter=typeLogStreamPrefix", info.Logs[1].Uri)
	assert.Equal(t, "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=system_log.var.log.containers.spark-app-name;streamFilter=typeLogStreamPrefix", info.Logs[2].Uri)
	assert.Equal(t, "https://spark-ui.flyte", info.Logs[3].Uri)

	assert.NoError(t, setSparkConfig(&SparkConfig{
		SparkHistoryServerURL: "spark-history.flyte",
	}))

	info, err = getEventInfoForSpark(dummySparkApplication(sj.FailedState))
	assert.NoError(t, err)
	assert.Equal(t, "spark-history.flyte/history/app-id", info.CustomInfo.Fields[sparkHistoryUI].GetStringValue())
	assert.Equal(t, "k8s.com/#!/log/spark-namespace/spark-pod/pod?namespace=spark-namespace", info.Logs[0].Uri)
	assert.Equal(t, "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.spark-pod;streamFilter=typeLogStreamPrefix", info.Logs[1].Uri)
	assert.Equal(t, "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=system_log.var.log.containers.spark-app-name;streamFilter=typeLogStreamPrefix", info.Logs[2].Uri)
	assert.Equal(t, "spark-history.flyte/history/app-id", info.Logs[3].Uri)
}

func TestGetTaskStatus(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}

	ctx := context.TODO()
	taskStatus, i, err := sparkResourceHandler.GetTaskStatus(ctx, nil, dummySparkApplication(sj.NewState))
	assert.NoError(t, err)
	assert.Equal(t, taskStatus.Phase, types.TaskPhaseQueued)
	assert.NotNil(t, i)
	assert.Nil(t, err)

	taskStatus, i, err = sparkResourceHandler.GetTaskStatus(ctx, nil, dummySparkApplication(sj.SubmittedState))
	assert.NoError(t, err)
	assert.Equal(t, taskStatus.Phase, types.TaskPhaseQueued)
	assert.NotNil(t, i)
	assert.Nil(t, err)

	taskStatus, i, err = sparkResourceHandler.GetTaskStatus(ctx, nil, dummySparkApplication(sj.RunningState))
	assert.NoError(t, err)
	assert.Equal(t, taskStatus.Phase, types.TaskPhaseRunning)
	assert.NotNil(t, i)
	assert.Nil(t, err)

	taskStatus, i, err = sparkResourceHandler.GetTaskStatus(ctx, nil, dummySparkApplication(sj.CompletedState))
	assert.NoError(t, err)
	assert.Equal(t, taskStatus.Phase, types.TaskPhaseSucceeded)
	assert.NotNil(t, i)
	assert.Nil(t, err)

	taskStatus, i, err = sparkResourceHandler.GetTaskStatus(ctx, nil, dummySparkApplication(sj.InvalidatingState))
	assert.NoError(t, err)
	assert.Equal(t, taskStatus.Phase, types.TaskPhaseRunning)
	assert.NotNil(t, i)
	assert.Nil(t, err)

	taskStatus, i, err = sparkResourceHandler.GetTaskStatus(ctx, nil, dummySparkApplication(sj.FailingState))
	assert.NoError(t, err)
	assert.Equal(t, taskStatus.Phase, types.TaskPhaseRunning)
	assert.NotNil(t, i)
	assert.Nil(t, err)

	taskStatus, i, err = sparkResourceHandler.GetTaskStatus(ctx, nil, dummySparkApplication(sj.PendingRerunState))
	assert.NoError(t, err)
	assert.Equal(t, taskStatus.Phase, types.TaskPhaseRunning)
	assert.NotNil(t, i)
	assert.Nil(t, err)

	taskStatus, i, err = sparkResourceHandler.GetTaskStatus(ctx, nil, dummySparkApplication(sj.SucceedingState))
	assert.NoError(t, err)
	assert.Equal(t, taskStatus.Phase, types.TaskPhaseRunning)
	assert.NotNil(t, i)
	assert.Nil(t, err)

	taskStatus, i, err = sparkResourceHandler.GetTaskStatus(ctx, nil, dummySparkApplication(sj.FailedSubmissionState))
	assert.NoError(t, err)
	assert.Equal(t, taskStatus.Phase, types.TaskPhaseRetryableFailure)
	assert.NotNil(t, i)
	assert.Nil(t, err)

	taskStatus, i, err = sparkResourceHandler.GetTaskStatus(ctx, nil, dummySparkApplication(sj.FailedState))
	assert.NoError(t, err)
	assert.Equal(t, taskStatus.Phase, types.TaskPhaseRetryableFailure)
	assert.NotNil(t, i)
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

func dummySparkCustomObj() *plugins.SparkJob {
	sparkJob := plugins.SparkJob{}

	sparkJob.MainApplicationFile = sparkApplicationFile
	sparkJob.SparkConf = dummySparkConf
	sparkJob.ApplicationType = plugins.SparkApplication_PYTHON
	return &sparkJob
}

func dummySparkTaskTemplate(id string) *core.TaskTemplate {

	sparkJob := dummySparkCustomObj()
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

func dummySparkTaskContext() types.TaskContext {
	taskCtx := &mocks.TaskContext{}
	taskCtx.On("GetNamespace").Return("test-namespace")
	taskCtx.On("GetAnnotations").Return(map[string]string{"annotation-1": "val1"})
	taskCtx.On("GetLabels").Return(map[string]string{"label-1": "val1"})
	taskCtx.On("GetOwnerReference").Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskCtx.On("GetDataDir").Return(storage.DataReference("/data/"))
	taskCtx.On("GetInputsFile").Return(storage.DataReference("/input"))

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
	taskCtx.On("GetTaskExecutionID").Return(tID)
	return taskCtx
}

func TestBuildResourceSpark(t *testing.T) {
	sparkResourceHandler := sparkResourceHandler{}

	// Case1: Valid Spark Task-Template
	taskTemplate := dummySparkTaskTemplate("blah-1")

	resource, err := sparkResourceHandler.BuildResource(context.TODO(), dummySparkTaskContext(), taskTemplate, nil)
	assert.Nil(t, err)

	assert.NotNil(t, resource)
	sparkApp, ok := resource.(*sj.SparkApplication)
	assert.True(t, ok)
	assert.Equal(t, sparkApplicationFile, *sparkApp.Spec.MainApplicationFile)
	assert.Equal(t, sj.PythonApplicationType, sparkApp.Spec.Type)
	assert.Equal(t, testArgs, sparkApp.Spec.Arguments)
	assert.Equal(t, testImage, *sparkApp.Spec.Image)

	for confKey, confVal := range dummySparkConf {
		exists := false
		for k, v := range sparkApp.Spec.SparkConf {
			if k == confKey {
				assert.Equal(t, v, confVal)
				exists = true
			}
		}
		assert.True(t, exists)
	}

	assert.Equal(t, dummySparkConf["spark.driver.cores"], sparkApp.Spec.SparkConf["spark.kubernetes.driver.limit.cores"])
	assert.Equal(t, dummySparkConf["spark.executor.cores"], sparkApp.Spec.SparkConf["spark.kubernetes.executor.limit.cores"])

	// Case2: Invalid Spark Task-Template
	taskTemplate.Custom = nil
	resource, err = sparkResourceHandler.BuildResource(context.TODO(), dummySparkTaskContext(), taskTemplate, nil)
	assert.NotNil(t, err)
	assert.Nil(t, resource)
}
