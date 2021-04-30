package pkce

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"golang.org/x/oauth2"
)

// BuildClientConfig builds OAuth2 config from information retrieved through the anonymous auth metadata service.
func BuildClientConfig(ctx context.Context, authMetadataClient service.AuthMetadataServiceClient) (clientConf *oauth2.Config, err error) {
	var clientResp *service.PublicClientAuthConfigResponse
	if clientResp, err = authMetadataClient.GetPublicClientConfig(ctx, &service.PublicClientAuthConfigRequest{}); err != nil {
		return nil, err
	}

	var oauthMetaResp *service.OAuth2MetadataResponse
	if oauthMetaResp, err = authMetadataClient.GetOAuth2Metadata(ctx, &service.OAuth2MetadataRequest{}); err != nil {
		return nil, err
	}

	clientConf = &oauth2.Config{
		ClientID:    clientResp.ClientId,
		RedirectURL: clientResp.RedirectUri,
		Scopes:      clientResp.Scopes,
		Endpoint: oauth2.Endpoint{
			TokenURL: oauthMetaResp.TokenEndpoint,
			AuthURL:  oauthMetaResp.AuthorizationEndpoint,
		},
	}

	return clientConf, nil
}
