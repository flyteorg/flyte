package admin

import (
	"context"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

func TestClientsetBuilder_Build(t *testing.T) {
	u, _ := url.Parse("localhost:8089")
	cb := NewClientsetBuilder().WithConfig(&Config{
		UseInsecureConnection: true,
		Endpoint:              config.URL{URL: *u},
	}).WithTokenCache(&cache.TokenCacheInMemoryProvider{})
	ctx := context.Background()
	_, err := cb.Build(ctx)
	assert.NoError(t, err)
	assert.True(t, reflect.TypeOf(cb.tokenCache) == reflect.TypeOf(&cache.TokenCacheInMemoryProvider{}))
}
