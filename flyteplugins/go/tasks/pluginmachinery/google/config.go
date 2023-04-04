package google

type TokenSourceFactoryType = string

const (
	TokenSourceTypeDefault = "default"
)

type TokenSourceFactoryConfig struct {
	// Type is type of TokenSourceFactory, possible values are 'default' or 'gke'.
	// - 'default' uses default credentials, see https://cloud.google.com/iam/docs/service-accounts#default
	Type TokenSourceFactoryType `json:"type" pflag:",Defines type of TokenSourceFactory, possible values are 'default'"`
}

func GetDefaultConfig() TokenSourceFactoryConfig {
	return TokenSourceFactoryConfig{
		Type: "default",
	}
}
