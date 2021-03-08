package config

import pluginsConfig "github.com/flyteorg/flyteplugins/go/tasks/config"

//go:generate pflags Config --default-var=defaultConfig

const sagemakerConfigSectionKey = "sagemaker"

var (
	defaultConfig = Config{
		RoleArn: "default_role",
		Region:  "us-east-1",
		// https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
		PrebuiltAlgorithms: []PrebuiltAlgorithmConfig{
			{
				Name: "xgboost",
				RegionalConfig: []RegionalConfig{
					{
						Region: "us-east-1",
						VersionConfigs: []VersionConfig{
							{
								Version: "0.90",
								Image:   "683313688378.dkr.ecr.us-east-1.amazonaws.com/sagemaker-xgboost:0.90-2-cpu-py3",
							},
						},
					},
				},
			},
		},
	}

	sagemakerConfigSection = pluginsConfig.MustRegisterSubSection(sagemakerConfigSectionKey, &defaultConfig)
)

// Sagemaker plugin configs
type Config struct {
	RoleArn            string                    `json:"roleArn" pflag:",The role the SageMaker plugin uses to communicate with the SageMaker service"`
	Region             string                    `json:"region" pflag:",The AWS region the SageMaker plugin communicates to"`
	RoleAnnotationKey  string                    `json:"roleAnnotationKey" pflag:",Map key to use to lookup role from task annotations."`
	PrebuiltAlgorithms []PrebuiltAlgorithmConfig `json:"prebuiltAlgorithms" pflag:"-,A List of PrebuiltAlgorithm configs"`
}
type PrebuiltAlgorithmConfig struct {
	Name           string           `json:"name" pflag:",The name of the ML algorithm. Should match Sagemaker"`
	RegionalConfig []RegionalConfig `json:"regionalConfigs" pflag:"-,Per region specific configuration for this algorithm"`
}
type RegionalConfig struct {
	Region         string          `json:"region" pflag:",Region for which this config is applicable"`
	VersionConfigs []VersionConfig `json:"versionConfigs" pflag:",Configuration for various versions of the algorithms'"`
}
type VersionConfig struct {
	Version string `json:"version" pflag:",version of the algorithm"`
	Image   string `json:"image" pflag:",Image URI of the algorithm"`
}

// Retrieves the current config value or default.
func GetSagemakerConfig() *Config {
	return sagemakerConfigSection.GetConfig().(*Config)
}

func SetSagemakerConfig(cfg *Config) error {
	return sagemakerConfigSection.SetConfig(cfg)
}

func ResetSagemakerConfig() error {
	return sagemakerConfigSection.SetConfig(&defaultConfig)
}
