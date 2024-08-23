package admin

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"testing"

	adminMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

func TestClientsetBuilder_Build(t *testing.T) {
	httpPort := rand.IntnRange(10000, 60000)
	grpcPort := rand.IntnRange(10000, 60000)
	m := &adminMocks.AuthMetadataServiceServer{}
	m.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{
		AuthorizationEndpoint: fmt.Sprintf("http://localhost:%d/oauth2/authorize", httpPort),
		TokenEndpoint:         fmt.Sprintf("http://localhost:%d/oauth2/token", httpPort),
		JwksUri:               fmt.Sprintf("http://localhost:%d/oauth2/jwks", httpPort),
	}, nil)
	m.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(&service.PublicClientAuthConfigResponse{
		Scopes: []string{"all"},
	}, nil)
	u, err := url.Parse(fmt.Sprintf("dns:///localhost:%d", grpcPort))
	assert.NoError(t, err)

	s := newAuthMetadataServer(t, grpcPort, httpPort, m)
	ctx := context.Background()
	assert.NoError(t, s.Start(ctx))
	defer s.Close()

	cb := NewClientsetBuilder().WithConfig(&Config{
		UseInsecureConnection: true,
		Endpoint:              config.URL{URL: *u},
	}).WithTokenCache(cache.NewTokenCacheInMemoryProvider())
	_, err = cb.Build(ctx)
	assert.NoError(t, err)
	assert.True(t, reflect.TypeOf(cb.tokenCache) == reflect.TypeOf(cache.NewTokenCacheInMemoryProvider()))
}
