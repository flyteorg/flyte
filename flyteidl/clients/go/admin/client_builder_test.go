package admin

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteidl/clients/go/admin/cache"
)

func TestClientsetBuilder_Build(t *testing.T) {
	cb := NewClientsetBuilder().WithConfig(&Config{
		UseInsecureConnection: true,
	}).WithTokenCache(&cache.TokenCacheInMemoryProvider{})
	_, err := cb.Build(context.Background())
	assert.NoError(t, err)
	assert.True(t, reflect.TypeOf(cb.tokenCache) == reflect.TypeOf(&cache.TokenCacheInMemoryProvider{}))
}
