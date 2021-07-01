package sandbox

//go:generate pflags Config --default-var DefaultConfig
var (
	DefaultConfig = &Config{}
)

//Config
type Config struct {
	Source string `json:"source" pflag:", Path of your source code"`
}
