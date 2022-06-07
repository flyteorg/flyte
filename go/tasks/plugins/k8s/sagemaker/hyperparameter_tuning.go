package sagemaker

import (
	"context"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	awsUtils "github.com/flyteorg/flyteplugins/go/tasks/plugins/awsutils"

	pluginErrors "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	commonv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/common"
	hpojobv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/hyperparametertuningjob"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	taskError "github.com/flyteorg/flyteplugins/go/tasks/errors"

	flyteSageMakerIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins/sagemaker"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"
)

const ReconcilingTuningJobStatus = "ReconcilingTuningJob"

func (m awsSagemakerPlugin) buildResourceForHyperparameterTuningJob(
	ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {

	logger.Infof(ctx, "Building a hyperparameter tuning job resource for task [%v]", taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())

	taskTemplate, err := getTaskTemplate(ctx, taskCtx)
	if err != nil {
		return nil, err
	}

	// Unmarshal the custom field of the task template back into the HyperparameterTuningJob struct generated in flyteidl
	sagemakerHPOJob := flyteSageMakerIdl.HyperparameterTuningJob{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &sagemakerHPOJob)
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "invalid HyperparameterTuningJob task specification: not able to unmarshal the custom field to [%s]", hyperparameterTuningJobTaskType)
	}
	if sagemakerHPOJob.GetTrainingJob() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "Required field [TrainingJob] of the HyperparameterTuningJob does not exist")
	}
	if sagemakerHPOJob.GetTrainingJob().GetAlgorithmSpecification() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "Required field [AlgorithmSpecification] of the HyperparameterTuningJob's underlying TrainingJob does not exist")
	}
	if sagemakerHPOJob.GetTrainingJob().GetTrainingJobResourceConfig() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "Required field [TrainingJobResourceConfig] of the HyperparameterTuningJob's underlying TrainingJob does not exist")
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
	validatePathLiteral := inputLiterals[ValidationPredefinedInputVariable]
	staticHyperparamsLiteral := inputLiterals[StaticHyperparametersPredefinedInputVariable]
	hpoJobConfigLiteral := inputLiterals["hyperparameter_tuning_job_config"]
	if trainPathLiteral.GetScalar() == nil || trainPathLiteral.GetScalar().GetBlob() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "[%v] Input is required and should be of Type [Scalar.Blob]", TrainPredefinedInputVariable)
	}
	if validatePathLiteral.GetScalar() == nil || validatePathLiteral.GetScalar().GetBlob() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "[%v] Input is required and should be of Type [Scalar.Blob]", ValidationPredefinedInputVariable)
	}
	// Convert the hyperparameters to the spec value
	staticHyperparams, err := convertStaticHyperparamsLiteralToSpecType(staticHyperparamsLiteral)
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "could not convert static hyperparameters to spec type")
	}

	// hyperparameter_tuning_job_config is marshaled into a struct in flytekit, so will have to unmarshal it back
	hpoJobConfig, err := convertHyperparameterTuningJobConfigToSpecType(hpoJobConfigLiteral)
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "failed to convert hyperparameter tuning job config literal to spec type")
	}

	outputPath := createOutputPath(taskCtx.OutputWriter().GetRawOutputPrefix().String(), HyperparameterOutputPathSubDir)
	if hpoJobConfig.GetTuningObjective() == nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "Required field [TuningObjective] does not exist")
	}

	// Deleting the conflicting static hyperparameters: if a hyperparameter exist in both the map of static hyperparameter
	// and the map of the tunable hyperparameter inside the Hyperparameter Tuning Job Config, we delete the entry
	// in the static map and let the one in the map of the tunable hyperparameters take precedence
	staticHyperparams = deleteConflictingStaticHyperparameters(ctx, staticHyperparams, hpoJobConfig.GetHyperparameterRanges().GetParameterRangeMap())

	jobName := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()

	trainingImageStr, err := getTrainingJobImage(ctx, taskCtx, sagemakerHPOJob.GetTrainingJob())
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "failed to find the training image")
	}

	hpoJobParameterRanges := buildParameterRanges(ctx, inputLiterals)
	logger.Infof(ctx, "The Sagemaker HyperparameterTuningJob Task plugin received the following inputs: \n"+
		"static hyperparameters: [%v]\n"+
		"hyperparameter tuning job config: [%v]\n"+
		"parameter ranges: [%v]", staticHyperparams, hpoJobConfig, hpoJobParameterRanges)

	cfg := config.GetSagemakerConfig()

	var metricDefinitions []commonv1.MetricDefinition
	idlMetricDefinitions := sagemakerHPOJob.GetTrainingJob().GetAlgorithmSpecification().GetMetricDefinitions()
	for _, md := range idlMetricDefinitions {
		metricDefinitions = append(metricDefinitions,
			commonv1.MetricDefinition{Name: ToStringPtr(md.Name), Regex: ToStringPtr(md.Regex)})
	}

	apiContentType, err := getAPIContentType(sagemakerHPOJob.GetTrainingJob().GetAlgorithmSpecification().GetInputContentType())
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "Unsupported input file type [%v]",
			sagemakerHPOJob.GetTrainingJob().GetAlgorithmSpecification().GetInputContentType().String())
	}

	inputModeString := strings.Title(strings.ToLower(sagemakerHPOJob.GetTrainingJob().GetAlgorithmSpecification().GetInputMode().String()))
	tuningStrategyString := strings.Title(strings.ToLower(hpoJobConfig.GetTuningStrategy().String()))
	tuningObjectiveTypeString := strings.Title(strings.ToLower(hpoJobConfig.GetTuningObjective().GetObjectiveType().String()))
	trainingJobEarlyStoppingTypeString := strings.Title(strings.ToLower(hpoJobConfig.TrainingJobEarlyStoppingType.String()))

	role := awsUtils.GetRoleFromSecurityContext(cfg.RoleAnnotationKey, taskCtx.TaskExecutionMetadata())

	if len(role) == 0 {
		role = cfg.RoleArn
	}

	hpoJob := &hpojobv1.HyperparameterTuningJob{
		Spec: hpojobv1.HyperparameterTuningJobSpec{
			HyperParameterTuningJobName: &jobName,
			HyperParameterTuningJobConfig: &commonv1.HyperParameterTuningJobConfig{
				ResourceLimits: &commonv1.ResourceLimits{
					MaxNumberOfTrainingJobs: ToInt64Ptr(sagemakerHPOJob.GetMaxNumberOfTrainingJobs()),
					MaxParallelTrainingJobs: ToInt64Ptr(sagemakerHPOJob.GetMaxParallelTrainingJobs()),
				},
				Strategy: commonv1.HyperParameterTuningJobStrategyType(tuningStrategyString),
				HyperParameterTuningJobObjective: &commonv1.HyperParameterTuningJobObjective{
					Type:       commonv1.HyperParameterTuningJobObjectiveType(tuningObjectiveTypeString),
					MetricName: ToStringPtr(hpoJobConfig.GetTuningObjective().GetMetricName()),
				},
				ParameterRanges:              hpoJobParameterRanges,
				TrainingJobEarlyStoppingType: commonv1.TrainingJobEarlyStoppingType(trainingJobEarlyStoppingTypeString),
			},
			TrainingJobDefinition: &commonv1.HyperParameterTrainingJobDefinition{
				StaticHyperParameters: staticHyperparams,
				AlgorithmSpecification: &commonv1.HyperParameterAlgorithmSpecification{
					TrainingImage:     ToStringPtr(trainingImageStr),
					TrainingInputMode: commonv1.TrainingInputMode(inputModeString),
					MetricDefinitions: metricDefinitions,
					AlgorithmName:     nil,
				},
				InputDataConfig: []commonv1.Channel{
					{
						ChannelName: ToStringPtr(TrainPredefinedInputVariable),
						DataSource: &commonv1.DataSource{
							S3DataSource: &commonv1.S3DataSource{
								S3DataType: "S3Prefix",
								S3Uri:      ToStringPtr(trainPathLiteral.GetScalar().GetBlob().GetUri()),
							},
						},
						ContentType: ToStringPtr(apiContentType), // TODO: can this be derived from the BlobMetadata
						InputMode:   inputModeString,
					},
					{
						ChannelName: ToStringPtr(ValidationPredefinedInputVariable),
						DataSource: &commonv1.DataSource{
							S3DataSource: &commonv1.S3DataSource{
								S3DataType: "S3Prefix",
								S3Uri:      ToStringPtr(validatePathLiteral.GetScalar().GetBlob().GetUri()),
							},
						},
						ContentType: ToStringPtr(apiContentType), // TODO: can this be derived from the BlobMetadata
						InputMode:   inputModeString,
					},
				},
				OutputDataConfig: &commonv1.OutputDataConfig{
					S3OutputPath: ToStringPtr(outputPath),
				},
				ResourceConfig: &commonv1.ResourceConfig{
					InstanceType:   sagemakerHPOJob.GetTrainingJob().GetTrainingJobResourceConfig().GetInstanceType(),
					InstanceCount:  ToInt64Ptr(sagemakerHPOJob.GetTrainingJob().GetTrainingJobResourceConfig().GetInstanceCount()),
					VolumeSizeInGB: ToInt64Ptr(sagemakerHPOJob.GetTrainingJob().GetTrainingJobResourceConfig().GetVolumeSizeInGb()),
					VolumeKmsKeyId: ToStringPtr(""), // TODO: Not yet supported. Need to add to proto and flytekit in the future
				},
				RoleArn: ToStringPtr(role),
				StoppingCondition: &commonv1.StoppingCondition{
					MaxRuntimeInSeconds:  ToInt64Ptr(86400),
					MaxWaitTimeInSeconds: nil,
				},
			},
			Region: ToStringPtr(cfg.Region),
		},
	}

	logger.Infof(ctx, "Successfully built a hyperparameter tuning job resource for task [%v]", taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())
	return hpoJob, nil
}

func (m awsSagemakerPlugin) getTaskPhaseForHyperparameterTuningJob(
	ctx context.Context, pluginContext k8s.PluginContext, hpoJob *hpojobv1.HyperparameterTuningJob) (pluginsCore.PhaseInfo, error) {

	logger.Infof(ctx, "Getting task phase for hyperparameter tuning job [%v]", hpoJob.Status.SageMakerHyperParameterTuningJobName)
	info, err := m.getEventInfoForHyperparameterTuningJob(ctx, hpoJob)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, pluginErrors.Wrapf(pluginErrors.RuntimeFailure, err, "Failed to get event info for the job")
	}

	occurredAt := time.Now()

	switch hpoJob.Status.HyperParameterTuningJobStatus {
	case ReconcilingTuningJobStatus:
		logger.Errorf(ctx, "Job stuck in reconciling status, assuming retryable failure [%s]", hpoJob.Status.Additional)
		// TODO talk to AWS about why there cannot be an explicit condition that signals AWS API call pluginErrors
		execError := &flyteIdlCore.ExecutionError{
			Message: hpoJob.Status.Additional,
			Kind:    flyteIdlCore.ExecutionError_USER,
			Code:    ReconcilingTuningJobStatus,
		}
		return pluginsCore.PhaseInfoFailed(pluginsCore.PhaseRetryableFailure, execError, info), nil
	case sagemaker.HyperParameterTuningJobStatusFailed:
		execError := &flyteIdlCore.ExecutionError{
			Message: hpoJob.Status.Additional,
			Kind:    flyteIdlCore.ExecutionError_USER,
			Code:    sagemaker.HyperParameterTuningJobStatusFailed,
		}
		return pluginsCore.PhaseInfoFailed(pluginsCore.PhasePermanentFailure, execError, info), nil
	case sagemaker.HyperParameterTuningJobStatusStopped:
		return pluginsCore.PhaseInfoRetryableFailure(taskError.DownstreamSystemError, "Hyperparameter tuning job stopped", info), nil
	case sagemaker.HyperParameterTuningJobStatusCompleted:
		// Now that it is a success we will set the outputs as expected by the task

		// TODO:
		// Check task template -> custom training job -> if custom: assume output.pb exist, and fail if it doesn't. If it exists, then
		//						 				      -> if not custom: check model.tar.gz
		out, err := getOutputLiteralMapFromTaskInterface(ctx, pluginContext.TaskReader(),
			createModelOutputPath(hpoJob, pluginContext.OutputWriter().GetRawOutputPrefix().String(),
				*hpoJob.Status.BestTrainingJob.TrainingJobName))
		if err != nil {
			logger.Errorf(ctx, "Failed to create outputs, err: %s", err)
			return pluginsCore.PhaseInfoUndefined, pluginErrors.Wrapf(pluginErrors.BadTaskSpecification, err, "failed to create outputs for the task")
		}
		if err := pluginContext.OutputWriter().Put(ctx, ioutils.NewInMemoryOutputReader(out, nil, nil)); err != nil {
			return pluginsCore.PhaseInfoUndefined, err
		}
		logger.Debugf(ctx, "Successfully produced and returned outputs")
		return pluginsCore.PhaseInfoSuccess(info), nil
	case "":
		return pluginsCore.PhaseInfoQueued(occurredAt, pluginsCore.DefaultPhaseVersion, "job submitted"), nil
	}

	return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, info), nil
}

func (m awsSagemakerPlugin) getEventInfoForHyperparameterTuningJob(ctx context.Context, hpoJob *hpojobv1.HyperparameterTuningJob) (*pluginsCore.TaskInfo, error) {

	var jobRegion, jobName, jobTypeInURL, sagemakerLinkName string
	jobRegion = *hpoJob.Spec.Region
	jobName = *hpoJob.Spec.HyperParameterTuningJobName
	jobTypeInURL = "hyper-tuning-jobs"
	sagemakerLinkName = HyperparameterTuningJobSageMakerLinkName

	logger.Infof(ctx, "Getting event information for SageMaker HyperparameterTuningJob task, job region: [%v], job name: [%v], "+
		"job type in url: [%v], sagemaker link name: [%v]", jobRegion, jobName, jobTypeInURL, sagemakerLinkName)

	return createTaskInfo(ctx, jobRegion, jobName, jobTypeInURL, sagemakerLinkName)
}
