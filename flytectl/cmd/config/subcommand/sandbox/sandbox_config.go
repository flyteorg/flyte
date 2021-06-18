package sandbox

//go:generate pflags Config --default-var DefaultConfig
var (
	DefaultConfig = &Config{}
)

// Config
type Config struct {
	SnacksRepo string `json:"flytesnacks" pflag:", Path of your flytesnacks repository"`
}
