package pkce

import "github.com/flyteorg/flytestdlib/config"

// Config defines settings used for PKCE flow.
type Config struct {
	BrowserSessionTimeout   config.Duration `json:"timeout"`
	TokenRefreshGracePeriod config.Duration `json:"refreshTime"`
}
