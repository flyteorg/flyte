package sagemaker

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	commonv1 "github.com/aws/amazon-sagemaker-operator-for-k8s/api/v1/common"
	sagemakerSpec "github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins/sagemaker"
	"github.com/lyft/flytestdlib/config/viper"
	"github.com/stretchr/testify/assert"

	stdConfig "github.com/lyft/flytestdlib/config"

	"github.com/lyft/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"
	sagemakerConfig "github.com/lyft/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"
)

func generateMockTunableHPMap() map[string]*sagemakerSpec.ParameterRangeOneOf {
	ret := map[string]*sagemakerSpec.ParameterRangeOneOf{
		"hp1": {ParameterRangeType: &sagemakerSpec.ParameterRangeOneOf_IntegerParameterRange{
			IntegerParameterRange: &sagemakerSpec.IntegerParameterRange{
				MaxValue: 10, MinValue: 0, ScalingType: sagemakerSpec.HyperparameterScalingType_AUTO}}},
		"hp2": {ParameterRangeType: &sagemakerSpec.ParameterRangeOneOf_ContinuousParameterRange{
			ContinuousParameterRange: &sagemakerSpec.ContinuousParameterRange{
				MaxValue: 5.0, MinValue: 3.0, ScalingType: sagemakerSpec.HyperparameterScalingType_LINEAR}}},
		"hp3": {ParameterRangeType: &sagemakerSpec.ParameterRangeOneOf_CategoricalParameterRange{
			CategoricalParameterRange: &sagemakerSpec.CategoricalParameterRange{
				Values: []string{"AAA", "BBB", "CCC"}}}},
	}
	return ret
}

func generatePartiallyConflictingStaticHPs() []*commonv1.KeyValuePair {
	ret := []*commonv1.KeyValuePair{
		{Name: "hp1", Value: "100"},
		{Name: "hp4", Value: "0.5"},
		{Name: "hp3", Value: "ddd,eee"},
	}
	return ret
}

func generateTotallyConflictingStaticHPs() []*commonv1.KeyValuePair {
	ret := []*commonv1.KeyValuePair{
		{Name: "hp1", Value: "100"},
		{Name: "hp2", Value: "0.5"},
		{Name: "hp3", Value: "ddd,eee"},
	}
	return ret
}

func generateNonConflictingStaticHPs() []*commonv1.KeyValuePair {
	ret := []*commonv1.KeyValuePair{
		{Name: "hp5", Value: "100"},
		{Name: "hp4", Value: "0.5"},
		{Name: "hp7", Value: "ddd,eee"},
	}
	return ret
}

func generateMockHyperparameterTuningJobConfig() *sagemakerSpec.HyperparameterTuningJobConfig {
	return &sagemakerSpec.HyperparameterTuningJobConfig{
		HyperparameterRanges: &sagemakerSpec.ParameterRanges{ParameterRangeMap: generateMockTunableHPMap()},
		TuningStrategy:       sagemakerSpec.HyperparameterTuningStrategy_BAYESIAN,
		TuningObjective: &sagemakerSpec.HyperparameterTuningObjective{
			ObjectiveType: sagemakerSpec.HyperparameterTuningObjectiveType_MINIMIZE,
			MetricName:    "validate:mse",
		},
		TrainingJobEarlyStoppingType: sagemakerSpec.TrainingJobEarlyStoppingType_AUTO,
	}
}

func generateMockSageMakerConfig() *sagemakerConfig.Config {
	return &sagemakerConfig.Config{
		RoleArn:           "default",
		Region:            "us-east-1",
		RoleAnnotationKey: "role_annotation_key",
		// https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
		PrebuiltAlgorithms: []sagemakerConfig.PrebuiltAlgorithmConfig{
			{
				Name: "xgboost",
				RegionalConfig: []sagemakerConfig.RegionalConfig{
					{
						Region: "us-east-1",
						VersionConfigs: []sagemakerConfig.VersionConfig{
							{
								Version: "0.9.1",
								Image:   "amazonaws.com/xgboost:latest",
							},
						},
					},
				},
			},
		},
	}
}

func Test_deleteConflictingStaticHyperparameters(t *testing.T) {
	mockCtx := context.TODO()
	type args struct {
		ctx          context.Context
		staticHPs    []*commonv1.KeyValuePair
		tunableHPMap map[string]*sagemakerSpec.ParameterRangeOneOf
	}
	tests := []struct {
		name string
		args args
		want []*commonv1.KeyValuePair
	}{
		{name: "Partially conflicting hyperparameter list", args: args{
			ctx:          mockCtx,
			staticHPs:    generatePartiallyConflictingStaticHPs(),
			tunableHPMap: generateMockTunableHPMap(),
		}, want: []*commonv1.KeyValuePair{{Name: "hp4", Value: "0.5"}}},
		{name: "Totally conflicting hyperparameter list", args: args{
			ctx:          mockCtx,
			staticHPs:    generateTotallyConflictingStaticHPs(),
			tunableHPMap: generateMockTunableHPMap(),
		}, want: []*commonv1.KeyValuePair{}},
		{name: "Non-conflicting hyperparameter list", args: args{
			ctx:          mockCtx,
			staticHPs:    generateNonConflictingStaticHPs(),
			tunableHPMap: generateMockTunableHPMap(),
		}, want: []*commonv1.KeyValuePair{{Name: "hp5", Value: "100"}, {Name: "hp4", Value: "0.5"}, {Name: "hp7", Value: "ddd,eee"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deleteConflictingStaticHyperparameters(tt.args.ctx, tt.args.staticHPs, tt.args.tunableHPMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deleteConflictingStaticHyperparameters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildParameterRanges(t *testing.T) {
	type args struct {
		hpoJobConfig *sagemakerSpec.HyperparameterTuningJobConfig
	}
	tests := []struct {
		name string
		args args
		want *commonv1.ParameterRanges
	}{
		{name: "Building a list of a mixture of all three types of parameter ranges",
			args: args{hpoJobConfig: generateMockHyperparameterTuningJobConfig()},
			want: &commonv1.ParameterRanges{
				CategoricalParameterRanges: []commonv1.CategoricalParameterRange{{Name: ToStringPtr("hp3"), Values: []string{"AAA", "BBB", "CCC"}}},
				ContinuousParameterRanges:  []commonv1.ContinuousParameterRange{{Name: ToStringPtr("hp2"), MinValue: ToStringPtr("3.0"), MaxValue: ToStringPtr("5.0")}},
				IntegerParameterRanges:     []commonv1.IntegerParameterRange{{Name: ToStringPtr("hp1"), MinValue: ToStringPtr("0"), MaxValue: ToStringPtr("10")}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildParameterRanges(tt.args.hpoJobConfig)

			wantCatPr := tt.want.CategoricalParameterRanges[0]
			gotCatPr := got.CategoricalParameterRanges[0]
			if *wantCatPr.Name != *gotCatPr.Name || !reflect.DeepEqual(wantCatPr.Values, gotCatPr.Values) {
				t.Errorf("buildParameterRanges(): CategoricalParameterRange: got [Name: %v, Value: %v], want [Name: %v, Value: %v]",
					*gotCatPr.Name, gotCatPr.Values, *wantCatPr.Name, wantCatPr.Values)
			}
			wantIntPr := tt.want.IntegerParameterRanges[0]
			gotIntPr := got.IntegerParameterRanges[0]
			if *wantIntPr.Name != *gotIntPr.Name || *wantIntPr.MinValue != *gotIntPr.MinValue || *wantIntPr.MaxValue != *gotIntPr.MaxValue {
				t.Errorf("buildParameterRanges(): IntegerParameterRange: got [Name: %v, MinValue: %v, MaxValue: %v], want [Name: %v, MinValue: %v, MaxValue: %v]",
					*gotIntPr.Name, *gotIntPr.MinValue, *gotIntPr.MaxValue, *wantIntPr.Name, *wantIntPr.MinValue, *wantIntPr.MaxValue)
			}
			wantConPr := tt.want.ContinuousParameterRanges[0]
			gotConPr := got.ContinuousParameterRanges[0]
			wantMin, _ := strconv.ParseFloat(*wantConPr.MinValue, 64)
			gotMin, err := strconv.ParseFloat(*gotConPr.MinValue, 64)
			if err != nil {
				t.Errorf("buildParameterRanges(): ContinuousParameterRange: got invalid min value [%v]", gotMin)
			}
			wantMax, _ := strconv.ParseFloat(*wantConPr.MaxValue, 64)
			gotMax, err := strconv.ParseFloat(*gotConPr.MaxValue, 64)
			if err != nil {
				t.Errorf("buildParameterRanges(): ContinuousParameterRange: got invalid max value [%v]", gotMax)
			}
			if *wantConPr.Name != *gotConPr.Name || wantMin != gotMin || wantMax != gotMax {
				t.Errorf("buildParameterRanges(): ContinuousParameterRange: got [Name: %v, MinValue: %v, MaxValue: %v], want [Name: %v, MinValue: %v, MaxValue: %v]",
					*gotConPr.Name, gotMin, gotMax, *wantConPr.Name, wantMin, wantMax)
			}
		})
	}
}

func Test_getLatestTrainingImage(t *testing.T) {
	type args struct {
		versionConfigs []config.VersionConfig
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "minor version", args: args{versionConfigs: []config.VersionConfig{
			{Version: "0.9", Image: "image1"}, {Version: "0.92", Image: "image2"}, {Version: "0.9.2", Image: "image3"},
		}}, want: "image2", wantErr: false},
		{name: "patch version", args: args{versionConfigs: []config.VersionConfig{
			{Version: "0.9", Image: "image1"}, {Version: "0.9.2", Image: "image3"},
		}}, want: "image3", wantErr: false},
		{name: "major version", args: args{versionConfigs: []config.VersionConfig{
			{Version: "1.0.0-3", Image: "image1"}, {Version: "0.9.2", Image: "image3"},
		}}, want: "image1", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLatestTrainingImage(tt.args.versionConfigs)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLatestTrainingImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getLatestTrainingImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTrainingImage(t *testing.T) {
	ctx := context.TODO()

	_ = sagemakerConfig.SetSagemakerConfig(generateMockSageMakerConfig())

	type args struct {
		ctx context.Context
		job *sagemakerSpec.TrainingJob
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "xgboost version found", args: args{ctx: ctx, job: &sagemakerSpec.TrainingJob{
			AlgorithmSpecification: &sagemakerSpec.AlgorithmSpecification{
				InputMode:         0,
				AlgorithmName:     sagemakerSpec.AlgorithmName_XGBOOST,
				AlgorithmVersion:  "0.9.1",
				MetricDefinitions: nil,
				InputContentType:  0,
			},
			TrainingJobResourceConfig: nil,
		}}, want: "amazonaws.com/xgboost:latest", wantErr: false},
		{name: "xgboost version not found", args: args{ctx: ctx, job: &sagemakerSpec.TrainingJob{
			AlgorithmSpecification: &sagemakerSpec.AlgorithmSpecification{
				InputMode:         0,
				AlgorithmName:     sagemakerSpec.AlgorithmName_XGBOOST,
				AlgorithmVersion:  "0.7",
				MetricDefinitions: nil,
				InputContentType:  0,
			},
			TrainingJobResourceConfig: nil,
		}}, want: "", wantErr: true},
		{name: "custom", args: args{ctx: ctx, job: &sagemakerSpec.TrainingJob{
			AlgorithmSpecification: &sagemakerSpec.AlgorithmSpecification{
				InputMode:         0,
				AlgorithmName:     sagemakerSpec.AlgorithmName_CUSTOM,
				AlgorithmVersion:  "0.7",
				MetricDefinitions: nil,
				InputContentType:  0,
			},
			TrainingJobResourceConfig: nil,
		}}, want: "custom image", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTrainingImage(tt.args.ctx, tt.args.job)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTrainingImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getTrainingImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTrainingImage_LoadConfig(t *testing.T) {
	configAccessor := viper.NewAccessor(stdConfig.Options{
		StrictMode:  true,
		SearchPaths: []string{"testdata/config.yaml"},
	})

	err := configAccessor.UpdateConfig(context.TODO())
	assert.NoError(t, err)

	assert.NotNil(t, config.GetSagemakerConfig())

	image, err := getTrainingImage(context.TODO(), &sagemakerSpec.TrainingJob{AlgorithmSpecification: &sagemakerSpec.AlgorithmSpecification{
		AlgorithmName:    sagemakerSpec.AlgorithmName_XGBOOST,
		AlgorithmVersion: "0.90",
	}})

	assert.NoError(t, err)
	assert.Equal(t, "image-0.90", image)

	image, err = getTrainingImage(context.TODO(), &sagemakerSpec.TrainingJob{AlgorithmSpecification: &sagemakerSpec.AlgorithmSpecification{
		AlgorithmName:    sagemakerSpec.AlgorithmName_XGBOOST,
		AlgorithmVersion: "1.0",
	}})

	assert.NoError(t, err)
	assert.Equal(t, "image-1.0", image)
}
