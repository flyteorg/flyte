package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/clients/go/admin/pkce"
	pkcemocks "github.com/flyteorg/flyteidl/clients/go/admin/pkce/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
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

	t.Run("illegal", func(t *testing.T) {
		once = sync.Once{}
		assert.NotNil(t, InitializeAdminClient(ctx, &Config{}))
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
	})

	t.Run("legal-from-config", func(t *testing.T) {
		clientSet, err := initializeClients(ctx, &Config{InsecureSkipVerify: true}, nil)
		assert.NoError(t, err)
		assert.NotNil(t, clientSet)
		assert.NotNil(t, clientSet.AuthMetadataClient())
		assert.NotNil(t, clientSet.AdminClient())
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
	t.Run("error during oauth2Metatdata", func(t *testing.T) {
		mockAuthClient := new(mocks.AuthMetadataServiceClient)
		mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
		dialOption, err := getAuthenticationDialOption(ctx, adminServiceConfig, nil, mockAuthClient)
		assert.Nil(t, dialOption)
		assert.NotNil(t, err)
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
	plan, _ := ioutil.ReadFile("pkce/testdata/token.json")
	var tokenData oauth2.Token
	err := json.Unmarshal(plan, &tokenData)
	assert.NoError(t, err)
	tokenData.Expiry = time.Now().Add(time.Minute)
	t.Run("cache hit", func(t *testing.T) {
		mockTokenCache := new(pkcemocks.TokenCache)
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
		mockTokenCache := new(pkcemocks.TokenCache)
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
		plan, _ := ioutil.ReadFile("pkce/testdata/token.json")
		var tokenData oauth2.Token
		err := json.Unmarshal(plan, &tokenData)
		assert.NoError(t, err)

		// populate the cache
		tokenCache := &pkce.TokenCacheInMemoryProvider{}
		assert.NoError(t, tokenCache.SaveToken(&tokenData))

		orchestrator, err := pkce.NewTokenOrchestrator(ctx, pkce.Config{}, tokenCache, mockAuthClient)
		assert.NoError(t, err)

		http.DefaultServeMux = http.NewServeMux()
		dialOption, err := GetPKCEAuthTokenSource(ctx, orchestrator)
		assert.Nil(t, dialOption)
		assert.Error(t, err)
	})
}

func ExampleClientSetBuilder() {
	ctx := context.Background()
	// Create a client set that initializes the connection with flyte admin and sets up Auth (if needed).
	// See AuthType for a list of supported authentication types.
	clientSet, err := NewClientsetBuilder().WithConfig(GetConfig(ctx)).Build(ctx)
	if err != nil {
		logger.Fatalf(ctx, "failed to initialize clientSet from config. Error: %v", err)
	}

	// Access and use the desired client:
	_ = clientSet.AdminClient()
	_ = clientSet.AuthMetadataClient()
	_ = clientSet.IdentityClient()
}
