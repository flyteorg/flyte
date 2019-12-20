package config

// This struct encapsulates config options for bootstrapping various Flyte applications with config values
// For example, FlyteClientConfig contains application-specific values to initialize the config required by flyte client
type ThirdPartyConfigOptions struct {
	FlyteClientConfig FlyteClientConfig `json:"flyteClient"`
}

type FlyteClientConfig struct {
	ClientID    string `json:"clientId" pflag:",public identifier for the app which handles authorization for a Flyte deployment"`
	RedirectURI string `json:"redirectUri" pflag:",This is the callback uri registered with the app which handles authorization for a Flyte deployment"`
}
