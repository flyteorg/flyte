package pkce

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyteidl/clients/go/admin/oauth"
	"github.com/flyteorg/flyteidl/clients/go/admin/tokenorchestrator"
)

func TestFetchFromAuthFlow(t *testing.T) {
	ctx := context.Background()
	t.Run("fetch from auth flow", func(t *testing.T) {
		tokenCache := &cache.TokenCacheInMemoryProvider{}
		orchestrator, err := NewTokenOrchestrator(tokenorchestrator.BaseTokenOrchestrator{
			ClientConfig: &oauth.Config{
				Config: &oauth2.Config{
					RedirectURL: "http://localhost:8089/redirect",
					Scopes:      []string{"code", "all"},
				},
			},
			TokenCache: tokenCache,
		}, Config{})
		assert.NoError(t, err)
		refreshedToken, err := orchestrator.FetchTokenFromAuthFlow(ctx)
		assert.Nil(t, refreshedToken)
		assert.NotNil(t, err)
	})
}
