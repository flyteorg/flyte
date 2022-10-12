package sandbox

import "github.com/flyteorg/flytectl/pkg/docker"

// Config holds configuration flags for sandbox command.
type Config struct {
	Source string `json:"source" pflag:",Path of your source code"`

	// Flytectl sandbox only supports Flyte version available in Github release https://github.com/flyteorg/flyte/tags.
	// Flytectl sandbox will only work for v0.10.0+.
	// Default value dind represents the latest release.
	Version string `json:"version" pflag:",Version of flyte. Only supports flyte releases greater than v0.10.0"`

	// Optionally it is possible to specify a specific fqn for the docker image with the tag. This should be
	// Flyte compliant sandbox image. Usually useful, if you want to push the image to your own registry and relaunch
	// from there.
	Image string `json:"image" pflag:",Optional. Provide a fully qualified path to a Flyte compliant docker image."`

	// Default value false represents that Flytectl will not use the latest pre-release if it exists.
	Prerelease bool `json:"pre" pflag:",Optional. Pre release Version of flyte will be used for sandbox."`

	// Optionally it is possible to pass in environment variables to sandbox container.
	Env []string `json:"env" pflag:",Optional. Provide Env variable in key=value format which can be passed to sandbox container."`

	// Optionally it is possible to use local sandbox image
	// Flytectl will not pull the image from the registry if the local flag passes. It is usually useful while testing your local images without pushing them to a registry.
	ImagePullPolicy docker.ImagePullPolicy `json:"imagePullPolicy" pflag:",Optional. Defines the image pull behavior [Always/IfNotPresent/Never]"`

	ImagePullOptions docker.ImagePullOptions `json:"imagePullOptions" pflag:",Optional. Defines image pull options (e.g. auth)"`

	// It's used for development. Users are able to start flyte locally via single binary and save the data to the minio or postgres in the sandbox.
	Dev bool `json:"dev" pflag:",Optional. Only start minio and postgres in the sandbox."`
}

//go:generate pflags Config --default-var DefaultConfig --bind-default-var
var (
	DefaultConfig = &Config{}
)
