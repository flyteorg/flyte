package deviceflow

import "github.com/flyteorg/flytestdlib/config"

// Config defines settings used for Device orchestration flow.
type Config struct {
	TokenRefreshGracePeriod config.Duration `json:"refreshTime" pflag:",grace period from the token expiry after which it would refresh the token."`
	Timeout                 config.Duration `json:"timeout" pflag:",amount of time the device flow should complete or else it will be cancelled."`
	PollInterval            config.Duration `json:"pollInterval" pflag:",amount of time the device flow would poll the token endpoint if auth server doesn't return a polling interval. Okta and google IDP do return an interval'"`
}
