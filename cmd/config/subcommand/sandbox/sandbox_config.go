package sandbox

//go:generate pflags Config --default-var DefaultConfig --bind-default-var
var (
	DefaultConfig = &Config{}
)

//Config
type Config struct {
	Source string `json:"source" pflag:",Path of your source code"`

	// Flytectl sandbox only support flyte version available in Github release https://github.com/flyteorg/flyte/tags
	// Flytectl sandbox will only work for v0.10.0+
	// Default value dind represent the latest release
	Version string `json:"version" pflag:",Version of flyte. Only support v0.10.0+ flyte release"`
}
