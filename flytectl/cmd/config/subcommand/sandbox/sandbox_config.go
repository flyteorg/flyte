package sandbox

//go:generate enumer -type=ImagePullPolicy -trimprefix=ImagePullPolicy --json
type ImagePullPolicy int

const (
	ImagePullPolicyAlways ImagePullPolicy = iota
	ImagePullPolicyIfNotPresent
	ImagePullPolicyNever
)

// Set implements PFlag's Value interface to attempt to set the value of the flag from string.
func (i *ImagePullPolicy) Set(val string) error {
	policy, err := ImagePullPolicyString(val)
	if err != nil {
		return err
	}

	*i = policy
	return nil
}

// Type implements PFlag's Value interface to return type name.
func (i ImagePullPolicy) Type() string {
	return "ImagePullPolicy"
}

//go:generate pflags Config --default-var DefaultConfig --bind-default-var
var (
	DefaultConfig = &Config{}
)

//Config holds configuration flags for sandbox command.
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

	// Default value false represents that flytectl will not use latest pre release if exist
	Prerelease bool `json:"pre" pflag:",Optional. Pre release Version of flyte will be used for sandbox."`

	// Optionally it is possible to pass in environment variables to sandbox container.
	Env []string `json:"env" pflag:",Optional. Provide Env variable in key=value format which can be passed to sandbox container."`

	// Optionally it is possible to use local sandbox image
	// If local flag pass then flytectl will not pull image from registry. Usually useful, if you want to test your local images without pushing them to a registry
	ImagePullPolicy ImagePullPolicy `json:"imagePullPolicy" pflag:",Optional. Defines the image pull behavior [Always/IfNotPresent/Never]"`
}
