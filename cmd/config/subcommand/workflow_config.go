package subcommand

//go:generate pflags WorkflowConfig --default-var DefaultWorklfowConfig

var (
	DefaultWorklfowConfig = &WorkflowConfig{}
)

// WorkflowConfig commandline configuration
type WorkflowConfig struct {
	Version string `json:"version" pflag:",version of the workflow to be fetched."`
	Latest  bool   `json:"latest" pflag:", flag to indicate to fetch the latest version, version flag will be ignored in this case"`
}
