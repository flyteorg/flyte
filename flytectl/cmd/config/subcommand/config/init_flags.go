package config

//go:generate pflags Config --default-var DefaultConfig --bind-default-var
var (
	DefaultConfig = &Config{
		Insecure: false,
		Force:    false,
	}
)

// Configs
type Config struct {
	Console  string `json:"console" pflag:",Endpoint of console, if different than flyte admin"`
	Host     string `json:"host" pflag:",Endpoint of flyte admin"`
	Insecure bool   `json:"insecure" pflag:",Enable insecure mode"`
	Force    bool   `json:"force" pflag:",Force to overwrite the default config file without confirmation"`
}
