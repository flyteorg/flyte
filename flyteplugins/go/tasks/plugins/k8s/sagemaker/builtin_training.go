package sagemaker

import (
	"context"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	trainingjobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/trainingjob"

	awsUtils "github.com/flyteorg/flyteplugins/go/tasks/plugins/awsutils"

	pluginErrors "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	commonv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/common"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	taskError "github.com/flyteorg/flyteplugins/go/tasks/errors"

	flyteSageMakerIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins/sagemaker"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"
)

const ReconcilingTrainingJobStatus = "ReconcilingTrainingJob"

func (m awsSagemakerPlugin) buildResourceForTrainingJob(
	ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {

	logger.Infof(ctx, "Building a training job resource for task [%v]", taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())
	taskTemplate, err := getTaskTemplate(ctx, taskCtx)
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "Failed to get the task template of the training job task")
	}

	// Unmarshal the custom field of the task template back into the Hyperparameter Tuning Job struct generated in flyteidl
	sagemakerTrainingJob := flyteSageMakerIdl.TrainingJob{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &sagemakerTrainingJob)
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "invalid TrainingJob task specification: not able to unmarshal the custom field to [%s]", m.TaskType)
	}
	if sagemakerTrainingJob.GetTrainingJobResourceConfig() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "Required field [TrainingJobResourceConfig] of the TrainingJob does not exist")
	}
	if sagemakerTrainingJob.GetAlgorithmSpecification() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "Required field [AlgorithmSpecification] does not exist")
	}
	if sagemakerTrainingJob.GetAlgorithmSpecification().GetAlgorithmName() == flyteSageMakerIdl.AlgorithmName_CUSTOM {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "Custom algorithm is not supported by the built-in training job plugin")
	}

	taskInput, err := taskCtx.InputReader().Get(ctx)
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "unable to fetch task inputs")
	}

	// Get inputs from literals
	inputLiterals := taskInput.GetLiterals()
	err = checkIfRequiredInputLiteralsExist(inputLiterals,
		[]string{TrainPredefinedInputVariable, ValidationPredefinedInputVariable, StaticHyperparametersPredefinedInputVariable})
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "Error occurred when checking if all the required inputs exist")
	}

	trainPathLiteral := inputLiterals[TrainPredefinedInputVariable]
	validationPathLiteral := inputLiterals[ValidationPredefinedInputVariable]
	staticHyperparamsLiteral := inputLiterals[StaticHyperparametersPredefinedInputVariable]

	if trainPathLiteral.GetScalar() == nil || trainPathLiteral.GetScalar().GetBlob() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "[train] Input is required and should be of Type [Scalar.Blob]")
	}
	if validationPathLiteral.GetScalar() == nil || validationPathLiteral.GetScalar().GetBlob() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "[validation] Input is required and should be of Type [Scalar.Blob]")
	}

	// Convert the hyperparameters to the spec value
	staticHyperparams, err := convertStaticHyperparamsLiteralToSpecType(staticHyperparamsLiteral)
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "could not convert static hyperparameters to spec type")
	}

	outputPath := createOutputPath(taskCtx.OutputWriter().GetRawOutputPrefix().String(), TrainingJobOutputPathSubDir)

	jobName := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()

	trainingImageStr, err := getTrainingJobImage(ctx, taskCtx, &sagemakerTrainingJob)
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "failed to find the training image")
	}

	logger.Infof(ctx, "The Sagemaker TrainingJob Task plugin received static hyperparameters [%v]", staticHyperparams)

	cfg := config.GetSagemakerConfig()

	var metricDefinitions []commonv1.MetricDefinition
	idlMetricDefinitions := sagemakerTrainingJob.GetAlgorithmSpecification().GetMetricDefinitions()
	for _, md := range idlMetricDefinitions {
		metricDefinitions = append(metricDefinitions,
			commonv1.MetricDefinition{Name: ToStringPtr(md.Name), Regex: ToStringPtr(md.Regex)})
	}

	apiContentType, err := getAPIContentType(sagemakerTrainingJob.GetAlgorithmSpecification().GetInputContentType())
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "Unsupported input file type [%v]", sagemakerTrainingJob.GetAlgorithmSpecification().GetInputContentType().String())
	}

	inputModeString := strings.Title(strings.ToLower(sagemakerTrainingJob.GetAlgorithmSpecification().GetInputMode().String()))

	role := awsUtils.GetRoleFromSecurityContext(cfg.RoleAnnotationKey, taskCtx.TaskExecutionMetadata())

	if len(role) == 0 {
		role = cfg.RoleArn
	}

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
			HyperParameters:           staticHyperparams,
			InputDataConfig: []commonv1.Channel{
				{
					ChannelName: ToStringPtr(TrainPredefinedInputVariable),
					DataSource: &commonv1.DataSource{
						S3DataSource: &commonv1.S3DataSource{
							S3DataType: "S3Prefix",
							S3Uri:      ToStringPtr(trainPathLiteral.GetScalar().GetBlob().GetUri()),
						},
					},
					ContentType: ToStringPtr(apiContentType),
					InputMode:   inputModeString,
				},
				{
					ChannelName: ToStringPtr(ValidationPredefinedInputVariable),
					DataSource: &commonv1.DataSource{
						S3DataSource: &commonv1.S3DataSource{
							S3DataType: "S3Prefix",
							S3Uri:      ToStringPtr(validationPathLiteral.GetScalar().GetBlob().GetUri()),
						},
					},
					ContentType: ToStringPtr(apiContentType),
					InputMode:   inputModeString,
				},
			},
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
			RoleArn: ToStringPtr(role),
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
	logger.Infof(ctx, "Successfully built a training job resource for task [%v]", taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())
	return trainingJob, nil
}

func (m awsSagemakerPlugin) getTaskPhaseForTrainingJob(
	ctx context.Context, pluginContext k8s.PluginContext, trainingJob *trainingjobv1.TrainingJob) (pluginsCore.PhaseInfo, error) {

	logger.Infof(ctx, "Getting task phase for sagemaker training job [%v]", trainingJob.Status.SageMakerTrainingJobName)
	info, err := m.getEventInfoForTrainingJob(ctx, trainingJob)
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

		// We have specified an output path in the CRD, and we know SageMaker will automatically upload the
		// model tarball to s3://<specified-output-path>/<training-job-name>/output/model.tar.gz

		// Therefore, here we create a output literal map, where we fill in the above path to the URI field of the
		// blob output, which will later be written out by the OutputWriter to the outputs.pb remotely on S3
		outputLiteralMap, err := getOutputLiteralMapFromTaskInterface(ctx, pluginContext.TaskReader(),
			createModelOutputPath(trainingJob, pluginContext.OutputWriter().GetRawOutputPrefix().String(), trainingJob.Status.SageMakerTrainingJobName))
		if err != nil {
			logger.Errorf(ctx, "Failed to create outputs, err: %s", err)
			return pluginsCore.PhaseInfoUndefined, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "failed to create outputs for the task")
		}
		// Instantiate a output reader with the literal map, and write the output to the remote location referred to by the OutputWriter
		if err := pluginContext.OutputWriter().Put(ctx, ioutils.NewInMemoryOutputReader(outputLiteralMap, nil, nil)); err != nil {
			return pluginsCore.PhaseInfoUndefined, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "Unable to write output to the remote location")
		}
		logger.Debugf(ctx, "Successfully produced and returned outputs")
		return pluginsCore.PhaseInfoSuccess(info), nil
	case "":
		return pluginsCore.PhaseInfoQueued(occurredAt, pluginsCore.DefaultPhaseVersion, "job submitted"), nil
	}

	return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, info), nil
}

func (m awsSagemakerPlugin) getEventInfoForTrainingJob(ctx context.Context, trainingJob *trainingjobv1.TrainingJob) (*pluginsCore.TaskInfo, error) {

	var jobRegion, jobName, jobTypeInURL, sagemakerLinkName string
	jobRegion = *trainingJob.Spec.Region
	jobName = *trainingJob.Spec.TrainingJobName
	jobTypeInURL = "jobs"
	sagemakerLinkName = TrainingJobSageMakerLinkName

	logger.Infof(ctx, "Getting event information for SageMaker BuiltinAlgorithmTrainingJob task, job region: [%v], job name: [%v], "+
		"job type in url: [%v], sagemaker link name: [%v]", jobRegion, jobName, jobTypeInURL, sagemakerLinkName)

	return createTaskInfo(ctx, jobRegion, jobName, jobTypeInURL, sagemakerLinkName)
}
