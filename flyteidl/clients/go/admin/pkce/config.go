package pkce

import "github.com/flyteorg/flytestdlib/config"

// Config defines settings used for PKCE flow.
type Config struct {
	BrowserSessionTimeout   config.Duration `json:"timeout" pflag:",Amount of time the browser session would be active for authentication from client app."`
	TokenRefreshGracePeriod config.Duration `json:"refreshTime" pflag:",grace period from the token expiry after which it would refresh the token."`
}
