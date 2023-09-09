package sagemaker

const (
	trainingJobTaskPluginID = "sagemaker_training"
	trainingJobTaskType     = "sagemaker_training_job_task"
)

const (
	customTrainingJobTaskPluginID = "sagemaker_custom_training"
	customTrainingJobTaskType     = "sagemaker_custom_training_job_task"
)

const (
	hyperparameterTuningJobTaskPluginID = "sagemaker_hyperparameter_tuning"
	hyperparameterTuningJobTaskType     = "sagemaker_hyperparameter_tuning_job_task"
)

const (
	TEXTCSVInputContentType string = "text/csv"
)

const (
	FlyteSageMakerEnvVarKeyPrefix         string = "__FLYTE_ENV_VAR_"
	FlyteSageMakerKeySuffix               string = "__"
	FlyteSageMakerCmdKeyPrefix            string = "__FLYTE_CMD_"
	FlyteSageMakerCmdDummyValue           string = "__FLYTE_CMD_DUMMY_VALUE__"
	FlyteSageMakerEnvVarKeyStatsdDisabled string = "FLYTE_STATSD_DISABLED"
	SageMakerMpiEnableEnvVarName          string = "sagemaker_mpi_enabled"
)

const (
	TrainingJobOutputPathSubDir    = "training_outputs"
	HyperparameterOutputPathSubDir = "hyperparameter_tuning_outputs"
)

const (
	// These constants are the default input channel names for built-in algorithms
	// When dealing with built-in algorithm training tasks, the plugin would assume these inputs exist and
	// access these keys in the input literal map
	// The same keys (except for the static hyperparameter key) would also be used when filling out the inputDataConfig fields of the CRD
	TrainPredefinedInputVariable                 = "train"
	ValidationPredefinedInputVariable            = "validation"
	StaticHyperparametersPredefinedInputVariable = "static_hyperparameters"
)

const (
	CloudWatchLogLinkName                    = "CloudWatch Logs"
	TrainingJobSageMakerLinkName             = "SageMaker Built-in Algorithm Training Job"
	CustomTrainingJobSageMakerLinkName       = "SageMaker Custom Training Job"
	HyperparameterTuningJobSageMakerLinkName = "SageMaker Hyperparameter Tuning Job"
)
