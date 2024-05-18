package docker

// Config holds configuration flags for docker command.
var (
	DefaultConfig = &Config{
		Force: false,
	}
)

type Config struct {
	Force bool `json:"force" pflag:",Optional. Forcefully delete existing sandbox cluster if it exists."`
}
