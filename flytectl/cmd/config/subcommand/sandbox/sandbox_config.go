package sandbox

//go:generate pflags Config --default-var DefaultConfig
var (
	DefaultConfig = &Config{}
)

// Config represents the config parameters exposed for the `sandbox` command.
type Config struct {
	SourcesPath string `json:"sourcesPath" pflag:",Path to your source code path where flyte workflows and tasks are."`
}
