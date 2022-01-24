package sagemaker

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-test/deep"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	taskError "github.com/flyteorg/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	commonv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/common"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"

	stdConfig "github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"

	trainingjobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/trainingjob"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	sagemakerIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins/sagemaker"
	"github.com/stretchr/testify/assert"
)

func Test_awsSagemakerPlugin_BuildResourceForTrainingJob(t *testing.T) {
	// Default config does not contain a roleAnnotationKey -> expecting to get the role from default config
	ctx := context.TODO()
	defaultCfg := config.GetSagemakerConfig()
	defer func() {
		_ = config.SetSagemakerConfig(defaultCfg)
	}()

	t.Run("If roleAnnotationKey has a match, the role from the metadata should be fetched", func(t *testing.T) {
		// Injecting a config which contains a matching roleAnnotationKey -> expecting to get the role from metadata
		configAccessor := viper.NewAccessor(stdConfig.Options{
			StrictMode:  true,
			SearchPaths: []string{"testdata/config.yaml"},
		})

		err := configAccessor.UpdateConfig(context.TODO())
		assert.NoError(t, err)

		awsSageMakerTrainingJobHandler := awsSagemakerPlugin{TaskType: trainingJobTaskType}

		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_XGBOOST, "0.90", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job x", tjObj)

		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, generateMockTrainingJobTaskContext(taskTemplate, false))
		assert.NoError(t, err)
		assert.NotNil(t, trainingJobResource)

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)
		assert.Equal(t, "metadata_role", *trainingJob.Spec.RoleArn)
	})

	t.Run("If roleAnnotationKey does not have a match, the role from the config should be fetched", func(t *testing.T) {
		// Injecting a config which contains a mismatched roleAnnotationKey -> expecting to get the role from the config
		configAccessor := viper.NewAccessor(stdConfig.Options{
			StrictMode: true,
			// Use a different
			SearchPaths: []string{"testdata/config2.yaml"},
		})

		err := configAccessor.UpdateConfig(context.TODO())
		assert.NoError(t, err)

		awsSageMakerTrainingJobHandler := awsSagemakerPlugin{TaskType: trainingJobTaskType}

		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_XGBOOST, "0.90", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job y", tjObj)

		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, generateMockTrainingJobTaskContext(taskTemplate, false))
		assert.NoError(t, err)
		assert.NotNil(t, trainingJobResource)

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)
		assert.Equal(t, "config_role", *trainingJob.Spec.RoleArn)
	})

	t.Run("In a custom training job we should see the FLYTE_SAGEMAKER_CMD being injected", func(t *testing.T) {
		// Injecting a config which contains a mismatched roleAnnotationKey -> expecting to get the role from the config
		configAccessor := viper.NewAccessor(stdConfig.Options{
			StrictMode: true,
			// Use a different
			SearchPaths: []string{"testdata/config2.yaml"},
		})

		err := configAccessor.UpdateConfig(context.TODO())
		assert.NoError(t, err)

		awsSageMakerTrainingJobHandler := awsSagemakerPlugin{TaskType: trainingJobTaskType}

		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_XGBOOST, "0.90", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job", tjObj)

		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, generateMockTrainingJobTaskContext(taskTemplate, false))
		assert.NoError(t, err)
		assert.NotNil(t, trainingJobResource)

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)
		assert.Equal(t, "config_role", *trainingJob.Spec.RoleArn)

		expectedHPs := []*commonv1.KeyValuePair{
			{Name: "a", Value: "1"},
			{Name: "b", Value: "2"},
		}

		assert.ElementsMatch(t,
			func(kvs []*commonv1.KeyValuePair) []commonv1.KeyValuePair {
				ret := make([]commonv1.KeyValuePair, 0, len(kvs))
				for _, kv := range kvs {
					ret = append(ret, *kv)
				}
				return ret
			}(expectedHPs),
			func(kvs []*commonv1.KeyValuePair) []commonv1.KeyValuePair {
				ret := make([]commonv1.KeyValuePair, 0, len(kvs))
				for _, kv := range kvs {
					ret = append(ret, *kv)
				}
				return ret
			}(trainingJob.Spec.HyperParameters))
	})
}

func Test_awsSagemakerPlugin_GetTaskPhaseForTrainingJob(t *testing.T) {
	ctx := context.TODO()
	// Injecting a config which contains a mismatched roleAnnotationKey -> expecting to get the role from the config
	configAccessor := viper.NewAccessor(stdConfig.Options{
		StrictMode: true,
		// Use a different
		SearchPaths: []string{"testdata/config2.yaml"},
	})

	err := configAccessor.UpdateConfig(context.TODO())
	assert.NoError(t, err)

	awsSageMakerTrainingJobHandler := awsSagemakerPlugin{TaskType: trainingJobTaskType}

	tjObj := generateMockTrainingJobCustomObj(
		sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_XGBOOST, "0.90", []*sagemakerIdl.MetricDefinition{},
		sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
	taskTemplate := generateMockTrainingJobTaskTemplate("the job", tjObj)
	taskCtx := generateMockTrainingJobTaskContext(taskTemplate, false)
	trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, taskCtx)
	assert.NoError(t, err)
	assert.NotNil(t, trainingJobResource)

	t.Run("ReconcilingTrainingJobStatus should lead to a retryable failure", func(t *testing.T) {

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)
		trainingJob.Status.TrainingJobStatus = ReconcilingTrainingJobStatus
		phaseInfo, err := awsSageMakerTrainingJobHandler.getTaskPhaseForTrainingJob(ctx, taskCtx, trainingJob)
		assert.Nil(t, err)
		assert.Equal(t, phaseInfo.Phase(), pluginsCore.PhaseRetryableFailure)
		assert.Equal(t, phaseInfo.Err().GetKind(), flyteIdlCore.ExecutionError_USER)
		assert.Equal(t, phaseInfo.Err().GetCode(), ReconcilingTrainingJobStatus)
		assert.Equal(t, phaseInfo.Err().GetMessage(), "")
	})
	t.Run("TrainingJobStatusFailed should be a permanent failure", func(t *testing.T) {
		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)
		trainingJob.Status.TrainingJobStatus = sagemaker.TrainingJobStatusFailed

		phaseInfo, err := awsSageMakerTrainingJobHandler.getTaskPhaseForTrainingJob(ctx, taskCtx, trainingJob)
		assert.Nil(t, err)
		assert.Equal(t, phaseInfo.Phase(), pluginsCore.PhasePermanentFailure)
		assert.Equal(t, phaseInfo.Err().GetKind(), flyteIdlCore.ExecutionError_USER)
		assert.Equal(t, phaseInfo.Err().GetCode(), sagemaker.TrainingJobStatusFailed)
		assert.Equal(t, phaseInfo.Err().GetMessage(), "")
	})
	t.Run("TrainingJobStatusFailed should be a permanent failure", func(t *testing.T) {

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)
		trainingJob.Status.TrainingJobStatus = sagemaker.TrainingJobStatusStopped

		phaseInfo, err := awsSageMakerTrainingJobHandler.getTaskPhaseForTrainingJob(ctx, taskCtx, trainingJob)
		assert.Nil(t, err)
		assert.Equal(t, phaseInfo.Phase(), pluginsCore.PhaseRetryableFailure)
		assert.Equal(t, phaseInfo.Err().GetKind(), flyteIdlCore.ExecutionError_USER)
		assert.Equal(t, phaseInfo.Err().GetCode(), taskError.DownstreamSystemError)
		// We have a default message for TrainingJobStatusStopped
		assert.Equal(t, phaseInfo.Err().GetMessage(), "Training Job Stopped")
	})
}

func Test_awsSagemakerPlugin_getEventInfoForTrainingJob(t *testing.T) {
	// Default config does not contain a roleAnnotationKey -> expecting to get the role from default config
	ctx := context.TODO()
	defaultCfg := config.GetSagemakerConfig()
	defer func() {
		_ = config.SetSagemakerConfig(defaultCfg)
	}()

	t.Run("get event info should return correctly formatted log links for training job", func(t *testing.T) {
		// Injecting a config which contains a mismatched roleAnnotationKey -> expecting to get the role from the config
		configAccessor := viper.NewAccessor(stdConfig.Options{
			StrictMode: true,
			// Use a different
			SearchPaths: []string{"testdata/config2.yaml"},
		})

		err := configAccessor.UpdateConfig(context.TODO())
		assert.NoError(t, err)

		awsSageMakerTrainingJobHandler := awsSagemakerPlugin{TaskType: trainingJobTaskType}

		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_XGBOOST, "0.90", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
		taskTemplate := generateMockTrainingJobTaskTemplate("the job", tjObj)
		taskCtx := generateMockTrainingJobTaskContext(taskTemplate, false)
		trainingJobResource, err := awsSageMakerTrainingJobHandler.BuildResource(ctx, taskCtx)
		assert.NoError(t, err)
		assert.NotNil(t, trainingJobResource)

		trainingJob, ok := trainingJobResource.(*trainingjobv1.TrainingJob)
		assert.True(t, ok)

		taskInfo, err := awsSageMakerTrainingJobHandler.getEventInfoForTrainingJob(ctx, trainingJob)
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
				Name:          TrainingJobSageMakerLinkName,
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
