package pkce

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/oauth"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/tokenorchestrator"
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
		// Throws an exitError since browser cannot be opened.
		// But this makes sure that at least code is tested till the browser action is invoked.
		// Earlier change had broken this by introducing a hard dependency on the http Client from the context
		_, isAnExitError := err.(*exec.ExitError)
		assert.True(t, isAnExitError)
	})
}
