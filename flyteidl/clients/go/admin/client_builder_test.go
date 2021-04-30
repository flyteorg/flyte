package admin

import (
	"context"
	"reflect"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/pkce"
	"github.com/stretchr/testify/assert"
)

func TestClientsetBuilder_Build(t *testing.T) {
	cb := NewClientsetBuilder().WithConfig(&Config{
		UseInsecureConnection: true,
	}).WithTokenCache(&pkce.TokenCacheInMemoryProvider{})
	_, err := cb.Build(context.Background())
	assert.NoError(t, err)
	assert.True(t, reflect.TypeOf(cb.tokenCache) == reflect.TypeOf(&pkce.TokenCacheInMemoryProvider{}))
}
