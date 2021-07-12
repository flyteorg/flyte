package config

//go:generate pflags Config --default-var DefaultConfig --bind-default-var
var (
	DefaultConfig = &Config{
		Insecure: true,
		Storage:  false,
	}
)

//Configs
type Config struct {
	Host     string `json:"host" pflag:",Endpoint of flyte admin"`
	Insecure bool   `json:"insecure" pflag:",Enable insecure mode"`
	Storage  bool   `json:"storage" pflag:",Enable storage provider config"`
}
