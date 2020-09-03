package sagemaker

import (
	"context"
	"fmt"
	"sort"
	"strings"

	pluginErrors "github.com/lyft/flyteplugins/go/tasks/errors"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/utils"

	pluginsCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/lyft/flytestdlib/logger"

	"github.com/Masterminds/semver"
	commonv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/common"
	awssagemaker "github.com/aws/amazon-sagemaker-operator-for-k8s/controllers/controllertest"
	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	flyteIdlCore "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	flyteSagemakerIdl "github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins/sagemaker"
	"github.com/lyft/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"
	"github.com/pkg/errors"
)

func getAPIContentType(fileType flyteSagemakerIdl.InputContentType_Value) (string, error) {
	if fileType == flyteSagemakerIdl.InputContentType_TEXT_CSV {
		return TEXTCSVInputContentType, nil
	}
	return "", errors.Errorf("Unsupported input file type [%v]", fileType.String())
}

func getLatestTrainingImage(versionConfigs []config.VersionConfig) (string, error) {
	latestSemVer, _ := semver.NewVersion("0.0.0")
	latestImg := ""
	for _, verCfg := range versionConfigs {
		semVer, err := semver.NewVersion(verCfg.Version)
		if err != nil {
			return "", errors.Wrapf(err, "Failed to cast version [%v] to a semver", verCfg.Version)
		}
		if semVer.GreaterThan(latestSemVer) {
			latestSemVer = semVer
			latestImg = verCfg.Image
		}
	}
	if latestImg == "" {
		return "", errors.Errorf("Failed to find the latest image")
	}
	return latestImg, nil
}

func getTrainingJobImage(ctx context.Context, _ pluginsCore.TaskExecutionContext, job *flyteSagemakerIdl.TrainingJob) (string, error) {
	image, err := getPrebuiltTrainingImage(ctx, job)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to get prebuilt image for job [%v]", *job)
	}
	return image, nil
}

func getPrebuiltTrainingImage(ctx context.Context, job *flyteSagemakerIdl.TrainingJob) (string, error) {
	// This function determines which image URI to put into the CRD of the training job and the hyperparameter tuning job

	cfg := config.GetSagemakerConfig()
	var foundAlgorithmCfg *config.PrebuiltAlgorithmConfig
	var foundRegionalCfg *config.RegionalConfig

	if specifiedAlg := job.GetAlgorithmSpecification().GetAlgorithmName(); specifiedAlg != flyteSagemakerIdl.AlgorithmName_CUSTOM {
		// Built-in algorithm mode
		apiAlgorithmName := specifiedAlg.String()

		for _, algorithmCfg := range cfg.PrebuiltAlgorithms {
			if strings.EqualFold(apiAlgorithmName, algorithmCfg.Name) {
				foundAlgorithmCfg = &algorithmCfg
				break
			}
		}
		if foundAlgorithmCfg == nil {
			return "", errors.Errorf("Failed to find an image for algorithm [%v]", apiAlgorithmName)
		}

		for _, regionalCfg := range foundAlgorithmCfg.RegionalConfig {
			if strings.EqualFold(cfg.Region, regionalCfg.Region) {
				foundRegionalCfg = &regionalCfg
				break
			}
		}
		if foundRegionalCfg == nil {
			return "", errors.Errorf("Failed to find an image for algorithm [%v] region [%v]",
				job.GetAlgorithmSpecification().GetAlgorithmName(), cfg.Region)
		}

		userSpecifiedVer := job.GetAlgorithmSpecification().GetAlgorithmVersion()
		// If the user does not specify a version -> use the latest version found in the config possible
		if userSpecifiedVer == "" {
			logger.Infof(ctx, "The version of the algorithm [%v] is not specified. "+
				"The plugin will try to pick the latest version available for the algorithm-region combination.", userSpecifiedVer, apiAlgorithmName, cfg.Region)
			latestTrainingImage, err := getLatestTrainingImage(foundRegionalCfg.VersionConfigs)
			if err != nil {
				return "", errors.Wrapf(err, "Failed to identify the latest image for algorithm:region [%v:%v]",
					apiAlgorithmName, cfg.Region)
			}
			return latestTrainingImage, nil
		}
		// If the user specified a version -> we have to translate it to semver and find an exact match
		for _, versionCfg := range foundRegionalCfg.VersionConfigs {
			configSemVer, err := semver.NewVersion(versionCfg.Version)
			if err != nil {
				return "", errors.Wrapf(err, "Unable to cast version listed in the config [%v] to a semver", versionCfg.Version)
			}
			userSpecifiedSemVer, err := semver.NewVersion(userSpecifiedVer)
			if err != nil {
				return "", errors.Wrapf(err, "Unable to cast version specified by the user [%v] to a semver", userSpecifiedVer)
			}
			if configSemVer.Equal(userSpecifiedSemVer) {
				logger.Infof(ctx, "Image [%v] is picked for algorithm [%v] region [%v] version [%v] ",
					versionCfg.Image, apiAlgorithmName, cfg.Region, userSpecifiedSemVer)
				return versionCfg.Image, nil
			}
		}
		logger.Errorf(ctx, "Failed to find an image for [%v]:[%v]:[%v]",
			job.GetAlgorithmSpecification().GetAlgorithmName(), cfg.Region, job.GetAlgorithmSpecification().GetAlgorithmVersion())

		return "", errors.Errorf("Failed to find an image for [%v]:[%v]:[%v]",
			job.GetAlgorithmSpecification().GetAlgorithmName(), cfg.Region, job.GetAlgorithmSpecification().GetAlgorithmVersion())
	}
	// Custom image
	return "", errors.Errorf("It is invalid to try getting a prebuilt image for AlgorithmName == CUSTOM ")
}

func buildParameterRanges(hpoJobConfig *flyteSagemakerIdl.HyperparameterTuningJobConfig) *commonv1.ParameterRanges {
	prMap := hpoJobConfig.GetHyperparameterRanges().GetParameterRangeMap()
	var retValue = &commonv1.ParameterRanges{
		CategoricalParameterRanges: []commonv1.CategoricalParameterRange{},
		ContinuousParameterRanges:  []commonv1.ContinuousParameterRange{},
		IntegerParameterRanges:     []commonv1.IntegerParameterRange{},
	}

	for prName, pr := range prMap {
		scalingTypeString := strings.Title(strings.ToLower(pr.GetContinuousParameterRange().GetScalingType().String()))
		switch pr.GetParameterRangeType().(type) {
		case *flyteSagemakerIdl.ParameterRangeOneOf_CategoricalParameterRange:
			var newElem = commonv1.CategoricalParameterRange{
				Name:   awssagemaker.ToStringPtr(prName),
				Values: pr.GetCategoricalParameterRange().GetValues(),
			}
			retValue.CategoricalParameterRanges = append(retValue.CategoricalParameterRanges, newElem)

		case *flyteSagemakerIdl.ParameterRangeOneOf_ContinuousParameterRange:
			var newElem = commonv1.ContinuousParameterRange{
				MaxValue:    awssagemaker.ToStringPtr(fmt.Sprintf("%f", pr.GetContinuousParameterRange().GetMaxValue())),
				MinValue:    awssagemaker.ToStringPtr(fmt.Sprintf("%f", pr.GetContinuousParameterRange().GetMinValue())),
				Name:        awssagemaker.ToStringPtr(prName),
				ScalingType: commonv1.HyperParameterScalingType(scalingTypeString),
			}
			retValue.ContinuousParameterRanges = append(retValue.ContinuousParameterRanges, newElem)

		case *flyteSagemakerIdl.ParameterRangeOneOf_IntegerParameterRange:
			var newElem = commonv1.IntegerParameterRange{
				MaxValue:    awssagemaker.ToStringPtr(fmt.Sprintf("%d", pr.GetIntegerParameterRange().GetMaxValue())),
				MinValue:    awssagemaker.ToStringPtr(fmt.Sprintf("%d", pr.GetIntegerParameterRange().GetMinValue())),
				Name:        awssagemaker.ToStringPtr(prName),
				ScalingType: commonv1.HyperParameterScalingType(scalingTypeString),
			}
			retValue.IntegerParameterRanges = append(retValue.IntegerParameterRanges, newElem)
		}
	}

	return retValue
}

func convertHyperparameterTuningJobConfigToSpecType(hpoJobConfigLiteral *core.Literal) (*flyteSagemakerIdl.HyperparameterTuningJobConfig, error) {
	var retValue = &flyteSagemakerIdl.HyperparameterTuningJobConfig{}
	if hpoJobConfigLiteral.GetScalar() == nil || hpoJobConfigLiteral.GetScalar().GetBinary() == nil {
		return nil, errors.Errorf("[Hyperparameters] should be of type [Scalar.Binary]")
	}
	hpoJobConfigByteArray := hpoJobConfigLiteral.GetScalar().GetBinary().GetValue()
	err := proto.Unmarshal(hpoJobConfigByteArray, retValue)
	if err != nil {
		return nil, errors.Errorf("Hyperparameter Tuning Job Config Literal in input cannot be unmarshalled into spec type")
	}
	return retValue, nil
}

func convertStaticHyperparamsLiteralToSpecType(hyperparamLiteral *core.Literal) ([]*commonv1.KeyValuePair, error) {
	var retValue []*commonv1.KeyValuePair
	if hyperparamLiteral.GetScalar() == nil || hyperparamLiteral.GetScalar().GetGeneric() == nil {
		return nil, errors.Errorf("[Hyperparameters] should be of type [Scalar.Generic]")
	}
	hyperFields := hyperparamLiteral.GetScalar().GetGeneric().GetFields()
	if hyperFields == nil {
		return nil, errors.Errorf("Failed to get the static hyperparameters field from the literal")
	}

	keys := make([]string, 0)
	for k := range hyperFields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("[%v]", keys)
	for _, k := range keys {
		v := hyperFields[k]
		var newElem = commonv1.KeyValuePair{
			Name:  k,
			Value: v.GetStringValue(),
		}
		retValue = append(retValue, &newElem)
	}

	return retValue, nil
}

func ToStringPtr(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

func ToInt64Ptr(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func ToIntPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func ToFloat64Ptr(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

func deleteConflictingStaticHyperparameters(
	ctx context.Context,
	staticHPs []*commonv1.KeyValuePair,
	tunableHPMap map[string]*flyteSagemakerIdl.ParameterRangeOneOf) []*commonv1.KeyValuePair {

	resolvedStaticHPs := make([]*commonv1.KeyValuePair, 0, len(staticHPs))

	for _, hp := range staticHPs {
		if _, found := tunableHPMap[hp.Name]; !found {
			resolvedStaticHPs = append(resolvedStaticHPs, hp)
		} else {
			logger.Infof(ctx,
				"Static hyperparameter [%v] is removed because the same hyperparameter can be found in the map of tunable hyperparameters", hp.Name)
		}
	}
	return resolvedStaticHPs
}

func makeHyperparametersKeysValuesFromArgs(_ context.Context, args []string) []*commonv1.KeyValuePair {
	ret := make([]*commonv1.KeyValuePair, 0)
	for argOrder, arg := range args {
		ret = append(ret, &commonv1.KeyValuePair{
			Name:  fmt.Sprintf("%s%d_%s%s", FlyteSageMakerCmdKeyPrefix, argOrder, arg, FlyteSageMakerKeySuffix),
			Value: FlyteSageMakerCmdDummyValue,
		})
	}
	return ret
}

func injectTaskTemplateEnvVarToHyperparameters(ctx context.Context, taskTemplate *flyteIdlCore.TaskTemplate, hps []*commonv1.KeyValuePair) ([]*commonv1.KeyValuePair, error) {
	if taskTemplate == nil || taskTemplate.GetContainer() == nil {
		return hps, errors.Errorf("The taskTemplate is nil or the container is nil")
	}

	if hps == nil {
		return nil, errors.Errorf("A nil slice of hyperparameters is passed in")
	}

	for _, ev := range taskTemplate.GetContainer().GetEnv() {
		hpKey := fmt.Sprintf("%s%s%s", FlyteSageMakerEnvVarKeyPrefix, ev.Key, FlyteSageMakerKeySuffix)
		logger.Infof(ctx, "Injecting env var {%v: %v} into the hyperparameter list", hpKey, ev.Value)
		hps = append(hps, &commonv1.KeyValuePair{
			Name:  hpKey,
			Value: ev.Value})
	}

	return hps, nil
}

func injectArgsAndEnvVars(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, taskTemplate *flyteIdlCore.TaskTemplate) ([]*commonv1.KeyValuePair, error) {
	templateArgs := taskTemplate.GetContainer().GetArgs()
	templateArgs, err := utils.ReplaceTemplateCommandArgs(ctx, templateArgs, taskCtx.InputReader(), taskCtx.OutputWriter())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to de-template the hyperparameter values")
	}
	hyperParameters := makeHyperparametersKeysValuesFromArgs(ctx, templateArgs)
	hyperParameters, err = injectTaskTemplateEnvVarToHyperparameters(ctx, taskTemplate, hyperParameters)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to inject the task template's container env vars to the hyperparameter list")
	}
	return hyperParameters, nil
}

func checkIfRequiredInputLiteralsExist(inputLiterals map[string]*flyteIdlCore.Literal, inputKeys []string) error {
	for _, inputKey := range inputKeys {
		_, ok := inputLiterals[inputKey]
		if !ok {
			return errors.Errorf("Required input not specified: [%v]", inputKey)
		}
	}
	return nil
}

func getTaskTemplate(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (*flyteIdlCore.TaskTemplate, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to fetch task specification")
	} else if taskTemplate == nil {
		return nil, errors.Errorf("nil task specification")
	}
	return taskTemplate, nil
}

func createTaskInfo(_ context.Context, jobRegion string, jobName string, jobTypeInURL string, sagemakerLinkName string) (*pluginsCore.TaskInfo, error) {
	cwLogURL := fmt.Sprintf("https://%s.console.aws.amazon.com/cloudwatch/home?region=%s#logStream:group=/aws/sagemaker/TrainingJobs;prefix=%s;streamFilter=typeLogStreamPrefix",
		jobRegion, jobRegion, jobName)
	smLogURL := fmt.Sprintf("https://%s.console.aws.amazon.com/sagemaker/home?region=%s#/%s/%s",
		jobRegion, jobRegion, jobTypeInURL, jobName)

	taskLogs := []*flyteIdlCore.TaskLog{
		{
			Uri:           cwLogURL,
			Name:          CloudWatchLogLinkName,
			MessageFormat: flyteIdlCore.TaskLog_JSON,
		},
		{
			Uri:           smLogURL,
			Name:          sagemakerLinkName,
			MessageFormat: flyteIdlCore.TaskLog_UNKNOWN,
		},
	}

	customInfoMap := make(map[string]string)

	customInfo, err := utils.MarshalObjToStruct(customInfoMap)
	if err != nil {
		return nil, pluginErrors.Wrapf(pluginErrors.RuntimeFailure, err, "Unable to create a custom info object")
	}

	return &pluginsCore.TaskInfo{
		Logs:       taskLogs,
		CustomInfo: customInfo,
	}, nil
}
