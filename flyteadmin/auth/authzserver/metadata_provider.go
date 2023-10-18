package authzserver

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/auth"
	authConfig "github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteadmin/pkg/async"
	flyteErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type OAuth2MetadataProvider struct {
	cfg *authConfig.Config
}

// Override auth func to enforce anonymous access on the implemented APIs
// Ref: https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/auth/auth.go#L31
func (s OAuth2MetadataProvider) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}

func (s OAuth2MetadataProvider) GetOAuth2Metadata(ctx context.Context, r *service.OAuth2MetadataRequest) (*service.OAuth2MetadataResponse, error) {
	switch s.cfg.AppAuth.AuthServerType {
	case authConfig.AuthorizationServerTypeSelf:
		u := auth.GetPublicURL(ctx, nil, s.cfg)
		doc := &service.OAuth2MetadataResponse{
			Issuer:                        GetIssuer(ctx, nil, s.cfg),
			AuthorizationEndpoint:         u.ResolveReference(authorizeRelativeURL).String(),
			TokenEndpoint:                 u.ResolveReference(tokenRelativeURL).String(),
			JwksUri:                       u.ResolveReference(jsonWebKeysURL).String(),
			CodeChallengeMethodsSupported: []string{"S256"},
			ResponseTypesSupported: []string{
				"code",
				"token",
				"code token",
			},
			GrantTypesSupported: supportedGrantTypes,
			ScopesSupported:     []string{auth.ScopeAll},
			TokenEndpointAuthMethodsSupported: []string{
				"client_secret_basic",
			},
		}

		return doc, nil
	default:
		baseURL := s.cfg.UserAuth.OpenID.BaseURL
		if len(s.cfg.AppAuth.ExternalAuthServer.BaseURL.String()) > 0 {
			baseURL = s.cfg.AppAuth.ExternalAuthServer.BaseURL
		}

		// issuer urls, conventionally, do not end with a '/', however, metadata urls are usually relative of those.
		// This adds a '/' to ensure ResolveReference behaves intuitively.
		baseURL.Path = strings.TrimSuffix(baseURL.Path, "/") + "/"

		var externalMetadataURL *url.URL
		if len(s.cfg.AppAuth.ExternalAuthServer.MetadataEndpointURL.String()) > 0 {
			externalMetadataURL = baseURL.ResolveReference(&s.cfg.AppAuth.ExternalAuthServer.MetadataEndpointURL.URL)
		} else {
			externalMetadataURL = baseURL.ResolveReference(oauth2MetadataEndpoint)
		}

		httpClient := &http.Client{}

		if len(s.cfg.HTTPProxyURL.String()) > 0 {
			// create a transport that uses the proxy
			transport := &http.Transport{
				Proxy: http.ProxyURL(&s.cfg.HTTPProxyURL.URL),
			}
			httpClient.Transport = transport
		}

		response, err := sendAndRetryHttpRequest(httpClient, externalMetadataURL.String(), s.cfg.AppAuth.ExternalAuthServer.RetryAttempts, s.cfg.AppAuth.ExternalAuthServer.RetryDelayMilliseconds.Duration)
		if err != nil {
			if response != nil {
				logger.Errorf(ctx, "Failed to get oauth metadata. Error code: %v. Err: %v", response.StatusCode, err)
				return nil, flyteErrors.NewFlyteAdminError(codes.Code(response.StatusCode), "Failed to get oauth metadata.")
			}
			logger.Errorf(ctx, "Failed to get oauth metadata. Err: %v", err)
			return nil, flyteErrors.NewFlyteAdminError(codes.Code(500), "Failed to get oauth metadata.")
		}

		raw, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		resp := &service.OAuth2MetadataResponse{}
		err = unmarshalResp(response, raw, resp)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

func (s OAuth2MetadataProvider) GetPublicClientConfig(context.Context, *service.PublicClientAuthConfigRequest) (*service.PublicClientAuthConfigResponse, error) {
	return &service.PublicClientAuthConfigResponse{
		ClientId:                 s.cfg.AppAuth.ThirdParty.FlyteClientConfig.ClientID,
		RedirectUri:              s.cfg.AppAuth.ThirdParty.FlyteClientConfig.RedirectURI,
		Scopes:                   s.cfg.AppAuth.ThirdParty.FlyteClientConfig.Scopes,
		AuthorizationMetadataKey: s.cfg.GrpcAuthorizationHeader,
		Audience:                 s.cfg.AppAuth.ThirdParty.FlyteClientConfig.Audience,
	}, nil
}

func NewService(config *authConfig.Config) OAuth2MetadataProvider {
	return OAuth2MetadataProvider{
		cfg: config,
	}
}

func sendAndRetryHttpRequest(client *http.Client, url string, retryAttempts int, retryDelay time.Duration) (*http.Response, error) {
	var response *http.Response
	var err error
	err = async.RetryOnSpecificErrorCodes(retryAttempts+1, retryDelay, func() (*http.Response, error) {
		response, err = client.Get(url)
		return response, err
	}, isTransientErrorCode)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return response, errors.New(fmt.Sprint("Failed to get oauth metadata"))
	}

	return response, nil
}

func isTransientErrorCode(resp *http.Response) bool {
	if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
		return true
	}
	return false
}
