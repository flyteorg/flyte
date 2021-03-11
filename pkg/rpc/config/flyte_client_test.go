package config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authConfig "github.com/flyteorg/flyteadmin/pkg/auth/config"
	"github.com/flyteorg/flyteadmin/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestHandleFlyteCliConfigFunc(t *testing.T) {
	testClientID := "12345"
	testRedirectURI := "localhost:12345/callback"
	testAuthMetadataKey := "flyte-authorization"

	handleFlyteCliConfigFunc := HandleFlyteCliConfigFunc(context.Background(), &config.ServerConfig{
		Security: config.ServerSecurityOptions{
			Oauth: authConfig.OAuthOptions{
				GrpcAuthorizationHeader: testAuthMetadataKey,
			},
		},
		ThirdPartyConfig: config.ThirdPartyConfigOptions{
			FlyteClientConfig: config.FlyteClientConfig{
				ClientID:    testClientID,
				RedirectURI: testRedirectURI,
			},
		},
	})

	responseRecorder := httptest.NewRecorder()
	handleFlyteCliConfigFunc(responseRecorder, nil)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseBody := responseRecorder.Body
	var responseBodyMap map[string]string
	err := json.Unmarshal(responseBody.Bytes(), &responseBodyMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body with err: %v", err)
	}
	assert.EqualValues(t, map[string]string{
		clientID:        testClientID,
		redirectURI:     testRedirectURI,
		authMetadataKey: testAuthMetadataKey,
		scopes:          "",
	}, responseBodyMap)
}
