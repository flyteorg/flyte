package register

import "github.com/flyteorg/flytestdlib/config"

//go:generate pflags FilesConfig --default-var DefaultFilesConfig --bind-default-var

var (
	DefaultFilesConfig = &FilesConfig{
		Version:         "",
		ContinueOnError: false,
	}

	cfg = config.MustRegisterSection("files", DefaultFilesConfig)
)

// FilesConfig containing flags used for registration
type FilesConfig struct {
	Version                    string `json:"version" pflag:",Version of the entity to be registered with flyte which are un-versioned after serialization."`
	Force                      bool   `json:"force" pflag:",Force use of version number on entities registered with flyte."`
	ContinueOnError            bool   `json:"continueOnError" pflag:",Continue on error when registering files."`
	Archive                    bool   `json:"archive" pflag:",Pass in archive file either an http link or local path."`
	AssumableIamRole           string `json:"assumableIamRole" pflag:",Custom assumable iam auth role to register launch plans with."`
	K8sServiceAccount          string `json:"k8sServiceAccount" pflag:",Custom kubernetes service account auth role to register launch plans with."`
	K8ServiceAccount           string `json:"k8ServiceAccount" pflag:",Deprecated. Please use --K8sServiceAccount"`
	OutputLocationPrefix       string `json:"outputLocationPrefix" pflag:",Custom output location prefix for offloaded types (files/schemas)."`
	DeprecatedSourceUploadPath string `json:"sourceUploadPath" pflag:",Deprecated: Update flyte admin to avoid having to configure storage access from flytectl."`
	DestinationDirectory       string `json:"destinationDirectory" pflag:",Location of source code in container."`
	DryRun                     bool   `json:"dryRun" pflag:",Execute command without making any modifications."`
}

func GetConfig() *FilesConfig {
	return cfg.GetConfig().(*FilesConfig)
}
