package sagemaker

import (
	"context"
	"fmt"
	"strings"

	"github.com/lyft/flytestdlib/logger"

	"github.com/Masterminds/semver"
	commonv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/common"
	awssagemaker "github.com/aws/amazon-sagemaker-operator-for-k8s/controllers/controllertest"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	sagemakerSpec "github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins/sagemaker"
	"github.com/lyft/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"
	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"
)

func getAPIContentType(fileType sagemakerSpec.InputContentType_Value) (string, error) {
	if fileType == sagemakerSpec.InputContentType_TEXT_CSV {
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

func getTrainingImage(ctx context.Context, job *sagemakerSpec.TrainingJob) (string, error) {
	cfg := config.GetSagemakerConfig()
	var foundAlgorithmCfg *config.PrebuiltAlgorithmConfig
	var foundRegionalCfg *config.RegionalConfig

	if specifiedAlg := job.GetAlgorithmSpecification().GetAlgorithmName(); specifiedAlg != sagemakerSpec.AlgorithmName_CUSTOM {
		// Built-in algorithm mode
		apiAlgorithmName := specifiedAlg.String()

		for _, algorithmCfg := range cfg.PrebuiltAlgorithms {
			if strings.EqualFold(apiAlgorithmName, algorithmCfg.Name) {
				foundAlgorithmCfg = &algorithmCfg
				break
			}
		}
		if foundAlgorithmCfg == nil {
			return "", errors.Errorf("Failed to find a image for algorithm [%v]", apiAlgorithmName)
		}

		for _, regionalCfg := range foundAlgorithmCfg.RegionalConfig {
			if strings.EqualFold(cfg.Region, regionalCfg.Region) {
				foundRegionalCfg = &regionalCfg
				break
			}
		}
		if foundRegionalCfg == nil {
			return "", errors.Errorf("Failed to find a image for algorithm [%v] region [%v]",
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
				logger.Infof(ctx, "Image [%v] is picked for algorithm [] region [] version [] ",
					versionCfg.Image, apiAlgorithmName, cfg.Region, userSpecifiedSemVer)
				return versionCfg.Image, nil
			}
		}
		logger.Errorf(ctx, "Failed to find a image for [%v]:[%v]:[%v]",
			job.GetAlgorithmSpecification().GetAlgorithmName(), cfg.Region, job.GetAlgorithmSpecification().GetAlgorithmVersion())

		return "", errors.Errorf("Failed to find a image for [%v]:[%v]:[%v]",
			job.GetAlgorithmSpecification().GetAlgorithmName(), cfg.Region, job.GetAlgorithmSpecification().GetAlgorithmVersion())
	}
	return "custom image", errors.Errorf("Custom images are not supported yet")
}

func buildParameterRanges(hpoJobConfig *sagemakerSpec.HyperparameterTuningJobConfig) *commonv1.ParameterRanges {
	prMap := hpoJobConfig.GetHyperparameterRanges().GetParameterRangeMap()
	var retValue = &commonv1.ParameterRanges{
		CategoricalParameterRanges: []commonv1.CategoricalParameterRange{},
		ContinuousParameterRanges:  []commonv1.ContinuousParameterRange{},
		IntegerParameterRanges:     []commonv1.IntegerParameterRange{},
	}

	for prName, pr := range prMap {
		scalingTypeString := strings.Title(strings.ToLower(pr.GetContinuousParameterRange().GetScalingType().String()))
		switch pr.GetParameterRangeType().(type) {
		case *sagemakerSpec.ParameterRangeOneOf_CategoricalParameterRange:
			var newElem = commonv1.CategoricalParameterRange{
				Name:   awssagemaker.ToStringPtr(prName),
				Values: pr.GetCategoricalParameterRange().GetValues(),
			}
			retValue.CategoricalParameterRanges = append(retValue.CategoricalParameterRanges, newElem)

		case *sagemakerSpec.ParameterRangeOneOf_ContinuousParameterRange:
			var newElem = commonv1.ContinuousParameterRange{
				MaxValue:    awssagemaker.ToStringPtr(fmt.Sprintf("%f", pr.GetContinuousParameterRange().GetMaxValue())),
				MinValue:    awssagemaker.ToStringPtr(fmt.Sprintf("%f", pr.GetContinuousParameterRange().GetMinValue())),
				Name:        awssagemaker.ToStringPtr(prName),
				ScalingType: commonv1.HyperParameterScalingType(scalingTypeString),
			}
			retValue.ContinuousParameterRanges = append(retValue.ContinuousParameterRanges, newElem)

		case *sagemakerSpec.ParameterRangeOneOf_IntegerParameterRange:
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

func convertHyperparameterTuningJobConfigToSpecType(hpoJobConfigLiteral *core.Literal) (*sagemakerSpec.HyperparameterTuningJobConfig, error) {
	var retValue = &sagemakerSpec.HyperparameterTuningJobConfig{}
	hpoJobConfigByteArray := hpoJobConfigLiteral.GetScalar().GetBinary().GetValue()
	err := proto.Unmarshal(hpoJobConfigByteArray, retValue)
	if err != nil {
		return nil, errors.Errorf("Hyperparameter Tuning Job Config Literal in input cannot be unmarshalled into spec type")
	}
	return retValue, nil
}

func convertStaticHyperparamsLiteralToSpecType(hyperparamLiteral *core.Literal) ([]*commonv1.KeyValuePair, error) {
	var retValue []*commonv1.KeyValuePair
	hyperFields := hyperparamLiteral.GetScalar().GetGeneric().GetFields()
	if hyperFields == nil {
		return nil, errors.Errorf("Failed to get the static hyperparameters field from the literal")
	}
	for k, v := range hyperFields {
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

func createOutputLiteralMap(tk *core.TaskTemplate, outputPath string) *core.LiteralMap {
	op := &core.LiteralMap{}
	for k := range tk.Interface.Outputs.Variables {
		// if v != core.LiteralType_Blob{}
		op.Literals = make(map[string]*core.Literal)
		op.Literals[k] = &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Blob{
						Blob: &core.Blob{
							Metadata: &core.BlobMetadata{
								Type: &core.BlobType{Dimensionality: core.BlobType_SINGLE},
							},
							Uri: outputPath,
						},
					},
				},
			},
		}
	}
	return op
}

func deleteConflictingStaticHyperparameters(
	ctx context.Context,
	staticHPs []*commonv1.KeyValuePair,
	tunableHPMap map[string]*sagemakerSpec.ParameterRangeOneOf) []*commonv1.KeyValuePair {

	finalStaticHPs := make([]*commonv1.KeyValuePair, 0, len(staticHPs))

	for _, hp := range staticHPs {
		if _, found := tunableHPMap[hp.Name]; !found {
			finalStaticHPs = append(finalStaticHPs, hp)
		} else {
			logger.Infof(ctx,
				"Static hyperparameter [%v] is removed because the same hyperparameter can be found in the map of tunable hyperparameters", hp.Name)
		}
	}
	return finalStaticHPs
}
