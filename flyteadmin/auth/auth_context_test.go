package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/flyteadmin/auth/config"
)

func TestOAuth2ClientConfig(t *testing.T) {
	authCtx := Context{
		oauth2Client: &oauth2.Config{},
	}

	type test struct {
		name                string
		url                 string
		expectedRedirectURL string
	}
	tests := []test{
		{
			name:                "simple publicUrl",
			url:                 "https://flyte.com",
			expectedRedirectURL: "https://flyte.com/callback",
		},
		{
			name:                "custom subpath",
			url:                 "https://flyte.com/custom-subpath/console",
			expectedRedirectURL: "https://flyte.com/custom-subpath/callback",
		},
		{
			name:                "complex publicUrl",
			url:                 "https://flyte.com/login?redirect_url=https://flyte.com/console/select-project",
			expectedRedirectURL: "https://flyte.com/callback",
		},
		{
			name:                "complex publicUrl with custom subpath",
			url:                 "https://flyte.com/custom-subpath/login?redirect_url=https://flyte.com/custom-subpath/console/select-project",
			expectedRedirectURL: "https://flyte.com/custom-subpath/callback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := authCtx.OAuth2ClientConfig(config.MustParseURL(tt.url))
			assert.Equal(t, tt.expectedRedirectURL, cfg.RedirectURL)
		})
	}
}
