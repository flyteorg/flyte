package docker

//go:generate pflags Config --default-var DefaultConfig --bind-default-var
var (
	DefaultConfig = &Config{
		Force: false,
	}
)

// Configs
type Config struct {
	Force bool `json:"force" pflag:",Optional. Forcefully delete existing sandbox cluster if it exists."`
}
