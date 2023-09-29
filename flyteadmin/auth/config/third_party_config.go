package config

// This struct encapsulates config options for bootstrapping various Flyte applications with config values
// For example, FlyteClientConfig contains application-specific values to initialize the config required by flyte client
type ThirdPartyConfigOptions struct {
	FlyteClientConfig FlyteClientConfig `json:"flyteClient"`
}

type FlyteClientConfig struct {
	ClientID    string   `json:"clientId" pflag:",public identifier for the app which handles authorization for a Flyte deployment"`
	RedirectURI string   `json:"redirectUri" pflag:",This is the callback uri registered with the app which handles authorization for a Flyte deployment"`
	Scopes      []string `json:"scopes" pflag:",Recommended scopes for the client to request."`
	Audience    string   `json:"audience" pflag:",Audience to use when initiating OAuth2 authorization requests."`
}

func (o ThirdPartyConfigOptions) IsEmpty() bool {
	return len(o.FlyteClientConfig.ClientID) == 0 && len(o.FlyteClientConfig.RedirectURI) == 0 && len(o.FlyteClientConfig.Scopes) == 0
}
