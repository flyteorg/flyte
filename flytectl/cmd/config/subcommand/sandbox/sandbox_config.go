package sandbox

//go:generate pflags Config --default-var DefaultConfig --bind-default-var
var (
	DefaultConfig = &Config{}
)

//Config
type Config struct {
	Source string `json:"source" pflag:",Path of your source code"`

	// Flytectl sandbox only supports flyte version available in Github release https://github.com/flyteorg/flyte/tags
	// Flytectl sandbox will only work for v0.10.0+
	// Default value dind represents the latest release
	Version string `json:"version" pflag:",Version of flyte. Only supports flyte releases greater than v0.10.0"`

	// Optionally it is possible to specify a specific fqn for the docker image with the tag. This should be
	// Flyte compliant sandbox image. Usually useful, if you want to push the image to your own registry and relaunch
	// from there.
	Image string `json:"image" pflag:",Optional. Provide a fully qualified path to a Flyte compliant docker image."`
}
