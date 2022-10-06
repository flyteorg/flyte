package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
	_ "google.golang.org/grpc/balancer/roundrobin" //nolint

	"github.com/flyteorg/flyteidl/clients/go/admin/cache"
	cachemocks "github.com/flyteorg/flyteidl/clients/go/admin/cache/mocks"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/clients/go/admin/oauth"
	"github.com/flyteorg/flyteidl/clients/go/admin/pkce"
	"github.com/flyteorg/flyteidl/clients/go/admin/tokenorchestrator"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

func TestInitializeAndGetAdminClient(t *testing.T) {

	ctx := context.TODO()
	t.Run("legal", func(t *testing.T) {
		u, err := url.Parse("http://localhost:8089")
		assert.NoError(t, err)
		assert.NotNil(t, InitializeAdminClient(ctx, &Config{
			Endpoint: config.URL{URL: *u},
		}))
	})
}

func TestInitializeMockClientset(t *testing.T) {
	c := InitializeMockClientset()
	assert.NotNil(t, c)
	assert.NotNil(t, c.adminServiceClient)
	assert.NotNil(t, c.authMetadataServiceClient)
}

func TestInitializeMockAdminClient(t *testing.T) {
	c := InitializeMockAdminClient()
	assert.NotNil(t, c)
}

func TestGetAdditionalAdminClientConfigOptions(t *testing.T) {
	u, _ := url.Parse("localhost:8089")
	adminServiceConfig := &Config{
		Endpoint:              config.URL{URL: *u},
		UseInsecureConnection: true,
		PerRetryTimeout:       config.Duration{Duration: 1 * time.Second},
	}

	assert.NoError(t, SetConfig(adminServiceConfig))

	ctx := context.Background()
	t.Run("legal", func(t *testing.T) {
		u, err := url.Parse("http://localhost:8089")
		assert.NoError(t, err)
		clientSet, err := ClientSetBuilder().WithConfig(&Config{Endpoint: config.URL{URL: *u}}).Build(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, clientSet)
		assert.NotNil(t, clientSet.AdminClient())
		assert.NotNil(t, clientSet.AuthMetadataClient())
		assert.NotNil(t, clientSet.IdentityClient())
		assert.NotNil(t, clientSet.HealthServiceClient())
	})

	t.Run("legal-from-config", func(t *testing.T) {
		clientSet, err := initializeClients(ctx, &Config{InsecureSkipVerify: true}, nil)
		assert.NoError(t, err)
		assert.NotNil(t, clientSet)
		assert.NotNil(t, clientSet.AuthMetadataClient())
		assert.NotNil(t, clientSet.AdminClient())
		assert.NotNil(t, clientSet.HealthServiceClient())
	})
	t.Run("legal-from-config-with-cacerts", func(t *testing.T) {
		clientSet, err := initializeClients(ctx, &Config{CACertFilePath: "testdata/root.pem"}, nil)
		assert.NoError(t, err)
		assert.NotNil(t, clientSet)
		assert.NotNil(t, clientSet.AuthMetadataClient())
		assert.NotNil(t, clientSet.AdminClient())
	})
	t.Run("legal-from-config-with-invalid-cacerts", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		newAdminServiceConfig := &Config{
			Endpoint:              config.URL{URL: *u},
			UseInsecureConnection: false,
			CACertFilePath:        "testdata/non-existent.pem",
			PerRetryTimeout:       config.Duration{Duration: 1 * time.Second},
		}

		assert.NoError(t, SetConfig(newAdminServiceConfig))
		clientSet, err := initializeClients(ctx, newAdminServiceConfig, nil)
		assert.NotNil(t, err)
		assert.Nil(t, clientSet)
	})
}

func TestGetAuthenticationDialOptionClientSecret(t *testing.T) {
	ctx := context.Background()

	u, _ := url.Parse("localhost:8089")
	adminServiceConfig := &Config{
		ClientSecretLocation:  "testdata/secret_key",
		Endpoint:              config.URL{URL: *u},
		UseInsecureConnection: true,
		AuthType:              AuthTypeClientSecret,
		PerRetryTimeout:       config.Duration{Duration: 1 * time.Second},
	}
	t.Run("legal", func(t *testing.T) {
		metatdata := &service.OAuth2MetadataResponse{
			TokenEndpoint:   "http://localhost:8089/token",
			ScopesSupported: []string{"code", "all"},
		}
		clientMetatadata := &service.PublicClientAuthConfigResponse{
			AuthorizationMetadataKey: "flyte_authorization",
		}
		mockAuthClient := new(mocks.AuthMetadataServiceClient)
		mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(metatdata, nil)
		mockAuthClient.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(clientMetatadata, nil)
		dialOption, err := getAuthenticationDialOption(ctx, adminServiceConfig, nil, mockAuthClient)
		assert.Nil(t, dialOption)
		assert.NotNil(t, err)
	})
	t.Run("legal-no-external-calls", func(t *testing.T) {
		mockAuthClient := new(mocks.AuthMetadataServiceClient)
		mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(nil, errors.New("unexpected call to get oauth2 metadata"))
		mockAuthClient.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(nil, errors.New("unexpected call to get public client config"))
		var adminCfg Config
		err := copier.Copy(&adminCfg, adminServiceConfig)
		assert.NoError(t, err)
		adminCfg.TokenURL = "http://localhost:1000/api/v1/token"
		adminCfg.Scopes = []string{"all"}
		adminCfg.AuthorizationHeader = "authorization"
		dialOption, err := getAuthenticationDialOption(ctx, &adminCfg, nil, mockAuthClient)
		assert.Nil(t, dialOption)
		assert.NotNil(t, err)
	})
	t.Run("error during oauth2Metatdata", func(t *testing.T) {
		mockAuthClient := new(mocks.AuthMetadataServiceClient)
		mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		dialOption, err := getAuthenticationDialOption(ctx, adminServiceConfig, nil, mockAuthClient)
		assert.Nil(t, dialOption)
		assert.NotNil(t, err)
	})
	t.Run("error during public client config", func(t *testing.T) {
		mockAuthClient := new(mocks.AuthMetadataServiceClient)
		mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(nil, errors.New("unexpected call to get oauth2 metadata"))
		failedPublicClientConfigLookup := errors.New("expected err")
		mockAuthClient.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(nil, failedPublicClientConfigLookup)
		var adminCfg Config
		err := copier.Copy(&adminCfg, adminServiceConfig)
		assert.NoError(t, err)
		adminCfg.TokenURL = "http://localhost:1000/api/v1/token"
		adminCfg.Scopes = []string{"all"}
		tokenProvider := ClientCredentialsTokenSourceProvider{}
		dialOption, err := getAuthenticationDialOption(ctx, &adminCfg, tokenProvider, mockAuthClient)
		assert.Nil(t, dialOption)
		assert.EqualError(t, err, "failed to fetch client metadata. Error: expected err")
	})
	t.Run("error during flyte client", func(t *testing.T) {
		metatdata := &service.OAuth2MetadataResponse{
			TokenEndpoint:   "/token",
			ScopesSupported: []string{"code", "all"},
		}
		mockAuthClient := new(mocks.AuthMetadataServiceClient)
		mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(metatdata, nil)
		mockAuthClient.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		dialOption, err := getAuthenticationDialOption(ctx, adminServiceConfig, nil, mockAuthClient)
		assert.Nil(t, dialOption)
		assert.NotNil(t, err)
	})
	incorrectSecretLocConfig := &Config{
		ClientSecretLocation:  "testdata/secret_key_invalid",
		Endpoint:              config.URL{URL: *u},
		UseInsecureConnection: true,
		AuthType:              AuthTypeClientSecret,
		PerRetryTimeout:       config.Duration{Duration: 1 * time.Second},
	}
	t.Run("incorrect client secret loc", func(t *testing.T) {
		metatdata := &service.OAuth2MetadataResponse{
			TokenEndpoint:   "http://localhost:8089/token",
			ScopesSupported: []string{"code", "all"},
		}
		clientMetatadata := &service.PublicClientAuthConfigResponse{
			AuthorizationMetadataKey: "flyte_authorization",
		}
		mockAuthClient := new(mocks.AuthMetadataServiceClient)
		mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(metatdata, nil)
		mockAuthClient.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(clientMetatadata, nil)
		dialOption, err := getAuthenticationDialOption(ctx, incorrectSecretLocConfig, nil, mockAuthClient)
		assert.Nil(t, dialOption)
		assert.NotNil(t, err)
	})
}

func TestGetAuthenticationDialOptionPkce(t *testing.T) {
	ctx := context.Background()

	u, _ := url.Parse("localhost:8089")
	adminServiceConfig := &Config{
		Endpoint:              config.URL{URL: *u},
		UseInsecureConnection: true,
		AuthType:              AuthTypePkce,
		PerRetryTimeout:       config.Duration{Duration: 1 * time.Second},
	}
	metatdata := &service.OAuth2MetadataResponse{
		TokenEndpoint:   "http://localhost:8089/token",
		ScopesSupported: []string{"code", "all"},
	}
	clientMetatadata := &service.PublicClientAuthConfigResponse{
		AuthorizationMetadataKey: "flyte_authorization",
		RedirectUri:              "http://localhost:54545/callback",
	}
	http.DefaultServeMux = http.NewServeMux()
	plan, _ := os.ReadFile("tokenorchestrator/testdata/token.json")
	var tokenData oauth2.Token
	err := json.Unmarshal(plan, &tokenData)
	assert.NoError(t, err)
	tokenData.Expiry = time.Now().Add(time.Minute)
	t.Run("cache hit", func(t *testing.T) {
		mockTokenCache := new(cachemocks.TokenCache)
		mockAuthClient := new(mocks.AuthMetadataServiceClient)
		mockTokenCache.OnGetTokenMatch().Return(&tokenData, nil)
		mockTokenCache.OnSaveTokenMatch(mock.Anything).Return(nil)
		mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(metatdata, nil)
		mockAuthClient.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(clientMetatadata, nil)
		tokenSourceProvider, err := NewTokenSourceProvider(ctx, adminServiceConfig, mockTokenCache, mockAuthClient)
		assert.Nil(t, err)
		dialOption, err := getAuthenticationDialOption(ctx, adminServiceConfig, tokenSourceProvider, mockAuthClient)
		assert.NotNil(t, dialOption)
		assert.Nil(t, err)
	})
	tokenData.Expiry = time.Now().Add(-time.Minute)
	t.Run("cache miss auth failure", func(t *testing.T) {
		mockTokenCache := new(cachemocks.TokenCache)
		mockAuthClient := new(mocks.AuthMetadataServiceClient)
		mockTokenCache.OnGetTokenMatch().Return(&tokenData, nil)
		mockTokenCache.OnSaveTokenMatch(mock.Anything).Return(nil)
		mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(metatdata, nil)
		mockAuthClient.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(clientMetatadata, nil)
		tokenSourceProvider, err := NewTokenSourceProvider(ctx, adminServiceConfig, mockTokenCache, mockAuthClient)
		assert.Nil(t, err)
		dialOption, err := getAuthenticationDialOption(ctx, adminServiceConfig, tokenSourceProvider, mockAuthClient)
		assert.Nil(t, dialOption)
		assert.NotNil(t, err)
	})
}

func Test_getPkceAuthTokenSource(t *testing.T) {
	ctx := context.Background()
	mockAuthClient := new(mocks.AuthMetadataServiceClient)
	metatdata := &service.OAuth2MetadataResponse{
		TokenEndpoint:   "http://localhost:8089/token",
		ScopesSupported: []string{"code", "all"},
	}

	clientMetatadata := &service.PublicClientAuthConfigResponse{
		AuthorizationMetadataKey: "flyte_authorization",
		RedirectUri:              "http://localhost:54546/callback",
	}

	mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(metatdata, nil)
	mockAuthClient.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(clientMetatadata, nil)

	t.Run("cached token expired", func(t *testing.T) {
		plan, _ := ioutil.ReadFile("tokenorchestrator/testdata/token.json")
		var tokenData oauth2.Token
		err := json.Unmarshal(plan, &tokenData)
		assert.NoError(t, err)

		// populate the cache
		tokenCache := &cache.TokenCacheInMemoryProvider{}
		assert.NoError(t, tokenCache.SaveToken(&tokenData))

		baseOrchestrator := tokenorchestrator.BaseTokenOrchestrator{
			ClientConfig: &oauth.Config{
				Config: &oauth2.Config{
					RedirectURL: "http://localhost:8089/redirect",
					Scopes:      []string{"code", "all"},
				},
			},
			TokenCache: tokenCache,
		}

		orchestrator, err := pkce.NewTokenOrchestrator(baseOrchestrator, pkce.Config{})
		assert.NoError(t, err)

		http.DefaultServeMux = http.NewServeMux()
		dialOption, err := GetPKCEAuthTokenSource(ctx, orchestrator)
		assert.Nil(t, dialOption)
		assert.Error(t, err)
	})
}

func TestGetDefaultServiceConfig(t *testing.T) {
	u, _ := url.Parse("localhost:8089")
	adminServiceConfig := &Config{
		Endpoint:             config.URL{URL: *u},
		DefaultServiceConfig: `{"loadBalancingConfig": [{"round_robin":{}}]}`,
	}

	assert.NoError(t, SetConfig(adminServiceConfig))

	ctx := context.Background()
	t.Run("legal", func(t *testing.T) {
		u, err := url.Parse("http://localhost:8089")
		assert.NoError(t, err)
		clientSet, err := ClientSetBuilder().WithConfig(&Config{Endpoint: config.URL{URL: *u}, DefaultServiceConfig: `{"loadBalancingConfig": [{"round_robin":{}}]}`}).Build(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, clientSet)
		assert.NotNil(t, clientSet.AdminClient())
		assert.NotNil(t, clientSet.AuthMetadataClient())
		assert.NotNil(t, clientSet.IdentityClient())
		assert.NotNil(t, clientSet.HealthServiceClient())
	})
	t.Run("illegal default service config", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()

		u, err := url.Parse("http://localhost:8089")
		assert.NoError(t, err)
		clientSet, err := ClientSetBuilder().WithConfig(&Config{Endpoint: config.URL{URL: *u}, DefaultServiceConfig: `{"loadBalancingConfig": [{"foo":{}}]}`}).Build(ctx)
		assert.Error(t, err)
		assert.Nil(t, clientSet)
	})
}

func ExampleClientSetBuilder() {
	ctx := context.Background()
	// Create a client set that initializes the connection with flyte admin and sets up Auth (if needed).
	// See AuthType for a list of supported authentication types.
	clientSet, err := NewClientsetBuilder().WithConfig(GetConfig(ctx)).Build(ctx)
	if err != nil {
		logger.Fatalf(ctx, "failed to initialized clientSet from config. Error: %v", err)
	}

	// Access and use the desired client:
	_ = clientSet.AdminClient()
	_ = clientSet.AuthMetadataClient()
	_ = clientSet.IdentityClient()
}
