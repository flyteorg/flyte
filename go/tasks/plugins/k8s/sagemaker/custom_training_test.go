package sagemaker

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/go-test/deep"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	commonv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/common"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"

	sagemakerIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins/sagemaker"
	"github.com/stretchr/testify/assert"

	trainingjobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/trainingjob"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	stdConfig "github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
)

func Test_awsSagemakerPlugin_BuildResourceForCustomTrainingJob(t *testing.T) {
	// Default config does not contain a roleAnnotationKey -> expecting to get the role from default config
	ctx := context.TODO()
	defaultCfg := config.GetSagemakerConfig()
	defer func() {
		_ = config.SetSagemakerConfig(defaultCfg)
	}()
	t.Run("In a custom training job we should see the FLYTE_SAGEMAKER_CMD being injected", func(t *testing.T) {
		// Injecting a config which contains a mismatched roleAnnotationKey -> expecting to get the role from the config
		configAccessor := viper.NewAccessor(stdConfig.Options{
			StrictMode: true,
			// Use a different
			SearchPaths: []string{"testdata/config2.yaml"},
		})

		err := configAccessor.UpdateConfig(context.TODO())
		assert.NoError(t, err)

		awsSageMakerTrainingJobHandler := awsSagemakerPlugin{TaskType: customTrainingJobTaskType}

		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_CUSTOM, "0.90", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job", tjObj)

		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, generateMockCustomTrainingJobTaskContext(taskTemplate, false))
		assert.NoError(t, err)
		assert.NotNil(t, trainingJobResource)

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)
		assert.Equal(t, "config_role", *trainingJob.Spec.RoleArn)
		//assert.Equal(t, 1, len(trainingJob.Spec.HyperParameters))
		fmt.Printf("%v", trainingJob.Spec.HyperParameters)
		expectedHPs := []*commonv1.KeyValuePair{
			{Name: fmt.Sprintf("%s%d_%s%s", FlyteSageMakerCmdKeyPrefix, 0, "service_venv", FlyteSageMakerKeySuffix), Value: FlyteSageMakerCmdDummyValue},
			{Name: fmt.Sprintf("%s%d_%s%s", FlyteSageMakerCmdKeyPrefix, 1, "pyflyte-execute", FlyteSageMakerKeySuffix), Value: FlyteSageMakerCmdDummyValue},
			{Name: fmt.Sprintf("%s%d_%s%s", FlyteSageMakerCmdKeyPrefix, 2, "--test-opt1", FlyteSageMakerKeySuffix), Value: FlyteSageMakerCmdDummyValue},
			{Name: fmt.Sprintf("%s%d_%s%s", FlyteSageMakerCmdKeyPrefix, 3, "value1", FlyteSageMakerKeySuffix), Value: FlyteSageMakerCmdDummyValue},
			{Name: fmt.Sprintf("%s%d_%s%s", FlyteSageMakerCmdKeyPrefix, 4, "--test-opt2", FlyteSageMakerKeySuffix), Value: FlyteSageMakerCmdDummyValue},
			{Name: fmt.Sprintf("%s%d_%s%s", FlyteSageMakerCmdKeyPrefix, 5, "value2", FlyteSageMakerKeySuffix), Value: FlyteSageMakerCmdDummyValue},
			{Name: fmt.Sprintf("%s%d_%s%s", FlyteSageMakerCmdKeyPrefix, 6, "--test-flag", FlyteSageMakerKeySuffix), Value: FlyteSageMakerCmdDummyValue},
			{Name: fmt.Sprintf("%s%s%s", FlyteSageMakerEnvVarKeyPrefix, "Env_Var", FlyteSageMakerKeySuffix), Value: "Env_Val"},
			{Name: fmt.Sprintf("%s%s%s", FlyteSageMakerEnvVarKeyPrefix, FlyteSageMakerEnvVarKeyStatsdDisabled, FlyteSageMakerKeySuffix), Value: strconv.FormatBool(true)},
			{Name: SageMakerMpiEnableEnvVarName, Value: strconv.FormatBool(false)},
		}
		assert.Equal(t, len(expectedHPs), len(trainingJob.Spec.HyperParameters))
		for i := range expectedHPs {
			assert.Equal(t, expectedHPs[i].Name, trainingJob.Spec.HyperParameters[i].Name)
			assert.Equal(t, expectedHPs[i].Value, trainingJob.Spec.HyperParameters[i].Value)
		}

		assert.Equal(t, testImage, *trainingJob.Spec.AlgorithmSpecification.TrainingImage)

		// Since the distributed protocol is UNSPECIFIED, we should find sagemaker_mpi_enabled=false in the hyperparameters
		count := 0
		for i := range trainingJob.Spec.HyperParameters {
			if trainingJob.Spec.HyperParameters[i].Name == SageMakerMpiEnableEnvVarName && trainingJob.Spec.HyperParameters[i].Value == strconv.FormatBool(false) {
				count++
			}
		}
		assert.Equal(t, 1, count)
	})

	t.Run("In a custom training job when users specify the MPI distributed protocol, even when the instance count is 1, we should find sagemaker_mpi_enabled=true in the hyperparameters", func(t *testing.T) {
		// Injecting a config which contains a mismatched roleAnnotationKey -> expecting to get the role from the config
		configAccessor := viper.NewAccessor(stdConfig.Options{
			StrictMode: true,
			// Use a different
			SearchPaths: []string{"testdata/config2.yaml"},
		})

		err := configAccessor.UpdateConfig(context.TODO())
		assert.NoError(t, err)

		awsSageMakerTrainingJobHandler := awsSagemakerPlugin{TaskType: customTrainingJobTaskType}

		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_CUSTOM, "0.90", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_MPI)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job", tjObj)

		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, generateMockCustomTrainingJobTaskContext(taskTemplate, false))
		assert.NoError(t, err)
		assert.NotNil(t, trainingJobResource)

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)

		count := 0
		for i := range trainingJob.Spec.HyperParameters {
			if trainingJob.Spec.HyperParameters[i].Name == SageMakerMpiEnableEnvVarName {
				count++
				assert.Equal(t, trainingJob.Spec.HyperParameters[i].Value, strconv.FormatBool(true))
			}
		}
		assert.Equal(t, 1, count)
	})

	t.Run("When users specify the MPI distributed protocol, we should find sagemaker_mpi_enabled=true in the hyperparameters", func(t *testing.T) {
		// Injecting a config which contains a mismatched roleAnnotationKey -> expecting to get the role from the config
		configAccessor := viper.NewAccessor(stdConfig.Options{
			StrictMode: true,
			// Use a different
			SearchPaths: []string{"testdata/config2.yaml"},
		})

		err := configAccessor.UpdateConfig(context.TODO())
		assert.NoError(t, err)

		awsSageMakerTrainingJobHandler := awsSagemakerPlugin{TaskType: customTrainingJobTaskType}

		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_CUSTOM, "0.90", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 2, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_MPI)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job", tjObj)

		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, generateMockCustomTrainingJobTaskContext(taskTemplate, false))
		assert.NoError(t, err)
		assert.NotNil(t, trainingJobResource)

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)

		count := 0
		for i := range trainingJob.Spec.HyperParameters {
			if trainingJob.Spec.HyperParameters[i].Name == SageMakerMpiEnableEnvVarName {
				count++
				assert.Equal(t, trainingJob.Spec.HyperParameters[i].Value, strconv.FormatBool(true))
			}
		}
		assert.Equal(t, 1, count)
	})
}

func Test_awsSagemakerPlugin_GetTaskPhaseForCustomTrainingJob(t *testing.T) {
	ctx := context.TODO()
	// Injecting a config which contains a mismatched roleAnnotationKey -> expecting to get the role from the config
	configAccessor := viper.NewAccessor(stdConfig.Options{
		StrictMode: true,
		// Use a different
		SearchPaths: []string{"testdata/config2.yaml"},
	})

	err := configAccessor.UpdateConfig(context.TODO())
	assert.NoError(t, err)

	awsSageMakerTrainingJobHandler := awsSagemakerPlugin{TaskType: customTrainingJobTaskType}

	t.Run("TrainingJobStatusCompleted", func(t *testing.T) {
		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_XGBOOST, "0.90", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job", tjObj)
		taskCtx := generateMockCustomTrainingJobTaskContext(taskTemplate, false)
		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, taskCtx)
		assert.Error(t, err)
		assert.Nil(t, trainingJobResource)
	})

	t.Run("TrainingJobStatusCompleted", func(t *testing.T) {
		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_CUSTOM, "", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job", tjObj)
		taskCtx := generateMockCustomTrainingJobTaskContext(taskTemplate, false)
		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, taskCtx)
		assert.NoError(t, err)
		assert.NotNil(t, trainingJobResource)

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)

		trainingJob.Status.TrainingJobStatus = sagemaker.TrainingJobStatusCompleted
		phaseInfo, err := awsSageMakerTrainingJobHandler.getTaskPhaseForCustomTrainingJob(ctx, taskCtx, trainingJob)
		assert.Nil(t, err)
		assert.Equal(t, phaseInfo.Phase(), pluginsCore.PhaseSuccess)
	})
	t.Run("OutputWriter.Put returns an error", func(t *testing.T) {
		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_CUSTOM, "", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job", tjObj)
		taskCtx := generateMockCustomTrainingJobTaskContext(taskTemplate, true)
		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, taskCtx)
		assert.NoError(t, err)
		assert.NotNil(t, trainingJobResource)

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)

		trainingJob.Status.TrainingJobStatus = sagemaker.TrainingJobStatusCompleted
		phaseInfo, err := awsSageMakerTrainingJobHandler.getTaskPhaseForCustomTrainingJob(ctx, taskCtx, trainingJob)
		assert.NotNil(t, err)
		assert.Equal(t, phaseInfo.Phase(), pluginsCore.PhaseUndefined)
	})
}

func Test_awsSagemakerPlugin_getEventInfoForCustomTrainingJob(t *testing.T) {
	// Default config does not contain a roleAnnotationKey -> expecting to get the role from default config
	ctx := context.TODO()
	defaultCfg := config.GetSagemakerConfig()
	defer func() {
		_ = config.SetSagemakerConfig(defaultCfg)
	}()

	t.Run("get event info should return correctly formatted log links for custom training job", func(t *testing.T) {
		// Injecting a config which contains a mismatched roleAnnotationKey -> expecting to get the role from the config
		configAccessor := viper.NewAccessor(stdConfig.Options{
			StrictMode: true,
			// Use a different
			SearchPaths: []string{"testdata/config2.yaml"},
		})

		err := configAccessor.UpdateConfig(context.TODO())
		assert.NoError(t, err)

		awsSageMakerTrainingJobHandler := awsSagemakerPlugin{TaskType: customTrainingJobTaskType}

		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_CUSTOM, "", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job", tjObj)
		taskCtx := generateMockCustomTrainingJobTaskContext(taskTemplate, false)
		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, taskCtx)
		assert.NoError(t, err)
		assert.NotNil(t, trainingJobResource)

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)

		taskInfo, err := awsSageMakerTrainingJobHandler.getEventInfoForCustomTrainingJob(ctx, trainingJob)
		if err != nil {
			panic(err)
		}

		expectedTaskLogs := []*flyteIdlCore.TaskLog{
			{
				Uri: fmt.Sprintf("https://%s.console.aws.amazon.com/cloudwatch/home?region=%s#logStream:group=/aws/sagemaker/TrainingJobs;prefix=%s;streamFilter=typeLogStreamPrefix",
					"us-west-2", "us-west-2", taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()),
				Name:          CloudWatchLogLinkName,
				MessageFormat: flyteIdlCore.TaskLog_JSON,
			},
			{
				Uri: fmt.Sprintf("https://%s.console.aws.amazon.com/sagemaker/home?region=%s#/%s/%s",
					"us-west-2", "us-west-2", "jobs", taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()),
				Name:          CustomTrainingJobSageMakerLinkName,
				MessageFormat: flyteIdlCore.TaskLog_UNKNOWN,
			},
		}

		expectedCustomInfo, _ := utils.MarshalObjToStruct(map[string]string{})
		assert.Equal(t,
			func(tis []*flyteIdlCore.TaskLog) []flyteIdlCore.TaskLog {
				ret := make([]flyteIdlCore.TaskLog, 0, len(tis))
				for _, ti := range tis {
					ret = append(ret, *ti)
				}
				return ret
			}(expectedTaskLogs),
			func(tis []*flyteIdlCore.TaskLog) []flyteIdlCore.TaskLog {
				ret := make([]flyteIdlCore.TaskLog, 0, len(tis))
				for _, ti := range tis {
					ret = append(ret, *ti)
				}
				return ret
			}(taskInfo.Logs))
		if diff := deep.Equal(expectedCustomInfo, taskInfo.CustomInfo); diff != nil {
			assert.FailNow(t, "Should be equal.", "Diff: %v", diff)
		}
		assert.Len(t, taskInfo.ExternalResources, 1)
		assert.Equal(t, taskInfo.ExternalResources[0].ExternalID, "some-acceptable-name")
	})
}
