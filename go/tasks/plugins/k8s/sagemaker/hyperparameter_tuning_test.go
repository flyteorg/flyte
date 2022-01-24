package sagemaker

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-test/deep"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	stdConfig "github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"

	hpojobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/hyperparametertuningjob"
	sagemakerIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins/sagemaker"
	"github.com/stretchr/testify/assert"
)

func Test_awsSagemakerPlugin_BuildResourceForHyperparameterTuningJob(t *testing.T) {
	// Default config does not contain a roleAnnotationKey -> expecting to get the role from default config
	ctx := context.TODO()
	err := config.ResetSagemakerConfig()
	if err != nil {
		panic(err)
	}
	defaultCfg := config.GetSagemakerConfig()
	awsSageMakerHPOJobHandler := awsSagemakerPlugin{TaskType: hyperparameterTuningJobTaskType}

	tjObj := generateMockTrainingJobCustomObj(
		sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_XGBOOST, "0.90", []*sagemakerIdl.MetricDefinition{},
		sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
	htObj := generateMockHyperparameterTuningJobCustomObj(tjObj, 10, 5)
	taskTemplate := generateMockHyperparameterTuningJobTaskTemplate("the job", htObj)
	hpoJobResource, err := awsSageMakerHPOJobHandler.BuildResource(ctx, generateMockHyperparameterTuningJobTaskContext(taskTemplate))
	assert.NoError(t, err)
	assert.NotNil(t, hpoJobResource)

	hpoJob, ok := hpoJobResource.(*hpojobv1.HyperparameterTuningJob)
	assert.True(t, ok)
	assert.NotNil(t, hpoJob.Spec.TrainingJobDefinition)
	assert.Equal(t, 1, len(hpoJob.Spec.HyperParameterTuningJobConfig.ParameterRanges.IntegerParameterRanges))
	assert.Equal(t, 0, len(hpoJob.Spec.HyperParameterTuningJobConfig.ParameterRanges.ContinuousParameterRanges))
	assert.Equal(t, 0, len(hpoJob.Spec.HyperParameterTuningJobConfig.ParameterRanges.CategoricalParameterRanges))
	assert.Equal(t, "us-east-1", *hpoJob.Spec.Region)
	assert.Equal(t, "default_role", *hpoJob.Spec.TrainingJobDefinition.RoleArn)

	err = config.SetSagemakerConfig(defaultCfg)
	if err != nil {
		panic(err)
	}
}

func Test_awsSagemakerPlugin_getEventInfoForHyperparameterTuningJob(t *testing.T) {
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

		awsSageMakerHPOJobHandler := awsSagemakerPlugin{TaskType: hyperparameterTuningJobTaskType}

		tjObj := generateMockTrainingJobCustomObj(
			sagemakerIdl.InputMode_FILE, sagemakerIdl.AlgorithmName_XGBOOST, "0.90", []*sagemakerIdl.MetricDefinition{},
			sagemakerIdl.InputContentType_TEXT_CSV, 1, "ml.m4.xlarge", 25, sagemakerIdl.DistributedProtocol_UNSPECIFIED)
		htObj := generateMockHyperparameterTuningJobCustomObj(tjObj, 10, 5)
		taskTemplate := generateMockHyperparameterTuningJobTaskTemplate("the job", htObj)
		taskCtx := generateMockHyperparameterTuningJobTaskContext(taskTemplate)
		hpoJobResource, err := awsSageMakerHPOJobHandler.BuildResource(ctx, taskCtx)
		assert.NoError(t, err)
		assert.NotNil(t, hpoJobResource)

		hpoJob, ok := hpoJobResource.(*hpojobv1.HyperparameterTuningJob)
		assert.True(t, ok)

		taskInfo, err := awsSageMakerHPOJobHandler.getEventInfoForHyperparameterTuningJob(ctx, hpoJob)
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
					"us-west-2", "us-west-2", "hyper-tuning-jobs", taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()),
				Name:          HyperparameterTuningJobSageMakerLinkName,
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
