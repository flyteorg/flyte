package sagemaker

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	taskError "github.com/flyteorg/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flytestdlib/logger"

	pluginErrors "github.com/flyteorg/flyteplugins/go/tasks/errors"

	commonv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/common"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"

	trainingjobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/trainingjob"
	flyteSageMakerIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins/sagemaker"
)

func (m awsSagemakerPlugin) buildResourceForCustomTrainingJob(
	ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {

	logger.Infof(ctx, "Building a training job resource for task [%v]", taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())
	taskTemplate, err := getTaskTemplate(ctx, taskCtx)
	if err != nil {
		return nil, err
	}

	// Unmarshal the custom field of the task template back into the Hyperparameter Tuning Job struct generated in flyteidl
	sagemakerTrainingJob := flyteSageMakerIdl.TrainingJob{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &sagemakerTrainingJob)
	if err != nil {

		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "invalid TrainingJob task specification: not able to unmarshal the custom field to [%s]", m.TaskType)
	}

	if sagemakerTrainingJob.GetAlgorithmSpecification() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "The unmarshaled training job does not have a AlgorithmSpecification field")
	}
	if sagemakerTrainingJob.GetAlgorithmSpecification().GetAlgorithmName() != flyteSageMakerIdl.AlgorithmName_CUSTOM {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "The algorithm name [%v] is not supported by the custom training job plugin",
			sagemakerTrainingJob.GetAlgorithmSpecification().GetAlgorithmName().String())
	}

	inputChannels := make([]commonv1.Channel, 0)
	inputModeString := strings.Title(strings.ToLower(sagemakerTrainingJob.GetAlgorithmSpecification().GetInputMode().String()))

	jobName := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	outputPath := taskCtx.OutputWriter().GetOutputPrefixPath().String()

	if taskTemplate.GetContainer() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "The task template points to a nil container")
	}

	if taskTemplate.GetContainer().GetImage() == "" {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "Invalid image of the container")
	}

	trainingImageStr := taskTemplate.GetContainer().GetImage()

	cfg := config.GetSagemakerConfig()

	var metricDefinitions []commonv1.MetricDefinition
	idlMetricDefinitions := sagemakerTrainingJob.GetAlgorithmSpecification().GetMetricDefinitions()
	for _, md := range idlMetricDefinitions {
		metricDefinitions = append(metricDefinitions,
			commonv1.MetricDefinition{Name: ToStringPtr(md.Name), Regex: ToStringPtr(md.Regex)})
	}

	// TODO: When dealing with HPO job, we will need to deal with the following
	// jobOutputPath := NewJobOutputPaths(ctx, taskCtx.DataStore(), taskCtx.OutputWriter().GetOutputPrefixPath(), jobName)

	hyperParameters, err := injectArgsAndEnvVars(ctx, taskCtx, taskTemplate)
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "Failed to inject the task template's container env vars to the hyperparameter list")
	}

	// Statsd is not available in SM because we are not allowed kick off a side car container nor are we allowed to customize our own AMI in an SM environment
	// Therefore we need to inject a env var to disable statsd for sagemaker tasks
	statsdDisableEnvVarName := fmt.Sprintf("%s%s%s", FlyteSageMakerEnvVarKeyPrefix, FlyteSageMakerEnvVarKeyStatsdDisabled, FlyteSageMakerKeySuffix)
	logger.Infof(ctx, "Injecting %v=%v to force disable statsd for SageMaker tasks only", statsdDisableEnvVarName, strconv.FormatBool(true))
	hyperParameters = append(hyperParameters, &commonv1.KeyValuePair{
		Name:  statsdDisableEnvVarName,
		Value: strconv.FormatBool(true),
	})

	if sagemakerTrainingJob.GetTrainingJobResourceConfig() == nil {
		logger.Errorf(ctx, "TrainingJobResourceConfig is nil")
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "TrainingJobResourceConfig is nil")
	}

	if sagemakerTrainingJob.GetTrainingJobResourceConfig().GetDistributedProtocol() == flyteSageMakerIdl.DistributedProtocol_MPI {
		// inject sagemaker_mpi_enabled=true into hyperparameters if the user code designates MPI as its distributed training framework
		logger.Infof(ctx, "MPI is enabled by the user. TrainingJob.TrainingJobResourceConfig.DistributedProtocol=[%v]", sagemakerTrainingJob.GetTrainingJobResourceConfig().GetDistributedProtocol().String())
		hyperParameters = append(hyperParameters, &commonv1.KeyValuePair{
			Name:  SageMakerMpiEnableEnvVarName,
			Value: strconv.FormatBool(true),
		})
	} else {
		// default value: injecting sagemaker_mpi_enabled=false
		logger.Infof(ctx, "Distributed protocol is unspecified or a non-MPI value [%v] in the training job", sagemakerTrainingJob.GetTrainingJobResourceConfig().GetDistributedProtocol())
		hyperParameters = append(hyperParameters, &commonv1.KeyValuePair{
			Name:  SageMakerMpiEnableEnvVarName,
			Value: strconv.FormatBool(false),
		})
	}
	logger.Infof(ctx, "The Sagemaker TrainingJob Task plugin received static hyperparameters [%v]", hyperParameters)

	trainingJob := &trainingjobv1.TrainingJob{
		Spec: trainingjobv1.TrainingJobSpec{
			AlgorithmSpecification: &commonv1.AlgorithmSpecification{
				// If the specify a value for this AlgorithmName parameter, the user can't specify a value for TrainingImage.
				// in this Flyte plugin, we always use the algorithm name and version the user provides via Flytekit to map to an image
				// so we intentionally leave this field nil
				AlgorithmName:     nil,
				TrainingImage:     ToStringPtr(trainingImageStr),
				TrainingInputMode: commonv1.TrainingInputMode(inputModeString),
				MetricDefinitions: metricDefinitions,
			},
			// The support of spot training will come in a later version
			EnableManagedSpotTraining: nil,
			HyperParameters:           hyperParameters,
			InputDataConfig:           inputChannels,
			OutputDataConfig: &commonv1.OutputDataConfig{
				S3OutputPath: ToStringPtr(outputPath),
			},
			CheckpointConfig: nil,
			ResourceConfig: &commonv1.ResourceConfig{
				InstanceType:   sagemakerTrainingJob.GetTrainingJobResourceConfig().GetInstanceType(),
				InstanceCount:  ToInt64Ptr(sagemakerTrainingJob.GetTrainingJobResourceConfig().GetInstanceCount()),
				VolumeSizeInGB: ToInt64Ptr(sagemakerTrainingJob.GetTrainingJobResourceConfig().GetVolumeSizeInGb()),
				VolumeKmsKeyId: ToStringPtr(""), // TODO: Not yet supported. Need to add to proto and flytekit in the future
			},
			RoleArn: ToStringPtr(cfg.RoleArn),
			Region:  ToStringPtr(cfg.Region),
			StoppingCondition: &commonv1.StoppingCondition{
				MaxRuntimeInSeconds:  ToInt64Ptr(86400), // TODO: decide how to coordinate this and Flyte's timeout
				MaxWaitTimeInSeconds: nil,               // TODO: decide how to coordinate this and Flyte's timeout and queueing budget
			},
			TensorBoardOutputConfig: nil,
			Tags:                    nil,
			TrainingJobName:         &jobName,
		},
	}
	logger.Infof(ctx, "Successfully built a custom training job resource for task [%v]", taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())
	return trainingJob, nil
}

func (m awsSagemakerPlugin) getTaskPhaseForCustomTrainingJob(
	ctx context.Context, pluginContext k8s.PluginContext, trainingJob *trainingjobv1.TrainingJob) (pluginsCore.PhaseInfo, error) {

	logger.Infof(ctx, "Getting task phase for sagemaker training job [%v]", trainingJob.Status.SageMakerTrainingJobName)
	info, err := m.getEventInfoForCustomTrainingJob(ctx, trainingJob)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, pluginErrors.Wrapf(pluginErrors.RuntimeFailure, err, "Failed to get event info for the job")
	}

	occurredAt := time.Now()

	switch trainingJob.Status.TrainingJobStatus {
	case ReconcilingTrainingJobStatus:
		logger.Errorf(ctx, "Job stuck in reconciling status, assuming retryable failure [%s]", trainingJob.Status.Additional)
		// TODO talk to AWS about why there cannot be an explicit condition that signals AWS API call pluginErrors
		execError := &flyteIdlCore.ExecutionError{
			Message: trainingJob.Status.Additional,
			Kind:    flyteIdlCore.ExecutionError_USER,
			Code:    ReconcilingTrainingJobStatus,
		}
		return pluginsCore.PhaseInfoFailed(pluginsCore.PhaseRetryableFailure, execError, info), nil
	case sagemaker.TrainingJobStatusFailed:
		execError := &flyteIdlCore.ExecutionError{
			Message: trainingJob.Status.Additional,
			Kind:    flyteIdlCore.ExecutionError_USER,
			Code:    sagemaker.TrainingJobStatusFailed,
		}
		return pluginsCore.PhaseInfoFailed(pluginsCore.PhasePermanentFailure, execError, info), nil
	case sagemaker.TrainingJobStatusStopped:
		return pluginsCore.PhaseInfoRetryableFailure(taskError.DownstreamSystemError, "Training Job Stopped", info), nil
	case sagemaker.TrainingJobStatusCompleted:
		// Now that it is a success we will set the outputs as expected by the task

		logger.Infof(ctx, "Looking for the output.pb under %s", pluginContext.OutputWriter().GetOutputPrefixPath())
		outputReader := ioutils.NewRemoteFileOutputReader(ctx, pluginContext.DataStore(), pluginContext.OutputWriter(), pluginContext.MaxDatasetSizeBytes())

		// Instantiate a output reader with the literal map, and write the output to the remote location referred to by the OutputWriter
		if err := pluginContext.OutputWriter().Put(ctx, outputReader); err != nil {
			return pluginsCore.PhaseInfoUndefined, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "Failed to write output to the remote location")
		}
		logger.Debugf(ctx, "Successfully produced and returned outputs")
		return pluginsCore.PhaseInfoSuccess(info), nil
	case "":
		return pluginsCore.PhaseInfoQueued(occurredAt, pluginsCore.DefaultPhaseVersion, "job submitted"), nil
	}

	return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, info), nil
}

func (m awsSagemakerPlugin) getEventInfoForCustomTrainingJob(ctx context.Context, trainingJob *trainingjobv1.TrainingJob) (*pluginsCore.TaskInfo, error) {

	var jobRegion, jobName, jobTypeInURL, sagemakerLinkName string
	jobRegion = *trainingJob.Spec.Region
	jobName = *trainingJob.Spec.TrainingJobName
	jobTypeInURL = "jobs"
	sagemakerLinkName = CustomTrainingJobSageMakerLinkName

	logger.Infof(ctx, "Getting event information for SageMaker CustomTrainingJob task, job region: [%v], job name: [%v], "+
		"job type in url: [%v], sagemaker link name: [%v]", jobRegion, jobName, jobTypeInURL, sagemakerLinkName)

	return createTaskInfo(ctx, jobRegion, jobName, jobTypeInURL, sagemakerLinkName)
}
