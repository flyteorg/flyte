package register

//go:generate pflags FilesConfig --default-var DefaultFilesConfig --bind-default-var

var (
	DefaultFilesConfig = &FilesConfig{
		Version:         "v1",
		ContinueOnError: false,
	}
)

// FilesConfig containing flags used for registration
type FilesConfig struct {
	Version              string `json:"version" pflag:",version of the entity to be registered with flyte."`
	ContinueOnError      bool   `json:"continueOnError" pflag:",continue on error when registering files."`
	Archive              bool   `json:"archive" pflag:",pass in archive file either an http link or local path."`
	AssumableIamRole     string `json:"assumableIamRole" pflag:", custom assumable iam auth role to register launch plans with."`
	K8ServiceAccount     string `json:"k8ServiceAccount" pflag:", custom kubernetes service account auth role to register launch plans with."`
	OutputLocationPrefix string `json:"outputLocationPrefix" pflag:", custom output location prefix for offloaded types (files/schemas)."`
	SourceUploadPath     string `json:"sourceUploadPath" pflag:", Location for source code in storage."`
	DryRun               bool   `json:"dryRun" pflag:",execute command without making any modifications."`
}
